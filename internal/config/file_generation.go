package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileGenerator handles configuration file generation and management
type FileGenerator struct {
	envDetector *EnvDetector
}

// NewFileGenerator creates a new configuration file generator
func NewFileGenerator() *FileGenerator {
	return &FileGenerator{
		envDetector: NewEnvDetector(),
	}
}

// EnsureConfigDirectory creates the configuration directory if it doesn't exist
func (f *FileGenerator) EnsureConfigDirectory(configPath string) error {
	configDir := filepath.Dir(configPath)

	// Check if directory already exists
	if _, err := os.Stat(configDir); err == nil {
		return nil // Directory already exists
	}

	// Create directory with proper permissions
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	return nil
}

// GenerateConfigFromEnvVars generates a configuration file from environment variables
func (f *FileGenerator) GenerateConfigFromEnvVars(envVars *UsacloudEnvVars, configPath string) error {
	// Ensure config directory exists
	if err := f.EnsureConfigDirectory(configPath); err != nil {
		return err
	}

	// Generate config content
	configContent := f.envDetector.GenerateConfigContent(envVars)

	// Write config file
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		return fmt.Errorf("設定ファイルの作成に失敗しました: %w", err)
	}

	return nil
}

// SetProperFilePermissions sets appropriate permissions for the config file
func (f *FileGenerator) SetProperFilePermissions(filePath string) error {
	err := os.Chmod(filePath, 0600)
	if err != nil {
		return fmt.Errorf("ファイル権限の設定に失敗しました: %w", err)
	}
	return nil
}

// PromptCreateFromEnvVars prompts the user to create config from environment variables
func (f *FileGenerator) PromptCreateFromEnvVars(envVars *UsacloudEnvVars) bool {
	fmt.Println()
	fmt.Println(f.envDetector.FormatEnvVarsDisplay(envVars))
	fmt.Println()
	fmt.Println("📝 環境変数から設定ファイルを自動生成しますか？")
	fmt.Print("   [y]es / [n]o: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// ShowCurrentConfig displays the current configuration file content
func (f *FileGenerator) ShowCurrentConfig(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み取りに失敗しました: %w", err)
	}

	fmt.Println("📂 現在の設定ファイル内容:")
	fmt.Println()
	fmt.Println(string(content))

	return nil
}

// CheckConsistencyWithEnvVars checks if config file is consistent with environment variables
func (f *FileGenerator) CheckConsistencyWithEnvVars(configPath string) {
	envVars, detected := f.envDetector.DetectUsacloudEnvVars()
	if !detected {
		fmt.Println("ℹ️  環境変数は設定されていません")
		return
	}

	fmt.Println()
	fmt.Println("🔍 環境変数との整合性チェック:")
	fmt.Println(f.envDetector.FormatEnvVarsDisplay(envVars))

	// Read current config and compare (simplified comparison)
	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("⚠️  設定ファイルの読み取りエラー: %v\n", err)
		return
	}

	configStr := string(content)

	// Simple consistency check
	if strings.Contains(configStr, envVars.AccessToken) {
		fmt.Println("✅ アクセストークンが一致しています")
	} else {
		fmt.Println("⚠️  アクセストークンが環境変数と異なります")
	}

	if strings.Contains(configStr, envVars.AccessTokenSecret) {
		fmt.Println("✅ アクセストークンシークレットが一致しています")
	} else {
		fmt.Println("⚠️  アクセストークンシークレットが環境変数と異なります")
	}
}

// ShowManualSetupGuide displays manual setup instructions with environment variable guidance
func (f *FileGenerator) ShowManualSetupGuide() {
	fmt.Println("📝 設定ファイルを作成するには:")
	fmt.Println("   usacloud-update --sandbox --in <script.sh>")
	fmt.Println("   または")
	fmt.Println("   cp usacloud-update.conf.sample ~/.config/usacloud-update/usacloud-update.conf")
	fmt.Println()
	fmt.Println("💡 より簡単な方法:")
	fmt.Println("   以下の環境変数を設定してから `usacloud-update config` を実行:")
	fmt.Println("   export SAKURACLOUD_ACCESS_TOKEN='your-access-token'")
	fmt.Println("   export SAKURACLOUD_ACCESS_TOKEN_SECRET='your-access-token-secret'")
	fmt.Println("   export SAKURACLOUD_ZONE='tk1v'  # オプショナル")
	fmt.Println()
	fmt.Println("📖 詳細: https://docs.usacloud.jp/usacloud/references/env/")
}
