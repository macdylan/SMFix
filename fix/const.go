package fix

import "regexp"

const (
	Mark = "; Postprocessed by smfix (https://github.com/macdylan/SMFix)"

	PrintModeDefault     = "Default"
	PrintModeBackup      = "IDEX Backup"
	PrintModeDuplication = "IDEX Duplication"
	PrintModeMirror      = "IDEX Mirror"

	ToolheadSingle = "singleExtruderToolheadForSM2"
	ToolheadDual   = "dualExtruderToolheadForSM2"

	ModelA150 = "Snapmaker 2.0 A150"
	ModelA250 = "Snapmaker 2.0 A250"
	ModelA350 = "Snapmaker 2.0 A350"
	ModelA400 = "A400"
	ModelJ1   = "Snapmaker J1"

	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
)

var (
	reThumb = regexp.MustCompile(`(?m)(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)`)
)
