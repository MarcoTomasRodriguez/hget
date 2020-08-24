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
	task := Task{URL: "localhost/my_file.file", Parts: []Part{}}
	task2 := Task{URL: "localhost/another-file", Parts: []Part{}}

	assert.NoError(t, RemoveAllTasks())

	assert.NoError(t, task.SaveTask())

	assert.True(t, utils.ExistDir(utils.FolderOf(task.URL)))
	assert.True(t, utils.ExistDir(filepath.Join(utils.FolderOf(task.URL), config.TaskFilename)))

	assert.NoError(t, task2.SaveTask())

	assert.True(t, utils.ExistDir(utils.FolderOf(task2.URL)))
	assert.True(t, utils.ExistDir(filepath.Join(utils.FolderOf(task2.URL), config.TaskFilename)))

	tasks, err := GetAllTasks()
	actualTasks := []string{utils.FilenameWithHash(task.URL), utils.FilenameWithHash(task2.URL)}
	sort.Strings(tasks)
	sort.Strings(actualTasks)
	assert.NoError(t, err)
	assert.Equal(t, tasks, actualTasks)

	savedTask, err := ReadTask(utils.FilenameWithHash(task.URL))
	assert.NoError(t, err)
	assert.Equal(t, task, *savedTask)
}
