# PBI-022: エンドツーエンドテストシナリオ

## 概要
実際のユーザーワークフローを模擬した包括的なエンドツーエンドテストシナリオを設計・実装する。CLI実行からファイル出力、エラーハンドリング、ヘルプシステムまで、完全なユーザー体験をテストし、現実的な使用場面での品質を保証する。

## 受け入れ条件
- [ ] 実際のユーザーワークフローを網羅的にカバーしている
- [ ] 複雑なエラーシナリオとその回復処理がテストされている
- [ ] 異なるユーザータイプ（初心者・中級者・上級者）に対応している
- [ ] ファイルI/O、設定管理、環境依存性が適切にテストされている
- [ ] CI/CD環境での安定した自動実行が実現されている

## 技術仕様

### エンドツーエンドテストアーキテクチャ

#### 1. テストシナリオ分類
```
tests/
├── e2e/
│   ├── user_workflows/                      # ユーザーワークフローテスト
│   │   ├── beginner_workflow_test.go       # 初心者ワークフロー
│   │   ├── expert_workflow_test.go         # エキスパートワークフロー
│   │   ├── ci_workflow_test.go             # CI/CD環境ワークフロー
│   │   └── migration_workflow_test.go      # 移行作業ワークフロー
│   ├── error_scenarios/                    # エラーシナリオテスト
│   │   ├── file_errors_test.go             # ファイル関連エラー
│   │   ├── validation_errors_test.go       # 検証エラー処理
│   │   ├── config_errors_test.go           # 設定エラー処理
│   │   └── recovery_scenarios_test.go      # エラー回復シナリオ
│   ├── integration_scenarios/              # 統合シナリオテスト
│   │   ├── profile_switching_test.go       # プロファイル切り替え
│   │   ├── interactive_mode_test.go        # インタラクティブモード
│   │   ├── batch_processing_test.go        # バッチ処理
│   │   └── help_system_test.go             # ヘルプシステム統合
│   └── testdata/
│       ├── workflows/
│       │   ├── beginner_scripts/           # 初心者向けスクリプト
│       │   ├── complex_scripts/            # 複雑なスクリプト
│       │   └── problematic_scripts/        # 問題のあるスクリプト
│       ├── environments/
│       │   ├── clean_env/                  # クリーン環境
│       │   ├── configured_env/             # 設定済み環境
│       │   └── broken_env/                 # 破損環境
│       └── expected_outputs/
│           ├── successful_runs/            # 成功時出力
│           ├── error_scenarios/            # エラー時出力
│           └── interactive_sessions/       # 対話セッション出力
```

#### 2. エンドツーエンドテストフレームワーク
```go
// tests/e2e/e2e_test_framework.go
package e2e

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

// E2ETestSuite はエンドツーエンドテストスイート
type E2ETestSuite struct {
    t           *testing.T
    testDir     string
    binaryPath  string
    
    // テスト環境
    tempHome    string
    tempConfig  string
    cleanupFuncs []func()
    
    // 実行設定
    timeout     time.Duration
    verbose     bool
}

// E2ETestOptions はE2Eテストオプション
type E2ETestOptions struct {
    // 実行設定
    Arguments      []string          `yaml:"arguments"`
    Environment    map[string]string `yaml:"environment"`
    WorkingDir     string            `yaml:"working_dir"`
    Timeout        string            `yaml:"timeout"`
    
    // 入力設定
    StdinInput     string            `yaml:"stdin_input"`
    InteractiveInputs []string       `yaml:"interactive_inputs"`
    InputFiles     []string          `yaml:"input_files"`
    
    // 期待結果
    ExpectedExitCode    int          `yaml:"expected_exit_code"`
    ExpectedStdout      []string     `yaml:"expected_stdout"`
    ExpectedStderr      []string     `yaml:"expected_stderr"`
    ExpectedFiles       []FileExpectation `yaml:"expected_files"`
    ExpectedNoFiles     []string     `yaml:"expected_no_files"`
    
    // 検証設定
    ValidateOutput      bool         `yaml:"validate_output"`
    ValidateFiles       bool         `yaml:"validate_files"`
    ValidatePerformance bool         `yaml:"validate_performance"`
    MaxExecutionTime    string       `yaml:"max_execution_time"`
    MaxMemoryUsage      string       `yaml:"max_memory_usage"`
}

// FileExpectation はファイル期待値
type FileExpectation struct {
    Path            string   `yaml:"path"`
    ShouldExist     bool     `yaml:"should_exist"`
    ContentContains []string `yaml:"content_contains"`
    ContentExact    string   `yaml:"content_exact"`
    MinSize         int64    `yaml:"min_size"`
    MaxSize         int64    `yaml:"max_size"`
}

// E2ETestResult はE2Eテスト結果
type E2ETestResult struct {
    ExitCode        int           `json:"exit_code"`
    Stdout          string        `json:"stdout"`
    Stderr          string        `json:"stderr"`
    ExecutionTime   time.Duration `json:"execution_time"`
    MemoryUsage     int64         `json:"memory_usage"`
    FilesCreated    []string      `json:"files_created"`
    FilesModified   []string      `json:"files_modified"`
    Error           error         `json:"error,omitempty"`
}

// NewE2ETestSuite は新しいE2Eテストスイートを作成
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
    suite := &E2ETestSuite{
        t:           t,
        testDir:     t.TempDir(),
        timeout:     5 * time.Minute,
        verbose:     testing.Verbose(),
    }
    
    suite.setupTestEnvironment()
    return suite
}

// setupTestEnvironment はテスト環境をセットアップ
func (e2e *E2ETestSuite) setupTestEnvironment() {
    e2e.t.Helper()
    
    // テストバイナリをビルド
    e2e.buildTestBinary()
    
    // 分離された環境をセットアップ
    e2e.setupIsolatedEnvironment()
    
    // クリーンアップ関数の登録
    e2e.t.Cleanup(e2e.cleanup)
}

// RunE2ETest はE2Eテストを実行
func (e2e *E2ETestSuite) RunE2ETest(testName string, options *E2ETestOptions) *E2ETestResult {
    e2e.t.Helper()
    
    // テスト前の状態スナップショット
    beforeSnapshot := e2e.takeEnvironmentSnapshot()
    
    // テスト実行
    result := e2e.executeTest(options)
    
    // テスト後の状態スナップショット
    afterSnapshot := e2e.takeEnvironmentSnapshot()
    
    // 環境変化を記録
    result.FilesCreated = e2e.findCreatedFiles(beforeSnapshot, afterSnapshot)
    result.FilesModified = e2e.findModifiedFiles(beforeSnapshot, afterSnapshot)
    
    // 期待結果との比較
    e2e.validateResult(testName, result, options)
    
    return result
}

// executeTest はテストを実行
func (e2e *E2ETestSuite) executeTest(options *E2ETestOptions) *E2ETestResult {
    // コマンド構築
    args := append([]string{}, options.Arguments...)
    cmd := exec.Command(e2e.binaryPath, args...)
    
    // 環境変数設定
    cmd.Env = e2e.buildEnvironment(options.Environment)
    
    // 作業ディレクトリ設定
    if options.WorkingDir != "" {
        cmd.Dir = filepath.Join(e2e.testDir, options.WorkingDir)
    } else {
        cmd.Dir = e2e.testDir
    }
    
    // 標準入力設定
    if options.StdinInput != "" {
        cmd.Stdin = strings.NewReader(options.StdinInput)
    }
    
    // 出力キャプチャ設定
    var stdout, stderr strings.Builder
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // インタラクティブ入力処理
    if len(options.InteractiveInputs) > 0 {
        return e2e.executeInteractiveTest(cmd, options, &stdout, &stderr)
    }
    
    // 実行時間測定
    startTime := time.Now()
    err := cmd.Run()
    executionTime := time.Since(startTime)
    
    // 結果構築
    result := &E2ETestResult{
        ExitCode:      cmd.ProcessState.ExitCode(),
        Stdout:        stdout.String(),
        Stderr:        stderr.String(),
        ExecutionTime: executionTime,
        Error:         err,
    }
    
    return result
}
```

### ユーザーワークフローテスト

#### 1. 初心者ユーザーワークフロー
```go
// tests/e2e/user_workflows/beginner_workflow_test.go
package user_workflows

import (
    "testing"
    "path/filepath"
)

// TestBeginnerWorkflow_FirstTimeUser は初回利用者のワークフローテスト
func TestBeginnerWorkflow_FirstTimeUser(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // シナリオ: 初心者がtypoを含むスクリプトを変換
    t.Run("TypoInScript", func(t *testing.T) {
        // 1. typoを含むスクリプト作成
        inputScript := `#!/bin/bash
usacloud serv list
usacloud iso-image lst
usacloud dsk create --size 100`
        
        inputFile := suite.CreateTempFile("input.sh", inputScript)
        outputFile := filepath.Join(suite.testDir, "output.sh")
        
        // 2. 初心者プロファイルで実行
        options := &E2ETestOptions{
            Arguments: []string{
                "--profile", "beginner",
                "--in", inputFile,
                "--out", outputFile,
                "--interactive",
            },
            InteractiveInputs: []string{
                "y", // "server" への修正を受け入れ
                "y", // "cdrom" への修正を受け入れ
                "y", // "list" への修正を受け入れ
                "y", // "disk" への修正を受け入れ
                "n", // ヘルプは不要
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "typoが検出されました",
                "修正候補",
                "server",
                "cdrom",
                "list",
                "disk",
                "変換完了",
            },
            ExpectedFiles: []FileExpectation{
                {
                    Path:        outputFile,
                    ShouldExist: true,
                    ContentContains: []string{
                        "usacloud server list",
                        "usacloud cdrom list",
                        "usacloud disk create",
                    },
                },
            },
        }
        
        result := suite.RunE2ETest("BeginnerTypoFix", options)
        
        // 追加検証: 学習効果の確認
        suite.ValidateHelpfulOutput(result)
        suite.ValidateLearningProgress(result)
    })
}

// TestBeginnerWorkflow_LearningProgress は学習進捗のワークフローテスト
func TestBeginnerWorkflow_LearningProgress(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // シナリオ: 初心者が段階的にスキルアップ
    learningSteps := []struct {
        name   string
        script string
        expectImprovement bool
    }{
        {
            name:   "Step1_BasicMistakes",
            script: "usacloud serv list",
            expectImprovement: false,
        },
        {
            name:   "Step2_SameMistake", 
            script: "usacloud serv read 123",
            expectImprovement: false,
        },
        {
            name:   "Step3_NewCommand",
            script: "usacloud server create --name test",
            expectImprovement: true,
        },
    }
    
    for i, step := range learningSteps {
        t.Run(step.name, func(t *testing.T) {
            inputFile := suite.CreateTempFile(fmt.Sprintf("step%d.sh", i+1), step.script)
            
            options := &E2ETestOptions{
                Arguments: []string{
                    "--profile", "beginner",
                    "--validate-only",
                    inputFile,
                },
                ExpectedExitCode: 1, // 検証エラーを期待
            }
            
            result := suite.RunE2ETest(step.name, options)
            
            if step.expectImprovement {
                // より詳細なアドバイスが減っているかチェック
                suite.ValidateReducedHelp(result)
            }
        })
    }
}
```

#### 2. エキスパートユーザーワークフロー
```go
// tests/e2e/user_workflows/expert_workflow_test.go  
package user_workflows

import (
    "testing"
)

// TestExpertWorkflow_BatchProcessing はバッチ処理のワークフローテスト
func TestExpertWorkflow_BatchProcessing(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // シナリオ: エキスパートが大量ファイルを高速処理
    t.Run("LargeBatchProcessing", func(t *testing.T) {
        // 複数のスクリプトファイル作成
        scriptFiles := suite.CreateMultipleScripts([]string{
            "script1.sh",
            "script2.sh", 
            "script3.sh",
        }, 100) // 各ファイル100行
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--profile", "expert",
                "--parallel", "true",
                "--batch-process",
                "--input-dir", suite.testDir,
                "--output-dir", filepath.Join(suite.testDir, "output"),
                "--stats",
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "バッチ処理開始",
                "並列処理: 有効",
                "処理完了",
                "統計情報",
            },
            ValidatePerformance: true,
            MaxExecutionTime:    "30s",
        }
        
        result := suite.RunE2ETest("ExpertBatchProcessing", options)
        
        // パフォーマンス検証
        suite.ValidateProcessingRate(result, 1000) // 1000 lines/sec
        
        // 出力ファイル検証
        suite.ValidateAllFilesProcessed(scriptFiles, result)
    })
}

// TestExpertWorkflow_CustomConfiguration はカスタム設定のワークフローテスト
func TestExpertWorkflow_CustomConfiguration(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // カスタム設定ファイル作成
    customConfig := `
[validation]
strict_mode = true
max_suggestions = 2
enable_typo_detection = false

[output]
show_progress = false
report_level = "minimal"

[performance]
parallel_processing = true
worker_count = 8
`
    
    configFile := suite.CreateTempFile("expert.conf", customConfig)
    
    t.Run("CustomConfigProcessing", func(t *testing.T) {
        inputScript := `usacloud server list --output-type csv
usacloud iso-image list`
        inputFile := suite.CreateTempFile("input.sh", inputScript)
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--config", configFile,
                "--strict-validation",
                inputFile,
            },
            ExpectedExitCode: 1, // 厳格モードでエラー
            ExpectedStderr: []string{
                "厳格モード",
                "iso-image",
                "廃止",
            },
        }
        
        result := suite.RunE2ETest("ExpertCustomConfig", options)
        
        // 最小出力の確認
        suite.ValidateMinimalOutput(result)
    })
}
```

### エラーシナリオテスト

#### 1. ファイル関連エラーテスト
```go
// tests/e2e/error_scenarios/file_errors_test.go
package error_scenarios

import (
    "os"
    "testing"
)

// TestFileErrors_InputFileIssues は入力ファイル問題のテスト
func TestFileErrors_InputFileIssues(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    errorScenarios := []struct {
        name        string
        setupFunc   func() string
        expectedMsg string
    }{
        {
            name: "NonExistentFile",
            setupFunc: func() string {
                return "/non/existent/file.sh"
            },
            expectedMsg: "ファイルが見つかりません",
        },
        {
            name: "NoReadPermission",
            setupFunc: func() string {
                file := suite.CreateTempFile("noaccess.sh", "usacloud server list")
                os.Chmod(file, 0000) // 読み取り権限削除
                return file
            },
            expectedMsg: "ファイルの読み取り権限がありません",
        },
        {
            name: "EmptyFile",
            setupFunc: func() string {
                return suite.CreateTempFile("empty.sh", "")
            },
            expectedMsg: "処理対象の行がありません",
        },
        {
            name: "BinaryFile",
            setupFunc: func() string {
                binaryData := []byte{0x00, 0x01, 0x02, 0xFF}
                return suite.CreateTempFileBytes("binary.bin", binaryData)
            },
            expectedMsg: "バイナリファイルは処理できません",
        },
    }
    
    for _, scenario := range errorScenarios {
        t.Run(scenario.name, func(t *testing.T) {
            inputFile := scenario.setupFunc()
            
            options := &E2ETestOptions{
                Arguments: []string{
                    "--in", inputFile,
                    "--out", "/tmp/output.sh",
                },
                ExpectedExitCode: 1,
                ExpectedStderr: []string{
                    scenario.expectedMsg,
                },
            }
            
            result := suite.RunE2ETest(scenario.name, options)
            suite.ValidateGracefulErrorHandling(result)
        })
    }
}

// TestFileErrors_OutputFileIssues は出力ファイル問題のテスト
func TestFileErrors_OutputFileIssues(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    inputScript := "usacloud server list"
    inputFile := suite.CreateTempFile("input.sh", inputScript)
    
    t.Run("NoWritePermission", func(t *testing.T) {
        // 書き込み権限のないディレクトリ
        noWriteDir := suite.CreateTempDir("nowrite")
        os.Chmod(noWriteDir, 0555) // 読み取り・実行のみ
        
        outputFile := filepath.Join(noWriteDir, "output.sh")
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--in", inputFile,
                "--out", outputFile,
            },
            ExpectedExitCode: 1,
            ExpectedStderr: []string{
                "出力ファイルの作成に失敗しました",
                "書き込み権限",
            },
        }
        
        result := suite.RunE2ETest("NoWritePermission", options)
        suite.ValidateErrorRecoveryAdvice(result)
    })
    
    t.Run("DiskSpaceFull", func(t *testing.T) {
        // ディスク容量不足をシミュレート（モック）
        if testing.Short() {
            t.Skip("ディスク容量テストは短時間モードではスキップ")
        }
        
        // 大きなファイル作成でディスク不足をシミュレート
        // 実装は環境依存のため、ここではスキップ
        t.Skip("ディスク容量不足テストは環境依存のためスキップ")
    })
}
```

#### 2. エラー回復シナリオテスト
```go
// tests/e2e/error_scenarios/recovery_scenarios_test.go
package error_scenarios

import (
    "testing"
)

// TestErrorRecovery_ValidationErrorsWithFix は検証エラーの修正テスト
func TestErrorRecovery_ValidationErrorsWithFix(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // 複数の問題を含むスクリプト
    problematicScript := `#!/bin/bash
# 複数の問題を含むスクリプト
usacloud serv list              # typo
usacloud iso-image list         # deprecated  
usacloud server lst --zone=all  # subcommand typo
usacloud invalid-cmd read       # invalid command
usacloud disk create size 100   # missing --
`
    
    inputFile := suite.CreateTempFile("problematic.sh", problematicScript)
    
    t.Run("InteractiveFixProcess", func(t *testing.T) {
        options := &E2ETestOptions{
            Arguments: []string{
                "--interactive",
                "--validate-and-fix",
                inputFile,
            },
            InteractiveInputs: []string{
                "y", // server への修正を受け入れ
                "y", // cdrom への修正を受け入れ
                "y", // list への修正を受け入れ
                "s", // invalid-cmd はスキップ
                "y", // --size への修正を受け入れ
                "y", // 変更を保存
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "5個の問題が検出されました",
                "修正候補",
                "4個の問題を修正しました",
                "1個の問題をスキップしました",
                "修正完了",
            },
        }
        
        result := suite.RunE2ETest("InteractiveFixProcess", options)
        
        // 修正後のファイル内容検証
        suite.ValidateFixedContent(inputFile, []string{
            "usacloud server list",
            "usacloud cdrom list", 
            "usacloud server list --zone=all",
            "usacloud invalid-cmd read", // そのまま
            "usacloud disk create --size 100",
        })
    })
    
    t.Run("AutoFixWithReport", func(t *testing.T) {
        options := &E2ETestOptions{
            Arguments: []string{
                "--auto-fix",
                "--report-level", "detailed",
                inputFile,
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "自動修正モード",
                "詳細レポート",
                "修正不可能な問題",
                "手動確認が必要",
            },
        }
        
        result := suite.RunE2ETest("AutoFixWithReport", options)
        
        // 自動修正されたものと手動対応が必要なものを区別
        suite.ValidatePartialAutoFix(result)
        suite.ValidateManualActionRequired(result)
    })
}

// TestErrorRecovery_ConfigurationErrors は設定エラーの回復テスト
func TestErrorRecovery_ConfigurationErrors(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    t.Run("CorruptedConfigFile", func(t *testing.T) {
        // 破損した設定ファイル
        corruptedConfig := `[general
color_output = true
language = invalid-language
profile = non-existent-profile

[validation]
strict_mode = maybe  # invalid boolean
max_suggestions = -5  # invalid number
`
        
        configFile := suite.CreateTempFile("corrupted.conf", corruptedConfig)
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--config", configFile,
                "--validate-only",
                "/dev/null", // ダミー入力
            },
            ExpectedExitCode: 1,
            ExpectedStderr: []string{
                "設定ファイルエラー",
                "構文エラー",
                "デフォルト設定を使用",
            },
        }
        
        result := suite.RunE2ETest("CorruptedConfigFile", options)
        
        // フォールバック機能の確認
        suite.ValidateConfigFallback(result)
        
        // 修正方法の提案確認
        suite.ValidateConfigFixSuggestions(result)
    })
}
```

### インタラクティブモードテスト

#### 1. 対話式処理テスト
```go
// tests/e2e/integration_scenarios/interactive_mode_test.go
package integration_scenarios

import (
    "testing"
    "time"
)

// TestInteractiveMode_CommandBuilding は対話式コマンド構築テスト
func TestInteractiveMode_CommandBuilding(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    t.Run("GuidedCommandCreation", func(t *testing.T) {
        options := &E2ETestOptions{
            Arguments: []string{
                "--interactive",
                "--command-builder",
            },
            InteractiveInputs: []string{
                "server",         // メインコマンド選択
                "create",         // サブコマンド選択  
                "test-server",    // サーバー名
                "2",              // CPU数
                "4",              // メモリ(GB)
                "is1a",           // ゾーン
                "y",              // コマンド実行確認
                "n",              // 実際の実行はしない
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "コマンド構築ヘルパー",
                "メインコマンドを選択",
                "サブコマンドを選択",
                "生成されたコマンド:",
                "usacloud server create",
                "--name test-server",
                "--cpu 2",
                "--memory 4",
                "--zone is1a",
            },
            Timeout: "2m", // インタラクティブモードは時間がかかる
        }
        
        result := suite.RunE2ETest("GuidedCommandCreation", options)
        suite.ValidateInteractiveFlow(result)
    })
    
    t.Run("HelpSystemIntegration", func(t *testing.T) {
        options := &E2ETestOptions{
            Arguments: []string{
                "--interactive",
                "--help-system",
            },
            InteractiveInputs: []string{
                "server",         // コマンドについて質問
                "create",         // サブコマンドについて質問
                "examples",       // 使用例を要求
                "q",              // 終了
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "ヘルプシステム",
                "server コマンドについて",
                "create サブコマンドについて",
                "使用例:",
                "usacloud server create",
            },
        }
        
        result := suite.RunE2ETest("HelpSystemIntegration", options)
        suite.ValidateHelpSystemResponsiveness(result)
    })
}

// TestInteractiveMode_ErrorHandling は対話式エラーハンドリングテスト
func TestInteractiveMode_ErrorHandling(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    t.Run("InvalidInputRecovery", func(t *testing.T) {
        options := &E2ETestOptions{
            Arguments: []string{
                "--interactive",
                "--command-builder",
            },
            InteractiveInputs: []string{
                "invalid-command", // 無効なコマンド
                "server",          // 有効なコマンドに修正
                "invalid-sub",     // 無効なサブコマンド
                "create",          // 有効なサブコマンドに修正
                "quit",            // 途中終了
            },
            ExpectedExitCode: 0,
            ExpectedStdout: []string{
                "無効なコマンドです",
                "候補:",
                "server",
                "無効なサブコマンドです",
                "利用可能なサブコマンド:",
                "create",
                "途中終了しました",
            },
        }
        
        result := suite.RunE2ETest("InvalidInputRecovery", options)
        suite.ValidateGracefulErrorRecovery(result)
    })
}
```

### CI/CD環境テスト

#### 1. CI/CD専用ワークフローテスト
```go
// tests/e2e/user_workflows/ci_workflow_test.go
package user_workflows

import (
    "testing"
)

// TestCIWorkflow_AutomatedProcessing はCI環境での自動処理テスト
func TestCIWorkflow_AutomatedProcessing(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // CI環境をシミュレート
    ciEnvironment := map[string]string{
        "CI":                 "true",
        "USACLOUD_NO_COLOR":  "1",
        "USACLOUD_NO_INTERACTIVE": "1",
    }
    
    t.Run("BatchValidationOnly", func(t *testing.T) {
        scriptFiles := []string{
            "deployment.sh",
            "maintenance.sh", 
            "monitoring.sh",
        }
        
        suite.CreateMultipleScripts(scriptFiles, 50)
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--profile", "ci",
                "--batch-validate",
                "--input-dir", suite.testDir,
                "--report-format", "json",
                "--exit-on-error",
            },
            Environment:      ciEnvironment,
            ExpectedExitCode: 1, // エラーがある場合
            ExpectedStdout: []string{
                "バッチ検証モード",
                "処理対象ファイル: 3",
                "検証結果レポート",
            },
        }
        
        result := suite.RunE2ETest("CIBatchValidation", options)
        
        // CI向け出力の検証
        suite.ValidateCIFriendlyOutput(result)
        suite.ValidateJSONReport(result)
    })
    
    t.Run("ZeroDowntimeValidation", func(t *testing.T) {
        // 本番環境で使用中のスクリプトの検証
        productionScript := `#!/bin/bash
# 本番環境のデプロイメントスクリプト
usacloud server list --output-type json | jq '.[] | .Name'
usacloud loadbalancer update --vip-id 12345 --server-id 67890
usacloud disk snapshot create --disk-id 111 --name backup-$(date +%Y%m%d)
`
        
        inputFile := suite.CreateTempFile("production.sh", productionScript)
        
        options := &E2ETestOptions{
            Arguments: []string{
                "--validate-only",
                "--strict-mode",
                "--no-suggestions", // 本番では提案不要
                inputFile,
            },
            Environment:      ciEnvironment,
            ExpectedExitCode: 0, // 問題なしを期待
            ExpectedStdout: []string{
                "検証完了",
                "問題は検出されませんでした",
            },
            ValidatePerformance: true,
            MaxExecutionTime:   "5s", // CI環境では高速化が重要
        }
        
        result := suite.RunE2ETest("ZeroDowntimeValidation", options)
        suite.ValidateProductionSafety(result)
    })
}
```

## テスト戦略
- **現実性**: 実際のユーザー体験を忠実に再現
- **網羅性**: 正常系・異常系・エッジケースを包括的にカバー
- **自動化**: CI/CD環境での安定した自動実行
- **隔離性**: テスト間の相互影響を排除
- **保守性**: テストケースの追加・変更が容易

## 依存関係
- 前提PBI: PBI-015～017 (統合システム), PBI-018～021 (他テスト戦略)
- 外部ツール: shell scripting, process monitoring tools

## 見積もり
- 開発工数: 18時間
  - E2Eテストフレームワーク実装: 5時間
  - ユーザーワークフローテスト実装: 5時間
  - エラーシナリオテスト実装: 4時間
  - インタラクティブモードテスト実装: 2時間
  - CI/CD統合テスト実装: 2時間

## 完了の定義
- [ ] E2Eテストフレームワークが実装されている
- [ ] ユーザータイプ別ワークフローテストが実装されている
- [ ] 包括的なエラーシナリオテストが実装されている
- [ ] インタラクティブモードのE2Eテストが実装されている
- [ ] CI/CD環境でのワークフローテストが実装されている
- [ ] ファイルI/O・設定管理の統合テストが実装されている
- [ ] エラー回復シナリオテストが実装されている
- [ ] CI/CDでの自動実行が安定している
- [ ] 全E2Eテストが継続的に通過している
- [ ] コードレビューが完了している

## 実装状況
❌ **PBI-022は未実装** (2025-09-11)

**現在の状況**:
- 包括的なE2Eテスト戦略とアーキテクチャが設計済み
- ユーザータイプ別ワークフロー、エラーシナリオ、インタラクティブモードの詳細設計完了
- ファイルI/O、設定管理、エラー回復シナリオの仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `tests/e2e/` - E2Eテストフレームワークとテストスイート
- ユーザータイプ別ワークフローテスト（初心者、上級者、CI/CD）
- 包括的なエラーシナリオとエラー回復テスト
- インタラクティブモードの自動テスト機能
- ファイルI/O、設定管理の統合テスト
- CI/CDパイプラインでの自動実行環境

**次のステップ**:
1. E2Eテストフレームワークの基盤実装
2. ユーザータイプ別ワークフローテストの実装
3. エラーシナリオとインタラクティブモードテストの実装
4. ファイルI/Oと設定管理の統合テスト実装
5. CI/CD統合とエラー回復シナリオの実装

## 備考
- E2Eテストは実行時間が長いため、効率的な並列実行と選択的実行が重要
- 実際のユーザーフィードバックを基にしたシナリオの継続的改善が必要
- テスト環境の分離により、他のテストへの影響を防ぐことが重要
- 複雑なE2Eテストの保守性を考慮した可読性の高い実装が必要

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- E2Eテストシナリオ拡張よりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存E2Eテストの安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のE2Eテストは引き続き保守・改善
- 新規E2Eテストシナリオの実装は延期
- 現在のテスト基盤の安定化を最優先