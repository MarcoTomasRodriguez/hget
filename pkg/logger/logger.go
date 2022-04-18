package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var logger = log.New(os.Stdout, "", 0)

// Info prints a message with the prefix "INFO: " in cyan.
func Info(message string, a ...interface{}) {
	logger.Println(color.CyanString("INFO:"), fmt.Sprintf(message, a...))
}

// Warn prints a message with the prefix "WARN: " in yellow.
func Warn(message string, a ...interface{}) {
	logger.Println(color.YellowString("WARN:"), fmt.Sprintf(message, a...))
}

// Error prints a message with the prefix "ERROR: " in red.
func Error(message string, a ...interface{}) {
	logger.Println(color.RedString("ERROR:"), fmt.Sprintf(message, a...))
}
