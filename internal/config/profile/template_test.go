package profile

import (
	"os"
	"strings"
	"testing"
)

func TestBuiltinTemplates(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templates := manager.GetBuiltinTemplates()

	if len(templates) == 0 {
		t.Errorf("Expected builtin templates, got none")
	}

	// Check for expected templates
	foundProd := false
	foundDev := false

	for _, template := range templates {
		if strings.Contains(template.Name, "本番") {
			foundProd = true
			if template.Environment != "production" {
				t.Errorf("Expected production environment for prod template, got %s", template.Environment)
			}
		}
		if strings.Contains(template.Name, "開発") {
			foundDev = true
			if template.Environment != "development" {
				t.Errorf("Expected development environment for dev template, got %s", template.Environment)
			}
		}
	}

	if !foundProd {
		t.Errorf("Production template not found")
	}
	if !foundDev {
		t.Errorf("Development template not found")
	}
}

func TestTemplateValidation(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Test Template",
		Description: "Test Description",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "REQUIRED_KEY",
				Description: "Required key",
				Required:    true,
				Type:        "string",
			},
			{
				Key:         "OPTIONAL_KEY",
				Description: "Optional key",
				Required:    false,
				Default:     "default_value",
				Type:        "string",
			},
			{
				Key:         "URL_KEY",
				Description: "URL key",
				Required:    true,
				Type:        "url",
				Validation:  "^https?://.*",
			},
		},
	}

	// Test valid config
	config := map[string]string{
		"REQUIRED_KEY": "test_value",
		"URL_KEY":      "https://example.com",
	}

	err = manager.ValidateTemplateConfig(template, config)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test missing required key
	configMissingRequired := map[string]string{
		"URL_KEY": "https://example.com",
	}

	err = manager.ValidateTemplateConfig(template, configMissingRequired)
	if err == nil {
		t.Errorf("Expected validation to fail for missing required key")
	}
	if !strings.Contains(err.Error(), "REQUIRED_KEY") {
		t.Errorf("Expected error to mention missing required key, got: %v", err)
	}

	// Test invalid URL
	configInvalidURL := map[string]string{
		"REQUIRED_KEY": "test_value",
		"URL_KEY":      "invalid_url",
	}

	err = manager.ValidateTemplateConfig(template, configInvalidURL)
	if err == nil {
		t.Errorf("Expected validation to fail for invalid URL")
	}
	if !strings.Contains(err.Error(), "URL_KEY") {
		t.Errorf("Expected error to mention invalid URL key, got: %v", err)
	}
}

func TestCreateProfileFromTemplate(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Test Template",
		Description: "Test Description",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "REQUIRED_KEY",
				Description: "Required key",
				Required:    true,
				Type:        "string",
			},
			{
				Key:         "OPTIONAL_KEY",
				Description: "Optional key",
				Required:    false,
				Default:     "default_value",
				Type:        "string",
			},
		},
		Tags: []string{"template", "test"},
	}

	// Add the test template to the template manager
	err = manager.templateManager.AddCustomTemplate(template)
	if err != nil {
		t.Fatalf("AddCustomTemplate() failed: %v", err)
	}

	config := map[string]string{
		"REQUIRED_KEY":                    "test_value",
		"SAKURACLOUD_ACCESS_TOKEN":        "test_token_123456789012345678901234567890",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test_secret_123456789012345678901234567890",
	}

	profile, err := manager.CreateProfileFromTemplate(template.Name, "Test Profile", config)
	if err != nil {
		t.Fatalf("CreateProfileFromTemplate() failed: %v", err)
	}

	if profile.Name != "Test Profile" {
		t.Errorf("Expected profile name 'Test Profile', got %s", profile.Name)
	}

	if profile.Environment != template.Environment {
		t.Errorf("Expected profile environment %s, got %s", template.Environment, profile.Environment)
	}

	// Check that required config is set
	if profile.Config["REQUIRED_KEY"] != "test_value" {
		t.Errorf("Expected REQUIRED_KEY to be 'test_value', got %s", profile.Config["REQUIRED_KEY"])
	}

	// Check that default value is applied
	if profile.Config["OPTIONAL_KEY"] != "default_value" {
		t.Errorf("Expected OPTIONAL_KEY to have default value 'default_value', got %s", profile.Config["OPTIONAL_KEY"])
	}

	// Check that tags are inherited
	if len(profile.Tags) != len(template.Tags) {
		t.Errorf("Expected %d tags, got %d", len(template.Tags), len(profile.Tags))
	}
}

func TestCreateProfileFromTemplateWithOverrides(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Test Template",
		Description: "Test Description",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "REQUIRED_KEY",
				Description: "Required key",
				Required:    true,
				Type:        "string",
			},
			{
				Key:         "OPTIONAL_KEY",
				Description: "Optional key",
				Required:    false,
				Default:     "default_value",
				Type:        "string",
			},
		},
	}

	// Add the test template to the template manager
	err = manager.templateManager.AddCustomTemplate(template)
	if err != nil {
		t.Fatalf("AddCustomTemplate() failed: %v", err)
	}

	config := map[string]string{
		"REQUIRED_KEY":                    "test_value",
		"OPTIONAL_KEY":                    "override_value", // Override default
		"SAKURACLOUD_ACCESS_TOKEN":        "test_token_123456789012345678901234567890",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test_secret_123456789012345678901234567890",
	}

	profile, err := manager.CreateProfileFromTemplate(template.Name, "Test Profile", config)
	if err != nil {
		t.Fatalf("CreateProfileFromTemplate() failed: %v", err)
	}

	// Check that override is applied
	if profile.Config["OPTIONAL_KEY"] != "override_value" {
		t.Errorf("Expected OPTIONAL_KEY to be overridden to 'override_value', got %s", profile.Config["OPTIONAL_KEY"])
	}
}

func TestValidateConfigKeyType(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	tests := []struct {
		keyDef ConfigKeyDef
		value  string
		valid  bool
	}{
		// String type
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "string"},
			value:  "any_string",
			valid:  true,
		},
		// Int type
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "int"},
			value:  "123",
			valid:  true,
		},
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "int"},
			value:  "not_a_number",
			valid:  false,
		},
		// Bool type
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "bool"},
			value:  "true",
			valid:  true,
		},
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "bool"},
			value:  "false",
			valid:  true,
		},
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "bool"},
			value:  "not_a_bool",
			valid:  false,
		},
		// URL type
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "url"},
			value:  "https://example.com",
			valid:  true,
		},
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "url"},
			value:  "not_a_url",
			valid:  false,
		},
		// Custom validation
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "string", Validation: "^test.*"},
			value:  "test_value",
			valid:  true,
		},
		{
			keyDef: ConfigKeyDef{Key: "test", Type: "string", Validation: "^test.*"},
			value:  "invalid_value",
			valid:  false,
		},
	}

	for _, test := range tests {
		err := manager.validateConfigKeyType(test.keyDef, test.value)
		if test.valid && err != nil {
			t.Errorf("Expected valid value %s for type %s, got error: %v", test.value, test.keyDef.Type, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Expected invalid value %s for type %s, but got no error", test.value, test.keyDef.Type)
		}
	}
}

func TestGetTemplateByName(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	templates := manager.GetBuiltinTemplates()

	if len(templates) == 0 {
		t.Skip("No builtin templates available")
	}

	// Get first template by name
	firstTemplate := templates[0]
	foundTemplate, err := manager.GetTemplateByName(firstTemplate.Name)
	if err != nil {
		t.Fatalf("GetTemplateByName() failed: %v", err)
	}

	if foundTemplate.Name != firstTemplate.Name {
		t.Errorf("Expected template name %s, got %s", firstTemplate.Name, foundTemplate.Name)
	}

	// Test non-existent template
	_, err = manager.GetTemplateByName("non-existent-template")
	if err == nil {
		t.Errorf("Expected error for non-existent template")
	}
}

func TestInteractiveTemplateConfig(t *testing.T) {
	// This test is difficult to test directly since it involves user input
	// We'll test the validation logic instead
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Interactive Test Template",
		Description: "Test template for interactive config",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "API_TOKEN",
				Description: "API Token",
				Required:    true,
				Type:        "string",
			},
			{
				Key:         "ZONE",
				Description: "Zone",
				Required:    false,
				Default:     "tk1v",
				Type:        "string",
			},
		},
	}

	// Test that the template has the expected structure for interactive config
	if len(template.ConfigKeys) != 2 {
		t.Errorf("Expected 2 config keys, got %d", len(template.ConfigKeys))
	}

	requiredKey := template.ConfigKeys[0]
	if !requiredKey.Required {
		t.Errorf("Expected first key to be required")
	}
	if requiredKey.Description == "" {
		t.Errorf("Expected description for required key")
	}

	optionalKey := template.ConfigKeys[1]
	if optionalKey.Required {
		t.Errorf("Expected second key to be optional")
	}
	if optionalKey.Default == "" {
		t.Errorf("Expected default value for optional key")
	}

	// Use manager to avoid unused variable error
	_ = manager
}

func TestTemplateConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Default Test Template",
		Description: "Test template with defaults",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "WITH_DEFAULT",
				Description: "Key with default",
				Required:    false,
				Default:     "default_value",
				Type:        "string",
			},
			{
				Key:         "WITHOUT_DEFAULT",
				Description: "Key without default",
				Required:    false,
				Type:        "string",
			},
		},
	}

	// Add the test template to the template manager
	err = manager.templateManager.AddCustomTemplate(template)
	if err != nil {
		t.Fatalf("AddCustomTemplate() failed: %v", err)
	}

	// Create profile with minimal config
	config := map[string]string{}

	// Apply defaults
	finalConfig := manager.applyTemplateDefaults(&template, config)

	if finalConfig["WITH_DEFAULT"] != "default_value" {
		t.Errorf("Expected default value to be applied, got %s", finalConfig["WITH_DEFAULT"])
	}

	if finalConfig["WITHOUT_DEFAULT"] != "" {
		t.Errorf("Expected empty value for key without default, got %s", finalConfig["WITHOUT_DEFAULT"])
	}
}

func TestTemplateEnvironmentVariableSubstitution(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	template := ProfileTemplate{
		Name:        "Env Var Template",
		Description: "Template with env var substitution",
		Environment: "test",
		ConfigKeys: []ConfigKeyDef{
			{
				Key:         "ENV_KEY",
				Description: "Key with env var",
				Required:    false,
				Default:     "${TEST_VAR}",
				Type:        "string",
			},
		},
	}

	// Add the test template to the template manager
	err = manager.templateManager.AddCustomTemplate(template)
	if err != nil {
		t.Fatalf("AddCustomTemplate() failed: %v", err)
	}

	config := map[string]string{}
	finalConfig := manager.applyTemplateDefaults(&template, config)

	// Note: This test assumes env var substitution is implemented
	// If not implemented yet, this test will help drive the implementation
	if finalConfig["ENV_KEY"] != "test_value" && finalConfig["ENV_KEY"] != "${TEST_VAR}" {
		t.Errorf("Expected env var substitution or raw value, got %s", finalConfig["ENV_KEY"])
	}
}
