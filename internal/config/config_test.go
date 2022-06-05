package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownloadFolder(t *testing.T) {
	testCases := []struct {
		programFolder          string
		expectedDownloadFolder string
	}{
		{"/home/user/.hget", "/home/user/.hget/downloads"},
		{"/home/user/.hget//", "/home/user/.hget/downloads"},
		{"/home/user/folder/", "/home/user/folder/downloads"},
	}

	for _, tc := range testCases {
		t.Run(tc.programFolder, func(t *testing.T) {
			cfg := Config{ProgramFolder: tc.programFolder}
			assert.Equal(t, tc.expectedDownloadFolder, cfg.DownloadFolder())
		})
	}

}
