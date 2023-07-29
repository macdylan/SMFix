package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/macdylan/SMFix/fix"
)

var (
	OutputPath string
	noTrim     bool
	noShutoff  bool
	noPreheat  bool
)

func init() {
	flag.StringVar(&OutputPath, "o", "", "output path, default is input path")
	flag.BoolVar(&noTrim, "notrim", false, "do not trim spaces in the gcode")
	flag.BoolVar(&noShutoff, "noshutoff", false, "do not shutoff nozzles that are no longer in use")
	flag.BoolVar(&noPreheat, "nopreheat", false, "do not pre-heat nozzles")
	flag.Parse()
}

func main() {
	var (
		in  *os.File
		err error
	)
	if len(flag.Args()) > 0 {
		in, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalln(err)
		}
	}

	if in == nil {
		flag_usage()
	}

	var headers [][]byte
	if headers, err = fix.ExtractHeader(in); err != nil {
		log.Fatalf("Parse params failed: %s", err)
	}

	// write headers
	tmpfile, err := os.CreateTemp("", "smfix")
	if err != nil {
		log.Fatalf("Can not create temp file: %s", err)
	}

	if _, err := tmpfile.Write(bytes.Join(headers, []byte("\n"))); err != nil {
		log.Fatalln(err)
	}

	// 4. append raw gcodes
	if _, err := io.Copy(tmpfile, in); err != nil {
		log.Fatalf("Can not write gcodes: %s", err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatalf("Temp file error: %s", err)
	}

	// 5. fix gcodes
	if !noTrim || !noPreheat || !noShutoff {
		gcodes := make([]string, 0, fix.Params.TotalLines+len(headers)+128)

		f, _ := os.Open(tmpfile.Name())
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			gcodes = append(gcodes, sc.Text())
		}
		if err := sc.Err(); err != nil {
			log.Fatalf("Read temp file error: %s", err)
		}

		if len(gcodes) > 30 {
			if !noTrim {
				gcodes = fix.GcodeTrimLines(gcodes)
			}
			if !noShutoff {
				gcodes = fix.GcodeFixShutoff(gcodes)
			}
			if !noPreheat {
				gcodes = fix.GcodeFixPreheat(gcodes)
			}
			os.WriteFile(tmpfile.Name(), []byte(strings.Join(gcodes, "\n")), 0644)
		}
	}

	// 6. finally, move tmpfile to in
	in.Close()
	if OutputPath == "" {
		OutputPath = in.Name()
	}
	if err := os.Rename(tmpfile.Name(), OutputPath); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
