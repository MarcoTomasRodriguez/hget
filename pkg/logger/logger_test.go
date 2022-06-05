package logger

import (
	"bytes"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
)

type LoggerSuite struct {
	suite.Suite
	buffer bytes.Buffer
}

func (s *LoggerSuite) SetupSuite() {
	do.ProvideValue[*log.Logger](do.DefaultInjector, log.New(&s.buffer, "", 0))
}

func (s *LoggerSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *LoggerSuite) TearDownSuite() {
	do.ProvideValue[*log.Logger](do.DefaultInjector, log.New(os.Stdout, "", 0))
}

func (s *LoggerSuite) TestInfo() {
	testCases := []struct {
		message  string
		args     []any
		expected string
	}{
		{
			"This is an info message with an argument: %s!",
			[]any{"1 argument"},
			"INFO: This is an info message with an argument: 1 argument!\n",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.expected, func() {
			Info(tc.message, tc.args...)
			assert.Equal(s.T(), s.buffer.String(), tc.expected)
		})
	}
}

func (s *LoggerSuite) TestWarn() {
	testCases := []struct {
		message  string
		args     []any
		expected string
	}{
		{
			"This is a warning message with an argument: %s!",
			[]any{"1 argument"},
			"WARN: This is a warning message with an argument: 1 argument!\n",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.expected, func() {
			Warn(tc.message, tc.args...)
			assert.Equal(s.T(), s.buffer.String(), tc.expected)
		})
	}
}

func (s *LoggerSuite) TestError() {
	testCases := []struct {
		message  string
		args     []any
		expected string
	}{
		{
			"This is an error message with an argument: %s!",
			[]any{"1 argument"},
			"ERROR: This is an error message with an argument: 1 argument!\n",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.expected, func() {
			Error(tc.message, tc.args...)
			assert.Equal(s.T(), s.buffer.String(), tc.expected)
		})
	}
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
