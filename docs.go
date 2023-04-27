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

func flag_usage() {
	ex, _ := os.Executable()
	usage := `Optimize G-code file for Snapmaker printers.
%s - https://github.com/macdylan/Snapmaker2Slic3rPostProcessor

Example configuration in PrusaSlicer,
Go to Print Settings -> Output options -> Post-processing scripts:

  %s;

DO NOT include spaces in the path.

`
	absPath, _ := filepath.Abs(ex)
	fmt.Printf(usage, Version, absPath)
	flag.PrintDefaults()
	os.Exit(1)
}
