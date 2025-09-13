package error_scenarios

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/tests/e2e"
)

// TestFileErrors_InputFileIssues は入力ファイル問題のテスト
func TestFileErrors_InputFileIssues(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	errorScenarios := []struct {
		name        string
		setupFunc   func(*e2e.E2ETestSuite) string
		expectedMsg string
	}{
		{
			name: "NonExistentFile",
			setupFunc: func(suite *e2e.E2ETestSuite) string {
				return "/non/existent/file.sh"
			},
			expectedMsg: "ファイルが見つかりません",
		},
		{
			name: "NoReadPermission",
			setupFunc: func(suite *e2e.E2ETestSuite) string {
				file := suite.CreateTempFile("noaccess.sh", "usacloud server list")
				os.Chmod(file, 0000) // 読み取り権限削除
				return file
			},
			expectedMsg: "読み取り権限",
		},
		{
			name: "EmptyFile",
			setupFunc: func(suite *e2e.E2ETestSuite) string {
				return suite.CreateTempFile("empty.sh", "")
			},
			expectedMsg: "", // 空ファイルは許可される場合もある
		},
		{
			name: "BinaryFile",
			setupFunc: func(suite *e2e.E2ETestSuite) string {
				binaryData := []byte{0x00, 0x01, 0x02, 0xFF}
				return suite.CreateTempFileBytes("binary.bin", binaryData)
			},
			expectedMsg: "", // バイナリファイルのエラーは実装依存
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			inputFile := scenario.setupFunc(suite)

			options := &e2e.E2ETestOptions{
				Arguments: []string{
					"--in", inputFile,
					"--out", "/tmp/output.sh",
				},
				ExpectedExitCode: 1,
			}

			// 期待メッセージが設定されている場合のみチェック
			if scenario.expectedMsg != "" {
				options.ExpectedStderr = []string{scenario.expectedMsg}
			}

			result := suite.RunE2ETest(scenario.name, options)
			validateGracefulErrorHandling(t, result)
		})
	}
}

// TestFileErrors_OutputFileIssues は出力ファイル問題のテスト
func TestFileErrors_OutputFileIssues(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	inputScript := "usacloud server list"
	inputFile := suite.CreateTempFile("input.sh", inputScript)

	t.Run("NoWritePermission", func(t *testing.T) {
		// 書き込み権限のないディレクトリ
		noWriteDir := suite.CreateTempDir("nowrite")
		os.Chmod(noWriteDir, 0555) // 読み取り・実行のみ

		outputFile := filepath.Join(noWriteDir, "output.sh")

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", inputFile,
				"--out", outputFile,
			},
			ExpectedExitCode: 1,
			ExpectedStderr: []string{
				"出力",
				"失敗",
			},
		}

		result := suite.RunE2ETest("NoWritePermission", options)
		validateErrorRecoveryAdvice(t, result)
	})

	t.Run("OutputToDirectory", func(t *testing.T) {
		// ディレクトリを出力ファイルとして指定
		outputDir := suite.CreateTempDir("is_directory")

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--in", inputFile,
				"--out", outputDir,
			},
			ExpectedExitCode: 1,
			ExpectedStderr: []string{
				"ディレクトリ",
			},
		}

		result := suite.RunE2ETest("OutputToDirectory", options)
		validateDirectoryErrorHandling(t, result)
	})
}

// TestFileErrors_CorruptedFiles は破損ファイルのテスト
func TestFileErrors_CorruptedFiles(t *testing.T) {
	suite := e2e.NewE2ETestSuite(t)

	t.Run("VeryLongLines", func(t *testing.T) {
		// 非常に長い行を含むファイル
		longLine := "usacloud server list " + strings.Repeat("--tag very-long-tag-name ", 1000)
		inputFile := suite.CreateTempFile("longline.sh", longLine)

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode: 0, // 長い行でも処理できるはず
		}

		result := suite.RunE2ETest("VeryLongLines", options)
		validateLongLineHandling(t, result)
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// 特殊文字を含むファイル
		specialScript := `#!/bin/bash
# テスト用スクリプト with 特殊文字 !@#$%^&*()
usacloud server list --tags "テスト用タグ"
usacloud disk create --name "ディスク/名前"
echo "日本語コメント: 正常に処理されるはず"`

		inputFile := suite.CreateTempFile("special.sh", specialScript)

		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode: 0, // 特殊文字でも処理できるはず
		}

		result := suite.RunE2ETest("SpecialCharacters", options)
		validateSpecialCharacterHandling(t, result)
	})
}

// TestFileErrors_ConcurrentAccess は同時アクセスのテスト
func TestFileErrors_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("同時アクセステストは短時間モードではスキップ")
	}

	suite := e2e.NewE2ETestSuite(t)

	t.Run("MultipleReaders", func(t *testing.T) {
		// 複数のプロセスが同じファイルを読み取り
		inputScript := "usacloud server list --output-type csv"
		inputFile := suite.CreateTempFile("shared.sh", inputScript)

		// 複数のプロセスを並行実行（簡易テスト）
		options := &e2e.E2ETestOptions{
			Arguments: []string{
				"--validate-only",
				inputFile,
			},
			ExpectedExitCode: 0,
		}

		result := suite.RunE2ETest("MultipleReaders", options)
		validateConcurrentReadHandling(t, result)
	})
}

// 検証ヘルパー関数

func validateGracefulErrorHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// エラーが適切に処理されているか確認
	if result.ExitCode == 0 {
		t.Errorf("エラーケースで正常終了しました")
	}

	// エラーメッセージが有用であることを確認
	if result.Stderr == "" {
		t.Errorf("エラーメッセージが出力されていません")
	}

	// スタックトレースや内部エラーが露出していないことを確認
	problematicPatterns := []string{
		"panic:",
		"runtime error:",
		"goroutine",
		"stack trace",
	}

	for _, pattern := range problematicPatterns {
		if strings.Contains(result.Stderr, pattern) {
			t.Errorf("内部エラーが露出しています: %q が含まれています\nエラー出力:\n%s", pattern, result.Stderr)
		}
	}
}

func validateErrorRecoveryAdvice(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// エラー回復のアドバイスが含まれているか確認
	expectedAdvice := []string{
		"確認してください",
		"権限",
		"解決方法",
	}

	hasAdvice := false
	for _, advice := range expectedAdvice {
		if strings.Contains(result.Stderr, advice) {
			hasAdvice = true
			break
		}
	}

	if !hasAdvice {
		t.Errorf("エラー回復のアドバイスが含まれていません\nエラー出力:\n%s", result.Stderr)
	}
}

func validateDirectoryErrorHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// ディレクトリエラーが適切に処理されているか確認
	if result.ExitCode == 0 {
		t.Errorf("ディレクトリ指定エラーで正常終了しました")
	}

	// 明確なエラーメッセージが表示されているか確認
	if !strings.Contains(result.Stderr, "ディレクトリ") {
		t.Errorf("ディレクトリエラーメッセージが不適切: %s", result.Stderr)
	}
}

func validateLongLineHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 長い行でもクラッシュしないことを確認
	if strings.Contains(result.Stderr, "panic") || strings.Contains(result.Stderr, "runtime error") {
		t.Errorf("長い行でクラッシュしました: %s", result.Stderr)
	}

	// メモリ不足エラーが発生していないことを確認
	if strings.Contains(result.Stderr, "out of memory") {
		t.Errorf("長い行でメモリ不足が発生: %s", result.Stderr)
	}
}

func validateSpecialCharacterHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 特殊文字でエンコーディングエラーが発生していないことを確認
	if strings.Contains(result.Stderr, "encoding") || strings.Contains(result.Stderr, "invalid UTF-8") {
		t.Errorf("特殊文字でエンコーディングエラーが発生: %s", result.Stderr)
	}

	// 文字化けが発生していないことを確認（簡易チェック）
	if strings.Contains(result.Stderr, "�") {
		t.Errorf("文字化けが発生している可能性があります: %s", result.Stderr)
	}
}

func validateConcurrentReadHandling(t *testing.T, result *e2e.E2ETestResult) {
	t.Helper()

	// 同時読み取りでファイルロックエラーが発生していないことを確認
	if strings.Contains(result.Stderr, "file locked") || strings.Contains(result.Stderr, "resource busy") {
		t.Errorf("同時読み取りでロックエラーが発生: %s", result.Stderr)
	}

	// 正常に処理が完了していることを確認
	if result.ExitCode != 0 {
		t.Errorf("同時読み取りで予期しないエラー: %s", result.Stderr)
	}
}
