package log

import "log"

func Fatal(v ...any) {
	log.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

func Fatalln(v ...any) {
	log.Fatalln(v...)
}

func Panic(v ...any) {
	log.Panic(v...)
}

func Panicf(format string, v ...any) {
	log.Panicf(format, v...)
}

func Panicln(v ...any) {
	log.Panicln(v...)
}

func Print(v ...any) {
	log.Print(v...)
}

func Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func Println(v ...any) {
	log.Println(v...)
}
