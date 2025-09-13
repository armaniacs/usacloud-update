package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ProfileManager struct {
	config      *IntegratedConfig
	profilesDir string
}

func NewProfileManager(config *IntegratedConfig) *ProfileManager {
	profilesDir := filepath.Dir(config.configPath)
	if config.configPath == "" {
		homeDir, _ := os.UserHomeDir()
		profilesDir = filepath.Join(homeDir, ".config", "usacloud-update")
	}

	return &ProfileManager{
		config:      config,
		profilesDir: profilesDir,
	}
}

func (pm *ProfileManager) CreateProfile(name, basedOn, description string) error {
	if name == "" {
		return fmt.Errorf("プロファイル名を指定してください")
	}

	if _, exists := pm.config.Profiles[name]; exists {
		return fmt.Errorf("プロファイル '%s' は既に存在します", name)
	}

	var baseProfile *ProfileConfig
	if basedOn != "" {
		var exists bool
		baseProfile, exists = pm.config.Profiles[basedOn]
		if !exists {
			return fmt.Errorf("ベースプロファイル '%s' が見つかりません", basedOn)
		}
	}

	newProfile := &ProfileConfig{
		Name:        name,
		Description: description,
		BasedOn:     basedOn,
		Overrides:   make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UsageCount:  0,
	}

	if baseProfile != nil {
		for key, value := range baseProfile.Overrides {
			newProfile.Overrides[key] = value
		}
	}

	pm.config.Profiles[name] = newProfile

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) DeleteProfile(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("プロファイル名を指定してください")
	}

	if profileName == "default" {
		return fmt.Errorf("デフォルトプロファイルは削除できません")
	}

	if _, exists := pm.config.Profiles[profileName]; !exists {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	for _, profile := range pm.config.Profiles {
		if profile.BasedOn == profileName {
			return fmt.Errorf("プロファイル '%s' は他のプロファイルのベースとして使用されているため削除できません", profileName)
		}
	}

	if pm.config.profileName == profileName {
		pm.config.profileName = "default"
		pm.config.General.Profile = "default"
		if err := pm.config.applyProfile("default"); err != nil {
			return fmt.Errorf("デフォルトプロファイルの適用に失敗: %w", err)
		}
	}

	delete(pm.config.Profiles, profileName)

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) SwitchProfile(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("プロファイル名を指定してください")
	}

	profile, exists := pm.config.Profiles[profileName]
	if !exists {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	if currentProfile, exists := pm.config.Profiles[pm.config.profileName]; exists {
		currentProfile.LastUsed = time.Now()
		currentProfile.UsageCount++
	}

	pm.config.profileName = profileName
	pm.config.General.Profile = profileName

	if err := pm.config.applyProfile(profileName); err != nil {
		return fmt.Errorf("プロファイル適用に失敗: %w", err)
	}

	profile.LastUsed = time.Now()
	profile.UsageCount++

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) GetProfile(profileName string) (*ProfileConfig, error) {
	if profileName == "" {
		profileName = pm.config.profileName
		if profileName == "" {
			profileName = "default"
		}
	}

	profile, exists := pm.config.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	return profile, nil
}

func (pm *ProfileManager) ListProfiles() []*ProfileConfig {
	profiles := make([]*ProfileConfig, 0, len(pm.config.Profiles))
	for _, profile := range pm.config.Profiles {
		profiles = append(profiles, profile)
	}

	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Name == "default" {
			return true
		}
		if profiles[j].Name == "default" {
			return false
		}
		return profiles[i].UsageCount > profiles[j].UsageCount
	})

	return profiles
}

func (pm *ProfileManager) GetCurrentProfile() *ProfileConfig {
	profileName := pm.config.profileName
	if profileName == "" {
		profileName = "default"
	}

	if profile, exists := pm.config.Profiles[profileName]; exists {
		return profile
	}

	return pm.config.Profiles["default"]
}

func (pm *ProfileManager) UpdateProfileSetting(profileName, key string, value interface{}) error {
	if profileName == "" {
		profileName = pm.config.profileName
		if profileName == "" {
			profileName = "default"
		}
	}

	profile, exists := pm.config.Profiles[profileName]
	if !exists {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	if profile.Overrides == nil {
		profile.Overrides = make(map[string]interface{})
	}

	profile.Overrides[key] = value

	if profileName == pm.config.profileName {
		pm.config.applyOverride(key, value)
	}

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) RemoveProfileSetting(profileName, key string) error {
	if profileName == "" {
		profileName = pm.config.profileName
		if profileName == "" {
			profileName = "default"
		}
	}

	profile, exists := pm.config.Profiles[profileName]
	if !exists {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	if profile.Overrides == nil {
		return nil
	}

	delete(profile.Overrides, key)

	if profileName == pm.config.profileName {
		if err := pm.config.applyProfile(profileName); err != nil {
			return fmt.Errorf("プロファイル再適用に失敗: %w", err)
		}
	}

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) ExportProfile(profileName, exportPath string) error {
	profile, err := pm.GetProfile(profileName)
	if err != nil {
		return err
	}

	exportConfig := NewIntegratedConfig()

	exportConfig.Profiles = map[string]*ProfileConfig{
		profileName: profile,
	}

	if profile.BasedOn != "" {
		baseProfile, exists := pm.config.Profiles[profile.BasedOn]
		if exists {
			exportConfig.Profiles[profile.BasedOn] = baseProfile
		}
	}

	return exportConfig.SaveAs(exportPath)
}

func (pm *ProfileManager) ImportProfile(importPath string) error {
	importConfig, err := LoadIntegratedConfig(importPath)
	if err != nil {
		return fmt.Errorf("プロファイルインポートに失敗: %w", err)
	}

	for name, profile := range importConfig.Profiles {
		if _, exists := pm.config.Profiles[name]; exists {
			return fmt.Errorf("プロファイル '%s' は既に存在します", name)
		}

		pm.config.Profiles[name] = profile
	}

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) CloneProfile(sourceName, targetName, description string) error {
	sourceProfile, err := pm.GetProfile(sourceName)
	if err != nil {
		return fmt.Errorf("ソースプロファイルの取得に失敗: %w", err)
	}

	if _, exists := pm.config.Profiles[targetName]; exists {
		return fmt.Errorf("プロファイル '%s' は既に存在します", targetName)
	}

	newProfile := &ProfileConfig{
		Name:        targetName,
		Description: description,
		BasedOn:     sourceProfile.BasedOn,
		Overrides:   make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UsageCount:  0,
	}

	for key, value := range sourceProfile.Overrides {
		newProfile.Overrides[key] = value
	}

	pm.config.Profiles[targetName] = newProfile

	if pm.config.autoSave {
		return pm.config.Save()
	}

	return nil
}

func (pm *ProfileManager) GetProfileUsageStats() map[string]ProfileStats {
	stats := make(map[string]ProfileStats)

	for name, profile := range pm.config.Profiles {
		stats[name] = ProfileStats{
			Name:          name,
			Description:   profile.Description,
			UsageCount:    profile.UsageCount,
			LastUsed:      profile.LastUsed,
			CreatedAt:     profile.CreatedAt,
			IsCurrent:     name == pm.config.profileName,
			OverrideCount: len(profile.Overrides),
		}
	}

	return stats
}

type ProfileStats struct {
	Name          string
	Description   string
	UsageCount    int
	LastUsed      time.Time
	CreatedAt     time.Time
	IsCurrent     bool
	OverrideCount int
}

func (pm *ProfileManager) ValidateProfile(profileName string) error {
	profile, err := pm.GetProfile(profileName)
	if err != nil {
		return err
	}

	if profile.BasedOn != "" {
		if _, exists := pm.config.Profiles[profile.BasedOn]; !exists {
			return fmt.Errorf("ベースプロファイル '%s' が見つかりません", profile.BasedOn)
		}
	}

	validKeys := map[string]bool{
		"interactive_by_default":  true,
		"verbose":                 true,
		"color_output":            true,
		"skill_level":             true,
		"max_suggestions":         true,
		"enable_interactive_help": true,
		"show_common_mistakes":    true,
		"strict_mode":             true,
		"parallel_processing":     true,
		"show_progress":           true,
		"report_level":            true,
	}

	for key := range profile.Overrides {
		if !validKeys[key] {
			return fmt.Errorf("不正な設定キー: %s", key)
		}
	}

	return nil
}
