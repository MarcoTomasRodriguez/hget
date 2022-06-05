package logger

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/do"
	"log"
)

// Info prints a message with the prefix "INFO: " in cyan.
func Info(message string, a ...interface{}) {
	logger := do.MustInvoke[*log.Logger](do.DefaultInjector)
	logger.Println(color.CyanString("INFO:"), fmt.Sprintf(message, a...))
}

// Warn prints a message with the prefix "WARN: " in yellow.
func Warn(message string, a ...interface{}) {
	logger := do.MustInvoke[*log.Logger](do.DefaultInjector)
	logger.Println(color.YellowString("WARN:"), fmt.Sprintf(message, a...))
}

// Error prints a message with the prefix "ERROR: " in red.
func Error(message string, a ...interface{}) {
	logger := do.MustInvoke[*log.Logger](do.DefaultInjector)
	logger.Println(color.RedString("ERROR:"), fmt.Sprintf(message, a...))
}
