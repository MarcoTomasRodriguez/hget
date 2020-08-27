package config

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Configuration is the shared object which is used to persist information both inside the program and outside of it.
type Configuration struct {
	// ProgramFolder is the folder in which the program will store his information
	// about the ongoing downloads. This path is relative to $HOME.
	ProgramFolder string `toml:"-"`

	// ConfigFilename is the
	ConfigFilename string `toml:"-"`

	// TaskFilename represents the state of a download. This file will be located
	// in $HOME/ProgramFolder/Download
	TaskFilename string `toml:"-"`

	// DisplayProgressBar enables/disables the display of the progress bar.
	DisplayProgressBar bool `toml:"display_progress_bar"`

	// UseHashLength sets the length of the hash used to prevent collisions.
	// Note that this can never be more than 32
	UseHashLength uint8 `toml:"use_hash_length"`

	// SaveWithHash enables/disables the collision protection using a hash
	// while moving the file from inside the program to outside.
	SaveWithHash bool `toml:"save_with_hash"`

	// CopyNBytes sets the bytes to copy in a row from the response body.
	CopyNBytes int64 `toml:"copy_n_bytes"`

	// LogLevel restricts the logs to what the user wants to get.
	// 0 means no logs, 1 only important logs and 2 all logs.
	LogLevel uint8 `toml:"log_level"`

	// DownloadFolder defines in which directory the downloaded file will be moved to.
	// If it is empty, then the download folder will be the terminal cwd.
	DownloadFolder string `toml:"download_folder"`
}

// DefaultConfig is the default configuration instance.
// The Config instance is based on these defaults.
var DefaultConfig = &Configuration{
	ProgramFolder:      filepath.Join(os.Getenv("HOME"), ".hget"),
	ConfigFilename:     "config.toml",
	TaskFilename:       "task.json",
	DisplayProgressBar: isatty.IsTerminal(os.Stdout.Fd()),
	UseHashLength:      uint8(16),
	SaveWithHash:       false,
	CopyNBytes:         int64(250),
	LogLevel:           uint8(2),
	DownloadFolder:     "",
}

// Config is the shared configuration instance.
var Config = DefaultConfig

// LoadConfig loads the configuration from a toml file merging it with the default config.
func LoadConfig(configFilepath string) (*Configuration, error) {
	config := DefaultConfig
	loadedConfig := &Configuration{}

	// Read config
	file, err := ioutil.ReadFile(configFilepath)
	if err != nil {
		return config, fmt.Errorf("unable to read configuration file: %v", err)
	}

	// Parse config
	if err = toml.Unmarshal(file, loadedConfig); err != nil {
		return config, fmt.Errorf("unable to read configuration file: %v", err)
	}

	// Write to shared object
	config.DisplayProgressBar = config.DisplayProgressBar && loadedConfig.DisplayProgressBar
	if loadedConfig.UseHashLength <= 32 {
		config.UseHashLength = loadedConfig.UseHashLength
	}
	if loadedConfig.CopyNBytes > 0 {
		config.CopyNBytes = loadedConfig.CopyNBytes
	}
	config.SaveWithHash = loadedConfig.SaveWithHash
	if loadedConfig.LogLevel <= 3 {
		config.LogLevel = loadedConfig.LogLevel
	}
	config.DownloadFolder = loadedConfig.DownloadFolder

	return config, nil
}
