//go:generate go tool aliaspkg -docs=all

package strings

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Dedent attempts to determine the indent-level from the first non-empty line,
// and trims the indents from all lines.
func Dedent(s string) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	var indent []byte
	var b bytes.Buffer
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if indent == nil {
			indent = detect(line).Bytes()
		}
		line = bytes.TrimPrefix(line, indent)
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		b.Write(line)
		b.WriteRune('\n')
	}
	return b.String()
}

func detect(line []byte) (indent indent) {
	if len(line) == 0 {
		return
	}
	switch rune(line[0]) {
	case '\t':
		indent.c = rune(line[0])
		indent.n += 1
	case ' ':
		indent.c = rune(line[0])
		indent.n += 1
	default:
		return
	}
	for _, r := range line[1:] {
		if rune(r) != indent.c {
			return
		}
		indent.n += 1
	}
	return
}

type indent struct {
	c rune
	n int
}

func (i indent) Len() int {
	return cmp.Or(i.n, 0)
}

func (i indent) Bytes() []byte {
	return bytes.Repeat(utf8.AppendRune(nil, i.c), i.n)
}

func ToSnakeCase(s string) string {
	input := []rune(s)
	isLower := func(idx int) bool {
		return idx < len(input)-1 && unicode.IsLower(input[idx])
	}
	var b strings.Builder
	for i, v := range input {
		if unicode.IsUpper(v) {
			if i > 0 && (isLower(i-1) || isLower(i+1)) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(v))
			continue
		}
		b.WriteRune(v)
	}
	return b.String()
}

func ToString[T fmt.Stringer](s T) string {
	return s.String()
}
