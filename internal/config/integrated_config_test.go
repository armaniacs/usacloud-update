package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewIntegratedConfig(t *testing.T) {
	config := NewIntegratedConfig()

	if config == nil {
		t.Fatal("Expected config to be created, got nil")
	}

	if config.General.Version != "1.9.0" {
		t.Errorf("Expected version 1.9.0, got %s", config.General.Version)
	}

	if !config.General.ColorOutput {
		t.Error("Expected color output to be enabled by default")
	}

	if config.General.Language != "ja" {
		t.Errorf("Expected language to be 'ja', got %s", config.General.Language)
	}

	if config.General.Profile != "default" {
		t.Errorf("Expected profile to be 'default', got %s", config.General.Profile)
	}

	if !config.Validation.EnableValidation {
		t.Error("Expected validation to be enabled by default")
	}

	if config.Validation.MaxSuggestions != 5 {
		t.Errorf("Expected max suggestions to be 5, got %d", config.Validation.MaxSuggestions)
	}

	if config.Profiles == nil {
		t.Error("Expected profiles map to be initialized")
	}

	if config.Environments == nil {
		t.Error("Expected environments map to be initialized")
	}
}

func TestLoadIntegratedConfig_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.conf")

	config, err := LoadIntegratedConfig(configPath)
	if err != nil {
		t.Fatalf("Expected config to be created when file doesn't exist, got error: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be created, got nil")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestCreateDefaultProfiles(t *testing.T) {
	config := NewIntegratedConfig()
	config.createDefaultProfiles()

	expectedProfiles := []string{"default", "beginner", "expert", "ci"}
	for _, profileName := range expectedProfiles {
		profile, exists := config.Profiles[profileName]
		if !exists {
			t.Errorf("Expected profile '%s' to exist", profileName)
			continue
		}

		if profile.Name != profileName {
			t.Errorf("Expected profile name to be '%s', got '%s'", profileName, profile.Name)
		}

		if profile.Description == "" {
			t.Errorf("Expected profile '%s' to have description", profileName)
		}

		if profile.CreatedAt.IsZero() {
			t.Errorf("Expected profile '%s' to have creation time", profileName)
		}
	}

	beginnerProfile := config.Profiles["beginner"]
	if beginnerProfile.BasedOn != "default" {
		t.Errorf("Expected beginner profile to be based on 'default', got '%s'", beginnerProfile.BasedOn)
	}

	if beginnerProfile.Overrides["verbose"] != "true" {
		t.Error("Expected beginner profile to have verbose=true override")
	}

	expertProfile := config.Profiles["expert"]
	if expertProfile.Overrides["strict_mode"] != "true" {
		t.Error("Expected expert profile to have strict_mode=true override")
	}
}

func TestCreateDefaultEnvironments(t *testing.T) {
	config := NewIntegratedConfig()
	config.createDefaultEnvironments()

	expectedEnvs := []string{"development", "production"}
	for _, envName := range expectedEnvs {
		env, exists := config.Environments[envName]
		if !exists {
			t.Errorf("Expected environment '%s' to exist", envName)
			continue
		}

		if env.Name != envName {
			t.Errorf("Expected environment name to be '%s', got '%s'", envName, env.Name)
		}

		if env.SakuraAPIEndpoint == "" {
			t.Errorf("Expected environment '%s' to have API endpoint", envName)
		}

		if env.TimeoutSeconds == 0 {
			t.Errorf("Expected environment '%s' to have timeout", envName)
		}
	}

	prodEnv := config.Environments["production"]
	if !prodEnv.StrictMode {
		t.Error("Expected production environment to have strict mode enabled")
	}

	devEnv := config.Environments["development"]
	if devEnv.StrictMode {
		t.Error("Expected development environment to have strict mode disabled")
	}
}

func TestApplyEnvironmentOverrides(t *testing.T) {
	config := NewIntegratedConfig()

	os.Setenv("USACLOUD_UPDATE_PROFILE", "expert")
	os.Setenv("USACLOUD_UPDATE_STRICT_MODE", "true")
	os.Setenv("USACLOUD_UPDATE_PARALLEL", "false")
	os.Setenv("USACLOUD_UPDATE_COLOR", "false")
	os.Setenv("USACLOUD_UPDATE_VERBOSE", "true")

	defer func() {
		os.Unsetenv("USACLOUD_UPDATE_PROFILE")
		os.Unsetenv("USACLOUD_UPDATE_STRICT_MODE")
		os.Unsetenv("USACLOUD_UPDATE_PARALLEL")
		os.Unsetenv("USACLOUD_UPDATE_COLOR")
		os.Unsetenv("USACLOUD_UPDATE_VERBOSE")
	}()

	config.applyEnvironmentOverrides()

	if config.General.Profile != "expert" {
		t.Errorf("Expected profile to be 'expert', got '%s'", config.General.Profile)
	}

	if !config.Validation.StrictMode {
		t.Error("Expected strict mode to be enabled")
	}

	if config.Performance.ParallelProcessing {
		t.Error("Expected parallel processing to be disabled")
	}

	if config.General.ColorOutput {
		t.Error("Expected color output to be disabled")
	}

	if !config.General.Verbose {
		t.Error("Expected verbose to be enabled")
	}
}

func TestApplyProfile(t *testing.T) {
	config := NewIntegratedConfig()
	config.createDefaultProfiles()

	err := config.applyProfile("beginner")
	if err != nil {
		t.Fatalf("Failed to apply beginner profile: %v", err)
	}

	if !config.General.InteractiveByDefault {
		t.Error("Expected interactive mode to be enabled for beginner profile")
	}

	if !config.General.Verbose {
		t.Error("Expected verbose to be enabled for beginner profile")
	}

	if config.HelpSystem.SkillLevel != "beginner" {
		t.Errorf("Expected skill level to be 'beginner', got '%s'", config.HelpSystem.SkillLevel)
	}

	if config.Validation.MaxSuggestions != 8 {
		t.Errorf("Expected max suggestions to be 8, got %d", config.Validation.MaxSuggestions)
	}
}

func TestApplyProfileWithBasedOn(t *testing.T) {
	config := NewIntegratedConfig()
	config.createDefaultProfiles()

	err := config.applyProfile("expert")
	if err != nil {
		t.Fatalf("Failed to apply expert profile: %v", err)
	}

	if !config.Validation.StrictMode {
		t.Error("Expected strict mode to be enabled for expert profile")
	}

	if config.Validation.MaxSuggestions != 3 {
		t.Errorf("Expected max suggestions to be 3, got %d", config.Validation.MaxSuggestions)
	}

	if !config.Performance.ParallelProcessing {
		t.Error("Expected parallel processing to be enabled for expert profile")
	}

	if config.Output.ShowProgress {
		t.Error("Expected progress display to be disabled for expert profile")
	}

	if config.Output.ReportLevel != "minimal" {
		t.Errorf("Expected report level to be 'minimal', got '%s'", config.Output.ReportLevel)
	}
}

func TestApplyProfileNotFound(t *testing.T) {
	config := NewIntegratedConfig()
	config.createDefaultProfiles()

	err := config.applyProfile("nonexistent")
	if err == nil {
		t.Error("Expected error when applying non-existent profile")
	}

	expectedError := "プロファイル 'nonexistent' が見つかりません"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.conf")

	originalConfig := NewIntegratedConfig()
	originalConfig.configPath = configPath
	originalConfig.createDefaultProfiles()
	originalConfig.General.ColorOutput = false
	originalConfig.General.Verbose = true
	originalConfig.Validation.StrictMode = true
	originalConfig.Validation.MaxSuggestions = 10

	err := originalConfig.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to get config file info: %v", err)
	}

	if info.Mode() != 0600 {
		t.Errorf("Expected config file to have mode 0600, got %o", info.Mode())
	}

	loadedConfig, err := LoadIntegratedConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.General.ColorOutput != originalConfig.General.ColorOutput {
		t.Error("ColorOutput not preserved after save/load")
	}

	if loadedConfig.General.Verbose != originalConfig.General.Verbose {
		t.Error("Verbose not preserved after save/load")
	}

	if loadedConfig.Validation.StrictMode != originalConfig.Validation.StrictMode {
		t.Error("StrictMode not preserved after save/load")
	}

	if loadedConfig.Validation.MaxSuggestions != originalConfig.Validation.MaxSuggestions {
		t.Error("MaxSuggestions not preserved after save/load")
	}
}

func TestConfigChangeNotification(t *testing.T) {
	config := NewIntegratedConfig()
	config.autoSave = false

	eventChan := make(chan ConfigChangeEvent, 1)
	config.watchers = append(config.watchers, eventChan)

	err := config.UpdateSetting("general", "color_output", false)
	if err != nil {
		t.Fatalf("Failed to update setting: %v", err)
	}

	select {
	case event := <-eventChan:
		if event.Section != "general" {
			t.Errorf("Expected section 'general', got '%s'", event.Section)
		}
		if event.Key != "color_output" {
			t.Errorf("Expected key 'color_output', got '%s'", event.Key)
		}
		if event.NewValue != false {
			t.Errorf("Expected new value false, got %v", event.NewValue)
		}
		if event.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	case <-time.After(time.Second):
		t.Error("Expected config change event to be sent")
	}
}

func TestConfigValidation(t *testing.T) {
	config := NewIntegratedConfig()

	config.Validation.MaxSuggestions = -1
	if config.Validation.MaxSuggestions < 0 {
		config.Validation.MaxSuggestions = 5
	}

	if config.Validation.MaxSuggestions != 5 {
		t.Error("Expected invalid max suggestions to be corrected")
	}

	config.Validation.MaxEditDistance = 100
	if config.Validation.MaxEditDistance > 10 {
		config.Validation.MaxEditDistance = 3
	}

	if config.Validation.MaxEditDistance != 3 {
		t.Error("Expected large edit distance to be corrected")
	}
}

func TestLoadProfilesFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.conf")

	configContent := `[general]
version = "1.9.0"
profile = "custom"

[profiles.default]
description = "標準設定"
based_on = ""

[profiles.custom]
description = "カスタムプロファイル"
based_on = "default"
verbose = true
strict_mode = false

[profiles.test]
description = "テスト用"
max_suggestions = 7
skill_level = "advanced"`

	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := LoadIntegratedConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Profiles) < 2 {
		t.Errorf("Expected at least 2 profiles, got %d", len(config.Profiles))
	}

	customProfile, exists := config.Profiles["custom"]
	if !exists {
		t.Error("Expected 'custom' profile to exist")
	} else {
		if customProfile.Description != "カスタムプロファイル" {
			t.Errorf("Expected description 'カスタムプロファイル', got '%s'", customProfile.Description)
		}
		if customProfile.BasedOn != "default" {
			t.Errorf("Expected based_on 'default', got '%s'", customProfile.BasedOn)
		}
		if customProfile.Overrides["verbose"] != "true" {
			t.Error("Expected verbose override to be 'true'")
		}
	}

	testProfile, exists := config.Profiles["test"]
	if !exists {
		t.Error("Expected 'test' profile to exist")
	} else {
		if testProfile.Overrides["max_suggestions"] != "7" {
			t.Error("Expected max_suggestions override to be '7'")
		}
		if testProfile.Overrides["skill_level"] != "advanced" {
			t.Error("Expected skill_level override to be 'advanced'")
		}
	}
}
