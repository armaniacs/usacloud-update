package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ConfigMigrator struct {
	fromVersion string
	toVersion   string
}

func NewConfigMigrator(fromVersion, toVersion string) *ConfigMigrator {
	return &ConfigMigrator{
		fromVersion: fromVersion,
		toVersion:   toVersion,
	}
}

func (cm *ConfigMigrator) MigrateConfig(configPath string) error {
	oldConfig, err := cm.loadOldConfig(configPath)
	if err != nil {
		return err
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102-150405")
	if err := cm.backupConfig(configPath, backupPath); err != nil {
		return fmt.Errorf("バックアップ作成に失敗: %w", err)
	}

	newConfig, err := cm.convertConfig(oldConfig)
	if err != nil {
		return fmt.Errorf("設定変換に失敗: %w", err)
	}

	if err := newConfig.SaveAs(configPath); err != nil {
		return fmt.Errorf("新設定保存に失敗: %w", err)
	}

	fmt.Printf("✅ 設定ファイルを v%s から v%s に更新しました\n", cm.fromVersion, cm.toVersion)
	fmt.Printf("   バックアップ: %s\n", backupPath)

	return nil
}

func (cm *ConfigMigrator) MigrateFromEnvFile(envPath, configPath string) error {
	fmt.Println("🔄 .envファイルから新設定形式への移行を開始します")

	envVars, err := cm.loadEnvFile(envPath)
	if err != nil {
		return err
	}

	config := NewIntegratedConfig()

	envMappings := map[string]func(string){
		"SAKURACLOUD_ACCESS_TOKEN": func(v string) {
		},
		"SAKURACLOUD_ACCESS_TOKEN_SECRET": func(v string) {
		},
		"USACLOUD_COLOR_OUTPUT": func(v string) {
			config.General.ColorOutput = (v == "true")
		},
		"USACLOUD_VERBOSE": func(v string) {
			config.General.Verbose = (v == "true")
		},
		"USACLOUD_INTERACTIVE": func(v string) {
			config.General.InteractiveByDefault = (v == "true")
		},
		"USACLOUD_STRICT_MODE": func(v string) {
			config.Validation.StrictMode = (v == "true")
		},
	}

	migratedCount := 0
	for envKey, value := range envVars {
		if mapper, exists := envMappings[envKey]; exists {
			mapper(value)
			migratedCount++
			fmt.Printf("  ✓ %s: %s\n", envKey, value)
		}
	}

	if migratedCount == 0 {
		fmt.Println("  💡 移行可能な設定項目が見つかりませんでした")
	} else {
		fmt.Printf("  📊 %d個の設定項目を移行しました\n", migratedCount)
	}

	if err := config.SaveAs(configPath); err != nil {
		return fmt.Errorf("設定保存に失敗: %w", err)
	}

	backupPath := envPath + ".migrated." + time.Now().Format("20060102-150405")
	if err := cm.backupConfig(envPath, backupPath); err == nil {
		fmt.Printf("  💾 元の.envファイルをバックアップしました: %s\n", backupPath)
	}

	fmt.Printf("✅ 新しい設定ファイルを作成しました: %s\n", configPath)
	return nil
}

func (cm *ConfigMigrator) loadOldConfig(configPath string) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			config[key] = value
		}
	}

	return config, scanner.Err()
}

func (cm *ConfigMigrator) loadEnvFile(envPath string) (map[string]string, error) {
	envVars := make(map[string]string)

	file, err := os.Open(envPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("  ⚠️  無効な行をスキップ (行 %d): %s\n", lineNumber, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, "\"'")

		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(".envファイル読み込みエラー: %w", err)
	}

	return envVars, nil
}

func (cm *ConfigMigrator) backupConfig(sourcePath, backupPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return err
	}

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer backupFile.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := sourceFile.Read(buf)
		if n > 0 {
			if _, writeErr := backupFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}

func (cm *ConfigMigrator) convertConfig(oldConfig map[string]interface{}) (*IntegratedConfig, error) {
	newConfig := NewIntegratedConfig()

	conversionMap := map[string]func(string){
		"color_output": func(v string) {
			newConfig.General.ColorOutput = (v == "true" || v == "1")
		},
		"verbose": func(v string) {
			newConfig.General.Verbose = (v == "true" || v == "1")
		},
		"language": func(v string) {
			newConfig.General.Language = v
		},
		"profile": func(v string) {
			newConfig.General.Profile = v
		},
		"preserve_comments": func(v string) {
			newConfig.Transform.PreserveComments = (v == "true" || v == "1")
		},
		"add_explanatory_comments": func(v string) {
			newConfig.Transform.AddExplanatoryComments = (v == "true" || v == "1")
		},
		"show_line_numbers": func(v string) {
			newConfig.Transform.ShowLineNumbers = (v == "true" || v == "1")
		},
		"backup_original": func(v string) {
			newConfig.Transform.BackupOriginal = (v == "true" || v == "1")
		},
		"enable_validation": func(v string) {
			newConfig.Validation.EnableValidation = (v == "true" || v == "1")
		},
		"strict_mode": func(v string) {
			newConfig.Validation.StrictMode = (v == "true" || v == "1")
		},
		"max_suggestions": func(v string) {
			if v == "3" {
				newConfig.Validation.MaxSuggestions = 3
			} else if v == "5" {
				newConfig.Validation.MaxSuggestions = 5
			} else if v == "8" {
				newConfig.Validation.MaxSuggestions = 8
			}
		},
	}

	for key, value := range oldConfig {
		strValue := fmt.Sprintf("%v", value)
		if converter, exists := conversionMap[key]; exists {
			converter(strValue)
		}
	}

	return newConfig, nil
}

func (cm *ConfigMigrator) HasEnvFile(envPath string) bool {
	if envPath == "" {
		return false
	}

	_, err := os.Stat(envPath)
	return err == nil
}

func (cm *ConfigMigrator) ShouldMigrate(configPath string) (bool, string, error) {
	if _, err := os.Stat(configPath); err == nil {
		return false, "", nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return false, "", err
	}

	envPaths := []string{
		filepath.Join(currentDir, ".env"),
		filepath.Join(currentDir, "usacloud-update.env"),
		filepath.Join(filepath.Dir(configPath), ".env"),
	}

	for _, envPath := range envPaths {
		if cm.HasEnvFile(envPath) {
			return true, envPath, nil
		}
	}

	return false, "", nil
}

func (cm *ConfigMigrator) GetMigrationSummary(envPath string) (MigrationSummary, error) {
	summary := MigrationSummary{
		EnvPath:             envPath,
		SupportedSettings:   make(map[string]string),
		UnsupportedSettings: make(map[string]string),
	}

	envVars, err := cm.loadEnvFile(envPath)
	if err != nil {
		return summary, err
	}

	supportedKeys := map[string]string{
		"SAKURACLOUD_ACCESS_TOKEN":        "認証トークン (環境変数で継続使用)",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET": "認証シークレット (環境変数で継続使用)",
		"USACLOUD_COLOR_OUTPUT":           "カラー出力設定",
		"USACLOUD_VERBOSE":                "詳細出力設定",
		"USACLOUD_INTERACTIVE":            "インタラクティブモード設定",
		"USACLOUD_STRICT_MODE":            "厳密モード設定",
	}

	for key, value := range envVars {
		if description, isSupported := supportedKeys[key]; isSupported {
			summary.SupportedSettings[key] = fmt.Sprintf("%s: %s", description, value)
		} else {
			summary.UnsupportedSettings[key] = value
		}
	}

	summary.TotalSettings = len(envVars)
	summary.SupportedCount = len(summary.SupportedSettings)
	summary.UnsupportedCount = len(summary.UnsupportedSettings)

	return summary, nil
}

type MigrationSummary struct {
	EnvPath             string
	TotalSettings       int
	SupportedCount      int
	UnsupportedCount    int
	SupportedSettings   map[string]string
	UnsupportedSettings map[string]string
}

func (ms MigrationSummary) PrintSummary() {
	fmt.Printf("📋 移行サマリー\n")
	fmt.Printf("===============\n")
	fmt.Printf("元ファイル: %s\n", ms.EnvPath)
	fmt.Printf("設定項目数: %d\n", ms.TotalSettings)
	fmt.Printf("移行可能:   %d\n", ms.SupportedCount)
	fmt.Printf("非対応:     %d\n", ms.UnsupportedCount)

	if len(ms.SupportedSettings) > 0 {
		fmt.Printf("\n✅ 移行可能な設定:\n")
		for key, description := range ms.SupportedSettings {
			fmt.Printf("  • %s: %s\n", key, description)
		}
	}

	if len(ms.UnsupportedSettings) > 0 {
		fmt.Printf("\n⚠️  非対応の設定:\n")
		for key, value := range ms.UnsupportedSettings {
			fmt.Printf("  • %s=%s (手動での設定が必要)\n", key, value)
		}
	}
}
