package fix

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrEmptyString = errors.New("empty string")
)

const (
	// GCODE_SEPARATOR is a separator used to separate the sections of the block when is exported as line string
	GCODE_SEPARATOR = " "
)

type Gcode struct {
	word byte
	addr string
}

func (g *Gcode) Word() byte {
	return g.word
}

func (g *Gcode) HasAddr() bool {
	return len(g.addr) > 0
}

func (g *Gcode) Addr() string {
	return g.addr
}

func (g *Gcode) AddrAs(target any) error {
	baddr := []byte(g.addr)
	switch typ := target.(type) {
	case *string:
		*typ = g.addr
	case *int:
		i64, err := ParseInt(baddr)
		if err == nil {
			*typ = int(i64)
		}
		return err
	case *int32:
		i64, err := ParseInt(baddr)
		if err == nil {
			*typ = int32(i64)
		}
		return err
	case *float32:
		f64, err := strconv.ParseFloat(g.addr, 32)
		if err == nil {
			*typ = float32(f64)
		}
		return err
	default:
		return fmt.Errorf("unsupported addr type as %T", typ)
	}
	return nil
}

func (g *Gcode) SetAddr(value any) error {
	switch typ := value.(type) {
	case string:
		g.addr = strings.TrimSpace(typ)
	case int, int32, int64, uint, uint32, uint64:
		g.addr = fmt.Sprintf("%d", typ)
	case float32, float64:
		var (
			v64 float64
			ok  bool
		)
		if v64, ok = value.(float64); !ok {
			v32, _ := value.(float32)
			v64 = float64(v32)
		}

		switch g.Word() {
		case 'E':
			switch {
			case math.Abs(v64) < 0.00001:
				g.addr = "0.00000"
			default:
				g.addr = fmt.Sprintf("%.5f", v64)
			}
		default:
			switch {
			case math.Abs(v64) < 0.001:
				g.addr = "0.000"
			default:
				g.addr = fmt.Sprintf("%.3f", v64)
			}
		}
	case nil:
		g.addr = ""
	default:
		return fmt.Errorf("unsupported addr type %T", typ)
	}
	return nil
}

// func (g *Gcode) Compare(other Gcode) bool {
// 	return g.word == other.word && g.addr == other.addr
// }

func (g *Gcode) Is(s string) bool {
	return g.word == s[0] && g.addr == s[1:]
}

func (g *Gcode) String() string {
	return string(append([]byte{g.word}, g.addr[:]...))
}

func (g *Gcode) Copy() *Gcode {
	return &Gcode{
		word: g.word,
		addr: g.addr,
	}
}

func NewGcode(word byte, addr string) (*Gcode, error) {
	if err := isValidWord(word); err != nil {
		return nil, err
	}
	return &Gcode{word: word, addr: addr}, nil
}

func ParseGcode(s string) (*Gcode, error) {
	if s == "" {
		return nil, ErrEmptyString
	}
	return NewGcode(s[0], s[1:])
}

// GcodeBlock
type GcodeBlock struct {
	cmd     *Gcode
	params  []*Gcode
	comment string
	next    *GcodeBlock
}

func (b *GcodeBlock) Cmd() *Gcode {
	if b.cmd == nil {
		return &Gcode{}
	}
	return b.cmd
}

func (b *GcodeBlock) Params() []*Gcode {
	if b.params == nil {
		return []*Gcode{}
	}
	return b.params
}

func (b *GcodeBlock) Comment() string {
	return b.comment
}

func (b *GcodeBlock) SetComment(comment string, args ...any) {
	if len(args) > 0 {
		b.comment = fmt.Sprintf(comment, args...)
	} else {
		b.comment = comment
	}
}

func (b *GcodeBlock) AppendComment(comment string, args ...any) {
	if b.comment == "" {
		b.comment = ";"
	}
	if len(args) > 0 {
		b.comment += fmt.Sprintf(comment, args...)
	} else {
		b.comment += comment
	}
}

func (b *GcodeBlock) String() string {
	return strings.TrimSpace(b.Format("%c %p %m"))
}

func (b *GcodeBlock) IsComment() bool {
	return b.cmd == nil && len(b.params) == 0 && b.comment != ""
}

func (b *GcodeBlock) InComment(s string) bool {
	return strings.Contains(b.comment, s)
}

func (b *GcodeBlock) Is(s string) bool {
	return b.Cmd().Is(s)
}

func (b *GcodeBlock) HasParam(p byte) bool {
	for _, g := range b.Params() {
		if g.Word() == p {
			return true
		}
	}
	return false
}

func (b *GcodeBlock) RemoveParam(p byte) {
	for i, g := range b.Params() {
		if g.Word() == p {
			b.params = append(b.params[:i], b.params[i+1:]...)
		}
	}
}

func (b *GcodeBlock) SetParam(p byte, v string) error {
	for _, g := range b.Params() {
		if g.Word() == p {
			return g.SetAddr(v)
		}
	}
	new, err := NewGcode(p, v)
	if err == nil {
		b.params = append(b.params, new)
	}
	return err
}

func (b *GcodeBlock) GetParam(p byte, target any) error {
	for _, g := range b.Params() {
		if g.Word() == p {
			return g.AddrAs(target)
		}
	}
	return fmt.Errorf("param %s not found", string(p))
}

func (b *GcodeBlock) GetToolNum() (t int32, err error) {
	t = -1
	switch b.Cmd().Word() {
	case 'T': // Tn
		err = b.Cmd().AddrAs(&t)
	case 'M':
		switch b.Cmd().Addr() {
		case "106", "107":
			err = b.GetParam('P', &t)
		case "301", "303":
			err = b.GetParam('E', &t)
		default:
			err = b.GetParam('T', &t)
		}
	default:
		err = fmt.Errorf("command %s not supported", b.Cmd())
	}
	if (t == -1 || err != nil) && len(b.Comment()) > 2 {
		// try T in comment
		if ele := strings.TrimSpace(take(b.Comment(), `\s*T\d+`).taken); ele != "" {
			var i64 int64
			if i64, err = strconv.ParseInt(ele[1:], 10, 32); err == nil {
				t = int32(i64)
			}
			return
		}
	}
	return t, err
}

/*
Format formats the command with the given format string.

%c : command
%p : series of params
%m : comments
*/
func (b *GcodeBlock) Format(format string) string {
	result := strings.Builder{}
	result.Grow(128)

	for i := 0; i < len(format); i++ {
		if format[i] == '%' {
			if i+1 < len(format) {
				switch format[i+1] {
				case 'c':
					if b.cmd != nil {
						result.WriteString(b.Cmd().String())
					}
					i++
				case 'p':
					if total := len(b.Params()); total > 0 {
						for i, g := range b.Params() {
							result.WriteString(g.String())
							if i < total-1 {
								result.WriteString(GCODE_SEPARATOR)
							}
						}
					}
					i++
				case 'm':
					result.WriteString(b.Comment())
					i++
				}
			}
		} else {
			result.WriteByte(format[i])
		}
	}

	return result.String()
}

func (b *GcodeBlock) Copy() *GcodeBlock {
	params := make([]*Gcode, len(b.Params()))
	copy(params, b.Params())
	return &GcodeBlock{
		cmd:     b.Cmd().Copy(),
		params:  params,
		comment: b.Comment(),
	}
}

func ParseGcodeBlock(source string) (*GcodeBlock, error) {
	if len(source) > 0 && source[0] == ' ' {
		source = strings.TrimSpace(source)
	}

	if source == "" {
		return nil, ErrEmptyString
	}

	block := &GcodeBlock{}

	// keep comments
	if i := strings.Index(source, ";"); i != -1 {
		comments := source[i:]
		source = source[:i]
		block.SetComment(strings.TrimSpace(comments))
	}

	parse := prepareGcodeLineToParse(source)

	if parse == "" {
		return block, nil // only comments
	}

	params := make([]*Gcode, 0, 8)

	total := len(parse)
	for i := 0; i < total; {
		start := i
		for i < total && parse[i] != ' ' {
			i++
		}
		g := parse[start:i]

		if g == "" {
			continue
		}
		if err := isValidWord(g[0]); err == nil {
			gcode, err := ParseGcode(g)
			if err != nil {
				return nil, err
			}

			params = append(params, gcode)
		}

		for i < total && parse[i] == ' ' {
			i++
		}
	}

	if len(params) > 0 {
		block.cmd = params[0]
		block.params = params[1:]
	}

	return block, nil
}

// } GcodeBlock

// isValidWord allow knowledge if a potential word value contains a value valid according to a specification gcode.
func isValidWord(word byte) error {
	if word >= 'A' && word <= 'Z' {
		return nil
	}
	/*
		switch word {
		case 'G', 'T', 'S', 'P', 'X', 'Y', 'Z', 'U', 'V', 'W', 'I', 'J', 'D', 'H', 'F', 'R', 'Q', 'E', 'N',
			'B', //M260
			'K', //M900
			'C', //M2000
			'L', //M2020
			'M':
			return nil
		}
	*/

	return fmt.Errorf("gcode's word has invalid value: %v", word)
}

func insertAfter(gcodes *[]*GcodeBlock, pos int, g *GcodeBlock) {
	*gcodes = append(*gcodes, nil)
	copy((*gcodes)[pos:], (*gcodes)[pos-1:])
	(*gcodes)[pos-1] = g
}

func insertBefore(gcodes *[]*GcodeBlock, pos int, g *GcodeBlock) {
	*gcodes = append(*gcodes, nil)
	copy((*gcodes)[pos:], (*gcodes)[pos-1:])
	(*gcodes)[pos] = g
}
