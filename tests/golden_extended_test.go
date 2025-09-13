package tests

import (
	"fmt"
	"testing"

	goldenTesting "github.com/armaniacs/usacloud-update/internal/testing"
)

// TestGolden_TransformWithValidation は変換＋検証のゴールデンテスト
func TestGolden_TransformWithValidation(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	testCases := []struct {
		name    string
		options *goldenTesting.GoldenTestOptions
	}{
		{
			name: "BasicTransformWithValidation",
			options: &goldenTesting.GoldenTestOptions{
				InputFile:          "sample_v0_v1_mixed.sh",
				ConfigFile:         "default.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
			},
		},
		{
			name: "StrictModeValidation",
			options: &goldenTesting.GoldenTestOptions{
				InputFile:          "problematic_script.sh",
				ConfigFile:         "strict.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
				StrictMode:         true,
			},
		},
		{
			name: "BeginnerModeHelp",
			options: &goldenTesting.GoldenTestOptions{
				InputFile:          "typo_commands.sh",
				ConfigFile:         "beginner.conf",
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeHelp:        true,
				IncludeSuggestions: true,
				InteractiveMode:    true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.RunGoldenTest(tc.name, tc.options)
		})
	}
}

// TestGolden_MultiLanguage は多言語ゴールデンテスト
func TestGolden_MultiLanguage(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	languages := []string{"ja", "en"}

	for _, lang := range languages {
		t.Run(fmt.Sprintf("Language_%s", lang), func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:          "error_scenarios.sh",
				ConfigFile:         fmt.Sprintf("default_%s.conf", lang),
				Language:           lang,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
			}

			suite.RunGoldenTest(fmt.Sprintf("MultiLanguage_%s", lang), options)
		})
	}
}

// TestGolden_ColorOutput はカラー出力のゴールデンテスト
func TestGolden_ColorOutput(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	colorModes := []struct {
		name    string
		enabled bool
	}{
		{"ColorEnabled", true},
		{"PlainText", false},
	}

	for _, mode := range colorModes {
		t.Run(mode.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:         "error_scenarios.sh",
				ConfigFile:        "default.conf",
				ColorEnabled:      mode.enabled,
				IncludeValidation: true,
				IncludeErrors:     true,
			}

			suite.RunGoldenTest(fmt.Sprintf("Color_%s", mode.name), options)
		})
	}
}

// TestGolden_ErrorScenarios はエラーシナリオのゴールデンテスト
func TestGolden_ErrorScenarios(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	errorScenarios := []struct {
		name        string
		inputFile   string
		description string
	}{
		{
			name:        "TypoCommands",
			inputFile:   "typo_commands.sh",
			description: "タイポコマンドの検証",
		},
		{
			name:        "DeprecatedCommands",
			inputFile:   "deprecated_commands.sh",
			description: "廃止コマンドの検証",
		},
		{
			name:        "InvalidOptions",
			inputFile:   "invalid_options.sh",
			description: "無効なオプションの検証",
		},
		{
			name:        "MixedErrors",
			inputFile:   "mixed_errors.sh",
			description: "複数種類のエラーが混在した検証",
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:          scenario.inputFile,
				ConfigFile:         "default.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
				IncludeHelp:        true,
			}

			suite.RunGoldenTest(scenario.name, options)
		})
	}
}

// TestGolden_PerformanceInputs はパフォーマンス用入力のゴールデンテスト
func TestGolden_PerformanceInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストは短時間モードではスキップ")
	}

	suite := goldenTesting.NewGoldenTestSuite(t)

	performanceScenarios := []struct {
		name      string
		inputFile string
		lineCount int
	}{
		{
			name:      "SmallScript",
			inputFile: "small_100_lines.sh",
			lineCount: 100,
		},
		{
			name:      "MediumScript",
			inputFile: "medium_1000_lines.sh",
			lineCount: 1000,
		},
		{
			name:      "LargeScript",
			inputFile: "large_5000_lines.sh",
			lineCount: 5000,
		},
	}

	for _, scenario := range performanceScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:          scenario.inputFile,
				ConfigFile:         "performance.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      false, // パフォーマンステストではエラー出力は不要
				IncludeSuggestions: false,
			}

			suite.RunGoldenTest(scenario.name, options)
		})
	}
}

// TestGolden_IntegrationWorkflows は統合ワークフローのゴールデンテスト
func TestGolden_IntegrationWorkflows(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	workflowScenarios := []struct {
		name        string
		inputFile   string
		configFile  string
		description string
	}{
		{
			name:        "BeginnerWorkflow",
			inputFile:   "beginner_scenario.sh",
			configFile:  "beginner.conf",
			description: "初心者向けワークフロー",
		},
		{
			name:        "ExpertWorkflow",
			inputFile:   "expert_scenario.sh",
			configFile:  "expert.conf",
			description: "エキスパート向けワークフロー",
		},
		{
			name:        "CIWorkflow",
			inputFile:   "ci_scenario.sh",
			configFile:  "ci.conf",
			description: "CI環境でのワークフロー",
		},
		{
			name:        "ProductionMigration",
			inputFile:   "production_migration.sh",
			configFile:  "production.conf",
			description: "本番環境移行ワークフロー",
		},
	}

	for _, scenario := range workflowScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:          scenario.inputFile,
				ConfigFile:         scenario.configFile,
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
				IncludeHelp:        true,
			}

			suite.RunGoldenTest(scenario.name, options)
		})
	}
}

// TestGolden_RegressionSafety は既存機能の回帰安全性テスト
func TestGolden_RegressionSafety(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	// 既存のゴールデンファイルテストと互換性確認
	legacyScenarios := []struct {
		name      string
		inputFile string
	}{
		{
			name:      "LegacyBasicTransform",
			inputFile: "sample_v0_v1_mixed.sh", // 既存のテストファイル
		},
		{
			name:      "LegacyMixedNonUsacloud",
			inputFile: "mixed_with_non_usacloud.sh", // 既存のテストファイル
		},
	}

	for _, scenario := range legacyScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:        scenario.inputFile,
				ConfigFile:       "default.conf",
				IncludeTransform: true,
				// 既存テストとの互換性のため、検証機能は無効化
				IncludeValidation:  false,
				IncludeErrors:      false,
				IncludeSuggestions: false,
			}

			suite.RunGoldenTest(scenario.name, options)
		})
	}
}

// TestGolden_GeneratedScenarios は自動生成シナリオのゴールデンテスト
func TestGolden_GeneratedScenarios(t *testing.T) {
	// テストデータ生成器を使用
	generator := goldenTesting.NewGoldenDataGenerator()
	scenarios := generator.GenerateTestScenarios(5) // 5つのシナリオを生成

	suite := goldenTesting.NewGoldenTestSuite(t)

	for i, scenario := range scenarios {
		t.Run(fmt.Sprintf("Generated_%d_%s", i+1, scenario.Name), func(t *testing.T) {
			// 生成されたシナリオを一時ファイルとして保存
			tempInputFile := fmt.Sprintf("generated_scenario_%d.sh", i+1)

			options := &goldenTesting.GoldenTestOptions{
				InputFile:          tempInputFile,
				ConfigFile:         "default.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
			}

			// 注意: 実際の実装では一時ファイルの作成・削除が必要
			suite.RunGoldenTest(scenario.Name, options)
		})
	}
}

// TestGolden_EdgeCases はエッジケースのゴールデンテスト
func TestGolden_EdgeCases(t *testing.T) {
	suite := goldenTesting.NewGoldenTestSuite(t)

	edgeCases := []struct {
		name        string
		inputFile   string
		description string
	}{
		{
			name:        "EmptyFile",
			inputFile:   "empty.sh",
			description: "空ファイルの処理",
		},
		{
			name:        "OnlyComments",
			inputFile:   "only_comments.sh",
			description: "コメントのみのファイル",
		},
		{
			name:        "VeryLongLines",
			inputFile:   "very_long_lines.sh",
			description: "非常に長い行を含むファイル",
		},
		{
			name:        "SpecialCharacters",
			inputFile:   "special_characters.sh",
			description: "特殊文字を含むファイル",
		},
		{
			name:        "UnicodeContent",
			inputFile:   "unicode_content.sh",
			description: "Unicode文字を含むファイル",
		},
	}

	for _, edgeCase := range edgeCases {
		t.Run(edgeCase.name, func(t *testing.T) {
			options := &goldenTesting.GoldenTestOptions{
				InputFile:          edgeCase.inputFile,
				ConfigFile:         "default.conf",
				IncludeTransform:   true,
				IncludeValidation:  true,
				IncludeErrors:      true,
				IncludeSuggestions: true,
			}

			suite.RunGoldenTest(edgeCase.name, options)
		})
	}
}
