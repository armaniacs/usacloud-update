# PBI-011: 日本語エラーメッセージ生成器

## 概要
コマンド検証エラーに対して、ユーザーフレンドリーで建設的な日本語エラーメッセージを自動生成する機能を実装する。技術的なエラーではなく、ユーザーが次に何をすべきかを明確に示すメッセージを提供する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 全ての検証エラータイプに対応したメッセージテンプレートが定義されている
- [x] 一貫性のある分かりやすい日本語メッセージが生成される
- [x] 建設的な提案やヒントが含まれている
- [x] エラーの重要度に応じて適切な表現が使い分けられている
- [x] カスタマイズ可能なメッセージフォーマットが実装されている

## 技術仕様

### エラーメッセージのカテゴリ

#### 1. メインコマンドエラー
```
存在しないコマンド: "usacloud invalidcommand list"
→ "エラー: 'invalidcommand' は有効なusacloudコマンドではありません。
   利用可能なコマンドを確認するには 'usacloud --help' を実行してください。"
```

#### 2. サブコマンドエラー  
```
存在しないサブコマンド: "usacloud server invalidaction"
→ "エラー: 'invalidaction' は server コマンドの有効なサブコマンドではありません。
   server コマンドで利用可能なサブコマンド: list, read, create, update, delete, boot, shutdown..."
```

#### 3. 廃止コマンドエラー
```
廃止されたコマンド: "usacloud iso-image list"
→ "注意: 'iso-image' コマンドはv1で廃止されました。
   代わりに 'usacloud cdrom list' を使用してください。"
```

#### 4. 構文エラー
```
不正な構文: "usacloud version list"
→ "エラー: 'version' コマンドはサブコマンドを受け付けません。
   正しい使用法: usacloud version"
```

### 実装構造
```go
// internal/validation/error_message_generator.go
package validation

import (
    "fmt"
    "strings"
)

// MessageTemplate はメッセージテンプレート
type MessageTemplate struct {
    Template    string            // メッセージテンプレート
    Severity    MessageSeverity   // メッセージの重要度
    Type        MessageType       // メッセージタイプ
    Suggestions bool              // 提案を含むかどうか
}

// MessageSeverity はメッセージの重要度
type MessageSeverity int

const (
    SeverityError   MessageSeverity = iota // エラー（赤）
    SeverityWarning                        // 警告（黄）
    SeverityInfo                           // 情報（青）
    SeveritySuccess                        // 成功（緑）
)

// MessageType はメッセージタイプ
type MessageType int

const (
    TypeInvalidCommand MessageType = iota
    TypeInvalidSubcommand
    TypeDeprecatedCommand
    TypeSyntaxError
    TypeMissingCommand
    TypeSuggestion
)

// ErrorMessageGenerator はエラーメッセージ生成器
type ErrorMessageGenerator struct {
    templates map[MessageType]*MessageTemplate
    colorEnabled bool
}

// NewErrorMessageGenerator は新しい生成器を作成
func NewErrorMessageGenerator(colorEnabled bool) *ErrorMessageGenerator {
    generator := &ErrorMessageGenerator{
        templates: make(map[MessageType]*MessageTemplate),
        colorEnabled: colorEnabled,
    }
    
    // メッセージテンプレートの初期化
    generator.initializeTemplates()
    
    return generator
}

// GenerateMessage はエラーメッセージを生成
func (g *ErrorMessageGenerator) GenerateMessage(msgType MessageType, params map[string]interface{}) string {
    // 実装詳細
}
```

### メッセージテンプレートの定義
```go
func (g *ErrorMessageGenerator) initializeTemplates() {
    g.templates[TypeInvalidCommand] = &MessageTemplate{
        Template: "エラー: '%s' は有効なusacloudコマンドではありません。\n" +
                  "利用可能なコマンドを確認するには 'usacloud --help' を実行してください。",
        Severity: SeverityError,
        Type:     TypeInvalidCommand,
        Suggestions: true,
    }
    
    g.templates[TypeInvalidSubcommand] = &MessageTemplate{
        Template: "エラー: '%s' は %s コマンドの有効なサブコマンドではありません。\n" +
                  "%s コマンドで利用可能なサブコマンド: %s",
        Severity: SeverityError,
        Type:     TypeInvalidSubcommand,
        Suggestions: true,
    }
    
    g.templates[TypeDeprecatedCommand] = &MessageTemplate{
        Template: "注意: '%s' コマンドはv1で廃止されました。\n" +
                  "代わりに '%s' を使用してください。",
        Severity: SeverityWarning,
        Type:     TypeDeprecatedCommand,
        Suggestions: true,
    }
    
    g.templates[TypeSyntaxError] = &MessageTemplate{
        Template: "エラー: '%s' コマンドはサブコマンドを受け付けません。\n" +
                  "正しい使用法: usacloud %s",
        Severity: SeverityError,
        Type:     TypeSyntaxError,
        Suggestions: true,
    }
    
    // ... その他のテンプレート
}
```

### カラー出力対応
```go
// ColorCode はカラーコード
type ColorCode string

const (
    ColorRed    ColorCode = "\033[31m"
    ColorYellow ColorCode = "\033[33m"
    ColorBlue   ColorCode = "\033[34m"
    ColorGreen  ColorCode = "\033[32m"
    ColorReset  ColorCode = "\033[0m"
)

// colorizeMessage はメッセージに色を付ける
func (g *ErrorMessageGenerator) colorizeMessage(message string, severity MessageSeverity) string {
    if !g.colorEnabled {
        return message
    }
    
    var color ColorCode
    switch severity {
    case SeverityError:
        color = ColorRed
    case SeverityWarning:
        color = ColorYellow
    case SeverityInfo:
        color = ColorBlue
    case SeveritySuccess:
        color = ColorGreen
    default:
        return message
    }
    
    return fmt.Sprintf("%s%s%s", color, message, ColorReset)
}
```

### メッセージ生成パラメータ
```go
// 使用例
params := map[string]interface{}{
    "command": "invalidcommand",
    "mainCommand": "server", 
    "subCommand": "invalidaction",
    "availableSubcommands": []string{"list", "read", "create"},
    "suggestions": []string{"server", "service"},
    "replacementCommand": "cdrom",
}

message := generator.GenerateMessage(TypeInvalidCommand, params)
```

## テスト戦略
- テンプレートテスト：全メッセージテンプレートが適切に定義されていることを確認
- パラメータテスト：様々なパラメータ組み合わせでメッセージが正しく生成されることを確認
- カラーテスト：カラー出力の有効/無効が正しく動作することを確認
- 日本語テスト：日本語メッセージが自然で分かりやすいことを確認
- 重要度テスト：エラー重要度に応じて適切な色分けがされることを確認
- 統合テスト：検証エンジンからの実際のエラーデータでメッセージが生成されることを確認

## 依存関係
- 前提PBI: PBI-008 (メインコマンド検証), PBI-009 (サブコマンド検証), PBI-010 (廃止コマンド検出)
- 関連PBI: PBI-012 (類似コマンド提案) - 候補提案をメッセージに含める

## 見積もり
- 開発工数: 4時間
  - メッセージテンプレート設計・実装: 2時間
  - カラー出力とフォーマット機能: 1時間
  - ユニットテスト作成: 1時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/error_message_generator.go`ファイルが作成されている
- [x] `ErrorMessageGenerator`構造体とメッセージ生成メソッドが実装されている
- [x] 全てのエラータイプに対応したテンプレートが定義されている
- [x] カラー出力機能が実装されている
- [x] パラメータ化されたメッセージ生成が正しく動作する
- [x] 自然で分かりやすい日本語メッセージが生成される
- [x] 建設的な提案やヒントが適切に含まれている
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] 実際の検証エラーでのテストが通過している
- [x] コードレビューが完了している

## 実装結果 📊

**実装ファイル:**
- `internal/validation/error_message_generator.go` - ErrorMessageGenerator構造体とメッセージ生成システムの完全実装
- `internal/validation/error_message_generator_test.go` - 25テスト関数による包括的テスト

**実装内容:**
- 9個のメッセージタイプ対応（無効コマンド、サブコマンドエラー、廃止コマンド等）
- 4つの重要度レベル（エラー、警告、情報、成功）
- カラー出力機能（赤、黄、青、緑）
- パラメータ化されたメッセージテンプレートシステム
- 提案機能付きメッセージ生成
- 検証結果からの自動メッセージ生成
- 自然な日本語メッセージと建設的な提案

**テスト結果:**
- 25のテスト関数すべて成功
- 全メッセージタイプの生成確認
- カラー出力機能の検証
- パラメータ置換の検証
- 日本語メッセージ品質の確認
- 実際の検証エラーとの統合テスト

**技術的特徴:**
- テンプレートベースの柔軟なメッセージ生成
- 重要度に応じたカラー出力制御
- 検証結果との完全統合
- ユーザビリティ重視の日本語メッセージ
- 拡張可能なテンプレート管理システム

## 備考
- ユーザーエクスペリエンスに直接影響する重要なコンポーネント
- 技術的なエラーメッセージではなく、ユーザーが理解しやすいメッセージを重視
- 一貫性のある表現とトーンを維持することが重要
- CLIツールの標準的なカラー出力規則に従う