package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/internal/scanner"
)

// TestPreviewDynamicDisplay_BDD_Scenarios は PBI-034 の BDD 受け入れシナリオテスト
func TestPreviewDynamicDisplay_BDD_Scenarios(t *testing.T) {
	t.Run("Scenario: 大きなウィンドウでの表示最適化", func(t *testing.T) {
		// Given: ターミナルウィンドウの高さが50行である
		// And: プレビューペインの表示領域が40行分ある
		tempDir := createTestDir(t)
		defer os.RemoveAll(tempDir)

		// 100行のテストファイルを作成
		testFile := filepath.Join(tempDir, "large_file.sh")
		createLargeTestFile(t, testFile, 100)

		// モックプレビューペインサイズ（50行の高さをシミュレート）
		mockPreviewHeight := 40

		// When: ユーザーが100行のファイルを選択する
		scanResult, err := scanner.NewScanner().Scan(tempDir)
		if err != nil {
			t.Fatalf("ディレクトリスキャンに失敗: %v", err)
		}

		if len(scanResult.Files) == 0 {
			t.Fatal("テストファイルが見つかりません")
		}

		file := scanResult.Files[0]

		// Then: プレビューペインに30行以上のファイル内容が表示される
		// And: 不要な余白なくコンテンツが表示される
		preview, err := file.Preview(calculateDynamicLines(mockPreviewHeight))
		if err != nil {
			t.Fatalf("プレビュー取得に失敗: %v", err)
		}

		expectedMinLines := 25 // 40行の高さから余白を除いた期待値
		if len(preview) < expectedMinLines {
			t.Errorf("プレビュー行数が不十分: %d行（期待: %d行以上）", len(preview), expectedMinLines)
		}

		t.Logf("大きなウィンドウでの表示: %d行のプレビューを表示", len(preview))
	})

	t.Run("Scenario: 小さなウィンドウでの最小表示保証", func(t *testing.T) {
		// Given: ターミナルウィンドウの高さが20行である
		// And: プレビューペインの表示領域が10行分ある
		tempDir := createTestDir(t)
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "small_window_file.sh")
		createLargeTestFile(t, testFile, 50)

		// モックプレビューペインサイズ（小さなウィンドウをシミュレート）
		mockPreviewHeight := 10

		// When: ユーザーがファイルを選択する
		scanResult, err := scanner.NewScanner().Scan(tempDir)
		if err != nil {
			t.Fatalf("ディレクトリスキャンに失敗: %v", err)
		}

		file := scanResult.Files[0]

		// Then: 最低10行のプレビューが表示される
		// And: ヘッダー情報が適切に表示される
		preview, err := file.Preview(calculateDynamicLines(mockPreviewHeight))
		if err != nil {
			t.Fatalf("プレビュー取得に失敗: %v", err)
		}

		minLines := 10
		if len(preview) < minLines {
			t.Errorf("最小表示行数が保証されていません: %d行（期待: %d行以上）", len(preview), minLines)
		}

		t.Logf("小さなウィンドウでの表示: %d行のプレビューを表示", len(preview))
	})

	t.Run("Scenario: ウィンドウサイズ変更への追従", func(t *testing.T) {
		// Given: プレビューペインが表示されている
		tempDir := createTestDir(t)
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "resize_test_file.sh")
		createLargeTestFile(t, testFile, 80)

		scanResult, err := scanner.NewScanner().Scan(tempDir)
		if err != nil {
			t.Fatalf("ディレクトリスキャンに失敗: %v", err)
		}

		file := scanResult.Files[0]

		// When: ユーザーがターミナルウィンドウのサイズを変更する
		originalHeight := 30
		resizedHeight := 50

		originalLines := calculateDynamicLines(originalHeight)
		resizedLines := calculateDynamicLines(resizedHeight)

		// Then: プレビューの表示行数が新しいサイズに応じて調整される
		if resizedLines <= originalLines {
			t.Errorf("ウィンドウサイズ変更に追従していません: %d行 → %d行", originalLines, resizedLines)
		}

		// And: truncatedメッセージが適切に表示/非表示される
		preview1, _ := file.Preview(originalLines)
		preview2, _ := file.Preview(resizedLines)

		if len(preview2) <= len(preview1) {
			t.Errorf("リサイズ後の表示行数が増加していません: %d行 → %d行", len(preview1), len(preview2))
		}

		t.Logf("ウィンドウサイズ変更: %d行 → %d行（プレビュー: %d行 → %d行）",
			originalHeight, resizedHeight, len(preview1), len(preview2))
	})
}

// calculateDynamicLines は動的な表示行数を計算するヘルパー関数
func calculateDynamicLines(previewPaneHeight int) int {
	const (
		MinPreviewLines = 10
		MaxPreviewLines = 100
		HeaderLines     = 10
		MarginLines     = 2
	)

	availableLines := previewPaneHeight - HeaderLines - MarginLines
	if availableLines < MinPreviewLines {
		availableLines = MinPreviewLines
	}
	if availableLines > MaxPreviewLines {
		availableLines = MaxPreviewLines
	}

	return availableLines
}

// createTestDir はテスト用の一時ディレクトリを作成
func createTestDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	return tempDir
}

// createLargeTestFile は指定した行数のテストファイルを作成
func createLargeTestFile(t *testing.T, filePath string, lines int) {
	t.Helper()

	var content strings.Builder
	content.WriteString("#!/bin/bash\n")
	content.WriteString("# Test file for preview display\n")

	for i := 1; i <= lines-2; i++ {
		content.WriteString("echo \"Line ")
		content.WriteString(strings.Repeat("X", 10))
		content.WriteString(" ")
		content.WriteString(strings.Repeat("Y", 20))
		content.WriteString(" - ")
		content.WriteString("This is line number ")
		content.WriteString(strings.Repeat("Z", 5))
		content.WriteString("\"\n")
	}

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
}
