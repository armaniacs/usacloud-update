package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type IntegratedConfig struct {
	configPath      string
	profileName     string
	environmentName string

	General       *GeneralConfig
	Transform     *TransformConfig
	Validation    *ValidationConfig
	ErrorFeedback *ErrorFeedbackConfig
	HelpSystem    *HelpSystemConfig
	Performance   *PerformanceConfig
	Output        *OutputConfig

	Profiles     map[string]*ProfileConfig
	Environments map[string]*EnvironmentConfig

	LastModified  time.Time
	ConfigVersion string
	autoSave      bool
	watchers      []chan ConfigChangeEvent
}

type GeneralConfig struct {
	Version              string `ini:"version"`
	ColorOutput          bool   `ini:"color_output"`
	Language             string `ini:"language"`
	Verbose              bool   `ini:"verbose"`
	InteractiveByDefault bool   `ini:"interactive_by_default"`
	Profile              string `ini:"profile"`
}

type TransformConfig struct {
	PreserveComments       bool `ini:"preserve_comments"`
	AddExplanatoryComments bool `ini:"add_explanatory_comments"`
	ShowLineNumbers        bool `ini:"show_line_numbers"`
	BackupOriginal         bool `ini:"backup_original"`
}

type ValidationConfig struct {
	EnableValidation        bool `ini:"enable_validation"`
	StrictMode              bool `ini:"strict_mode"`
	ValidateBeforeTransform bool `ini:"validate_before_transform"`
	ValidateAfterTransform  bool `ini:"validate_after_transform"`
	MaxSuggestions          int  `ini:"max_suggestions"`
	MaxEditDistance         int  `ini:"max_edit_distance"`
	SkipDeprecatedWarnings  bool `ini:"skip_deprecated_warnings"`
	TypoDetectionEnabled    bool `ini:"typo_detection_enabled"`
}

type ErrorFeedbackConfig struct {
	ErrorFormat                   string  `ini:"error_format"`
	ShowSuggestions               bool    `ini:"show_suggestions"`
	ShowAlternatives              bool    `ini:"show_alternatives"`
	ShowMigrationGuide            bool    `ini:"show_migration_guide"`
	SuggestionConfidenceThreshold float64 `ini:"suggestion_confidence_threshold"`
}

type HelpSystemConfig struct {
	EnableInteractiveHelp  bool   `ini:"enable_interactive_help"`
	SkillLevel             string `ini:"skill_level"`
	PreferredHelpFormat    string `ini:"preferred_help_format"`
	ShowCommonMistakes     bool   `ini:"show_common_mistakes"`
	EnableLearningTracking bool   `ini:"enable_learning_tracking"`
}

type PerformanceConfig struct {
	ParallelProcessing bool `ini:"parallel_processing"`
	CacheEnabled       bool `ini:"cache_enabled"`
	CacheSizeMB        int  `ini:"cache_size_mb"`
	BatchSize          int  `ini:"batch_size"`
	WorkerCount        int  `ini:"worker_count"`
}

type OutputConfig struct {
	Format        string `ini:"format"`
	ShowProgress  bool   `ini:"show_progress"`
	ProgressStyle string `ini:"progress_style"`
	ReportLevel   string `ini:"report_level"`
}

type ProfileConfig struct {
	Name        string                 `ini:"name"`
	Description string                 `ini:"description"`
	BasedOn     string                 `ini:"based_on"`
	Overrides   map[string]interface{} `ini:"-"`
	CreatedAt   time.Time              `ini:"-"`
	LastUsed    time.Time              `ini:"-"`
	UsageCount  int                    `ini:"-"`
}

type EnvironmentConfig struct {
	Name              string                 `ini:"name"`
	SakuraAPIEndpoint string                 `ini:"sakura_api_endpoint"`
	TimeoutSeconds    int                    `ini:"timeout_seconds"`
	RetryCount        int                    `ini:"retry_count"`
	StrictMode        bool                   `ini:"strict_mode"`
	Overrides         map[string]interface{} `ini:"-"`
}

type ConfigChangeEvent struct {
	Section   string
	Key       string
	OldValue  interface{}
	NewValue  interface{}
	Timestamp time.Time
}

func NewIntegratedConfig() *IntegratedConfig {
	return &IntegratedConfig{
		General: &GeneralConfig{
			Version:              "1.9.0",
			ColorOutput:          true,
			Language:             "ja",
			Verbose:              false,
			InteractiveByDefault: false,
			Profile:              "default",
		},
		Transform: &TransformConfig{
			PreserveComments:       true,
			AddExplanatoryComments: true,
			ShowLineNumbers:        true,
			BackupOriginal:         false,
		},
		Validation: &ValidationConfig{
			EnableValidation:        true,
			StrictMode:              false,
			ValidateBeforeTransform: true,
			ValidateAfterTransform:  true,
			MaxSuggestions:          5,
			MaxEditDistance:         3,
			SkipDeprecatedWarnings:  false,
			TypoDetectionEnabled:    true,
		},
		ErrorFeedback: &ErrorFeedbackConfig{
			ErrorFormat:                   "comprehensive",
			ShowSuggestions:               true,
			ShowAlternatives:              true,
			ShowMigrationGuide:            true,
			SuggestionConfidenceThreshold: 0.5,
		},
		HelpSystem: &HelpSystemConfig{
			EnableInteractiveHelp:  true,
			SkillLevel:             "intermediate",
			PreferredHelpFormat:    "detailed",
			ShowCommonMistakes:     true,
			EnableLearningTracking: true,
		},
		Performance: &PerformanceConfig{
			ParallelProcessing: true,
			CacheEnabled:       true,
			CacheSizeMB:        100,
			BatchSize:          1000,
			WorkerCount:        0,
		},
		Output: &OutputConfig{
			Format:        "auto",
			ShowProgress:  true,
			ProgressStyle: "bar",
			ReportLevel:   "summary",
		},
		Profiles:      make(map[string]*ProfileConfig),
		Environments:  make(map[string]*EnvironmentConfig),
		ConfigVersion: "1.9.0",
		autoSave:      true,
	}
}

func LoadIntegratedConfig(configPath string) (*IntegratedConfig, error) {
	config := NewIntegratedConfig()
	config.configPath = configPath

	if err := config.loadFromFile(); err != nil {
		if os.IsNotExist(err) {
			if err := config.createDefaultConfig(); err != nil {
				return nil, fmt.Errorf("デフォルト設定作成に失敗: %w", err)
			}
		} else {
			return nil, fmt.Errorf("設定ファイル読み込みに失敗: %w", err)
		}
	}

	config.applyEnvironmentOverrides()

	if err := config.applyProfile(config.General.Profile); err != nil {
		return nil, fmt.Errorf("プロファイル適用に失敗: %w", err)
	}

	return config, nil
}

func (ic *IntegratedConfig) loadFromFile() error {
	if ic.configPath == "" {
		return fmt.Errorf("設定ファイルパスが指定されていません")
	}

	cfg, err := ini.Load(ic.configPath)
	if err != nil {
		return err
	}

	if err := cfg.Section("general").MapTo(ic.General); err != nil {
		return fmt.Errorf("general セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("transform").MapTo(ic.Transform); err != nil {
		return fmt.Errorf("transform セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("validation").MapTo(ic.Validation); err != nil {
		return fmt.Errorf("validation セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("error_feedback").MapTo(ic.ErrorFeedback); err != nil {
		return fmt.Errorf("error_feedback セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("help_system").MapTo(ic.HelpSystem); err != nil {
		return fmt.Errorf("help_system セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("performance").MapTo(ic.Performance); err != nil {
		return fmt.Errorf("performance セクションの読み込みに失敗: %w", err)
	}

	if err := cfg.Section("output").MapTo(ic.Output); err != nil {
		return fmt.Errorf("output セクションの読み込みに失敗: %w", err)
	}

	if err := ic.loadProfiles(cfg); err != nil {
		return fmt.Errorf("プロファイル読み込みに失敗: %w", err)
	}

	if err := ic.loadEnvironments(cfg); err != nil {
		return fmt.Errorf("環境設定読み込みに失敗: %w", err)
	}

	stat, err := os.Stat(ic.configPath)
	if err == nil {
		ic.LastModified = stat.ModTime()
	}

	return nil
}

func (ic *IntegratedConfig) loadProfiles(cfg *ini.File) error {
	for _, sectionName := range cfg.SectionStrings() {
		if strings.HasPrefix(sectionName, "profiles.") {
			profileName := strings.TrimPrefix(sectionName, "profiles.")
			section := cfg.Section(sectionName)

			profile := &ProfileConfig{
				Name:      profileName,
				Overrides: make(map[string]interface{}),
			}

			if description := section.Key("description").String(); description != "" {
				profile.Description = description
			}
			if basedOn := section.Key("based_on").String(); basedOn != "" {
				profile.BasedOn = basedOn
			}

			for _, key := range section.Keys() {
				if key.Name() != "description" && key.Name() != "based_on" {
					profile.Overrides[key.Name()] = key.String()
				}
			}

			ic.Profiles[profileName] = profile
		}
	}
	return nil
}

func (ic *IntegratedConfig) loadEnvironments(cfg *ini.File) error {
	for _, sectionName := range cfg.SectionStrings() {
		if strings.HasPrefix(sectionName, "environments.") {
			envName := strings.TrimPrefix(sectionName, "environments.")
			env := &EnvironmentConfig{
				Name:      envName,
				Overrides: make(map[string]interface{}),
			}

			section := cfg.Section(sectionName)
			if err := section.MapTo(env); err != nil {
				return fmt.Errorf("環境 %s の読み込みに失敗: %w", envName, err)
			}

			ic.Environments[envName] = env
		}
	}
	return nil
}

func (ic *IntegratedConfig) createDefaultConfig() error {
	ic.createDefaultProfiles()
	ic.createDefaultEnvironments()

	if err := ic.Save(); err != nil {
		return fmt.Errorf("デフォルト設定の保存に失敗: %w", err)
	}

	return nil
}

func (ic *IntegratedConfig) createDefaultProfiles() {
	ic.Profiles["default"] = &ProfileConfig{
		Name:        "default",
		Description: "標準設定",
		BasedOn:     "",
		Overrides:   make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}

	ic.Profiles["beginner"] = &ProfileConfig{
		Name:        "beginner",
		Description: "初心者向け設定",
		BasedOn:     "default",
		Overrides: map[string]interface{}{
			"interactive_by_default":  "true",
			"verbose":                 "true",
			"skill_level":             "beginner",
			"max_suggestions":         "8",
			"enable_interactive_help": "true",
			"show_common_mistakes":    "true",
		},
		CreatedAt: time.Now(),
	}

	ic.Profiles["expert"] = &ProfileConfig{
		Name:        "expert",
		Description: "エキスパート向け設定",
		BasedOn:     "default",
		Overrides: map[string]interface{}{
			"strict_mode":         "true",
			"max_suggestions":     "3",
			"parallel_processing": "true",
			"show_progress":       "false",
			"report_level":        "minimal",
		},
		CreatedAt: time.Now(),
	}

	ic.Profiles["ci"] = &ProfileConfig{
		Name:        "ci",
		Description: "CI/CD環境向け設定",
		BasedOn:     "expert",
		Overrides: map[string]interface{}{
			"color_output":           "false",
			"verbose":                "false",
			"interactive_by_default": "false",
			"show_progress":          "false",
			"report_level":           "detailed",
		},
		CreatedAt: time.Now(),
	}
}

func (ic *IntegratedConfig) createDefaultEnvironments() {
	ic.Environments["development"] = &EnvironmentConfig{
		Name:              "development",
		SakuraAPIEndpoint: "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/",
		TimeoutSeconds:    30,
		RetryCount:        3,
		StrictMode:        false,
		Overrides:         make(map[string]interface{}),
	}

	ic.Environments["production"] = &EnvironmentConfig{
		Name:              "production",
		SakuraAPIEndpoint: "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/",
		TimeoutSeconds:    60,
		RetryCount:        5,
		StrictMode:        true,
		Overrides:         make(map[string]interface{}),
	}
}

func (ic *IntegratedConfig) applyEnvironmentOverrides() {
	envMappings := map[string]func(string){
		"USACLOUD_UPDATE_PROFILE": func(v string) {
			ic.General.Profile = v
		},
		"USACLOUD_UPDATE_STRICT_MODE": func(v string) {
			ic.Validation.StrictMode = (v == "true" || v == "1")
		},
		"USACLOUD_UPDATE_PARALLEL": func(v string) {
			ic.Performance.ParallelProcessing = (v == "true" || v == "1")
		},
		"USACLOUD_UPDATE_COLOR": func(v string) {
			ic.General.ColorOutput = (v == "true" || v == "1")
		},
		"USACLOUD_UPDATE_VERBOSE": func(v string) {
			ic.General.Verbose = (v == "true" || v == "1")
		},
	}

	for envVar, setter := range envMappings {
		if value := os.Getenv(envVar); value != "" {
			setter(value)
		}
	}
}

func (ic *IntegratedConfig) applyProfile(profileName string) error {
	if profileName == "" {
		return nil
	}

	profile, exists := ic.Profiles[profileName]
	if !exists {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	baseProfile := profile
	if profile.BasedOn != "" {
		var exists bool
		baseProfile, exists = ic.Profiles[profile.BasedOn]
		if !exists {
			return fmt.Errorf("ベースプロファイル '%s' が見つかりません", profile.BasedOn)
		}
	}

	if baseProfile != profile && baseProfile.Overrides != nil {
		for key, value := range baseProfile.Overrides {
			ic.applyOverride(key, value)
		}
	}

	for key, value := range profile.Overrides {
		ic.applyOverride(key, value)
	}

	ic.profileName = profileName
	return nil
}

func (ic *IntegratedConfig) applyOverride(key string, value interface{}) {
	strValue := fmt.Sprintf("%v", value)

	switch key {
	case "interactive_by_default":
		ic.General.InteractiveByDefault = (strValue == "true")
	case "verbose":
		ic.General.Verbose = (strValue == "true")
	case "color_output":
		ic.General.ColorOutput = (strValue == "true")
	case "skill_level":
		ic.HelpSystem.SkillLevel = strValue
	case "max_suggestions":
		if intVal, err := strconv.Atoi(strValue); err == nil {
			ic.Validation.MaxSuggestions = intVal
		}
	case "enable_interactive_help":
		ic.HelpSystem.EnableInteractiveHelp = (strValue == "true")
	case "show_common_mistakes":
		ic.HelpSystem.ShowCommonMistakes = (strValue == "true")
	case "strict_mode":
		ic.Validation.StrictMode = (strValue == "true")
	case "parallel_processing":
		ic.Performance.ParallelProcessing = (strValue == "true")
	case "show_progress":
		ic.Output.ShowProgress = (strValue == "true")
	case "report_level":
		ic.Output.ReportLevel = strValue
	}
}

func (ic *IntegratedConfig) Save() error {
	return ic.SaveAs(ic.configPath)
}

func (ic *IntegratedConfig) SaveAs(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("設定ファイルパスが指定されていません")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("設定ディレクトリ作成に失敗: %w", err)
	}

	cfg := ini.Empty()

	generalSec, err := cfg.NewSection("general")
	if err != nil {
		return err
	}
	ic.writeStructToSection(generalSec, ic.General)

	transformSec, err := cfg.NewSection("transform")
	if err != nil {
		return err
	}
	ic.writeStructToSection(transformSec, ic.Transform)

	validationSec, err := cfg.NewSection("validation")
	if err != nil {
		return err
	}
	ic.writeStructToSection(validationSec, ic.Validation)

	errorFeedbackSec, err := cfg.NewSection("error_feedback")
	if err != nil {
		return err
	}
	ic.writeStructToSection(errorFeedbackSec, ic.ErrorFeedback)

	helpSystemSec, err := cfg.NewSection("help_system")
	if err != nil {
		return err
	}
	ic.writeStructToSection(helpSystemSec, ic.HelpSystem)

	performanceSec, err := cfg.NewSection("performance")
	if err != nil {
		return err
	}
	ic.writeStructToSection(performanceSec, ic.Performance)

	outputSec, err := cfg.NewSection("output")
	if err != nil {
		return err
	}
	ic.writeStructToSection(outputSec, ic.Output)

	for name, profile := range ic.Profiles {
		sectionName := "profiles." + name
		profileSec, err := cfg.NewSection(sectionName)
		if err != nil {
			return err
		}

		if profile.Description != "" {
			profileSec.Key("description").SetValue(profile.Description)
		}
		if profile.BasedOn != "" {
			profileSec.Key("based_on").SetValue(profile.BasedOn)
		}

		for key, value := range profile.Overrides {
			profileSec.Key(key).SetValue(fmt.Sprintf("%v", value))
		}
	}

	for name, env := range ic.Environments {
		sectionName := "environments." + name
		envSec, err := cfg.NewSection(sectionName)
		if err != nil {
			return err
		}
		ic.writeStructToSection(envSec, env)
	}

	if err := cfg.SaveTo(configPath); err != nil {
		return fmt.Errorf("設定ファイル保存に失敗: %w", err)
	}

	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("設定ファイル権限設定に失敗: %w", err)
	}

	ic.LastModified = time.Now()
	return nil
}

func (ic *IntegratedConfig) writeStructToSection(section *ini.Section, data interface{}) {
	switch v := data.(type) {
	case *GeneralConfig:
		section.Key("version").SetValue(v.Version)
		section.Key("color_output").SetValue(fmt.Sprintf("%t", v.ColorOutput))
		section.Key("language").SetValue(v.Language)
		section.Key("verbose").SetValue(fmt.Sprintf("%t", v.Verbose))
		section.Key("interactive_by_default").SetValue(fmt.Sprintf("%t", v.InteractiveByDefault))
		section.Key("profile").SetValue(v.Profile)
	case *TransformConfig:
		section.Key("preserve_comments").SetValue(fmt.Sprintf("%t", v.PreserveComments))
		section.Key("add_explanatory_comments").SetValue(fmt.Sprintf("%t", v.AddExplanatoryComments))
		section.Key("show_line_numbers").SetValue(fmt.Sprintf("%t", v.ShowLineNumbers))
		section.Key("backup_original").SetValue(fmt.Sprintf("%t", v.BackupOriginal))
	case *ValidationConfig:
		section.Key("enable_validation").SetValue(fmt.Sprintf("%t", v.EnableValidation))
		section.Key("strict_mode").SetValue(fmt.Sprintf("%t", v.StrictMode))
		section.Key("validate_before_transform").SetValue(fmt.Sprintf("%t", v.ValidateBeforeTransform))
		section.Key("validate_after_transform").SetValue(fmt.Sprintf("%t", v.ValidateAfterTransform))
		section.Key("max_suggestions").SetValue(fmt.Sprintf("%d", v.MaxSuggestions))
		section.Key("max_edit_distance").SetValue(fmt.Sprintf("%d", v.MaxEditDistance))
		section.Key("skip_deprecated_warnings").SetValue(fmt.Sprintf("%t", v.SkipDeprecatedWarnings))
		section.Key("typo_detection_enabled").SetValue(fmt.Sprintf("%t", v.TypoDetectionEnabled))
	case *ErrorFeedbackConfig:
		section.Key("error_format").SetValue(v.ErrorFormat)
		section.Key("show_suggestions").SetValue(fmt.Sprintf("%t", v.ShowSuggestions))
		section.Key("show_alternatives").SetValue(fmt.Sprintf("%t", v.ShowAlternatives))
		section.Key("show_migration_guide").SetValue(fmt.Sprintf("%t", v.ShowMigrationGuide))
		section.Key("suggestion_confidence_threshold").SetValue(fmt.Sprintf("%.2f", v.SuggestionConfidenceThreshold))
	case *HelpSystemConfig:
		section.Key("enable_interactive_help").SetValue(fmt.Sprintf("%t", v.EnableInteractiveHelp))
		section.Key("skill_level").SetValue(v.SkillLevel)
		section.Key("preferred_help_format").SetValue(v.PreferredHelpFormat)
		section.Key("show_common_mistakes").SetValue(fmt.Sprintf("%t", v.ShowCommonMistakes))
		section.Key("enable_learning_tracking").SetValue(fmt.Sprintf("%t", v.EnableLearningTracking))
	case *PerformanceConfig:
		section.Key("parallel_processing").SetValue(fmt.Sprintf("%t", v.ParallelProcessing))
		section.Key("cache_enabled").SetValue(fmt.Sprintf("%t", v.CacheEnabled))
		section.Key("cache_size_mb").SetValue(fmt.Sprintf("%d", v.CacheSizeMB))
		section.Key("batch_size").SetValue(fmt.Sprintf("%d", v.BatchSize))
		section.Key("worker_count").SetValue(fmt.Sprintf("%d", v.WorkerCount))
	case *OutputConfig:
		section.Key("format").SetValue(v.Format)
		section.Key("show_progress").SetValue(fmt.Sprintf("%t", v.ShowProgress))
		section.Key("progress_style").SetValue(v.ProgressStyle)
		section.Key("report_level").SetValue(v.ReportLevel)
	case *EnvironmentConfig:
		section.Key("sakura_api_endpoint").SetValue(v.SakuraAPIEndpoint)
		section.Key("timeout_seconds").SetValue(fmt.Sprintf("%d", v.TimeoutSeconds))
		section.Key("retry_count").SetValue(fmt.Sprintf("%d", v.RetryCount))
		section.Key("strict_mode").SetValue(fmt.Sprintf("%t", v.StrictMode))
	}
}

func (ic *IntegratedConfig) UpdateSetting(sectionName, key string, value interface{}) error {
	oldValue := ic.getSetting(sectionName, key)

	if err := ic.setSetting(sectionName, key, value); err != nil {
		return err
	}

	event := ConfigChangeEvent{
		Section:   sectionName,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
	}

	ic.notifyConfigChange(event)

	if ic.autoSave {
		return ic.Save()
	}

	return nil
}

func (ic *IntegratedConfig) getSetting(section, key string) interface{} {
	return nil
}

func (ic *IntegratedConfig) setSetting(section, key string, value interface{}) error {
	return nil
}

func (ic *IntegratedConfig) notifyConfigChange(event ConfigChangeEvent) {
	for _, watcher := range ic.watchers {
		select {
		case watcher <- event:
		default:
		}
	}
}
