package user_workflows

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/tests/e2e"
)

// TestBeginnerWorkflow_FirstTimeUser は初回利用者のワークフローテスト
func TestBeginnerWorkflow_FirstTimeUser(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	// シナリオ: 初心者がtypoを含むスクリプトを変換
	t.Run("TypoInScript", func(t *testing.T) {
		// 1. typoを含むスクリプト作成
		inputScript := `#!/bin/bash
usacloud serv list
usacloud iso-image lst
usacloud dsk create --size 100`

		inputFile := suite.CreateTempFile("input.sh", inputScript)

		// 2. 初心者プロファイルで実行（非インタラクティブ）
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode: 1, // 検証エラーを期待
			ExpectedStderr: []string{
				"serv",
				"iso-image",
				"dsk",
				"lst",
			},
		}

		result := suite.RunE2ETest("BeginnerTypoDetection", options)

		// 追加検証: エラーメッセージの確認
		validateBeginnerFriendlyErrors(t, result)
	})

	t.Run("SimpleTransformation", func(t *testing.T) {
		// 簡単な変換タスク
		inputScript := `#!/bin/bash
usacloud server list --output-type csv
usacloud disk list --output-type tsv`

		inputFile := suite.CreateTempFile("simple.sh", inputScript)
		outputFile := filepath.Join(suite.GetTestDir(), "simple_output.sh")

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", inputFile,
				"--out", outputFile,
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"変換完了",
			},
			ExpectedFiles: []e2e.FileExpectation{
				{
					Path:        "simple_output.sh",
					ShouldExist: true,
					ContentContains: []string{
						"usacloud server list --output-type json",
						"usacloud disk list --output-type json",
						"# usacloud-update:",
					},
				},
			},
		}

		result := suite.RunE2ETest("SimpleTransformation", options)
		validateTransformationOutput(t, result)
	})
}

// TestBeginnerWorkflow_LearningProgress は学習進捗のワークフローテスト
func TestBeginnerWorkflow_LearningProgress(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	// シナリオ: 初心者が段階的にスキルアップ
	learningSteps := []struct {
		name              string
		script            string
		expectImprovement bool
	}{
		{
			name:              "Step1_BasicMistakes",
			script:            "usacloud serv list",
			expectImprovement: false,
		},
		{
			name:              "Step2_SameMistake",
			script:            "usacloud serv read 123",
			expectImprovement: false,
		},
		{
			name:              "Step3_NewCommand",
			script:            "usacloud server create --name test",
			expectImprovement: true,
		},
	}

	for i, step := range learningSteps {
		t.Run(step.name, func(t *testing.T) {
			inputFile := suite.CreateTempFile(fmt.Sprintf("step%d.sh", i+1), step.script)

			options := &e2e.E2ETestOptions{
				Arguments: []string{
					"--validate-only",
					inputFile,
				},
				ExpectedExitCode: 1, // 検証エラーを期待（step3以外）
			}

			if step.expectImprovement {
				options.ExpectedExitCode = 0 // step3は正常終了を期待
			}

			result := suite.RunE2ETest(step.name, options)

			if step.expectImprovement {
				// より詳細なアドバイスが減っているかチェック
				validateReducedHelp(t, result)
			}
		})
	}
}

// TestBeginnerWorkflow_HelpSystem はヘルプシステムのワークフローテスト
func TestBeginnerWorkflow_HelpSystem(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	t.Run("BasicHelp", func(t *testing.T) {
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--help",
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"usacloud-update",
				"Usage",
				"Flags",
				"使用例:",
			},
		}

		result := suite.RunE2ETest("BasicHelp", options)
		validateHelpOutput(t, result)
	})

	t.Run("ErrorWithHelpHint", func(t *testing.T) {
		// 無効な引数でヘルプヒントが表示されるかテスト
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--invalid-option",
			},
			ExpectedExitCode: 1, // Cobraは無効なフラグでexit code 1を返す
			ExpectedStderr: []string{
				"Error: unknown flag",
				"Usage",
			},
		}

		result := suite.RunE2ETest("ErrorWithHelpHint", options)
		validateErrorHelpHints(t, result)
	})
}

// TestBeginnerWorkflow_FileHandling はファイル処理のワークフローテスト
func TestBeginnerWorkflow_FileHandling(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	t.Run("StandardInputOutput", func(t *testing.T) {
		inputScript := "usacloud server list --output-type csv"

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", "-", // 標準入力
				"--out", "-", // 標準出力
			},
			StdinInput:       inputScript,
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"usacloud server list --output-type json",
				"# usacloud-update:",
			},
		}

		result := suite.RunE2ETest("StandardInputOutput", options)
		validateStandardIOHandling(t, result)
	})

	t.Run("FileToFileProcessing", func(t *testing.T) {
		// ファイルからファイルへの変換処理
		script := "usacloud server list --output-type csv\nusacloud disk list --output-type tsv"

		inputFile := suite.CreateTempFile("input_script.sh", script)
		outputFile := filepath.Join(suite.GetTestDir(), "output_script.sh")

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", inputFile,
				"--out", outputFile,
			},
			ExpectedExitCode: 0,
			ExpectedFiles: []e2e.FileExpectation{
				{
					Path:        "output_script.sh",
					ShouldExist: true,
					ContentContains: []string{
						"--output-type json",
						"# usacloud-update:",
					},
				},
			},
		}

		result := suite.RunE2ETest("FileToFileProcessing", options)
		validateFileProcessing(t, result)
	})
}

// 検証ヘルパー関数

func validateBeginnerFriendlyErrors(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 初心者向けの丁寧なエラーメッセージが含まれているか確認
	expectedMessages := []string{
		"もしかして以下のコマンドですか？",
		"詳細情報",
	}

	for _, msg := range expectedMessages {
		if !containsIgnoreCase(result.Stderr, msg) {
			t.Errorf("初心者向けエラーメッセージが不足: %q が含まれていません\nエラー出力:\n%s", msg, result.Stderr)
		}
	}
}

func validateTransformationOutput(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 変換処理が正常に完了したことを確認
	if result.ExitCode != 0 {
		t.Errorf("変換処理が失敗しました: 終了コード=%d, エラー=%s", result.ExitCode, result.Stderr)
	}

	// 統計情報が含まれているか確認
	expectedStats := []string{
		"変換",
		"完了",
	}

	for _, stat := range expectedStats {
		if !containsIgnoreCase(result.Stdout, stat) {
			t.Errorf("変換統計情報が不足: %q が含まれていません", stat)
		}
	}
}

func validateReducedHelp(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 学習が進んでいる場合は、より簡潔なメッセージになることを期待
	// ここでは正常なコマンドとして扱われることを確認
	if result.ExitCode != 0 {
		t.Errorf("学習が進んだコマンドで予期しないエラー: %s", result.Stderr)
	}
}

func validateHelpOutput(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// ヘルプ出力の品質を確認（Cobra標準出力に合わせて修正）
	requiredSections := []string{
		"usacloud-update は異なるバージョン", // 概要部分
		"Usage", // 使用方法 → Usage
		"Flags", // オプション → Flags
		"使用例",   // 例 → 使用例（実際の出力に合わせる）
	}

	for _, section := range requiredSections {
		if !containsIgnoreCase(result.Stdout, section) {
			t.Errorf("ヘルプ出力にセクションが不足: %q", section)
		}
	}
}

func validateErrorHelpHints(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// エラー時のヘルプヒントが適切に表示されているか確認（Cobra標準に合わせて修正）
	expectedHints := []string{
		"Usage",
		"Use \"usacloud-update",
	}

	for _, hint := range expectedHints {
		if !containsIgnoreCase(result.Stderr, hint) {
			t.Errorf("エラー時のヘルプヒントが不足: %q", hint)
		}
	}
}

func validateStandardIOHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 標準入出力が正しく処理されているか確認
	if result.ExitCode != 0 {
		t.Errorf("標準入出力処理が失敗: %s", result.Stderr)
	}

	// 変換結果が標準出力に表示されているか確認
	if !containsIgnoreCase(result.Stdout, "--output-type json") {
		t.Errorf("変換結果が標準出力に表示されていません")
	}
}

func validateFileProcessing(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// ファイル処理が正常に完了したことを確認
	if result.ExitCode != 0 {
		t.Errorf("ファイル処理が失敗: %s", result.Stderr)
	}

	// 変換処理が実行されたことを確認
	// 統計出力は stderr に出力されるため、そちらをチェック
	if result.Stderr != "" && !containsIgnoreCase(result.Stderr, "#L") {
		t.Logf("統計情報が期待されますが見つかりませんでした: %s", result.Stderr)
	}
}

func validateBatchProcessing(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// バッチ処理が正常に完了したことを確認
	if result.ExitCode != 0 {
		t.Errorf("バッチ処理が失敗: %s", result.Stderr)
	}

	// 処理統計が表示されているか確認
	expectedStats := []string{
		"バッチ処理",
		"ファイル",
		"処理完了",
	}

	for _, stat := range expectedStats {
		if !containsIgnoreCase(result.Stdout, stat) {
			t.Errorf("バッチ処理統計が不足: %q", stat)
		}
	}
}

// ユーティリティ関数

func containsIgnoreCase(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}

// GetTestDirを追加するためのE2ETestSuiteの拡張
// この関数は実際のe2e_test_framework.goに追加する必要があります
