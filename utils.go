package main

import (
	"bytes"
	"os"
	"strconv"
	"strings"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func config(gcodes [][]byte, key string) string {
	// from env
	if v, ok := os.LookupEnv("SLIC3R_" + strings.ToUpper(key)); ok {
		return v
	}

	// from prusaslicer_config
	key_b := []byte("; " + key + " =")
	i := len(gcodes) - 1
	j := max(i-1000, 0) // tail 1k lines
	for ; i >= j; i-- {
		if 0 == bytes.Index(gcodes[i], key_b) {
			return string(gcodes[i][bytes.Index(gcodes[i], []byte("= "))+2:])
		}
	}

	return ""
}

func getProperty(gcodes [][]byte, keys ...string) (v string) {
	for _, key := range keys {
		v = config(gcodes, key)
		if v != "" {
			return v
		}
	}
	return v
}

func split(s string) []string {
	var x []string
	if strings.Contains(s, ";") {
		x = strings.Split(s, ";")
	} else {
		x = strings.Split(s, ",")
	}
	if len(x) == 1 {
		x = append(x, "0")
	}
	return x
}

func convertThumbnail(gcodes [][]byte) []byte {
	comments := bytes.NewBuffer([]byte{})
	for _, line := range gcodes {
		if len(line) > 0 && line[0] == ';' {
			comments.Write(line)
			comments.WriteRune('\n')
		}
	}
	matches := reThumb.FindAllSubmatch(comments.Bytes(), -1)
	if matches != nil {
		none := []byte(nil)
		data := matches[len(matches)-1][1]
		data = bytes.ReplaceAll(data, []byte("\r\n"), none)
		data = bytes.ReplaceAll(data, []byte("\n"), none)
		data = bytes.ReplaceAll(data, []byte("; "), none)
		b := []byte("data:image/png;base64,")
		return append(b, data...)
	}
	return nil
}

func findEstimatedTime(gcodes [][]byte) int {
	for _, line := range gcodes {
		if 0 == bytes.Index(line, []byte("; estimated printing time")) {
			est := line[bytes.Index(line, []byte("= "))+2:] // 2d 12h 8m 58s
			est = bytes.ReplaceAll(est, []byte(" "), []byte(nil))
			t := map[byte]int{'d': 0, 'h': 0, 'm': 0, 's': 0}
			for _, p := range []byte("dhms") {
				if i := bytes.IndexByte(est, p); i >= 0 {
					t[p], _ = strconv.Atoi(string(est[0:i]))
					est = est[i+1:]
				}
			}
			return t['d']*86400 +
				t['h']*3600 +
				t['m']*60 +
				t['s']
		}
	}
	return 0
}

func startWith(b []byte, prefix ...string) bool {
	for _, p := range prefix {
		if bytes.HasPrefix(b, []byte(p)) {
			return true
		}
	}
	return false
}
