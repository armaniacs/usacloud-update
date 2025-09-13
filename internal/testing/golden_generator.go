package testing

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GoldenDataGenerator はゴールデンテストデータ生成器
type GoldenDataGenerator struct {
	commandPatterns []CommandPattern
	errorPatterns   []ErrorPattern
	typoPatterns    map[string][]string
}

// secureRandInt はcrypto/randを使用して安全な乱数を生成
func secureRandInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// フォールバック: 時刻ベースの疑似乱数
		return int(time.Now().UnixNano()) % max
	}
	return int(n.Int64())
}

// secureRandFloat はcrypto/randを使用して0-1の乱数を生成
func secureRandFloat() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// フォールバック
		return float32(time.Now().UnixNano()%1000000) / 1000000.0
	}
	return float32(n.Int64()) / 1000000.0
}

// CommandPattern はコマンドパターン
type CommandPattern struct {
	Template    string   // "usacloud {command} {subcommand} {options}"
	Commands    []string // ["server", "disk", "database"]
	Subcommands []string // ["list", "read", "create"]
	Options     []string // ["--output-type json", "--zone is1a"]
}

// ErrorPattern はエラーパターン
type ErrorPattern struct {
	Type        string // "typo", "deprecated", "invalid"
	Pattern     string // エラーパターン
	Severity    string // "error", "warning"
	Description string // パターンの説明
}

// TestScenario はテストシナリオ
type TestScenario struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Input       ScenarioInput `json:"input"`
	Expected    interface{}   `json:"expected"`
}

// ScenarioInput はシナリオ入力
type ScenarioInput struct {
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// NewGoldenDataGenerator は新しいゴールデンデータ生成器を作成
func NewGoldenDataGenerator() *GoldenDataGenerator {
	return &GoldenDataGenerator{
		commandPatterns: []CommandPattern{
			{
				Template:    "usacloud {command} {subcommand} {options}",
				Commands:    []string{"server", "disk", "database", "note", "cdrom"},
				Subcommands: []string{"list", "read", "create", "update", "delete"},
				Options:     []string{"--output-type json", "--zone is1a", "--zone tk1v", "--tags web"},
			},
			{
				Template:    "usacloud {command} {options}",
				Commands:    []string{"profile", "config", "version"},
				Subcommands: []string{},
				Options:     []string{"--help", "--version", "--config-path ~/.usacloud"},
			},
		},
		errorPatterns: []ErrorPattern{
			{
				Type:        "typo",
				Pattern:     "usacloud serv list",
				Severity:    "error",
				Description: "serverコマンドのタイポ",
			},
			{
				Type:        "deprecated",
				Pattern:     "usacloud iso-image list",
				Severity:    "error",
				Description: "廃止されたiso-imageコマンド",
			},
			{
				Type:        "deprecated_option",
				Pattern:     "usacloud server list --output-type csv",
				Severity:    "warning",
				Description: "非推奨のcsv出力形式",
			},
			{
				Type:        "invalid_option",
				Pattern:     "usacloud server list --invalid-option",
				Severity:    "error",
				Description: "存在しないオプション",
			},
		},
		typoPatterns: map[string][]string{
			"server":   {"serv", "sever", "servr"},
			"database": {"databse", "datbase", "db"},
			"cdrom":    {"cd-rom", "cdrom-image", "cd"},
			"list":     {"lst", "listt", "lsit"},
			"create":   {"creat", "crete", "crate"},
		},
	}
}

// GenerateTestScenarios はテストシナリオを生成
func (gdg *GoldenDataGenerator) GenerateTestScenarios(count int) []TestScenario {
	scenarios := make([]TestScenario, count)

	for i := 0; i < count; i++ {
		scenario := TestScenario{
			Name:        fmt.Sprintf("GeneratedScenario_%03d", i+1),
			Description: "自動生成されたテストシナリオ",
			Input:       gdg.generateRandomInput(),
			Expected:    gdg.generateExpectedOutput(),
		}
		scenarios[i] = scenario
	}

	return scenarios
}

// generateRandomInput はランダムな入力を生成
func (gdg *GoldenDataGenerator) generateRandomInput() ScenarioInput {
	inputTypes := []string{"valid_command", "typo_command", "deprecated_command", "invalid_command"}
	inputType := inputTypes[secureRandInt(len(inputTypes))]

	switch inputType {
	case "valid_command":
		return gdg.generateValidCommand()
	case "typo_command":
		return gdg.generateTypoCommand()
	case "deprecated_command":
		return gdg.generateDeprecatedCommand()
	case "invalid_command":
		return gdg.generateInvalidCommand()
	}

	return ScenarioInput{}
}

// generateValidCommand は有効なコマンドを生成
func (gdg *GoldenDataGenerator) generateValidCommand() ScenarioInput {
	pattern := gdg.commandPatterns[secureRandInt(len(gdg.commandPatterns))]

	command := pattern.Commands[secureRandInt(len(pattern.Commands))]

	var subcommand string
	if len(pattern.Subcommands) > 0 {
		subcommand = pattern.Subcommands[secureRandInt(len(pattern.Subcommands))]
	}

	var option string
	if len(pattern.Options) > 0 {
		option = pattern.Options[secureRandInt(len(pattern.Options))]
	}

	// テンプレートに値を代入
	content := strings.Replace(pattern.Template, "{command}", command, -1)
	content = strings.Replace(content, "{subcommand}", subcommand, -1)
	content = strings.Replace(content, "{options}", option, -1)

	// 空の部分を削除
	content = strings.Replace(content, "  ", " ", -1)
	content = strings.TrimSpace(content)

	return ScenarioInput{
		Type:    "command",
		Content: content,
	}
}

// generateTypoCommand はタイポコマンドを生成
func (gdg *GoldenDataGenerator) generateTypoCommand() ScenarioInput {
	// ランダムなコマンドを選択
	commands := []string{"server", "database", "cdrom", "list", "create"}
	originalCommand := commands[secureRandInt(len(commands))]

	// タイポパターンを取得
	typos, exists := gdg.typoPatterns[originalCommand]
	if !exists || len(typos) == 0 {
		// タイポパターンがない場合は文字を入れ替え
		typos = []string{gdg.generateCharacterSwap(originalCommand)}
	}

	typo := typos[secureRandInt(len(typos))]

	// コマンド生成
	var content string
	if originalCommand == "list" || originalCommand == "create" {
		// サブコマンドの場合
		commands := []string{"server", "disk", "database"}
		mainCommand := commands[secureRandInt(len(commands))]
		content = fmt.Sprintf("usacloud %s %s", mainCommand, typo)
	} else {
		// メインコマンドの場合
		subcommands := []string{"list", "read", "create"}
		subcommand := subcommands[secureRandInt(len(subcommands))]
		content = fmt.Sprintf("usacloud %s %s", typo, subcommand)
	}

	return ScenarioInput{
		Type:    "command",
		Content: content,
	}
}

// generateDeprecatedCommand は廃止コマンドを生成
func (gdg *GoldenDataGenerator) generateDeprecatedCommand() ScenarioInput {
	deprecatedCommands := []string{
		"usacloud iso-image list",
		"usacloud startup-script list",
		"usacloud ipv4 list",
		"usacloud product-disk list",
		"usacloud summary",
		"usacloud object-storage list",
		"usacloud server list --output-type csv",
		"usacloud disk list --output-type tsv",
	}

	content := deprecatedCommands[secureRandInt(len(deprecatedCommands))]

	return ScenarioInput{
		Type:    "command",
		Content: content,
	}
}

// generateInvalidCommand は無効なコマンドを生成
func (gdg *GoldenDataGenerator) generateInvalidCommand() ScenarioInput {
	invalidPatterns := []string{
		"usacloud invalid-command list",
		"usacloud server invalid-subcommand",
		"usacloud disk list --invalid-option",
		"usacloud xyz abc",
		"usacloud server list --zone = invalid",
	}

	content := invalidPatterns[secureRandInt(len(invalidPatterns))]

	return ScenarioInput{
		Type:    "command",
		Content: content,
	}
}

// generateCharacterSwap は文字入れ替えでタイポを生成
func (gdg *GoldenDataGenerator) generateCharacterSwap(word string) string {
	if len(word) < 2 {
		return word
	}

	runes := []rune(word)
	i := secureRandInt(len(runes) - 1)

	// 隣接する文字を入れ替え
	runes[i], runes[i+1] = runes[i+1], runes[i]

	return string(runes)
}

// generateExpectedOutput は期待される出力を生成
func (gdg *GoldenDataGenerator) generateExpectedOutput() interface{} {
	// 簡易実装：実際の検証結果に基づいた期待値を生成
	return map[string]interface{}{
		"validation_status": "pending",
		"generated":         true,
		"timestamp":         time.Now().Format(time.RFC3339),
	}
}

// GenerateComplexScript は複雑なスクリプトを生成
func (gdg *GoldenDataGenerator) GenerateComplexScript(lineCount int) string {
	var lines []string

	// シェバン行
	lines = append(lines, "#!/bin/bash")
	lines = append(lines, "# 自動生成されたテストスクリプト")
	lines = append(lines, "")

	// ランダムなコマンドを生成
	for i := 0; i < lineCount; i++ {
		var line string

		switch secureRandInt(5) {
		case 0:
			// 有効なusacloudコマンド
			input := gdg.generateValidCommand()
			line = input.Content
		case 1:
			// タイポコマンド
			input := gdg.generateTypoCommand()
			line = input.Content
		case 2:
			// 廃止コマンド
			input := gdg.generateDeprecatedCommand()
			line = input.Content
		case 3:
			// 非usacloudコマンド
			nonUsacloudCommands := []string{
				"echo 'Hello World'",
				"curl -X GET https://api.example.com",
				"python3 script.py",
				"docker run nginx",
				"ssh user@server",
			}
			line = nonUsacloudCommands[secureRandInt(len(nonUsacloudCommands))]
		case 4:
			// コメント行
			comments := []string{
				"# サーバー操作",
				"# データベース設定",
				"# バックアップ処理",
				"# 監視設定",
			}
			line = comments[secureRandInt(len(comments))]
		}

		lines = append(lines, line)

		// ランダムで空行追加
		if secureRandFloat() < 0.2 {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// GenerateMultiLanguageTestData は多言語テストデータを生成
func (gdg *GoldenDataGenerator) GenerateMultiLanguageTestData() map[string][]TestScenario {
	languages := map[string]string{
		"ja": "日本語",
		"en": "English",
	}

	result := make(map[string][]TestScenario)

	for lang, langName := range languages {
		scenarios := []TestScenario{}

		// 基本的なエラーシナリオ
		scenarios = append(scenarios, TestScenario{
			Name:        fmt.Sprintf("BasicError_%s", lang),
			Description: fmt.Sprintf("%sでの基本エラーテスト", langName),
			Input: ScenarioInput{
				Type:    "command",
				Content: "usacloud serv list",
			},
			Expected: map[string]interface{}{
				"language":   lang,
				"error_type": "command_not_found",
			},
		})

		// 廃止コマンドシナリオ
		scenarios = append(scenarios, TestScenario{
			Name:        fmt.Sprintf("DeprecatedCommand_%s", lang),
			Description: fmt.Sprintf("%sでの廃止コマンドテスト", langName),
			Input: ScenarioInput{
				Type:    "command",
				Content: "usacloud iso-image list",
			},
			Expected: map[string]interface{}{
				"language":   lang,
				"error_type": "deprecated_command",
			},
		})

		result[lang] = scenarios
	}

	return result
}

// GeneratePerformanceTestData はパフォーマンステストデータを生成
func (gdg *GoldenDataGenerator) GeneratePerformanceTestData(sizes []int) map[string]string {
	result := make(map[string]string)

	for _, size := range sizes {
		sizeName := ""
		switch {
		case size <= 100:
			sizeName = "small"
		case size <= 1000:
			sizeName = "medium"
		case size <= 10000:
			sizeName = "large"
		default:
			sizeName = "huge"
		}

		script := gdg.GenerateComplexScript(size)
		result[fmt.Sprintf("%s_%d_lines", sizeName, size)] = script
	}

	return result
}
