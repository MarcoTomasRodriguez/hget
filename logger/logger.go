package logger

import (
	"fmt"

	"github.com/fatih/color"
)

func LogInfo(message string, a ...interface{}) {
	fmt.Println(color.CyanString("INFO: "), fmt.Sprintf(message, a...))
}

func LogWarn(message string, a ...interface{}) {
	fmt.Println(color.YellowString("WARN: "), fmt.Sprintf(message, a...))
}

func LogError(message string, a ...interface{}) {
	fmt.Println(color.RedString("ERROR: "), fmt.Sprintf(message, a...))
}
