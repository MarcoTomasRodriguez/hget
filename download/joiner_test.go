package download

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestJoiner(t *testing.T) {
	files := []string{"file1", "file2"}

	for _, file := range files {
		assert.NoError(t, ioutil.WriteFile(file, []byte(file), 0600))
	}

	assert.NoError(t, JoinFile(files[:], "join"))

	content, err := ioutil.ReadFile("join")
	assert.NoError(t, err)

	assert.Equal(t, strings.Join(files, ""), string(content))

	for _, file := range files {
		assert.NoError(t, os.Remove(file))
	}

	assert.NoError(t, os.Remove("join"))
}