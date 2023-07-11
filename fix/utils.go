package fix

import (
	"bytes"
	"strconv"
	"strings"
)

func split(s string) []string {
	var x []string
	if strings.Contains(s, ";") {
		x = strings.Split(s, ";")
	} else {
		x = strings.Split(s, ",")
	}
	if len(x) == 1 {
		x = append(x, "")
	}
	for i, str := range x {
		x[i] = strings.TrimSpace(str)
	}
	return x
}

func splitFloat(s string) []float64 {
	var x []float64
	for _, v := range split(s) {
		f, _ := strconv.ParseFloat(v, 64)
		x = append(x, f)
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

func convertEstimatedTime(s string) int {
	// est := s[strings.Index(s, "= ")+2:] // 2d 12h 8m 58s
	est := strings.ReplaceAll(s, " ", "")
	t := map[byte]int{'d': 0, 'h': 0, 'm': 0, 's': 0}
	for _, p := range []byte("dhms") {
		if i := strings.IndexByte(est, p); i >= 0 {
			t[p], _ = strconv.Atoi(est[0:i])
			est = est[i+1:]
		}
	}
	return t['d']*86400 +
		t['h']*3600 +
		t['m']*60 +
		t['s']
}

func parseFloat(s string) float64 {
	var f float64
	f, _ = strconv.ParseFloat(s, 64)
	return f
}

func parseInt(s string) int {
	var i int
	i, _ = strconv.Atoi(s)
	return i
}

func getSetting(s string, key ...string) (v string, ok bool) {
	if s[0] == ';' {
		for _, p := range key {
			prefix := "; " + p + " ="
			if strings.HasPrefix(s, prefix) {
				return strings.TrimSpace(s[len(prefix):]), true
			}
		}
	}
	return "", false
}
