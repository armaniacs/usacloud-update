# PBI-021: ゴールデンファイルテスト拡張

## 概要
既存のusacloud-updateゴールデンファイルテストシステムを拡張し、新しいコマンド検証・エラーフィードバック機能を包含する包括的なテストフレームワークに発展させる。変換結果だけでなく、検証結果、エラーメッセージ、ヘルプ出力も含む多次元ゴールデンファイルテストを実装する。

## 受け入れ条件
- [ ] 既存のゴールデンファイルテストとの完全な後方互換性が保たれている
- [ ] 変換結果、検証結果、エラーメッセージの統合テストが実装されている
- [ ] テストデータの自動生成と更新機能が実装されている
- [ ] 多言語対応（日本語/英語）のゴールデンファイルが整備されている
- [ ] CI/CDでの自動更新と差分確認が実装されている

## 技術仕様

### 拡張ゴールデンファイル構造

#### 1. 拡張テストディレクトリ構造
```
testdata/
├── golden/
│   ├── transforms/                          # 既存の変換テスト
│   │   ├── expected_v1_1.sh               # 既存ファイル（維持）
│   │   ├── expected_v1_1_with_validation.sh # 検証付き変換結果
│   │   └── expected_v1_1_strict.sh         # 厳格モード結果
│   ├── validations/                         # 新しい検証テスト
│   │   ├── command_validation.golden       # コマンド検証結果
│   │   ├── error_detection.golden          # エラー検出結果
│   │   ├── suggestions.golden              # 提案結果
│   │   └── deprecated_warnings.golden      # 廃止警告結果
│   ├── errors/                             # エラーメッセージテスト
│   │   ├── japanese_errors.golden          # 日本語エラーメッセージ
│   │   ├── english_errors.golden           # 英語エラーメッセージ
│   │   ├── colored_errors.golden           # カラー付きエラー
│   │   └── plain_errors.golden             # プレーンテキストエラー
│   ├── help/                               # ヘルプ出力テスト
│   │   ├── interactive_help.golden         # インタラクティブヘルプ
│   │   ├── contextual_help.golden          # コンテキストヘルプ
│   │   └── beginner_help.golden            # 初心者向けヘルプ
│   └── integration/                        # 統合テスト
│       ├── full_pipeline.golden            # 完全パイプライン結果
│       ├── error_recovery.golden           # エラー回復結果
│       └── multi_issue.golden              # 複数問題処理結果
├── inputs/                                 # テスト入力
│   ├── sample_v0_v1_mixed.sh              # 既存サンプル（維持）
│   ├── complex_mixed_versions.sh           # 複雑な混在スクリプト
│   ├── error_scenarios.sh                  # エラーシナリオ
│   ├── deprecated_commands.sh              # 廃止コマンド集
│   └── typo_commands.sh                    # typoパターン集
└── configs/                               # テスト用設定
    ├── default.conf                       # デフォルト設定
    ├── strict.conf                        # 厳格モード設定
    ├── beginner.conf                      # 初心者設定
    └── ci.conf                            # CI環境設定
```

#### 2. 拡張ゴールデンファイルフレームワーク
```go
// internal/testing/golden_test_framework.go
package testing

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "testing"
    
    "github.com/armaniacs/usacloud-update/internal/config"
    "github.com/armaniacs/usacloud-update/internal/transform"
    "github.com/armaniacs/usacloud-update/internal/validation"
)

// GoldenTestSuite は拡張ゴールデンテストスイート
type GoldenTestSuite struct {
    t           *testing.T
    testDataDir string
    updateFlag  bool    // -update フラグの状態
    
    // テスト対象システム
    integratedCLI *IntegratedCLI
    config        *config.IntegratedConfig
}

// GoldenTestResult はゴールデンテスト結果
type GoldenTestResult struct {
    // 変換結果
    TransformOutput     string                    `json:"transform_output"`
    TransformStats      *TransformStats           `json:"transform_stats"`
    
    // 検証結果
    ValidationResults   []ValidationResult        `json:"validation_results"`
    ValidationSummary   *ValidationSummary        `json:"validation_summary"`
    
    // エラー・警告
    ErrorMessages       []ErrorMessage            `json:"error_messages"`
    WarningMessages     []WarningMessage          `json:"warning_messages"`
    
    // 提案
    Suggestions         []SuggestionResult        `json:"suggestions"`
    DeprecationWarnings []DeprecationWarning      `json:"deprecation_warnings"`
    
    // ヘルプ出力
    HelpOutput          string                    `json:"help_output"`
    InteractiveOutput   string                    `json:"interactive_output"`
    
    // メタデータ
    TestMetadata        *TestMetadata             `json:"test_metadata"`
}

// TestMetadata はテストメタデータ
type TestMetadata struct {
    TestName      string    `json:"test_name"`
    InputFile     string    `json:"input_file"`
    ConfigUsed    string    `json:"config_used"`
    TestDate      string    `json:"test_date"`
    ToolVersion   string    `json:"tool_version"`
    Language      string    `json:"language"`
    ColorEnabled  bool      `json:"color_enabled"`
}

// NewGoldenTestSuite は新しいゴールデンテストスイートを作成
func NewGoldenTestSuite(t *testing.T) *GoldenTestSuite {
    return &GoldenTestSuite{
        t:           t,
        testDataDir: "testdata",
        updateFlag:  updateGoldenFiles(), // フラグから取得
    }
}

// RunGoldenTest はゴールデンテストを実行
func (gts *GoldenTestSuite) RunGoldenTest(testName string, options *GoldenTestOptions) {
    gts.t.Helper()
    
    // 入力ファイル読み込み
    inputPath := filepath.Join(gts.testDataDir, "inputs", options.InputFile)
    input, err := ioutil.ReadFile(inputPath)
    if err != nil {
        gts.t.Fatalf("入力ファイル読み込みエラー %s: %v", inputPath, err)
    }
    
    // テスト実行
    result := gts.executeTest(testName, string(input), options)
    
    // ゴールデンファイル比較
    gts.compareWithGoldenFile(testName, result, options)
}

// executeTest はテストを実行
func (gts *GoldenTestSuite) executeTest(
    testName, input string,
    options *GoldenTestOptions,
) *GoldenTestResult {
    // 設定読み込み
    config := gts.loadTestConfig(options.ConfigFile)
    
    // 統合CLIの初期化
    cli := NewIntegratedCLI(config)
    
    // テスト実行
    var result GoldenTestResult
    
    // 1. 変換処理実行
    if options.IncludeTransform {
        transformResult := cli.ProcessInput(input)
        result.TransformOutput = transformResult.Output
        result.TransformStats = transformResult.Stats
    }
    
    // 2. 検証処理実行
    if options.IncludeValidation {
        validationResults := cli.ValidateInput(input)
        result.ValidationResults = validationResults
        result.ValidationSummary = cli.SummarizeValidation(validationResults)
    }
    
    // 3. エラーメッセージ生成
    if options.IncludeErrors {
        errorMessages := cli.GenerateErrorMessages(result.ValidationResults)
        result.ErrorMessages = errorMessages
    }
    
    // 4. 提案生成
    if options.IncludeSuggestions {
        suggestions := cli.GenerateSuggestions(result.ValidationResults)
        result.Suggestions = suggestions
    }
    
    // 5. ヘルプ出力生成
    if options.IncludeHelp {
        helpOutput := cli.GenerateHelp(input, result.ValidationResults)
        result.HelpOutput = helpOutput
    }
    
    // メタデータ設定
    result.TestMetadata = &TestMetadata{
        TestName:     testName,
        InputFile:    options.InputFile,
        ConfigUsed:   options.ConfigFile,
        TestDate:     getCurrentTimestamp(),
        ToolVersion:  getToolVersion(),
        Language:     config.General.Language,
        ColorEnabled: config.General.ColorOutput,
    }
    
    return &result
}
```

### 多次元ゴールデンファイル管理

#### 1. 構造化ゴールデンファイル形式
```json
// testdata/golden/integration/full_pipeline.golden
{
  "test_metadata": {
    "test_name": "full_pipeline",
    "input_file": "sample_v0_v1_mixed.sh",
    "config_used": "default.conf",
    "test_date": "2025-01-15T10:30:00Z",
    "tool_version": "1.9.0",
    "language": "ja",
    "color_enabled": false
  },
  "transform_output": "#!/bin/bash\n# usacloud-update で変換されたスクリプト\n# 元ファイル: sample_v0_v1_mixed.sh\n# 変換日時: 2025-01-15T10:30:00Z\n# 変換ルール適用数: 5\n\nusacloud server list --output-type json  # usacloud-update: csv → json 形式変更 https://docs.usacloud.jp/\nusacloud cdrom list  # usacloud-update: iso-image → cdrom 名称変更 https://docs.usacloud.jp/",
  "transform_stats": {
    "total_lines": 10,
    "processed_lines": 8,
    "transformed_lines": 5,
    "skipped_lines": 2,
    "rules_applied": [
      "output_format_csv_to_json",
      "resource_rename_iso_image"
    ]
  },
  "validation_results": [
    {
      "line_number": 3,
      "original_line": "usacloud server list --output-type csv",
      "validation_status": "warning",
      "issues": [
        {
          "type": "deprecated_parameter",
          "severity": "warning",
          "message": "csv出力形式は非推奨です。json形式の使用を推奨します。",
          "suggestion": "--output-type json"
        }
      ]
    },
    {
      "line_number": 5,
      "original_line": "usacloud iso-image list",
      "validation_status": "error",
      "issues": [
        {
          "type": "deprecated_command",
          "severity": "error", 
          "message": "iso-imageコマンドは廃止されました。cdromコマンドを使用してください。",
          "replacement_command": "cdrom"
        }
      ]
    }
  ],
  "validation_summary": {
    "total_issues": 2,
    "errors": 1,
    "warnings": 1,
    "suggestions": 2,
    "deprecated_commands": 1
  },
  "error_messages": [
    {
      "type": "deprecated_command_error",
      "formatted_message": "❌ エラー: 'iso-image' コマンドはv1で廃止されました。\n\n🔄 代わりに以下を使用してください:\n   usacloud cdrom list\n\nℹ️  詳細な移行ガイド: https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/"
    }
  ],
  "suggestions": [
    {
      "line_number": 3,
      "original_command": "server list --output-type csv",
      "suggested_command": "server list --output-type json",
      "reason": "JSON形式の方が構造化データの処理に適しています",
      "confidence": 0.9
    }
  ],
  "deprecation_warnings": [
    {
      "deprecated_command": "iso-image",
      "replacement_command": "cdrom",
      "deprecation_version": "v1.0.0",
      "migration_guide_url": "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/"
    }
  ]
}
```

#### 2. 差分検出とレポート生成
```go
// compareWithGoldenFile はゴールデンファイルとの比較
func (gts *GoldenTestSuite) compareWithGoldenFile(
    testName string,
    result *GoldenTestResult,
    options *GoldenTestOptions,
) {
    goldenPath := gts.getGoldenFilePath(testName, options)
    
    // 現在の結果をJSONに変換
    currentJSON, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        gts.t.Fatalf("結果のJSON変換エラー: %v", err)
    }
    
    if gts.updateFlag {
        // ゴールデンファイル更新
        gts.updateGoldenFile(goldenPath, currentJSON)
        return
    }
    
    // 既存ゴールデンファイル読み込み
    expectedJSON, err := ioutil.ReadFile(goldenPath)
    if err != nil {
        if os.IsNotExist(err) {
            gts.t.Fatalf("ゴールデンファイルが存在しません: %s\n"+
                "-update フラグを使用してファイルを作成してください", goldenPath)
        }
        gts.t.Fatalf("ゴールデンファイル読み込みエラー: %v", err)
    }
    
    // 差分検出
    diff := gts.generateDetailedDiff(expectedJSON, currentJSON)
    if diff != nil {
        gts.t.Errorf("ゴールデンファイルテスト失敗: %s\n\n%s\n\n"+
            "ファイルを更新する場合は -update フラグを使用してください",
            testName, diff.Report())
    }
}

// DetailedDiff は詳細差分情報
type DetailedDiff struct {
    HasDifferences      bool                    `json:"has_differences"`
    TransformDiff       *SectionDiff            `json:"transform_diff"`
    ValidationDiff      *SectionDiff            `json:"validation_diff"`
    ErrorMessageDiff    *SectionDiff            `json:"error_message_diff"`
    SuggestionDiff      *SectionDiff            `json:"suggestion_diff"`
    MetadataDiff        *SectionDiff            `json:"metadata_diff"`
}

// SectionDiff はセクション別差分
type SectionDiff struct {
    SectionName string      `json:"section_name"`
    HasChanges  bool        `json:"has_changes"`
    AddedLines  []string    `json:"added_lines"`
    RemovedLines []string   `json:"removed_lines"`
    ModifiedLines []LineDiff `json:"modified_lines"`
}

// LineDiff は行レベル差分
type LineDiff struct {
    LineNumber int    `json:"line_number"`
    Expected   string `json:"expected"`
    Actual     string `json:"actual"`
}

// generateDetailedDiff は詳細差分を生成
func (gts *GoldenTestSuite) generateDetailedDiff(expected, actual []byte) *DetailedDiff {
    var expectedResult, actualResult GoldenTestResult
    
    json.Unmarshal(expected, &expectedResult)
    json.Unmarshal(actual, &actualResult)
    
    diff := &DetailedDiff{}
    
    // 変換結果の差分
    diff.TransformDiff = gts.compareTransformOutput(
        expectedResult.TransformOutput,
        actualResult.TransformOutput)
    
    // 検証結果の差分
    diff.ValidationDiff = gts.compareValidationResults(
        expectedResult.ValidationResults,
        actualResult.ValidationResults)
    
    // エラーメッセージの差分
    diff.ErrorMessageDiff = gts.compareErrorMessages(
        expectedResult.ErrorMessages,
        actualResult.ErrorMessages)
    
    // 提案の差分
    diff.SuggestionDiff = gts.compareSuggestions(
        expectedResult.Suggestions,
        actualResult.Suggestions)
    
    // 差分の有無を判定
    diff.HasDifferences = diff.TransformDiff.HasChanges ||
        diff.ValidationDiff.HasChanges ||
        diff.ErrorMessageDiff.HasChanges ||
        diff.SuggestionDiff.HasChanges
    
    return diff
}

// Report は差分レポートを生成
func (dd *DetailedDiff) Report() string {
    if !dd.HasDifferences {
        return "差分なし"
    }
    
    var report strings.Builder
    report.WriteString("📊 ゴールデンファイル差分レポート\n")
    report.WriteString("================================\n\n")
    
    if dd.TransformDiff.HasChanges {
        report.WriteString("🔄 変換結果の差分:\n")
        report.WriteString(dd.TransformDiff.formatDiff())
        report.WriteString("\n")
    }
    
    if dd.ValidationDiff.HasChanges {
        report.WriteString("🔍 検証結果の差分:\n")
        report.WriteString(dd.ValidationDiff.formatDiff())
        report.WriteString("\n")
    }
    
    if dd.ErrorMessageDiff.HasChanges {
        report.WriteString("❌ エラーメッセージの差分:\n")
        report.WriteString(dd.ErrorMessageDiff.formatDiff())
        report.WriteString("\n")
    }
    
    if dd.SuggestionDiff.HasChanges {
        report.WriteString("💡 提案の差分:\n")
        report.WriteString(dd.SuggestionDiff.formatDiff())
        report.WriteString("\n")
    }
    
    return report.String()
}
```

### テスト実行とCI/CD統合

#### 1. 拡張ゴールデンテスト実装
```go
// tests/golden_extended_test.go
package tests

import (
    "testing"
)

// TestGolden_TransformWithValidation は変換＋検証のゴールデンテスト
func TestGolden_TransformWithValidation(t *testing.T) {
    suite := NewGoldenTestSuite(t)
    
    testCases := []struct {
        name    string
        options *GoldenTestOptions
    }{
        {
            name: "BasicTransformWithValidation",
            options: &GoldenTestOptions{
                InputFile:         "sample_v0_v1_mixed.sh",
                ConfigFile:        "default.conf",
                IncludeTransform:  true,
                IncludeValidation: true,
                IncludeErrors:     true,
                IncludeSuggestions: true,
            },
        },
        {
            name: "StrictModeValidation",
            options: &GoldenTestOptions{
                InputFile:         "problematic_script.sh",
                ConfigFile:        "strict.conf",
                IncludeTransform:  true,
                IncludeValidation: true,
                IncludeErrors:     true,
                IncludeSuggestions: true,
                StrictMode:        true,
            },
        },
        {
            name: "BeginnerModeHelp",
            options: &GoldenTestOptions{
                InputFile:         "typo_commands.sh",
                ConfigFile:        "beginner.conf",
                IncludeValidation: true,
                IncludeErrors:     true,
                IncludeHelp:       true,
                IncludeSuggestions: true,
                InteractiveMode:   true,
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            suite.RunGoldenTest(tc.name, tc.options)
        })
    }
}

// TestGolden_MultiLanguage は多言語ゴールデンテスト
func TestGolden_MultiLanguage(t *testing.T) {
    suite := NewGoldenTestSuite(t)
    
    languages := []string{"ja", "en"}
    
    for _, lang := range languages {
        t.Run(fmt.Sprintf("Language_%s", lang), func(t *testing.T) {
            options := &GoldenTestOptions{
                InputFile:         "error_scenarios.sh",
                ConfigFile:        fmt.Sprintf("default_%s.conf", lang),
                Language:          lang,
                IncludeValidation: true,
                IncludeErrors:     true,
                IncludeSuggestions: true,
            }
            
            suite.RunGoldenTest(fmt.Sprintf("MultiLanguage_%s", lang), options)
        })
    }
}

// TestGolden_ColorOutput はカラー出力のゴールデンテスト
func TestGolden_ColorOutput(t *testing.T) {
    suite := NewGoldenTestSuite(t)
    
    colorModes := []struct {
        name    string
        enabled bool
    }{
        {"ColorEnabled", true},
        {"PlainText", false},
    }
    
    for _, mode := range colorModes {
        t.Run(mode.name, func(t *testing.T) {
            options := &GoldenTestOptions{
                InputFile:         "error_scenarios.sh",
                ConfigFile:        "default.conf",
                ColorEnabled:      mode.enabled,
                IncludeValidation: true,
                IncludeErrors:     true,
            }
            
            suite.RunGoldenTest(fmt.Sprintf("Color_%s", mode.name), options)
        })
    }
}
```

#### 2. CI/CD統合とゴールデンファイル管理
```yaml
# .github/workflows/golden-tests.yml
name: Golden File Tests

on: [push, pull_request]

jobs:
  golden-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run golden file tests
      run: |
        go test ./tests/... -run=TestGolden -v
    
    - name: Check for golden file changes
      if: failure()
      run: |
        echo "## 🔍 ゴールデンファイルテスト失敗" >> $GITHUB_STEP_SUMMARY
        echo "ゴールデンファイルと実際の出力に差分があります。" >> $GITHUB_STEP_SUMMARY
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "### 修正方法:" >> $GITHUB_STEP_SUMMARY
        echo "1. ローカルで \`go test ./tests/... -run=TestGolden -update\` を実行" >> $GITHUB_STEP_SUMMARY
        echo "2. 更新されたゴールデンファイルを確認・コミット" >> $GITHUB_STEP_SUMMARY
    
    - name: Update golden files (on main branch)
      if: github.ref == 'refs/heads/main' && failure()
      run: |
        go test ./tests/... -run=TestGolden -update
        git config --local user.email "action@github.com"
        git config --local user.name "GitHub Action"
        git add testdata/golden/
        git commit -m "自動更新: ゴールデンファイル [skip ci]" || exit 0
        git push
```

### ゴールデンファイル自動生成

#### 1. テストデータジェネレーター
```go
// internal/testing/golden_generator.go
package testing

import (
    "fmt"
    "math/rand"
    "strings"
    "time"
)

// GoldenDataGenerator はゴールデンテストデータ生成器
type GoldenDataGenerator struct {
    commandPatterns    []CommandPattern
    errorPatterns      []ErrorPattern
    typoPatterns       map[string][]string
    
    rand *rand.Rand
}

// CommandPattern はコマンドパターン
type CommandPattern struct {
    Template    string   // "usacloud {command} {subcommand} {options}"
    Commands    []string // ["server", "disk", "database"]
    Subcommands []string // ["list", "read", "create"]
    Options     []string // ["--output-type json", "--zone is1a"]
}

// GenerateTestScenarios はテストシナリオを生成
func (gdg *GoldenDataGenerator) GenerateTestScenarios(count int) []TestScenario {
    scenarios := make([]TestScenario, count)
    
    for i := 0; i < count; i++ {
        scenario := TestScenario{
            Name:        fmt.Sprintf("GeneratedScenario_%03d", i+1),
            Description: "自動生成されたテストシナリオ",
            Input:       gdg.generateRandomInput(),
            Expected:    gdg.generateExpectedOutput(),
        }
        scenarios[i] = scenario
    }
    
    return scenarios
}

// generateRandomInput はランダムな入力を生成
func (gdg *GoldenDataGenerator) generateRandomInput() ScenarioInput {
    inputTypes := []string{"valid_command", "typo_command", "deprecated_command", "invalid_command"}
    inputType := inputTypes[gdg.rand.Intn(len(inputTypes))]
    
    switch inputType {
    case "valid_command":
        return gdg.generateValidCommand()
    case "typo_command":
        return gdg.generateTypoCommand()
    case "deprecated_command":
        return gdg.generateDeprecatedCommand()
    case "invalid_command":
        return gdg.generateInvalidCommand()
    }
    
    return ScenarioInput{}
}
```

## テスト戦略
- **後方互換性**: 既存のゴールデンファイルテストシステムとの完全な互換性維持
- **多次元検証**: 変換、検証、エラー、ヘルプの統合テスト
- **自動化**: CI/CDでの自動実行と差分検出
- **保守性**: 構造化された差分レポートによる効率的なメンテナンス
- **拡張性**: 新機能追加時の容易なテスト拡張

## 依存関係
- 前提PBI: PBI-001～017 (全実装), 既存のゴールデンファイルテストシステム
- 外部ツール: jq (JSON処理), git (差分管理)

## 見積もり
- 開発工数: 10時間
  - ゴールデンファイルフレームワーク拡張: 4時間
  - 多次元差分検出実装: 2時間
  - テストケース作成: 2時間
  - 自動生成機能実装: 1.5時間
  - CI/CD統合: 0.5時間

## 完了の定義
- [ ] 拡張ゴールデンファイルフレームワークが実装されている
- [ ] 既存のゴールデンファイルテストとの完全な後方互換性が保たれている
- [ ] 多次元ゴールデンファイル（変換、検証、エラー、ヘルプ）が実装されている
- [ ] 詳細差分検出とレポート生成機能が実装されている
- [ ] 多言語対応のゴールデンファイルが整備されている
- [ ] カラー出力・プレーンテキスト両方のテストが実装されている
- [ ] テストデータ自動生成機能が実装されている
- [ ] CI/CDでの自動実行と更新が設定されている
- [ ] 全ゴールデンファイルテストが継続的に通過している
- [ ] コードレビューが完了している

## 実装状況
❌ **PBI-021は未実装** (2025-09-11)

**現在の状況**:
- 拡張ゴールデンファイルテスト戦略とアーキテクチャが設計済み
- 多次元橋渡機能、詳細差分検出、自動生成機能の詳細設計完了
- 既存システムとの後方互換性を保つ拡張方針が確定
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `tests/golden/` - 拡張ゴールデンファイルフレームワーク
- 多次元ゴールデンファイル（変換、検証、エラー、ヘルプ）の実装
- 詳細差分検出とレポート生成機能
- テストデータ自動生成機能とシナリオジェネレーター
- CI/CDパイプラインとの統合設定
- 多言語対応とカラー出力テスト

**次のステップ**:
1. 拡張ゴールデンファイルフレームワークの基盤実装
2. 多次元ゴールデンファイルシステムの構築
3. 詳細差分検出とレポート機能の実装
4. テストデータ自動生成機能の実装
5. CI/CD統合と多言語・カラー対応の実装

## 備考
- 既存のテストケースへの影響を最小限に抑える段階的移行が重要
- ゴールデンファイルのサイズが大きくなるため、効率的な差分表示が必要
- 多言語対応により、文字エンコーディングへの適切な対応が重要
- テスト結果の可視化により、回帰の早期発見を促進することが重要

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- ゴールデンファイルテスト拡張よりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存ゴールデンファイルテストの安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のゴールデンファイルテストは引き続き保守・改善
- 拡張ゴールデンファイルテスト機能の実装は延期
- 現在のテスト基盤の安定化を最優先