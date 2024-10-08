package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/macdylan/SMFix/fix"
)

var (
	OutputPath       string
	noTrim           bool
	noShutoff        bool
	noPreheat        bool
	noReinforceTower bool
	noReplaceTool    bool
)

func init() {
	flag.StringVar(&OutputPath, "o", "", "output path, default is input path")
	flag.BoolVar(&noTrim, "notrim", false, "do not trim spaces in the gcode")
	flag.BoolVar(&noShutoff, "noshutoff", false, "do not shutoff nozzles that are no longer in use")
	flag.BoolVar(&noPreheat, "nopreheat", true, "do not pre-heat nozzles")
	flag.BoolVar(&noReinforceTower, "noreinforcetower", true, "do not reinforce the prime tower")
	flag.BoolVar(&noReplaceTool, "noreplacetool", false, "do not replace the tool number")
	flag.Parse()
}

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	var (
		in  *os.File
		err error
	)
	if len(flag.Args()) > 0 {
		in, err = os.OpenFile(flag.Arg(0), os.O_RDONLY, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		defer in.Close()
	}

	if in == nil {
		flag_usage()
	}

	startCPUProfile()
	defer func() {
		writeMemProfile()
		stopCPUProfile()
	}()

	// read gcodes form file
	gcodes := []*fix.GcodeBlock{}
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		line := sc.Text()

		if strings.HasPrefix(line, "; Postprocessed by smfix") {
			log.Fatalln(fix.ErrIsFixed)
		}

		g, err := fix.ParseGcodeBlock(line)
		if err == nil {
			// ignore G4 S0
			if g.Is("G4") {
				var s int
				if err := g.GetParam('S', &s); err == nil && s == 0 {
					continue
				}
			}

			gcodes = append(gcodes, g)
			continue
		}
		if err != fix.ErrEmptyString {
			log.Fatalf("Parse gcode error: %s", err)
		}
	}
	in.Close()
	if err := sc.Err(); err != nil {
		log.Fatalf("Read input file error: %s", err)
	}

	// fix gcodes
	funcs := make([]fix.GcodeModifier, 0, 6)
	if !noTrim {
		// funcs = append(funcs, fix.GcodeTrimLines)
	}
	if !noShutoff {
		funcs = append(funcs, fix.GcodeFixShutoff)
	}
	if !noPreheat {
		funcs = append(funcs, fix.GcodeFixPreheat)
	}
	if !noReplaceTool {
		funcs = append(funcs, fix.GcodeReplaceToolNum)
	}
	if !noReinforceTower {
		funcs = append(funcs, fix.GcodeReinforceTower)
	}
	funcs = append(funcs, fix.GcodeFixOrcaToolUnload)

	for _, fn := range funcs {
		gcodes = fn(gcodes)
	}

	// extract headers
	var headers [][]byte
	if headers, err = fix.ExtractHeader(gcodes); err != nil {
		log.Fatalf("Parse params failed: %s", err)
	}

	// prepare for output file
	if len(OutputPath) == 0 {
		OutputPath = flag.Arg(0)
	}
	out, err := os.Create(OutputPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	bufWriter := bufio.NewWriterSize(out, 64*1024)
	defer bufWriter.Flush()

	// write headers
	if _, err := bufWriter.Write(bytes.Join(headers, []byte("\n"))); err != nil {
		log.Fatalln(err)
	}

	// write gcodes
	for _, gcode := range gcodes {
		if _, err := bufWriter.WriteString(gcode.String() + "\n"); err != nil {
			log.Fatalln(err)
		}
	}
}
