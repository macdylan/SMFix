# AGENTS.md

This file provides guidance to the AI agent when working with code in this repository.

## Project

SMFix is a G-code post-processor for Snapmaker 2/Artisan/J1/U1 printers. It reads sliced G-code from PrusaSlicer/SuperSlicer/OrcaSlicer, injects metadata headers, converts thumbnails, and optimizes multi-extruder behavior (preheat, shutoff, tool number remapping, Orca tool-unload cleanup).

**Performance is a project goal.** Input files range from hundreds of KB to **more than 1 GB**; the tool must handle them with bounded time and memory. Keep per-line allocations low on hot paths (parsing, `ParseParams`, modifiers) and verify changes with the benchmarks (see below). Output defaults to overwriting the input file in place.

## Build & Test

```bash
make test                   # unit tests: root module + fix/ (separate module; ./... from root never includes it)
go test -race ./...         # race detector (modifiers run in parallel); run the same inside fix/
make benchmark              # benchmarks for both modules
make darwin-arm64           # single-platform build → dist/smfix-darwin-arm64
make all                    # cross-compile all 7 targets
make all-zip                # build + zip for release
make pprof                  # build with profiling (uses -tags pprof)
```

## Architecture

- **Two modules**: root (`main`) and `fix/` (core library, its own `go.mod`). A local `go.work` exists for development but is **gitignored** — repo builds resolve `fix/` through the `replace github.com/macdylan/SMFix/fix => ./fix` directive in the root `go.mod`. Do not remove it: without it, fresh-clone/release builds silently compile a stale published `fix/`.
- All domain logic (G-code parsing, header extraction, modifiers) is in `fix/`. The root package is a thin CLI wrapper (`readGcodes` → `fixGcodes` → `writeOutput`; `process` chains them for in-memory use).
- **Pipeline order matters**: modifiers run *first*, then `ExtractHeader`/`ParseParams` — so `ParseParams` reads the *rewritten* slicer comments (e.g. after `GcodeReplaceToolNum` remaps multi-tool values down to 2). Changing that order breaks header generation.
- G-code modifiers (`GcodeFixShutoff`, `GcodeFixPreheat`, `GcodeReplaceToolNum`) are chained as `GcodeModifier` functions (`func([]*GcodeBlock) []*GcodeBlock`). `GcodeFixOrcaToolUnload` is always appended and not flag-controlled. `noTrim`/`noReinforceTower` flags exist but their modifiers are commented out (dead flags).
- `-nopreheat` defaults to `true`: preheat is off unless slicer preheat is absent (slicers ≥ PrusaSlicer 2.8 / Orca 2.1.1 implement it natively).
- **U1 passthrough** (`fix.IsU1Model`, raw scan in `smfix.go`): files whose `; printer_model` contains `U1` skip the whole pipeline — only the processed-mark is prepended and the input is copied **byte-for-byte** (no parsing, so spacing/blank lines/`G4 S0` survive). The U1 is a Klipper-based 4-toolhead machine: `%2` tool folding would corrupt its start gcode (`M104 S0 T2 A0`, `M106 P2`), the shutoff modifier could cancel required `M109 T<n>` waits inside toolchange macros, and the U1 firmware reads the slicer's native metadata instead of the SM2 header. Verified against real Snapmaker Orca and OrcaSlicer output (`fix/testdata/u1_*.gcode`).
- `slicerParams` is a global singleton (`fix.Params`) populated by `ParseParams` from G-code comment headers.
- Conditional compilation: `pprof_enabled.go` / `pprof_disabled.go` controlled by build tag `pprof`.
- Tests live in both modules: `fix/internal_test.go` + `fix/regression_test.go` (regressions for fixed bugs), `smfix_test.go` (CLI pipeline, first root-package tests).

## Performance baseline

Apple M1, synthetic 20k-line dual-extruder fixture (`fix/bench_test.go`), 2026-09:

| Benchmark | time | bytes | allocs |
|---|---|---|---|
| `BenchmarkParseParams` | 98.7 µs | 2.2 KB | 35 |
| `BenchmarkGcodeFixShutoffB` | 1.48 ms | 470 KB | 28,009 |
| `BenchmarkGcodeReplaceToolNumB` | 0.76 ms | 228 KB | 22,055 |
| `BenchmarkProcess` (root, ~1 MB in) | 7.95 ms (~60 MB/s) | 8.1 MB | 225,769 |

`ParseParams` used to allocate a string per gcode line (~5.6 allocs/line, 44× slower); it now inspects `Comment()`/cmd directly. Remaining hot spots are per-line `ParseGcodeBlock` and `GcodeBlock.String()` in the shutoff/replace modifiers and output writer — optimize before scaling beyond ~60 MB/s. Lines up to 1 MB are supported (Scanner buffer raised from the 64 KB default).

## CI

GitHub Actions triggers on tag push (`v*.*`): runs `make test`, then `make all-zip`, and creates a draft release. CI toolchain: Go 1.25.x (the `go` directive in go.mod stays 1.20).
