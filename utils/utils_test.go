package utils

import (
	"errors"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"
)

func TestFatalCheck(t *testing.T) {
	assert.Panics(t, func() { FatalCheck(errors.New("error")) }, "FatalCheck should panic on error")
	assert.NotPanics(t, func() { FatalCheck(nil) }, "FatalCheck should not panic on error nil")
}

func TestStringifyIpsV4(t *testing.T) {}

func TestMkdirIfNotExist(t *testing.T) {}

func TestExistDir(t *testing.T) {}

func TestHashOf(t *testing.T) {
	data := []string{"test", "file", "test-test-test-test-test-test", "test-test-test-test-test-test-test-test-test-test-test"}

	for _, str := range data {
		assert.Equal(t, HashOf(str), HashOf(str))
	}

	assert.NotEqual(t, HashOf(data[0]), HashOf(data[1]))
}

/*
// RemoveHashFromFilename returns the basename + the hash of the url.
func FilenameWithHash(url string) string {
	base := filepath.Base(url)
	hash := HashOf(url)[:config.UseHashLength]
	if base == "." {
		logger.Panic(errors.New("there is no basename for the url"))
	}

	filename := hash + "-" + base
	if len(filename) > FilenameLengthLimit {
		logger.Panic(fmt.Errorf("the filename length should never exceed the limit of %d",
			FilenameLengthLimit - len(hash) + 1))
	}

	return filename
}
*/
func TestFilenameWithHash(t *testing.T) {
	data := []string{"localhost", "localhost/my-file.file"}

	for _, url := range data {
		assert.Equal(t, HashOf(url)[:config.UseHashLength] + "-" + filepath.Base(url), FilenameWithHash(url))
	}

	assert.Panics(t, func() { FilenameWithHash(".") } )
	assert.Panics(t, func() { FilenameWithHash(strings.Repeat("-", 255)) } )

}

func TestFolderOf(t *testing.T) {
	// HashOf("localhost")[:config.UseHashLength] + "-localhost"
	data := []string{"localhost", "localhost/my-file.file"}

	for _, url := range data {
		assert.Equal(t, filepath.Join(config.Home, config.ProgramFolder, FilenameWithHash(url)), FolderOf(url))
	}

	assert.Panics(t, func() { FolderOf("localhost/../") })
	assert.Panics(t, func() { FolderOf("localhost/.") })
	assert.Panics(t, func() { FolderOf("") })
}

func TestIsUrl(t *testing.T) {
	assert.True(t, IsUrl("http://gooGle.com.ar"))
	assert.True(t, IsUrl("https://google.Com"))
	assert.True(t, IsUrl("GoOgLe.c0m.ar"))
	assert.False(t, IsUrl("https://not a url.com"))
}

func TestReadableMemorySize(t *testing.T) {
	assert.Equal(t, "0.0 KB", ReadableMemorySize(Byte * 10))
	assert.Equal(t, "0.5 KB", ReadableMemorySize(KiloByte * 0.5))
	assert.Equal(t, "1.0 KB", ReadableMemorySize(KiloByte))
	assert.Equal(t, "1024.0 KB", ReadableMemorySize(MegaByte-1))
	assert.Equal(t, "1.0 MB", ReadableMemorySize(MegaByte))
	assert.Equal(t, "2.5 MB", ReadableMemorySize(MegaByte*2.5))
	assert.Equal(t, "1.0 GB", ReadableMemorySize(GigaByte))
	assert.Equal(t, "1.0 TB", ReadableMemorySize(TeraByte))
}