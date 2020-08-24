package logger

import (
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/fatih/color"
)

// log prints a message with a coloured header.
func log(logLevel uint8, colorFn func(format string, a ...interface{}) string, header string, format string, a ...interface{}) {
	if logLevel <= config.LogLevel {
		fmt.Printf(colorFn(header)+format, a...)
	}
}

// Info prints a message with the header "INFO: " in cyan.
func Info(format string, a ...interface{}) { log(2, color.CyanString, "INFO: ", format, a...) }

// Warn prints a message with the header "WARN: " in yellow.
func Warn(format string, a ...interface{}) { log(2, color.YellowString, "WARN: ", format, a...) }

// Error prints a message with the header "ERROR: " in red.
func Error(format string, a ...interface{}) { log(1, color.RedString, "ERROR: ", format, a...) }

// Panic automatically exits the program giving a traceback.
func Panic(err error) { panic(err) }
