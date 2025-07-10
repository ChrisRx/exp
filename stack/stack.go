// Package stack provides functions for getting caller information from the Go
// runtime.
package stack

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type Source struct {
	File     string
	Line     int
	FullName string
}

func (s Source) Name() string {
	return s.FullName[strings.LastIndex(s.FullName, "/")+1:]
}

func (s Source) String() string {
	name := s.Name()
	if name == "" {
		return fmt.Sprintf("%s:%d", s.File, s.Line)
	}
	return fmt.Sprintf("%s:%d %s", s.File, s.Line, name)
}

func GetSource(skip int) Source {
	pc, file, line, _ := runtime.Caller(1 + skip)
	s := Source{
		File: filepath.Base(file),
		Line: line,
	}
	if fn := runtime.FuncForPC(pc); fn != nil {
		s.FullName = fn.Name()
	}
	return s
}

const maxStackDepth = 10

func GetLocation(ignore func(Source) bool) string {
	for i := 1; i < maxStackDepth; i++ {
		s := GetSource(i + 1)
		if ignore(s) {
			continue
		}
		return s.String()
	}
	return "<unknown>"
}
