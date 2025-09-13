# PBI-023: 自動回帰テスト

## 概要
コードベースの変更が既存機能に悪影響を与えないことを継続的に監視する包括的な自動回帰テストシステムを実装する。機能回帰、パフォーマンス回帰、互換性回帰を早期発見し、品質の継続的維持を保証する自動化された検証プロセスを構築する。

## 受け入れ条件
- [ ] 既存機能の完全な後方互換性が継続的に検証されている
- [ ] パフォーマンス回帰が自動検出され早期に警告される
- [ ] API・CLI互換性の破壊的変更が確実に検出される
- [ ] 回帰テスト結果が分かりやすいレポートで提供される
- [ ] CI/CDパイプラインでの自動実行が安定している

## 技術仕様

### 回帰テストアーキテクチャ

#### 1. 回帰テスト分類体系
```
tests/
├── regression/
│   ├── functional/                         # 機能回帰テスト
│   │   ├── transform_regression_test.go    # 変換機能回帰
│   │   ├── validation_regression_test.go   # 検証機能回帰
│   │   ├── cli_regression_test.go          # CLI機能回帰
│   │   └── integration_regression_test.go  # 統合機能回帰
│   ├── performance/                        # パフォーマンス回帰テスト
│   │   ├── speed_regression_test.go        # 実行速度回帰
│   │   ├── memory_regression_test.go       # メモリ使用量回帰
│   │   ├── scalability_regression_test.go  # スケーラビリティ回帰
│   │   └── resource_regression_test.go     # リソース効率回帰
│   ├── compatibility/                      # 互換性回帰テスト
│   │   ├── api_compatibility_test.go       # API互換性回帰
│   │   ├── cli_compatibility_test.go       # CLI互換性回帰
│   │   ├── config_compatibility_test.go    # 設定互換性回帰
│   │   └── output_compatibility_test.go    # 出力互換性回帰
│   └── baseline/                           # ベースライン管理
│       ├── v1.0/                          # v1.0基準値
│       ├── v1.1/                          # v1.1基準値
│       ├── current/                       # 現在の基準値
│       └── snapshots/                     # 定期スナップショット
├── testdata/
│   ├── regression_scenarios/              # 回帰シナリオ
│   │   ├── legacy_scripts/                # レガシースクリプト
│   │   ├── edge_cases/                    # エッジケース
│   │   ├── performance_benchmarks/       # パフォーマンスベンチマーク
│   │   └── compatibility_samples/        # 互換性サンプル
│   └── baselines/                         # ベースラインデータ
└── reports/                               # テストレポート
    ├── daily/                             # 日次レポート
    ├── release/                           # リリース前レポート
    └── historical/                        # 履歴データ
```

#### 2. 回帰テストフレームワーク
```go
// tests/regression/regression_test_framework.go
package regression

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/armaniacs/usacloud-update/internal/config"
    "github.com/armaniacs/usacloud-update/internal/transform" 
    "github.com/armaniacs/usacloud-update/internal/validation"
)

// RegressionTestSuite は回帰テストスイート
type RegressionTestSuite struct {
    t              *testing.T
    testDataDir    string
    baselineDir    string
    reportsDir     string
    
    // ベースライン管理
    currentBaseline   *RegressionBaseline
    previousBaseline  *RegressionBaseline
    
    // 回帰検出設定
    tolerance         *RegressionTolerance
    
    // テスト実行環境
    testEnvironment   *TestEnvironment
}

// RegressionBaseline は回帰テストベースライン
type RegressionBaseline struct {
    Version          string                 `json:"version"`
    CreatedAt        time.Time             `json:"created_at"`
    
    // 機能ベースライン
    FunctionalTests  map[string]*FunctionalResult `json:"functional_tests"`
    
    // パフォーマンスベースライン
    PerformanceTests map[string]*PerformanceResult `json:"performance_tests"`
    
    // 互換性ベースライン
    CompatibilityTests map[string]*CompatibilityResult `json:"compatibility_tests"`
    
    // メタデータ
    TestEnvironment  *TestEnvironment       `json:"test_environment"`
    GitCommit        string                `json:"git_commit"`
    BuildInfo        *BuildInfo            `json:"build_info"`
}

// FunctionalResult は機能テスト結果
type FunctionalResult struct {
    TestName        string            `json:"test_name"`
    InputHash       string            `json:"input_hash"`
    OutputHash      string            `json:"output_hash"`
    TransformRules  []string          `json:"transform_rules"`
    ValidationIssues int              `json:"validation_issues"`
    ExecutionTime   time.Duration     `json:"execution_time"`
    Success         bool              `json:"success"`
    ErrorMessage    string            `json:"error_message,omitempty"`
}

// PerformanceResult はパフォーマンステスト結果
type PerformanceResult struct {
    TestName         string            `json:"test_name"`
    ExecutionTime    time.Duration     `json:"execution_time"`
    MemoryUsage      int64             `json:"memory_usage"`
    ProcessingRate   float64           `json:"processing_rate"`
    Throughput       float64           `json:"throughput"`
    ResourceEfficiency float64         `json:"resource_efficiency"`
    
    // 詳細メトリクス
    CPUUsage         float64           `json:"cpu_usage"`
    GCCount          int               `json:"gc_count"`
    GCPauseTime      time.Duration     `json:"gc_pause_time"`
}

// CompatibilityResult は互換性テスト結果
type CompatibilityResult struct {
    TestName         string            `json:"test_name"`
    APIVersion       string            `json:"api_version"`
    CLICompatible    bool              `json:"cli_compatible"`
    ConfigCompatible bool              `json:"config_compatible"`
    OutputCompatible bool              `json:"output_compatible"`
    
    // 互換性問題詳細
    BreakingChanges  []BreakingChange  `json:"breaking_changes"`
    Warnings         []CompatibilityWarning `json:"warnings"`
}

// RegressionTolerance は回帰許容値
type RegressionTolerance struct {
    // 機能回帰許容値
    FunctionalErrorRate    float64 `json:"functional_error_rate"`    // 0.01 = 1%
    
    // パフォーマンス回帰許容値
    PerformanceRegression  float64 `json:"performance_regression"`   // 0.05 = 5%
    MemoryRegression       float64 `json:"memory_regression"`        // 0.10 = 10%
    
    // 互換性回帰許容値
    BreakingChangeCount    int     `json:"breaking_change_count"`    // 0
    WarningThreshold       int     `json:"warning_threshold"`        // 5
}

// NewRegressionTestSuite は新しい回帰テストスイートを作成
func NewRegressionTestSuite(t *testing.T) *RegressionTestSuite {
    suite := &RegressionTestSuite{
        t:           t,
        testDataDir: "testdata/regression_scenarios",
        baselineDir: "tests/regression/baseline",
        reportsDir:  "tests/regression/reports",
        
        tolerance: &RegressionTolerance{
            FunctionalErrorRate:   0.01, // 1%
            PerformanceRegression: 0.05, // 5%
            MemoryRegression:      0.10, // 10%
            BreakingChangeCount:   0,
            WarningThreshold:      5,
        },
    }
    
    suite.loadBaselines()
    suite.setupTestEnvironment()
    return suite
}

// RunRegressionTest は回帰テストを実行
func (rts *RegressionTestSuite) RunRegressionTest(
    testType RegressionTestType,
    testName string,
) *RegressionTestResult {
    rts.t.Helper()
    
    switch testType {
    case TypeFunctional:
        return rts.runFunctionalRegressionTest(testName)
    case TypePerformance:
        return rts.runPerformanceRegressionTest(testName)
    case TypeCompatibility:
        return rts.runCompatibilityRegressionTest(testName)
    default:
        rts.t.Fatalf("不明な回帰テストタイプ: %v", testType)
        return nil
    }
}
```

### 機能回帰テスト

#### 1. 変換機能回帰テスト
```go
// tests/regression/functional/transform_regression_test.go
package functional

import (
    "crypto/sha256"
    "fmt"
    "io/ioutil"
    "testing"
)

// TestTransformRegression_CoreRules はコア変換ルールの回帰テスト
func TestTransformRegression_CoreRules(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    // 既存の変換ルールが正確に動作することを確認
    regressionCases := []struct {
        name        string
        inputFile   string
        ruleApplied string
    }{
        {
            name:        "CSV_to_JSON_conversion",
            inputFile:   "csv_output_scripts.sh", 
            ruleApplied: "output_format_csv_to_json",
        },
        {
            name:        "ISO_image_to_CDROM_rename",
            inputFile:   "iso_image_scripts.sh",
            ruleApplied: "resource_rename_iso_image",
        },
        {
            name:        "Startup_script_to_Note_rename", 
            inputFile:   "startup_script_scripts.sh",
            ruleApplied: "resource_rename_startup_script",
        },
        {
            name:        "Selector_deprecation",
            inputFile:   "selector_scripts.sh",
            ruleApplied: "selector_migration",
        },
    }
    
    for _, rc := range regressionCases {
        t.Run(rc.name, func(t *testing.T) {
            // 現在の実行結果
            currentResult := suite.ExecuteTransform(rc.inputFile)
            
            // ベースラインとの比較
            baselineResult := suite.GetFunctionalBaseline(rc.name)
            if baselineResult == nil {
                suite.CreateNewBaseline(rc.name, currentResult)
                return
            }
            
            // 出力ハッシュの比較
            if currentResult.OutputHash != baselineResult.OutputHash {
                // 詳細差分分析
                diff := suite.AnalyzeTransformDiff(baselineResult, currentResult)
                
                if diff.IsSignificant() {
                    t.Errorf("変換結果に回帰が検出されました:\n%s", diff.Report())
                } else {
                    t.Logf("軽微な差分が検出されました（許容範囲内）:\n%s", diff.Report())
                }
            }
            
            // 性能回帰チェック
            performanceRegression := suite.CheckPerformanceRegression(
                baselineResult.ExecutionTime,
                currentResult.ExecutionTime,
            )
            
            if performanceRegression > suite.tolerance.PerformanceRegression {
                t.Errorf("パフォーマンス回帰が検出されました: %.2f%% 悪化", 
                    performanceRegression*100)
            }
        })
    }
}

// TestTransformRegression_EdgeCases はエッジケースの回帰テスト
func TestTransformRegression_EdgeCases(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    edgeCases := []struct {
        name        string
        description string
        inputData   string
    }{
        {
            name:        "EmptyLines",
            description: "空行とコメントのみのファイル",
            inputData: `#!/bin/bash

# これはコメントです

# もう一つのコメント
`,
        },
        {
            name:        "VeryLongLine",
            description: "非常に長い行の処理",
            inputData:   fmt.Sprintf("usacloud server list %s", strings.Repeat("--tag key=value ", 100)),
        },
        {
            name:        "UnicodeCharacters",
            description: "Unicode文字を含む行",
            inputData:   "# 日本語コメント\nusacloud server create --name サーバー名",
        },
        {
            name:        "MixedLineEndings",
            description: "混在する改行コード",
            inputData:   "usacloud server list\r\nusacloud disk list\n",
        },
    }
    
    for _, ec := range edgeCases {
        t.Run(ec.name, func(t *testing.T) {
            currentResult := suite.ExecuteTransformOnData(ec.inputData)
            baselineResult := suite.GetFunctionalBaseline(ec.name)
            
            if baselineResult == nil {
                // 初回実行時はベースライン作成
                suite.CreateNewBaseline(ec.name, currentResult)
                return
            }
            
            // エッジケースでの一貫性確認
            suite.ValidateEdgeCaseConsistency(baselineResult, currentResult)
        })
    }
}
```

#### 2. 検証機能回帰テスト
```go
// tests/regression/functional/validation_regression_test.go
package functional

import (
    "testing"
)

// TestValidationRegression_ErrorDetection はエラー検出の回帰テスト
func TestValidationRegression_ErrorDetection(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    // 様々なエラーパターンが一貫して検出されることを確認
    validationCases := []struct {
        name           string
        inputCommand   string
        expectedErrors int
        errorTypes     []string
    }{
        {
            name:           "InvalidMainCommand",
            inputCommand:   "usacloud invalid-command list",
            expectedErrors: 1,
            errorTypes:     []string{"invalid_main_command"},
        },
        {
            name:           "InvalidSubCommand", 
            inputCommand:   "usacloud server invalid-action",
            expectedErrors: 1,
            errorTypes:     []string{"invalid_sub_command"},
        },
        {
            name:           "DeprecatedCommand",
            inputCommand:   "usacloud iso-image list",
            expectedErrors: 1,
            errorTypes:     []string{"deprecated_command"},
        },
        {
            name:           "MultipleErrors",
            inputCommand:   "usacloud invalid-cmd invalid-sub --deprecated-flag",
            expectedErrors: 3,
            errorTypes:     []string{"invalid_main_command", "invalid_sub_command", "deprecated_flag"},
        },
    }
    
    for _, vc := range validationCases {
        t.Run(vc.name, func(t *testing.T) {
            currentResult := suite.ExecuteValidation(vc.inputCommand)
            baselineResult := suite.GetFunctionalBaseline(vc.name)
            
            if baselineResult == nil {
                suite.CreateNewBaseline(vc.name, currentResult)
                return
            }
            
            // エラー検出数の一貫性確認
            if currentResult.ValidationIssues != baselineResult.ValidationIssues {
                t.Errorf("検証エラー数に回帰: 期待 %d, 実際 %d", 
                    baselineResult.ValidationIssues, currentResult.ValidationIssues)
            }
            
            // エラータイプの一貫性確認
            suite.ValidateErrorTypeConsistency(baselineResult, currentResult)
        })
    }
}

// TestValidationRegression_SuggestionQuality は提案品質の回帰テスト
func TestValidationRegression_SuggestionQuality(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    suggestionCases := []struct {
        name            string
        typoCommand     string
        expectedSuggestion string
        minConfidence   float64
    }{
        {
            name:            "ServerTypo",
            typoCommand:     "serv",
            expectedSuggestion: "server",
            minConfidence:   0.8,
        },
        {
            name:            "DiskTypo", 
            typoCommand:     "dsk",
            expectedSuggestion: "disk",
            minConfidence:   0.7,
        },
        {
            name:            "DatabaseTypo",
            typoCommand:     "databse",
            expectedSuggestion: "database", 
            minConfidence:   0.9,
        },
    }
    
    for _, sc := range suggestionCases {
        t.Run(sc.name, func(t *testing.T) {
            currentResult := suite.ExecuteSuggestion(sc.typoCommand)
            baselineResult := suite.GetFunctionalBaseline(sc.name)
            
            if baselineResult == nil {
                suite.CreateNewBaseline(sc.name, currentResult)
                return
            }
            
            // 提案品質の維持確認
            suite.ValidateSuggestionQuality(baselineResult, currentResult, sc.minConfidence)
        })
    }
}
```

### パフォーマンス回帰テスト

#### 1. 実行速度回帰テスト
```go
// tests/regression/performance/speed_regression_test.go
package performance

import (
    "testing"
    "time"
)

// TestSpeedRegression_ProcessingRate は処理速度の回帰テスト
func TestSpeedRegression_ProcessingRate(t *testing.T) {
    if testing.Short() {
        t.Skip("パフォーマンス回帰テストは短時間モードではスキップ")
    }
    
    suite := NewRegressionTestSuite(t)
    
    speedTestCases := []struct {
        name         string
        fileSize     int    // 行数
        maxDuration  time.Duration
    }{
        {
            name:        "SmallFile_100_lines",
            fileSize:    100,
            maxDuration: 1 * time.Second,
        },
        {
            name:        "MediumFile_1000_lines", 
            fileSize:    1000,
            maxDuration: 5 * time.Second,
        },
        {
            name:        "LargeFile_10000_lines",
            fileSize:    10000,
            maxDuration: 30 * time.Second,
        },
    }
    
    for _, stc := range speedTestCases {
        t.Run(stc.name, func(t *testing.T) {
            // テストファイル生成
            inputFile := suite.GenerateTestFile(stc.fileSize)
            
            // 現在の実行時間測定
            startTime := time.Now()
            currentResult := suite.ExecuteFullPipeline(inputFile)
            currentDuration := time.Since(startTime)
            
            // ベースライン取得
            baselineResult := suite.GetPerformanceBaseline(stc.name)
            
            if baselineResult == nil {
                // 初回実行時はベースライン作成
                newBaseline := &PerformanceResult{
                    TestName:      stc.name,
                    ExecutionTime: currentDuration,
                    ProcessingRate: float64(stc.fileSize) / currentDuration.Seconds(),
                }
                suite.CreatePerformanceBaseline(stc.name, newBaseline)
                return
            }
            
            // 処理速度回帰チェック
            performanceChange := float64(currentDuration-baselineResult.ExecutionTime) / 
                               float64(baselineResult.ExecutionTime)
            
            if performanceChange > suite.tolerance.PerformanceRegression {
                t.Errorf("処理速度回帰が検出されました: %.2f%% 悪化 (%v → %v)", 
                    performanceChange*100, 
                    baselineResult.ExecutionTime, 
                    currentDuration)
            }
            
            // 絶対的な制限チェック
            if currentDuration > stc.maxDuration {
                t.Errorf("処理時間が制限を超過: %v > %v", currentDuration, stc.maxDuration)
            }
            
            // 処理レート確認
            currentRate := float64(stc.fileSize) / currentDuration.Seconds()
            baselineRate := baselineResult.ProcessingRate
            
            rateChange := (baselineRate - currentRate) / baselineRate
            if rateChange > suite.tolerance.PerformanceRegression {
                t.Errorf("処理レート回帰が検出されました: %.2f lines/sec → %.2f lines/sec", 
                    baselineRate, currentRate)
            }
        })
    }
}

// TestSpeedRegression_ParallelProcessing は並列処理速度の回帰テスト
func TestSpeedRegression_ParallelProcessing(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    parallelTestCases := []struct {
        name        string
        workerCount int
        fileSize    int
        minSpeedup  float64
    }{
        {
            name:        "Parallel_2_workers",
            workerCount: 2,
            fileSize:    2000,
            minSpeedup:  1.3, // 30%の高速化を期待
        },
        {
            name:        "Parallel_4_workers",
            workerCount: 4,
            fileSize:    4000, 
            minSpeedup:  2.0, // 100%の高速化を期待
        },
        {
            name:        "Parallel_8_workers",
            workerCount: 8,
            fileSize:    8000,
            minSpeedup:  3.0, // 300%の高速化を期待
        },
    }
    
    for _, ptc := range parallelTestCases {
        t.Run(ptc.name, func(t *testing.T) {
            inputFile := suite.GenerateTestFile(ptc.fileSize)
            
            // シーケンシャル実行
            seqTime := suite.ExecuteSequential(inputFile)
            
            // 並列実行
            parTime := suite.ExecuteParallel(inputFile, ptc.workerCount)
            
            // 高速化率計算
            speedup := float64(seqTime) / float64(parTime)
            
            // ベースラインとの比較
            baselineResult := suite.GetPerformanceBaseline(ptc.name)
            if baselineResult != nil {
                baselineSpeedup := baselineResult.Throughput
                speedupRegression := (baselineSpeedup - speedup) / baselineSpeedup
                
                if speedupRegression > suite.tolerance.PerformanceRegression {
                    t.Errorf("並列処理高速化に回帰: %.2fx → %.2fx", baselineSpeedup, speedup)
                }
            }
            
            // 最小高速化率チェック
            if speedup < ptc.minSpeedup {
                t.Errorf("並列処理高速化が不十分: %.2fx < %.2fx", speedup, ptc.minSpeedup)
            }
        })
    }
}
```

### 互換性回帰テスト

#### 1. CLI互換性回帰テスト
```go
// tests/regression/compatibility/cli_compatibility_test.go
package compatibility

import (
    "testing"
)

// TestCLICompatibility_FlagBehavior はフラグ動作の互換性回帰テスト
func TestCLICompatibility_FlagBehavior(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    // 重要なCLIフラグの動作が維持されていることを確認
    compatibilityTests := []struct {
        name           string
        cliArgs        []string
        expectedOutput []string
        mustWork       bool
    }{
        {
            name:           "HelpFlag",
            cliArgs:        []string{"--help"},
            expectedOutput: []string{"usacloud-update", "使用法", "オプション"},
            mustWork:       true,
        },
        {
            name:           "VersionFlag", 
            cliArgs:        []string{"--version"},
            expectedOutput: []string{"usacloud-update", "version"},
            mustWork:       true,
        },
        {
            name:           "StatsFlag",
            cliArgs:        []string{"--stats", "testfile.sh"},
            expectedOutput: []string{"統計情報", "処理済み行数"},
            mustWork:       true,
        },
        {
            name:           "InputOutputFlags",
            cliArgs:        []string{"--in", "input.sh", "--out", "output.sh"},
            expectedOutput: []string{}, // ファイル作成のみ
            mustWork:       true,
        },
    }
    
    for _, ct := range compatibilityTests {
        t.Run(ct.name, func(t *testing.T) {
            currentResult := suite.ExecuteCLI(ct.cliArgs)
            baselineResult := suite.GetCompatibilityBaseline(ct.name)
            
            if baselineResult == nil {
                suite.CreateCompatibilityBaseline(ct.name, currentResult)
                return
            }
            
            // CLI動作の一貫性確認
            if ct.mustWork && !currentResult.Success {
                t.Errorf("重要なCLI機能が動作しません: %s", ct.name)
            }
            
            // 出力の一貫性確認（重要部分のみ）
            suite.ValidateOutputCompatibility(baselineResult, currentResult, ct.expectedOutput)
            
            // 終了コードの一貫性確認
            if currentResult.ExitCode != baselineResult.ExitCode {
                t.Errorf("終了コードの互換性問題: %d → %d", 
                    baselineResult.ExitCode, currentResult.ExitCode)
            }
        })
    }
}

// TestCLICompatibility_BreakingChanges は破壊的変更の検出テスト
func TestCLICompatibility_BreakingChanges(t *testing.T) {
    suite := NewRegressionTestSuite(t)
    
    // 破壊的変更が発生していないことを確認
    breakingChangeTests := []struct {
        name        string
        testCommand []string
        changeType  string
    }{
        {
            name:        "RemovedFlag",
            testCommand: []string{"--old-flag"},
            changeType:  "flag_removal",
        },
        {
            name:        "ChangedDefaultBehavior",
            testCommand: []string{"input.sh"},
            changeType:  "default_behavior_change",
        },
        {
            name:        "OutputFormatChange",
            testCommand: []string{"--stats", "input.sh"},
            changeType:  "output_format_change",
        },
    }
    
    for _, bct := range breakingChangeTests {
        t.Run(bct.name, func(t *testing.T) {
            currentResult := suite.ExecuteCLI(bct.testCommand)
            baselineResult := suite.GetCompatibilityBaseline(bct.name)
            
            if baselineResult == nil {
                suite.CreateCompatibilityBaseline(bct.name, currentResult)
                return
            }
            
            // 破壊的変更の検出
            breakingChanges := suite.DetectBreakingChanges(baselineResult, currentResult)
            
            if len(breakingChanges) > suite.tolerance.BreakingChangeCount {
                t.Errorf("破壊的変更が検出されました (%s):", bct.changeType)
                for _, change := range breakingChanges {
                    t.Errorf("  - %s", change.Description)
                }
            }
        })
    }
}
```

### 回帰レポートシステム

#### 1. 自動レポート生成
```go
// tests/regression/regression_reporter.go
package regression

import (
    "fmt"
    "html/template"
    "os"
    "path/filepath"
    "time"
)

// RegressionReporter は回帰テストレポーター
type RegressionReporter struct {
    outputDir     string
    templateDir   string
    
    // レポートデータ
    testResults   []*RegressionTestResult
    summary       *RegressionSummary
    
    // レポート設定
    includeGraphs bool
    includeDetails bool
}

// RegressionSummary は回帰テスト概要
type RegressionSummary struct {
    TotalTests       int               `json:"total_tests"`
    PassedTests      int               `json:"passed_tests"`
    FailedTests      int               `json:"failed_tests"`
    RegressionCount  int               `json:"regression_count"`
    
    // 分類別統計
    FunctionalRegressions  int         `json:"functional_regressions"`
    PerformanceRegressions int         `json:"performance_regressions"`
    CompatibilityRegressions int       `json:"compatibility_regressions"`
    
    // 全体評価
    OverallHealth    string            `json:"overall_health"` // "良好", "注意", "警告", "危険"
    RiskLevel        int               `json:"risk_level"`     // 1-5
    
    // 時系列データ
    TrendData        []TrendPoint      `json:"trend_data"`
    
    // 推奨アクション
    RecommendedActions []string        `json:"recommended_actions"`
}

// TrendPoint はトレンドデータポイント
type TrendPoint struct {
    Date             time.Time         `json:"date"`
    RegressionCount  int               `json:"regression_count"`
    OverallHealth    string            `json:"overall_health"`
}

// GenerateReport は包括的な回帰レポートを生成
func (rr *RegressionReporter) GenerateReport() error {
    // 1. HTML概要レポート生成
    if err := rr.generateHTMLSummaryReport(); err != nil {
        return fmt.Errorf("HTML概要レポート生成エラー: %w", err)
    }
    
    // 2. 詳細JSONレポート生成
    if err := rr.generateJSONDetailedReport(); err != nil {
        return fmt.Errorf("JSON詳細レポート生成エラー: %w", err)
    }
    
    // 3. マークダウンサマリー生成（GitHub用）
    if err := rr.generateMarkdownSummary(); err != nil {
        return fmt.Errorf("マークダウンサマリー生成エラー: %w", err)
    }
    
    // 4. CI/CDメトリクス出力
    if err := rr.generateCIMetrics(); err != nil {
        return fmt.Errorf("CI/CDメトリクス生成エラー: %w", err)
    }
    
    return nil
}

// generateMarkdownSummary はGitHub PR用のマークダウンサマリーを生成
func (rr *RegressionReporter) generateMarkdownSummary() error {
    summary := rr.summary
    
    markdownContent := fmt.Sprintf(`# 🔍 回帰テスト結果レポート

## 📊 概要

- **総テスト数**: %d
- **成功**: %d ✅
- **失敗**: %d ❌  
- **回帰検出**: %d 🚨

## 🎯 全体評価: %s

### 分類別回帰状況
- **機能回帰**: %d個
- **パフォーマンス回帰**: %d個  
- **互換性回帰**: %d個

### 💡 推奨アクション
`,
        summary.TotalTests,
        summary.PassedTests, 
        summary.FailedTests,
        summary.RegressionCount,
        summary.OverallHealth,
        summary.FunctionalRegressions,
        summary.PerformanceRegressions,
        summary.CompatibilityRegressions)
    
    for i, action := range summary.RecommendedActions {
        markdownContent += fmt.Sprintf("%d. %s\n", i+1, action)
    }
    
    // 詳細リンク追加
    markdownContent += fmt.Sprintf(`
### 📈 詳細レポート

- [詳細HTMLレポート](./reports/%s_detailed.html)
- [JSONデータ](./reports/%s_data.json)
- [トレンドグラフ](./reports/%s_trends.png)
`,
        time.Now().Format("2006-01-02"),
        time.Now().Format("2006-01-02"), 
        time.Now().Format("2006-01-02"))
    
    // ファイル書き込み
    outputPath := filepath.Join(rr.outputDir, "regression_summary.md")
    return os.WriteFile(outputPath, []byte(markdownContent), 0644)
}
```

### CI/CD統合

#### 1. 自動回帰テストワークフロー
```yaml
# .github/workflows/regression-tests.yml
name: Regression Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    # 毎日午前2時に実行
    - cron: '0 2 * * *'

jobs:
  regression-tests:
    runs-on: ubuntu-latest
    timeout-minutes: 60
    
    strategy:
      matrix:
        test-type: [functional, performance, compatibility]
    
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0  # 履歴が必要
    
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Download regression baselines
      run: |
        # 前回のベースラインをダウンロード
        aws s3 sync s3://usacloud-update-baselines/latest ./tests/regression/baseline/
      env:
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    
    - name: Run regression tests
      run: |
        go test -tags=regression ./tests/regression/${{ matrix.test-type }}/... \
          -timeout=45m \
          -v \
          -run=TestRegression \
          -args -update-baseline=false
    
    - name: Generate regression report
      if: always()
      run: |
        go run ./tests/regression/cmd/report-generator \
          --type=${{ matrix.test-type }} \
          --output=./reports/regression_${{ matrix.test-type }}.json
    
    - name: Upload regression reports
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: regression-reports-${{ matrix.test-type }}
        path: ./reports/
    
    - name: Comment PR with results
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const reportPath = `./reports/regression_${{ matrix.test-type }}.json`;
          if (fs.existsSync(reportPath)) {
            const report = JSON.parse(fs.readFileSync(reportPath, 'utf8'));
            let comment = `## 🔍 ${{ matrix.test-type }} 回帰テスト結果\n\n`;
            
            if (report.regression_count > 0) {
              comment += `❌ **${report.regression_count}個の回帰が検出されました**\n\n`;
              comment += `詳細: [回帰レポート](${process.env.GITHUB_SERVER_URL}/${process.env.GITHUB_REPOSITORY}/actions/runs/${process.env.GITHUB_RUN_ID})\n`;
            } else {
              comment += `✅ 回帰は検出されませんでした\n`;
            }
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
          }

  consolidate-results:
    needs: [regression-tests]
    runs-on: ubuntu-latest
    if: always()
    
    steps:
    - name: Download all reports
      uses: actions/download-artifact@v3
      with:
        path: ./reports/
    
    - name: Generate consolidated report
      run: |
        # 全ての回帰テスト結果を統合
        python3 scripts/consolidate_regression_reports.py \
          --input-dir ./reports/ \
          --output ./consolidated_regression_report.md
    
    - name: Upload consolidated report
      uses: actions/upload-artifact@v3
      with:
        name: consolidated-regression-report
        path: ./consolidated_regression_report.md
```

## テスト戦略
- **継続監視**: 毎日の自動実行による継続的な品質監視
- **早期発見**: 小さな変更でも回帰を早期に検出
- **分類管理**: 機能・性能・互換性の分類別回帰管理
- **トレンド分析**: 時系列での品質トレンド分析
- **自動化**: 人間の介入なしでの完全自動実行

## 依存関係
- 前提PBI: PBI-018～022 (全テスト戦略), PBI-001～017 (全実装)
- 外部ツール: Git (履歴管理), CI/CDシステム

## 見積もり
- 開発工数: 20時間
  - 回帰テストフレームワーク実装: 6時間
  - 機能回帰テスト実装: 4時間
  - パフォーマンス回帰テスト実装: 4時間
  - 互換性回帰テスト実装: 3時間
  - 自動レポートシステム実装: 2時間
  - CI/CD統合: 1時間

## 完了の定義
- [ ] 回帰テストフレームワークが実装されている
- [ ] 機能回帰テストが網羅的に実装されている
- [ ] パフォーマンス回帰テストが実装されている
- [ ] 互換性回帰テストが実装されている
- [ ] ベースライン管理システムが実装されている
- [ ] 自動レポート生成システムが実装されている
- [ ] CI/CDでの自動実行が安定している
- [ ] 回帰検出時の通知システムが実装されている
- [ ] トレンド分析とダッシュボードが実装されている
- [ ] 全回帰テストが継続的に実行されている
- [ ] コードレビューが完了している

## 実装状況
❌ **PBI-023は未実装** (2025-09-11)

**現在の状況**:
- 包括的な自動回帰テスト戦略とアーキテクチャが設計済み
- 機能、パフォーマンス、互換性回帰テストの詳細設計完了
- ベースライン管理、自動レポート、トレンド分析システムの仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `tests/regression/` - 回帰テストフレームワークとテストスイート
- 機能、パフォーマンス、互換性回帰テストの実装
- ベースライン管理システムと自動更新機能
- 自動レポート生成と通知システム
- トレンド分析とダッシュボード機能
- CI/CDパイプラインでの継続的実行環境

**次のステップ**:
1. 回帰テストフレームワークの基盤実装
2. 機能回帰テストとベースライン管理の実装
3. パフォーマンスと互換性回帰テストの実装
4. 自動レポートと通知システムの実装
5. トレンド分析とCI/CD統合の実装

## 備考
- 回帰テストは長期間の品質維持に重要で、継続的なメンテナンスが必要
- ベースラインの適切な更新タイミングの判断が品質維持の鍵
- 回帰検出時の迅速な対応プロセスの確立が重要
- パフォーマンス回帰は環境依存性があるため、安定した測定環境の確保が重要

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 新規自動回帰テストよりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存回帰テストの安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存の回帰テストは引き続き保守・改善
- 新規自動回帰テストシステムの実装は延期
- 現在のテスト基盤の安定化を最優先