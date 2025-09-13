package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PresetManager manages filter presets
type PresetManager struct {
	presets map[string]*FilterSet
	storage PresetStorage
}

// PresetStorage defines the interface for preset storage
type PresetStorage interface {
	Save(preset *FilterSet) error
	Load(id string) (*FilterSet, error)
	LoadAll() ([]*FilterSet, error)
	List() []string
	Delete(id string) error
	Exists(id string) bool
}

// FilePresetStorage implements preset storage using files
type FilePresetStorage struct {
	baseDir string
}

// NewPresetManager creates a new preset manager
func NewPresetManager(storage PresetStorage) *PresetManager {
	pm := &PresetManager{
		presets: make(map[string]*FilterSet),
		storage: storage,
	}

	// Load existing presets
	if presets, err := storage.LoadAll(); err == nil {
		for _, preset := range presets {
			pm.presets[preset.ID] = preset
		}
	}

	return pm
}

// NewFilePresetStorage creates a file-based preset storage
func NewFilePresetStorage(baseDir string) (*FilePresetStorage, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".config", "usacloud-update", "filter-presets")
	}

	// Ensure directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create preset directory: %w", err)
	}

	return &FilePresetStorage{baseDir: baseDir}, nil
}

// SaveCurrentAsPreset saves the current filter state as a new preset
func (pm *PresetManager) SaveCurrentAsPreset(name string, fs *FilterSystem) error {
	// Generate a unique ID
	id := pm.generateID(name)

	// Get FilterExport directly
	exports := fs.ExportConfig()

	preset := &FilterSet{
		ID:      id,
		Name:    name,
		Filters: exports,
		Created: time.Now(),
	}

	// Save to storage
	if err := pm.storage.Save(preset); err != nil {
		return fmt.Errorf("failed to save preset: %w", err)
	}

	// Add to memory
	pm.presets[preset.ID] = preset

	return nil
}

// ApplyPreset applies a preset to the filter system
func (pm *PresetManager) ApplyPreset(id string, fs *FilterSystem) error {
	preset, exists := pm.presets[id]
	if !exists {
		// Try to load from storage
		var err error
		preset, err = pm.storage.Load(id)
		if err != nil {
			return fmt.Errorf("preset not found: %s", id)
		}
		pm.presets[id] = preset
	}

	// Use FilterExport directly
	return fs.ImportConfig(preset.Filters)
}

// GetPreset returns a preset by ID
func (pm *PresetManager) GetPreset(id string) (*FilterSet, error) {
	if preset, exists := pm.presets[id]; exists {
		return preset, nil
	}

	// Try to load from storage
	preset, err := pm.storage.Load(id)
	if err != nil {
		return nil, fmt.Errorf("preset not found: %s", id)
	}

	pm.presets[id] = preset
	return preset, nil
}

// ListPresets returns all available presets
func (pm *PresetManager) ListPresets() []*FilterSet {
	var presets []*FilterSet
	for _, preset := range pm.presets {
		presets = append(presets, preset)
	}
	return presets
}

// GetPresetNames returns names of all presets
func (pm *PresetManager) GetPresetNames() []string {
	var names []string
	for _, preset := range pm.presets {
		names = append(names, preset.Name)
	}
	return names
}

// DeletePreset deletes a preset
func (pm *PresetManager) DeletePreset(id string) error {
	if err := pm.storage.Delete(id); err != nil {
		return fmt.Errorf("failed to delete preset: %w", err)
	}

	delete(pm.presets, id)
	return nil
}

// UpdatePreset updates an existing preset
func (pm *PresetManager) UpdatePreset(id string, fs *FilterSystem) error {
	preset, exists := pm.presets[id]
	if !exists {
		return fmt.Errorf("preset not found: %s", id)
	}

	// Update the filters but keep the name and ID
	exports := fs.ExportConfig()
	preset.Filters = exports

	// Save to storage
	if err := pm.storage.Save(preset); err != nil {
		return fmt.Errorf("failed to update preset: %w", err)
	}

	return nil
}

// RenamePreset renames a preset
func (pm *PresetManager) RenamePreset(id, newName string) error {
	preset, exists := pm.presets[id]
	if !exists {
		return fmt.Errorf("preset not found: %s", id)
	}

	preset.Name = newName

	// Save to storage
	if err := pm.storage.Save(preset); err != nil {
		return fmt.Errorf("failed to rename preset: %w", err)
	}

	return nil
}

// ExportPresets exports all presets to a JSON file
func (pm *PresetManager) ExportPresets(filename string) error {
	presets := pm.ListPresets()

	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal presets: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportPresets imports presets from a JSON file
func (pm *PresetManager) ImportPresets(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var presets []*FilterSet
	if err := json.Unmarshal(data, &presets); err != nil {
		return fmt.Errorf("failed to unmarshal presets: %w", err)
	}

	for _, preset := range presets {
		// Generate new ID to avoid conflicts
		preset.ID = pm.generateID(preset.Name)

		if err := pm.storage.Save(preset); err != nil {
			return fmt.Errorf("failed to save imported preset %s: %w", preset.Name, err)
		}

		pm.presets[preset.ID] = preset
	}

	return nil
}

// generateID generates a unique ID for a preset
func (pm *PresetManager) generateID(name string) string {
	// Create a base ID from the name
	baseID := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	baseID = strings.ReplaceAll(baseID, "_", "-")

	// Remove invalid characters
	var validChars strings.Builder
	for _, r := range baseID {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			validChars.WriteRune(r)
		}
	}
	baseID = validChars.String()

	// Ensure uniqueness
	id := baseID
	counter := 1
	for pm.storage.Exists(id) {
		id = fmt.Sprintf("%s-%d", baseID, counter)
		counter++
	}

	return id
}

// FilePresetStorage implementation

// Save saves a preset to a file
func (fps *FilePresetStorage) Save(preset *FilterSet) error {
	filename := filepath.Join(fps.baseDir, preset.ID+".json")

	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}

	return nil
}

// Load loads a preset from a file
func (fps *FilePresetStorage) Load(id string) (*FilterSet, error) {
	filename := filepath.Join(fps.baseDir, id+".json")

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset file: %w", err)
	}

	var preset FilterSet
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preset: %w", err)
	}

	return &preset, nil
}

// LoadAll loads all presets from files
func (fps *FilePresetStorage) LoadAll() ([]*FilterSet, error) {
	files, err := ioutil.ReadDir(fps.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset directory: %w", err)
	}

	var presets []*FilterSet
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			id := strings.TrimSuffix(file.Name(), ".json")
			preset, err := fps.Load(id)
			if err != nil {
				// Skip invalid files
				continue
			}
			presets = append(presets, preset)
		}
	}

	return presets, nil
}

// List returns a list of all preset IDs
func (fps *FilePresetStorage) List() []string {
	files, err := ioutil.ReadDir(fps.baseDir)
	if err != nil {
		return nil
	}

	var ids []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			id := strings.TrimSuffix(file.Name(), ".json")
			ids = append(ids, id)
		}
	}

	return ids
}

// Delete deletes a preset file
func (fps *FilePresetStorage) Delete(id string) error {
	filename := filepath.Join(fps.baseDir, id+".json")

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete preset file: %w", err)
	}

	return nil
}

// Exists checks if a preset file exists
func (fps *FilePresetStorage) Exists(id string) bool {
	filename := filepath.Join(fps.baseDir, id+".json")
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
