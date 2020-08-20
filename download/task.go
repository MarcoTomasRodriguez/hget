package download

import (
	"encoding/json"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Task represents a download.
type Task struct {
	Url   string
	Parts []Part
}

// Part is a slice of the file downloaded.
type Part struct {
	Url       string
	Path      string
	RangeFrom int64
	RangeTo   int64
}

// SaveTask saves the current task as json into $HOME/ProgramFolder/Filename/TaskFilename
func (task *Task) SaveTask() error {
	// make temp folder
	// only working in unix with env HOME
	folder := utils.FolderOf(task.Url)
	logger.Info("Saving current download data in %s\n", folder)
	if err := utils.MkdirIfNotExist(folder); err != nil {
		return err
	}

	// move current downloading file to data folder
	for _, part := range task.Parts {
		if err := os.Rename(part.Path, filepath.Join(folder, filepath.Base(part.Path))); err != nil {
			return err
		}
	}

	// save task file
	jsonTask, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(folder, config.TaskFilename), jsonTask, 0644)
}

// ReadTask reads the task from $HOME/ProgramFolder/Filename/TaskFilename
func ReadTask(taskName string) (*Task, error) {
	file := filepath.Join(config.Home, config.ProgramFolder, taskName, config.TaskFilename)
	logger.Info("Getting data from %s\n", file)

	jsonTask, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	task := new(Task)
	err = json.Unmarshal(jsonTask, task)

	return task, err
}

// GetAllTasks returns all the saved tasks
func GetAllTasks() ([]string, error) {
	tasks := make([]string, 0)

	tasksFolder, err := ioutil.ReadDir(filepath.Join(config.Home, config.ProgramFolder))
	if err != nil {
		return tasks, err
	}

	for _, t := range tasksFolder {
		if t.IsDir() { tasks = append(tasks, t.Name()) }
	}

	return tasks, nil
}