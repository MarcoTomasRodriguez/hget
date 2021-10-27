package logger

import (
	"fmt"

	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/fatih/color"
)

type Level uint8

const (
	Info Level = iota
	Warn
	Error
)

func Log(level Level, message string) {
	if uint8(level) <= config.Config.LogLevel {
		switch level {
		case Info:
			fmt.Println(color.CyanString("INFO: "), message)
		case Warn:
			fmt.Println(color.YellowString("WARN: "), message)
		case Error:
			fmt.Println(color.RedString("ERROR: "), message)
		}
	}
}

func LogInfo(message string, a ...interface{}) {
	Log(Info, fmt.Sprintf(message, a...))
}

func LogWarn(message string, a ...interface{}) {
	Log(Warn, fmt.Sprintf(message, a...))
}

func LogError(message string, a ...interface{}) {
	Log(Error, fmt.Sprintf(message, a...))
}
