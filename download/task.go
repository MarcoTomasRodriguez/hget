package download

import (
	"errors"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Task represents a download.
type Task struct {
	URL   string
	Parts []Part
}

// Part is a slice of the file downloaded.
type Part struct {
	Index     int64
	Path      string
	RangeFrom int64
	RangeTo   int64
}

// SaveTask saves the current task as json into $HOME/ProgramFolder/Filename/TaskFilename
func (task *Task) SaveTask() error {
	// Make temp folder. Only working in unix with env HOME.
	folder := utils.FolderOf(task.URL)
	logger.Info("Saving current download data in %s\n", folder)
	if err := utils.MkdirIfNotExist(folder); err != nil {
		return err
	}

	// Move downloaded files to the task folder
	for _, part := range task.Parts {
		if err := os.Rename(part.Path, filepath.Join(folder, filepath.Base(part.Path))); err != nil {
			return err
		}
	}

	// Save task file
	tomlTask, err := toml.Marshal(task)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(folder, config.Config.TaskFilename), tomlTask, 0644)
}

// ReadTask reads the task from $HOME/ProgramFolder/Filename/TaskFilename
func ReadTask(taskName string) (*Task, error) {
	file := filepath.Join(config.Config.ProgramFolder, taskName, config.Config.TaskFilename)
	logger.Info("Getting data from %s\n", file)

	// Load the task
	tomlTask, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Parse the task
	task := new(Task)
	err = toml.Unmarshal(tomlTask, task)

	return task, err
}

// FindTask finds a tasks by his identifier (this can be the task-name or the URL).
func FindTask(identifier string) (string, error) {
	if utils.Exists(filepath.Join(config.Config.ProgramFolder, identifier, config.Config.TaskFilename)) {
		return identifier, nil
	}

	URL, err := utils.ResolveURL(identifier)
	if err != nil {
		return "", err
	}

	taskName := utils.FilenameWithHash(URL)

	if utils.Exists(filepath.Join(config.Config.ProgramFolder, taskName, config.Config.TaskFilename)) {
		return taskName, nil
	}

	return "", errors.New("task not found")
}

// GetAllTasks returns all the saved tasks
func GetAllTasks() ([]string, error) {
	tasks := make([]string, 0)

	tasksFolder, err := ioutil.ReadDir(config.Config.ProgramFolder)
	if err != nil {
		return tasks, err
	}

	for _, t := range tasksFolder {
		if t.IsDir() {
			tasks = append(tasks, t.Name())
		}
	}

	return tasks, nil
}

// RemoveTask removes a task by taskName.
func RemoveTask(taskName string) error {
	if !strings.Contains(taskName, "..") {
		return os.RemoveAll(filepath.Join(config.Config.ProgramFolder, taskName))
	}
	return fmt.Errorf("illegal task name")
}

// RemoveAllTasks removes all the tasks.
func RemoveAllTasks() error {
	tasks, err := GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err := RemoveTask(task)
		if err != nil {
			return err
		}
	}

	return nil
}
