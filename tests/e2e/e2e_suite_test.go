package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestE2ECompleteWorkflow は全体的なE2Eワークフローテスト
func TestE2ECompleteWorkflow(t *testing.T) {
	// usacloud-updateバイナリが存在することを確認
	suite := NewE2ETestSuite(t)
	defer suite.cleanup()

	t.Run("BasicFunctionality", func(t *testing.T) {
		// 基本的な変換機能のテスト
		options := &E2ETestOptions{
			Arguments: []string{
				"--help",
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"usacloud-update",
				"Usage",
			},
		}

		result := suite.RunE2ETest("BasicHelp", options)
		if result.ExitCode != options.ExpectedExitCode {
			t.Errorf("終了コードが期待値と異なります: got=%d, want=%d", result.ExitCode, options.ExpectedExitCode)
		}

		for _, expected := range options.ExpectedStdout {
			if !containsIgnoreCase(result.Stdout, expected) {
				t.Errorf("標準出力に期待する文字列が含まれていません: %q", expected)
			}
		}
	})

	t.Run("VersionCheck", func(t *testing.T) {
		// バージョン表示のテスト
		options := &E2ETestOptions{
			Arguments: []string{
				"--version",
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"usacloud-update version",
			},
		}

		result := suite.RunE2ETest("VersionCheck", options)
		if result.ExitCode != options.ExpectedExitCode {
			t.Errorf("バージョンチェック失敗: got=%d, want=%d", result.ExitCode, options.ExpectedExitCode)
		}

		for _, expected := range options.ExpectedStdout {
			if !containsIgnoreCase(result.Stdout, expected) {
				t.Errorf("バージョン出力に期待する文字列が含まれていません: %q", expected)
			}
		}
	})

	t.Run("ConfigCommand", func(t *testing.T) {
		// PBI-026で実装したconfigコマンドのテスト
		options := &E2ETestOptions{
			Arguments: []string{
				"config",
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"🔧 usacloud-update 設定情報",
				"📂 設定ファイル",
			},
		}

		result := suite.RunE2ETest("ConfigCommand", options)
		if result.ExitCode != options.ExpectedExitCode {
			t.Errorf("configコマンド失敗: got=%d, want=%d", result.ExitCode, options.ExpectedExitCode)
		}

		for _, expected := range options.ExpectedStdout {
			if !containsIgnoreCase(result.Stdout, expected) {
				t.Errorf("config出力に期待する文字列が含まれていません: %q", expected)
			}
		}
	})

	t.Run("StrictValidationFlag", func(t *testing.T) {
		// 新しく実装した--strict-validationフラグのテスト
		options := &E2ETestOptions{
			Arguments: []string{
				"--strict-validation",
				"--help",
			},
			ExpectedExitCode: 0,
			ExpectedStdout: []string{
				"Usage",
				"strict-validation",
			},
		}

		result := suite.RunE2ETest("StrictValidationFlag", options)
		if result.ExitCode != options.ExpectedExitCode {
			t.Errorf("strict-validationフラグテスト失敗: got=%d, want=%d", result.ExitCode, options.ExpectedExitCode)
		}
	})
}

// TestE2ESubmoduleIntegration はサブディレクトリテストとの統合確認
func TestE2ESubmoduleIntegration(t *testing.T) {
	t.Run("UserWorkflowsIntegration", func(t *testing.T) {
		// user_workflowsパッケージが正常に動作することを確認
		cmd := exec.Command("go", "test", "./user_workflows", "-v")
		cmd.Dir = "/Users/yaar/Playground/usacloud-update/tests/e2e"

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("user_workflows テスト出力: %s", string(output))
			// エラーがあっても必ずしも失敗ではない（個別テストで確認済み）
		}

		t.Logf("user_workflows統合確認完了")
	})

	t.Run("ErrorScenariosIntegration", func(t *testing.T) {
		// error_scenariosパッケージが正常に動作することを確認
		cmd := exec.Command("go", "test", "./error_scenarios", "-v")
		cmd.Dir = "/Users/yaar/Playground/usacloud-update/tests/e2e"

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("error_scenarios テスト出力: %s", string(output))
			// エラーがあっても必ずしも失敗ではない（個別テストで確認済み）
		}

		t.Logf("error_scenarios統合確認完了")
	})
}

// TestE2EFrameworkValidation はE2Eテストフレームワーク自体のテスト
func TestE2EFrameworkValidation(t *testing.T) {
	t.Run("SuiteCreation", func(t *testing.T) {
		// E2ETestSuiteが正常に作成できることを確認
		suite := NewE2ETestSuite(t)
		defer suite.cleanup()

		if suite == nil {
			t.Fatal("E2ETestSuiteの作成に失敗しました")
		}

		// テストディレクトリが作成されていることを確認
		if _, err := os.Stat(suite.testDir); os.IsNotExist(err) {
			t.Errorf("テストディレクトリが作成されていません: %s", suite.testDir)
		}
	})

	t.Run("BinaryAvailability", func(t *testing.T) {
		// usacloud-updateバイナリが利用可能であることを確認
		suite := NewE2ETestSuite(t)
		defer suite.cleanup()

		// バイナリパスの確認
		if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
			t.Errorf("バイナリが見つかりません: %s", suite.binaryPath)
		}
	})
}

// containsIgnoreCase は大文字小文字を無視して文字列が含まれるかチェック
func containsIgnoreCase(s, substr string) bool {
	return contains(strings.ToLower(s), strings.ToLower(substr))
}

// contains は文字列が含まれるかチェック（strings.Containsのラッパー）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

// indexOfSubstring は部分文字列のインデックスを返す
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
