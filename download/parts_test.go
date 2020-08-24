package download

import (
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJoiner(t *testing.T) {
	files := []string{"file1", "file2"}

	for _, file := range files {
		assert.NoError(t, ioutil.WriteFile(file, []byte(file), 0600))
	}

	assert.NoError(t, JoinParts(files[:], "join"))

	content, err := ioutil.ReadFile("join")
	assert.NoError(t, err)

	assert.Equal(t, strings.Join(files, ""), string(content))

	for _, file := range files {
		assert.NoError(t, os.Remove(file))
	}

	assert.NoError(t, os.Remove("join"))
}

func TestCalculateParts(t *testing.T) {
	assert.NoError(t, RemoveAllTasks())
	assert.Equal(t, []Part{}, CalculateParts("url", 0, 150))
	assert.Equal(t, []Part{
		{Path: filepath.Join(utils.FolderOf("my-url"), utils.MakePartName(0, 3)), RangeFrom: 0, RangeTo: 150},
	}, CalculateParts("my-url", 1, 150))
	assert.Equal(t, []Part{
		{Path: filepath.Join(utils.FolderOf("google.com"), utils.MakePartName(0, 3)), RangeFrom: 0, RangeTo: 49},
		{Path: filepath.Join(utils.FolderOf("google.com"), utils.MakePartName(1, 3)), RangeFrom: 50, RangeTo: 99},
		{Path: filepath.Join(utils.FolderOf("google.com"), utils.MakePartName(2, 3)), RangeFrom: 100, RangeTo: 150},
	}, CalculateParts("google.com", 3, 150))
}
