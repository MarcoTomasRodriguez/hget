package fsutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckFilenameValidity_Invalid(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"", false},
		{".", false},
		{"/test", false},
		{"test/test.txt", false},
		{"test.txt", true},
		{strings.Repeat("a", 263), false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			assert.Equal(t, tc.expected, ValidateFilename(tc.filename))
		})
	}
}

func TestReadableMemorySize(t *testing.T) {
	cases := []struct {
		bytes    uint64
		expected string
	}{
		{10, "10 B"},
		{512, "512 B"},
		{1 * KB, "1.0 kB"},
		{1000*KB - 1, "1000.0 kB"},
		{1 * MB, "1.0 MB"},
		{2.5 * MB, "2.5 MB"},
		{1 * GB, "1.0 GB"},
		{1 * TB, "1.0 TB"},
	}

	for _, v := range cases {
		t.Run(v.expected, func(t *testing.T) {
			assert.Equal(t, v.expected, ReadableMemorySize(v.bytes))
		})
	}
}
