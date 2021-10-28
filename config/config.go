package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Configuration defines the behaviour of the application.
type Configuration struct {
	// ProgramFolder is the folder used by the program to save temporal files, such as ongoing and paused downloads.
	ProgramFolder string `mapstructure:"program_folder"`

	// LogLevel restricts the logs to what the user wants to get. 0 means no logs, 1 only important logs and 2 all logs.
	LogLevel uint8 `mapstructure:"log_level"`

	// Download ...
	Download struct {
		// Folder defines the directory in which the downloaded file will be moved. Default: current execution path.
		Folder string `mapstructure:"folder"`

		// CopyNBytes sets the bytes to copy in a row from the response body.
		CopyNBytes int64 `mapstructure:"copy_n_bytes"`

		// UrlChecksumLength sets the length of the checksum used to prevent collisions. Maximum: 32.
		UrlChecksumLength uint8 `mapstructure:"url_checksum_length"`

		// CollisionProtection enables/disables the collision protection using a hash when saving the file to the final destination.
		CollisionProtection bool `mapstructure:"collision_protection"`
	}
}

// Filepath is the path of the configuration file.
var Filepath string

// Config is the shared configuration instance.
var Config = &Configuration{}

// DownloadFolder ...
func (config Configuration) DownloadFolder() string {
	return filepath.Join(config.ProgramFolder, "downloads")
}

// Validate validates the config.
func (config Configuration) Validate() error {
	if config.Download.UrlChecksumLength > 32 {
		return fmt.Errorf("UrlChecksumLength should be between 0 and 32")
	}

	if config.Download.CopyNBytes < 0 {
		return fmt.Errorf("CopyNBytes should be greater than 0")
	}

	if config.LogLevel > 2 {
		return fmt.Errorf("LogLevel should be between 0 and 2")
	}

	return nil
}

// LoadConfig loads the config from the configuration file.
func LoadConfig() {
	// Get home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: could not get home directory %v\n", err)
		panic(err)
	}

	// Get working directory.
	workingDir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: could not get working directory %v\n", err)
		panic(err)
	}

	// Set defaults.
	viper.SetDefault("program_folder", filepath.Join(homeDir, ".hget"))
	viper.SetDefault("download.folder", workingDir)
	viper.SetDefault("download.copy_n_bytes", 300)
	viper.SetDefault("download.url_checksum_length", 16)
	viper.SetDefault("download.collision_protection", false)

	// Check if config file exists.
	if _, err := os.Stat(Filepath); !os.IsNotExist(err) {
		// Set config file.
		viper.SetConfigFile(Filepath)

		// Read config.
		if err := viper.ReadInConfig(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: unable to read configuration file: %v\n", err)
			panic(err)
		}
	}

	// Unmarshal configuration into the shared config struct.
	if err := viper.Unmarshal(Config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: unable to decode configuration into struct: %v\n", err)
		panic(err)
	}

	if err := Config.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: invalid configuration: %v\n", err)
		panic(err)
	}

	// Create internal program folders.
	_ = os.MkdirAll(filepath.Join(Config.ProgramFolder, "downloads"), os.ModePerm)
}
