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

func TestMkdirIfNotExist(t *testing.T) {}

func TestExistDir(t *testing.T) {}

func TestHashOf(t *testing.T) {
	data := []string{"test", "file", "test-test-test-test-test-test", "test-test-test-test-test-test-test-test-test-test-test"}

	for _, str := range data {
		assert.Equal(t, HashOf(str), HashOf(str))
	}

	assert.NotEqual(t, HashOf(data[0]), HashOf(data[1]))
}

func TestFilenameWithHash(t *testing.T) {
	data := []string{"localhost", "localhost/my-file.file"}

	for _, url := range data {
		assert.Equal(t, HashOf(url)[:config.UseHashLength]+"-"+filepath.Base(url), FilenameWithHash(url))
	}

	assert.Panics(t, func() { FilenameWithHash(".") })
	assert.Panics(t, func() { FilenameWithHash(strings.Repeat("-", 255)) })
}

func TestFilenameWithoutHash(t *testing.T) {
	data := []string{"localhost", "localhost/my-file.file"}

	for _, url := range data {
		assert.Equal(t, filepath.Base(url), FilenameWithoutHash(url))
	}

	assert.Panics(t, func() { FilenameWithHash(".") })
	assert.Panics(t, func() { FilenameWithHash(strings.Repeat("-", 256)) })
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
	assert.Equal(t, "0.0 KB", ReadableMemorySize(Byte*10))
	assert.Equal(t, "0.5 KB", ReadableMemorySize(KiloByte*0.5))
	assert.Equal(t, "1.0 KB", ReadableMemorySize(KiloByte))
	assert.Equal(t, "1024.0 KB", ReadableMemorySize(MegaByte-1))
	assert.Equal(t, "1.0 MB", ReadableMemorySize(MegaByte))
	assert.Equal(t, "2.5 MB", ReadableMemorySize(MegaByte*2.5))
	assert.Equal(t, "1.0 GB", ReadableMemorySize(GigaByte))
	assert.Equal(t, "1.0 TB", ReadableMemorySize(TeraByte))
}

func TestPartName(t *testing.T) {
	assert.Equal(t, "part.0", PartName(0, 1))
	assert.Equal(t, "part.9", PartName(9, 10))
	assert.Equal(t, "part.00", PartName(0, 100))
	assert.Equal(t, "part.00", PartName(0, 100))
	assert.Equal(t, "part.99", PartName(99, 100))
	assert.Equal(t, "part.100", PartName(100, 101))
	assert.Equal(t, "part.500", PartName(500, 1000))
	assert.Equal(t, "part.12034", PartName(12034, 14623))
}
