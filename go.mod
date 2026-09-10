module github.com/macdylan/SMFix

go 1.20

require github.com/macdylan/SMFix/fix v0.0.0-20240325141746-70877a3c65b4

// the workspace is gitignored: without this replace, release builds
// (fresh clone, no go.work) silently compile a stale published fix/.
replace github.com/macdylan/SMFix/fix => ./fix
