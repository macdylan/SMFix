package fix

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type GcodeModifier func([]*GcodeBlock) []*GcodeBlock

func GcodeFixShutoff(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	nGcodes := len(gcodes)
	output = make([]*GcodeBlock, 0, nGcodes+8)

	// track of the last line each tool was used
	toolLastLine := make(map[int32]int)
	toolShutted := make(map[int32]bool)
	mutex := sync.Mutex{}

	// look for tool changes
	work := func(wi, wn int) {
		for n := wi; n < nGcodes; n += wn {
			gcode := gcodes[n]
			cmd := gcode.Cmd()
			if cmd.Word() == 'T' {
				var tool int32
				if err := cmd.AddrAs(&tool); err == nil {
					mutex.Lock()
					if toolLastLine[tool] < n {
						toolLastLine[tool] = n
					}
					mutex.Unlock()
				}
			}
		}
	}
	GoInParallelAndWait(work)

	// add M104 S0
	var (
		curTool int32 = -1
		offset  int
	)
	for n, line := range gcodes {
		output = append(output, line)
		cmd := line.Cmd()

		if cmd.Word() == 'T' {
			var nextTool int32
			cmd.AddrAs(&nextTool)

			if curTool != -1 && curTool != nextTool {
				if toolLastLine[curTool] < n {
					if g, err := ParseGcodeBlock(fmt.Sprintf("M104 S0 T%d ; (Fixed: Shutoff T%d)", curTool, curTool)); err == nil {
						output = append(output, g)
						toolShutted[curTool] = true
						offset++
					}
				}
			}
			curTool = nextTool
		}

		// M104 or M109
		if cmd.Is("M104") || cmd.Is("M109") {
			var (
				tool int32
				temp float32
				err  error
			)
			if tool, err = line.GetToolNum(); err == nil {
				if err = line.GetParam('R', &temp); err != nil {
					err = line.GetParam('S', &temp)
				}
				if err == nil && temp > 0 && n > toolLastLine[tool] && toolShutted[tool] {
					for i := n + offset; i > n; i-- {
						if g, err := ParseGcodeBlock(fmt.Sprintf(";(Fixed: T%d has been shutted off: %s)", tool, output[i].Format("%c %p"))); err == nil {
							output[i] = g
						}
					}
				}
			}
		}
	}

	return output
}

var (
	PreheatShort int64 = 1
	PreheatLong  int64 = 3
)

func GcodeFixPreheat(gcodes []*GcodeBlock) []*GcodeBlock {
	hasM73 := false // must enable "Supports remaining times" in the Printer settings
	for n, gcode := range gcodes {

		if gcode.Is("M109") && gcode.HasParam('S') {
			if tool, err := gcode.GetToolNum(); err == nil {
				var (
					preheatTemp        float32
					remaining          float32 = -1
					remainingDiffCount int64   = 0
					nShortPreheat              = -1
					nLongPreheat               = -1
				)
				gcode.GetParam('S', &preheatTemp)

				// find line to place preheat command
				for pn := n - 1; pn >= 0; pn-- {
					checkLine := gcodes[pn]

					// find an earlier M109 for this tool, stop looking
					if checkLine.Is("M109") { // T maybe in the comment : && checkLine.HasParam('T') {
						if checkTool, err := checkLine.GetToolNum(); err == nil && tool == checkTool {
							break
						}
					}

					if checkLine.Is("M73") && checkLine.HasParam('R') {
						hasM73 = true
						var remain float32
						if err := checkLine.GetParam('R', &remain); err == nil {
							if remaining == -1 { // first remaining
								remaining = remain

							} else if remaining != remain {
								remaining = remain
								remainingDiffCount++
								if remainingDiffCount == PreheatShort {
									nShortPreheat = pn
								} else if remainingDiffCount == PreheatLong {
									nLongPreheat = pn
								}
							}
						}
					} // m73Match

					// find an earlier M104 for this tool, determine where to place pre-heat command
					if checkLine.Is("M104") { // && checkLine.HasParam('T') {
						if checkTool, err := checkLine.GetToolNum(); err == nil && tool == checkTool && hasM73 {
							if checkLine.InComment("cooldown") || checkLine.InComment(fmt.Sprintf("standby T%d", tool)) {
								// build preheat command
								preheat, _ := ParseGcodeBlock(fmt.Sprintf("M104 T%d S%g", checkTool, preheatTemp))

								if remainingDiffCount < PreheatShort {
									nl, _ := ParseGcodeBlock(fmt.Sprintf(";(Fixed: remove cooldown: %s)", checkLine.Format("%c %p")))
									gcodes[pn] = nl
								} else if remainingDiffCount < PreheatLong {
									preheat.SetComment(";(Fixed: pre-heat short)")
									// output = append(output[:nShortPreheat], append([]*GcodeBlock{preheat}, output[nShortPreheat:]...)...)
									insertBefore(&gcodes, nShortPreheat, preheat)
								} else {
									preheat.SetComment(";(Fixed: pre-heat long)")
									// output = append(output[:nLongPreheat], append([]*GcodeBlock{preheat}, output[nLongPreheat:]...)...)
									insertBefore(&gcodes, nLongPreheat, preheat)

									deepfreeze, _ := ParseGcodeBlock(fmt.Sprintf(
										"M104 T%d S110 ;(Fixed: deep freeze instead of: %s)",
										tool, checkLine.Format("%c %p")))
									gcodes[pn] = deepfreeze
								}
							}
							break
						}
					}

				}
			}
		} // m109Match
	}

	if !hasM73 {
		// Do not modify anything if there is no M73
		return gcodes
	}

	// remove any unnecessary M104 and 109 commands
	curToolTemp := make(map[int32]float32)
	curToolTempGuaranteed := make(map[int32]bool)
	for n, line := range gcodes {
		if (line.Is("M104") || line.Is("M109")) && line.HasParam('S') {
			var (
				temp float32
				tool int32
				err  error
			)
			if tool, err = line.GetToolNum(); err != nil {
				continue
			}
			line.GetParam('S', &temp)

			if curTemp, ok := curToolTemp[tool]; ok && curTemp == temp {
				switch line.Cmd().Addr() {
				case "104":
					requested, _ := ParseGcodeBlock(fmt.Sprintf(";(Fixed: already requested temp: %s)", line.Format("%c %p")))
					gcodes[n] = requested
				case "109":
					if curToolTempGuaranteed[tool] {
						stabilized, _ := ParseGcodeBlock(fmt.Sprintf(";(Fixed: already stabilized temp: %s)", line.Format("%c %p")))
						gcodes[n] = stabilized
					} else {
						curToolTempGuaranteed[tool] = true
					}
				}
			} else {
				curToolTemp[tool] = temp
				curToolTempGuaranteed[tool] = line.Is("M109")
			}
		}
	}

	return gcodes
}

/*
func GcodeTrimLines(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	prevLineEmpty := false
	output = make([]*GcodeBlock, 0, len(gcodes))
	for _, gcode := range gcodes {
		if gcode.Format("%c %p") != "G4 S0" {
			if gcode.IsEmpty() && prevLineEmpty {
				continue
			}
			output = append(output, gcode)
			prevLineEmpty = gcode.IsEmpty()
		}
	}
	return output
}
*/

func GcodeReinforceTower(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	output = make([]*GcodeBlock, 0, len(gcodes)+2048)

	var (
		wiping bool
		e      float32
		f      float32
		cmd    *GcodeBlock
		z      float32
	)
	for _, gcode := range gcodes {
		if gcode.IsComment() {
			if gcode.InComment("; CP TOOLCHANGE WIPE") {
				wiping = true
			}
			if gcode.InComment("; CP TOOLCHANGE END") {
				wiping = false
				e = 0.0
			}
			if ele := take(gcode.Comment(), `^;Z:[\d\.]+`).taken; ele != "" {
				if v, err := strconv.ParseFloat(ele[3:], 32); err == nil {
					z = float32(v)
				}
			}
		}
		if wiping && z > 0.3 {
			if gcode.Is("G1") && gcode.HasParam('E') && gcode.HasParam('F') {
				if e < 0.01 {
					gcode.GetParam('E', &e)
					gcode.GetParam('F', &f)
					if e > 0.0 {
						e = e * 0.45
					}
					// if f > 0.0 {
					// 	f = f * 0.7
					// }
				}
				cmd, _ = ParseGcodeBlock(fmt.Sprintf("G1 E%g F%g ;(Fixed: reinforce tower)", e, f))
				output = append(output, cmd)
			}
		}
		output = append(output, gcode)
	}
	return output
}

// GcodeReplaceToolNum 查找 Gcode 中的 T/M104/M106/M107/M109 指令，将参数中的 Tnum/Pnum 替换为 num % 2 的结果
// Snapmaker 打印机最多只有2个喷嘴，T > 1 无效，但在 OrcaSlicer 中可以简化多材料的配置
// T0 -> T0, T1 -> T1
// T2 -> T0, T3 -> T1
// T4 -> T0, T5 -> T1
func GcodeReplaceToolNum(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	var (
		idxT0, idxT1 int
	)
	nGcodes := len(gcodes)
	work := func(wi, wn int) {
		for n := wi; n < nGcodes; n += wn {
			gcode := gcodes[n]
			switch gcode.Cmd().Word() {
			case 'T':
				{
					tool, _ := gcode.GetToolNum()
					if num, t := tool%2, int(tool); num == 0 {
						idxT0 = t
					} else {
						idxT1 = t
					}
					gcodes[n].Cmd().SetAddr(tool % 2)
				}
			case 'M':
				{
					tool, _ := gcode.GetToolNum()
					str_tool := strconv.Itoa(int(tool) % 2)
					switch gcode.Cmd().Addr() {
					case "106", "107": // fan use P
						if gcode.HasParam('P') {
							gcodes[n].SetParam('P', str_tool)
						}

					case "104", "109": // temp use T
						if gcode.HasParam('T') {
							gcodes[n].SetParam('T', str_tool)
						}

					case "301", "303":
						if gcode.HasParam('E') {
							gcodes[n].SetParam('E', str_tool)
						}

					}
				} // M
			} // switch word
		} // for
	} // work

	GoInParallelAndWait(work)

	// remove unused values for ParseParams()
	prefixes := []string{
		"; filament used [",
		"; filament_type = ",
		"; filament_retraction_length = ",
		"; nozzle_temperature_initial_layer = ",
		"; hot_plate_temp_initial_layer = ",
	}
	work2 := func(wi, wn int) {
		defer func() {
			if r := recover(); r != nil {
				// fmt.Println(r, comment)
			}
		}()
		for n := wi; n < nGcodes; n += wn {
			gcode := gcodes[n]
			if gcode.IsComment() {
				comment := gcode.Comment()
				if len(comment) > 15 {
					for _, prefix := range prefixes {
						if strings.HasPrefix(comment, prefix) {
							i := strings.Index(comment, "=")
							if i != -1 {
								v := comment[i+1:]
								var (
									vs        []string
									delimiter = ","
								)
								if strings.Contains(v, ";") {
									delimiter = ";"
								}
								vs = strings.Split(v, delimiter)
								l := len(vs)
								if l >= idxT0 {
									vs[0] = strings.TrimSpace(vs[idxT0])
								}
								if l >= idxT1 {
									vs[1] = strings.TrimSpace(vs[idxT1])
								}
								nv := strings.Join(vs[:2], delimiter)
								var buf bytes.Buffer
								buf.WriteString(comment[:i+2])
								buf.WriteString(nv)
								gcodes[n].SetComment(buf.String())
							}
							break
						}
					}
				}
			}
		}
	}
	GoInParallelAndWait(work2)
	return gcodes
}

func GcodeFixOrcaToolUnload(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	output = make([]*GcodeBlock, 0, len(gcodes)+2048)

	var (
		check bool
		cmd   *GcodeBlock
	)
	for _, gcode := range gcodes {
		if gcode.IsComment() {
			if gcode.InComment("; CP TOOLCHANGE START") {
				check = true
			}
			if gcode.InComment("; CP TOOLCHANGE END") {
				check = false
			}
		}
		if check && gcode.Is("M104") {
			// no tool num is an invalid cmd
			if _, err := gcode.GetToolNum(); err != nil {
				cmd, _ = ParseGcodeBlock(fmt.Sprintf(";(Fixed: remove: %s)", gcode.Format("%c %p")))
				output = append(output, cmd)
				continue
			}
		}
		output = append(output, gcode)
	}
	return output
}
