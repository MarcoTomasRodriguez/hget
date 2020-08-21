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
	task := Task{Url: "localhost/my_file.file", Parts: []Part{}}
	task2 := Task{Url: "localhost/another-file", Parts: []Part{}}

	assert.NoError(t, RemoveAllTasks())

	assert.NoError(t, task.SaveTask())

	assert.True(t, utils.ExistDir(utils.FolderOf(task.Url)))
	assert.True(t, utils.ExistDir(filepath.Join(utils.FolderOf(task.Url), config.TaskFilename)))

	assert.NoError(t, task2.SaveTask())

	assert.True(t, utils.ExistDir(utils.FolderOf(task2.Url)))
	assert.True(t, utils.ExistDir(filepath.Join(utils.FolderOf(task2.Url), config.TaskFilename)))

	tasks, err := GetAllTasks()
	actualTasks := []string{utils.FilenameWithHash(task.Url), utils.FilenameWithHash(task2.Url)}
	sort.Strings(tasks)
	sort.Strings(actualTasks)
	assert.NoError(t, err)
	assert.Equal(t, tasks, actualTasks)

	savedTask, err := ReadTask(utils.FilenameWithHash(task.Url))
	assert.NoError(t, err)
	assert.Equal(t, task, *savedTask)
}
