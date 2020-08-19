package download

import (
	"encoding/json"
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/config"
	"github.com/MarcoTomasRodriguez/hget/logger"
	"github.com/MarcoTomasRodriguez/hget/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

// State represents a file in state.
type State struct {
	Url   string
	Parts []Part
}

// Part is a slice of the file.
type Part struct {
	Url       string
	Path      string
	RangeFrom int64
	RangeTo   int64
}

// Save saves the current state as json into $HOME/ProgramFolder/Filename/StateFilename
func (state *State) Save() error {
	// make temp folder
	// only working in unix with env HOME
	folder := utils.FolderOf(state.Url)
	logger.Info("Saving current download data in %s\n", folder)
	if err := utils.MkdirIfNotExist(folder); err != nil {
		return err
	}

	// move current downloading file to data folder
	for _, part := range state.Parts {
		if err := os.Rename(part.Path, filepath.Join(folder, filepath.Base(part.Path))); err != nil {
			return err
		}
	}

	// save state file
	jsonState, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(folder, config.StateFilename), jsonState, 0644)
}

// Read reads the current state from $HOME/ProgramFolder/Filename/StateFilename
func Read(task string) (*State, error) {
	file := filepath.Join(os.Getenv("HOME"), config.ProgramFolder, task, config.StateFilename)
	logger.Info("Getting data from %s\n", file)

	jsonState, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	state := new(State)
	err = json.Unmarshal(jsonState, state)

	return state, err
}

// PrintTasks prints all the saved tasks
func PrintTasks() error {
	downloading, err := ioutil.ReadDir(filepath.Join(os.Getenv("HOME"), config.ProgramFolder))
	if err != nil {
		return err
	}

	logger.Info("Currently on going download:\n")

	for _, d := range downloading {
		if d.IsDir() { fmt.Println(d.Name()) }
	}

	return nil
}