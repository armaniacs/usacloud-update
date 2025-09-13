package profile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TemplateManager manages profile templates
type TemplateManager struct {
	builtinTemplates []ProfileTemplate
	customTemplates  []ProfileTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		builtinTemplates: getBuiltinTemplates(),
		customTemplates:  []ProfileTemplate{},
	}
}

// GetBuiltinTemplates returns built-in profile templates
func getBuiltinTemplates() []ProfileTemplate {
	return []ProfileTemplate{
		{
			Name:        "Sakura Cloud 本番環境",
			Description: "本番環境用の標準設定テンプレート",
			Environment: EnvironmentProduction,
			ConfigKeys: []ConfigKeyDef{
				{
					Key:         ConfigKeyAccessToken,
					Description: "SakuraCloud APIアクセストークン",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9]{20,}$",
				},
				{
					Key:         ConfigKeyAccessTokenSecret,
					Description: "SakuraCloud APIアクセストークンシークレット",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9+/]{40,}={0,2}$",
				},
				{
					Key:         ConfigKeyZone,
					Description: "操作対象ゾーン",
					Required:    true,
					Default:     "tk1v",
					Type:        "string",
					Validation:  "^(tk1v|is1a|is1b|tk1a|tk1b)$",
				},
			},
			Tags: []string{"production", "sakura-cloud"},
		},
		{
			Name:        "開発・検証環境",
			Description: "開発・検証用の安全な設定テンプレート",
			Environment: EnvironmentDevelopment,
			ConfigKeys: []ConfigKeyDef{
				{
					Key:         ConfigKeyAccessToken,
					Description: "開発用APIアクセストークン",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9]{20,}$",
				},
				{
					Key:         ConfigKeyAccessTokenSecret,
					Description: "開発用APIアクセストークンシークレット",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9+/]{40,}={0,2}$",
				},
				{
					Key:         ConfigKeyZone,
					Description: "検証用ゾーン（tk1v推奨）",
					Required:    true,
					Default:     "tk1v",
					Type:        "string",
					Validation:  "^(tk1v|is1a|is1b|tk1a|tk1b)$",
				},
				{
					Key:         ConfigKeyDryRun,
					Description: "デフォルトでドライラン実行",
					Required:    false,
					Default:     "true",
					Type:        "bool",
				},
				{
					Key:         ConfigKeyBatchMode,
					Description: "バッチモードでの実行",
					Required:    false,
					Default:     "false",
					Type:        "bool",
				},
			},
			Tags: []string{"development", "safe"},
		},
		{
			Name:        "ステージング環境",
			Description: "ステージング環境用の設定テンプレート",
			Environment: EnvironmentStaging,
			ConfigKeys: []ConfigKeyDef{
				{
					Key:         ConfigKeyAccessToken,
					Description: "ステージング用APIアクセストークン",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9]{20,}$",
				},
				{
					Key:         ConfigKeyAccessTokenSecret,
					Description: "ステージング用APIアクセストークンシークレット",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9+/]{40,}={0,2}$",
				},
				{
					Key:         ConfigKeyZone,
					Description: "ステージング用ゾーン",
					Required:    true,
					Default:     "tk1v",
					Type:        "string",
					Validation:  "^(tk1v|is1a|is1b|tk1a|tk1b)$",
				},
				{
					Key:         ConfigKeyInteractive,
					Description: "インタラクティブモードを有効化",
					Required:    false,
					Default:     "true",
					Type:        "bool",
				},
			},
			Tags: []string{"staging", "pre-production"},
		},
		{
			Name:        "テスト環境",
			Description: "自動テスト用の設定テンプレート",
			Environment: EnvironmentTest,
			ConfigKeys: []ConfigKeyDef{
				{
					Key:         ConfigKeyAccessToken,
					Description: "テスト用APIアクセストークン",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9]{20,}$",
				},
				{
					Key:         ConfigKeyAccessTokenSecret,
					Description: "テスト用APIアクセストークンシークレット",
					Required:    true,
					Type:        "string",
					Validation:  "^[a-zA-Z0-9+/]{40,}={0,2}$",
				},
				{
					Key:         ConfigKeyZone,
					Description: "テスト用ゾーン（tk1v固定）",
					Required:    true,
					Default:     "tk1v",
					Type:        "string",
					Validation:  "^tk1v$",
				},
				{
					Key:         ConfigKeyDryRun,
					Description: "テストでは常にドライラン",
					Required:    true,
					Default:     "true",
					Type:        "bool",
				},
				{
					Key:         ConfigKeyBatchMode,
					Description: "テストはバッチモードで実行",
					Required:    true,
					Default:     "true",
					Type:        "bool",
				},
			},
			Tags: []string{"test", "automation", "ci-cd"},
		},
	}
}

// GetAllTemplates returns all available templates (builtin + custom)
func (tm *TemplateManager) GetAllTemplates() []ProfileTemplate {
	templates := make([]ProfileTemplate, 0, len(tm.builtinTemplates)+len(tm.customTemplates))
	templates = append(templates, tm.builtinTemplates...)
	templates = append(templates, tm.customTemplates...)
	return templates
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*ProfileTemplate, error) {
	for _, template := range tm.GetAllTemplates() {
		if template.Name == name {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("template not found: %s", name)
}

// CreateProfileFromTemplate creates a profile from a template
func (tm *TemplateManager) CreateProfileFromTemplate(templateName string, profileName string, config map[string]string) (*Profile, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// Validate required config keys
	if err := tm.validateTemplateConfig(template, config); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Apply defaults for missing config
	finalConfig := make(map[string]string)
	for _, keyDef := range template.ConfigKeys {
		if value, exists := config[keyDef.Key]; exists {
			finalConfig[keyDef.Key] = value
		} else if keyDef.Default != "" {
			finalConfig[keyDef.Key] = keyDef.Default
		}
	}

	// Add any additional config not in template
	for key, value := range config {
		if _, exists := finalConfig[key]; !exists {
			finalConfig[key] = value
		}
	}

	profile := &Profile{
		ID:          generateProfileID(),
		Name:        profileName,
		Description: fmt.Sprintf("Created from template: %s", template.Name),
		Environment: template.Environment,
		Config:      finalConfig,
		Tags:        append([]string{}, template.Tags...),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return profile, nil
}

// ValidateTemplateConfig validates config against a template
func (tm *TemplateManager) validateTemplateConfig(template *ProfileTemplate, config map[string]string) error {
	for _, keyDef := range template.ConfigKeys {
		value, exists := config[keyDef.Key]

		// Check required keys
		if keyDef.Required && (!exists || value == "") {
			return fmt.Errorf("required configuration key missing: %s (%s)", keyDef.Key, keyDef.Description)
		}

		// Skip validation if value is empty and not required
		if !exists || value == "" {
			continue
		}

		// Validate type
		if err := tm.validateConfigType(keyDef, value); err != nil {
			return fmt.Errorf("validation failed for key %s: %w", keyDef.Key, err)
		}

		// Validate pattern
		if keyDef.Validation != "" {
			if err := tm.validateConfigPattern(keyDef, value); err != nil {
				return fmt.Errorf("validation failed for key %s: %w", keyDef.Key, err)
			}
		}
	}

	return nil
}

// ValidateConfigType validates the type of a config value
func (tm *TemplateManager) validateConfigType(keyDef ConfigKeyDef, value string) error {
	switch keyDef.Type {
	case "string":
		// String is always valid
		return nil

	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value must be an integer")
		}
		return nil

	case "bool":
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			return fmt.Errorf("value must be 'true' or 'false'")
		}
		return nil

	case "url":
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("value must be a valid URL (http:// or https://)")
		}
		return nil

	default:
		return fmt.Errorf("unknown type: %s", keyDef.Type)
	}
}

// ValidateConfigPattern validates a config value against a regex pattern
func (tm *TemplateManager) validateConfigPattern(keyDef ConfigKeyDef, value string) error {
	regex, err := regexp.Compile(keyDef.Validation)
	if err != nil {
		return fmt.Errorf("invalid validation pattern: %w", err)
	}

	if !regex.MatchString(value) {
		return fmt.Errorf("value does not match required pattern")
	}

	return nil
}

// AddCustomTemplate adds a custom template
func (tm *TemplateManager) AddCustomTemplate(template ProfileTemplate) error {
	// Validate template
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Environment == "" {
		return fmt.Errorf("template environment is required")
	}

	// Check for duplicate names
	for _, existing := range tm.GetAllTemplates() {
		if existing.Name == template.Name {
			return fmt.Errorf("template with name '%s' already exists", template.Name)
		}
	}

	// Validate config key definitions
	for _, keyDef := range template.ConfigKeys {
		if keyDef.Key == "" {
			return fmt.Errorf("config key name is required")
		}

		if keyDef.Type == "" {
			keyDef.Type = "string" // Default to string
		}

		// Validate regex pattern if provided
		if keyDef.Validation != "" {
			if _, err := regexp.Compile(keyDef.Validation); err != nil {
				return fmt.Errorf("invalid validation pattern for key %s: %w", keyDef.Key, err)
			}
		}
	}

	tm.customTemplates = append(tm.customTemplates, template)
	return nil
}

// RemoveCustomTemplate removes a custom template
func (tm *TemplateManager) RemoveCustomTemplate(name string) error {
	for i, template := range tm.customTemplates {
		if template.Name == name {
			tm.customTemplates = append(tm.customTemplates[:i], tm.customTemplates[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("custom template not found: %s", name)
}

// GetTemplateByEnvironment returns templates for a specific environment
func (tm *TemplateManager) GetTemplateByEnvironment(environment string) []ProfileTemplate {
	var templates []ProfileTemplate
	for _, template := range tm.GetAllTemplates() {
		if template.Environment == environment {
			templates = append(templates, template)
		}
	}
	return templates
}

// GetRequiredConfigKeys returns required config keys for a template
func (tm *TemplateManager) GetRequiredConfigKeys(templateName string) ([]ConfigKeyDef, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	var required []ConfigKeyDef
	for _, keyDef := range template.ConfigKeys {
		if keyDef.Required {
			required = append(required, keyDef)
		}
	}

	return required, nil
}

// GetDefaultConfig returns default configuration for a template
func (tm *TemplateManager) GetDefaultConfig(templateName string) (map[string]string, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	for _, keyDef := range template.ConfigKeys {
		if keyDef.Default != "" {
			config[keyDef.Key] = keyDef.Default
		}
	}

	return config, nil
}

// InteractiveConfigPrompt provides interactive configuration for a template
type InteractiveConfigPrompt struct {
	template *ProfileTemplate
	config   map[string]string
}

// NewInteractiveConfigPrompt creates a new interactive config prompt
func NewInteractiveConfigPrompt(template *ProfileTemplate) *InteractiveConfigPrompt {
	return &InteractiveConfigPrompt{
		template: template,
		config:   make(map[string]string),
	}
}

// PromptForConfig prompts for all required and optional configuration
func (icp *InteractiveConfigPrompt) PromptForConfig() (map[string]string, error) {
	for _, keyDef := range icp.template.ConfigKeys {
		value, err := icp.promptForKey(keyDef)
		if err != nil {
			return nil, err
		}

		if value != "" {
			icp.config[keyDef.Key] = value
		}
	}

	return icp.config, nil
}

func (icp *InteractiveConfigPrompt) promptForKey(keyDef ConfigKeyDef) (string, error) {
	// This would be implemented with actual user input in a real CLI
	// For now, return default or empty
	return keyDef.Default, nil
}

// GetBuiltinTemplates returns all builtin templates
func (tm *TemplateManager) GetBuiltinTemplates() []ProfileTemplate {
	return tm.builtinTemplates
}
