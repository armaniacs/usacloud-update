package config

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	// Check default values
	if config.Zone != "tk1v" {
		t.Errorf("Expected default zone 'tk1v', got '%s'", config.Zone)
	}

	if config.APIEndpoint == "" {
		t.Error("Default API endpoint should not be empty")
	}

	if !strings.HasPrefix(config.APIEndpoint, "https://") {
		t.Error("Default API endpoint should use HTTPS")
	}

	if config.Enabled {
		t.Error("Default config should have Enabled=false")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Timeout)
	}

	if config.Debug {
		t.Error("Default config should have Debug=false")
	}

	if config.DryRun {
		t.Error("Default config should have DryRun=false")
	}

	if !config.Interactive {
		t.Error("Default config should have Interactive=true")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"SAKURACLOUD_ACCESS_TOKEN",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET",
		"SAKURACLOUD_ZONE",
		"SAKURACLOUD_API_URL",
		"USACLOUD_UPDATE_SANDBOX_ENABLED",
		"USACLOUD_UPDATE_DEBUG",
		"USACLOUD_UPDATE_DRY_RUN",
		"USACLOUD_UPDATE_INTERACTIVE",
		"USACLOUD_UPDATE_TIMEOUT",
	}

	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Test with default values (no environment variables set)
	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not return error: %v", err)
	}

	if config.AccessToken != "" {
		t.Error("AccessToken should be empty when not set")
	}

	if config.Zone != "tk1v" {
		t.Error("Should use default zone when not set")
	}

	// Test with environment variables set
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "test-token")
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "test-secret")
	os.Setenv("SAKURACLOUD_ZONE", "is1a")
	os.Setenv("USACLOUD_UPDATE_SANDBOX_ENABLED", "true")
	os.Setenv("USACLOUD_UPDATE_DEBUG", "true")
	os.Setenv("USACLOUD_UPDATE_DRY_RUN", "true")
	os.Setenv("USACLOUD_UPDATE_INTERACTIVE", "false")
	os.Setenv("USACLOUD_UPDATE_TIMEOUT", "60")

	config, err = LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not return error: %v", err)
	}

	if config.AccessToken != "test-token" {
		t.Errorf("Expected AccessToken 'test-token', got '%s'", config.AccessToken)
	}

	if config.AccessTokenSecret != "test-secret" {
		t.Errorf("Expected AccessTokenSecret 'test-secret', got '%s'", config.AccessTokenSecret)
	}

	if config.Zone != "is1a" {
		t.Errorf("Expected Zone 'is1a', got '%s'", config.Zone)
	}

	if !config.Enabled {
		t.Error("Expected Enabled=true")
	}

	if !config.Debug {
		t.Error("Expected Debug=true")
	}

	if !config.DryRun {
		t.Error("Expected DryRun=true")
	}

	if config.Interactive {
		t.Error("Expected Interactive=false")
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout 60s, got %v", config.Timeout)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test with non-existing environment variable
	result = getEnv("NON_EXISTING_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}

	// Test with empty environment variable
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = getEnv("EMPTY_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default' for empty env var, got '%s'", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	testCases := []struct {
		envValue     string
		defaultValue bool
		expected     bool
		description  string
	}{
		{"true", false, true, "true string"},
		{"false", true, false, "false string"},
		{"True", false, true, "True with capital"},
		{"FALSE", true, false, "FALSE all caps"},
		{"1", false, true, "1 as true"},
		{"0", true, false, "0 as false"},
		{"yes", false, false, "invalid value should use default"},
		{"", true, true, "empty value should use default"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.envValue != "" {
				os.Setenv("TEST_BOOL_VAR", tc.envValue)
				defer os.Unsetenv("TEST_BOOL_VAR")
			} else {
				os.Unsetenv("TEST_BOOL_VAR")
			}

			result := getEnvBool("TEST_BOOL_VAR", tc.defaultValue)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for env value '%s'", tc.expected, result, tc.envValue)
			}
		})
	}

	// Test with non-existing environment variable
	result := getEnvBool("NON_EXISTING_BOOL_VAR", true)
	if !result {
		t.Error("Expected true for non-existing env var with default true")
	}

	result = getEnvBool("NON_EXISTING_BOOL_VAR", false)
	if result {
		t.Error("Expected false for non-existing env var with default false")
	}
}

func TestSandboxConfigValidate(t *testing.T) {
	// Test disabled config (should always be valid)
	config := &SandboxConfig{
		Enabled: false,
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Disabled config should be valid, got error: %v", err)
	}

	// Test enabled config without credentials (should be invalid)
	config = &SandboxConfig{
		Enabled:           true,
		AccessToken:       "",
		AccessTokenSecret: "",
	}

	if err := config.Validate(); err == nil {
		t.Error("Enabled config without credentials should be invalid")
	}

	// Test enabled config with only token (should be invalid)
	config = &SandboxConfig{
		Enabled:           true,
		AccessToken:       "test-token",
		AccessTokenSecret: "",
	}

	if err := config.Validate(); err == nil {
		t.Error("Enabled config without secret should be invalid")
	}

	// Test enabled config with only secret (should be invalid)
	config = &SandboxConfig{
		Enabled:           true,
		AccessToken:       "",
		AccessTokenSecret: "test-secret",
	}

	if err := config.Validate(); err == nil {
		t.Error("Enabled config without token should be invalid")
	}

	// Test enabled config with both credentials (should be valid)
	config = &SandboxConfig{
		Enabled:           true,
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Enabled config with credentials should be valid, got error: %v", err)
	}
}

func TestLoadFromEnvTimeout(t *testing.T) {
	// Save original environment
	originalTimeout := os.Getenv("USACLOUD_UPDATE_TIMEOUT")
	defer func() {
		if originalTimeout != "" {
			os.Setenv("USACLOUD_UPDATE_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("USACLOUD_UPDATE_TIMEOUT")
		}
	}()

	testCases := []struct {
		envValue        string
		expectedTimeout time.Duration
		description     string
	}{
		{"60", 60 * time.Second, "valid integer"},
		{"120", 120 * time.Second, "different valid integer"},
		{"invalid", 30 * time.Second, "invalid string should use default"},
		{"", 30 * time.Second, "empty string should use default"},
		{"-1", -1 * time.Second, "negative number"},
		{"0", 0, "zero"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.envValue != "" {
				os.Setenv("USACLOUD_UPDATE_TIMEOUT", tc.envValue)
			} else {
				os.Unsetenv("USACLOUD_UPDATE_TIMEOUT")
			}

			config, err := LoadFromEnv()
			if err != nil {
				t.Fatalf("LoadFromEnv should not return error: %v", err)
			}

			if config.Timeout != tc.expectedTimeout {
				t.Errorf("Expected timeout %v, got %v for env value '%s'",
					tc.expectedTimeout, config.Timeout, tc.envValue)
			}
		})
	}
}

func TestLoadFromEnvCustomAPIEndpoint(t *testing.T) {
	originalEndpoint := os.Getenv("SAKURACLOUD_API_URL")
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("SAKURACLOUD_API_URL", originalEndpoint)
		} else {
			os.Unsetenv("SAKURACLOUD_API_URL")
		}
	}()

	// Test with custom API endpoint
	customEndpoint := "https://api.custom.sakura.io/cloud/1.1/"
	os.Setenv("SAKURACLOUD_API_URL", customEndpoint)

	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not return error: %v", err)
	}

	if config.APIEndpoint != customEndpoint {
		t.Errorf("Expected API endpoint '%s', got '%s'", customEndpoint, config.APIEndpoint)
	}

	// Test without custom API endpoint (should use default)
	os.Unsetenv("SAKURACLOUD_API_URL")

	config, err = LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not return error: %v", err)
	}

	defaultConfig := DefaultConfig()
	if config.APIEndpoint != defaultConfig.APIEndpoint {
		t.Errorf("Expected default API endpoint '%s', got '%s'",
			defaultConfig.APIEndpoint, config.APIEndpoint)
	}
}

func TestConfigFieldTypes(t *testing.T) {
	config := DefaultConfig()

	// Test that all fields have expected types
	var _ string = config.AccessToken
	var _ string = config.AccessTokenSecret
	var _ string = config.Zone
	var _ string = config.APIEndpoint
	var _ bool = config.Enabled
	var _ time.Duration = config.Timeout
	var _ bool = config.Debug
	var _ bool = config.DryRun
	var _ bool = config.Interactive
}

func TestEnvironmentVariableNames(t *testing.T) {
	// Test that expected environment variable names are used
	expectedVars := []string{
		"SAKURACLOUD_ACCESS_TOKEN",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET",
		"SAKURACLOUD_ZONE",
		"SAKURACLOUD_API_URL",
		"USACLOUD_UPDATE_SANDBOX_ENABLED",
		"USACLOUD_UPDATE_DEBUG",
		"USACLOUD_UPDATE_DRY_RUN",
		"USACLOUD_UPDATE_INTERACTIVE",
		"USACLOUD_UPDATE_TIMEOUT",
	}

	// Save original environment
	originalEnv := make(map[string]string)
	for _, key := range expectedVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set test values for all expected variables
	for i, key := range expectedVars {
		testValue := "test_value_" + strconv.Itoa(i)
		if key == "USACLOUD_UPDATE_SANDBOX_ENABLED" ||
			key == "USACLOUD_UPDATE_DEBUG" ||
			key == "USACLOUD_UPDATE_DRY_RUN" ||
			key == "USACLOUD_UPDATE_INTERACTIVE" {
			testValue = "true"
		}
		if key == "USACLOUD_UPDATE_TIMEOUT" {
			testValue = "60"
		}
		os.Setenv(key, testValue)
	}

	// Load configuration and verify all variables were read
	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv should not return error: %v", err)
	}

	// Verify that config is not using all default values
	defaultConfig := DefaultConfig()

	if config.AccessToken == defaultConfig.AccessToken {
		t.Error("SAKURACLOUD_ACCESS_TOKEN was not read")
	}

	if config.Enabled == defaultConfig.Enabled {
		t.Error("USACLOUD_UPDATE_SANDBOX_ENABLED was not read")
	}

	if config.Timeout == defaultConfig.Timeout {
		t.Error("USACLOUD_UPDATE_TIMEOUT was not read")
	}
}
