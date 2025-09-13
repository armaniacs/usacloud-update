package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigPath(t *testing.T) {
	// Save original env vars
	originalConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalAppData := os.Getenv("APPDATA")
	defer func() {
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", originalConfigDir)
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		os.Setenv("APPDATA", originalAppData)
	}()

	t.Run("CustomConfigDir", func(t *testing.T) {
		// Create temp directory for testing
		tempDir, err := os.MkdirTemp("", "usacloud-config-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set custom config dir
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", tempDir)

		configPath, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() failed: %v", err)
		}

		expected := filepath.Join(tempDir, "usacloud-update.conf")
		if configPath != expected {
			t.Errorf("ConfigPath() = %s, expected %s", configPath, expected)
		}
	})

	t.Run("CustomConfigDir_RelativePath", func(t *testing.T) {
		// Set relative path (should fail)
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", "relative/path")

		_, err := ConfigPath()
		if err == nil {
			t.Error("ConfigPath() should fail with relative path")
		}
		if !strings.Contains(err.Error(), "directory path must be absolute") {
			t.Errorf("Expected 'directory path must be absolute' error, got: %v", err)
		}
	})

	t.Run("CustomConfigDir_NonExistent", func(t *testing.T) {
		// Set non-existent path (should create it)
		tempDir, err := os.MkdirTemp("", "usacloud-config-test-parent")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		nonExistentDir := filepath.Join(tempDir, "non-existent")
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", nonExistentDir)

		configPath, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() failed: %v", err)
		}

		// Check that directory was created
		if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
			t.Error("Directory should have been created")
		}

		expected := filepath.Join(nonExistentDir, "usacloud-update.conf")
		if configPath != expected {
			t.Errorf("ConfigPath() = %s, expected %s", configPath, expected)
		}
	})

	t.Run("DefaultPath_Unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix test on Windows")
		}

		// Clear custom config dir
		os.Unsetenv("USACLOUD_UPDATE_CONFIG_DIR")
		os.Unsetenv("XDG_CONFIG_HOME")

		configPath, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() failed: %v", err)
		}

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "usacloud-update", "usacloud-update.conf")
		if configPath != expected {
			t.Errorf("ConfigPath() = %s, expected %s", configPath, expected)
		}
	})

	t.Run("XDG_CONFIG_HOME", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping XDG test on Windows")
		}

		// Clear custom config dir and set XDG_CONFIG_HOME
		os.Unsetenv("USACLOUD_UPDATE_CONFIG_DIR")

		tempDir, err := os.MkdirTemp("", "xdg-config-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		os.Setenv("XDG_CONFIG_HOME", tempDir)

		configPath, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() failed: %v", err)
		}

		expected := filepath.Join(tempDir, "usacloud-update", "usacloud-update.conf")
		if configPath != expected {
			t.Errorf("ConfigPath() = %s, expected %s", configPath, expected)
		}
	})
}

func TestValidateConfigDir(t *testing.T) {
	t.Run("ValidAbsolutePath", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "validate-config-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		err = validateConfigDir(tempDir)
		if err != nil {
			t.Errorf("validateConfigDir() failed for valid path: %v", err)
		}
	})

	t.Run("RelativePath", func(t *testing.T) {
		err := validateConfigDir("relative/path")
		if err == nil {
			t.Error("validateConfigDir() should fail with relative path")
		}
		if !strings.Contains(err.Error(), "directory path must be absolute") {
			t.Errorf("Expected 'directory path must be absolute' error, got: %v", err)
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "validate-config-parent")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		nonExistentDir := filepath.Join(tempDir, "non-existent")

		err = validateConfigDir(nonExistentDir)
		if err != nil {
			t.Errorf("validateConfigDir() should create directory: %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
			t.Error("Directory should have been created")
		}
	})

	t.Run("WritePermissionTest", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "validate-write-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		err = validateConfigDir(tempDir)
		if err != nil {
			t.Errorf("validateConfigDir() failed: %v", err)
		}

		// Verify test file was cleaned up
		testFile := filepath.Join(tempDir, ".usacloud-update-test")
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("Test file should have been cleaned up")
		}
	})
}

func TestLoadFromFile(t *testing.T) {
	t.Run("FileNotFound", func(t *testing.T) {
		// Create temp directory without config file
		tempDir, err := os.MkdirTemp("", "load-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set custom config dir to temp directory
		originalConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
		defer os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", originalConfigDir)
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", tempDir)

		_, err = LoadFromFile()
		if err == nil {
			t.Error("LoadFromFile() should fail when file doesn't exist")
		}
		if !IsConfigNotFound(err) {
			t.Errorf("Expected ConfigNotFoundError, got: %v", err)
		}
	})

	t.Run("ValidConfigFile", func(t *testing.T) {
		// Create temp directory with config file
		tempDir, err := os.MkdirTemp("", "load-valid-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		configFile := filepath.Join(tempDir, "usacloud-update.conf")
		configContent := `[sakura-cloud]
access_token = "test-token"
access_token_secret = "test-secret"
zone = "tk1v"
api_endpoint = "https://test.example.com"

[sandbox]
enabled = true
debug = true
dry_run = false
interactive = false
timeout = 60
`
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Set custom config dir
		originalConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
		defer os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", originalConfigDir)
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", tempDir)

		config, err := LoadFromFile()
		if err != nil {
			t.Fatalf("LoadFromFile() failed: %v", err)
		}

		// Verify loaded values
		if config.AccessToken != "test-token" {
			t.Errorf("AccessToken = %s, expected test-token", config.AccessToken)
		}
		if config.AccessTokenSecret != "test-secret" {
			t.Errorf("AccessTokenSecret = %s, expected test-secret", config.AccessTokenSecret)
		}
		if config.Zone != "tk1v" {
			t.Errorf("Zone = %s, expected tk1v", config.Zone)
		}
		if config.APIEndpoint != "https://test.example.com" {
			t.Errorf("APIEndpoint = %s, expected https://test.example.com", config.APIEndpoint)
		}
		if !config.Enabled {
			t.Error("Enabled should be true")
		}
		if !config.Debug {
			t.Error("Debug should be true")
		}
		if config.DryRun {
			t.Error("DryRun should be false")
		}
		if config.Interactive {
			t.Error("Interactive should be false")
		}
		if config.Timeout.Seconds() != 60 {
			t.Errorf("Timeout = %v, expected 60s", config.Timeout)
		}
	})

	t.Run("InvalidSyntax", func(t *testing.T) {
		// Create temp directory with invalid config file
		tempDir, err := os.MkdirTemp("", "load-invalid-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		configFile := filepath.Join(tempDir, "usacloud-update.conf")
		invalidContent := `[sakura-cloud]
invalid-line-without-equals
access_token = "test-token"
`
		err = os.WriteFile(configFile, []byte(invalidContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Set custom config dir
		originalConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
		defer os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", originalConfigDir)
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", tempDir)

		_, err = LoadFromFile()
		if err == nil {
			t.Error("LoadFromFile() should fail with invalid syntax")
		}
		if !strings.Contains(err.Error(), "invalid syntax") {
			t.Errorf("Expected 'invalid syntax' error, got: %v", err)
		}
	})
}

func TestSaveToFile(t *testing.T) {
	t.Run("SaveAndLoad", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "save-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set custom config dir
		originalConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
		defer os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", originalConfigDir)
		os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", tempDir)

		// Create config
		config := DefaultConfig()
		config.AccessToken = "save-test-token"
		config.AccessTokenSecret = "save-test-secret"
		config.Enabled = true
		config.Debug = true

		// Save config
		err = config.SaveToFile()
		if err != nil {
			t.Fatalf("SaveToFile() failed: %v", err)
		}

		// Verify file exists with correct permissions
		configFile := filepath.Join(tempDir, "usacloud-update.conf")
		info, err := os.Stat(configFile)
		if err != nil {
			t.Fatalf("Config file not created: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("Config file permissions = %v, expected 0600", info.Mode().Perm())
		}

		// Load and verify
		loadedConfig, err := LoadFromFile()
		if err != nil {
			t.Fatalf("LoadFromFile() failed: %v", err)
		}

		if loadedConfig.AccessToken != config.AccessToken {
			t.Errorf("AccessToken = %s, expected %s", loadedConfig.AccessToken, config.AccessToken)
		}
		if loadedConfig.AccessTokenSecret != config.AccessTokenSecret {
			t.Errorf("AccessTokenSecret = %s, expected %s", loadedConfig.AccessTokenSecret, config.AccessTokenSecret)
		}
		if loadedConfig.Enabled != config.Enabled {
			t.Errorf("Enabled = %t, expected %t", loadedConfig.Enabled, config.Enabled)
		}
	})
}

func TestIsConfigNotFound(t *testing.T) {
	t.Run("ConfigNotFoundError", func(t *testing.T) {
		err := &ConfigNotFoundError{Path: "/test/path"}
		if !IsConfigNotFound(err) {
			t.Error("IsConfigNotFound() should return true for ConfigNotFoundError")
		}
	})

	t.Run("OtherError", func(t *testing.T) {
		err := os.ErrNotExist
		if IsConfigNotFound(err) {
			t.Error("IsConfigNotFound() should return false for other errors")
		}
	})
}
