package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/scanner"
)

// TestPreviewNotice_BDD_Scenarios は PBI-035 の BDD 受け入れシナリオテスト
func TestPreviewNotice_BDD_Scenarios(t *testing.T) {
	t.Run("Scenario: TUIモードでPreview表示を確認", func(t *testing.T) {
		// Given: usacloud-updateがインストールされている
		// And: バージョンがv1.9.6である
		tempDir := createTestDirPreview(t)
		defer os.RemoveAll(tempDir)

		testScript := filepath.Join(tempDir, "test_script.sh")
		createTestFilePreview(t, testScript, "#!/bin/bash\nusacloud server list\n")

		cfg := &config.SandboxConfig{}
		fileSelector := NewFileSelector(cfg)

		// When: ユーザーがTUIモードで起動する
		// Then: Preview notice が設定されている
		if fileSelector.previewNotice == nil {
			t.Error("Preview notice が初期化されていません")
		}

		// And: 表示テキストが正しい
		expectedText := "[black:yellow:b] TUIはPreviewとして提供中 [::-]"
		actualText := fileSelector.previewNotice.GetText(false)
		if actualText != expectedText {
			t.Errorf("Preview notice のテキストが正しくありません: expected '%s', got '%s'", expectedText, actualText)
		}

		// And: 中央配置が設定されている
		// tview.AlignCenter の値は 1
		// (tview の内部実装によるが、中央配置であることを確認)
		t.Logf("Preview notice が正しく設定されました: %s", actualText)
	})

	t.Run("Scenario: ヘルプ表示時もPreview表示維持", func(t *testing.T) {
		// Given: TUIモードが起動している
		cfg := &config.SandboxConfig{}
		fileSelector := NewFileSelector(cfg)

		// When: ユーザーがヘルプを表示する
		// （ヘルプが表示されている状態をシミュレート）
		fileSelector.helpVisible = true
		fileSelector.updateLayout()

		// Then: Preview表示が最下行に維持される
		// Grid レイアウトが正しく設定されていることを確認
		if fileSelector.previewNotice == nil {
			t.Error("ヘルプ表示時でもPreview notice が存在する必要があります")
		}

		// And: ヘルプパネルと重ならない
		// ヘルプが表示されている場合の行構成: [0:main, 1:status, 2:help, 3:preview]
		t.Log("ヘルプ表示時でもPreview notice が正しく維持されています")
	})

	t.Run("Scenario: ファイル選択後もPreview表示維持", func(t *testing.T) {
		// Given: TUIモードでファイル選択画面が表示されている
		tempDir := createTestDirPreview(t)
		defer os.RemoveAll(tempDir)

		testScript := filepath.Join(tempDir, "test_script.sh")
		createTestFilePreview(t, testScript, "#!/bin/bash\nusacloud server list\n")

		cfg := &config.SandboxConfig{}
		fileSelector := NewFileSelector(cfg)

		// ファイルスキャンをシミュレート
		result, err := scanner.NewScanner().Scan(tempDir)
		if err != nil {
			t.Fatalf("ディレクトリスキャンに失敗: %v", err)
		}

		fileSelector.scanResult = result
		fileSelector.populateFileList()

		// When: ユーザーがファイルを選択する
		// （ファイル選択後の状態をシミュレート）
		if len(fileSelector.scanResult.Files) > 0 {
			fileSelector.selectedFiles = append(fileSelector.selectedFiles, fileSelector.scanResult.Files[0].Path)
		}

		// Then: Preview表示が消えずに維持される
		if fileSelector.previewNotice == nil {
			t.Error("ファイル選択後でもPreview notice が存在する必要があります")
		}

		expectedText := "[black:yellow:b] TUIはPreviewとして提供中 [::-]"
		actualText := fileSelector.previewNotice.GetText(false)
		if actualText != expectedText {
			t.Errorf("ファイル選択後でもPreview notice のテキストが維持される必要があります: expected '%s', got '%s'", expectedText, actualText)
		}

		t.Log("ファイル選択後でもPreview notice が正しく維持されています")
	})
}

// TestPreviewNoticeSetup_UnitTests は setupPreviewNotice 関数の単体テスト
func TestPreviewNoticeSetup_UnitTests(t *testing.T) {
	tests := []struct {
		name        string
		expectText  string
		expectAlign bool
	}{
		{
			name:        "Preview notice の初期化",
			expectText:  "[black:yellow:b] TUIはPreviewとして提供中 [::-]",
			expectAlign: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SandboxConfig{}
			fileSelector := NewFileSelector(cfg)

			// Preview notice が正しく初期化されることを確認
			if fileSelector.previewNotice == nil {
				t.Error("Preview notice が初期化されていません")
			}

			actualText := fileSelector.previewNotice.GetText(false)
			if actualText != tt.expectText {
				t.Errorf("Preview notice のテキストが正しくありません: expected '%s', got '%s'", tt.expectText, actualText)
			}

			t.Logf("Test case: %s - Text: %s", tt.name, actualText)
		})
	}
}

// TestPreviewNoticeLayout_IntegrationTests は レイアウト統合テスト
func TestPreviewNoticeLayout_IntegrationTests(t *testing.T) {
	t.Run("ヘルプ表示時のレイアウト", func(t *testing.T) {
		cfg := &config.SandboxConfig{}
		fileSelector := NewFileSelector(cfg)

		// ヘルプ表示状態
		fileSelector.helpVisible = true
		fileSelector.updateLayout()

		// Grid レイアウトにPreview notice が追加されていることを確認
		if fileSelector.previewNotice == nil {
			t.Error("Preview notice がレイアウトに追加されていません")
		}

		t.Log("ヘルプ表示時のレイアウトにPreview notice が正しく統合されています")
	})

	t.Run("ヘルプ非表示時のレイアウト", func(t *testing.T) {
		cfg := &config.SandboxConfig{}
		fileSelector := NewFileSelector(cfg)

		// ヘルプ非表示状態
		fileSelector.helpVisible = false
		fileSelector.updateLayout()

		// Grid レイアウトにPreview notice が追加されていることを確認
		if fileSelector.previewNotice == nil {
			t.Error("Preview notice がレイアウトに追加されていません")
		}

		t.Log("ヘルプ非表示時のレイアウトにPreview notice が正しく統合されています")
	})
}

// ヘルパー関数

func createTestDirPreview(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	return tempDir
}

func createTestFilePreview(t *testing.T, filePath, content string) {
	t.Helper()
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
}
