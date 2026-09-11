package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	Version = "dev"
)

const usageTemplate = `smfix %s - Optimize G-code for Snapmaker 2.0 / Artisan / J1 / U1 printers
https://github.com/macdylan/SMFix

Usage:
  smfix [flags] <input.gcode>

The input file is overwritten in place unless -o is given. Files that
were already processed by smfix are rejected. Only the first file
argument is processed.

What it does:
  - injects the Snapmaker metadata header (machine, toolhead, temperatures,
    retraction, thumbnail, ...)
  - shuts off (M104 S0) extruders no longer in use and cancels stale
    reheat commands (unless -noshutoff)
  - remaps tool numbers T2+ down to T0/T1 so multi-spool slicer profiles
    work on 2-nozzle printers (unless -noreplacetool)
  - pre-heats the idle extruder before tool changes when the slicer does
    not implement it natively (see -nopreheat)
  - removes OrcaSlicer tool-unload temperature commands without a tool number

Snapmaker U1 files (4 toolheads) are passed through unchanged: the U1
firmware reads the slicer's native metadata and tool numbers T0-T3 are
already correct.

Supported slicers: PrusaSlicer, SuperSlicer, OrcaSlicer / Bambu Studio.

PrusaSlicer integration:
  Print Settings -> Output options -> Post-processing scripts:
    %s
  DO NOT include spaces in the path.

Flags:
`

// usageText builds the full help output; kept pure for testing.
func usageText(version, binPath string) string {
	return fmt.Sprintf(usageTemplate, version, binPath)
}

func printUsage() {
	ex, _ := os.Executable()
	absPath, _ := filepath.Abs(ex)
	flag.CommandLine.SetOutput(os.Stdout)
	fmt.Print(usageText(Version, absPath))
	flag.PrintDefaults()
}

func flag_usage() {
	printUsage()
	os.Exit(1)
}
