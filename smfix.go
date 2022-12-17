package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	silent  = false
	reThumb = regexp.MustCompile(`(?m)(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)`)
	in      *os.File
	out     io.Writer
)

var (
	sliLayerHeight           string
	sliPrintSpeedSec         string
	sliPrinterNotes          string
	sliFirstBedTemp          []string
	sliTemp                  []string
	sliFirstTemp             []string
	sliNozzleDiameters       []string
	sliFilamentTypes         []string
	sliFirstLayerTemperature []string
	sliBedTemperature        []string
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getProperty(gcodes [][]byte, key string) string {
	// from env
	_env := os.Getenv("SLIC3R_" + strings.ToUpper(key))
	if _env != "" {
		return _env
	}

	// from prusaslicer_config
	key_b := []byte("; " + key + " =")
	i := len(gcodes) - 1
	m := min(i-1000, 0) // tail 1k lines
	for ; m >= i; i-- {
		if 0 == bytes.Index(gcodes[i], key_b) {
			return string(gcodes[i][bytes.Index(gcodes[i], []byte("= "))+2:])
		}
	}

	return ""
}

func split(s string) []string {
	var x []string
	if strings.Contains(s, ";") {
		x = strings.Split(s, ";")
	} else {
		x = strings.Split(s, ",")
	}
	if len(x) == 1 {
		x = append(x, "0")
	}
	return x
}

func convertThumbnail(gcodes [][]byte) []byte {
	comments := bytes.NewBuffer([]byte{})
	for _, line := range gcodes {
		if len(line) > 0 && line[0] == ';' {
			comments.Write(line)
			comments.WriteRune('\n')
		}
	}
	matches := reThumb.FindAllSubmatch(comments.Bytes(), -1)
	if matches != nil {
		none := []byte(nil)
		data := matches[len(matches)-1][1]
		data = bytes.ReplaceAll(data, []byte("\r\n"), none)
		data = bytes.ReplaceAll(data, []byte("\n"), none)
		data = bytes.ReplaceAll(data, []byte("; "), none)
		b := []byte("data:image/png;base64,")
		return append(b, data...)
	}
	return nil
}

func findEstimatedTime(gcodes [][]byte) int {
	for _, line := range gcodes {
		if 0 == bytes.Index(line, []byte("; estimated printing time")) {
			est := line[bytes.Index(line, []byte("= "))+2:] // 2d 12h 8m 58s
			est = bytes.ReplaceAll(est, []byte(" "), []byte(nil))
			t := map[byte]int{'d': 0, 'h': 0, 'm': 0, 's': 0}
			for _, p := range []byte("dhms") {
				if i := bytes.IndexByte(est, p); i >= 0 {
					t[p], _ = strconv.Atoi(string(est[0:i]))
					est = est[i+1:]
				}
			}
			return t['d']*86400 +
				t['h']*3600 +
				t['m']*60 +
				t['s']
		}
	}
	return 0
}

func fix() {
	var (
		useV1 bool

		cmdL = []byte("T0")
		extL bool
		cmdR = []byte("T1")
		extR bool

		lineCount = 0
	)

	buf := &bytes.Buffer{}
	buf.ReadFrom(in)
	gcodes := [][]byte{}
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		gcodes = append(gcodes, line[0:len(line)-1])
		lineCount++
		if 0 == bytes.Index(line, cmdL) {
			extL = true
		}
		if 0 == bytes.Index(line, cmdR) {
			extR = true
		}
	}
	in.Close()

	if lineCount == 0 {
		usage()
	}

	sliLayerHeight = getProperty(gcodes, "layer_height")
	sliPrintSpeedSec = getProperty(gcodes, "max_print_speed")
	sliPrinterNotes = getProperty(gcodes, "printer_notes")

	sliTemp = split(getProperty(gcodes, "temperature"))
	sliFirstTemp = split(getProperty(gcodes, "first_layer_temperature"))
	sliFilamentTypes = split(getProperty(gcodes, "filament_type"))
	sliNozzleDiameters = split(getProperty(gcodes, "nozzle_diameter"))
	sliFirstLayerTemperature = split(getProperty(gcodes, "first_layer_temperature"))
	sliFirstBedTemp = split(getProperty(gcodes, "first_layer_bed_temperature"))
	sliBedTemperature = split(getProperty(gcodes, "bed_temperature"))

	if extL || extR ||
		// settings: Printer Settings - Notes - add line(PRINTER_GCODE_V1)
		strings.Contains(sliPrinterNotes, "PRINTER_GCODE_V1") {
		useV1 = true
	}

	speed, _ := strconv.Atoi(sliPrintSpeedSec)
	thumbnail := convertThumbnail(gcodes)
	headers := [][]byte{
		[]byte("; Postprocessed by smfix (https://github.com/macdylan/Snapmaker2Slic3rPostProcessor)"),
		[]byte(";Header Start"),
	}

	// V1
	if useV1 {
		var (
			btempL, _      = strconv.Atoi(sliBedTemperature[0])
			btempR, _      = strconv.Atoi(sliBedTemperature[1])
			bedTemperature = max(btempL, btempR)

			ptempL, ptempR = sliTemp[0], sliTemp[1]
		)

		ext := [][]byte{
			[]byte(";Version:1"),
			[]byte(";Slicer:PrusaSlicer"),
			[]byte(fmt.Sprintf(";Estimated Print Time:%.0f", float64(findEstimatedTime(gcodes))*1.07)),
			[]byte(fmt.Sprintf(";Lines:%d", lineCount)),
		}

		if extL || extR {
			ext = append(ext,
				[]byte(";Printer:Snapmaker J1"),
				[]byte(";Extruder Mode:IDEX Full Control"),
			)

			if extL && extR {
				ext = append(ext, []byte(";Extruder(s) Used:2"))
			} else {
				ext = append(ext, []byte(";Extruder(s) Used:1"))
				if extL {
					bedTemperature = btempL
					// disable R nozzle
					ptempR = "0"
				} else {
					bedTemperature = btempR
					// disable L nozzle
					ptempL = "0"
				}
			}

			ext = append(ext,
				[]byte(fmt.Sprintf(";Extruder 0 Nozzle Size:%s", sliNozzleDiameters[0])),
				[]byte(fmt.Sprintf(";Extruder 0 Material:%s", sliFilamentTypes[0])),
				[]byte(fmt.Sprintf(";Extruder 0 Print Temperature:%s", ptempL)),

				[]byte(fmt.Sprintf(";Extruder 1 Nozzle Size:%s", sliNozzleDiameters[1])),
				[]byte(fmt.Sprintf(";Extruder 1 Material:%s", sliFilamentTypes[1])),
				[]byte(fmt.Sprintf(";Extruder 1 Print Temperature:%s", ptempR)),
			)

		} else {
			ext = append(ext,
				[]byte(";Printer:Snapmaker 2"),
			)
		}

		ext = append(ext,
			[]byte(fmt.Sprintf(";Bed Temperature:%d", bedTemperature)),
			[]byte(";Work Range - Min X:0"),
			[]byte(";Work Range - Min Y:0"),
			[]byte(";Work Range - Min Z:0"),
			[]byte(";Work Range - Min X:0"),
			[]byte(";Work Range - Max Y:0"),
			[]byte(";Work Range - Max Z:0"),
			[]byte(";no thumbnail, Printer Settings / Firmware / G-code thumbnails, add '600x600'"),
			[]byte(";Header End\n\n"),
		)

		if thumbnail != nil {
			ext[len(ext)-2] = append([]byte(";Thumbnail:"), thumbnail...)
		}

		headers = append(headers, ext...)

		// V0
	} else {
		ext := [][]byte{
			[]byte(";FAVOR:Marlin"),
			[]byte(";TIME:666"),
			[]byte(fmt.Sprintf(";Layer height: %s", sliLayerHeight)),
			[]byte(";header_type: 3dp"),
			[]byte(";no thumbnail, Printer Settings / Firmware / G-code thumbnails, add '300x150'"), // slot for thumbnail
			[]byte(fmt.Sprintf(";file_total_lines: %d", lineCount+19)),
			[]byte(fmt.Sprintf(";estimated_time(s): %.0f", float64(findEstimatedTime(gcodes))*1.07)),
			[]byte(fmt.Sprintf(";nozzle_temperature(°C): %s", sliTemp[0])),
			[]byte(fmt.Sprintf(";build_plate_temperature(°C): %s", sliBedTemperature[0])),
			[]byte(fmt.Sprintf(";work_speed(mm/minute): %d", speed*60)),
			[]byte(";max_x(mm): 0"),
			[]byte(";max_y(mm): 0"),
			[]byte(";max_z(mm): 0"),
			[]byte(";min_x(mm): 0"),
			[]byte(";min_y(mm): 0"),
			[]byte(";min_z(mm): 0"),
			[]byte(";Header End\n\n"),
		}

		if thumbnail != nil {
			ext[4] = append([]byte(";thumbnail: "), thumbnail...)
		}

		headers = append(headers, ext...)
	}

	bw := bufio.NewWriter(out)
	bw.Write(bytes.Join(headers, []byte("\n")))
	bw.Write(bytes.Join(gcodes, []byte("\n")))
	bw.Flush()
}

func usage() {
	fmt.Println("smfix, optimize G-code file for Snapmaker 2.")
	fmt.Println("<https://github.com/macdylan/Snapmaker2Slic3rPostProcessor>")
	fmt.Println("Example:")
	fmt.Println("  # smfix a.gcode")
	fmt.Println("or")
	fmt.Println("  # cat a.gcode | smfix > b.gcode")
	fmt.Println("")
	os.Exit(1)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Err: %s", r)
			os.Exit(2)
		}
	}()

	if len(os.Args) > 1 {
		var err error
		in, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}

		out, err = os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println(err)
			return
		}

	} else if st, _ := os.Stdin.Stat(); (st.Mode() & os.ModeCharDevice) == 0 {
		silent = true
		in = os.Stdin
		out = os.Stdout
	}

	if in == nil {
		usage()
	}

	if !silent {
		fmt.Print("Starting SMFix ... ")
	}

	fix()

	if !silent {
		fmt.Println("done")
	}
}
