// Package stack provides functions for getting caller information from the Go
// runtime.
package stack

import (
	"fmt"
	"go/build"
	"reflect"
	"runtime"
	"strings"
)

const maxStackDepth = 10

type dummy struct{}

var packageName = reflect.TypeOf(dummy{}).PkgPath()

func Trace(skip int) (frames []Frame) {
	for i := 1; i < maxStackDepth; i++ {
		pc, file, line, ok := runtime.Caller(i + skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		name := fn.Name()
		file = strings.TrimPrefix(file, build.Default.GOROOT+"/src/")

		// Filter out all frames from this package.
		if strings.HasPrefix(name, packageName) {
			continue
		}
		GOROOT := build.Default.GOROOT
		if len(GOROOT) > 0 && strings.Contains(file, GOROOT) {
			continue
		}
		frames = append(frames, Frame{
			pc:   pc,
			file: file,
			line: line,
			name: name,
		})
	}
	return frames
}

type Frame struct {
	pc   uintptr
	file string
	line int
	name string
}

func (f Frame) String() string {
	s := fmt.Sprintf("%s:%d", f.file, f.line)
	if f.name == "" {
		return s
	}
	return s + fmt.Sprintf(" -- %s()", shortFuncName(f.name))
}

func (f Frame) Name() string {
	return f.name
}

func shortFuncName(name string) string {
	name = name[strings.LastIndex(name, "/")+1:]
	name = name[strings.Index(name, ".")+1:]
	name = strings.Replace(name, "(", "", 1)
	name = strings.Replace(name, "*", "", 1)
	name = strings.Replace(name, ")", "", 1)
	return name
}

func Location(ignore func(Frame) bool) string {
	for _, frame := range Trace(1) {
		if ignore(frame) {
			continue
		}
		return frame.String()
	}
	return "<unknown>"
}
