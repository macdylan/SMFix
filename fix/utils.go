package fix

import (
	"bytes"
	"errors"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrValueSyntax  = errors.New("invalid syntax")
	ErrIntegerRange = errors.New("value out of range")
)

func split(s string) []string {
	delimiter := ","
	if strings.Contains(s, ";") {
		delimiter = ";"
	}
	x := strings.Split(s, delimiter)
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

func ParseInt(b []byte) (int64, error) {
	if v, ok, overflow := _parseInt(b); !ok {
		if overflow {
			return 0, ErrIntegerRange
		}
		return 0, ErrValueSyntax
	} else {
		return v, nil
	}
}

func getSetting(s string, key ...string) (v string, ok bool) {
	strlen := len(s)
	if strlen > 5 && s[0] == ';' {
		for _, p := range key {
			if strlen < len(p)+4 {
				continue
			}
			prefix := "; " + p + " ="
			if strings.HasPrefix(s, prefix) {
				if v := strings.TrimSpace(s[len(prefix):]); v != "" {
					return v, true
				}
			}
		}
	}
	return "", false
}

func GoInParallelAndWait(work func(wi, wn int)) {
	var wg sync.WaitGroup
	wn := runtime.NumCPU()
	for wi := 0; wi < wn; wi++ {
		wg.Add(1)
		go func(wi, wn int) {
			work(wi, wn)
			wg.Done()
		}(wi, wn)
	}
	wg.Wait()
}

// removeDuplicateSpaces removes all consecutive spaces in a string
func removeDuplicateSpaces(s string) string {
	var (
		sb        strings.Builder
		prevSpace = false
	)

	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			if !prevSpace {
				sb.WriteByte(s[i])
				prevSpace = true
			}
		} else {
			sb.WriteByte(s[i])
			prevSpace = false
		}
	}

	return sb.String()
}

// removeSpecialChars removes only the escape characters \n, \t, and \r from the given string
func removeSpecialChars(s string) string {
	var result strings.Builder
	for _, c := range s {
		if c != '\n' && c != '\t' && c != '\r' {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// prepareGcodeLineToParse modify a string to can be parsed for the Parse function
// It doesn't verify if s strings is a gcode line valid
func prepareGcodeLineToParse(s string) string {
	s = strings.TrimSpace(s)
	s = removeSpecialChars(s)
	s = removeDuplicateSpaces(s)

	return s
}

type elementTaken struct {
	taken     string
	remainder string
}

var takeRegexp sync.Map

func take(source string, regex string) elementTaken {
	var (
		re *regexp.Regexp
	)
	if load, ok := takeRegexp.Load(regex); ok {
		re = load.(*regexp.Regexp)
	} else {
		re = regexp.MustCompile(regex)
		takeRegexp.Store(regex, re)
	}
	match := re.FindStringIndex(source)
	if match == nil {
		return elementTaken{remainder: source}
	}

	return elementTaken{taken: source[match[0]:match[1]], remainder: source[:match[0]] + source[match[1]:]}
}

// About 2x faster then strconv.ParseInt because it only supports base 10
func _parseInt(bytes []byte) (v int64, ok bool, overflow bool) {
	if len(bytes) == 0 {
		return 0, false, false
	}

	var neg bool = false
	if bytes[0] == '-' {
		neg = true
		bytes = bytes[1:]
	}

	var n uint64 = 0
	for _, c := range bytes {
		if c < '0' || c > '9' {
			return 0, false, false
		}
		if n > maxUint64/10 {
			return 0, false, true
		}
		n *= 10
		n1 := n + uint64(c-'0')
		if n1 < n {
			return 0, false, true
		}
		n = n1
	}

	if n > maxInt64 {
		if neg && n == absMinInt64 {
			return -absMinInt64, true, false
		}
		return 0, false, true
	}

	if neg {
		return -int64(n), true, false
	} else {
		return int64(n), true, false
	}
}
