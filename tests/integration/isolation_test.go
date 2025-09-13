package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config/profile"
)

// IsolatedTestEnvironment provides a completely isolated test environment
type IsolatedTestEnvironment struct {
	tempDir     string
	configDir   string
	originalEnv map[string]string
	cleanup     func()
	t           *testing.T
}

// NewIsolatedTestEnvironment creates a new isolated test environment
func NewIsolatedTestEnvironment(t *testing.T) *IsolatedTestEnvironment {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// 完全に分離された環境変数設定
	originalEnv := make(map[string]string)
	envVars := []string{
		"HOME",
		"USACLOUD_UPDATE_CONFIG_DIR",
		"SAKURACLOUD_ACCESS_TOKEN",
		"SAKURACLOUD_ACCESS_TOKEN_SECRET",
		"SAKURACLOUD_ZONE",
		"USACLOUD_UPDATE_DRY_RUN",
		"USACLOUD_UPDATE_BATCH",
		"USACLOUD_UPDATE_INTERACTIVE",
	}

	// 既存環境変数を保存
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}

	cleanup := func() {
		// 環境変数の復元
		for env, value := range originalEnv {
			if value == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, value)
			}
		}
	}

	// テスト用環境変数設定
	os.Setenv("HOME", tempDir)
	os.Setenv("USACLOUD_UPDATE_CONFIG_DIR", configDir)
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "test_token_isolated")
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "test_secret_isolated")
	os.Setenv("SAKURACLOUD_ZONE", "tk1v")
	os.Setenv("USACLOUD_UPDATE_DRY_RUN", "true")
	os.Setenv("USACLOUD_UPDATE_BATCH", "false")
	os.Setenv("USACLOUD_UPDATE_INTERACTIVE", "false")

	// クリーンアップをt.Cleanupに登録
	t.Cleanup(cleanup)

	return &IsolatedTestEnvironment{
		tempDir:     tempDir,
		configDir:   configDir,
		originalEnv: originalEnv,
		cleanup:     cleanup,
		t:           t,
	}
}

// CreateTestProfile creates a test profile in the isolated environment
func (env *IsolatedTestEnvironment) CreateTestProfile(name string) (*profile.Profile, error) {
	manager, err := profile.NewProfileManager(env.configDir)
	if err != nil {
		return nil, err
	}

	return manager.CreateProfile(profile.ProfileCreateOptions{
		Name:        name,
		Description: "Isolated test profile",
		Environment: "test",
		Config: map[string]string{
			"SAKURACLOUD_ACCESS_TOKEN":        "test_token_" + name,
			"SAKURACLOUD_ACCESS_TOKEN_SECRET": "test_secret_" + name,
			"SAKURACLOUD_ZONE":                "tk1v",
			"test_key":                        "test_value_" + name,
		},
	})
}

// GetTempDir returns the temporary directory path
func (env *IsolatedTestEnvironment) GetTempDir() string {
	return env.tempDir
}

// GetConfigDir returns the config directory path
func (env *IsolatedTestEnvironment) GetConfigDir() string {
	return env.configDir
}

// GetOriginalEnv returns the original environment variable value
func (env *IsolatedTestEnvironment) GetOriginalEnv(key string) string {
	return env.originalEnv[key]
}

// TestIsolatedEnvironment tests the isolation functionality
func TestIsolatedEnvironment(t *testing.T) {
	// 元の環境変数を記録
	originalHome := os.Getenv("HOME")
	_ = os.Getenv("USACLOUD_UPDATE_CONFIG_DIR") // 記録のみ

	// 分離環境作成
	env1 := NewIsolatedTestEnvironment(t)

	// 環境変数が分離されているか確認
	currentHome := os.Getenv("HOME")
	if currentHome == originalHome && originalHome != "" {
		t.Error("HOME environment variable was not isolated")
	}

	currentConfigDir := os.Getenv("USACLOUD_UPDATE_CONFIG_DIR")
	if currentConfigDir != env1.GetConfigDir() {
		t.Errorf("USACLOUD_UPDATE_CONFIG_DIR mismatch: expected %s, got %s",
			env1.GetConfigDir(), currentConfigDir)
	}

	// プロファイル作成テスト
	profile1, err := env1.CreateTestProfile("isolated-test-1")
	if err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	if profile1.Name != "isolated-test-1" {
		t.Errorf("Profile name mismatch: expected isolated-test-1, got %s", profile1.Name)
	}

	// 別の分離環境を作成
	env2 := NewIsolatedTestEnvironment(t)

	// 異なる環境であることを確認
	if env1.GetTempDir() == env2.GetTempDir() {
		t.Error("Two isolated environments have the same temp directory")
	}

	// 2つ目の環境でもプロファイル作成
	_, err = env2.CreateTestProfile("isolated-test-2")
	if err != nil {
		t.Fatalf("Failed to create test profile in second environment: %v", err)
	}

	// プロファイルが分離されていることを確認
	manager1, err := profile.NewProfileManager(env1.GetConfigDir())
	if err != nil {
		t.Fatalf("Failed to create profile manager for env1: %v", err)
	}

	manager2, err := profile.NewProfileManager(env2.GetConfigDir())
	if err != nil {
		t.Fatalf("Failed to create profile manager for env2: %v", err)
	}

	// env1にはprofile1のみ、env2にはprofile2のみ存在することを確認
	profiles1 := manager1.ListProfiles()
	profiles2 := manager2.ListProfiles()

	found1InEnv1 := false
	found2InEnv1 := false
	for _, p := range profiles1 {
		if p.Name == "isolated-test-1" {
			found1InEnv1 = true
		}
		if p.Name == "isolated-test-2" {
			found2InEnv1 = true
		}
	}

	found1InEnv2 := false
	found2InEnv2 := false
	for _, p := range profiles2 {
		if p.Name == "isolated-test-1" {
			found1InEnv2 = true
		}
		if p.Name == "isolated-test-2" {
			found2InEnv2 = true
		}
	}

	if !found1InEnv1 {
		t.Error("Profile 1 not found in environment 1")
	}
	if found2InEnv1 {
		t.Error("Profile 2 found in environment 1 (should be isolated)")
	}
	if found1InEnv2 {
		t.Error("Profile 1 found in environment 2 (should be isolated)")
	}
	if !found2InEnv2 {
		t.Error("Profile 2 not found in environment 2")
	}
}

func TestEnvironmentVariableRestoration(t *testing.T) {
	// 元の値を記録
	originalHome := os.Getenv("HOME")
	originalToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")

	// カスタム値設定
	testHome := "/custom/test/home"
	testToken := "custom_test_token"
	os.Setenv("HOME", testHome)
	os.Setenv("SAKURACLOUD_ACCESS_TOKEN", testToken)

	// 確認
	if os.Getenv("HOME") != testHome {
		t.Fatal("Failed to set custom HOME")
	}
	if os.Getenv("SAKURACLOUD_ACCESS_TOKEN") != testToken {
		t.Fatal("Failed to set custom token")
	}

	// 分離環境作成（内部で別の値を設定）
	{
		env := NewIsolatedTestEnvironment(t)

		// 分離環境内での値確認
		isolatedHome := os.Getenv("HOME")
		isolatedToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")

		if isolatedHome == testHome {
			t.Error("Environment was not properly isolated")
		}
		if isolatedToken == testToken {
			t.Error("Token was not properly isolated")
		}

		// 明示的クリーンアップ（t.Cleanupも動作するがテスト用）
		env.cleanup()
	}

	// 環境変数が復元されていることを確認
	restoredHome := os.Getenv("HOME")
	restoredToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")

	if restoredHome != testHome {
		t.Errorf("HOME not restored: expected %s, got %s", testHome, restoredHome)
	}
	if restoredToken != testToken {
		t.Errorf("Token not restored: expected %s, got %s", testToken, restoredToken)
	}

	// 元の値に戻す
	if originalHome == "" {
		os.Unsetenv("HOME")
	} else {
		os.Setenv("HOME", originalHome)
	}
	if originalToken == "" {
		os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN")
	} else {
		os.Setenv("SAKURACLOUD_ACCESS_TOKEN", originalToken)
	}
}

func TestConcurrentIsolatedEnvironments(t *testing.T) {
	// 複数の分離環境を並行して使用
	const numEnvironments = 5

	results := make(chan bool, numEnvironments)

	for i := 0; i < numEnvironments; i++ {
		go func(envID int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Environment %d panicked: %v", envID, r)
					results <- false
					return
				}
				results <- true
			}()

			// 各goroutineで独立したテスト実行
			env := NewIsolatedTestEnvironment(t)

			// プロファイル作成
			profileName := fmt.Sprintf("concurrent-test-%d", envID)
			profile, err := env.CreateTestProfile(profileName)
			if err != nil {
				t.Errorf("Environment %d failed to create profile: %v", envID, err)
				return
			}

			// 正しく作成されたか確認
			if profile.Name != profileName {
				t.Errorf("Environment %d profile name mismatch: expected %s, got %s",
					envID, profileName, profile.Name)
				return
			}

			// プロファイルの内容確認
			expectedConfigValue := "test_value_" + profileName
			if profile.Config["test_key"] != expectedConfigValue {
				t.Errorf("Environment %d config mismatch: expected %s, got %s",
					envID, expectedConfigValue, profile.Config["test_key"])
				return
			}
		}(i)
	}

	// 全ての環境が成功することを確認
	for i := 0; i < numEnvironments; i++ {
		select {
		case success := <-results:
			if !success {
				t.Errorf("Environment %d failed", i)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("Environment %d timed out", i)
		}
	}
}
