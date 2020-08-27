package config

import (
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const configFile = "config.toml"

func TestLoadConfig(t *testing.T) {
	err := os.Remove(configFile)
	assert.Error(t, err)

	config, err := LoadConfig(configFile)
	assert.Error(t, err)
	assert.Equal(t, DefaultConfig, config)

	file, err := os.Create(configFile)
	assert.NoError(t, err)
	_, err = file.WriteString("this is not toml")
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	config, err = LoadConfig(configFile)
	assert.Error(t, err)
	assert.Equal(t, DefaultConfig, config)

	file, err = os.OpenFile(configFile, os.O_WRONLY, 0644)
	assert.NoError(t, err)

	testConfig := &Configuration{
		ProgramFolder:      DefaultConfig.ProgramFolder,
		ConfigFilename:     DefaultConfig.ConfigFilename,
		TaskFilename:       DefaultConfig.TaskFilename,
		DisplayProgressBar: false,
		UseHashLength:      uint8(32),
		SaveWithHash:       true,
		CopyNBytes:         int64(300),
		LogLevel:           uint8(1),
		DownloadFolder:     "/home/User/Download",
	}
	configBytes, err := toml.Marshal(testConfig)
	assert.NoError(t, err)
	_, err = file.Write(configBytes)
	assert.NoError(t, err)

	config, err = LoadConfig(configFile)
	assert.NoError(t, err)
	assert.Equal(t, testConfig, config)

	assert.NoError(t, os.Remove(configFile))
}
