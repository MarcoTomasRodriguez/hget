package logger

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

// log prints a message with a coloured header.
func log(colorFn func(format string, a ...interface{}) string, header string, format string, a ...interface{}) {
	fmt.Printf(colorFn(header) + format, a...)
}

// Info prints a message with the header "INFO: " in cyan.
func Info(format string, a ...interface{}) { log(color.CyanString, "INFO: ", format, a...) }

// Warn prints a message with the header "WARN: " in yellow.
func Warn(format string, a ...interface{}) { log(color.YellowString, "WARN: ", format, a...) }

// Error prints a message with the header "ERROR: " in red.
func Error(format string, a ...interface{}) { log(color.RedString, "ERROR: ", format, a...) }

// Panic prints the error and exits with the code 1.
func Panic(err error) { Error("%v\n", err); os.Exit(1) }
