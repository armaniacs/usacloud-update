package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ConfigPath returns the path to the configuration file
func ConfigPath() (string, error) {
	var configDir string

	// Priority 1: USACLOUD_UPDATE_CONFIG_DIR environment variable (highest priority)
	if customDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR"); customDir != "" {
		// Validate the custom directory
		if err := validateConfigDir(customDir); err != nil {
			return "", fmt.Errorf("invalid USACLOUD_UPDATE_CONFIG_DIR: %w", err)
		}
		configDir = customDir
	} else {
		// Priority 2-3: OS-specific standard directories
		switch runtime.GOOS {
		case "windows":
			configDir = os.Getenv("APPDATA")
			if configDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}
				configDir = filepath.Join(home, "AppData", "Roaming")
			}
			configDir = filepath.Join(configDir, "usacloud-update")
		default:
			// Unix-like systems (Linux, macOS, etc.)
			configDir = os.Getenv("XDG_CONFIG_HOME")
			if configDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}
				configDir = filepath.Join(home, ".config")
			}
			configDir = filepath.Join(configDir, "usacloud-update")
		}
	}

	return filepath.Join(configDir, "usacloud-update.conf"), nil
}

// validateConfigDir validates the custom configuration directory
func validateConfigDir(dir string) error {
	// Check if the path is absolute
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("directory path must be absolute: %s", dir)
	}

	// Check if the directory exists, create if it doesn't
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}

	// Test write permissions by creating a temporary file
	testFile := filepath.Join(dir, ".usacloud-update-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// LoadFromFile loads configuration from the configuration file
func LoadFromFile() (*SandboxConfig, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	return LoadFromFileWithPath(configPath)
}

// LoadFromFileWithPath loads configuration from a specific file path
func LoadFromFileWithPath(configPath string) (*SandboxConfig, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, &ConfigNotFoundError{Path: configPath}
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Parse INI-style configuration
	scanner := bufio.NewScanner(file)
	lineNum := 0
	currentSection := ""

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Handle sections
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid syntax at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Apply configuration based on current section
		if err := applyConfigValue(config, currentSection, key, value); err != nil {
			return nil, fmt.Errorf("error at line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return config, nil
}

// applyConfigValue applies a configuration key-value pair to the config
func applyConfigValue(config *SandboxConfig, section, key, value string) error {
	switch section {
	case "", "sakura-cloud", "sakuracloud":
		// Sakura Cloud API settings
		switch strings.ToLower(key) {
		case "access_token", "accesstoken":
			config.AccessToken = value
		case "access_token_secret", "accesstokensecret":
			config.AccessTokenSecret = value
		case "zone":
			config.Zone = value
		case "api_endpoint", "apiendpoint", "api_url", "apiurl":
			config.APIEndpoint = value
		default:
			return fmt.Errorf("unknown sakura-cloud key: %s", key)
		}
	case "sandbox", "usacloud-update":
		// Application settings
		switch strings.ToLower(key) {
		case "enabled":
			if parsed, err := strconv.ParseBool(value); err == nil {
				config.Enabled = parsed
			} else {
				return fmt.Errorf("invalid boolean value for %s: %s", key, value)
			}
		case "debug":
			if parsed, err := strconv.ParseBool(value); err == nil {
				config.Debug = parsed
			} else {
				return fmt.Errorf("invalid boolean value for %s: %s", key, value)
			}
		case "dry_run", "dryrun":
			if parsed, err := strconv.ParseBool(value); err == nil {
				config.DryRun = parsed
			} else {
				return fmt.Errorf("invalid boolean value for %s: %s", key, value)
			}
		case "interactive":
			if parsed, err := strconv.ParseBool(value); err == nil {
				config.Interactive = parsed
			} else {
				return fmt.Errorf("invalid boolean value for %s: %s", key, value)
			}
		case "timeout":
			if timeout, err := strconv.Atoi(value); err == nil {
				config.Timeout = time.Duration(timeout) * time.Second
			} else {
				return fmt.Errorf("invalid timeout value: %s", value)
			}
		default:
			return fmt.Errorf("unknown sandbox key: %s", key)
		}
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
	return nil
}

// SaveToFile saves configuration to the configuration file
func (c *SandboxConfig) SaveToFile() error {
	configPath, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create configuration content
	content := c.generateConfigContent()

	// Write to file with appropriate permissions
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateConfigContent generates the configuration file content
func (c *SandboxConfig) generateConfigContent() string {
	var content strings.Builder

	content.WriteString("# usacloud-update Configuration File\n")
	content.WriteString("# This file contains settings for usacloud-update sandbox functionality\n")
	content.WriteString("\n")

	// Sakura Cloud API settings
	content.WriteString("[sakura-cloud]\n")
	content.WriteString("# Sakura Cloud API credentials (required for sandbox)\n")
	content.WriteString(fmt.Sprintf("access_token = \"%s\"\n", c.AccessToken))
	content.WriteString(fmt.Sprintf("access_token_secret = \"%s\"\n", c.AccessTokenSecret))
	content.WriteString("\n")
	content.WriteString("# Target zone for sandbox operations (must be tk1v)\n")
	content.WriteString(fmt.Sprintf("zone = \"%s\"\n", c.Zone))
	content.WriteString("\n")
	content.WriteString("# API endpoint for sandbox environment\n")
	content.WriteString(fmt.Sprintf("api_endpoint = \"%s\"\n", c.APIEndpoint))
	content.WriteString("\n")

	// Application settings
	content.WriteString("[sandbox]\n")
	content.WriteString("# Sandbox functionality settings\n")
	content.WriteString(fmt.Sprintf("enabled = %t\n", c.Enabled))
	content.WriteString(fmt.Sprintf("debug = %t\n", c.Debug))
	content.WriteString(fmt.Sprintf("dry_run = %t\n", c.DryRun))
	content.WriteString(fmt.Sprintf("interactive = %t\n", c.Interactive))
	content.WriteString(fmt.Sprintf("timeout = %d\n", int(c.Timeout.Seconds())))
	content.WriteString("\n")

	content.WriteString("# Configuration notes:\n")
	content.WriteString("# - This file contains sensitive API credentials\n")
	content.WriteString("# - File permissions are set to 600 (owner read/write only)\n")
	content.WriteString("# - For API key setup, visit: https://manual.sakura.ad.jp/cloud/api/apikey.html\n")

	return content.String()
}

// ConfigNotFoundError represents an error when configuration file is not found
type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("configuration file not found at %s", e.Path)
}

// IsConfigNotFound checks if an error is a ConfigNotFoundError
func IsConfigNotFound(err error) bool {
	_, ok := err.(*ConfigNotFoundError)
	return ok
}
