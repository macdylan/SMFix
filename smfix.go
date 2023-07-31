package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"

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

	if OutputPath == "" {
		OutputPath = in.Name()
	}

	var headers [][]byte
	if headers, err = fix.ExtractHeader(in); err != nil {
		log.Fatalf("Parse params failed: %s", err)
	}

	tmpfile, err := os.CreateTemp("", "smfix")
	if err != nil {
		log.Fatalf("Can not create temp file: %s", err)
	}

	gcodes := make([]string, 0, fix.Params.TotalLines+len(headers)+128)
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		gcodes = append(gcodes, sc.Text())
	}
	in.Close()
	if err := sc.Err(); err != nil {
		log.Fatalf("Read input file error: %s", err)
	}

	// fix gcodes
	if !noTrim {
		gcodes = fix.GcodeTrimLines(gcodes)
	}
	if !noShutoff {
		gcodes = fix.GcodeFixShutoff(gcodes)
	}
	if !noPreheat {
		gcodes = fix.GcodeFixPreheat(gcodes)
	}

	// write headers
	if _, err := tmpfile.Write(bytes.Join(headers, []byte("\n"))); err != nil {
		log.Fatalln(err)
	}

	// write gcodes
	for _, gcode := range gcodes {
		if _, err := tmpfile.WriteString(gcode + "\n"); err != nil {
			log.Fatalln(err)
		}
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatalf("Temp file error: %s", err)
	}

	// finally, move tmpfile to in
	if err := os.Rename(tmpfile.Name(), OutputPath); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
