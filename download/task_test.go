package download

import (
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestTask(t *testing.T) {
	task := Task{URL: "localhost/my_file.file", Parts: nil}
	task2 := Task{URL: "localhost/another-file", Parts: nil}

	assert.NoError(t, RemoveAllTasks())

	assert.NoError(t, task.SaveTask())

	assert.True(t, utils.Exists(utils.FolderOf(task.URL)))
	assert.True(t, utils.Exists(filepath.Join(utils.FolderOf(task.URL), config.Config.TaskFilename)))

	assert.NoError(t, task2.SaveTask())

	assert.True(t, utils.Exists(utils.FolderOf(task2.URL)))
	assert.True(t, utils.Exists(filepath.Join(utils.FolderOf(task2.URL), config.Config.TaskFilename)))

	tasks, err := GetAllTasks()
	actualTasks := []string{utils.FilenameWithHash(task.URL), utils.FilenameWithHash(task2.URL)}
	sort.Strings(tasks)
	sort.Strings(actualTasks)
	assert.NoError(t, err)
	assert.Equal(t, tasks, actualTasks)

	savedTask, err := ReadTask(utils.FilenameWithHash(task.URL))
	assert.NoError(t, err)
	assert.Equal(t, task, *savedTask)

	assert.NoError(t, os.RemoveAll(utils.FolderOf(task.URL)))
	assert.NoError(t, os.RemoveAll(utils.FolderOf(task2.URL)))
}
