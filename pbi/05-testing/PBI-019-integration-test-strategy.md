# PBI-019: 統合テスト戦略

## 概要
コマンド検証・エラーフィードバックシステム全体の統合動作を検証するため、コンポーネント間の相互作用、データフロー、エンドツーエンドシナリオを包括的にテストする統合テストスイートを設計・実装する。実際のユーザーシナリオに基づいた現実的なテスト環境を構築する。

## 受け入れ条件
- [ ] 全主要コンポーネント間の統合が正しくテストされている
- [ ] 実際のユーザーシナリオベースのテストが実装されている
- [ ] 異常系・エラーケースの統合処理が適切にテストされている
- [ ] パフォーマンス要件を満たす統合処理が確認されている
- [ ] CI/CD環境での自動実行が可能である

## 技術仕様

### 統合テストアーキテクチャ

#### 1. テスト環境構成
```
tests/
├── integration/
│   ├── end_to_end_test.go                    # エンドツーエンドテスト
│   ├── component_integration_test.go         # コンポーネント統合テスト
│   ├── cli_integration_test.go               # CLI統合テスト
│   ├── config_integration_test.go            # 設定統合テスト
│   ├── performance_integration_test.go       # パフォーマンス統合テスト
│   └── testdata/
│       ├── scenarios/
│       │   ├── user_scenarios.yaml          # ユーザーシナリオ定義
│       │   ├── error_scenarios.yaml         # エラーシナリオ定義
│       │   └── edge_cases.yaml              # エッジケースシナリオ
│       ├── sample_files/
│       │   ├── complex_script.sh            # 複雑なスクリプトサンプル
│       │   ├── mixed_versions.sh            # 混在バージョンスクリプト
│       │   └── problematic_script.sh        # 問題のあるスクリプト
│       ├── configs/
│       │   ├── test_config.conf             # テスト用設定
│       │   ├── minimal_config.conf          # 最小設定
│       │   └── full_config.conf             # 全機能設定
│       └── expected/
│           ├── transformed_outputs/         # 期待変換結果
│           ├── error_messages/              # 期待エラーメッセージ
│           └── help_outputs/                # 期待ヘルプ出力
├── fixtures/                                # テストフィクスチャ
└── helpers/                                 # テストヘルパー
    ├── test_runner.go                       # テスト実行器
    ├── mock_setup.go                        # モック設定
    └── assertion_helpers.go                 # アサーションヘルパー
```

#### 2. 統合テストフレームワーク
```go
// tests/integration/integration_test_framework.go
package integration

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"
    "yaml"
    
    "github.com/armaniacs/usacloud-update/internal/config"
    "github.com/armaniacs/usacloud-update/internal/transform"
    "github.com/armaniacs/usacloud-update/internal/validation"
)

// IntegrationTestSuite は統合テストスイート
type IntegrationTestSuite struct {
    t              *testing.T
    tempDir        string
    binaryPath     string
    configPath     string
    testDataDir    string
    
    // テスト対象システム
    integratedCLI  *IntegratedCLI
    config         *config.IntegratedConfig
    
    // テスト実行環境
    timeout        time.Duration
    parallelism    int
    verbose        bool
}

// NewIntegrationTestSuite は新しい統合テストスイートを作成
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
    suite := &IntegrationTestSuite{
        t:           t,
        tempDir:     t.TempDir(),
        testDataDir: "testdata",
        timeout:     30 * time.Second,
        parallelism: 4,
        verbose:     testing.Verbose(),
    }
    
    suite.setupTestEnvironment()
    return suite
}

// setupTestEnvironment はテスト環境をセットアップ
func (its *IntegrationTestSuite) setupTestEnvironment() {
    its.t.Helper()
    
    // テストバイナリのビルド
    its.buildTestBinary()
    
    // テスト設定ファイルの作成
    its.createTestConfig()
    
    // 統合CLIの初期化
    its.initializeIntegratedCLI()
}

// buildTestBinary はテスト用バイナリをビルド
func (its *IntegrationTestSuite) buildTestBinary() {
    its.t.Helper()
    
    its.binaryPath = filepath.Join(its.tempDir, "usacloud-update-test")
    
    cmd := exec.Command("go", "build", "-o", its.binaryPath, 
        "../../cmd/usacloud-update")
    cmd.Dir = its.tempDir
    
    if output, err := cmd.CombinedOutput(); err != nil {
        its.t.Fatalf("テストバイナリビルドエラー: %v\n出力: %s", err, output)
    }
}

// TestScenario はテストシナリオ定義
type TestScenario struct {
    Name         string            `yaml:"name"`
    Description  string            `yaml:"description"`
    Input        ScenarioInput     `yaml:"input"`
    Expected     ScenarioExpected  `yaml:"expected"`
    Config       map[string]interface{} `yaml:"config"`
    Environment  map[string]string `yaml:"environment"`
    Timeout      string            `yaml:"timeout"`
    Tags         []string          `yaml:"tags"`
}

// ScenarioInput はシナリオ入力定義
type ScenarioInput struct {
    Type        string            `yaml:"type"`        // "file", "command", "stdin"
    Content     string            `yaml:"content"`     // ファイル内容またはコマンド
    FilePath    string            `yaml:"file_path"`   // 入力ファイルパス
    Arguments   []string          `yaml:"arguments"`   // CLI引数
    Environment map[string]string `yaml:"environment"` // 環境変数
}

// ScenarioExpected は期待結果定義
type ScenarioExpected struct {
    ExitCode      int               `yaml:"exit_code"`
    OutputContains []string         `yaml:"output_contains"`
    OutputNotContains []string      `yaml:"output_not_contains"`
    ErrorContains []string          `yaml:"error_contains"`
    FilesCreated  []string          `yaml:"files_created"`
    FilesModified []string          `yaml:"files_modified"`
    Metrics       ExpectedMetrics   `yaml:"metrics"`
}

// ExpectedMetrics は期待メトリクス
type ExpectedMetrics struct {
    ProcessedLines   int     `yaml:"processed_lines"`
    TransformedLines int     `yaml:"transformed_lines"`
    ErrorsFound      int     `yaml:"errors_found"`
    SuggestionsShown int     `yaml:"suggestions_shown"`
    ExecutionTimeMs  int     `yaml:"execution_time_ms"`
}
```

### 主要統合テストシナリオ

#### 1. エンドツーエンドテスト
```go
// tests/integration/end_to_end_test.go
package integration

import (
    "testing"
)

// TestEndToEnd_CompleteWorkflow は完全なワークフローのテスト
func TestEndToEnd_CompleteWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("短時間テストモードではスキップ")
    }
    
    suite := NewIntegrationTestSuite(t)
    defer suite.Cleanup()
    
    scenarios := []struct {
        name         string
        scenarioFile string
    }{
        {
            name:         "Beginner user with typos",
            scenarioFile: "scenarios/beginner_user_typos.yaml",
        },
        {
            name:         "Expert user batch processing",
            scenarioFile: "scenarios/expert_batch_processing.yaml",
        },
        {
            name:         "Mixed version script conversion", 
            scenarioFile: "scenarios/mixed_version_conversion.yaml",
        },
        {
            name:         "Interactive command building",
            scenarioFile: "scenarios/interactive_command_building.yaml",
        },
    }
    
    for _, sc := range scenarios {
        t.Run(sc.name, func(t *testing.T) {
            suite.RunScenarioFromFile(sc.scenarioFile)
        })
    }
}

// TestEndToEnd_ErrorHandling はエラーハンドリングのテスト
func TestEndToEnd_ErrorHandling(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    defer suite.Cleanup()
    
    // 複雑なエラーシナリオ
    errorScenarios := []TestScenario{
        {
            Name: "Multiple validation errors in single line",
            Input: ScenarioInput{
                Type:      "command",
                Arguments: []string{"--validate-only"},
                Content:   "usacloud invalid-cmd invalid-sub --deprecated-flag",
            },
            Expected: ScenarioExpected{
                ExitCode: 1,
                ErrorContains: []string{
                    "invalid-cmd",
                    "有効なusacloudコマンドではありません", 
                    "候補",
                },
            },
        },
        {
            Name: "Deprecated command with migration guide",
            Input: ScenarioInput{
                Type:      "command", 
                Arguments: []string{"--strict-validation"},
                Content:   "usacloud iso-image list",
            },
            Expected: ScenarioExpected{
                ExitCode: 1,
                ErrorContains: []string{
                    "iso-image",
                    "廃止",
                    "cdrom",
                    "移行方法",
                },
            },
        },
    }
    
    for _, scenario := range errorScenarios {
        t.Run(scenario.Name, func(t *testing.T) {
            suite.RunScenario(scenario)
        })
    }
}
```

#### 2. コンポーネント統合テスト
```go
// tests/integration/component_integration_test.go
package integration

import (
    "testing"
)

// TestComponentIntegration_ValidationToErrorFormatting は検証からエラーフォーマットまでの統合テスト
func TestComponentIntegration_ValidationToErrorFormatting(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    
    // 検証器 -> エラーメッセージ生成器 -> フォーマッターの連携テスト
    testCases := []struct {
        name          string
        inputCommand  string
        expectStages  []string
        expectResult  string
    }{
        {
            name:         "Invalid main command with suggestions",
            inputCommand: "usacloud serv list",
            expectStages: []string{
                "main_command_validation",
                "similar_command_suggestion", 
                "error_message_generation",
                "comprehensive_formatting",
            },
            expectResult: "server",
        },
        {
            name:         "Invalid subcommand with alternatives",
            inputCommand: "usacloud server lst",
            expectStages: []string{
                "main_command_validation",
                "subcommand_validation",
                "similar_subcommand_suggestion",
                "error_message_generation", 
                "comprehensive_formatting",
            },
            expectResult: "list",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := suite.ProcessCommandThroughPipeline(tc.inputCommand)
            
            // 各段階が実行されたことを確認
            for _, stage := range tc.expectStages {
                if !result.StageExecuted(stage) {
                    t.Errorf("期待される段階 '%s' が実行されませんでした", stage)
                }
            }
            
            // 最終結果の確認
            if !strings.Contains(result.FinalOutput, tc.expectResult) {
                t.Errorf("最終出力に期待する内容が含まれていません: %s", tc.expectResult)
            }
        })
    }
}

// TestComponentIntegration_ConfigProfileSystem は設定・プロファイルシステムの統合テスト
func TestComponentIntegration_ConfigProfileSystem(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    
    // プロファイル切り替えが全システムに反映されることをテスト
    profiles := []struct {
        name     string
        profile  string
        command  string
        expectBehavior string
    }{
        {
            name:     "Beginner profile enables verbose help",
            profile:  "beginner",
            command:  "invalid-command",
            expectBehavior: "verbose_help_shown",
        },
        {
            name:     "Expert profile shows minimal output",
            profile:  "expert", 
            command:  "invalid-command",
            expectBehavior: "minimal_output",
        },
        {
            name:     "CI profile disables color and interactive",
            profile:  "ci",
            command:  "invalid-command", 
            expectBehavior: "no_color_no_interactive",
        },
    }
    
    for _, profile := range profiles {
        t.Run(profile.name, func(t *testing.T) {
            // プロファイル設定
            suite.SetProfile(profile.profile)
            
            // コマンド実行
            result := suite.ExecuteCommand(profile.command)
            
            // 期待する動作の確認
            suite.AssertBehavior(result, profile.expectBehavior)
        })
    }
}
```

#### 3. パフォーマンス統合テスト
```go
// tests/integration/performance_integration_test.go
package integration

import (
    "testing"
    "time"
)

// TestPerformanceIntegration_LargeFiles は大きなファイルでのパフォーマンステスト
func TestPerformanceIntegration_LargeFiles(t *testing.T) {
    if testing.Short() {
        t.Skip("パフォーマンステストは短時間モードではスキップ")
    }
    
    suite := NewIntegrationTestSuite(t)
    
    // 異なるサイズのファイルでテスト
    testCases := []struct {
        name           string
        lines          int
        maxTimeSeconds int
        maxMemoryMB    int
    }{
        {
            name:           "Small file (100 lines)",
            lines:          100,
            maxTimeSeconds: 1,
            maxMemoryMB:    50,
        },
        {
            name:           "Medium file (1,000 lines)",
            lines:          1000,
            maxTimeSeconds: 3,
            maxMemoryMB:    100,
        },
        {
            name:           "Large file (10,000 lines)",
            lines:          10000,
            maxTimeSeconds: 10,
            maxMemoryMB:    200,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // テストファイル生成
            inputFile := suite.GenerateTestFile(tc.lines)
            
            // パフォーマンス測定実行
            result := suite.MeasurePerformance(func() {
                suite.ExecuteCommandWithFile(inputFile)
            })
            
            // パフォーマンス要件の確認
            if result.ExecutionTime > time.Duration(tc.maxTimeSeconds)*time.Second {
                t.Errorf("実行時間が制限を超過: %v > %ds", 
                    result.ExecutionTime, tc.maxTimeSeconds)
            }
            
            if result.MaxMemoryMB > tc.maxMemoryMB {
                t.Errorf("メモリ使用量が制限を超過: %dMB > %dMB", 
                    result.MaxMemoryMB, tc.maxMemoryMB)
            }
        })
    }
}

// TestPerformanceIntegration_ConcurrentProcessing は並行処理のパフォーマンステスト
func TestPerformanceIntegration_ConcurrentProcessing(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    
    // 並行処理の効果をテスト
    testFile := suite.GenerateTestFile(5000) // 5000行のファイル
    
    // シーケンシャル処理
    seqStart := time.Now()
    suite.ExecuteCommand("--parallel=false", testFile)
    seqTime := time.Since(seqStart)
    
    // 並行処理
    parStart := time.Now()
    suite.ExecuteCommand("--parallel=true", testFile)
    parTime := time.Since(parStart)
    
    // 並行処理が効果的であることを確認
    speedup := float64(seqTime) / float64(parTime)
    if speedup < 1.5 {
        t.Logf("並行処理による高速化が期待値以下: %.2fx", speedup)
        // エラーではなく警告として記録
    } else {
        t.Logf("並行処理による高速化: %.2fx", speedup)
    }
}
```

### テストシナリオ定義

#### 1. YAML形式のシナリオ定義
```yaml
# scenarios/beginner_user_typos.yaml
name: "初心者ユーザーのtypoシナリオ"
description: "初心者がよく犯すtypoに対する支援機能のテスト"

scenarios:
  - name: "Server command typo"
    input:
      type: "command"
      arguments: ["--profile", "beginner", "--interactive"]
      content: "usacloud serv list"
    expected:
      exit_code: 1
      error_contains:
        - "serv"
        - "有効なusacloudコマンドではありません"
        - "もしかして以下のコマンドですか"
        - "server"
      metrics:
        suggestions_shown: 1

  - name: "Subcommand typo with help"
    input:
      type: "command"
      arguments: ["--profile", "beginner"] 
      content: "usacloud server lst"
    expected:
      exit_code: 1
      error_contains:
        - "lst"
        - "有効なサブコマンドではありません"
        - "list"
        - "利用可能なサブコマンド"

  - name: "Deprecated command guidance"
    input:
      type: "command"
      arguments: ["--profile", "beginner"]
      content: "usacloud iso-image list" 
    expected:
      exit_code: 1
      error_contains:
        - "iso-image"
        - "廃止されました"
        - "cdrom"
        - "移行方法"
```

#### 2. エッジケースシナリオ
```yaml
# scenarios/edge_cases.yaml
name: "エッジケースシナリオ"
description: "境界値や特殊な状況でのシステム動作テスト"

scenarios:
  - name: "Empty input handling"
    input:
      type: "stdin"
      content: ""
    expected:
      exit_code: 0
      output_contains:
        - "処理対象の行がありません"

  - name: "Very long command line"
    input:
      type: "command"
      content: "usacloud server list --very-long-argument-name-that-exceeds-normal-limits-and-tests-buffer-handling"
    expected:
      exit_code: 1
      error_contains:
        - "very-long-argument-name"

  - name: "Unicode characters in command"
    input:
      type: "command"
      content: "usacloud ｓｅｒｖｅｒ ｌｉｓｔ"  # 全角文字
    expected:
      exit_code: 1
      error_contains:
        - "有効なusacloudコマンドではありません"
```

### テスト実行環境

#### 1. CI/CD統合
```yaml
# .github/workflows/integration-test.yml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21]
        test-suite: [end-to-end, component, performance]
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Build
      run: make build
    
    - name: Run Integration Tests
      run: |
        go test -tags=integration ./tests/integration/... \
          -run="Test.*${{ matrix.test-suite }}" \
          -timeout=10m \
          -v
    
    - name: Upload Test Results
      if: always()
      uses: actions/upload-artifact@v3
      with:
        name: integration-test-results-${{ matrix.test-suite }}
        path: tests/integration/results/
```

## テスト戦略
- **現実的シナリオ**: 実際のユーザー使用パターンに基づくテストケース
- **段階的実行**: コンポーネントの統合を段階的にテスト
- **エラー網羅**: 正常系だけでなく異常系の統合処理を重点テスト
- **パフォーマンス検証**: 統合システム全体のパフォーマンス要件確認
- **環境可搬性**: 異なる環境での一貫した動作確認

## 依存関係
- 前提PBI: PBI-018 (ユニットテスト), PBI-001～017 (全実装)
- 外部ツール: Docker (テスト環境分離), YAML (シナリオ定義)

## 見積もり
- 開発工数: 16時間
  - 統合テストフレームワーク実装: 4時間
  - エンドツーエンドテスト実装: 4時間
  - コンポーネント統合テスト実装: 3時間
  - パフォーマンステスト実装: 2時間
  - テストシナリオ作成: 2時間
  - CI/CD統合: 1時間

## 完了の定義
- [ ] 統合テストフレームワークが実装されている
- [ ] 主要ユーザーシナリオの統合テストが実装されている
- [ ] コンポーネント間統合テストが網羅的に実装されている
- [ ] パフォーマンス統合テストが実装されている
- [ ] エラーハンドリング統合テストが実装されている
- [ ] YAMLベースのシナリオ定義システムが実装されている
- [ ] CI/CDパイプラインでの自動実行が設定されている
- [ ] 全統合テストが安定して通過している
- [ ] テストドキュメントが作成されている
- [ ] コードレビューが完了している

## 実装状況
❌ **PBI-019は未実装** (2025-09-11)

**現在の状況**:
- 包括的な統合テスト戦略とアーキテクチャが設計済み
- YAMLベースのシナリオ定義システムの仕様が完成
- エンドツーエンド・コンポーネント統合・パフォーマンステストの詳細設計完了
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `tests/integration/` - 統合テストフレームワークとテストスイート
- `tests/integration/testdata/scenarios/` - YAMLベースのテストシナリオ定義
- `tests/helpers/` - テストヘルパーとアサーション機能
- CI/CDパイプラインとの統合設定
- ユーザーシナリオベースのテストケース実装

**次のステップ**:
1. 統合テストフレームワークの基盤実装
2. YAMLシナリオ定義システムの構築
3. エンドツーエンドテストの実装
4. コンポーネント統合テストの実装
5. CI/CD統合とテスト自動化の設定

## 備考
- 統合テストは実行時間が長いため、効率的な並列実行が重要
- 実際のユーザーフィードバックを基にしたシナリオの継続的改善が必要
- テスト環境の分離により、CI/CD環境での安定実行を確保
- 複雑な統合テストの保守性を考慮した設計が重要

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 新規統合テスト戦略よりも既存テストの修復が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存統合テストの安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存の統合テストは引き続き保守・改善
- 新規統合テスト戦略の実装は延期
- 現在のテスト基盤の安定化を最優先