package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	t.Run("Load valid config file", func(t *testing.T) {
		filename := filepath.Join(dir, "../../config.yaml")
		assert.FileExists(t, filename)
		assert.NotEmpty(t, filename)

		cfg, err := LoadConfig(filename)
		assert.Nil(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("Load invalid config filepath", func(t *testing.T) {
		filename := filepath.Join(dir, "config.yaml")
		assert.NotNil(t, filename)
		assert.NoFileExists(t, filename)

		cfg, err := LoadConfig(filename)
		assert.NotNil(t, err)
		assert.Nil(t, cfg)
	})
}

func TestGetConfig(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	filename := filepath.Join(dir, "../../config.yaml")
	assert.FileExists(t, filename)
	assert.NotEmpty(t, filename)

	cfg, err := LoadConfig(filename)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	t.Run("GetConfig with valid configCache", func(t *testing.T) {
		gCfg, err := GetConfig()
		assert.Nil(t, err)
		assert.NotNil(t, gCfg)
		assert.Equal(t, gCfg.App.Name, cfg.App.Name)
	})

	t.Run("GetConfig with invalid configCache", func(t *testing.T) {
		configCache = nil
		gCfg, err := GetConfig()
		assert.NotNil(t, err)
		assert.Nil(t, gCfg)
	})
}
