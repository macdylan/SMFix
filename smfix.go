package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

var (
	silent  = false
	reThumb = regexp.MustCompile(`(?m)(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)`)
	in      *os.File
	out     io.Writer
)

var (
	sliLayerHeight   = os.Getenv("SLIC3R_LAYER_HEIGHT")
	sliBedTemp       = os.Getenv("SLIC3R_BED_TEMPERATURE")
	sliFirstBedTemp  = os.Getenv("SLIC3R_FIRST_LAYER_BED_TEMPERATURE")
	sliTemp          = os.Getenv("SLIC3R_TEMPERATURE")
	sliFirstTemp     = os.Getenv("SLIC3R_FIRST_LAYER_TEMPERATURE")
	sliPrintSpeedSec = os.Getenv("SLIC3R_MAX_PRINT_SPEED")
)

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
	lineCount := 0
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
	}
	in.Close()

	if lineCount == 0 {
		usage()
	}

	speed, _ := strconv.Atoi(sliPrintSpeedSec)

	headers := [][]byte{
		[]byte(";Header Start"),
		[]byte(";FAVOR:Marlin"),
		[]byte(fmt.Sprintf(";Layer height: %s", sliLayerHeight)),
		[]byte(";header_type: 3dp"),
		[]byte(";"), // slot for thumbnail
		[]byte(fmt.Sprintf(";file_total_lines: %d", lineCount)),
		[]byte(fmt.Sprintf(";estimated_time(s): %d", findEstimatedTime(gcodes))),
		[]byte(fmt.Sprintf(";nozzle_temperature(°C): %s", sliTemp)),
		[]byte(fmt.Sprintf(";build_plate_temperature(°C): %s", sliBedTemp)),
		[]byte(fmt.Sprintf(";work_speed(mm/minute): %d", speed*60)),
		[]byte(";Header End\n\n"),
	}

	thumbnail := convertThumbnail(gcodes)
	if thumbnail != nil {
		headers[4] = append([]byte(";thumbnail: "), thumbnail...)
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
	fmt.Println("# smfix a.gcode")
	fmt.Println("or")
	fmt.Println("# cat a.gcode | smfix > b.gcode")
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
