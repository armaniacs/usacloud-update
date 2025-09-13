# PBI-013: 包括的エラーフォーマッター

## 概要
検証エンジンからの結果、類似コマンド提案、エラーメッセージ生成器を統合し、一貫性のある包括的なエラー出力を生成するフォーマッターを実装する。ユーザーにとって分かりやすく、次のアクションが明確に示される統合されたエラー体験を提供する。

## 受け入れ条件 ✅ **完了 2025-01-09**
- [x] 全てのエラーコンポーネントが統一されたフォーマットで出力される
- [x] エラーの重要度に応じた視覚的階層が実装されている
- [x] 建設的な提案が適切に優先順位付けされて表示される
- [x] カラー出力とプレーンテキスト出力の両方に対応している
- [x] 国際化対応（日本語/英語）が実装されている

## 技術仕様

### 統合エラーフォーマット

#### 1. エラー出力の基本構造
```
┌─ エラーヘッダー（重要度に応じた色付け）
├─ 問題の詳細説明
├─ 類似コマンド提案（あれば）
├─ 代替手段・移行ガイド（あれば）
└─ 追加情報・ヘルプリンク
```

#### 2. 実装構造
```go
// internal/validation/comprehensive_error_formatter.go
package validation

import (
    "fmt"
    "strings"
)

// ErrorContext はエラーコンテキスト情報
type ErrorContext struct {
    InputCommand     string   // 入力されたコマンド
    CommandParts     []string // 分解されたコマンド部分
    DetectedIssues   []ValidationIssue // 検出された問題
    Suggestions      []SimilarityResult // 類似コマンド提案
    DeprecationInfo  *DeprecationInfo   // 廃止情報
    HelpURL          string             // ヘルプURL
}

// ValidationIssue は検証で発見された問題
type ValidationIssue struct {
    Type        IssueType        // 問題タイプ
    Severity    MessageSeverity  // 重要度
    Component   string           // 問題のあるコンポーネント
    Message     string           // 問題の説明
    Expected    []string         // 期待される値
}

// IssueType は問題タイプ
type IssueType int

const (
    IssueInvalidMainCommand IssueType = iota
    IssueInvalidSubCommand
    IssueDeprecatedCommand
    IssueSyntaxError
    IssueAmbiguousCommand
)

// ComprehensiveErrorFormatter は包括的エラーフォーマッター
type ComprehensiveErrorFormatter struct {
    messageGenerator   *ErrorMessageGenerator
    commandSuggester   *SimilarCommandSuggester
    deprecatedDetector *DeprecatedCommandDetector
    colorEnabled       bool
    language          string // "ja" または "en"
}

// NewComprehensiveErrorFormatter は新しいフォーマッターを作成
func NewComprehensiveErrorFormatter(
    msgGen *ErrorMessageGenerator,
    suggester *SimilarCommandSuggester, 
    detector *DeprecatedCommandDetector,
    colorEnabled bool,
    language string,
) *ComprehensiveErrorFormatter {
    return &ComprehensiveErrorFormatter{
        messageGenerator:   msgGen,
        commandSuggester:   suggester,
        deprecatedDetector: detector,
        colorEnabled:       colorEnabled,
        language:          language,
    }
}

// FormatError は包括的なエラーメッセージをフォーマット
func (f *ComprehensiveErrorFormatter) FormatError(context *ErrorContext) string {
    // 実装詳細
}
```

### エラー出力パターン

#### 1. 無効なメインコマンド
```
❌ エラー: 'invalidcommand' は有効なusacloudコマンドではありません

💡 もしかして以下のコマンドですか？
   • server (類似度: 67%)
   • service (類似度: 58%)
   
ℹ️  利用可能なコマンド一覧: usacloud --help
   詳細情報: https://docs.usacloud.jp/usacloud/commands/
```

#### 2. 無効なサブコマンド（類似提案あり）
```
❌ エラー: 'invalidaction' は server コマンドの有効なサブコマンドではありません

💡 もしかして以下のサブコマンドですか？
   • list (類似度: 45%)
   • create (類似度: 38%)
   
ℹ️  server で利用可能なサブコマンド:
   • list, read, create, update, delete
   • boot, shutdown, reset, send-nmi
   • monitor-cpu, ssh, vnc, rdp
   • wait-until-ready, wait-until-shutdown
   
詳細情報: usacloud server --help
```

#### 3. 廃止コマンド
```
⚠️  注意: 'iso-image' コマンドはv1で廃止されました

🔄 代わりに以下を使用してください:
   usacloud cdrom list

📋 移行方法:
   • iso-image list  → cdrom list
   • iso-image read  → cdrom read
   • iso-image create → cdrom create
   
詳細な移行ガイド: https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/
```

#### 4. 複合エラー（複数の問題）
```
❌ 複数の問題が検出されました: 'iso-imag lst'

⚠️  'iso-imag' は廃止されたコマンド 'iso-image' のtypoです
    → 'cdrom' を使用してください

❌ 'lst' は有効なサブコマンドではありません
    💡 もしかして 'list' ですか？

✅ 修正例:
   usacloud cdrom list
   
詳細情報: https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/
```

### フォーマット機能

#### 1. 視覚的要素の統一
```go
// VisualElements は視覚的要素定義
type VisualElements struct {
    ErrorIcon     string // ❌
    WarningIcon   string // ⚠️
    InfoIcon      string // ℹ️
    SuggestionIcon string // 💡
    SuccessIcon   string // ✅
    MigrationIcon string // 🔄
    ListIcon      string // 📋
}

// getVisualElements は言語に応じた視覚的要素を取得
func (f *ComprehensiveErrorFormatter) getVisualElements() *VisualElements {
    return &VisualElements{
        ErrorIcon:     "❌",
        WarningIcon:   "⚠️",
        InfoIcon:      "ℹ️",
        SuggestionIcon: "💡",
        SuccessIcon:   "✅", 
        MigrationIcon: "🔄",
        ListIcon:      "📋",
    }
}
```

#### 2. 多言語対応
```go
// Messages は多言語メッセージ
type Messages struct {
    InvalidCommand      string
    InvalidSubcommand   string
    DeprecatedCommand   string
    SuggestionsHeader   string
    AlternativesHeader  string
    MigrationHeader     string
    AvailableCommands   string
    SeeAlso            string
}

// getMessages は言語に応じたメッセージを取得
func (f *ComprehensiveErrorFormatter) getMessages() *Messages {
    if f.language == "en" {
        return &Messages{
            InvalidCommand:      "Error: '%s' is not a valid usacloud command",
            InvalidSubcommand:   "Error: '%s' is not a valid subcommand for %s command",
            DeprecatedCommand:   "Warning: '%s' command was deprecated in v1",
            SuggestionsHeader:   "Did you mean one of these?",
            AlternativesHeader:  "Use this instead:",
            MigrationHeader:     "Migration guide:",
            AvailableCommands:   "Available commands for %s:",
            SeeAlso:            "See also: %s",
        }
    }
    
    // デフォルト日本語
    return &Messages{
        InvalidCommand:      "エラー: '%s' は有効なusacloudコマンドではありません",
        InvalidSubcommand:   "エラー: '%s' は %s コマンドの有効なサブコマンドではありません",
        DeprecatedCommand:   "注意: '%s' コマンドはv1で廃止されました",
        SuggestionsHeader:   "もしかして以下のコマンドですか？",
        AlternativesHeader:  "代わりに以下を使用してください:",
        MigrationHeader:     "移行方法:",
        AvailableCommands:   "%s で利用可能なサブコマンド:",
        SeeAlso:            "詳細情報: %s",
    }
}
```

### 高度な機能

#### 1. コンテキスト分析
```go
// analyzeErrorContext はエラーコンテキストを分析
func (f *ComprehensiveErrorFormatter) analyzeErrorContext(context *ErrorContext) *ErrorAnalysis {
    analysis := &ErrorAnalysis{
        PrimaryIssue:    f.identifyPrimaryIssue(context.DetectedIssues),
        SecondaryIssues: f.identifySecondaryIssues(context.DetectedIssues),
        UserIntent:      f.inferUserIntent(context),
        RecommendedAction: f.recommendAction(context),
    }
    
    return analysis
}

// inferUserIntent はユーザーの意図を推測
func (f *ComprehensiveErrorFormatter) inferUserIntent(context *ErrorContext) UserIntent {
    // typo、学習、探索などの意図を分析
}
```

#### 2. 動的ヘルプ生成
```go
// generateDynamicHelp は動的ヘルプを生成
func (f *ComprehensiveErrorFormatter) generateDynamicHelp(context *ErrorContext) string {
    var helpSections []string
    
    // 基本的なヘルプ
    if len(context.CommandParts) == 1 {
        helpSections = append(helpSections, "usacloud --help")
    } else if len(context.CommandParts) >= 2 {
        helpSections = append(helpSections, 
            fmt.Sprintf("usacloud %s --help", context.CommandParts[1]))
    }
    
    // 関連ドキュメント
    if context.HelpURL != "" {
        helpSections = append(helpSections, context.HelpURL)
    }
    
    return strings.Join(helpSections, "\n   ")
}
```

## テスト戦略
- 統合テスト：全エラータイプが適切にフォーマットされることを確認
- 視覚テスト：カラー出力とプレーンテキスト出力が正しく動作することを確認
- 多言語テスト：日本語・英語両方で自然なメッセージが生成されることを確認
- レイアウトテスト：様々な端末幅での出力が読みやすいことを確認
- コンテキストテスト：異なるエラー組み合わせで適切な出力が生成されることを確認
- ユーザビリティテスト：実際のユーザーシナリオで有用性を検証

## 依存関係
- 前提PBI: PBI-008～012 (全検証・エラー生成コンポーネント)
- 関連PBI: PBI-014 (ユーザーフレンドリーヘルプシステム)

## 見積もり
- 開発工数: 6時間
  - 統合フォーマッター実装: 2.5時間
  - 視覚的要素とレイアウト: 1.5時間
  - 多言語対応実装: 1時間
  - ユニットテスト作成: 1時間

## 完了の定義 ✅ **完了 2025-01-09**
- [x] `internal/validation/comprehensive_error_formatter.go`ファイルが作成されている
- [x] `ComprehensiveErrorFormatter`構造体とフォーマット機能が実装されている
- [x] 全エラータイプが統一されたフォーマットで出力される
- [x] 視覚的階層とカラー出力が正しく実装されている
- [x] 日本語・英語両方の多言語対応が実装されている
- [x] 建設的な提案が適切に優先順位付けされている
- [x] 動的ヘルプ生成機能が実装されている
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] 統合テストが全エラーパターンで通過している
- [x] コードレビューが完了している

## 実装結果 📊

**実装ファイル:**
- `internal/validation/comprehensive_error_formatter.go` - ComprehensiveErrorFormatter構造体と統合フォーマット機能の完全実装
- `internal/validation/comprehensive_error_formatter_test.go` - 21テスト関数による包括的テスト

**実装内容:**
- 統合エラーコンテキスト分析システム（ErrorContext, ValidationIssue構造）
- 5つの問題タイプ分類（無効コマンド、サブコマンド、廃止、構文、曖昧）
- 4つのユーザー意図推測（タイポ、探索、移行、学習）
- 視覚的要素統一（7種類のアイコン：❌⚠️ℹ️💡✅🔄📋）
- 多言語対応（日本語・英語完全対応）
- 動的ヘルプ生成機能
- カラー出力制御機能
- 既存検証コンポーネント完全統合

**テスト結果:**
- 21のテスト関数すべて成功
- 検証結果フォーマット機能の検証
- サブコマンド結果フォーマット機能の検証
- エラーコンテキスト分析機能の検証
- 多言語出力の検証
- 廃止コマンド統合処理の検証
- ユーザー意図推測の検証
- エッジケース処理の検証

**技術的特徴:**
- 階層化されたエラー情報構造
- 重要度ベースの問題優先順位付け
- スコアベースの類似コマンド提案統合
- 適応的ヘルプ情報生成
- 完全な国際化フレームワーク
- カラーとプレーンテキスト双方対応
- 既存全検証コンポーネントとの完全統合
- 拡張可能なメッセージ・視覚要素管理

## 備考
- この機能はユーザーエクスペリエンスの最終的な品質を決定する重要コンポーネント
- 一貫性のあるデザインと分かりやすいメッセージが最も重要
- 国際化対応により将来的な言語追加が容易
- 視覚的要素の使用は端末環境への配慮が必要