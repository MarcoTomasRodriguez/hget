package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

// Info prints a message with the prefix "INFO: " in cyan.
func (l *Logger) Info(message string, a ...interface{}) {
	l.logger.Println(color.CyanString("INFO:"), fmt.Sprintf(message, a...))
}

// Warn prints a message with the prefix "WARN: " in yellow.
func (l *Logger) Warn(message string, a ...interface{}) {
	l.logger.Println(color.YellowString("WARN:"), fmt.Sprintf(message, a...))
}

// Error prints a message with the prefix "ERROR: " in red.
func (l *Logger) Error(message string, a ...interface{}) {
	l.logger.Println(color.RedString("ERROR:"), fmt.Sprintf(message, a...))
}

func NewConsoleLogger() *Logger {
	return &Logger{logger: log.New(os.Stdout, "", 0)}
}
