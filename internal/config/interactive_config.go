package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type InteractiveConfigManager struct {
	config *IntegratedConfig
	ui     *ConfigUI
}

type ConfigUI struct {
	reader *bufio.Reader
}

func NewInteractiveConfigManager(config *IntegratedConfig) *InteractiveConfigManager {
	return &InteractiveConfigManager{
		config: config,
		ui: &ConfigUI{
			reader: bufio.NewReader(os.Stdin),
		},
	}
}

func (icm *InteractiveConfigManager) RunInteractiveSetup() error {
	fmt.Println("🚀 usacloud-update 設定セットアップ")
	fmt.Println("==================================")
	fmt.Println()

	if err := icm.setupBasicConfig(); err != nil {
		return err
	}

	if err := icm.selectProfile(); err != nil {
		return err
	}

	if err := icm.setupValidationConfig(); err != nil {
		return err
	}

	if err := icm.setupOutputConfig(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("💾 設定を保存しています...")
	if err := icm.config.Save(); err != nil {
		return fmt.Errorf("設定保存に失敗: %w", err)
	}

	fmt.Println("✅ 設定セットアップが完了しました！")
	fmt.Printf("   設定ファイル: %s\n", icm.config.configPath)

	return nil
}

func (icm *InteractiveConfigManager) setupBasicConfig() error {
	fmt.Println("📋 基本設定")
	fmt.Println("-----------")

	colorChoice := icm.ui.promptChoice(
		"カラー出力を有効にしますか？ [Y/n]",
		[]string{"y", "Y", "n", "N", ""},
	)
	icm.config.General.ColorOutput = (colorChoice == "" || strings.ToLower(colorChoice) == "y")

	verboseChoice := icm.ui.promptChoice(
		"詳細出力を有効にしますか？ [y/N]",
		[]string{"y", "Y", "n", "N", ""},
	)
	icm.config.General.Verbose = (strings.ToLower(verboseChoice) == "y")

	langChoice := icm.ui.promptChoice(
		"言語を選択してください [1: 日本語, 2: English]",
		[]string{"1", "2"},
	)
	if langChoice == "2" {
		icm.config.General.Language = "en"
	} else {
		icm.config.General.Language = "ja"
	}

	fmt.Println()
	return nil
}

func (icm *InteractiveConfigManager) selectProfile() error {
	fmt.Println("📋 プロファイル選択")
	fmt.Println("------------------")
	fmt.Println("   1. default    - 標準設定")
	fmt.Println("   2. beginner   - 初心者向け（丁寧なヘルプ）")
	fmt.Println("   3. expert     - エキスパート向け（最小出力）")
	fmt.Println("   4. ci         - CI/CD環境向け")
	fmt.Println("   5. custom     - カスタムプロファイル作成")

	choice := icm.ui.promptChoice("選択してください [1-5]", []string{"1", "2", "3", "4", "5"})

	profileMap := map[string]string{
		"1": "default",
		"2": "beginner",
		"3": "expert",
		"4": "ci",
	}

	if profileName, exists := profileMap[choice]; exists {
		icm.config.General.Profile = profileName
		fmt.Printf("✅ プロファイル '%s' を選択しました\n", profileName)
	} else if choice == "5" {
		return icm.createCustomProfile()
	}

	fmt.Println()
	return nil
}

func (icm *InteractiveConfigManager) createCustomProfile() error {
	fmt.Println()
	fmt.Println("🛠 カスタムプロファイル作成")
	fmt.Println("-------------------------")

	profileName := icm.ui.promptString("プロファイル名を入力してください")
	if profileName == "" {
		return fmt.Errorf("プロファイル名は必須です")
	}

	description := icm.ui.promptString("プロファイルの説明を入力してください（省略可）")

	fmt.Println("ベースとするプロファイルを選択してください:")
	fmt.Println("   1. default    - 標準設定")
	fmt.Println("   2. beginner   - 初心者向け")
	fmt.Println("   3. expert     - エキスパート向け")
	fmt.Println("   4. なし       - 空のプロファイル")

	baseChoice := icm.ui.promptChoice("選択してください [1-4]", []string{"1", "2", "3", "4"})

	baseProfileMap := map[string]string{
		"1": "default",
		"2": "beginner",
		"3": "expert",
		"4": "",
	}

	basedOn := baseProfileMap[baseChoice]

	pm := NewProfileManager(icm.config)
	if err := pm.CreateProfile(profileName, basedOn, description); err != nil {
		return fmt.Errorf("プロファイル作成に失敗: %w", err)
	}

	icm.config.General.Profile = profileName
	fmt.Printf("✅ カスタムプロファイル '%s' を作成・選択しました\n", profileName)

	return nil
}

func (icm *InteractiveConfigManager) setupValidationConfig() error {
	fmt.Println("🔍 検証設定")
	fmt.Println("-----------")

	enableChoice := icm.ui.promptChoice(
		"コマンド検証を有効にしますか？ [Y/n]",
		[]string{"y", "Y", "n", "N", ""},
	)
	icm.config.Validation.EnableValidation = (enableChoice == "" || strings.ToLower(enableChoice) == "y")

	if icm.config.Validation.EnableValidation {
		strictChoice := icm.ui.promptChoice(
			"厳密モードを有効にしますか？ [y/N]",
			[]string{"y", "Y", "n", "N", ""},
		)
		icm.config.Validation.StrictMode = (strings.ToLower(strictChoice) == "y")

		fmt.Println("提案する類似コマンドの最大数を選択してください:")
		fmt.Println("   1. 3個（最小）")
		fmt.Println("   2. 5個（標準）")
		fmt.Println("   3. 8個（最大）")

		suggestionChoice := icm.ui.promptChoice("選択してください [1-3]", []string{"1", "2", "3"})
		suggestionMap := map[string]int{
			"1": 3,
			"2": 5,
			"3": 8,
		}
		icm.config.Validation.MaxSuggestions = suggestionMap[suggestionChoice]

		typoChoice := icm.ui.promptChoice(
			"タイポ検出を有効にしますか？ [Y/n]",
			[]string{"y", "Y", "n", "N", ""},
		)
		icm.config.Validation.TypoDetectionEnabled = (typoChoice == "" || strings.ToLower(typoChoice) == "y")
	}

	fmt.Println()
	return nil
}

func (icm *InteractiveConfigManager) setupOutputConfig() error {
	fmt.Println("📤 出力設定")
	fmt.Println("-----------")

	fmt.Println("出力形式を選択してください:")
	fmt.Println("   1. auto       - 自動（ターミナルに応じて）")
	fmt.Println("   2. plain      - プレーンテキスト")
	fmt.Println("   3. colored    - カラー出力")
	fmt.Println("   4. json       - JSON形式")

	formatChoice := icm.ui.promptChoice("選択してください [1-4]", []string{"1", "2", "3", "4"})
	formatMap := map[string]string{
		"1": "auto",
		"2": "plain",
		"3": "colored",
		"4": "json",
	}
	icm.config.Output.Format = formatMap[formatChoice]

	progressChoice := icm.ui.promptChoice(
		"進捗表示を有効にしますか？ [Y/n]",
		[]string{"y", "Y", "n", "N", ""},
	)
	icm.config.Output.ShowProgress = (progressChoice == "" || strings.ToLower(progressChoice) == "y")

	if icm.config.Output.ShowProgress {
		fmt.Println("進捗表示スタイルを選択してください:")
		fmt.Println("   1. bar        - プログレスバー")
		fmt.Println("   2. percentage - パーセンテージ")
		fmt.Println("   3. dots       - ドット")

		styleChoice := icm.ui.promptChoice("選択してください [1-3]", []string{"1", "2", "3"})
		styleMap := map[string]string{
			"1": "bar",
			"2": "percentage",
			"3": "dots",
		}
		icm.config.Output.ProgressStyle = styleMap[styleChoice]
	}

	fmt.Println("レポートレベルを選択してください:")
	fmt.Println("   1. minimal    - 最小限")
	fmt.Println("   2. summary    - サマリー")
	fmt.Println("   3. detailed   - 詳細")

	reportChoice := icm.ui.promptChoice("選択してください [1-3]", []string{"1", "2", "3"})
	reportMap := map[string]string{
		"1": "minimal",
		"2": "summary",
		"3": "detailed",
	}
	icm.config.Output.ReportLevel = reportMap[reportChoice]

	fmt.Println()
	return nil
}

func (ui *ConfigUI) promptString(prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)
		input, err := ui.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("入力エラー: %v\n", err)
			continue
		}
		return strings.TrimSpace(input)
	}
}

func (ui *ConfigUI) promptChoice(prompt string, validChoices []string) string {
	for {
		fmt.Printf("%s: ", prompt)
		input, err := ui.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("入力エラー: %v\n", err)
			continue
		}

		choice := strings.TrimSpace(input)

		if len(validChoices) == 0 {
			return choice
		}

		for _, valid := range validChoices {
			if choice == valid {
				return choice
			}
		}

		fmt.Printf("無効な選択です。有効な選択肢: %v\n", validChoices)
	}
}

func (ui *ConfigUI) promptPassword(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)

	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()
	return string(bytePassword), nil
}

func (ui *ConfigUI) promptInt(prompt string, min, max int) int {
	for {
		input := ui.promptString(fmt.Sprintf("%s [%d-%d]", prompt, min, max))

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("数値を入力してください\n")
			continue
		}

		if value < min || value > max {
			fmt.Printf("値は %d から %d の間で入力してください\n", min, max)
			continue
		}

		return value
	}
}

func (ui *ConfigUI) promptConfirm(prompt string, defaultValue bool) bool {
	defaultStr := "Y/n"
	if !defaultValue {
		defaultStr = "y/N"
	}

	choice := ui.promptChoice(
		fmt.Sprintf("%s [%s]", prompt, defaultStr),
		[]string{"y", "Y", "n", "N", ""},
	)

	if choice == "" {
		return defaultValue
	}

	return strings.ToLower(choice) == "y"
}

func (icm *InteractiveConfigManager) RunQuickSetup() error {
	fmt.Println("⚡ usacloud-update クイックセットアップ")
	fmt.Println("=====================================")

	fmt.Println("使用する設定プロファイルを選択してください:")
	fmt.Println("   1. beginner   - 初心者向け（推奨）")
	fmt.Println("   2. default    - 標準設定")
	fmt.Println("   3. expert     - エキスパート向け")

	choice := icm.ui.promptChoice("選択してください [1-3]", []string{"1", "2", "3"})

	profileMap := map[string]string{
		"1": "beginner",
		"2": "default",
		"3": "expert",
	}

	profileName := profileMap[choice]
	icm.config.General.Profile = profileName

	if err := icm.config.applyProfile(profileName); err != nil {
		return fmt.Errorf("プロファイル適用に失敗: %w", err)
	}

	fmt.Printf("✅ プロファイル '%s' で設定しました\n", profileName)

	if err := icm.config.Save(); err != nil {
		return fmt.Errorf("設定保存に失敗: %w", err)
	}

	fmt.Println("✅ クイックセットアップが完了しました！")
	fmt.Printf("   設定ファイル: %s\n", icm.config.configPath)
	fmt.Println("   詳細な設定は 'usacloud-update config edit' で変更できます")

	return nil
}

func (icm *InteractiveConfigManager) EditSetting(section, key string) error {
	currentValue := icm.config.getSetting(section, key)

	fmt.Printf("現在の値: %v\n", currentValue)
	newValue := icm.ui.promptString("新しい値を入力してください")

	if newValue == "" {
		fmt.Println("変更をキャンセルしました")
		return nil
	}

	return icm.config.UpdateSetting(section, key, newValue)
}
