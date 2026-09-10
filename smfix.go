package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
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
	// flag.BoolVar(&noReinforceTower, "noreinforcetower", true, "do not reinforce the prime tower")
	flag.BoolVar(&noReplaceTool, "noreplacetool", false, "do not replace the tool number")
}

// maxScanTokenSize allows metadata/comment lines up to 1 MB; the bufio
// default (64 KB) aborts the whole scan on longer lines.
const maxScanTokenSize = 1024 * 1024

func readGcodes(r io.Reader) ([]*fix.GcodeBlock, error) {
	gcodes := []*fix.GcodeBlock{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)
	for sc.Scan() {
		line := sc.Text()

		if strings.HasPrefix(line, "; Postprocessed by smfix") {
			return nil, fix.ErrIsFixed
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
			return nil, fmt.Errorf("parse gcode error: %w", err)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read input file error: %w", err)
	}
	return gcodes, nil
}

// fixGcodes runs the enabled modifier chain over the parsed blocks.
func fixGcodes(gcodes []*fix.GcodeBlock) []*fix.GcodeBlock {
	funcs := make([]fix.GcodeModifier, 0, 4)
	if !noShutoff {
		funcs = append(funcs, fix.GcodeFixShutoff)
	}
	if !noPreheat {
		funcs = append(funcs, fix.GcodeFixPreheat)
	}
	if !noReplaceTool {
		funcs = append(funcs, fix.GcodeReplaceToolNum)
	}
	funcs = append(funcs, fix.GcodeFixOrcaToolUnload)

	for _, fn := range funcs {
		gcodes = fn(gcodes)
	}
	return gcodes
}

// writeOutput emits the Snapmaker header followed by all gcode blocks.
func writeOutput(out io.Writer, gcodes []*fix.GcodeBlock) error {
	// extract headers
	headers, err := fix.ExtractHeader(gcodes)
	if err != nil {
		return fmt.Errorf("parse params failed: %w", err)
	}

	bufWriter := bufio.NewWriterSize(out, 64*1024)

	// write headers
	if _, err := bufWriter.Write(bytes.Join(headers, []byte("\n"))); err != nil {
		return err
	}

	// write gcodes
	for _, gcode := range gcodes {
		if _, err := bufWriter.WriteString(gcode.String()); err != nil {
			return err
		}
		if err := bufWriter.WriteByte('\n'); err != nil {
			return err
		}
	}
	return bufWriter.Flush()
}

// process pipelines input gcode into fixed gcode with a Snapmaker header.
// in and out must not refer to the same file.
func process(in io.Reader, out io.Writer) error {
	gcodes, err := readGcodes(in)
	if err != nil {
		return err
	}
	gcodes = fixGcodes(gcodes)
	return writeOutput(out, gcodes)
}

func main() {
	flag.Parse()

	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	if len(flag.Args()) < 1 {
		flag_usage()
	}

	in, err := os.OpenFile(flag.Arg(0), os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	startCPUProfile()
	defer func() {
		writeMemProfile()
		stopCPUProfile()
	}()

	// read gcodes form file
	gcodes, err := readGcodes(in)
	if err != nil {
		log.Fatalln(err)
	}
	// close before creating the output file: it defaults to the input
	// path, and Windows refuses to truncate an open file.
	in.Close()

	gcodes = fixGcodes(gcodes)

	// prepare for output file
	if len(OutputPath) == 0 {
		OutputPath = flag.Arg(0)
	}
	out, err := os.Create(OutputPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	if err := writeOutput(out, gcodes); err != nil {
		log.Fatalln(err)
	}
}
