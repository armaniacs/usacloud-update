package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestFileStorage_Save(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "Test Description",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        []string{"test"},
	}

	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "profiles", "test-profile.yaml")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Profile file was not created")
	}

	// Verify file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read profile file: %v", err)
	}

	var savedProfile Profile
	err = yaml.Unmarshal(data, &savedProfile)
	if err != nil {
		t.Fatalf("Failed to unmarshal profile: %v", err)
	}

	if savedProfile.Name != profile.Name {
		t.Errorf("Expected name %s, got %s", profile.Name, savedProfile.Name)
	}
}

func TestFileStorage_Load(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create test profile file
	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "Test Description",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        []string{"test"},
	}

	// Save first
	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load back
	loadedProfile, err := storage.Load("test-profile")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loadedProfile.Name != profile.Name {
		t.Errorf("Expected name %s, got %s", profile.Name, loadedProfile.Name)
	}
	if loadedProfile.Description != profile.Description {
		t.Errorf("Expected description %s, got %s", profile.Description, loadedProfile.Description)
	}
	if loadedProfile.Config["key"] != "value" {
		t.Errorf("Expected config key value, got %s", loadedProfile.Config["key"])
	}
}

func TestFileStorage_LoadNotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	_, err := storage.Load("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestFileStorage_LoadAll(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create multiple profiles
	profiles := []*Profile{
		{
			ID:          "profile1",
			Name:        "Profile 1",
			Environment: "test",
			Config:      map[string]string{"key1": "value1"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "profile2",
			Name:        "Profile 2",
			Environment: "prod",
			Config:      map[string]string{"key2": "value2"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Save all profiles
	for _, profile := range profiles {
		err := storage.Save(profile)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// Load all
	loadedProfiles, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() failed: %v", err)
	}

	if len(loadedProfiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(loadedProfiles))
	}

	// Verify profiles exist
	if _, exists := loadedProfiles["profile1"]; !exists {
		t.Errorf("Profile1 not found in loaded profiles")
	}
	if _, exists := loadedProfiles["profile2"]; !exists {
		t.Errorf("Profile2 not found in loaded profiles")
	}
}

func TestFileStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save first
	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "profiles", "test-profile.yaml")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Profile file was not created")
	}

	// Delete
	err = storage.Delete("test-profile")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify file deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Profile file was not deleted")
	}
}

func TestFileStorage_SetGetDefault(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create test profile first
	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Set default
	err = storage.SetDefault("test-profile")
	if err != nil {
		t.Fatalf("SetDefault() failed: %v", err)
	}

	// Get default
	defaultProfile, err := storage.GetDefault()
	if err != nil {
		t.Fatalf("GetDefault() failed: %v", err)
	}

	if defaultProfile.ID != "test-profile" {
		t.Errorf("Expected default profile ID test-profile, got %s", defaultProfile.ID)
	}
}

func TestFileStorage_GetDefaultNotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	_, err := storage.GetDefault()
	if err == nil {
		t.Errorf("Expected error when no default profile exists")
	}
}

func TestFileStorage_BackupRestore(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save profile
	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Create backup
	backupPath, err := storage.Backup()
	if err != nil {
		t.Fatalf("Backup() failed: %v", err)
	}

	// Read backup file
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	// Verify backup contains profile
	var backupProfiles map[string]*Profile
	err = yaml.Unmarshal(backupData, &backupProfiles)
	if err != nil {
		t.Fatalf("Failed to unmarshal backup data: %v", err)
	}

	if len(backupProfiles) != 1 {
		t.Errorf("Expected 1 profile in backup, got %d", len(backupProfiles))
	}

	// Find the backed up profile
	var foundProfile *Profile
	for _, p := range backupProfiles {
		if p.Name == profile.Name {
			foundProfile = p
			break
		}
	}
	if foundProfile == nil {
		t.Errorf("Expected profile name %s in backup, but not found", profile.Name)
	}

	// Delete all profiles
	err = storage.Delete("test-profile")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Restore from backup
	err = storage.Restore(backupPath)
	if err != nil {
		t.Fatalf("Restore() failed: %v", err)
	}

	// Verify profile restored
	restoredProfile, err := storage.Load("test-profile")
	if err != nil {
		t.Fatalf("Load() after restore failed: %v", err)
	}

	if restoredProfile.Name != profile.Name {
		t.Errorf("Expected restored profile name %s, got %s", profile.Name, restoredProfile.Name)
	}
}

func TestFileStorage_Export(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Environment: "test",
		Config:      map[string]string{"key": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save profile
	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Export profile
	exportData, err := storage.Export("test-profile")
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	// Verify export data
	var exportedProfile Profile
	err = yaml.Unmarshal(exportData, &exportedProfile)
	if err != nil {
		t.Fatalf("Failed to unmarshal export data: %v", err)
	}

	if exportedProfile.Name != profile.Name {
		t.Errorf("Expected exported profile name %s, got %s", profile.Name, exportedProfile.Name)
	}
}

func TestFileStorage_Import(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "imported-profile",
		Name:        "Imported Profile",
		Environment: "test",
		Config:      map[string]string{"imported": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create export data
	exportData, err := yaml.Marshal(profile)
	if err != nil {
		t.Fatalf("Failed to marshal profile: %v", err)
	}

	// Import profile
	importedProfile, err := storage.Import(exportData)
	if err != nil {
		t.Fatalf("Import() failed: %v", err)
	}

	if importedProfile.Name != profile.Name {
		t.Errorf("Expected imported profile name %s, got %s", profile.Name, importedProfile.Name)
	}

	// Verify profile saved using the new ID from import
	loadedProfile, err := storage.Load(importedProfile.ID)
	if err != nil {
		t.Fatalf("Load() after import failed: %v", err)
	}

	if loadedProfile.Name != profile.Name {
		t.Errorf("Expected loaded profile name %s, got %s", profile.Name, loadedProfile.Name)
	}
}

func TestFileStorage_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// Create invalid YAML file
	filePath := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(filePath, []byte("invalid: yaml: content:"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	_, err = storage.Load("invalid")
	if err == nil {
		t.Errorf("Expected error when loading invalid YAML")
	}
}

func TestFileStorage_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Environment: "test",
		Config:      map[string]string{"secret": "value"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(profile)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check file permissions
	filePath := filepath.Join(tempDir, "profiles", "test-profile.yaml")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat profile file: %v", err)
	}

	// Check that file is readable and writable by owner only
	mode := fileInfo.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", mode.Perm())
	}
}
