package jiji

import (
	"log"
	"os"
)

// Compatible with standard package "log"
type StdLogger interface {
	Print(...any)
	Printf(string, ...any)
	Println(...any)

	Fatal(...any)
	Fatalf(string, ...any)
	Fatalln(...any)

	Panic(...any)
	Panicf(string, ...any)
	Panicln(...any)
}

// May be replaced by another logger
var Logger StdLogger

var Verbose = false

func init() {
	Logger = log.New(os.Stderr, "jiji:", log.LstdFlags)
}

func debug(args ...any) {
	if Verbose {
		Logger.Print(args...)
	}
}

func debugf(s string, args ...any) {
	if Verbose {
		Logger.Printf(s, args...)
	}
}

func debugln(args ...any) {
	if Verbose {
		Logger.Println(args...)
	}
}
