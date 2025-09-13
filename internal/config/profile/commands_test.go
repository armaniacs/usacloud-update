package profile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestProfileCommand_ListProfiles(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profiles
	profile1, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Profile 1",
		Description: "Test profile 1",
		Environment: "test",
		Config: map[string]string{
			"key1":                            "value1",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token-1",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret-1",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	profile2, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Profile 2",
		Description: "Test profile 2",
		Environment: "production",
		Config: map[string]string{
			"key2":                            "value2",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token-2",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret-2",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Set one as default
	err = manager.SetDefault(profile1.ID)
	if err != nil {
		t.Fatalf("SetDefault() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Capture output
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	cmd := &cobra.Command{}
	err = pc.ListProfiles(cmd, []string{})

	// Restore stdout and get output
	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("ListProfiles() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Profile 1") {
		t.Errorf("Expected output to contain 'Profile 1', got: %s", output)
	}
	if !strings.Contains(output, "Profile 2") {
		t.Errorf("Expected output to contain 'Profile 2', got: %s", output)
	}

	// Use variables to avoid unused variable error
	_ = profile2
}

func TestProfileCommand_ShowProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profile
	profile, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Test Profile",
		Description: "Test description",
		Environment: "test",
		Config: map[string]string{
			"API_TOKEN":                       "secret_token",
			"ZONE":                            "tk1v",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Capture output
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	cmd := &cobra.Command{}
	err = pc.ShowProfile(cmd, []string{profile.ID})

	// Restore stdout and get output
	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("ShowProfile() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Profile") {
		t.Errorf("Expected output to contain profile name, got: %s", output)
	}
	if !strings.Contains(output, "Test description") {
		t.Errorf("Expected output to contain profile description, got: %s", output)
	}
	// Should show non-sensitive info
	if !strings.Contains(output, "tk1v") {
		t.Errorf("Expected non-sensitive data to be shown, got: %s", output)
	}
}

func TestProfileCommand_ShowProfileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	cmd := &cobra.Command{}
	err = pc.ShowProfile(cmd, []string{"non-existent"})

	if err == nil {
		t.Errorf("Expected error for non-existent profile")
	}
}

func TestProfileCommand_ShowProfileNoArgs(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	cmd := &cobra.Command{}
	err = pc.ShowProfile(cmd, []string{})

	if err == nil {
		t.Errorf("Expected error when no profile ID provided")
	}
	if !strings.Contains(err.Error(), "指定") {
		t.Errorf("Expected error to mention required argument, got: %v", err)
	}
}

func TestProfileCommand_CreateProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Set up flags
	cmd := &cobra.Command{}
	cmd.Flags().String("name", "Test Profile", "Profile name")
	cmd.Flags().String("description", "Test description", "Profile description")
	cmd.Flags().String("environment", "test", "Environment")
	cmd.Flags().StringSlice("config", []string{
		"key=value",
		"SAKURACLOUD_ACCESS_TOKEN=test-token",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET=test-secret",
	}, "Configuration")

	err = pc.CreateProfile(cmd, []string{"Test Profile"})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	// Verify profile was created
	profiles := manager.ListProfiles(ProfileListOptions{})
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Name != "Test Profile" {
		t.Errorf("Expected profile name 'Test Profile', got %s", profiles[0].Name)
	}
}

func TestProfileCommand_DeleteProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profile
	profile, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Test Profile",
		Description: "Test description",
		Environment: "test",
		Config: map[string]string{
			"key":                             "value",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", true, "Force deletion")
	err = pc.DeleteProfile(cmd, []string{profile.ID})
	if err != nil {
		t.Fatalf("DeleteProfile() failed: %v", err)
	}

	// Verify profile was deleted
	profiles := manager.ListProfiles(ProfileListOptions{})
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles after deletion, got %d", len(profiles))
	}
}

func TestProfileCommand_SwitchProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profile
	profile, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Test Profile",
		Description: "Test description",
		Environment: "test",
		Config: map[string]string{
			"key":                             "value",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	cmd := &cobra.Command{}
	err = pc.SwitchProfile(cmd, []string{profile.ID})
	if err != nil {
		t.Fatalf("SwitchProfile() failed: %v", err)
	}

	// Verify profile is active
	activeProfile := manager.GetActiveProfile()
	if activeProfile == nil {
		t.Errorf("Expected active profile to be set")
	} else if activeProfile.ID != profile.ID {
		t.Errorf("Expected active profile ID %s, got %s", profile.ID, activeProfile.ID)
	}
}

func TestProfileCommand_ExportProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profile
	profile, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Test Profile",
		Description: "Test description",
		Environment: "test",
		Config: map[string]string{
			"key":                             "value",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Create temporary export file
	exportFile := filepath.Join(tempDir, "export.yaml")

	cmd := &cobra.Command{}
	cmd.Flags().String("output", exportFile, "Output file")

	err = pc.ExportProfile(cmd, []string{profile.ID})
	if err != nil {
		t.Fatalf("ExportProfile() failed: %v", err)
	}

	// Verify export file was created
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Errorf("Export file was not created")
	}
}

func TestProfileCommand_ImportProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Create test import data with required SAKURACLOUD_ACCESS_TOKEN
	importData := `
id: imported-profile
name: Imported Profile
description: Imported from file
environment: test
config:
  SAKURACLOUD_ACCESS_TOKEN: test-token
  SAKURACLOUD_ACCESS_TOKEN_SECRET: test-secret
  imported: value
created_at: 2023-01-01T00:00:00Z
updated_at: 2023-01-01T00:00:00Z
tags: []
`

	importFile := filepath.Join(tempDir, "import.yaml")
	err = os.WriteFile(importFile, []byte(importData), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	cmd := &cobra.Command{}
	err = pc.ImportProfile(cmd, []string{importFile})
	if err != nil {
		t.Fatalf("ImportProfile() failed: %v", err)
	}

	// Verify profile was imported
	profiles := manager.ListProfiles(ProfileListOptions{})
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile after import, got %d", len(profiles))
	}

	if profiles[0].Name != "Imported Profile" {
		t.Errorf("Expected imported profile name 'Imported Profile', got %s", profiles[0].Name)
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key       string
		sensitive bool
	}{
		{"API_TOKEN", true},
		{"ACCESS_TOKEN", true},
		{"SECRET", true},
		{"PASSWORD", true},
		{"KEY", true},
		{"ZONE", false},
		{"REGION", false},
		{"NAME", false},
		{"api_token", true}, // case insensitive
		{"My_Password", true},
		{"NORMAL_CONFIG", false},
	}

	for _, test := range tests {
		result := isSensitiveKey(test.key)
		if result != test.sensitive {
			t.Errorf("Expected isSensitiveKey(%s) = %v, got %v", test.key, test.sensitive, result)
		}
	}
}

func TestMaskValueInCommands(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "a**"},
		{"abcd", "a***"},
		{"abcde", "a***e"},
		{"abcdef", "a****f"},
		{"very_long_secret_value", "v*****************e"},
	}

	for _, test := range tests {
		result := maskValue(test.value)
		if result != test.expected {
			t.Errorf("Expected maskValue(%s) = %s, got %s", test.value, test.expected, result)
		}
	}
}

func TestProfileCommand_UpdateProfile(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// Create test profile
	profile, err := manager.CreateProfile(ProfileCreateOptions{
		Name:        "Test Profile",
		Description: "Test description",
		Environment: "test",
		Config: map[string]string{
			"key1":                            "value1",
			"key2":                            "value2",
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Set up flags for update
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("config", []string{
		"key1=updated_value1",
		"key3=new_value3",
	}, "Configuration updates")
	cmd.Flags().String("name", "Updated Profile", "Updated name")
	cmd.Flags().String("description", "Updated description", "Updated description")

	err = pc.UpdateProfile(cmd, []string{profile.ID})
	if err != nil {
		t.Fatalf("UpdateProfile() failed: %v", err)
	}

	// Verify profile was updated
	updatedProfile, err := manager.GetProfile(profile.ID)
	if err != nil {
		t.Fatalf("GetProfile() failed: %v", err)
	}

	if updatedProfile.Name != "Updated Profile" {
		t.Errorf("Expected updated name 'Updated Profile', got %s", updatedProfile.Name)
	}

	if updatedProfile.Config["key1"] != "updated_value1" {
		t.Errorf("Expected updated config key1 'updated_value1', got %s", updatedProfile.Config["key1"])
	}

	if updatedProfile.Config["key3"] != "new_value3" {
		t.Errorf("Expected new config key3 'new_value3', got %s", updatedProfile.Config["key3"])
	}
}

func TestProfileCommand_ListTemplates(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templateManager := NewTemplateManager()
	pc := NewProfileCommand(manager, templateManager)

	// Capture output
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	err = pc.ListTemplates(cmd, []string{})

	// Restore stdout and get output
	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	output := buf.String()
	templates := manager.GetBuiltinTemplates()

	for _, template := range templates {
		if !strings.Contains(output, template.Name) {
			t.Errorf("Expected output to contain template name '%s', got: %s", template.Name, output)
		}
	}
}
