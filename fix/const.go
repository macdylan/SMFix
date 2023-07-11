package fix

import "regexp"

const (
	Mark = "; Postprocessed by smfix (https://github.com/macdylan/Snapmaker2Slic3rPostProcessor)"

	PrintModeDefault     = "Default"
	PrintModeBackup      = "IDEX Backup"
	PrintModeDuplication = "IDEX Duplication"
	PrintModeMirror      = "IDEX Mirror"

	ToolheadSingle = "singleExtruderToolheadForSM2"
	ToolheadDual   = "dualExtruderToolheadForSM2"

	ModelA150 = "A150"
	ModelA250 = "A250"
	ModelA350 = "A350"
	ModelA400 = "A400"
	ModelJ1   = "Snapmaker J1"
)

var (
	reThumb = regexp.MustCompile(`(?m)(?:^; thumbnail begin \d+[x ]\d+ \d+)(?:\n|\r\n?)((?:.+(?:\n|\r\n?))+?)(?:^; thumbnail end)`)
)
