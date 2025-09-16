package config

import (
	"os"
	"path/filepath"
)

// GetConfigFilePath returns the configuration file path, considering custom config directory
func GetConfigFilePath() (string, error) {
	configDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = homeDir
	}

	return filepath.Join(configDir, ".config", "usacloud-update", "usacloud-update.conf"), nil
}

// EnsureConfigDirectory creates the config directory if it doesn't exist
func EnsureConfigDirectory(configPath string) error {
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0700)
}
