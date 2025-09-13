package error_scenarios

import (
	"fmt"
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/tests/e2e"
)

// TestErrorRecovery_ValidationErrorsWithFix は検証エラーの修正テスト
func TestErrorRecovery_ValidationErrorsWithFix(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	// 複数の問題を含むスクリプト
	problematicScript := `#!/bin/bash
# 複数の問題を含むスクリプト
usacloud serv list              # typo
usacloud iso-image list         # deprecated  
usacloud server lst --zone=all  # subcommand typo
usacloud invalid-cmd read       # invalid command
usacloud disk create size 100   # missing --
`

	inputFile := suite.CreateTempFile("problematic.sh", problematicScript)

	t.Run("ValidationOnly", func(t *testing.T) {
		// 検証のみモードでエラー検出
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode: 1, // 検証エラーを期待
			ExpectedStderr: []string{
				"serv",
				"iso-image",
				"lst",
				"invalid-cmd",
			},
		}

		result := suite.RunE2ETest("ValidationOnly", options)
		validateDetailedErrorReporting(t, result)
	})

	t.Run("TransformWithPartialFix", func(t *testing.T) {
		// 通常の変換処理での修正可能な部分のみ修正
		outputFile := suite.GetTestDir() + "/fixed.sh"

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", inputFile,
				"--out", outputFile,
			},
			ExpectedExitCode: 0, // 正常終了
			ExpectedStdout: []string{
				"変換完了",
			},
			ExpectedFiles: []e2e.FileExpectation{
				{
					Path:        "fixed.sh",
					ShouldExist: true,
					ContentContains: []string{
						"usacloud cdrom list", // iso-imageが修正される
						"--output-type json",  // csvが修正される
						"--zone=all",          // zone形式が修正される
						"--size 100",          // --が追加される
					},
				},
			},
		}

		result := suite.RunE2ETest("TransformWithPartialFix", options)
		validatePartialAutoFix(t, result)
	})
}

// TestErrorRecovery_ConfigurationErrors は設定エラーの回復テスト
func TestErrorRecovery_ConfigurationErrors(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	t.Run("CorruptedConfigFile", func(t *testing.T) {
		// 破損した設定ファイル
		corruptedConfig := `[general
color_output = true
language = invalid-language
profile = non-existent-profile

[validation]
strict_mode = maybe  # invalid boolean
max_suggestions = -5  # invalid number
`

		configFile := suite.CreateTempFile("corrupted.conf", corruptedConfig)

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--config", configFile,
				"--validate-only",
				"/dev/null", // ダミー入力
			},
			ExpectedExitCode: 1,
			ExpectedStderr: []string{
				"設定",
				"エラー",
			},
		}

		result := suite.RunE2ETest("CorruptedConfigFile", options)
		validateConfigFallback(t, result)
		validateConfigFixSuggestions(t, result)
	})

	t.Run("MissingConfigFile", func(t *testing.T) {
		// 存在しない設定ファイル
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--config", "/non/existent/config.conf",
				"--validate-only",
				"/dev/null",
			},
			ExpectedExitCode: 1,
			ExpectedStderr: []string{
				"設定ファイル",
				"見つかりません",
			},
		}

		result := suite.RunE2ETest("MissingConfigFile", options)
		validateMissingConfigHandling(t, result)
	})
}

// TestErrorRecovery_NetworkIssues はネットワーク問題の回復テスト
func TestErrorRecovery_NetworkIssues(t *testing.T) {
	if testing.Short() {
		t.Skip("ネットワークテストは短時間モードではスキップ")
	}

	suite := e2e.NewE2ETestSuite(t)

	t.Run("UpdateCheckFailure", func(t *testing.T) {
		// アップデートチェックの失敗をシミュレート
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--help",
			},
			Environment: map[string]string{
				"USACLOUD_UPDATE_CHECK_URL": "http://invalid-url-for-testing.example.com",
			},
			ExpectedExitCode: 0, // ヘルプは正常表示されるべき
			ExpectedStdout: []string{
				"usacloud-update",
				"使用方法",
			},
		}

		result := suite.RunE2ETest("UpdateCheckFailure", options)
		validateNetworkFailureResilience(t, result)
	})
}

// TestErrorRecovery_MemoryPressure はメモリ不足時の回復テスト
func TestErrorRecovery_MemoryPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("メモリ負荷テストは短時間モードではスキップ")
	}

	suite := e2e.NewE2ETestSuite(t)

	t.Run("LargeFileHandling", func(t *testing.T) {
		// 大きなファイルでメモリ不足をテスト
		// 実際の実装では、巨大なファイルを作成してテスト
		largeScript := generateLargeScript(10000) // 1万行のスクリプト
		inputFile := suite.CreateTempFile("large.sh", largeScript)

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode:    0, // メモリ不足でもクラッシュしないことを期待
			ValidatePerformance: true,
			MaxExecutionTime:    "30s",
		}

		result := suite.RunE2ETest("LargeFileHandling", options)
		validateMemoryEfficiency(t, result)
	})
}

// TestErrorRecovery_PartialFailures は部分的失敗の回復テスト
func TestErrorRecovery_PartialFailures(t *testing.T) {
	t.Run("BatchProcessingWithErrors", func(t *testing.T) {
		// 通常モードでの複数ファイル一括処理機能は現在未実装
		// このテストは将来の実装のためのプレースホルダー
		t.Skip("通常モードでの複数ファイル一括処理機能は現在未実装です")
	})
}

// 検証ヘルパー関数

func validateDetailedErrorReporting(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 詳細なエラーレポートが含まれているか確認
	expectedSections := []string{
		"検証結果",
		"エラー",
		"警告",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result.Stderr, section) {
			t.Errorf("エラーレポートにセクションが不足: %q", section)
		}
	}

	// エラーの重要度が適切に分類されているか確認
	if !strings.Contains(result.Stderr, "重要度") && !strings.Contains(result.Stderr, "優先度") {
		t.Logf("エラーの重要度分類が明示的でない可能性があります")
	}
}

func validatePartialAutoFix(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 変換完了が報告されているか確認
	expectedReports := []string{
		"変換完了",
	}

	hasReport := false
	for _, report := range expectedReports {
		if strings.Contains(result.Stdout, report) {
			hasReport = true
			break
		}
	}

	if !hasReport {
		t.Errorf("変換完了の報告が不足\n出力:\n%s", result.Stdout)
	}
}

func validateConfigFallback(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 設定エラー時のフォールバック動作を確認
	fallbackIndicators := []string{
		"デフォルト設定",
		"フォールバック",
		"デフォルト値",
	}

	hasFallback := false
	for _, indicator := range fallbackIndicators {
		if strings.Contains(result.Stderr, indicator) {
			hasFallback = true
			break
		}
	}

	if !hasFallback {
		t.Errorf("設定エラー時のフォールバック動作が確認できません")
	}
}

func validateConfigFixSuggestions(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 設定修正の提案が含まれているか確認
	suggestionIndicators := []string{
		"修正方法",
		"確認してください",
		"設定例",
	}

	hasSuggestion := false
	for _, indicator := range suggestionIndicators {
		if strings.Contains(result.Stderr, indicator) {
			hasSuggestion = true
			break
		}
	}

	if !hasSuggestion {
		t.Errorf("設定修正の提案が含まれていません")
	}
}

func validateMissingConfigHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 設定ファイルが見つからない場合の適切な処理を確認
	if result.ExitCode == 0 {
		t.Errorf("存在しない設定ファイルで正常終了しました")
	}

	// 有用なエラーメッセージが表示されているか確認
	if !strings.Contains(result.Stderr, "見つかりません") {
		t.Errorf("設定ファイルが見つからないエラーメッセージが不適切")
	}
}

func validateNetworkFailureResilience(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// ネットワーク失敗時もコア機能が動作することを確認
	if result.ExitCode != 0 {
		t.Errorf("ネットワーク失敗時にコア機能が停止しました: %s", result.Stderr)
	}

	// ネットワークエラーが適切に無視されていることを確認
	if strings.Contains(result.Stderr, "network error") && strings.Contains(result.Stderr, "fatal") {
		t.Errorf("ネットワークエラーが致命的エラーとして扱われています")
	}
}

func validateMemoryEfficiency(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// メモリ不足でクラッシュしていないことを確認
	crashIndicators := []string{
		"out of memory",
		"killed",
		"signal",
		"panic",
	}

	for _, indicator := range crashIndicators {
		if strings.Contains(result.Stderr, indicator) {
			t.Errorf("メモリ不足でクラッシュした可能性: %q が含まれています", indicator)
		}
	}
}

func validatePartialFailureHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 部分失敗が適切に報告されているか確認
	reportElements := []string{
		"正常に処理",
		"エラー",
		"バッチ処理",
	}

	for _, element := range reportElements {
		if !strings.Contains(result.Stdout, element) {
			t.Errorf("部分失敗レポートに要素が不足: %q", element)
		}
	}

	// 終了コードが部分失敗を示していることを確認
	if result.ExitCode == 0 {
		t.Errorf("部分失敗で正常終了しました")
	}
}

// ユーティリティ関数

func generateLargeScript(lines int) string {
	var script strings.Builder
	script.WriteString("#!/bin/bash\n")
	script.WriteString("# Large script for memory testing\n\n")

	commands := []string{
		"usacloud server list",
		"usacloud disk list --output-type csv",
		"usacloud database list",
		"echo 'Processing step %d'",
	}

	for i := 0; i < lines; i++ {
		cmd := commands[i%len(commands)]
		if strings.Contains(cmd, "%d") {
			script.WriteString(fmt.Sprintf(cmd+"\n", i+1))
		} else {
			script.WriteString(cmd + "\n")
		}
	}

	return script.String()
}
