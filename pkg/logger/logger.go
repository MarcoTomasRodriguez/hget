package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
)

type Logger interface {
	Info(message string, a ...interface{})
	Warn(message string, a ...interface{})
	Error(message string, a ...interface{})
}

type consoleLogger struct {
	logger *log.Logger
}

// Info prints a message with the prefix "INFO: " in cyan.
func (l *consoleLogger) Info(message string, a ...interface{}) {
	l.logger.Println(color.CyanString("INFO:"), fmt.Sprintf(message, a...))
}

// Warn prints a message with the prefix "WARN: " in yellow.
func (l *consoleLogger) Warn(message string, a ...interface{}) {
	l.logger.Println(color.YellowString("WARN:"), fmt.Sprintf(message, a...))
}

// Error prints a message with the prefix "ERROR: " in red.
func (l *consoleLogger) Error(message string, a ...interface{}) {
	l.logger.Println(color.RedString("ERROR:"), fmt.Sprintf(message, a...))
}

type NoopConsoleLogger struct{}

func (n NoopConsoleLogger) Info(message string, a ...interface{}) {}

func (n NoopConsoleLogger) Warn(message string, a ...interface{}) {}

func (n NoopConsoleLogger) Error(message string, a ...interface{}) {}

// NewConsoleLogger creates a logger using the console as output.
func NewConsoleLogger() Logger {
	return &consoleLogger{logger: log.New(os.Stdout, "", 0)}
}

var _ Logger = (*consoleLogger)(nil)
