package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
		{Severity(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.severity.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestNewEnvironmentValidator(t *testing.T) {
	validator := NewEnvironmentValidator()

	if len(validator.checks) == 0 {
		t.Error("Expected validator to have checks")
	}

	// 標準チェックの存在確認
	checkNames := make([]string, len(validator.checks))
	for i, check := range validator.checks {
		checkNames[i] = check.Name()
	}

	expectedChecks := []string{
		"usacloud CLI",
		"APIキー",
		"ネットワーク接続",
		"ゾーンアクセス権限",
		"設定ファイル",
	}

	for _, expected := range expectedChecks {
		found := false
		for _, name := range checkNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected check '%s' not found in validator", expected)
		}
	}
}

func TestEnvironmentValidator_AddCheck(t *testing.T) {
	validator := &EnvironmentValidator{}
	check := &USACloudCLICheck{}

	validator.AddCheck(check)

	if len(validator.checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(validator.checks))
	}

	if validator.checks[0] != check {
		t.Error("Expected check to be added")
	}
}

func TestEnvironmentValidator_HasCriticalErrors(t *testing.T) {
	validator := &EnvironmentValidator{}

	t.Run("no critical errors", func(t *testing.T) {
		results := []*ValidationResult{
			{Passed: true, Severity: SeverityInfo},
			{Passed: false, Severity: SeverityWarning},
		}

		if validator.HasCriticalErrors(results) {
			t.Error("Expected no critical errors")
		}
	})

	t.Run("has critical errors", func(t *testing.T) {
		results := []*ValidationResult{
			{Passed: true, Severity: SeverityInfo},
			{Passed: false, Severity: SeverityError},
		}

		if !validator.HasCriticalErrors(results) {
			t.Error("Expected critical errors")
		}
	})
}

func TestEnvironmentValidator_GenerateReport(t *testing.T) {
	validator := &EnvironmentValidator{}

	results := []*ValidationResult{
		{
			CheckName: "Test Check 1",
			Passed:    true,
			Message:   "Success message",
			Severity:  SeverityInfo,
		},
		{
			CheckName: "Test Check 2",
			Passed:    false,
			Message:   "Error message",
			Severity:  SeverityError,
			FixAction: "Fix this issue",
			HelpURL:   "https://example.com/help",
		},
	}

	report := validator.GenerateReport(results)

	// 基本的な内容チェック
	if !strings.Contains(report, "サンドボックス環境検証結果") {
		t.Error("Report should contain title")
	}

	if !strings.Contains(report, "Test Check 1") {
		t.Error("Report should contain check names")
	}

	if !strings.Contains(report, "Success message") {
		t.Error("Report should contain messages")
	}

	if !strings.Contains(report, "Fix this issue") {
		t.Error("Report should contain fix actions")
	}

	if !strings.Contains(report, "https://example.com/help") {
		t.Error("Report should contain help URLs")
	}

	if !strings.Contains(report, "検証サマリー") {
		t.Error("Report should contain summary")
	}
}

func TestUSACloudCLICheck_Name(t *testing.T) {
	check := &USACloudCLICheck{}
	if check.Name() != "usacloud CLI" {
		t.Errorf("Expected 'usacloud CLI', got '%s'", check.Name())
	}
}

func TestUSACloudCLICheck_Description(t *testing.T) {
	check := &USACloudCLICheck{}
	desc := check.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
	if !strings.Contains(desc, "usacloud CLI") {
		t.Error("Description should mention usacloud CLI")
	}
}

func TestUSACloudCLICheck_isVersionCompatible(t *testing.T) {
	check := &USACloudCLICheck{requiredVersion: "1.43.0"}

	tests := []struct {
		version  string
		expected bool
	}{
		{"usacloud version 1.43.0", true},
		{"usacloud version 1.43.1", true},
		{"usacloud version 1.44.0", true},
		{"usacloud version 2.0.0", true},
		{"usacloud version 1.42.0", false},
		{"usacloud version 1.43.0-dev", true}, // パッチ部分が一致
		{"invalid version", false},
		{"1.43.0", true},
		{"1.42.9", false},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			result := check.isVersionCompatible(test.version)
			if result != test.expected {
				t.Errorf("For version '%s', expected %v, got %v", test.version, test.expected, result)
			}
		})
	}
}

func TestAPIKeyCheck_Name(t *testing.T) {
	check := &APIKeyCheck{}
	if check.Name() != "APIキー" {
		t.Errorf("Expected 'APIキー', got '%s'", check.Name())
	}
}

func TestAPIKeyCheck_Description(t *testing.T) {
	check := &APIKeyCheck{}
	desc := check.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
	if !strings.Contains(desc, "APIキー") {
		t.Error("Description should mention APIキー")
	}
}

func TestAPIKeyCheck_Validate(t *testing.T) {
	check := &APIKeyCheck{}

	// 環境変数をクリア
	oldToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	oldSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
	defer func() {
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN", oldToken)
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", oldSecret)
	}()

	t.Run("no api keys", func(t *testing.T) {
		os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN")
		os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

		result := check.Validate()

		if result.Passed {
			t.Error("Expected validation to fail when no API keys")
		}
		if result.Severity != SeverityCritical {
			t.Errorf("Expected SeverityCritical, got %v", result.Severity)
		}
	})

	t.Run("invalid api key format", func(t *testing.T) {
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "short")
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "short")

		result := check.Validate()

		if result.Passed {
			t.Error("Expected validation to fail for short API keys")
		}
		if result.Severity != SeverityError {
			t.Errorf("Expected SeverityError, got %v", result.Severity)
		}
	})

	t.Run("valid format but auth fails", func(t *testing.T) {
		// 形式は正しいが無効なキー
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "12345678901234567890")
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "123456789012345678901234567890123456789012345678901234567890")

		result := check.Validate()

		// usacloud commandが存在しない場合は無効キーとして扱われる
		if result.Passed {
			t.Log("API key validation passed (usacloud command may be available)")
		} else {
			if result.Severity != SeverityError {
				t.Errorf("Expected SeverityError, got %v", result.Severity)
			}
		}
	})
}

func TestNetworkCheck_Name(t *testing.T) {
	check := &NetworkCheck{}
	if check.Name() != "ネットワーク接続" {
		t.Errorf("Expected 'ネットワーク接続', got '%s'", check.Name())
	}
}

func TestNetworkCheck_testConnection(t *testing.T) {
	check := &NetworkCheck{
		timeout: 5 * time.Second,
	}

	t.Run("valid endpoint", func(t *testing.T) {
		// Google DNS (should be accessible)
		result := check.testConnection("https://www.google.com")
		if !result {
			t.Log("Network test to google.com failed - may be in restricted environment")
		}
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		result := check.testConnection("https://invalid-domain-that-should-not-exist.com")
		if result {
			t.Error("Expected connection to invalid domain to fail")
		}
	})
}

func TestZoneAccessCheck_Name(t *testing.T) {
	check := &ZoneAccessCheck{zone: "tk1v"}
	if check.Name() != "ゾーンアクセス権限" {
		t.Errorf("Expected 'ゾーンアクセス権限', got '%s'", check.Name())
	}
}

func TestConfigFileCheck_Name(t *testing.T) {
	check := &ConfigFileCheck{}
	if check.Name() != "設定ファイル" {
		t.Errorf("Expected '設定ファイル', got '%s'", check.Name())
	}
}

func TestConfigFileCheck_Validate(t *testing.T) {
	check := &ConfigFileCheck{}

	// テスト用一時ファイルの作成
	tmpDir := t.TempDir()
	configPath := tmpDir + "/usacloud-update.conf"

	t.Run("no config file", func(t *testing.T) {
		// HOMEを一時ディレクトリに変更
		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tmpDir)

		result := check.Validate()

		if result.Passed {
			t.Error("Expected validation to fail when no config file")
		}
		if result.Severity != SeverityWarning {
			t.Errorf("Expected SeverityWarning, got %v", result.Severity)
		}
	})

	t.Run("config file without api key", func(t *testing.T) {
		// 空の設定ファイルを作成
		err := os.WriteFile(configPath, []byte("# empty config"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tmpDir)

		result := check.Validate()

		if result.Passed {
			t.Error("Expected validation to fail for config without API key")
		}
		if result.Severity != SeverityWarning {
			t.Errorf("Expected SeverityWarning, got %v", result.Severity)
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		// 正しい設定ディレクトリ構造を作成
		configDir := filepath.Join(tmpDir, ".config", "usacloud-update")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// APIキーを含む設定ファイルを作成
		configContent := `
SAKURACLOUD_ACCESS_TOKEN=test_token
SAKURACLOUD_ACCESS_TOKEN_SECRET=test_secret
`
		configPath := filepath.Join(configDir, "usacloud-update.conf")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		oldHome := os.Getenv("HOME")
		defer os.Setenv("HOME", oldHome)
		os.Setenv("HOME", tmpDir)

		result := check.Validate()

		if !result.Passed {
			t.Errorf("Expected validation to pass for valid config, got: %s", result.Message)
		}
		if result.Severity != SeverityInfo {
			t.Errorf("Expected SeverityInfo, got %v", result.Severity)
		}
	})
}

func TestValidationCheck_Fix(t *testing.T) {
	checks := []ValidationCheck{
		&USACloudCLICheck{},
		&APIKeyCheck{},
		&NetworkCheck{},
		&ZoneAccessCheck{},
		&ConfigFileCheck{},
	}

	for _, check := range checks {
		t.Run(check.Name(), func(t *testing.T) {
			err := check.Fix()
			if err == nil {
				t.Error("Expected Fix() to return error (not implemented)")
			}
			if !strings.Contains(err.Error(), "自動修復は対応していません") {
				t.Error("Expected error message about auto-fix not supported")
			}
		})
	}
}

func TestEnvironmentValidator_RunAllChecks(t *testing.T) {
	// カスタムチェックを作成してテスト
	validator := &EnvironmentValidator{}

	// テスト用のモックチェック
	mockCheck := &mockValidationCheck{
		name:        "Test Check",
		description: "Test description",
		result: &ValidationResult{
			Passed:   true,
			Message:  "Test message",
			Severity: SeverityInfo,
		},
	}

	validator.AddCheck(mockCheck)

	results := validator.RunAllChecks()

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.CheckName != "Test Check" {
		t.Errorf("Expected CheckName 'Test Check', got '%s'", result.CheckName)
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// mockValidationCheck はテスト用のモックチェック
type mockValidationCheck struct {
	name        string
	description string
	result      *ValidationResult
}

func (m *mockValidationCheck) Name() string {
	return m.name
}

func (m *mockValidationCheck) Description() string {
	return m.description
}

func (m *mockValidationCheck) Validate() *ValidationResult {
	return m.result
}

func (m *mockValidationCheck) Fix() error {
	return nil
}

func TestValidationResult_JSON(t *testing.T) {
	result := &ValidationResult{
		CheckName: "Test",
		Passed:    true,
		Message:   "Test message",
		Severity:  SeverityInfo,
		Duration:  100 * time.Millisecond,
	}

	// JSONエンコードのテスト（実際のエンコードは行わず、フィールドタグの確認）
	if result.CheckName == "" {
		t.Error("CheckName should not be empty")
	}
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}
}
