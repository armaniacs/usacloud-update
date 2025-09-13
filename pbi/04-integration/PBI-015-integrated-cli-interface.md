# PBI-015: 統合CLIインターフェース

## 概要
開発されたコマンド検証・エラーフィードバックシステムを既存のusacloud-updateツールに統合し、シームレスなユーザー体験を提供する統合CLIインターフェースを実装する。既存の変換機能と新しい検証機能が調和して動作する設計を実現する。

## 受け入れ条件
- [ ] 既存のusacloud-updateフラグ体系と整合性が保たれている
- [ ] 変換前の検証オプションが適切に実装されている
- [ ] インタラクティブモードとバッチモードの両方に対応している
- [ ] エラー処理が統一されており一貫したUXを提供している
- [ ] 既存機能に影響を与えることなく新機能が統合されている

## 技術仕様

### CLI拡張設計

#### 1. 新しいフラグオプション
```bash
# 既存フラグ（変更なし）
usacloud-update --in input.sh --out output.sh --stats

# 新しい検証フラグ
usacloud-update --validate-only input.sh              # 検証のみ実行
usacloud-update --strict-validation input.sh          # 厳格検証モード
usacloud-update --interactive input.sh                # インタラクティブモード
usacloud-update --help-mode enhanced                  # 拡張ヘルプモード
usacloud-update --suggestion-level 3                  # 提案レベル設定
usacloud-update --skip-deprecated                     # 廃止コマンド警告をスキップ
```

#### 2. 実装構造
```go
// cmd/usacloud-update/main.go への統合
package main

import (
    "flag"
    "fmt"
    "os"
    
    "github.com/armaniacs/usacloud-update/internal/transform"
    "github.com/armaniacs/usacloud-update/internal/validation"
)

// Config は統合された設定
type Config struct {
    // 既存設定
    InputPath  string
    OutputPath string
    ShowStats  bool
    
    // 新しい検証設定
    ValidateOnly      bool
    StrictValidation  bool 
    InteractiveMode   bool
    HelpMode         string
    SuggestionLevel  int
    SkipDeprecated   bool
    ColorEnabled     bool
    LanguageCode     string
}

// ValidationConfig は検証システム設定
type ValidationConfig struct {
    MaxSuggestions    int
    MaxDistance       int
    EnableTypoDetection bool
    EnableInteractiveHelp bool
    ErrorFormat       string
    LogLevel          string
}

// IntegratedCLI は統合CLIインターフェース
type IntegratedCLI struct {
    config           *Config
    validationConfig *ValidationConfig
    transformEngine  *transform.Engine
    validationSystem *validation.ValidationSystem
    helpSystem       *validation.UserFriendlyHelpSystem
}

// NewIntegratedCLI は新しい統合CLIを作成
func NewIntegratedCLI() *IntegratedCLI {
    return &IntegratedCLI{
        config:           parseFlags(),
        validationConfig: loadValidationConfig(),
    }
}

// Run はメインの実行フロー
func (cli *IntegratedCLI) Run() error {
    // 初期化
    if err := cli.initialize(); err != nil {
        return fmt.Errorf("初期化に失敗しました: %w", err)
    }
    
    // インタラクティブモード
    if cli.config.InteractiveMode {
        return cli.runInteractiveMode()
    }
    
    // バッチモード
    return cli.runBatchMode()
}
```

### 統合実行フロー

#### 1. バッチモード処理フロー
```go
func (cli *IntegratedCLI) runBatchMode() error {
    // Step 1: 入力ファイル読み込み
    content, err := cli.readInputFile()
    if err != nil {
        return err
    }
    
    // Step 2: 事前検証（オプション）
    if cli.config.ValidateOnly || cli.config.StrictValidation {
        if err := cli.performPreValidation(content); err != nil {
            if cli.config.ValidateOnly {
                return nil // 検証のみの場合はここで終了
            }
            if cli.config.StrictValidation {
                return err // 厳格モードではエラーで停止
            }
        }
    }
    
    // Step 3: 行ごと処理（既存ロジック拡張）
    results, err := cli.processLines(content)
    if err != nil {
        return err
    }
    
    // Step 4: 出力生成
    return cli.generateOutput(results)
}

func (cli *IntegratedCLI) processLines(lines []string) ([]*transform.Result, error) {
    var results []*transform.Result
    lineNumber := 0
    
    for _, line := range lines {
        lineNumber++
        
        // 既存の変換処理
        transformResult := cli.transformEngine.ApplyRules(line)
        
        // 新しい検証処理（変換前）
        if !cli.config.SkipDeprecated {
            validationResult := cli.validateLine(line, lineNumber)
            if validationResult != nil {
                // 検証結果をtransformResultに統合
                transformResult = cli.mergeValidationResult(transformResult, validationResult)
            }
        }
        
        results = append(results, transformResult)
        
        // リアルタイム出力（既存機能）
        if transformResult.Changed {
            cli.outputColorizedChange(transformResult, lineNumber)
        }
    }
    
    return results, nil
}
```

#### 2. インタラクティブモード処理
```go
func (cli *IntegratedCLI) runInteractiveMode() error {
    fmt.Println("🚀 usacloud-update インタラクティブモードを開始します")
    fmt.Println("   ファイル全体を分析し、推奨される変更を提案します。")
    
    // ファイル分析
    analysis, err := cli.analyzeFile()
    if err != nil {
        return err
    }
    
    // 問題点の表示と選択
    issues := cli.identifyIssues(analysis)
    selectedIssues := cli.selectIssuesInteractively(issues)
    
    // 推奨変更の適用
    return cli.applySelectedChanges(selectedIssues)
}

func (cli *IntegratedCLI) selectIssuesInteractively(issues []ValidationIssue) []ValidationIssue {
    var selected []ValidationIssue
    
    fmt.Printf("\n📋 %d個の問題が検出されました:\n\n", len(issues))
    
    for i, issue := range issues {
        fmt.Printf("  %d. %s (行: %d)\n", i+1, issue.Description, issue.LineNumber)
        fmt.Printf("     現在: %s\n", issue.CurrentCode)
        fmt.Printf("     推奨: %s\n", issue.SuggestedCode)
        fmt.Printf("     理由: %s\n", issue.Reason)
        
        fmt.Printf("\n     この変更を適用しますか？ [y/N/s(skip)/q(quit)]: ")
        
        response := cli.readUserInput()
        switch strings.ToLower(response) {
        case "y", "yes":
            selected = append(selected, issue)
            fmt.Printf("     ✅ 適用予定に追加しました\n\n")
        case "s", "skip":
            fmt.Printf("     ⏭️  スキップしました\n\n")
        case "q", "quit":
            fmt.Printf("     🚪 インタラクティブモードを終了します\n")
            return selected
        default:
            fmt.Printf("     ❌ 適用しませんでした\n\n")
        }
    }
    
    return selected
}
```

### エラーハンドリング統合

#### 1. 統一エラー処理
```go
// ErrorHandler は統一エラーハンドラー
type ErrorHandler struct {
    formatter    *validation.ComprehensiveErrorFormatter
    colorEnabled bool
    verboseMode  bool
}

// HandleError はエラーを統一的に処理
func (h *ErrorHandler) HandleError(err error, context *ErrorContext) {
    switch e := err.(type) {
    case *ValidationError:
        h.handleValidationError(e, context)
    case *TransformationError:
        h.handleTransformationError(e, context)
    case *FileError:
        h.handleFileError(e, context)
    default:
        h.handleGenericError(e, context)
    }
}

func (h *ErrorHandler) handleValidationError(err *ValidationError, context *ErrorContext) {
    // 検証エラーの包括的な処理
    errorMessage := h.formatter.FormatError(&validation.ErrorContext{
        InputCommand:    context.Command,
        DetectedIssues: err.Issues,
        Suggestions:    err.Suggestions,
    })
    
    if h.colorEnabled {
        fmt.Fprintf(os.Stderr, "%s\n", errorMessage)
    } else {
        fmt.Fprintf(os.Stderr, "%s\n", stripAnsiCodes(errorMessage))
    }
    
    // 必要に応じて詳細情報を表示
    if h.verboseMode && len(err.Details) > 0 {
        fmt.Fprintf(os.Stderr, "\n詳細情報:\n%s\n", err.Details)
    }
}
```

#### 2. プログレス表示統合
```go
// ProgressReporter は進捗表示
type ProgressReporter struct {
    totalLines    int
    currentLine   int
    issuesFound   int
    changesApplied int
    colorEnabled  bool
}

// ReportProgress は進捗を報告
func (p *ProgressReporter) ReportProgress(lineNum int, result *ProcessResult) {
    p.currentLine = lineNum
    
    if result.HasIssues() {
        p.issuesFound++
    }
    if result.HasChanges() {
        p.changesApplied++
    }
    
    // プログレスバーの表示
    progress := float64(p.currentLine) / float64(p.totalLines)
    p.displayProgressBar(progress)
    
    // 統計情報の表示
    if p.currentLine%100 == 0 || p.currentLine == p.totalLines {
        p.displayStats()
    }
}

func (p *ProgressReporter) displayProgressBar(progress float64) {
    if !p.colorEnabled {
        return
    }
    
    barLength := 40
    filledLength := int(progress * float64(barLength))
    
    bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
    fmt.Printf("\r進捗: [%s] %.1f%% (%d/%d行)", bar, progress*100, p.currentLine, p.totalLines)
}
```

### 設定ファイル統合

#### 1. 統合設定ファイル
```ini
# usacloud-update.conf

[general]
color_output = true
language = ja
verbose = false
interactive_by_default = false

[validation]
enable_validation = true
strict_mode = false
max_suggestions = 5
max_edit_distance = 3
skip_deprecated_warnings = false

[output]
format = auto
show_line_numbers = true
show_progress = true
preserve_comments = true

[help]
enable_interactive_help = true
skill_level = intermediate
preferred_format = detailed
```

#### 2. 設定読み込み処理
```go
// loadConfig は統合設定を読み込み
func loadConfig() (*Config, error) {
    config := &Config{
        // デフォルト値
        ColorEnabled:    true,
        LanguageCode:    "ja",
        SuggestionLevel: 5,
    }
    
    // 設定ファイルから読み込み
    if err := config.loadFromFile(); err != nil {
        // 設定ファイルが見つからない場合はデフォルト値を使用
        if !os.IsNotExist(err) {
            return nil, err
        }
    }
    
    // 環境変数から読み込み
    config.loadFromEnvironment()
    
    // コマンドライン引数から読み込み（最優先）
    config.loadFromFlags()
    
    return config, nil
}
```

## テスト戦略
- 統合テスト：既存機能と新機能の組み合わせが正しく動作することを確認
- 回帰テスト：既存のゴールデンファイルテストが継続して通過することを確認
- インタラクティブテスト：インタラクティブモードの各種操作パターンを検証
- パフォーマンステスト：大きなファイルでの処理性能が許容範囲内であることを確認
- エラーハンドリングテスト：様々なエラー条件での適切な処理を確認
- 設定テスト：設定ファイル、環境変数、フラグの優先順位が正しいことを確認

## 依存関係
- 前提PBI: PBI-001～014 (全コンポーネント)
- 既存コード: 既存のusacloud-updateエンジンとの統合
- 関連PBI: PBI-016 (変換エンジン統合), PBI-017 (設定統合)

## 見積もり
- 開発工数: 10時間
  - CLI統合実装: 3時間
  - インタラクティブモード実装: 2.5時間
  - エラーハンドリング統合: 2時間
  - 設定システム統合: 1.5時間
  - 統合テスト作成: 1時間

## 完了の定義
- [ ] `cmd/usacloud-update/main.go`が新機能で拡張されている
- [ ] 既存のフラグ体系と整合性を保ちつつ新フラグが追加されている
- [ ] インタラクティブモードが完全に実装されている
- [ ] バッチモードでの検証機能が正しく統合されている
- [ ] 統一されたエラーハンドリングが実装されている
- [ ] プログレス表示が既存機能と統合されている
- [ ] 設定ファイル統合が完了している
- [ ] 全ての既存テストが継続して通過している
- [ ] 新機能の包括的テストが作成され、すべて通過している
- [ ] 統合後のパフォーマンスが許容範囲内である
- [ ] コードレビューが完了している

## 備考
- 既存のusacloud-updateユーザーへの影響を最小限に抑える設計が重要
- 新機能はオプトイン方式で提供し、既存の動作を変更しない
- インタラクティブモードは初心者ユーザーの学習支援に重点を置く
- パフォーマンスの回帰がないことを保証する継続的な監視が必要

## 実装状況
🟠 **PBI-015は部分実装** (2025-09-11)

### 現在の状況
- 基本的のCLIインターフェースは実装済み（`cmd/usacloud-update/main.go`）
- 既存の --in, --out, --stats フラグは動作中
- 基本的な変換エンジンは統合済み
- サンドボックス機能とTUI機能の一部が実装済み

### 未実装の要素
1. **拡張検証フラグ**
   - --validate-only: 検証のみ実行フラグ
   - --strict-validation: 厳格検証モード
   - --suggestion-level: 提案レベル設定
   - --skip-deprecated: 廃止コマンド警告スキップ
   - --help-mode enhanced: 拡張ヘルプモード

2. **高度なインタラクティブモード**
   - 対話式問題選択インターフェース
   - ファイル全体分析と推奨変更提案
   - ユーザー選択ベースの変更適用
   - リアルタイムフィードバックシステム

3. **統合エラーハンドリング**
   - ErrorHandler クラスと統一エラー処理
   - 検証エラー、変換エラー、ファイルエラーの統合処理
   - ComprehensiveErrorFormatter との連携
   - カラー出力とプレーンテキストの動的切り替え

4. **高度なプログレス表示**
   - ProgressReporter クラスの実装
   - リアルタイムプログレスバー表示
   - 統計情報の動的表示
   - パフォーマンスメトリクスの可視化

5. **設定系統合**
   - 統合設定ファイルシステム
   - 設定ファイル・環境変数・フラグの優先順位管理
   - ValidationConfig 構造体と設定ローダー
   - ユーザープリファレンス管理

### 部分実装済みの要素
✅ **既存CLIフラグ体系**: --in, --out, --stats, --sandbox, --interactive, --dry-run, --batch
✅ **基本的な変換エンジン統合**: transform.Engine との連携
✅ **基本的な設定管理**: internal/config/ パッケージ
✅ **TUIモードの一部**: サンドボックス用インタラクティブTUI

### 次のステップ
1. 拡張検証フラグの`cmd/usacloud-update/main.go`への追加
2. ValidationSystem との統合インターフェース実装
3. 高度なインタラクティブモードの実装
4. 統合ErrorHandlerクラスの作成
5. ProgressReporterシステムの実装
6. 設定システムの完全統合
7. 既存機能との互換性テスト
8. パフォーマンスリグレッションテスト

### 関連ファイル
- 主要統合先: `cmd/usacloud-update/main.go` ✅
- 拡張予定: `cmd/usacloud-update/interactive.go`
- 拡張予定: `cmd/usacloud-update/validation.go`
- 拡張予定: `cmd/usacloud-update/error_handler.go`
- 統合対象: `internal/validation/` パッケージ
- 統合対象: `internal/transform/engine.go` ✅
- 設定連携: `internal/config/` パッケージ ✅