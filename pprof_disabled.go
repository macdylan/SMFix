//go:build !pprof

package main

func startCPUProfile() {
	// No-op when pprof is disabled
}

func stopCPUProfile() {
	// No-op when pprof is disabled
}

func writeMemProfile() {
	// No-op when pprof is disabled
}
