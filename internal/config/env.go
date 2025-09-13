package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SandboxConfig holds all configuration for sandbox functionality
type SandboxConfig struct {
	// Sakura Cloud API settings
	AccessToken       string
	AccessTokenSecret string
	Zone              string
	APIEndpoint       string

	// Application settings
	Enabled     bool
	Timeout     time.Duration
	Debug       bool
	DryRun      bool
	Interactive bool
}

// DefaultConfig returns the default sandbox configuration
func DefaultConfig() *SandboxConfig {
	return &SandboxConfig{
		Zone:        "tk1v",
		APIEndpoint: "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/",
		Enabled:     false,
		Timeout:     30 * time.Second,
		Debug:       false,
		DryRun:      false,
		Interactive: true,
	}
}

// LoadConfig loads configuration with the following priority:
// 1. Configuration file (custom path if provided, otherwise default location)
// 2. Environment variables (legacy .env file support)
// 3. Interactive creation if no configuration exists
func LoadConfig(customConfigPath ...string) (*SandboxConfig, error) {
	// Try to load from configuration file first
	var config *SandboxConfig
	var err error

	if len(customConfigPath) > 0 && customConfigPath[0] != "" {
		// Load from custom path
		config, err = LoadFromFileWithPath(customConfigPath[0])
	} else {
		// Load from default path
		config, err = LoadFromFile()
	}

	if err != nil {
		if !IsConfigNotFound(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}

		// Config file not found - check for .env file (only if using default path)
		if len(customConfigPath) == 0 || customConfigPath[0] == "" {
			if _, envErr := os.Stat(".env"); envErr == nil {
				// .env exists, offer migration
				return MigrateFromEnv()
			}

			// No configuration exists - create interactively
			return CreateInteractiveConfig()
		} else {
			// Custom config path specified but file not found
			return nil, fmt.Errorf("設定ファイルが見つかりません: %s", customConfigPath[0])
		}
	}

	return config, nil
}

// LoadFromEnv loads configuration from environment variables (legacy support)
func LoadFromEnv() (*SandboxConfig, error) {
	config := DefaultConfig()

	// Try to load .env file if it exists
	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Load Sakura Cloud API settings
	config.AccessToken = getEnv("SAKURACLOUD_ACCESS_TOKEN", "")
	config.AccessTokenSecret = getEnv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "")
	config.Zone = getEnv("SAKURACLOUD_ZONE", config.Zone)
	config.APIEndpoint = getEnv("SAKURACLOUD_API_URL", config.APIEndpoint)

	// Load application settings
	config.Enabled = getEnvBool("USACLOUD_UPDATE_SANDBOX_ENABLED", config.Enabled)
	config.Debug = getEnvBool("USACLOUD_UPDATE_DEBUG", config.Debug)
	config.DryRun = getEnvBool("USACLOUD_UPDATE_DRY_RUN", config.DryRun)
	config.Interactive = getEnvBool("USACLOUD_UPDATE_INTERACTIVE", config.Interactive)

	// Load timeout setting
	if timeoutStr := getEnv("USACLOUD_UPDATE_TIMEOUT", ""); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}
	}

	return config, nil
}

// Validate checks if the configuration is valid for sandbox operations
func (c *SandboxConfig) Validate() error {
	if !c.Enabled {
		return nil // Validation not required if sandbox is disabled
	}

	var errors []string

	if c.AccessToken == "" {
		errors = append(errors, "SAKURACLOUD_ACCESS_TOKEN is required")
	}

	if c.AccessTokenSecret == "" {
		errors = append(errors, "SAKURACLOUD_ACCESS_TOKEN_SECRET is required")
	}

	if c.Zone != "tk1v" {
		errors = append(errors, "SAKURACLOUD_ZONE must be 'tk1v' for sandbox operations")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// GetUsacloudEnv returns environment variables formatted for usacloud command execution
func (c *SandboxConfig) GetUsacloudEnv() []string {
	env := os.Environ()

	// Set usacloud-specific environment variables
	env = setEnvVar(env, "SAKURACLOUD_ACCESS_TOKEN", c.AccessToken)
	env = setEnvVar(env, "SAKURACLOUD_ACCESS_TOKEN_SECRET", c.AccessTokenSecret)
	env = setEnvVar(env, "SAKURACLOUD_ZONE", c.Zone)

	return env
}

// PrintGuide prints a helpful guide for setting up environment variables
func (c *SandboxConfig) PrintGuide() {
	fmt.Fprintf(os.Stderr, `
🔧 サンドボックス機能を使用するには以下の環境変数が必要です:

必須設定:
  SAKURACLOUD_ACCESS_TOKEN      - さくらのクラウド APIアクセストークン
  SAKURACLOUD_ACCESS_TOKEN_SECRET - さくらのクラウド APIシークレット

現在の設定:
  SAKURACLOUD_ZONE             = %s
  SAKURACLOUD_API_URL          = %s
  USACLOUD_UPDATE_SANDBOX_ENABLED = %t
  USACLOUD_UPDATE_DEBUG        = %t
  USACLOUD_UPDATE_DRY_RUN      = %t

設定方法:
【推奨】新しい設定ファイル方式:
1. usacloud-update.conf.sample を参考にしてください
2. ~/.config/usacloud-update/usacloud-update.conf を作成してください
3. 初回実行時は対話的に設定を作成することも可能です

【レガシー】環境変数方式:
1. 環境変数を直接設定してください:
   export SAKURACLOUD_ACCESS_TOKEN="your-token"
   export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-secret"
2. 必要に応じて他の設定を調整してください

APIキー取得方法:
  https://manual.sakura.ad.jp/cloud/api/apikey.html

`, c.Zone, c.APIEndpoint, c.Enabled, c.Debug, c.DryRun)
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Set environment variable if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// setEnvVar sets an environment variable in the provided slice
func setEnvVar(env []string, key, value string) []string {
	prefix := key + "="

	// Update existing variable
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}

	// Add new variable
	return append(env, prefix+value)
}
