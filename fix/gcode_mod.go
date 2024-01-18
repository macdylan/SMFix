package fix

import (
	"fmt"
	"strconv"
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

func GcodeFixPreheat(gcodes []*GcodeBlock) (output []*GcodeBlock) {
	output = make([]*GcodeBlock, 0, len(gcodes)+512)

	hasM73 := false // must enable "Supports remaining times" in the Printer settings
	for _, gcode := range gcodes {
		output = append(output, gcode)

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
				for n := len(output) - 2; n >= 0; n-- {
					checkLine := output[n]

					// find an earlier M109 for this tool, stop looking
					if checkLine.Is("M109") { // T maybe in the comment : && checkLine.HasParam('T') {
						if checkTool, err := checkLine.GetToolNum(); err == nil && tool == checkTool {
							break
						}
					}

					// find an earlier M104 for this tool, determine where to place pre-heat command
					if checkLine.Is("M104") { // && checkLine.HasParam('T') {
						if checkTool, err := checkLine.GetToolNum(); err == nil && tool == checkTool && hasM73 {
							if checkLine.InComment("cooldown") || checkLine.InComment(fmt.Sprintf("standby T%d", tool)) {
								// build preheat command
								preheat, _ := ParseGcodeBlock(fmt.Sprintf("M104 T%d S%g", checkTool, preheatTemp))

								if remainingDiffCount < PreheatShort {
									nl, _ := ParseGcodeBlock(fmt.Sprintf(";(Fixed: remove cooldown: %s)", checkLine.Format("%c %p")))
									output[n] = nl
								} else if remainingDiffCount < PreheatLong {
									preheat.SetComment(";(Fixed: pre-heat short)")
									output = append(output[:nShortPreheat], append([]*GcodeBlock{preheat}, output[nShortPreheat:]...)...)
								} else {
									preheat.SetComment(";(Fixed: pre-heat long)")
									output = append(output[:nLongPreheat], append([]*GcodeBlock{preheat}, output[nLongPreheat:]...)...)

									deepfreeze, _ := ParseGcodeBlock(fmt.Sprintf(
										"M104 T%d S110 ;(Fixed: deep freeze instead of: %s)",
										tool, checkLine.Format("%c %p")))
									output[n] = deepfreeze
								}
							}
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
									nShortPreheat = n
								} else if remainingDiffCount == PreheatLong {
									nLongPreheat = n
								}
							}
						}
					} // m73Match
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
	for n, line := range output {
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
					output[n] = requested
				case "109":
					if curToolTempGuaranteed[tool] {
						stabilized, _ := ParseGcodeBlock(fmt.Sprintf(";(Fixed: already stabilized temp: %s)", line.Format("%c %p")))
						output[n] = stabilized
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

	return output
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
	nGcodes := len(gcodes)
	mutex := sync.Mutex{}
	work := func(wi, wn int) {
		for n := wi; n < nGcodes; n += wn {
			gcode := gcodes[n]
			switch gcode.Cmd().Word() {
			case 'T':
				{
					tool, _ := gcode.GetToolNum()
					mutex.Lock()
					gcodes[n].Cmd().SetAddr(tool % 2)
					mutex.Unlock()
				}
			case 'M':
				{
					tool, _ := gcode.GetToolNum()
					str_tool := strconv.Itoa(int(tool) % 2)
					switch gcode.Cmd().Addr() {
					case "106", "107": // fan use P
						if gcode.HasParam('P') {
							mutex.Lock()
							gcodes[n].SetParam('P', str_tool)
							mutex.Unlock()
						}

					case "104", "109": // temp use T
						if gcode.HasParam('T') {
							mutex.Lock()
							gcodes[n].SetParam('T', str_tool)
							mutex.Unlock()
						}

					case "301", "303":
						if gcode.HasParam('E') {
							mutex.Lock()
							gcodes[n].SetParam('E', str_tool)
							mutex.Unlock()
						}

					}
				} // M
			} // switch word
		} // for
	} // work

	GoInParallelAndWait(work)
	return gcodes
}
