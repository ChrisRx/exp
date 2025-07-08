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
	File string
	Line int
	Name string
}

func (s Source) String() string {
	if s.Name == "" {
		return fmt.Sprintf("%s:%d", s.File, s.Line)
	}
	return fmt.Sprintf("%s:%d %s", s.File, s.Line, s.Name)
}

func GetSource(skip int) Source {
	pc, file, line, _ := runtime.Caller(1 + skip)
	s := Source{
		File: filepath.Base(file),
		Line: line,
	}
	if fn := runtime.FuncForPC(pc); fn != nil {
		name := fn.Name()
		s.Name = name[strings.LastIndex(name, "/")+1:]
	}
	return s
}
