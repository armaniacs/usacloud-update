package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// CreateInteractiveConfig creates a new configuration file through interactive prompts
func CreateInteractiveConfig() (*SandboxConfig, error) {
	config := DefaultConfig()
	reader := bufio.NewReader(os.Stdin)

	// Display welcome message
	fmt.Fprint(os.Stderr, color.CyanString("🔧 usacloud-update 初期設定\n"))
	fmt.Fprint(os.Stderr, "設定ファイルが見つかりません。対話式で設定を作成します。\n\n")

	// Get configuration path for display
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	fmt.Fprintf(os.Stderr, "設定ファイル保存先: %s\n\n", configPath)

	// Get API credentials
	fmt.Fprint(os.Stderr, color.YellowString("=== Sakura Cloud API 設定 ===\n"))
	fmt.Fprint(os.Stderr, "APIキーの取得方法: https://manual.sakura.ad.jp/cloud/api/apikey.html\n\n")

	// Access Token
	accessToken, err := promptString(reader, "APIアクセストークン", "", true)
	if err != nil {
		return nil, err
	}
	config.AccessToken = accessToken

	// Access Token Secret
	accessTokenSecret, err := promptPassword("APIアクセストークンシークレット")
	if err != nil {
		return nil, err
	}
	config.AccessTokenSecret = accessTokenSecret

	// Zone (with validation)
	fmt.Fprintf(os.Stderr, "\nゾーン設定 (サンドボックスでは %s 固定)\n", color.GreenString("tk1v"))
	zone, err := promptString(reader, "ゾーン", "tk1v", false)
	if err != nil {
		return nil, err
	}
	if zone != "tk1v" {
		fmt.Fprint(os.Stderr, color.YellowString("⚠️  サンドボックス環境では tk1v のみ使用可能です。tk1v に設定します。\n"))
		zone = "tk1v"
	}
	config.Zone = zone

	// Sandbox settings
	fmt.Fprint(os.Stderr, color.YellowString("\n=== サンドボックス設定 ===\n"))

	// Enabled
	enabled, err := promptBool(reader, "サンドボックス機能を有効にする", true)
	if err != nil {
		return nil, err
	}
	config.Enabled = enabled

	// Debug
	debug, err := promptBool(reader, "デバッグモードを有効にする", false)
	if err != nil {
		return nil, err
	}
	config.Debug = debug

	// Dry Run
	dryRun, err := promptBool(reader, "ドライランモード（実際の実行を行わない）", false)
	if err != nil {
		return nil, err
	}
	config.DryRun = dryRun

	// Interactive
	interactive, err := promptBool(reader, "インタラクティブTUIを使用する", true)
	if err != nil {
		return nil, err
	}
	config.Interactive = interactive

	// Timeout
	timeout, err := promptInt(reader, "タイムアウト（秒）", 30)
	if err != nil {
		return nil, err
	}
	config.Timeout = time.Duration(timeout) * time.Second

	// Confirmation
	fmt.Fprint(os.Stderr, color.YellowString("\n=== 設定確認 ===\n"))
	displayConfig(config)

	confirm, err := promptBool(reader, "\nこの設定で保存しますか", true)
	if err != nil {
		return nil, err
	}

	if !confirm {
		fmt.Fprint(os.Stderr, color.YellowString("設定をキャンセルしました。\n"))
		return nil, fmt.Errorf("configuration cancelled by user")
	}

	// Save configuration
	if err := config.SaveToFile(); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Fprint(os.Stderr, color.GreenString("✅ 設定ファイルを作成しました: %s\n"), configPath)
	fmt.Fprint(os.Stderr, "これで usacloud-update --sandbox が使用できます。\n\n")

	return config, nil
}

// promptString prompts for a string value
func promptString(reader *bufio.Reader, prompt, defaultValue string, required bool) (string, error) {
	for {
		if defaultValue != "" {
			fmt.Fprintf(os.Stderr, "%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Fprintf(os.Stderr, "%s: ", prompt)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			if defaultValue != "" {
				return defaultValue, nil
			}
			if required {
				fmt.Fprint(os.Stderr, color.RedString("この項目は必須です。入力してください。\n"))
				continue
			}
		}

		if input != "" {
			return input, nil
		}
	}
}

// promptPassword prompts for a password (hidden input)
func promptPassword(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", prompt)

	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Fprint(os.Stderr, "\n")

	passwordStr := strings.TrimSpace(string(password))
	if passwordStr == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return passwordStr, nil
}

// promptBool prompts for a boolean value
func promptBool(reader *bufio.Reader, prompt string, defaultValue bool) (bool, error) {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Fprintf(os.Stderr, "%s [y/n] [%s]: ", prompt, defaultStr)

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" {
			return defaultValue, nil
		}

		switch input {
		case "y", "yes", "true", "1":
			return true, nil
		case "n", "no", "false", "0":
			return false, nil
		default:
			fmt.Fprint(os.Stderr, color.RedString("y または n で答えてください。\n"))
		}
	}
}

// promptInt prompts for an integer value
func promptInt(reader *bufio.Reader, prompt string, defaultValue int) (int, error) {
	for {
		fmt.Fprintf(os.Stderr, "%s [%d]: ", prompt, defaultValue)

		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue, nil
		}

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("数値を入力してください。\n"))
			continue
		}

		if value <= 0 {
			fmt.Fprint(os.Stderr, color.RedString("正の数値を入力してください。\n"))
			continue
		}

		return value, nil
	}
}

// displayConfig displays the current configuration for confirmation
func displayConfig(config *SandboxConfig) {
	fmt.Fprintf(os.Stderr, "  APIアクセストークン: %s\n", maskString(config.AccessToken))
	fmt.Fprintf(os.Stderr, "  APIシークレット: %s\n", maskString(config.AccessTokenSecret))
	fmt.Fprintf(os.Stderr, "  ゾーン: %s\n", config.Zone)
	fmt.Fprintf(os.Stderr, "  サンドボックス有効: %t\n", config.Enabled)
	fmt.Fprintf(os.Stderr, "  デバッグモード: %t\n", config.Debug)
	fmt.Fprintf(os.Stderr, "  ドライランモード: %t\n", config.DryRun)
	fmt.Fprintf(os.Stderr, "  インタラクティブモード: %t\n", config.Interactive)
	fmt.Fprintf(os.Stderr, "  タイムアウト: %d秒\n", int(config.Timeout.Seconds()))
}

// maskString masks a string for display (shows first 4 and last 4 characters)
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// MigrateFromEnv prompts user to migrate from .env file
func MigrateFromEnv() (*SandboxConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprint(os.Stderr, color.CyanString("🔄 設定移行\n"))
	fmt.Fprint(os.Stderr, ".env ファイルが見つかりました。新しい設定ファイル形式に移行しますか？\n\n")

	migrate, err := promptBool(reader, "移行を実行する", true)
	if err != nil {
		return nil, err
	}

	if !migrate {
		fmt.Fprint(os.Stderr, color.YellowString("移行をスキップしました。\n"))
		return CreateInteractiveConfig()
	}

	// Load from .env
	config, err := LoadFromEnv()
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("⚠️  .env ファイルの読み込みに失敗しました: %v\n"), err)
		fmt.Fprint(os.Stderr, "対話式設定に切り替えます。\n\n")
		return CreateInteractiveConfig()
	}

	// Display migrated settings
	fmt.Fprint(os.Stderr, color.YellowString("=== 移行される設定 ===\n"))
	displayConfig(config)

	confirm, err := promptBool(reader, "\nこの設定で保存しますか", true)
	if err != nil {
		return nil, err
	}

	if !confirm {
		fmt.Fprint(os.Stderr, "対話式設定に切り替えます。\n\n")
		return CreateInteractiveConfig()
	}

	// Save configuration
	if err := config.SaveToFile(); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	configPath, _ := ConfigPath()
	fmt.Fprint(os.Stderr, color.GreenString("✅ 設定ファイルを作成しました: %s\n"), configPath)
	fmt.Fprint(os.Stderr, "これで usacloud-update --sandbox が使用できます。\n\n")

	// Ask about .env file cleanup
	cleanup, err := promptBool(reader, ".env ファイルを削除しますか", false)
	if err != nil {
		return config, nil // Don't fail if cleanup prompt fails
	}

	if cleanup {
		if err := os.Remove(".env"); err != nil {
			fmt.Fprint(os.Stderr, color.YellowString("⚠️  .env ファイルの削除に失敗しました: %v\n"), err)
		} else {
			fmt.Fprint(os.Stderr, color.GreenString("✅ .env ファイルを削除しました。\n"))
		}
	}

	return config, nil
}
