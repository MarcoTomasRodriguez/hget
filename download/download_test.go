package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

// RDA means range download allowed.
const RDA = "https://raw.githubusercontent.com/MarcoTomasRodriguez/hget/master/README.md"

// RDNA means range download not allowed.
const RDNA = "https://github.com/MarcoTomasRodriguez/hget"

func DownloadTest(t *testing.T, url string, parallelism int) {
	config.Config.LogLevel = uint8(0)

	// Get future filename of the download
	filenameWithHash := utils.FilenameWithHash(url)
	filenameWithoutHash := utils.FilenameWithoutHash(url)

	// Send request to url (just to retrieve useful headers such as Content-Length)
	resp, err := http.Get(url)
	assert.NoError(t, err)

	// Download with hash
	config.Config.SaveWithHash = true
	Download(url, nil, parallelism)
	fileInfo, err := os.Stat(filenameWithHash)

	// Check if file was created
	assert.False(t, os.IsNotExist(err))

	// Verify length (not integrity)
	if resp.ContentLength != -1 {
		assert.Equal(t, resp.ContentLength, fileInfo.Size())
	}

	// Remove file
	assert.NoError(t, os.Remove(filenameWithHash))

	// Download without hash
	config.Config.SaveWithHash = false
	Download(url, nil, parallelism)
	fileInfo, err = os.Stat(filenameWithoutHash)

	// Check if file was created
	assert.False(t, os.IsNotExist(err))

	// Verify integrity (not integrity)
	if resp.ContentLength != -1 {
		assert.Equal(t, resp.ContentLength, fileInfo.Size())
	}

	// Remove file
	assert.NoError(t, os.Remove(filenameWithoutHash))
}

func TestDownload_Single_Thread_RDNA(t *testing.T) {
	DownloadTest(t, RDNA, 1)
}

func TestDownload_Multiple_Thread_RDNA(t *testing.T) {
	DownloadTest(t, RDNA, 8)
	DownloadTest(t, RDNA, 101)
}

func TestDownload_Single_Thread_RDA(t *testing.T) {
	DownloadTest(t, RDA, 1)
}

func TestDownload_Multiple_Thread_RDA(t *testing.T) {
	DownloadTest(t, RDNA, 8)
	DownloadTest(t, RDNA, 101)
}
