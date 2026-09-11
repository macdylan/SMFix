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
	showHelp         bool
	showVersion      bool
)

func init() {
	flag.Usage = flag_usage
	flag.StringVar(&OutputPath, "o", "", "output path; by default the input file is overwritten in place")
	flag.BoolVar(&noTrim, "notrim", false, "reserved; line trimming is currently disabled (no-op)")
	flag.BoolVar(&noShutoff, "noshutoff", false, "keep extruders hot after their last use; by default idle nozzles are turned off with M104 S0")
	flag.BoolVar(&noPreheat, "nopreheat", true, "do not pre-heat the idle extruder before tool changes; slicers >= PrusaSlicer 2.8 / OrcaSlicer 2.1.1 handle preheat natively")
	flag.BoolVar(&noReplaceTool, "noreplacetool", false, "keep original tool numbers; by default T2+ are remapped to T0/T1 for 2-nozzle printers")
	flag.BoolVar(&showHelp, "h", false, "show this help and exit")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
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

// u1ModelLine reports whether a raw gcode line marks the file as a
// Snapmaker U1 slice (see IsU1Model for why U1 needs passthrough).
func u1ModelLine(line []byte) bool {
	return bytes.HasPrefix(line, []byte("; printer_model =")) && bytes.Contains(line, []byte("U1"))
}

var markLine = []byte("; Postprocessed by smfix")

// isU1Reader scans raw input lines (no parsing, no allocations) to
// decide between the U1 passthrough and the SM2 pipeline. It returns
// fix.ErrIsFixed when the file was already processed. r is consumed;
// the caller must rewind seekable readers afterwards.
func isU1Reader(r io.Reader) (u1 bool, err error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)
	for sc.Scan() {
		line := sc.Bytes()
		if bytes.HasPrefix(line, markLine) {
			return false, fix.ErrIsFixed
		}
		if u1ModelLine(line) {
			return true, nil
		}
	}
	return false, sc.Err()
}

// writeU1 emits the processed-mark followed by the input verbatim.
func writeU1(out io.Writer, in io.Reader) error {
	bufWriter := bufio.NewWriterSize(out, 64*1024)
	if _, err := bufWriter.WriteString(fix.Mark + "\n"); err != nil {
		return err
	}
	if _, err := io.Copy(bufWriter, in); err != nil {
		return err
	}
	return bufWriter.Flush()
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

// writeBlocks emits every gcode block on its own line.
func writeBlocks(bufWriter *bufio.Writer, gcodes []*fix.GcodeBlock) error {
	for _, gcode := range gcodes {
		if _, err := bufWriter.WriteString(gcode.String()); err != nil {
			return err
		}
		if err := bufWriter.WriteByte('\n'); err != nil {
			return err
		}
	}
	return nil
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
	if err := writeBlocks(bufWriter, gcodes); err != nil {
		return err
	}
	return bufWriter.Flush()
}

// seekableReader can be scanned and rewound (files, strings.Reader, ...).
type seekableReader interface {
	io.Reader
	io.Seeker
}

// process pipelines input gcode into fixed gcode with a Snapmaker
// header. U1 files bypass the pipeline entirely: the input is copied
// verbatim after the processed-mark (raw passthrough, no parsing).
// The input is read twice; non-seekable readers are buffered.
func process(in io.Reader, out io.Writer) error {
	if s, ok := in.(seekableReader); ok {
		u1, err := isU1Reader(s)
		if err != nil {
			return err
		}
		if _, err := s.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if u1 {
			return writeU1(out, s)
		}
		gcodes, err := readGcodes(s)
		if err != nil {
			return err
		}
		gcodes = fixGcodes(gcodes)
		return writeOutput(out, gcodes)
	}

	// non-seekable input: buffer once so it can be inspected and,
	// for U1, emitted verbatim
	buf, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	u1, err := isU1Reader(bytes.NewReader(buf))
	if err != nil {
		return err
	}
	if u1 {
		return writeU1(out, bytes.NewReader(buf))
	}
	gcodes, err := readGcodes(bytes.NewReader(buf))
	if err != nil {
		return err
	}
	gcodes = fixGcodes(gcodes)
	return writeOutput(out, gcodes)
}

func main() {
	flag.Parse()

	if showHelp {
		printUsage()
		return
	}
	if showVersion {
		fmt.Println("smfix", Version)
		return
	}

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

	// raw scan decides between U1 passthrough and the SM2 pipeline;
	// consumes the file, so seek back before reading again
	u1, err := isU1Reader(in)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		log.Fatalln(err)
	}

	// prepare for output file
	if len(OutputPath) == 0 {
		OutputPath = flag.Arg(0)
	}
	inPlace := OutputPath == flag.Arg(0)

	if u1 {
		// raw passthrough; in-place requires buffering because a
		// file cannot be prepended in place
		var src io.Reader = in
		if inPlace {
			buf, err := io.ReadAll(in)
			if err != nil {
				log.Fatalln(err)
			}
			src = bytes.NewReader(buf)
			// Windows refuses to truncate the file while it is open
			in.Close()
		}

		out, err := os.Create(OutputPath)
		if err != nil {
			log.Fatalln(err)
		}
		defer in.Close()
		defer out.Close()
		if err := writeU1(out, src); err != nil {
			log.Fatalln(err)
		}
		return
	}

	// read gcodes form file
	gcodes, err := readGcodes(in)
	if err != nil {
		log.Fatalln(err)
	}
	// close before creating the output file: it defaults to the input
	// path, and Windows refuses to truncate an open file.
	in.Close()

	gcodes = fixGcodes(gcodes)

	out, err := os.Create(OutputPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	if err := writeOutput(out, gcodes); err != nil {
		log.Fatalln(err)
	}
}
