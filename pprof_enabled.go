//go:build pprof

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
)

func startCPUProfile() {
	cpuFile, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(cpuFile)
}

func stopCPUProfile() {
	pprof.StopCPUProfile()
}

func writeMemProfile() {
	memFile, err := os.Create("mem.pprof")
	if err != nil {
		panic(err)
	}
	defer memFile.Close()
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		panic(err)
	}
}
