package jiji

import (
	"log"
	"os"
)

// Compatible with standard package "log"
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

// May be replaced by another logger
var Logger StdLogger

var Verbose = false

func init() {
	Logger = log.New(os.Stderr, "jiji:", log.LstdFlags)
}

func debug(args ...interface{}) {
	if Verbose {
		Logger.Print(args...)
	}
}

func debugf(s string, args ...interface{}) {
	if Verbose {
		Logger.Printf(s, args...)
	}
}

func debugln(args ...interface{}) {
	if Verbose {
		Logger.Println(args...)
	}
}
