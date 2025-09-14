package config

import (
	"fmt"
	"os"
	"strings"
)

// UsacloudEnvVars represents the environment variables used by usacloud
type UsacloudEnvVars struct {
	AccessToken       string
	AccessTokenSecret string
	Zone              string // Optional
}

// EnvDetector handles environment variable detection and validation
type EnvDetector struct {
	RequiredVars []string
	OptionalVars []string
}

// NewEnvDetector creates a new environment variable detector
func NewEnvDetector() *EnvDetector {
	return &EnvDetector{
		RequiredVars: []string{
			"SAKURACLOUD_ACCESS_TOKEN",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET",
		},
		OptionalVars: []string{
			"SAKURACLOUD_ZONE",
		},
	}
}

// DetectUsacloudEnvVars detects and returns usacloud environment variables
func (e *EnvDetector) DetectUsacloudEnvVars() (*UsacloudEnvVars, bool) {
	envVars := &UsacloudEnvVars{
		AccessToken:       os.Getenv("SAKURACLOUD_ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET"),
		Zone:              os.Getenv("SAKURACLOUD_ZONE"),
	}

	// Check if required variables are present
	if envVars.AccessToken == "" || envVars.AccessTokenSecret == "" {
		return nil, false
	}

	return envVars, true
}

// ValidateEnvVarValues validates the format and content of environment variables
func (e *EnvDetector) ValidateEnvVarValues(envVars *UsacloudEnvVars) error {
	if envVars == nil {
		return fmt.Errorf("environment variables not detected")
	}

	// Validate access token format (basic validation)
	if len(envVars.AccessToken) < 10 {
		return fmt.Errorf("SAKURACLOUD_ACCESS_TOKEN appears to be too short (minimum 10 characters)")
	}

	// Validate access token secret format (basic validation)
	if len(envVars.AccessTokenSecret) < 20 {
		return fmt.Errorf("SAKURACLOUD_ACCESS_TOKEN_SECRET appears to be too short (minimum 20 characters)")
	}

	// Validate zone format if provided
	if envVars.Zone != "" {
		validZones := []string{"tk1v", "tk1a", "is1a", "is1b", "tk1b"}
		isValidZone := false
		for _, zone := range validZones {
			if envVars.Zone == zone {
				isValidZone = true
				break
			}
		}
		if !isValidZone {
			return fmt.Errorf("SAKURACLOUD_ZONE '%s' is not a valid zone. Valid zones are: %s",
				envVars.Zone, strings.Join(validZones, ", "))
		}
	}

	return nil
}

// FormatEnvVarsDisplay creates a formatted display string for detected environment variables
func (e *EnvDetector) FormatEnvVarsDisplay(envVars *UsacloudEnvVars) string {
	if envVars == nil {
		return "環境変数が検出されませんでした"
	}

	var display strings.Builder
	display.WriteString("🔍 検出された環境変数:\n")
	display.WriteString(fmt.Sprintf("  ✓ SAKURACLOUD_ACCESS_TOKEN: %s...%s\n",
		envVars.AccessToken[:8], envVars.AccessToken[len(envVars.AccessToken)-4:]))
	display.WriteString(fmt.Sprintf("  ✓ SAKURACLOUD_ACCESS_TOKEN_SECRET: %s...%s\n",
		envVars.AccessTokenSecret[:8], envVars.AccessTokenSecret[len(envVars.AccessTokenSecret)-4:]))

	if envVars.Zone != "" {
		display.WriteString(fmt.Sprintf("  ✓ SAKURACLOUD_ZONE: %s\n", envVars.Zone))
	} else {
		display.WriteString("  - SAKURACLOUD_ZONE: (未設定 - tk1vを使用)\n")
	}

	return display.String()
}

// GetDefaultZone returns the default zone if none is specified in environment variables
func (e *EnvDetector) GetDefaultZone() string {
	return "tk1v"
}

// GenerateConfigContent generates the configuration file content from environment variables
func (e *EnvDetector) GenerateConfigContent(envVars *UsacloudEnvVars) string {
	zone := envVars.Zone
	if zone == "" {
		zone = e.GetDefaultZone()
	}

	configContent := fmt.Sprintf(`[sakuracloud]
access_token = %s
access_token_secret = %s
zone = %s

# Generated from environment variables
# SAKURACLOUD_ACCESS_TOKEN
# SAKURACLOUD_ACCESS_TOKEN_SECRET
# SAKURACLOUD_ZONE (optional, defaults to tk1v)
`, envVars.AccessToken, envVars.AccessTokenSecret, zone)

	return configContent
}
