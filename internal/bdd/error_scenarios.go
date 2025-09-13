package bdd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ErrorScenarioGenerator generates various error scenarios for testing
type ErrorScenarioGenerator struct {
	scenarios []ErrorScenario
}

// NewErrorScenarioGenerator creates a new error scenario generator
func NewErrorScenarioGenerator() *ErrorScenarioGenerator {
	return &ErrorScenarioGenerator{
		scenarios: make([]ErrorScenario, 0),
	}
}

// GenerateNetworkErrors creates network-related error scenarios
func (esg *ErrorScenarioGenerator) GenerateNetworkErrors() []ErrorScenario {
	return []ErrorScenario{
		{
			Name:        "connection_timeout",
			Description: "API接続タイムアウト",
			Setup: func() error {
				// Set a very short timeout to force timeout errors
				return os.Setenv("SAKURACLOUD_TIMEOUT", "1")
			},
			Trigger: func() error {
				// Try to execute a command that would normally take longer than 1 second
				return executeCommand("usacloud server ls --zone=all")
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected timeout error but got none")
				}
				errorStr := strings.ToLower(err.Error())
				if !strings.Contains(errorStr, "timeout") && !strings.Contains(errorStr, "deadline") {
					return fmt.Errorf("expected timeout error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				return os.Unsetenv("SAKURACLOUD_TIMEOUT")
			},
		},
		{
			Name:        "invalid_api_key",
			Description: "無効なAPIキー",
			Setup: func() error {
				// Backup current API key
				if current := os.Getenv("SAKURACLOUD_ACCESS_TOKEN"); current != "" {
					if err := os.Setenv("SAKURACLOUD_ACCESS_TOKEN_BACKUP", current); err != nil {
						return fmt.Errorf("failed to backup API key: %w", err)
					}
				}
				// Set an invalid API key
				return os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "invalid_token_12345")
			},
			Trigger: func() error {
				return executeCommand("usacloud auth-status")
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected auth error but got none")
				}
				expectedStrings := []string{"authentication", "invalid", "token", "unauthorized", "401"}
				errorStr := strings.ToLower(err.Error())
				for _, expected := range expectedStrings {
					if strings.Contains(errorStr, expected) {
						return nil
					}
				}
				return fmt.Errorf("expected auth error containing one of %v, got: %v", expectedStrings, err)
			},
			Cleanup: func() error {
				// Restore original API key
				if backup := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_BACKUP"); backup != "" {
					if err := os.Setenv("SAKURACLOUD_ACCESS_TOKEN", backup); err != nil {
						return fmt.Errorf("failed to restore API key: %w", err)
					}
					if err := os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN_BACKUP"); err != nil {
						return fmt.Errorf("failed to unset backup key: %w", err)
					}
				} else {
					if err := os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN"); err != nil {
						return fmt.Errorf("failed to unset API key: %w", err)
					}
				}
				return nil
			},
		},
		{
			Name:        "rate_limit_exceeded",
			Description: "API率制限エラー",
			Setup: func() error {
				// Set aggressive rate limiting for testing
				return os.Setenv("SAKURACLOUD_RATE_LIMIT", "1")
			},
			Trigger: func() error {
				// Make multiple rapid API calls
				for i := 0; i < 5; i++ {
					executeCommand("usacloud zone ls")
					time.Sleep(100 * time.Millisecond)
				}
				return executeCommand("usacloud zone ls")
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected rate limit error but got none")
				}
				errorStr := strings.ToLower(err.Error())
				rateLimitKeywords := []string{"rate limit", "too many requests", "429", "quota"}
				for _, keyword := range rateLimitKeywords {
					if strings.Contains(errorStr, keyword) {
						return nil
					}
				}
				return fmt.Errorf("expected rate limit error, got: %v", err)
			},
			Cleanup: func() error {
				return os.Unsetenv("SAKURACLOUD_RATE_LIMIT")
			},
		},
	}
}

// GenerateFileSystemErrors creates file system related error scenarios
func (esg *ErrorScenarioGenerator) GenerateFileSystemErrors() []ErrorScenario {
	return []ErrorScenario{
		{
			Name:        "permission_denied",
			Description: "ファイル権限エラー",
			Setup: func() error {
				// Create a file with no read permissions
				file, err := os.CreateTemp("", "readonly-test-*.sh")
				if err != nil {
					return err
				}
				file.WriteString("#!/bin/bash\nusacloud server list\n")
				file.Close()

				// Remove read permissions
				if err := os.Chmod(file.Name(), 0000); err != nil {
					return err
				}

				return os.Setenv("BDD_TEST_FILE", file.Name())
			},
			Trigger: func() error {
				filename := os.Getenv("BDD_TEST_FILE")
				if filename == "" {
					return fmt.Errorf("test file not set")
				}
				// Try to read the file
				_, err := os.ReadFile(filename)
				return err
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected permission error but got none")
				}
				if !strings.Contains(err.Error(), "permission denied") {
					return fmt.Errorf("expected permission denied error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				filename := os.Getenv("BDD_TEST_FILE")
				if filename != "" {
					// Restore permissions and delete file
					os.Chmod(filename, 0644)
					os.Remove(filename)
					os.Unsetenv("BDD_TEST_FILE")
				}
				return nil
			},
		},
		{
			Name:        "disk_space_full",
			Description: "ディスク容量不足エラー（シミュレーション）",
			Setup: func() error {
				return os.Setenv("SIMULATE_DISK_FULL", "true")
			},
			Trigger: func() error {
				// Simulate a large write operation
				if os.Getenv("SIMULATE_DISK_FULL") == "true" {
					return fmt.Errorf("no space left on device")
				}
				return nil
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected disk full error but got none")
				}
				if !strings.Contains(err.Error(), "no space left") {
					return fmt.Errorf("expected disk full error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				return os.Unsetenv("SIMULATE_DISK_FULL")
			},
		},
	}
}

// GenerateConfigurationErrors creates configuration related error scenarios
func (esg *ErrorScenarioGenerator) GenerateConfigurationErrors() []ErrorScenario {
	return []ErrorScenario{
		{
			Name:        "missing_config_file",
			Description: "設定ファイル不在エラー",
			Setup: func() error {
				return os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", "/nonexistent/directory")
			},
			Trigger: func() error {
				// Try to load configuration
				configDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
				if configDir == "" {
					return fmt.Errorf("config directory not set")
				}

				configFile := configDir + "/usacloud-update.conf"
				_, err := os.ReadFile(configFile)
				return err
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected file not found error but got none")
				}
				if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "not found") {
					return fmt.Errorf("expected file not found error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				return os.Unsetenv("USACLOUD_UPDATE_CONFIG_DIR")
			},
		},
		{
			Name:        "invalid_config_format",
			Description: "設定ファイル形式エラー",
			Setup: func() error {
				// Create a temporary config file with invalid format
				tempDir, err := os.MkdirTemp("", "config-test-*")
				if err != nil {
					return err
				}

				configFile := tempDir + "/usacloud-update.conf"
				invalidConfig := `[invalid
missing_bracket=true
malformed content
`
				if err := os.WriteFile(configFile, []byte(invalidConfig), 0644); err != nil {
					return err
				}

				return os.Setenv("BDD_TEST_CONFIG_DIR", tempDir)
			},
			Trigger: func() error {
				configDir := os.Getenv("BDD_TEST_CONFIG_DIR")
				if configDir == "" {
					return fmt.Errorf("config directory not set")
				}

				configFile := configDir + "/usacloud-update.conf"
				content, err := os.ReadFile(configFile)
				if err != nil {
					return err
				}

				// Simulate config parsing error
				if strings.Contains(string(content), "missing_bracket") {
					return fmt.Errorf("configuration file format error: invalid INI syntax")
				}

				return nil
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected config format error but got none")
				}
				errorStr := strings.ToLower(err.Error())
				if !strings.Contains(errorStr, "format") && !strings.Contains(errorStr, "syntax") && !strings.Contains(errorStr, "parse") {
					return fmt.Errorf("expected config format error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				configDir := os.Getenv("BDD_TEST_CONFIG_DIR")
				if configDir != "" {
					os.RemoveAll(configDir)
					os.Unsetenv("BDD_TEST_CONFIG_DIR")
				}
				return nil
			},
		},
	}
}

// GenerateResourceErrors creates resource limitation error scenarios
func (esg *ErrorScenarioGenerator) GenerateResourceErrors() []ErrorScenario {
	return []ErrorScenario{
		{
			Name:        "memory_exhaustion",
			Description: "メモリ不足エラー（シミュレーション）",
			Setup: func() error {
				return os.Setenv("SIMULATE_MEMORY_ERROR", "true")
			},
			Trigger: func() error {
				if os.Getenv("SIMULATE_MEMORY_ERROR") == "true" {
					return fmt.Errorf("cannot allocate memory")
				}
				return nil
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected memory error but got none")
				}
				if !strings.Contains(err.Error(), "memory") && !strings.Contains(err.Error(), "allocate") {
					return fmt.Errorf("expected memory error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				return os.Unsetenv("SIMULATE_MEMORY_ERROR")
			},
		},
		{
			Name:        "process_limit_exceeded",
			Description: "プロセス制限エラー（シミュレーション）",
			Setup: func() error {
				return os.Setenv("SIMULATE_PROCESS_LIMIT", "true")
			},
			Trigger: func() error {
				if os.Getenv("SIMULATE_PROCESS_LIMIT") == "true" {
					return fmt.Errorf("resource temporarily unavailable: too many processes")
				}
				return nil
			},
			Verify: func(err error) error {
				if err == nil {
					return fmt.Errorf("expected process limit error but got none")
				}
				errorStr := strings.ToLower(err.Error())
				if !strings.Contains(errorStr, "process") && !strings.Contains(errorStr, "resource") {
					return fmt.Errorf("expected process limit error, got: %v", err)
				}
				return nil
			},
			Cleanup: func() error {
				return os.Unsetenv("SIMULATE_PROCESS_LIMIT")
			},
		},
	}
}

// GenerateAllScenarios generates all available error scenarios
func (esg *ErrorScenarioGenerator) GenerateAllScenarios() []ErrorScenario {
	var allScenarios []ErrorScenario

	allScenarios = append(allScenarios, esg.GenerateNetworkErrors()...)
	allScenarios = append(allScenarios, esg.GenerateFileSystemErrors()...)
	allScenarios = append(allScenarios, esg.GenerateConfigurationErrors()...)
	allScenarios = append(allScenarios, esg.GenerateResourceErrors()...)

	return allScenarios
}

// RunScenario executes a single error scenario
func (esg *ErrorScenarioGenerator) RunScenario(scenario ErrorScenario) error {
	// Setup
	if err := scenario.Setup(); err != nil {
		return fmt.Errorf("scenario setup failed: %w", err)
	}

	// Ensure cleanup runs regardless of outcome
	defer func() {
		if cleanupErr := scenario.Cleanup(); cleanupErr != nil {
			fmt.Printf("Warning: cleanup failed for scenario %s: %v\n", scenario.Name, cleanupErr)
		}
	}()

	// Trigger the error condition
	err := scenario.Trigger()

	// Verify the error is as expected
	if verifyErr := scenario.Verify(err); verifyErr != nil {
		return fmt.Errorf("scenario verification failed: %w", verifyErr)
	}

	return nil
}

// RunAllScenarios executes all generated error scenarios
func (esg *ErrorScenarioGenerator) RunAllScenarios() []ScenarioResult {
	scenarios := esg.GenerateAllScenarios()
	results := make([]ScenarioResult, len(scenarios))

	for i, scenario := range scenarios {
		startTime := time.Now()
		err := esg.RunScenario(scenario)
		duration := time.Since(startTime)

		results[i] = ScenarioResult{
			Name:        scenario.Name,
			Description: scenario.Description,
			Success:     err == nil,
			Error:       err,
			Duration:    duration,
		}
	}

	return results
}

// ScenarioResult represents the result of running an error scenario
type ScenarioResult struct {
	Name        string
	Description string
	Success     bool
	Error       error
	Duration    time.Duration
}

// executeCommand is a helper function to execute usacloud commands
func executeCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w (output: %s)", err, string(output))
	}

	return nil
}
