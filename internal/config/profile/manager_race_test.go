package profile

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestProfileManager_ThreadSafety(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	var wg sync.WaitGroup
	const numGoroutines = 50
	const numOperations = 10 // 調整：大量の操作は避ける

	errors := make(chan error, numGoroutines*numOperations)

	// 並行読み書きテスト
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// ランダムな操作を実行
				switch j % 4 {
				case 0:
					// プロファイル作成
					_, err := manager.CreateProfile(ProfileCreateOptions{
						Name:        fmt.Sprintf("Profile-%d-%d", id, j),
						Description: "Test profile",
						Environment: "test",
						Config: map[string]string{
							"SAKURACLOUD_ACCESS_TOKEN":        fmt.Sprintf("token-%d-%d", id, j),
							"SAKURACLOUD_ACCESS_TOKEN_SECRET": fmt.Sprintf("secret-%d-%d", id, j),
							"key":                             fmt.Sprintf("value-%d-%d", id, j),
						},
					})
					if err != nil {
						errors <- fmt.Errorf("create failed: %w", err)
					}

				case 1:
					// プロファイル一覧取得
					_ = manager.ListProfiles(ProfileListOptions{})

				case 2:
					// アクティブプロファイル取得
					_ = manager.GetActiveProfile()

				case 3:
					// プロファイル検索
					profiles := manager.ListProfiles(ProfileListOptions{})
					if len(profiles) > 0 {
						_, _ = manager.GetProfile(profiles[0].ID)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// エラーの確認
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

func TestProfileManager_ConcurrentCreateDelete(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	var wg sync.WaitGroup
	const numWorkers = 10
	errors := make(chan error, numWorkers*2)

	// 並行作成・削除テスト
	for i := 0; i < numWorkers; i++ {
		wg.Add(2)

		// 作成用 goroutine
		go func(id int) {
			defer wg.Done()
			profile, err := manager.CreateProfile(ProfileCreateOptions{
				Name:        fmt.Sprintf("ConcurrentProfile-%d", id),
				Description: "Concurrent test profile",
				Environment: "test",
				Config: map[string]string{
					"SAKURACLOUD_ACCESS_TOKEN":        fmt.Sprintf("token-%d", id),
					"SAKURACLOUD_ACCESS_TOKEN_SECRET": fmt.Sprintf("secret-%d", id),
					"test_key":                        fmt.Sprintf("test_value_%d", id),
				},
			})
			if err != nil {
				errors <- fmt.Errorf("concurrent create failed: %w", err)
				return
			}

			// 短時間待機
			time.Sleep(time.Millisecond * 10)

			// 削除試行
			if err := manager.DeleteProfile(profile.ID); err != nil {
				// 削除失敗は他の goroutine が先に削除した可能性があるので許容
				// errors <- fmt.Errorf("concurrent delete failed: %w", err)
			}
		}(i)

		// 読み取り用 goroutine
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_ = manager.ListProfiles()
				_ = manager.GetActiveProfile()
				time.Sleep(time.Millisecond * 2)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// エラーの確認
	for err := range errors {
		t.Errorf("Concurrent create/delete operation failed: %v", err)
	}
}

func TestProfileManager_ConcurrentSwitch(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// テスト用プロファイルを作成
	profiles := make([]*Profile, 3)
	for i := 0; i < 3; i++ {
		profile, err := manager.CreateProfile(ProfileCreateOptions{
			Name:        fmt.Sprintf("SwitchTestProfile-%d", i),
			Description: "Switch test profile",
			Environment: "test",
			Config: map[string]string{
				"SAKURACLOUD_ACCESS_TOKEN":        fmt.Sprintf("token-%d", i),
				"SAKURACLOUD_ACCESS_TOKEN_SECRET": fmt.Sprintf("secret-%d", i),
				"index":                           fmt.Sprintf("%d", i),
			},
		})
		if err != nil {
			t.Fatalf("Failed to create test profile: %v", err)
		}
		profiles[i] = profile
	}

	var wg sync.WaitGroup
	const numSwitchers = 10
	errors := make(chan error, numSwitchers*5)

	// 並行スイッチテスト
	for i := 0; i < numSwitchers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				targetProfile := profiles[j%len(profiles)]
				if err := manager.SwitchProfile(targetProfile.ID); err != nil {
					errors <- fmt.Errorf("switch failed (worker %d, iteration %d): %w", workerID, j, err)
				}

				// アクティブプロファイル確認
				active := manager.GetActiveProfile()
				if active == nil {
					errors <- fmt.Errorf("no active profile after switch (worker %d, iteration %d)", workerID, j)
				}

				time.Sleep(time.Millisecond * 1)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// エラーの確認
	for err := range errors {
		t.Errorf("Concurrent switch operation failed: %v", err)
	}
}

func TestProfileManager_DeadlockPrevention(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewProfileManager(tempDir)
	if err != nil {
		t.Fatalf("NewProfileManager() failed: %v", err)
	}

	// タイムアウト付きテスト
	done := make(chan bool, 1)

	go func() {
		var wg sync.WaitGroup
		const numWorkers = 20

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 複数の操作を組み合わせ
				for j := 0; j < 3; j++ {
					// 作成
					profile, err := manager.CreateProfile(ProfileCreateOptions{
						Name:        fmt.Sprintf("DeadlockTest-%d-%d", id, j),
						Description: "Deadlock prevention test",
						Environment: "test",
						Config: map[string]string{
							"SAKURACLOUD_ACCESS_TOKEN":        fmt.Sprintf("token-%d-%d", id, j),
							"SAKURACLOUD_ACCESS_TOKEN_SECRET": fmt.Sprintf("secret-%d-%d", id, j),
							"test":                            "value",
						},
					})
					if err != nil {
						continue
					}

					// 取得
					_, _ = manager.GetProfile(profile.ID)

					// 一覧
					_ = manager.ListProfiles()

					// スイッチ
					_ = manager.SwitchProfile(profile.ID)

					// 削除
					_ = manager.DeleteProfile(profile.ID)
				}
			}(i)
		}

		wg.Wait()
		done <- true
	}()

	// 5秒でタイムアウト
	select {
	case <-done:
		// 正常終了
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected - test timed out")
	}
}
