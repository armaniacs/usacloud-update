package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FileStorage implements ProfileStorage using YAML files
type FileStorage struct {
	configDir string
}

// NewFileStorage creates a new file-based profile storage
func NewFileStorage(configDir string) *FileStorage {
	return &FileStorage{
		configDir: configDir,
	}
}

// Save saves a profile to a YAML file
func (fs *FileStorage) Save(profile *Profile) error {
	if err := fs.ensureConfigDir(); err != nil {
		return err
	}

	filename := fs.getProfileFilename(profile.ID)
	return writeYAMLFile(filename, profile)
}

// Load loads a profile from a YAML file
func (fs *FileStorage) Load(id string) (*Profile, error) {
	filename := fs.getProfileFilename(id)

	var profile Profile
	if err := readYAMLFile(filename, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// LoadAll loads all profiles from the config directory
func (fs *FileStorage) LoadAll() (map[string]*Profile, error) {
	if err := fs.ensureConfigDir(); err != nil {
		return nil, err
	}

	profiles := make(map[string]*Profile)

	// Read all .yaml files in the profiles directory
	profilesDir := fs.getProfilesDir()
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil // Empty profiles map if directory doesn't exist
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filename := filepath.Join(profilesDir, entry.Name())
		var profile Profile

		if err := readYAMLFile(filename, &profile); err != nil {
			// Log error but continue loading other profiles
			fmt.Fprintf(os.Stderr, "Warning: failed to load profile %s: %v\n", filename, err)
			continue
		}

		profiles[profile.ID] = &profile
	}

	return profiles, nil
}

// Delete removes a profile file
func (fs *FileStorage) Delete(id string) error {
	filename := fs.getProfileFilename(id)
	return os.Remove(filename)
}

// SetDefault sets the default profile
func (fs *FileStorage) SetDefault(id string) error {
	defaultFile := fs.getDefaultFilename()
	return os.WriteFile(defaultFile, []byte(id), 0644)
}

// GetDefault gets the default profile
func (fs *FileStorage) GetDefault() (*Profile, error) {
	defaultFile := fs.getDefaultFilename()

	data, err := os.ReadFile(defaultFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no default profile set")
		}
		return nil, fmt.Errorf("failed to read default profile file: %w", err)
	}

	id := strings.TrimSpace(string(data))
	return fs.Load(id)
}

// Helper methods

func (fs *FileStorage) ensureConfigDir() error {
	profilesDir := fs.getProfilesDir()
	return os.MkdirAll(profilesDir, 0755)
}

func (fs *FileStorage) getProfilesDir() string {
	return filepath.Join(fs.configDir, "profiles")
}

func (fs *FileStorage) getProfileFilename(id string) string {
	return filepath.Join(fs.getProfilesDir(), fmt.Sprintf("%s.yaml", id))
}

func (fs *FileStorage) getDefaultFilename() string {
	return filepath.Join(fs.configDir, "default-profile")
}

// Utility functions for YAML I/O

func writeYAMLFile(filename string, data interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.CreateTemp(dir, "profile-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tempName := file.Name()
	var success bool
	defer func() {
		file.Close()
		if !success {
			os.Remove(tempName) // Clean up on failure
		}
	}()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempName, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true

	// Set appropriate permissions for profile files (contains sensitive data)
	return os.Chmod(filename, 0600)
}

func readYAMLFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile not found: %s", filename)
		}
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode YAML from %s: %w", filename, err)
	}

	return nil
}

// BackupStorage provides backup and restore functionality
type BackupStorage struct {
	storage   ProfileStorage
	backupDir string
}

// NewBackupStorage creates a new backup storage wrapper
func NewBackupStorage(storage ProfileStorage, backupDir string) *BackupStorage {
	return &BackupStorage{
		storage:   storage,
		backupDir: backupDir,
	}
}

// Backup creates a backup of all profiles
func (bs *BackupStorage) Backup() (string, error) {
	if err := os.MkdirAll(bs.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	profiles, err := bs.storage.LoadAll()
	if err != nil {
		return "", fmt.Errorf("failed to load profiles for backup: %w", err)
	}

	// Create timestamped backup file
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	backupFile := filepath.Join(bs.backupDir, fmt.Sprintf("profiles-backup-%s.yaml", timestamp))

	if err := writeYAMLFile(backupFile, profiles); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupFile, nil
}

// Restore restores profiles from a backup file
func (bs *BackupStorage) Restore(backupFile string) error {
	var profiles map[string]*Profile
	if err := readYAMLFile(backupFile, &profiles); err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Save each profile
	for _, profile := range profiles {
		if err := bs.storage.Save(profile); err != nil {
			return fmt.Errorf("failed to restore profile %s: %w", profile.Name, err)
		}
	}

	return nil
}

// ListBackups lists available backup files
func (bs *BackupStorage) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(bs.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "profiles-backup-") && strings.HasSuffix(entry.Name(), ".yaml") {
			backups = append(backups, filepath.Join(bs.backupDir, entry.Name()))
		}
	}

	return backups, nil
}

// Backup creates a backup of current profiles
func (fs *FileStorage) Backup() (string, error) {
	backupFile := fmt.Sprintf("profiles-backup-%s.yaml", time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(fs.configDir, backupFile)

	// Read current profiles
	profiles, err := fs.LoadAll()
	if err != nil {
		return "", fmt.Errorf("failed to load profiles for backup: %w", err)
	}

	// Save backup
	data, err := yaml.Marshal(profiles)
	if err != nil {
		return "", fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// Restore restores profiles from a backup file
func (fs *FileStorage) Restore(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var profiles map[string]*Profile
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	// Save each profile
	for _, profile := range profiles {
		if err := fs.Save(profile); err != nil {
			return fmt.Errorf("failed to restore profile %s: %w", profile.ID, err)
		}
	}

	return nil
}

// Export exports a single profile and returns the data
func (fs *FileStorage) Export(profileID string) ([]byte, error) {
	profile, err := fs.Load(profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile for export: %w", err)
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	return data, nil
}

// Import imports a profile from data and returns the imported profile
func (fs *FileStorage) Import(data []byte) (*Profile, error) {
	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Generate new ID to avoid conflicts
	profile.ID = generateProfileID()
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	if err := fs.Save(&profile); err != nil {
		return nil, fmt.Errorf("failed to save imported profile %s: %w", profile.Name, err)
	}

	return &profile, nil
}

// ExportToFile exports profiles to a specified file
func (fs *FileStorage) ExportToFile(exportPath string) error {
	profiles, err := fs.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load profiles for export: %w", err)
	}

	data, err := yaml.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportFromFile imports profiles from a specified file
func (fs *FileStorage) ImportFromFile(importPath string) error {
	data, err := os.ReadFile(importPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var profiles map[string]*Profile
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Save each profile
	for _, profile := range profiles {
		// Generate new ID to avoid conflicts
		profile.ID = generateProfileID()
		profile.CreatedAt = time.Now()
		profile.UpdatedAt = time.Now()

		if err := fs.Save(profile); err != nil {
			return fmt.Errorf("failed to import profile %s: %w", profile.Name, err)
		}
	}

	return nil
}
