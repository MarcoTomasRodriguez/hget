package utils

import (
	"strings"
	"testing"

	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/stretchr/testify/assert"
)

func TestHashFilename(t *testing.T) {
	config.Config.Download.UrlChecksumLength = 16

	cases := []struct {
		url      string
		filename string
		expected string
	}{
		{"localhost", "index.html", "49960de5880e8c68-index.html"},
		{"localhost/my-file.file", "my-file.file", "9832caf176bed618-my-file.file"},
	}

	for _, v := range cases {
		t.Run(v.expected, func(t *testing.T) {
			assert.Equal(t, v.expected, HashFilename(v.url, v.filename))
		})
	}
}

func TestCheckFilenameValidity_Valid(t *testing.T) {
	cases := []struct {
		filename string
	}{
		{"index.html"},
		{"go1.17.2.src.tar.gz"},
		{"binary"},
	}

	for _, v := range cases {
		t.Run(v.filename, func(t *testing.T) {
			err := CheckFilenameValidity(v.filename)
			assert.NoError(t, err)
		})
	}
}

func TestCheckFilenameValidity_Invalid(t *testing.T) {
	cases := []struct {
		filename string
		expected error
	}{
		{"", ErrFilenameEmpty},
		{".", ErrFilenameEmpty},
		{"..", ErrFilenameEmpty},
		{"/", ErrFilenameEmpty},
		{strings.Repeat("a", 263), ErrFilenameTooLong},
	}

	for _, v := range cases {
		t.Run(v.filename, func(t *testing.T) {
			err := CheckFilenameValidity(v.filename)
			assert.Equal(t, v.expected, err)
		})
	}
}

func TestResolveURL_ValidURL(t *testing.T) {
	cases := []struct {
		rawURL   string
		expected string
	}{
		{"google.com", "https://google.com"},
		{"https://google.com", "https://google.com"},
		{"http://google.com", "http://google.com"},
	}

	for _, v := range cases {
		t.Run(v.expected, func(t *testing.T) {
			url, err := ResolveURL(v.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, v.expected, url)
		})
	}
}

func TestResolveURL_InvalidURL(t *testing.T) {
	cases := []struct{ rawURL string }{
		{"https://not a rawURL"},
		{"stMHaWpBSem0OgfcVi6M"},
		{""},
	}

	for _, v := range cases {
		t.Run(v.rawURL, func(t *testing.T) {
			url, err := ResolveURL(v.rawURL)
			assert.Error(t, err)
			assert.Empty(t, url)
		})
	}
}

func TestReadableMemorySize(t *testing.T) {
	cases := []struct {
		bytes    uint64
		expected string
	}{
		{10 * B, "10 B"},
		{512 * B, "512 B"},
		{1 * KB, "1.0 KB"},
		{1024*KB - 1, "1024.0 KB"},
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
