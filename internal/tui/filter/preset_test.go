package filter

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPresetManager(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)
	if manager == nil {
		t.Fatal("NewPresetManager should not return nil")
	}

	if manager.storage != storage {
		t.Error("PresetManager should store the provided storage")
	}

	if manager.presets == nil {
		t.Error("PresetManager should initialize presets map")
	}
}

func TestPresetManager_SaveCurrentAsPreset(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create a filter system with some configuration
	fs := NewFilterSystem()
	textFilter := fs.GetFilter("text")
	if textFilter != nil {
		textFilter.SetActive(true)
		textFilter.SetConfig(FilterConfig{"query": "test"})
	}

	// Save current state as preset
	err = manager.SaveCurrentAsPreset("Test Preset", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	// Verify preset was created
	presets := manager.ListPresets()
	if len(presets) != 1 {
		t.Errorf("Expected 1 preset, got %d", len(presets))
	}

	if presets[0].Name != "Test Preset" {
		t.Errorf("Expected preset name 'Test Preset', got %s", presets[0].Name)
	}
}

func TestPresetManager_ApplyPreset(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create and save a preset
	fs1 := NewFilterSystem()
	textFilter := fs1.GetFilter("テキスト検索")
	if textFilter != nil {
		textFilter.SetActive(true)
		textFilter.SetConfig(FilterConfig{"query": "test"})
	}

	err = manager.SaveCurrentAsPreset("Test Preset", fs1)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	presets := manager.ListPresets()
	if len(presets) == 0 {
		t.Fatal("No presets found")
	}

	presetID := presets[0].ID

	// Create new filter system and apply preset
	fs2 := NewFilterSystem()
	err = manager.ApplyPreset(presetID, fs2)
	if err != nil {
		t.Fatalf("ApplyPreset() failed: %v", err)
	}

	// Verify the preset was applied
	textFilter2 := fs2.GetFilter("テキスト検索")
	if textFilter2 == nil {
		t.Fatal("Text filter not found")
	}

	if !textFilter2.IsActive() {
		t.Error("Text filter should be active after applying preset")
	}

	config := textFilter2.GetConfig()
	if query, ok := config["query"]; !ok || query != "test" {
		t.Errorf("Expected query 'test', got %v", query)
	}
}

func TestPresetManager_GetPreset(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create and save a preset
	fs := NewFilterSystem()
	err = manager.SaveCurrentAsPreset("Test Preset", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	presets := manager.ListPresets()
	if len(presets) == 0 {
		t.Fatal("No presets found")
	}

	presetID := presets[0].ID

	// Get preset by ID
	preset, err := manager.GetPreset(presetID)
	if err != nil {
		t.Fatalf("GetPreset() failed: %v", err)
	}

	if preset.Name != "Test Preset" {
		t.Errorf("Expected preset name 'Test Preset', got %s", preset.Name)
	}

	// Try to get non-existent preset
	_, err = manager.GetPreset("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent preset")
	}
}

func TestPresetManager_DeletePreset(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create and save a preset
	fs := NewFilterSystem()
	err = manager.SaveCurrentAsPreset("Test Preset", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	presets := manager.ListPresets()
	if len(presets) != 1 {
		t.Fatalf("Expected 1 preset, got %d", len(presets))
	}

	presetID := presets[0].ID

	// Delete the preset
	err = manager.DeletePreset(presetID)
	if err != nil {
		t.Fatalf("DeletePreset() failed: %v", err)
	}

	// Verify preset was deleted
	presets = manager.ListPresets()
	if len(presets) != 0 {
		t.Errorf("Expected 0 presets after deletion, got %d", len(presets))
	}

	// Verify file was deleted
	filename := filepath.Join(tmpDir, presetID+".json")
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Error("Preset file should be deleted")
	}
}

func TestPresetManager_RenamePreset(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create and save a preset
	fs := NewFilterSystem()
	err = manager.SaveCurrentAsPreset("Old Name", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	presets := manager.ListPresets()
	if len(presets) != 1 {
		t.Fatalf("Expected 1 preset, got %d", len(presets))
	}

	presetID := presets[0].ID

	// Rename the preset
	err = manager.RenamePreset(presetID, "New Name")
	if err != nil {
		t.Fatalf("RenamePreset() failed: %v", err)
	}

	// Verify the name was changed
	preset, err := manager.GetPreset(presetID)
	if err != nil {
		t.Fatalf("GetPreset() failed: %v", err)
	}

	if preset.Name != "New Name" {
		t.Errorf("Expected preset name 'New Name', got %s", preset.Name)
	}
}

func TestPresetManager_ExportImportPresets(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create and save multiple presets
	fs1 := NewFilterSystem()
	err = manager.SaveCurrentAsPreset("Preset 1", fs1)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	fs2 := NewFilterSystem()
	err = manager.SaveCurrentAsPreset("Preset 2", fs2)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	// Export presets
	exportFile := filepath.Join(tmpDir, "export.json")
	err = manager.ExportPresets(exportFile)
	if err != nil {
		t.Fatalf("ExportPresets() failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Error("Export file should exist")
	}

	// Create new manager and import presets
	tmpDir2 := t.TempDir()
	storage2, err := NewFilePresetStorage(tmpDir2)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager2 := NewPresetManager(storage2)

	err = manager2.ImportPresets(exportFile)
	if err != nil {
		t.Fatalf("ImportPresets() failed: %v", err)
	}

	// Verify presets were imported
	importedPresets := manager2.ListPresets()
	if len(importedPresets) != 2 {
		t.Errorf("Expected 2 imported presets, got %d", len(importedPresets))
	}

	names := manager2.GetPresetNames()
	if !contains(names, "Preset 1") || !contains(names, "Preset 2") {
		t.Errorf("Expected imported preset names, got %v", names)
	}
}

func TestFilePresetStorage_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	// Create a test preset
	preset := &FilterSet{
		ID:      "test-preset",
		Name:    "Test Preset",
		Filters: []FilterExport{{Name: "test", Active: true, Config: FilterConfig{"key": "value"}}},
		Created: time.Now(),
	}

	// Test Save
	err = storage.Save(preset)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Test Exists
	if !storage.Exists("test-preset") {
		t.Error("Preset should exist after saving")
	}

	// Test Load
	loadedPreset, err := storage.Load("test-preset")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loadedPreset.Name != preset.Name {
		t.Errorf("Expected loaded preset name %s, got %s", preset.Name, loadedPreset.Name)
	}

	// Test List
	ids := storage.List()
	if !contains(ids, "test-preset") {
		t.Errorf("Expected preset ID in list, got %v", ids)
	}

	// Test LoadAll
	allPresets, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() failed: %v", err)
	}

	if len(allPresets) != 1 {
		t.Errorf("Expected 1 preset, got %d", len(allPresets))
	}

	// Test Delete
	err = storage.Delete("test-preset")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	if storage.Exists("test-preset") {
		t.Error("Preset should not exist after deletion")
	}
}

func TestPresetManager_GenerateID(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	tests := []struct {
		name     string
		expected string
	}{
		{"Simple Name", "simple-name"},
		{"Name_With_Underscores", "name-with-underscores"},
		{"Name With Spaces", "name-with-spaces"},
		{"Name@#$%^&*()", "name"},
		{"123 Numbers", "123-numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := manager.generateID(tt.name)
			if id != tt.expected {
				t.Errorf("Expected ID %s, got %s", tt.expected, id)
			}
		})
	}
}

func TestPresetManager_IDUniqueness(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewFilePresetStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewFilePresetStorage() failed: %v", err)
	}

	manager := NewPresetManager(storage)

	// Create multiple presets with the same name
	fs := NewFilterSystem()

	err = manager.SaveCurrentAsPreset("Same Name", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	err = manager.SaveCurrentAsPreset("Same Name", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	err = manager.SaveCurrentAsPreset("Same Name", fs)
	if err != nil {
		t.Fatalf("SaveCurrentAsPreset() failed: %v", err)
	}

	// Verify all presets were created with unique IDs
	presets := manager.ListPresets()
	if len(presets) != 3 {
		t.Errorf("Expected 3 presets, got %d", len(presets))
	}

	// Check that all IDs are unique
	ids := make(map[string]bool)
	for _, preset := range presets {
		if ids[preset.ID] {
			t.Errorf("Duplicate ID found: %s", preset.ID)
		}
		ids[preset.ID] = true
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
