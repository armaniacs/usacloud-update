package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewProfileManager(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	if manager == nil {
		t.Fatal("ProfileManager should not be nil")
	}

	if manager.configDir != tmpDir {
		t.Errorf("Expected config dir %s, got %s", tmpDir, manager.configDir)
	}
}

func TestCreateProfile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	opts := ProfileCreateOptions{
		Name:        "test-profile",
		Description: "Test profile",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "test-token",
			ConfigKeyAccessTokenSecret: "test-secret",
			ConfigKeyZone:              "tk1v",
		},
		Tags:       []string{"test"},
		SetDefault: true,
	}

	profile, err := manager.CreateProfile(opts)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	if profile.Name != opts.Name {
		t.Errorf("Expected name %s, got %s", opts.Name, profile.Name)
	}

	if profile.Environment != opts.Environment {
		t.Errorf("Expected environment %s, got %s", opts.Environment, profile.Environment)
	}

	if !profile.IsDefault {
		t.Error("Profile should be default")
	}

	if len(profile.Config) != 3 {
		t.Errorf("Expected 3 config items, got %d", len(profile.Config))
	}
}

func TestCreateProfileWithDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	opts := ProfileCreateOptions{
		Name:        "duplicate-profile",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "test-token",
			ConfigKeyAccessTokenSecret: "test-secret",
		},
	}

	// Create first profile
	_, err = manager.CreateProfile(opts)
	if err != nil {
		t.Fatalf("Failed to create first profile: %v", err)
	}

	// Try to create second profile with same name
	_, err = manager.CreateProfile(opts)
	if err == nil {
		t.Error("Expected error for duplicate profile name")
	}
}

func TestCreateProfileFromParent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create parent profile
	parentOpts := ProfileCreateOptions{
		Name:        "parent-profile",
		Environment: EnvironmentProduction,
		Config: map[string]string{
			ConfigKeyAccessToken:       "parent-token",
			ConfigKeyAccessTokenSecret: "parent-secret",
			ConfigKeyZone:              "tk1v",
		},
		Tags: []string{"parent"},
	}

	parent, err := manager.CreateProfile(parentOpts)
	if err != nil {
		t.Fatalf("Failed to create parent profile: %v", err)
	}

	// Create child profile
	overrides := map[string]string{
		ConfigKeyAccessToken: "child-token",
		ConfigKeyDryRun:      "true",
	}

	child, err := manager.CreateProfileFromParent("child-profile", "Child profile", parent.ID, overrides)
	if err != nil {
		t.Fatalf("Failed to create child profile: %v", err)
	}

	if child.ParentID != parent.ID {
		t.Errorf("Expected parent ID %s, got %s", parent.ID, child.ParentID)
	}

	if child.Environment != parent.Environment {
		t.Errorf("Expected environment %s, got %s", parent.Environment, child.Environment)
	}

	// Check inherited config
	if child.Config[ConfigKeyAccessToken] != "child-token" {
		t.Errorf("Expected overridden token, got %s", child.Config[ConfigKeyAccessToken])
	}

	if child.Config[ConfigKeyAccessTokenSecret] != "parent-secret" {
		t.Errorf("Expected inherited secret, got %s", child.Config[ConfigKeyAccessTokenSecret])
	}

	if child.Config[ConfigKeyDryRun] != "true" {
		t.Errorf("Expected added config, got %s", child.Config[ConfigKeyDryRun])
	}
}

func TestUpdateProfile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create profile
	opts := ProfileCreateOptions{
		Name:        "update-test",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "test-token",
			ConfigKeyAccessTokenSecret: "test-secret",
		},
	}

	profile, err := manager.CreateProfile(opts)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Update profile
	newDescription := "Updated description"
	updateOpts := ProfileUpdateOptions{
		Description: &newDescription,
		Config: map[string]string{
			ConfigKeyZone: "is1a",
		},
	}

	err = manager.UpdateProfile(profile.ID, updateOpts)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Retrieve updated profile
	updated, err := manager.GetProfile(profile.ID)
	if err != nil {
		t.Fatalf("Failed to get updated profile: %v", err)
	}

	if updated.Description != newDescription {
		t.Errorf("Expected description %s, got %s", newDescription, updated.Description)
	}

	if updated.Config[ConfigKeyZone] != "is1a" {
		t.Errorf("Expected zone is1a, got %s", updated.Config[ConfigKeyZone])
	}
}

func TestSwitchProfile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create two profiles
	opts1 := ProfileCreateOptions{
		Name:        "profile1",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "token1",
			ConfigKeyAccessTokenSecret: "secret1",
			ConfigKeyZone:              "tk1v",
		},
	}

	profile1, err := manager.CreateProfile(opts1)
	if err != nil {
		t.Fatalf("Failed to create profile1: %v", err)
	}

	opts2 := ProfileCreateOptions{
		Name:        "profile2",
		Environment: EnvironmentProduction,
		Config: map[string]string{
			ConfigKeyAccessToken:       "token2",
			ConfigKeyAccessTokenSecret: "secret2",
			ConfigKeyZone:              "is1a",
		},
	}

	profile2, err := manager.CreateProfile(opts2)
	if err != nil {
		t.Fatalf("Failed to create profile2: %v", err)
	}

	// Switch to profile2
	err = manager.SwitchProfile(profile2.ID)
	if err != nil {
		t.Fatalf("Failed to switch profile: %v", err)
	}

	// Check active profile
	active := manager.GetActiveProfile()
	if active == nil {
		t.Fatal("Active profile should not be nil")
	}

	if active.ID != profile2.ID {
		t.Errorf("Expected active profile %s, got %s", profile2.ID, active.ID)
	}

	// Check environment variables
	if os.Getenv(ConfigKeyAccessToken) != "token2" {
		t.Errorf("Expected env var %s, got %s", "token2", os.Getenv(ConfigKeyAccessToken))
	}

	// Check config file
	configFile := filepath.Join(tmpDir, "current.conf")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file should exist after switching profile")
	}

	// Switch back to profile1
	err = manager.SwitchProfile(profile1.ID)
	if err != nil {
		t.Fatalf("Failed to switch back to profile1: %v", err)
	}

	// Check updated last used time
	updated1, _ := manager.GetProfile(profile1.ID)
	updated2, _ := manager.GetProfile(profile2.ID)

	if updated1.LastUsedAt.Before(updated2.LastUsedAt) {
		t.Error("Profile1 should have more recent LastUsedAt")
	}
}

func TestDeleteProfile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create profile
	opts := ProfileCreateOptions{
		Name:        "delete-test",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "test-token",
			ConfigKeyAccessTokenSecret: "test-secret",
		},
	}

	profile, err := manager.CreateProfile(opts)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Delete profile
	err = manager.DeleteProfile(profile.ID)
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Try to get deleted profile
	_, err = manager.GetProfile(profile.ID)
	if err == nil {
		t.Error("Expected error when getting deleted profile")
	}
}

func TestDeleteProfileWithChildren(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create parent profile
	parentOpts := ProfileCreateOptions{
		Name:        "parent",
		Environment: EnvironmentProduction,
		Config: map[string]string{
			ConfigKeyAccessToken:       "parent-token",
			ConfigKeyAccessTokenSecret: "parent-secret",
		},
	}

	parent, err := manager.CreateProfile(parentOpts)
	if err != nil {
		t.Fatalf("Failed to create parent profile: %v", err)
	}

	// Create child profile
	_, err = manager.CreateProfileFromParent("child", "Child profile", parent.ID, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to create child profile: %v", err)
	}

	// Try to delete parent (should fail)
	err = manager.DeleteProfile(parent.ID)
	if err == nil {
		t.Error("Expected error when deleting profile with children")
	}
}

func TestListProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create multiple profiles
	profiles := []ProfileCreateOptions{
		{
			Name:        "dev-profile",
			Environment: EnvironmentDevelopment,
			Tags:        []string{"dev", "test"},
			Config: map[string]string{
				ConfigKeyAccessToken:       "dev-token",
				ConfigKeyAccessTokenSecret: "dev-secret",
			},
		},
		{
			Name:        "prod-profile",
			Environment: EnvironmentProduction,
			Tags:        []string{"prod"},
			Config: map[string]string{
				ConfigKeyAccessToken:       "prod-token",
				ConfigKeyAccessTokenSecret: "prod-secret",
			},
		},
		{
			Name:        "staging-profile",
			Environment: EnvironmentStaging,
			Tags:        []string{"staging"},
			Config: map[string]string{
				ConfigKeyAccessToken:       "staging-token",
				ConfigKeyAccessTokenSecret: "staging-secret",
			},
		},
	}

	for _, opts := range profiles {
		_, err := manager.CreateProfile(opts)
		if err != nil {
			t.Fatalf("Failed to create profile %s: %v", opts.Name, err)
		}
	}

	// Test listing all profiles
	allProfiles := manager.ListProfiles()
	if len(allProfiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(allProfiles))
	}

	// Test filtering by environment
	devProfiles := manager.ListProfiles(ProfileListOptions{
		Environment: EnvironmentDevelopment,
	})
	if len(devProfiles) != 1 {
		t.Errorf("Expected 1 dev profile, got %d", len(devProfiles))
	}

	// Test filtering by tags
	testProfiles := manager.ListProfiles(ProfileListOptions{
		Tags: []string{"test"},
	})
	if len(testProfiles) != 1 {
		t.Errorf("Expected 1 profile with test tag, got %d", len(testProfiles))
	}

	// Test sorting by name
	sortedProfiles := manager.ListProfiles(ProfileListOptions{
		SortBy: "name",
	})
	if sortedProfiles[0].Name != "dev-profile" {
		t.Errorf("Expected first profile to be dev-profile, got %s", sortedProfiles[0].Name)
	}
}

func TestExportImportProfile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	// Create profile
	opts := ProfileCreateOptions{
		Name:        "export-test",
		Description: "Export test profile",
		Environment: EnvironmentDevelopment,
		Config: map[string]string{
			ConfigKeyAccessToken:       "export-token",
			ConfigKeyAccessTokenSecret: "export-secret",
			ConfigKeyZone:              "tk1v",
		},
		Tags: []string{"export", "test"},
	}

	original, err := manager.CreateProfile(opts)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Export profile
	exportFile := filepath.Join(tmpDir, "exported-profile.yaml")
	err = manager.ExportProfile(original.ID, exportFile)
	if err != nil {
		t.Fatalf("Failed to export profile: %v", err)
	}

	// Check export file exists
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Error("Export file should exist")
	}

	// Import profile
	imported, err := manager.ImportProfile(exportFile)
	if err != nil {
		t.Fatalf("Failed to import profile: %v", err)
	}

	// Verify imported profile (name might have suffix for duplicates)
	if !strings.HasPrefix(imported.Name, original.Name) {
		t.Errorf("Expected name to start with %s, got %s", original.Name, imported.Name)
	}

	if imported.Environment != original.Environment {
		t.Errorf("Expected environment %s, got %s", original.Environment, imported.Environment)
	}

	if len(imported.Config) != len(original.Config) {
		t.Errorf("Expected %d config items, got %d", len(original.Config), len(imported.Config))
	}

	for key, value := range original.Config {
		if imported.Config[key] != value {
			t.Errorf("Expected config %s=%s, got %s", key, value, imported.Config[key])
		}
	}

	// Verify different ID
	if imported.ID == original.ID {
		t.Error("Imported profile should have different ID")
	}

	// Verify not default
	if imported.IsDefault {
		t.Error("Imported profile should not be default")
	}
}

func TestProfileValidation(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewProfileManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create profile manager: %v", err)
	}

	testCases := []struct {
		name        string
		opts        ProfileCreateOptions
		expectError bool
	}{
		{
			name: "valid profile",
			opts: ProfileCreateOptions{
				Name:        "valid-profile",
				Environment: EnvironmentDevelopment,
				Config: map[string]string{
					ConfigKeyAccessToken:       "valid-token",
					ConfigKeyAccessTokenSecret: "valid-secret",
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			opts: ProfileCreateOptions{
				Environment: EnvironmentDevelopment,
				Config: map[string]string{
					ConfigKeyAccessToken:       "valid-token",
					ConfigKeyAccessTokenSecret: "valid-secret",
				},
			},
			expectError: true,
		},
		{
			name: "missing environment",
			opts: ProfileCreateOptions{
				Name: "no-env-profile",
				Config: map[string]string{
					ConfigKeyAccessToken:       "valid-token",
					ConfigKeyAccessTokenSecret: "valid-secret",
				},
			},
			expectError: true,
		},
		{
			name: "invalid environment",
			opts: ProfileCreateOptions{
				Name:        "invalid-env-profile",
				Environment: "invalid-env",
				Config: map[string]string{
					ConfigKeyAccessToken:       "valid-token",
					ConfigKeyAccessTokenSecret: "valid-secret",
				},
			},
			expectError: true,
		},
		{
			name: "missing required config",
			opts: ProfileCreateOptions{
				Name:        "missing-config-profile",
				Environment: EnvironmentDevelopment,
				Config:      map[string]string{},
			},
			expectError: true,
		},
		{
			name: "invalid zone",
			opts: ProfileCreateOptions{
				Name:        "invalid-zone-profile",
				Environment: EnvironmentDevelopment,
				Config: map[string]string{
					ConfigKeyAccessToken:       "valid-token",
					ConfigKeyAccessTokenSecret: "valid-secret",
					ConfigKeyZone:              "invalid-zone",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := manager.CreateProfile(tc.opts)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSensitiveKeyMasking(t *testing.T) {
	testCases := []struct {
		key      string
		expected bool
	}{
		{"SAKURACLOUD_ACCESS_TOKEN", true},
		{"SAKURACLOUD_ACCESS_TOKEN_SECRET", true},
		{"API_KEY", true},
		{"PASSWORD", true},
		{"SAKURACLOUD_ZONE", false},
		{"DEBUG", false},
		{"TIMEOUT", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := IsSensitiveKey(tc.key)
			if result != tc.expected {
				t.Errorf("Expected %t for key %s, got %t", tc.expected, tc.key, result)
			}
		})
	}
}

func TestMaskValue(t *testing.T) {
	testCases := []struct {
		value    string
		expected string
	}{
		{"short", "****"},
		{"12345678", "****"},
		{"1234567890", "1234**7890"},
		{"very-long-secret-token-value", "very************************alue"},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			result := MaskValue(tc.value)
			if result != tc.expected {
				t.Errorf("Expected %s for value %s, got %s", tc.expected, tc.value, result)
			}
		})
	}
}
