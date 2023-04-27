package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	sliLayerHeight     string
	sliPrintSpeedSec   string
	sliPrinterNote     string
	sliFilamentTypes   []string
	sliNozzleTemp      []string
	sliNozzleDiameters []string
	sliBedTemp         []string
	sliMinX            float64
	sliMinY            float64
	sliMaxX            float64
	sliMaxY            float64
	sliMaxZ            float64
)

func fix(in io.ReadCloser, out io.WriteCloser) {
	var (
		useV1  bool
		hasT0  bool
		hasT1  bool
		gcodes [][]byte

		lineCount = 0
		printMode = PrintModeDefault
	)

	buf := &bytes.Buffer{}
	buf.ReadFrom(in)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		line = bytes.TrimSpace(line)
		if len(line) < 1 {
			continue
		}

		if startWith(line, "; Postprocessed by smfix") {
			os.Exit(0)
		} else if startWith(line, "; SNAPMAKER_GCODE_V0") {
			useV1 = false
		} else if startWith(line, "; SNAPMAKER_GCODE_V1") {
			useV1 = true
		} else if startWith(line, "G4 S0") {
			// this is a bug with prusaslicer it's conflict with the firmware of J1@V2.2.13
			continue
		} else if startWith(line, "T0") {
			hasT0 = true
		} else if startWith(line, "T1") {
			hasT1 = true
		} else if startWith(line, "M605 S2") {
			printMode = PrintModeDuplication
		} else if startWith(line, "M605 S3") {
			printMode = PrintModeMirror
		} else if startWith(line, "M605 S4") {
			printMode = PrintModeBackup
		}
		gcodes = append(gcodes, line)
		lineCount++
	}
	in.Close()

	if lineCount < 2 {
		flag_usage()
	}

	{
		sliFilamentTypes = split(getProperty(gcodes, "filament_type"))
		if !hasT0 {
			sliFilamentTypes[0] = "-"
		}
		if !hasT1 {
			sliFilamentTypes[1] = "-"
		}

		sliNozzleDiameters = split(getProperty(gcodes, "nozzle_diameter"))
		if sliNozzleDiameters[1] == "0" {
			sliNozzleDiameters[1] = sliNozzleDiameters[0]
		}

		sliLayerHeight = getProperty(gcodes, "layer_height")
		sliPrinterNote = getProperty(gcodes, "printer_notes")
		sliPrintSpeedSec = getProperty(gcodes,
			"max_print_speed",
			"outer_wall_speed", // bbs
		)
	}

	{ // xyz
		// add '; min_x = [first_layer_print_min_0] ...' to the end-gcode
		sliMinX, _ = strconv.ParseFloat(getProperty(gcodes, "min_x"), 32)
		sliMinY, _ = strconv.ParseFloat(getProperty(gcodes, "min_y"), 32)
		sliMaxX, _ = strconv.ParseFloat(getProperty(gcodes, "max_x"), 32)
		sliMaxY, _ = strconv.ParseFloat(getProperty(gcodes, "max_y"), 32)
		sliMaxZ, _ = strconv.ParseFloat(getProperty(gcodes, "max_z"), 32)
	}

	{ // temperature
		sliNozzleTemp = split(getProperty(gcodes,
			"temperature",
			"nozzle_temperature", // bbs
		))
		sliBedTemp = split(getProperty(gcodes,
			"bed_temperature",
			"hot_plate_temp",  // bbs
			"cool_plate_temp", // bbs
			"eng_plate_temp",  // bbs
		))
	}

	if hasT0 || hasT1 || printMode != PrintModeDefault ||
		// settings: Printer Settings - Notes - add line(PRINTER_GCODE_V1)
		strings.Contains(sliPrinterNote, "PRINTER_GCODE_V1") {
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
			bedTempT0, _   = strconv.Atoi(sliBedTemp[0])
			bedTempT1, _   = strconv.Atoi(sliBedTemp[1])
			bedTemperature = max(bedTempT0, bedTempT1)

			nozzleTempT0, nozzleTempT1 = sliNozzleTemp[0], sliNozzleTemp[1]
		)

		ext := [][]byte{
			[]byte(";Version:1"),
			// []byte(";Slicer:PrusaSlicer"),
			[]byte(fmt.Sprintf(";Estimated Print Time:%d", findEstimatedTime(gcodes))),
			[]byte(fmt.Sprintf(";Lines:%d", lineCount)),
		}

		if hasT0 || hasT1 {
			ext = append(ext,
				[]byte(";Printer:Snapmaker J1"),
				[]byte(fmt.Sprintf(";Extruder Mode:%s", printMode)),
			)

			if hasT0 && hasT1 {
				ext = append(ext, []byte(";Extruder(s) Used:2"))
			} else {
				ext = append(ext, []byte(";Extruder(s) Used:1"))
				if hasT0 {
					bedTemperature = bedTempT0
					// disable T1
					nozzleTempT1 = "0"
				} else {
					bedTemperature = bedTempT1
					// disable T0
					nozzleTempT0 = "0"
				}
			}

			ext = append(ext,
				[]byte(fmt.Sprintf(";Extruder 0 Nozzle Size:%s", sliNozzleDiameters[0])),
				[]byte(fmt.Sprintf(";Extruder 0 Material:%s", sliFilamentTypes[0])),
				[]byte(fmt.Sprintf(";Extruder 0 Print Temperature:%s", nozzleTempT0)),

				[]byte(fmt.Sprintf(";Extruder 1 Nozzle Size:%s", sliNozzleDiameters[1])),
				[]byte(fmt.Sprintf(";Extruder 1 Material:%s", sliFilamentTypes[1])),
				[]byte(fmt.Sprintf(";Extruder 1 Print Temperature:%s", nozzleTempT1)),
			)

		} else {
			ext = append(ext,
				[]byte(";Printer:Snapmaker 2"),
			)
		}

		ext = append(ext,
			[]byte(fmt.Sprintf(";Bed Temperature:%d", bedTemperature)),
			[]byte(fmt.Sprintf(";Work Range - Max X:%.4f", sliMaxX)),
			[]byte(fmt.Sprintf(";Work Range - Max Y:%.4f", sliMaxY)),
			[]byte(fmt.Sprintf(";Work Range - Max Z:%.4f", sliMaxZ)),
			[]byte(fmt.Sprintf(";Work Range - Min X:%.4f", sliMinX)),
			[]byte(fmt.Sprintf(";Work Range - Min Y:%.4f", sliMinY)),
			[]byte(";Work Range - Min Z:0.0"),
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
			[]byte(fmt.Sprintf(";nozzle_temperature(°C): %s", sliNozzleTemp[0])),
			[]byte(fmt.Sprintf(";build_plate_temperature(°C): %s", sliBedTemp[0])),
			[]byte(fmt.Sprintf(";work_speed(mm/minute): %d", speed*60)),
			[]byte(fmt.Sprintf(";max_x(mm): %.4f", sliMaxX)),
			[]byte(fmt.Sprintf(";max_y(mm): %.4f", sliMaxY)),
			[]byte(fmt.Sprintf(";max_z(mm): %.4f", sliMaxZ)),
			[]byte(fmt.Sprintf(";min_x(mm): %.4f", sliMinX)),
			[]byte(fmt.Sprintf(";min_y(mm): %.4f", sliMinY)),
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
	out.Close()
}

func main() {
	var (
		in  *os.File
		out io.WriteCloser
		err error
	)
	if len(os.Args) > 1 {
		in, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}

		out, err = os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatalln(err)
		}

	} else if st, _ := os.Stdin.Stat(); (st.Mode() & os.ModeCharDevice) == 0 {
		in = os.Stdin
		out = os.Stdout
	}

	if in == nil {
		flag_usage()
	}

	fix(in, out)
}
