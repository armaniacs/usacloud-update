# PBI-030: Regression Test安定化

## 概要
回帰テストの安定性を向上させ、継続的品質監視体制を確立する。現在のテスト実行で発生する間欠的失敗や環境依存の問題を解消し、CI/CDパイプラインでの信頼性の高い品質ゲートを実現する。

## 受け入れ条件
- [ ] 回帰テストが100%安定して実行できること
- [ ] 環境依存による間欠的失敗が完全に解消されること
- [ ] CI/CD環境での継続的テスト実行が安定すること
- [ ] テスト実行時間が合理的な範囲（5分以内）に収まること

## 技術仕様

### 現在の問題
```bash
# 確認される間欠的失敗パターン
=== RUN   TestGolden_TransformWithValidation
    engine_test.go:XX: Random failure in golden file comparison
--- FAIL: TestGolden_TransformWithValidation (flaky)

=== RUN   TestProfileManager_ConcurrentAccess  
    manager_test.go:XX: Race condition detected
--- FAIL: TestProfileManager_ConcurrentAccess (race condition)

=== RUN   TestIntegration_FileOperations
    integration_test.go:XX: Temporary file cleanup race
--- FAIL: TestIntegration_FileOperations (cleanup race)
```

### 1. 間欠的失敗の根本原因分析・修正
#### レースコンディション対策
```go
// internal/config/profile/manager_race_test.go (新規作成)
func TestProfileManager_ThreadSafety(t *testing.T) {
    tempDir := t.TempDir()
    manager, err := NewProfileManager(tempDir)
    if err != nil {
        t.Fatalf("NewProfileManager() failed: %v", err)
    }
    
    var wg sync.WaitGroup
    const numGoroutines = 50
    const numOperations = 100
    
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
                            "key": fmt.Sprintf("value-%d-%d", id, j),
                        },
                    })
                    if err != nil {
                        errors <- fmt.Errorf("create failed: %w", err)
                    }
                    
                case 1:
                    // プロファイル一覧取得
                    _ = manager.ListProfiles(ProfileListOptions{})
                    
                case 2:
                    // デフォルトプロファイル取得
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
```

#### ファイル操作の原子性確保
```go
// internal/config/profile/manager_atomic.go (修正)
func (pm *ProfileManager) saveProfileAtomic(profile *Profile) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    data, err := yaml.Marshal(profile)
    if err != nil {
        return fmt.Errorf("YAML marshal error: %w", err)
    }
    
    profilePath := filepath.Join(pm.configDir, profile.ID+".yaml")
    tempPath := profilePath + ".tmp." + generateTempSuffix()
    
    // 原子的書き込み
    if err := ioutil.WriteFile(tempPath, data, 0600); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    
    // 原子的移動
    if err := os.Rename(tempPath, profilePath); err != nil {
        os.Remove(tempPath) // 失敗時のクリーンアップ
        return fmt.Errorf("failed to move temp file: %w", err)
    }
    
    return nil
}

func generateTempSuffix() string {
    // 一意な一時ファイル名生成
    return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
}
```

### 2. テスト環境の分離・清掃強化
#### 完全なテスト分離
```go
// tests/integration/isolation_test.go (新規作成)
type IsolatedTestEnvironment struct {
    tempDir     string
    configDir   string
    cleanup     func()
    t          *testing.T
}

func NewIsolatedTestEnvironment(t *testing.T) *IsolatedTestEnvironment {
    tempDir := t.TempDir()
    configDir := filepath.Join(tempDir, "config")
    
    // 完全に分離された環境変数設定
    oldEnv := make(map[string]string)
    envVars := []string{
        "HOME",
        "USACLOUD_UPDATE_CONFIG_DIR",
        "SAKURACLOUD_ACCESS_TOKEN",
        "SAKURACLOUD_ACCESS_TOKEN_SECRET",
    }
    
    for _, env := range envVars {
        oldEnv[env] = os.Getenv(env)
    }
    
    cleanup := func() {
        // 環境変数の復元
        for env, value := range oldEnv {
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
    os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "test_token")
    os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "test_secret")
    
    // クリーンアップをt.Cleanupに登録
    t.Cleanup(cleanup)
    
    return &IsolatedTestEnvironment{
        tempDir:   tempDir,
        configDir: configDir,
        cleanup:   cleanup,
        t:         t,
    }
}

func (env *IsolatedTestEnvironment) CreateTestProfile(name string) (*Profile, error) {
    manager, err := NewProfileManager(env.configDir)
    if err != nil {
        return nil, err
    }
    
    return manager.CreateProfile(ProfileCreateOptions{
        Name:        name,
        Description: "Isolated test profile",
        Environment: "test",
        Config: map[string]string{
            "test_key": "test_value",
        },
    })
}
```

### 3. Golden File テストの安定化
#### 決定論的出力の保証
```go
// internal/transform/golden_stable_test.go (修正)
func TestGolden_TransformWithValidation_Stable(t *testing.T) {
    // 決定論的な時刻設定（テスト用固定時刻）
    fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
    
    engine := NewEngine()
    
    // 入力データの正規化
    inputFile := "testdata/sample_v0_v1_mixed.sh"
    inputData, err := ioutil.ReadFile(inputFile)
    if err != nil {
        t.Fatalf("Failed to read input file: %v", err)
    }
    
    // 行区切りの正規化（CRLF → LF）
    normalizedInput := strings.ReplaceAll(string(inputData), "\r\n", "\n")
    lines := strings.Split(normalizedInput, "\n")
    
    var results []string
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            results = append(results, line)
            continue
        }
        
        result := engine.Transform(line)
        if result.Error != nil {
            t.Fatalf("Transform failed for line '%s': %v", line, result.Error)
        }
        
        // タイムスタンプの正規化
        output := result.Line
        if result.Changed {
            // 決定論的なコメント生成
            output += fmt.Sprintf(" # usacloud-update: converted at %s", fixedTime.Format("2006-01-02"))
        }
        
        results = append(results, output)
    }
    
    // 出力の正規化と比較
    actualOutput := strings.Join(results, "\n")
    
    goldenFile := "testdata/expected_v1_1.sh"
    
    if *updateGolden {
        // Golden file 更新時も正規化されたデータを書き込み
        if err := ioutil.WriteFile(goldenFile, []byte(actualOutput), 0644); err != nil {
            t.Fatalf("Failed to update golden file: %v", err)
        }
        return
    }
    
    expectedData, err := ioutil.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("Failed to read golden file: %v", err)
    }
    
    expectedOutput := strings.ReplaceAll(string(expectedData), "\r\n", "\n")
    
    if actualOutput != expectedOutput {
        t.Errorf("Output differs from golden file")
        
        // 詳細な差分表示
        actualLines := strings.Split(actualOutput, "\n")
        expectedLines := strings.Split(expectedOutput, "\n")
        
        maxLines := len(actualLines)
        if len(expectedLines) > maxLines {
            maxLines = len(expectedLines)
        }
        
        for i := 0; i < maxLines; i++ {
            actualLine := ""
            expectedLine := ""
            
            if i < len(actualLines) {
                actualLine = actualLines[i]
            }
            if i < len(expectedLines) {
                expectedLine = expectedLines[i]
            }
            
            if actualLine != expectedLine {
                t.Errorf("Line %d differs:\nActual:   %q\nExpected: %q", i+1, actualLine, expectedLine)
            }
        }
    }
}
```

### 4. CI/CD統合とモニタリング
#### 継続的品質監視
```go
// tests/regression/ci_stability_test.go (新規作成)
func TestCI_StabilityMetrics(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stability test in short mode")
    }
    
    const numRuns = 10
    var successCount int
    var totalDuration time.Duration
    
    for i := 0; i < numRuns; i++ {
        start := time.Now()
        
        // 全テストスイート実行
        err := runFullTestSuite()
        duration := time.Since(start)
        totalDuration += duration
        
        if err == nil {
            successCount++
        } else {
            t.Logf("Test run %d failed: %v", i+1, err)
        }
    }
    
    successRate := float64(successCount) / float64(numRuns)
    avgDuration := totalDuration / numRuns
    
    // 成功率95%以上を要求
    if successRate < 0.95 {
        t.Errorf("Success rate too low: %.2f%% (expected >= 95%%)", successRate*100)
    }
    
    // 平均実行時間5分以内を要求
    if avgDuration > 5*time.Minute {
        t.Errorf("Average duration too long: %v (expected <= 5m)", avgDuration)
    }
    
    t.Logf("Stability metrics: success rate %.2f%%, avg duration %v", successRate*100, avgDuration)
}

func runFullTestSuite() error {
    cmd := exec.Command("go", "test", "./...")
    cmd.Env = append(os.Environ(), 
        "GOCACHE=off",     // キャッシュ無効化
        "GOFLAGS=-count=1", // 実行結果キャッシュ無効化
    )
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("test suite failed: %w\nOutput: %s", err, string(output))
    }
    
    return nil
}
```

## テスト戦略
- **安定性テスト**: 複数回実行での成功率測定
- **パフォーマンス監視**: 実行時間の継続的監視
- **環境分離テスト**: 完全に分離された環境での動作確認
- **CI/CD統合テスト**: 継続的統合環境での安定性確認

## 依存関係
- 前提PBI: PBI-024〜029（全テスト修復・改善完了後）
- 関連PBI: なし（最終段階のPBI）
- 既存コード: 全テストコード、CI/CDスクリプト

## 見積もり
- 開発工数: 6時間
  - レースコンディション対策: 2時間
  - テスト環境分離強化: 2時間
  - Golden Fileテスト安定化: 1時間
  - CI/CD統合・モニタリング: 1時間

## 完了の定義
- [ ] 回帰テストの成功率が95%以上達成
- [ ] テスト実行時間が5分以内に安定
- [ ] 間欠的失敗が完全解消
- [ ] CI/CD環境での継続的実行が安定
- [ ] 品質メトリクスの継続監視体制確立

## 備考
- 回帰テストの安定化は品質保証体制の最終仕上げ
- CI/CDパイプラインの信頼性向上により、開発速度と品質を両立
- 継続的品質監視により、将来の品質劣化を早期検出可能

---

## 実装状況 (2025-09-11)

🟠 **PBI-030は未実装** (2025-09-11)

### 現在の状況
- 回帰テストで間欠的失敗や環境依存の問題が発生
- レースコンディションやファイルクリーンアップの競合状態が存在
- CI/CDパイプラインでの信頼性の高い品質ゲートが未確立
- 継続的品質監視体制が未構築

### 未実装要素
1. **間欠的失敗の根本原因分析・修正**
   - レースコンディション対策の実装
   - ファイルクリーンアップ競合状態の解決
   - Golden Fileテストの安定化
   - 環境依存問題の完全解消

2. **テスト環境分離強化**
   - 完全に分離された環境での動作確認
   - 並列実行時の競合状態防止
   - テストデータの安全な分離と管理
   - メモリリークやリソースリークの防止

3. **CI/CD統合・モニタリング**
   - 継続的品質監視体制の確立
   - テスト実行時間の最適化（5分以内）
   - 品質メトリクスの継続監視
   - CI/CDパイプラインの信頼性向上

### 次のステップ
1. PBI-024〜029の全テスト修復・改善完了待ち
2. 間欠的失敗の根本原因分析とレースコンディション対策
3. テスト環境分離強化と競合状態防止
4. Golden Fileテストの安定化と環境依存問題解決
5. CI/CD統合・モニタリング体制の確立

### 技術要件
- Go 1.24.1対応
- マルチスレッド安全性とレースコンディション対策
- CI/CDパイプラインでの安定した継続実行
- 6時間の作業見積もり

### 受け入れ条件の進捗
- [ ] 回帰テストが100%安定して実行できること
- [ ] 環境依存による間欠的失敗が完全に解消されること
- [ ] CI/CD環境での継続的テスト実行が安定すること
- [ ] テスト実行時間が合理的な範囲（5分以内）に収まること