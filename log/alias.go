package log

import (
	"fmt"
	"log/slog"
	"os"
)

func Fatal(v ...any) {
	slog.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	slog.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...any) {
	slog.Error(fmt.Sprintln(v...))
	os.Exit(1)
}

func Panic(v ...any) {
	s := fmt.Sprint(v...)
	slog.Error(s)
	panic(s)
}

func Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	slog.Error(s)
	panic(s)
}

func Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	slog.Error(s)
	panic(s)
}

func Print(v ...any) {
	slog.Info(fmt.Sprint(v...))
}

func Printf(format string, v ...any) {
	slog.Info(fmt.Sprintf(format, v...))
}

func Println(v ...any) {
	slog.Info(fmt.Sprintln(v...))
}
