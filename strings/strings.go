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
	var indent, line []byte
	var b bytes.Buffer
	for i := 0; scanner.Scan(); i++ {
		line = scanner.Bytes()

		if i == 0 {
			// The first line is blank so do not include it in the output.
			if len(line) == 0 {
				continue
			}
			if v := detect(line); v.Len() > 0 {
				indent = v.Bytes()
			}
		} else {
			// Only set the indent if the line is non-empty and the indent hasn't
			// already been set.
			if len(line) != 0 && indent == nil {
				indent = detect(line).Bytes()
			}
		}

		line = bytes.TrimPrefix(line, indent)
		b.Write(line)
		b.WriteRune('\n')
	}
	out := b.Bytes()

	// Don't include the last line if blank.
	if len(bytes.TrimSpace(line)) == 0 {
		out = bytes.TrimSuffix(out, append(line, '\n'))
	}
	return string(out)
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

var transliterations = map[rune]string{
	'&': "and",
	'à': "a",
	'é': "e",
}

func Slug(s string) string {
	var sb strings.Builder
	for _, b := range s {
		switch {
		case ('a' <= b && b <= 'z') || ('0' <= b && b <= '9'):
			sb.WriteRune(b)
		case ('A' <= b && b <= 'Z'):
			sb.WriteRune(b + 'a' - 'A')
		default:
			if b, ok := transliterations[b]; ok {
				sb.WriteString(b)
				continue
			}
			sb.WriteRune('-')
		}
	}
	return strings.Trim(sb.String(), "-")
}

func ToCamelCase(s string) string {
	var b strings.Builder
	for i, v := range strings.TrimSpace(s) {
		if !unicode.IsLetter(v) && !unicode.IsDigit(v) {
			continue
		}
		switch {
		case i == 0 || s[i-1] == '_':
			b.WriteRune(unicode.ToUpper(v))
		default:
			b.WriteRune(unicode.ToLower(v))
		}
	}
	return b.String()
}
