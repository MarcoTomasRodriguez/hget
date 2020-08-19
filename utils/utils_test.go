package utils

import (
	"errors"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestFatalCheck(t *testing.T) {
	assert.Panics(t, func() { FatalCheck(errors.New("error")) }, "FatalCheck should panic on error")
	assert.NotPanics(t, func() { FatalCheck(nil) }, "FatalCheck should not panic on error nil")
}

func TestStringifyIpsV4(t *testing.T) {}

func TestMkdirIfNotExist(t *testing.T) {}

func TestExistDir(t *testing.T) {}

func TestFolderOf(t *testing.T) {
	assert.Equal(t, FolderOf("localhost/my-file.file"), filepath.Join(config.Home, config.ProgramFolder, "my-file.file"))
	assert.Equal(t, FolderOf("localhost"), filepath.Join(config.Home, config.ProgramFolder, "localhost"))
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