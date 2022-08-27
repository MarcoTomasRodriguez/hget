package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config defines the behaviour of the application.
type Config struct {
	// ProgramFolder is the folder used by the program to save temporal files, such as ongoing and paused downloads.
	ProgramFolder string `mapstructure:"program_folder"`

	// LogLevel restricts the logs to what the user wants to get. 0 means no logs, 1 only important logs and 2 all logs.
	LogLevel int `mapstructure:"log_level"`

	// Download defines the configuration relative to a download.
	Download struct {
		// Folder defines the directory in which the downloaded file will be moved.
		Folder string `mapstructure:"folder"`

		// CopyNBytes sets the bytes to copy in a row from the response body.
		CopyNBytes int64 `mapstructure:"copy_n_bytes"`
	}
}

// DownloadFolder returns the path to the internal download folder.
func (config *Config) DownloadFolder() string {
	return filepath.Join(config.ProgramFolder, "downloads")
}

// NewConfig initializes the config object from a toml file.
func NewConfig(filename string) (*Config, error) {
	// Get home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get home directory: %v", err)
	}

	// Get working directory.
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get working directory: %v", err)
	}

	// Set defaults.
	viper.SetDefault("program_folder", filepath.Join(homeDir, ".hget"))
	viper.SetDefault("download.folder", workingDir)
	viper.SetDefault("download.copy_n_bytes", 300)
	viper.SetDefault("download.collision_protection", false)

	// Check if config file exists.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		// Set config file.
		viper.SetConfigFile(filename)

		// Read config.
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("unable to read configuration file: %v", err)
		}
	}

	// Unmarshal configuration into the shared config struct.
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration into struct: %v", err)
	}

	// Validate configuration.
	if cfg.Download.CopyNBytes < 0 {
		return nil, fmt.Errorf("CopyNBytes should be greater than 0: %v", err)
	}

	if cfg.LogLevel > 2 {
		return nil, fmt.Errorf("LogLevel should be between 0 and 2: %v", err)
	}

	// Create internal program folders.
	_ = os.MkdirAll(filepath.Join(cfg.ProgramFolder, "downloads"), 0755)

	return cfg, nil
}
