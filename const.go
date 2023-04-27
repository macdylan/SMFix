package main

import "regexp"

const (
	PrintModeDefault     = "Default"
	PrintModeBackup      = "IDEX Backup"
	PrintModeDuplication = "IDEX Duplication"
	PrintModeMirror      = "IDEX Mirror"
)

var (
	reThumb = regexp.MustCompile(`(?m)(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)`)
)
