package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"sort"
	"testing"
)

func TestTask(t *testing.T) {
	pathToProgram := filepath.Join(config.Home, config.ProgramFolder)
	task := Task{Url: "localhost/my_file.file", Parts: []Part{}}
	task2 := Task{Url: "localhost/another-file", Parts: []Part{}}

	assert.NoError(t, task.SaveTask())

	assert.True(t, utils.ExistDir(filepath.Join(pathToProgram, "my_file.file")))
	assert.True(t, utils.ExistDir(filepath.Join(pathToProgram, "my_file.file", config.TaskFilename)))

	assert.NoError(t, task2.SaveTask())

	assert.True(t, utils.ExistDir(filepath.Join(pathToProgram, "another-file")))
	assert.True(t, utils.ExistDir(filepath.Join(pathToProgram, "another-file", config.TaskFilename)))

	tasks, err := GetAllTasks()
	actualTasks := []string{"file", "my_file.file", "another-file"}
	sort.Strings(tasks)
	sort.Strings(actualTasks)
	assert.NoError(t, err)
	assert.Equal(t, tasks, actualTasks)

	savedTask, err := ReadTask("my_file.file")
	assert.NoError(t, err)
	assert.Equal(t, task, *savedTask)
}
