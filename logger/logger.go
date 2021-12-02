package logger

import (
	"fmt"

	"github.com/fatih/color"
)

// LogInfo prints a message with the prefix "INFO: " in cyan.
func LogInfo(message string, a ...interface{}) {
	fmt.Println(color.CyanString("INFO: "), fmt.Sprintf(message, a...))
}

// LogWarn prints a message with the prefix "WARN: " in yellow.
func LogWarn(message string, a ...interface{}) {
	fmt.Println(color.YellowString("WARN: "), fmt.Sprintf(message, a...))
}

// LogError prints a message with the prefix "ERROR: " in red.
func LogError(message string, a ...interface{}) {
	fmt.Println(color.RedString("ERROR: "), fmt.Sprintf(message, a...))
}
