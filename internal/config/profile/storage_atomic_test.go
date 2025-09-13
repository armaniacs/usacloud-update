package profile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileStorage_AtomicWrites(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	// テスト用プロファイル
	profile := &Profile{
		ID:          "test-atomic",
		Name:        "Atomic Test Profile",
		Description: "Testing atomic writes",
		Environment: "test",
		Config: map[string]string{
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 並行書き込みテスト
	var wg sync.WaitGroup
	const numWriters = 10
	errors := make(chan error, numWriters)

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			// 異なる内容で同じプロファイルを書き込み
			testProfile := *profile
			testProfile.Description = fmt.Sprintf("Writer %d content", writerID)
			testProfile.UpdatedAt = time.Now()

			if err := storage.Save(&testProfile); err != nil {
				errors <- fmt.Errorf("writer %d failed: %w", writerID, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// エラー確認
	for err := range errors {
		t.Errorf("Concurrent write error: %v", err)
	}

	// ファイルが正常に読み取れることを確認
	loadedProfile, err := storage.Load(profile.ID)
	if err != nil {
		t.Fatalf("Failed to load profile after concurrent writes: %v", err)
	}

	if loadedProfile.ID != profile.ID {
		t.Errorf("Profile ID mismatch: expected %s, got %s", profile.ID, loadedProfile.ID)
	}

	// ファイルが有効なYAMLであることを確認
	filename := storage.getProfileFilename(profile.ID)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read profile file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Profile file is empty")
	}

	// YAMLの基本構造を確認
	content := string(data)
	if !strings.Contains(content, "id: test-atomic") {
		t.Error("Profile file does not contain expected ID")
	}
}

func TestFileStorage_CorruptionPrevention(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "corruption-test",
		Name:        "Corruption Prevention Test",
		Description: "Testing corruption prevention",
		Environment: "test",
		Config: map[string]string{
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 正常な保存
	if err := storage.Save(profile); err != nil {
		t.Fatalf("Initial save failed: %v", err)
	}

	filename := storage.getProfileFilename(profile.ID)

	// ファイルの存在確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("Profile file was not created")
	}

	// 権限確認
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat profile file: %v", err)
	}

	expectedMode := os.FileMode(0600)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Incorrect file permissions: expected %v, got %v", expectedMode, info.Mode().Perm())
	}

	// 読み取り確認
	loadedProfile, err := storage.Load(profile.ID)
	if err != nil {
		t.Fatalf("Failed to load saved profile: %v", err)
	}

	if loadedProfile.Name != profile.Name {
		t.Errorf("Profile name mismatch: expected %s, got %s", profile.Name, loadedProfile.Name)
	}
}

func TestFileStorage_TempFileCleanup(t *testing.T) {
	tempDir := t.TempDir()
	_ = NewFileStorage(tempDir)

	// 無効なデータでセーブ操作を試行（エラーを発生させる）
	invalidProfile := struct {
		Name     string
		Channel  chan string // YAMLにシリアライズできない型
		Channels []chan string
	}{
		Name:     "Invalid",
		Channel:  make(chan string),
		Channels: []chan string{make(chan string)},
	}

	// 保存試行（失敗すべき - パニックまたはエラー）
	defer func() {
		if r := recover(); r != nil {
			// パニックが発生した場合は期待通り
			if !strings.Contains(fmt.Sprintf("%v", r), "cannot marshal type") {
				t.Errorf("Expected marshalling panic, got: %v", r)
			}
		}
	}()

	err := writeYAMLFile(filepath.Join(tempDir, "invalid.yaml"), invalidProfile)
	if err == nil {
		t.Error("Expected YAML encoding to fail for invalid data")
		return
	}

	// エラーが期待通り発生したことを確認
	if !strings.Contains(err.Error(), "cannot marshal type") && !strings.Contains(err.Error(), "failed to encode YAML") {
		t.Errorf("Expected marshalling error, got: %v", err)
	}

	// 一時ファイルがクリーンアップされていることを確認
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "profile-") && strings.Contains(file.Name(), ".tmp") {
			t.Errorf("Temporary file was not cleaned up: %s", file.Name())
		}
	}
}

func TestFileStorage_ConcurrentReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	storage := NewFileStorage(tempDir)

	profile := &Profile{
		ID:          "concurrent-rw",
		Name:        "Concurrent Read-Write Test",
		Description: "Testing concurrent read and write operations",
		Environment: "test",
		Config: map[string]string{
			"SAKURACLOUD_ACCESS_TOKEN":        "test-token",
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 初期保存
	if err := storage.Save(profile); err != nil {
		t.Fatalf("Initial save failed: %v", err)
	}

	var wg sync.WaitGroup
	const numReaders = 5
	const numWriters = 3
	const iterations = 10

	errors := make(chan error, (numReaders+numWriters)*iterations)

	// リーダー goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				_, err := storage.Load(profile.ID)
				if err != nil {
					errors <- fmt.Errorf("reader %d iteration %d failed: %w", readerID, j, err)
				}
				time.Sleep(time.Millisecond * 1)
			}
		}(i)
	}

	// ライター goroutines
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				updateProfile := *profile
				updateProfile.Description = fmt.Sprintf("Writer %d iteration %d", writerID, j)
				updateProfile.UpdatedAt = time.Now()

				if err := storage.Save(&updateProfile); err != nil {
					errors <- fmt.Errorf("writer %d iteration %d failed: %w", writerID, j, err)
				}
				time.Sleep(time.Millisecond * 2)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// エラー確認
	for err := range errors {
		t.Errorf("Concurrent read-write error: %v", err)
	}

	// 最終状態確認
	finalProfile, err := storage.Load(profile.ID)
	if err != nil {
		t.Fatalf("Failed to load final profile: %v", err)
	}

	if finalProfile.ID != profile.ID {
		t.Errorf("Final profile ID mismatch: expected %s, got %s", profile.ID, finalProfile.ID)
	}
}
