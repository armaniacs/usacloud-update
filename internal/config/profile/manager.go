package profile

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NewProfileManager creates a new profile manager
func NewProfileManager(configDir string) (*ProfileManager, error) {
	storage := NewFileStorage(configDir)

	pm := &ProfileManager{
		profiles:        make(map[string]*Profile),
		configDir:       configDir,
		storage:         storage,
		templateManager: NewTemplateManager(),
	}

	// Load existing profiles
	if err := pm.loadProfiles(); err != nil {
		return nil, fmt.Errorf("failed to load profiles: %w", err)
	}

	// Set active profile to default if available
	if defaultProfile, err := pm.storage.GetDefault(); err == nil {
		pm.activeProfile = defaultProfile
	}

	return pm, nil
}

// CreateProfile creates a new profile
func (pm *ProfileManager) CreateProfile(opts ProfileCreateOptions) (*Profile, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	profile := &Profile{
		ID:          generateProfileID(),
		Name:        opts.Name,
		Description: opts.Description,
		Environment: opts.Environment,
		ParentID:    opts.ParentID,
		Config:      make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        opts.Tags,
		IsDefault:   opts.SetDefault,
	}

	// Copy config
	for k, v := range opts.Config {
		profile.Config[k] = v
	}

	// Check for duplicate names
	if pm.profileExists(opts.Name) {
		return nil, fmt.Errorf("profile with name '%s' already exists", opts.Name)
	}

	// Apply inheritance if parent is specified
	if opts.ParentID != "" {
		if err := pm.applyInheritance(profile, opts.ParentID); err != nil {
			return nil, fmt.Errorf("failed to apply inheritance: %w", err)
		}
	}

	// Validate profile
	if err := pm.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("profile validation failed: %w", err)
	}

	// If this is the first profile or set as default, make it default
	if len(pm.profiles) == 0 || opts.SetDefault {
		// Unset previous default
		for _, p := range pm.profiles {
			if p.IsDefault {
				p.IsDefault = false
				pm.storage.Save(p)
			}
		}
		profile.IsDefault = true
	}

	pm.profiles[profile.ID] = profile

	if err := pm.storage.Save(profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	if profile.IsDefault {
		pm.storage.SetDefault(profile.ID)
	}

	return profile, nil
}

// CreateProfileFromParent creates a new profile inheriting from a parent
func (pm *ProfileManager) CreateProfileFromParent(name, description string, parentID string, overrides map[string]string) (*Profile, error) {
	parent, exists := pm.profiles[parentID]
	if !exists {
		return nil, fmt.Errorf("parent profile not found: %s", parentID)
	}

	// Create base config from parent
	config := make(map[string]string)
	for k, v := range parent.Config {
		config[k] = v
	}

	// Apply overrides
	for k, v := range overrides {
		config[k] = v
	}

	opts := ProfileCreateOptions{
		Name:        name,
		Description: description,
		Environment: parent.Environment,
		Config:      config,
		ParentID:    parentID,
		Tags:        append([]string{}, parent.Tags...),
	}

	return pm.CreateProfile(opts)
}

// UpdateProfile updates an existing profile
func (pm *ProfileManager) UpdateProfile(id string, opts ProfileUpdateOptions) error {
	profile, exists := pm.profiles[id]
	if !exists {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Update fields if provided
	if opts.Name != nil {
		// Check for duplicate names (excluding current profile)
		for _, p := range pm.profiles {
			if p.ID != id && p.Name == *opts.Name {
				return fmt.Errorf("profile with name '%s' already exists", *opts.Name)
			}
		}
		profile.Name = *opts.Name
	}

	if opts.Description != nil {
		profile.Description = *opts.Description
	}

	if opts.Environment != nil {
		profile.Environment = *opts.Environment
	}

	if opts.Tags != nil {
		profile.Tags = opts.Tags
	}

	if opts.SetDefault != nil && *opts.SetDefault {
		// Unset previous default
		for _, p := range pm.profiles {
			if p.IsDefault {
				p.IsDefault = false
				pm.storage.Save(p)
			}
		}
		profile.IsDefault = true
		pm.storage.SetDefault(profile.ID)
	}

	// Update config
	for key, value := range opts.Config {
		if value == "" {
			delete(profile.Config, key)
		} else {
			profile.Config[key] = value
		}
	}

	profile.UpdatedAt = time.Now()

	// Update child profiles if needed
	if err := pm.updateChildProfiles(profile); err != nil {
		return fmt.Errorf("failed to update child profiles: %w", err)
	}

	return pm.storage.Save(profile)
}

// SwitchProfile switches to the specified profile
func (pm *ProfileManager) SwitchProfile(id string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	profile, exists := pm.profiles[id]
	if !exists {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Update last used time for current profile
	if pm.activeProfile != nil {
		pm.activeProfile.LastUsedAt = time.Now()
		pm.storage.Save(pm.activeProfile)
	}

	// Set new active profile
	pm.activeProfile = profile
	profile.LastUsedAt = time.Now()

	// Apply profile to environment
	if err := pm.applyProfileToEnvironment(profile); err != nil {
		return fmt.Errorf("failed to apply profile to environment: %w", err)
	}

	return pm.storage.Save(profile)
}

// DeleteProfile deletes a profile
func (pm *ProfileManager) DeleteProfile(id string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	profile, exists := pm.profiles[id]
	if !exists {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Check if profile has children
	children := pm.getChildProfiles(id)
	if len(children) > 0 {
		return fmt.Errorf("cannot delete profile with child profiles: %v",
			pm.getProfileNames(children))
	}

	// If this is the active profile, clear it
	if pm.activeProfile != nil && pm.activeProfile.ID == id {
		pm.activeProfile = nil
	}

	// If this is the default profile, unset default
	if profile.IsDefault {
		// Try to set another profile as default
		for _, p := range pm.profiles {
			if p.ID != id {
				p.IsDefault = true
				pm.storage.SetDefault(p.ID)
				pm.storage.Save(p)
				break
			}
		}
	}

	delete(pm.profiles, id)
	return pm.storage.Delete(id)
}

// GetProfile retrieves a profile by ID or name
func (pm *ProfileManager) GetProfile(idOrName string) (*Profile, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	// Try by ID first
	if profile, exists := pm.profiles[idOrName]; exists {
		return profile, nil
	}

	// Try by name
	for _, profile := range pm.profiles {
		if profile.Name == idOrName {
			return profile, nil
		}
	}

	return nil, fmt.Errorf("profile not found: %s", idOrName)
}

// ListProfiles returns a list of profiles based on options
func (pm *ProfileManager) ListProfiles(opts ...ProfileListOptions) []*Profile {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	var options ProfileListOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	profiles := make([]*Profile, 0, len(pm.profiles))

	// Filter profiles
	for _, profile := range pm.profiles {
		// Filter by environment
		if options.Environment != "" && profile.Environment != options.Environment {
			continue
		}

		// Filter by tags
		if len(options.Tags) > 0 {
			hasTag := false
			for _, tag := range options.Tags {
				for _, profileTag := range profile.Tags {
					if profileTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		profiles = append(profiles, profile)
	}

	// Sort profiles
	sort.Slice(profiles, func(i, j int) bool {
		switch options.SortBy {
		case "created_at":
			if options.SortOrder == "desc" {
				return profiles[i].CreatedAt.After(profiles[j].CreatedAt)
			}
			return profiles[i].CreatedAt.Before(profiles[j].CreatedAt)
		case "last_used_at":
			if options.SortOrder == "desc" {
				return profiles[i].LastUsedAt.After(profiles[j].LastUsedAt)
			}
			return profiles[i].LastUsedAt.Before(profiles[j].LastUsedAt)
		default: // name
			if options.SortOrder == "desc" {
				return profiles[i].Name > profiles[j].Name
			}
			return profiles[i].Name < profiles[j].Name
		}
	})

	return profiles
}

// GetActiveProfile returns the currently active profile
func (pm *ProfileManager) GetActiveProfile() *Profile {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.activeProfile
}

// GetDefaultProfile returns the default profile
func (pm *ProfileManager) GetDefaultProfile() *Profile {
	for _, profile := range pm.profiles {
		if profile.IsDefault {
			return profile
		}
	}
	return nil
}

// ExportProfile exports a profile to a file
func (pm *ProfileManager) ExportProfile(id, filepath string) error {
	profile, exists := pm.profiles[id]
	if !exists {
		return fmt.Errorf("profile not found: %s", id)
	}

	return writeYAMLFile(filepath, profile)
}

// ImportProfile imports a profile from a file
func (pm *ProfileManager) ImportProfile(filepath string) (*Profile, error) {
	var profile Profile
	if err := readYAMLFile(filepath, &profile); err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Generate new ID to avoid conflicts
	profile.ID = generateProfileID()
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	profile.LastUsedAt = time.Time{}
	profile.IsDefault = false

	// Check for duplicate names
	originalName := profile.Name
	counter := 1
	for pm.profileExists(profile.Name) {
		profile.Name = fmt.Sprintf("%s (%d)", originalName, counter)
		counter++
	}

	// Validate profile
	if err := pm.validateProfile(&profile); err != nil {
		return nil, fmt.Errorf("imported profile validation failed: %w", err)
	}

	pm.profiles[profile.ID] = &profile

	if err := pm.storage.Save(&profile); err != nil {
		return nil, fmt.Errorf("failed to save imported profile: %w", err)
	}

	return &profile, nil
}

// Helper methods

func (pm *ProfileManager) loadProfiles() error {
	profiles, err := pm.storage.LoadAll()
	if err != nil {
		return err
	}

	pm.profiles = profiles
	return nil
}

func (pm *ProfileManager) profileExists(name string) bool {
	for _, profile := range pm.profiles {
		if profile.Name == name {
			return true
		}
	}
	return false
}

func (pm *ProfileManager) applyInheritance(profile *Profile, parentID string) error {
	parent, exists := pm.profiles[parentID]
	if !exists {
		return fmt.Errorf("parent profile not found: %s", parentID)
	}

	// Inherit config from parent (don't override existing values)
	for k, v := range parent.Config {
		if _, exists := profile.Config[k]; !exists {
			profile.Config[k] = v
		}
	}

	// Inherit environment if not set
	if profile.Environment == "" {
		profile.Environment = parent.Environment
	}

	// Merge tags
	tagSet := make(map[string]bool)
	for _, tag := range profile.Tags {
		tagSet[tag] = true
	}
	for _, tag := range parent.Tags {
		if !tagSet[tag] {
			profile.Tags = append(profile.Tags, tag)
		}
	}

	return nil
}

func (pm *ProfileManager) validateProfile(profile *Profile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.Environment == "" {
		return fmt.Errorf("profile environment is required")
	}

	// Validate environment
	validEnvs := []string{EnvironmentProduction, EnvironmentStaging, EnvironmentDevelopment, EnvironmentTest}
	validEnv := false
	for _, env := range validEnvs {
		if profile.Environment == env {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("invalid environment: %s (must be one of: %s)",
			profile.Environment, strings.Join(validEnvs, ", "))
	}

	// Validate required config keys based on environment
	if err := pm.validateRequiredConfig(profile); err != nil {
		return err
	}

	return nil
}

func (pm *ProfileManager) validateRequiredConfig(profile *Profile) error {
	// Basic required keys for all environments
	requiredKeys := []string{ConfigKeyAccessToken, ConfigKeyAccessTokenSecret}

	for _, key := range requiredKeys {
		if value, exists := profile.Config[key]; !exists || value == "" {
			return fmt.Errorf("required configuration key missing: %s", key)
		}
	}

	// Validate specific keys
	if zone, exists := profile.Config[ConfigKeyZone]; exists && zone != "" {
		validZones := []string{"tk1v", "is1a", "is1b", "tk1a", "tk1b"}
		validZone := false
		for _, validZ := range validZones {
			if zone == validZ {
				validZone = true
				break
			}
		}
		if !validZone {
			return fmt.Errorf("invalid zone: %s", zone)
		}
	}

	return nil
}

func (pm *ProfileManager) updateChildProfiles(parent *Profile) error {
	children := pm.getChildProfiles(parent.ID)

	for _, child := range children {
		// Update inherited config (don't override child's explicit config)
		for k, v := range parent.Config {
			// Only update if child doesn't have explicit override
			if _, hasOverride := child.Config[k]; !hasOverride {
				child.Config[k] = v
			}
		}

		child.UpdatedAt = time.Now()
		if err := pm.storage.Save(child); err != nil {
			return fmt.Errorf("failed to update child profile %s: %w", child.Name, err)
		}
	}

	return nil
}

func (pm *ProfileManager) getChildProfiles(parentID string) []*Profile {
	var children []*Profile
	for _, profile := range pm.profiles {
		if profile.ParentID == parentID {
			children = append(children, profile)
		}
	}
	return children
}

func (pm *ProfileManager) getProfileNames(profiles []*Profile) []string {
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

func (pm *ProfileManager) applyProfileToEnvironment(profile *Profile) error {
	// Apply environment variables
	for key, value := range profile.Config {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	// Write current config file
	configFile := filepath.Join(pm.configDir, "current.conf")
	return pm.writeConfigFile(configFile, profile.Config)
}

func (pm *ProfileManager) writeConfigFile(filepath string, config map[string]string) error {
	var lines []string
	for key, value := range config {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filepath, []byte(content), 0600)
}

func generateProfileID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// IsSensitiveKey returns true if the key contains sensitive information
func IsSensitiveKey(key string) bool {
	sensitivePatterns := []string{
		"TOKEN", "SECRET", "KEY", "PASSWORD", "PASS",
	}

	upperKey := strings.ToUpper(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(upperKey, pattern) {
			return true
		}
	}
	return false
}

// MaskValue masks sensitive values for display
func MaskValue(value string) string {
	if len(value) == 0 {
		return ""
	}

	// For strings <= 8 characters, return fixed mask
	if len(value) <= 8 {
		return "****"
	}

	// For strings 9-12 characters, preserve original length
	if len(value) <= 12 {
		starCount := len(value) - 8
		stars := strings.Repeat("*", starCount)
		return value[:4] + stars + value[len(value)-4:]
	}

	// For very long strings, use fixed 24 stars
	stars := strings.Repeat("*", 24)
	return value[:4] + stars + value[len(value)-4:]
}

// SetDefault sets a profile as the default
func (pm *ProfileManager) SetDefault(profileID string) error {
	profile, exists := pm.profiles[profileID]
	if !exists {
		return fmt.Errorf("profile not found: %s", profileID)
	}

	// Remove default flag from all profiles
	for _, p := range pm.profiles {
		p.IsDefault = false
	}

	// Set the specified profile as default
	profile.IsDefault = true
	pm.activeProfile = profile

	// Save to storage
	return pm.storage.SetDefault(profileID)
}

// GetBuiltinTemplates returns builtin profile templates
func (pm *ProfileManager) GetBuiltinTemplates() []ProfileTemplate {
	return pm.templateManager.GetBuiltinTemplates()
}

// ValidateTemplateConfig validates template configuration
func (pm *ProfileManager) ValidateTemplateConfig(template ProfileTemplate, config map[string]string) error {
	return pm.templateManager.validateTemplateConfig(&template, config)
}

// CreateProfileFromTemplate creates a profile from a template
func (pm *ProfileManager) CreateProfileFromTemplate(templateName string, profileName string, config map[string]string) (*Profile, error) {
	profile, err := pm.templateManager.CreateProfileFromTemplate(templateName, profileName, config)
	if err != nil {
		return nil, err
	}

	// Check for duplicate names
	if pm.profileExists(profile.Name) {
		return nil, fmt.Errorf("profile with name '%s' already exists", profile.Name)
	}

	// Validate profile
	if err := pm.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("profile validation failed: %w", err)
	}

	// Add to profiles map
	pm.profiles[profile.ID] = profile

	// Save to storage
	if err := pm.storage.Save(profile); err != nil {
		delete(pm.profiles, profile.ID) // Rollback
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// validateConfigKeyType validates the type of a config value
func (pm *ProfileManager) validateConfigKeyType(keyDef ConfigKeyDef, value string) error {
	// Validate type
	if err := pm.templateManager.validateConfigType(keyDef, value); err != nil {
		return err
	}

	// Validate pattern if specified
	if keyDef.Validation != "" {
		if err := pm.templateManager.validateConfigPattern(keyDef, value); err != nil {
			return err
		}
	}

	return nil
}

// GetTemplateByName gets a template by name
func (pm *ProfileManager) GetTemplateByName(name string) (*ProfileTemplate, error) {
	return pm.templateManager.GetTemplate(name)
}

// applyTemplateDefaults applies default values from a template
func (pm *ProfileManager) applyTemplateDefaults(template *ProfileTemplate, config map[string]string) map[string]string {
	// Get default config from template manager
	defaultConfig, err := pm.templateManager.GetDefaultConfig(template.Name)
	if err != nil {
		// If we can't get defaults, return original config
		return config
	}

	// Apply defaults for missing keys
	finalConfig := make(map[string]string)

	// First add defaults
	for key, value := range defaultConfig {
		finalConfig[key] = value
	}

	// Then override with provided config
	for key, value := range config {
		finalConfig[key] = value
	}

	return finalConfig
}

// Utility functions for tests

// isSensitiveKey checks if a key contains sensitive information
func isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"token", "secret", "key", "password", "pass", "credential",
		"auth", "api_key", "access_token", "sakuracloud_access_token",
	}

	keyLower := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}
	return false
}

// maskValue masks sensitive values for command display (different algorithm than MaskValue)
func maskValue(value string) string {
	if len(value) == 0 {
		return ""
	}
	if len(value) == 1 {
		return "*"
	}
	if len(value) == 2 {
		return "**"
	}
	if len(value) == 3 {
		return value[:1] + "**"
	}
	if len(value) == 4 {
		return value[:1] + "***"
	}
	if len(value) == 5 || len(value) == 6 {
		// Show first and last char with stars in between (preserve length)
		stars := strings.Repeat("*", len(value)-2)
		return value[:1] + stars + value[len(value)-1:]
	}
	if len(value) >= 7 {
		// For longer strings, compress to fixed format: first + 17 stars + last
		stars := strings.Repeat("*", 17)
		return value[:1] + stars + value[len(value)-1:]
	}
	return "****"
}
