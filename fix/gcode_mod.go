package fix

import (
	"regexp"
	"strings"
)

var (
	// toolchange command
	reToolchange = regexp.MustCompile(`^\s*(?P<tool>T[0-9]+).*`)
	// M109 heat and wait
	reM109 = regexp.MustCompile(`^\s*M109 .*(?P<tool>T[\d]+).*`)
	// M104 cool and continue
	reM104 = regexp.MustCompile(`^\s*M104 .*(?P<tool>T[\d]+).*`)
	// M73 progress update command
	reM73 = regexp.MustCompile(`^\s*M73 .*(?P<remain>R[\d]+).*`)

	reTemp = regexp.MustCompile(`^\s*(?P<cmd>M10[49]) .*(?P<arg1>[ST][\d]+).*(?P<arg2>[ST][\d]+).*`)
)

func GcodeFixShutoff(gcodes []string) (output []string) {
	// track of the last line each tool was used
	toolLastLine := make(map[string]int)

	// look for tool changes
	for n, line := range gcodes {
		toolMatch := reToolchange.FindStringSubmatch(line)
		if len(toolMatch) > 0 {
			tool := toolMatch[1]
			toolLastLine[tool] = n
		}
	}

	// add M104 S0
	var curTool string
	for n, line := range gcodes {
		output = append(output, line)

		toolMatch := reToolchange.FindStringSubmatch(line)
		if len(toolMatch) > 0 {
			nextTool := toolMatch[1]

			if curTool != "" && curTool != nextTool {
				if toolLastLine[curTool] < n {
					output = append(output, "M104 S0 "+curTool+" ; (Fixed: Shutoff "+curTool+")")
				}
			}
			curTool = nextTool
		}
	}

	return output
}

var (
	PreheatShort = 1
	PreheatLong  = 3
)

func GcodeFixPreheat(gcodes []string) (output []string) {
	hasM73 := false // must enable "Supports remaining times" in the Printer settings
	for _, line := range gcodes {
		output = append(output, line)

		m109Match := reM109.FindStringSubmatch(line)
		if len(m109Match) > 0 {
			tool := m109Match[1]
			remaining := ""
			remainingDiffCount := 0
			nShortPreheat := -1
			nLongPreheat := -1

			// build preheat command
			preheat := strings.Replace(line, "M109", "M104", 1)
			if comment := strings.Index(preheat, ";"); comment == -1 { // no comment
				preheat = preheat + " ;"
			} else {
				cmd := preheat[:comment]
				// T is not in cmd but in the comment
				if !strings.Contains(cmd, tool) {
					preheat = strings.TrimRight(cmd, " ") + " " + tool + " " + preheat[comment:]
				}
			}

			// find line to place preheat command
			for n := len(output) - 2; n >= 0; n-- {
				checkLine := output[n]

				// find an earlier M109 for this tool, stop looking
				earlierM109 := reM109.FindStringSubmatch(checkLine)
				if len(earlierM109) > 0 && earlierM109[1] == tool {
					break
				}
				// find an earlier M104 for this tool, determine where to place pre-heat command
				m104Match := reM104.FindStringSubmatch(checkLine)
				if len(m104Match) > 0 && m104Match[1] == tool && hasM73 {
					if strings.Contains(checkLine, ";cooldown") || strings.Contains(checkLine, ";standby "+tool) {
						if remainingDiffCount < PreheatShort {
							output[n] = ";(Fixed: remove cooldown: " + checkLine + ")"
						} else if remainingDiffCount < PreheatLong {
							output = append(output[:nShortPreheat], append([]string{preheat + "(Fixed: pre-heat short)"}, output[nShortPreheat:]...)...)
						} else {
							output = append(output[:nLongPreheat], append([]string{preheat + "(Fixed: pre-heat long)"}, output[nLongPreheat:]...)...)
							output[n] = "M104 S110 " + tool + " ; (Fixed: deep freeze instead of: " + checkLine + ")"
						}
					}
					break
				}

				m73Match := reM73.FindStringSubmatch(checkLine)
				if len(m73Match) > 0 {
					hasM73 = true
					remain := m73Match[1]
					if remaining == "" { // first remaining
						remaining = remain
					} else if remaining != remain {
						remaining = remain
						remainingDiffCount++
						if remainingDiffCount == PreheatShort {
							nShortPreheat = n
						} else if remainingDiffCount == PreheatLong {
							nLongPreheat = n
						}
					}
				} // m73Match
			}
		} // m109Match
	}

	if !hasM73 {
		return output
	}

	// remove any unnecessary M104 and 109 commands
	curToolTemp := make(map[string]string)
	curToolTempGuaranteed := make(map[string]bool)
	for n, line := range output {
		tempMatch := reTemp.FindStringSubmatch(line)
		if len(tempMatch) > 0 {
			cmd := tempMatch[1]
			var temp, tool string
			if strings.HasPrefix(tempMatch[2], "S") && strings.HasPrefix(tempMatch[3], "T") {
				// M104 S255 T1
				temp = tempMatch[2]
				tool = tempMatch[3]
			} else {
				// M104 T1 S255
				temp = tempMatch[3]
				tool = tempMatch[2]

			}

			if curTemp, ok := curToolTemp[tool]; ok && curTemp == temp {
				if cmd == "M104" {
					output[n] = ";(Fixed: already requested temp: " + line + ")"
				} else if cmd == "M109" {
					if curToolTempGuaranteed[tool] {
						output[n] = ";(Fixed: already stabilized temp: " + line + ")"
					} else {
						curToolTempGuaranteed[tool] = true
					}
				}
			} else {
				curToolTemp[tool] = temp
				curToolTempGuaranteed[tool] = (cmd == "M109")
			}
		}
	}

	return output
}

func GcodeTrimLines(gcodes []string) (output []string) {
	blank := false
	for _, line := range gcodes {
		line = strings.TrimSpace(line)
		if line == "" {
			if !blank {
				output = append(output, line)
				blank = true
			}
		} else {
			output = append(output, line)
			blank = false
		}
	}
	return output
}
