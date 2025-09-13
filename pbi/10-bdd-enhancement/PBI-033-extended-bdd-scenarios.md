# PBI-033: 拡張BDDシナリオテストスイート

## 概要
現在のBDD（行動駆動開発）テストスイートを大幅に拡張し、より包括的なユーザーシナリオをカバーします。エラーケース、パフォーマンステスト、複雑なワークフローなど、実際の運用で遭遇する様々な状況を自動テスト化します。

## 受け入れ条件
- [x] エラー処理シナリオを包括的にテストできる
- [x] パフォーマンス要件をBDDシナリオで検証できる
- [x] 複数ファイル・バッチ処理シナリオをテストできる
- [x] ユーザー操作の複雑なフローをテストできる
- [x] 回帰テストとして継続的に実行できる

## 技術仕様

### 1. 拡張シナリオ定義
```gherkin
# features/extended_scenarios.feature

Feature: 拡張ユーザーシナリオ
  usacloud-updateツールの複雑な使用ケースを検証する

  Background:
    Given usacloud CLIがインストールされている
    And 有効なAPIキーが設定されている
    And テスト用のサンドボックス環境が利用可能

  @error-handling
  Scenario: ネットワークエラー時の適切なエラーハンドリング
    Given ネットワーク接続が不安定な状況
    When サンドボックスモードでusacloudコマンドを実行する
    Then 適切なエラーメッセージが表示される
    And リトライ機能が提案される
    And エラーログが記録される

  @performance
  Scenario: 大量ファイル処理のパフォーマンス要件
    Given 100個のスクリプトファイルが存在する
    When バッチモードで全ファイルを処理する
    Then 5分以内に処理が完了する
    And 並行実行数が適切に制御される
    And メモリ使用量が1GB以下に保たれる

  @complex-workflow
  Scenario: マルチプロファイル環境での複雑なワークフロー
    Given 本番用とテスト用のプロファイルが存在する
    When テスト用プロファイルに切り替える
    And スクリプトファイルを変換・実行する
    And 本番用プロファイルに切り替える
    And 同じスクリプトを本番環境で実行する
    Then 各環境で適切な設定が適用される
    And 実行結果が環境ごとに記録される

  @regression
  Scenario: 既知の問題の回帰テスト
    Given 過去に修正されたバグのテストケース
    When 同じ条件でツールを実行する
    Then バグが再発していない
    And 期待通りの結果が得られる
```

### 2. ステップ実装の拡張
```go
// internal/bdd/extended_steps.go

func (s *Steps) networkConnectionIsUnstable() error {
    // ネットワーク遅延やタイムアウトをシミュレート
    s.networkSimulator = &NetworkSimulator{
        latency:     5 * time.Second,
        failureRate: 0.3, // 30%の確率で失敗
        timeout:     10 * time.Second,
    }
    return nil
}

func (s *Steps) sandboxCommandIsExecuted() error {
    // ネットワークシミュレータを有効にして実行
    ctx := context.WithValue(context.Background(), "network_simulator", s.networkSimulator)
    
    result, err := s.executor.ExecuteWithContext(ctx, s.lastCommand)
    s.lastResult = result
    s.lastError = err
    
    return nil
}

func (s *Steps) appropriateErrorMessageIsDisplayed() error {
    if s.lastError == nil {
        return fmt.Errorf("expected error but got none")
    }
    
    expectedMessages := []string{
        "ネットワーク接続に問題があります",
        "タイムアウトが発生しました",
        "再試行することをお勧めします",
    }
    
    errorMessage := s.lastError.Error()
    for _, expected := range expectedMessages {
        if !strings.Contains(errorMessage, expected) {
            return fmt.Errorf("error message missing expected content: %s", expected)
        }
    }
    
    return nil
}

func (s *Steps) scriptsExist(count int) error {
    s.testFiles = make([]string, count)
    
    for i := 0; i < count; i++ {
        filename := fmt.Sprintf("test_script_%03d.sh", i+1)
        filepath := path.Join(s.testDir, filename)
        
        content := s.generateTestScript(i)
        if err := ioutil.WriteFile(filepath, []byte(content), 0644); err != nil {
            return fmt.Errorf("failed to create test file %s: %w", filename, err)
        }
        
        s.testFiles[i] = filepath
    }
    
    return nil
}

func (s *Steps) batchModeProcessesAllFiles() error {
    startTime := time.Now()
    
    // バッチ実行
    results, err := s.executor.ExecuteBatch(s.testFiles)
    if err != nil {
        return fmt.Errorf("batch execution failed: %w", err)
    }
    
    s.batchResults = results
    s.batchDuration = time.Since(startTime)
    
    return nil
}

func (s *Steps) processingCompletesWithinMinutes(minutes int) error {
    maxDuration := time.Duration(minutes) * time.Minute
    
    if s.batchDuration > maxDuration {
        return fmt.Errorf("processing took %v, expected under %v", 
            s.batchDuration, maxDuration)
    }
    
    return nil
}

func (s *Steps) concurrencyIsControlledAppropriately() error {
    // 実行メトリクスから並行数を確認
    metrics := s.executor.GetExecutionMetrics()
    
    if metrics.MaxConcurrentJobs > 10 {
        return fmt.Errorf("too many concurrent jobs: %d (max 10)", 
            metrics.MaxConcurrentJobs)
    }
    
    if metrics.AverageConcurrentJobs < 1 {
        return fmt.Errorf("insufficient parallelism: %f", 
            metrics.AverageConcurrentJobs)
    }
    
    return nil
}

func (s *Steps) memoryUsageIsKeptUnderGB(maxGB int) error {
    metrics := s.executor.GetResourceMetrics()
    maxBytes := int64(maxGB) * 1024 * 1024 * 1024
    
    if metrics.PeakMemoryUsage > maxBytes {
        return fmt.Errorf("memory usage exceeded limit: %d bytes (max %d)", 
            metrics.PeakMemoryUsage, maxBytes)
    }
    
    return nil
}
```

### 3. パフォーマンステスト支援
```go
type PerformanceMetrics struct {
    StartTime           time.Time
    EndTime             time.Time
    Duration            time.Duration
    TotalCommands       int
    SuccessfulCommands  int
    FailedCommands      int
    MaxConcurrentJobs   int
    AverageConcurrentJobs float64
    PeakMemoryUsage     int64
    TotalCPUTime        time.Duration
}

type PerformanceMonitor struct {
    metrics     *PerformanceMetrics
    monitoring  bool
    samples     []ResourceSample
    sampleRate  time.Duration
}

type ResourceSample struct {
    Timestamp     time.Time
    MemoryUsage   int64
    CPUUsage      float64
    GoroutineCount int
    OpenFileCount int
}

func (pm *PerformanceMonitor) StartMonitoring() {
    pm.metrics = &PerformanceMetrics{
        StartTime: time.Now(),
    }
    pm.monitoring = true
    
    // リソース使用量を定期的にサンプリング
    go pm.sampleResources()
}

func (pm *PerformanceMonitor) StopMonitoring() *PerformanceMetrics {
    pm.monitoring = false
    pm.metrics.EndTime = time.Now()
    pm.metrics.Duration = pm.metrics.EndTime.Sub(pm.metrics.StartTime)
    
    // サンプルから統計を計算
    pm.calculateStatistics()
    
    return pm.metrics
}

func (pm *PerformanceMonitor) sampleResources() {
    ticker := time.NewTicker(pm.sampleRate)
    defer ticker.Stop()
    
    for pm.monitoring {
        select {
        case <-ticker.C:
            sample := ResourceSample{
                Timestamp:      time.Now(),
                MemoryUsage:    pm.getCurrentMemoryUsage(),
                CPUUsage:       pm.getCurrentCPUUsage(),
                GoroutineCount: runtime.NumGoroutine(),
                OpenFileCount:  pm.getOpenFileCount(),
            }
            pm.samples = append(pm.samples, sample)
        }
    }
}
```

### 4. エラーシナリオジェネレーター
```go
type ErrorScenarioGenerator struct {
    scenarios []ErrorScenario
}

type ErrorScenario struct {
    Name        string
    Description string
    Setup       func() error
    Trigger     func() error
    Verify      func(error) error
    Cleanup     func() error
}

func (esg *ErrorScenarioGenerator) GenerateNetworkErrors() []ErrorScenario {
    return []ErrorScenario{
        {
            Name:        "connection_timeout",
            Description: "API接続タイムアウト",
            Setup: func() error {
                // タイムアウト設定を短くする
                return os.Setenv("SAKURACLOUD_TIMEOUT", "1")
            },
            Trigger: func() error {
                // 長時間実行されるコマンドを実行
                return executeCommand("usacloud server ls --zone=all")
            },
            Verify: func(err error) error {
                if err == nil {
                    return fmt.Errorf("expected timeout error but got none")
                }
                if !strings.Contains(err.Error(), "timeout") {
                    return fmt.Errorf("expected timeout error, got: %v", err)
                }
                return nil
            },
            Cleanup: func() error {
                return os.Unsetenv("SAKURACLOUD_TIMEOUT")
            },
        },
        {
            Name:        "invalid_api_key",
            Description: "無効なAPIキー",
            Setup: func() error {
                return os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "invalid_token")
            },
            Trigger: func() error {
                return executeCommand("usacloud auth-status")
            },
            Verify: func(err error) error {
                if err == nil {
                    return fmt.Errorf("expected auth error but got none")
                }
                expectedStrings := []string{"authentication", "invalid", "token"}
                errorStr := strings.ToLower(err.Error())
                for _, expected := range expectedStrings {
                    if strings.Contains(errorStr, expected) {
                        return nil
                    }
                }
                return fmt.Errorf("expected auth error, got: %v", err)
            },
            Cleanup: func() error {
                return os.Unsetenv("SAKURACLOUD_ACCESS_TOKEN")
            },
        },
    }
}
```

### 5. 継続的テスト実行システム
```go
type ContinuousTestRunner struct {
    scenarios    []string
    schedule     *cron.Cron
    results      *TestResultStore
    notifications *NotificationService
}

func (ctr *ContinuousTestRunner) ScheduleTests() {
    // 毎日午前3時に実行
    ctr.schedule.AddFunc("0 3 * * *", ctr.runFullTestSuite)
    
    // 毎時実行（軽量テストのみ）
    ctr.schedule.AddFunc("0 * * * *", ctr.runSmokeTests)
    
    ctr.schedule.Start()
}

func (ctr *ContinuousTestRunner) runFullTestSuite() {
    log.Println("Starting full BDD test suite...")
    
    startTime := time.Now()
    results := make(map[string]TestResult)
    
    for _, scenario := range ctr.scenarios {
        result := ctr.runScenario(scenario)
        results[scenario] = result
        
        // 失敗した場合は即座に通知
        if !result.Passed {
            ctr.notifications.SendAlert(fmt.Sprintf(
                "BDD Test Failed: %s - %s", scenario, result.Error))
        }
    }
    
    duration := time.Since(startTime)
    
    // 結果をストアに保存
    testRun := TestRun{
        ID:        generateRunID(),
        Timestamp: startTime,
        Duration:  duration,
        Results:   results,
    }
    
    ctr.results.Save(testRun)
    
    // サマリー通知
    ctr.sendSummaryNotification(testRun)
}

func (ctr *ContinuousTestRunner) sendSummaryNotification(run TestRun) {
    passed := 0
    failed := 0
    
    for _, result := range run.Results {
        if result.Passed {
            passed++
        } else {
            failed++
        }
    }
    
    message := fmt.Sprintf(
        "BDD Test Run Complete\nPassed: %d, Failed: %d, Duration: %v",
        passed, failed, run.Duration)
    
    if failed > 0 {
        ctr.notifications.SendAlert(message)
    } else {
        ctr.notifications.SendInfo(message)
    }
}
```

## テスト戦略
- **シナリオ網羅性テスト**: 実際のユーザーワークフローとの対応確認
- **エラーケーステスト**: 各種エラー条件での適切な処理確認
- **パフォーマンス回帰テスト**: 性能劣化の早期検出
- **並行実行テスト**: BDDシナリオの並行実行安定性確認

## 依存関係
- 前提PBI: なし（既存BDD機能を拡張）
- 関連PBI: PBI-034（テスト報告自動化）、PBI-035（テストデータ管理）
- 既存コード: internal/bdd/、features/

## 見積もり
- 開発工数: 15時間
  - 拡張シナリオ定義・実装: 6時間
  - パフォーマンステスト機能: 4時間
  - エラーシナリオジェネレーター: 3時間
  - 継続的実行システム: 2時間

## 完了の定義
- [x] 20個以上の拡張BDDシナリオが実装される
- [x] パフォーマンス要件がBDDで自動検証される
- [x] エラーケースが包括的にテストされる
- [x] 継続的テスト実行が自動化される
- [x] テスト結果の通知システムが機能する

## 実装状況
✅ **PBI-033は完全に実装済み** (2025-09-11)

以下のファイルで完全に実装されています：
- `internal/bdd/extended_steps.go` - 拡張BDDステップ定義
- `internal/bdd/error_scenarios.go` - エラーシナリオテスト
- `internal/bdd/continuous_test.go` - 継続的テスト実行システム
- `features/extended_scenarios.feature` - 拡張BDDシナリオ定義
- `features/sandbox.feature` - サンドボックス機能シナリオ

実装内容：
- 20+の包括的なBDDシナリオスイート
- エラー処理・パフォーマンス・ワークフローテスト
- 自動的なテスト結果通知システム
- CI/CDパイプライン統合
- 実際のユーザーシナリオに基づいたテストケース
- Godogフレームワークの最大活用

## 備考
- Godog（Go用BDDフレームワーク）を最大限活用
- CI/CDパイプラインとの統合を考慮
- テストデータの管理とクリーンアップを自動化
- 実際のユーザーフィードバックを元にシナリオを継続更新