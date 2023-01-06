package logger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type LoggerSuite struct {
	suite.Suite
	tl  Logger
	buf bytes.Buffer
}

func (s *LoggerSuite) SetupSuite() {
	s.tl = &consoleLogger{logger: log.New(&s.buf, "", 0)}
}

func (s *LoggerSuite) SetupTest() {
	s.buf.Reset()
}

func (s *LoggerSuite) TestLogger_Info() {
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
			s.tl.Info(tc.message, tc.args...)
			assert.Equal(s.T(), s.buf.String(), tc.expected)
		})
	}
}

func (s *LoggerSuite) TestLogger_Warn() {
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
			s.tl.Warn(tc.message, tc.args...)
			assert.Equal(s.T(), s.buf.String(), tc.expected)
		})
	}
}

func (s *LoggerSuite) TestLogger_Error() {
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
			s.tl.Error(tc.message, tc.args...)
			assert.Equal(s.T(), s.buf.String(), tc.expected)
		})
	}
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
