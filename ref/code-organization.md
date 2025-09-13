# Code Organization Reference

このドキュメントでは、usacloud-updateプロジェクトのファイル・ディレクトリ構造、命名規則、およびコーディング規約について詳しく説明します。

## プロジェクト構造概要

```
usacloud-update/
├── cmd/                        # アプリケーションエントリーポイント
│   └── usacloud-update/       # メインアプリケーション
│       ├── main.go            # CLIエントリーポイント
│       └── main_test.go       # CLIテスト
├── internal/                   # 内部ライブラリ（非公開パッケージ）
│   ├── bdd/                   # BDD（行動駆動開発）テスト
│   ├── config/                # 設定管理システム
│   ├── sandbox/               # サンドボックス実行環境
│   ├── scanner/               # ファイルスキャン機能
│   ├── security/              # セキュリティコンポーネント
│   ├── testing/               # テスト支援ツール
│   ├── transform/             # 変換エンジン
│   ├── tui/                   # ターミナルUI
│   └── validation/            # 検証システム
├── testdata/                   # テストデータ
├── features/                   # BDD機能仕様
├── ref/                       # リファレンスドキュメント
├── pbi/                       # プロダクトバックログアイテム
├── tests/                     # 統合テスト
└── bin/                       # ビルド成果物
```

## ディレクトリ構造詳細

### `/cmd` - アプリケーションディレクトリ

**目的**: アプリケーションのエントリーポイントを格納

```
cmd/
└── usacloud-update/           # メインアプリケーション
    ├── main.go                # CLIインターフェースとメイン処理
    └── main_test.go           # メイン処理のテスト
```

**命名規則**:
- ディレクトリ名：実行可能ファイル名と同一
- メインファイル：`main.go`
- テストファイル：`main_test.go`

### `/internal` - 内部ライブラリディレクトリ

**目的**: プロジェクト専用の内部パッケージを格納（外部からimport不可）

#### `/internal/transform` - 変換エンジン
```
transform/
├── engine.go                  # 変換エンジン本体
├── engine_test.go             # エンジンテスト
├── engine_edge_test.go        # エッジケーステスト
├── engine_error_test.go       # エラーケーステスト
├── ruledefs.go                # ルール定義
├── ruledefs_test.go           # ルール定義テスト
├── rules.go                   # ルール実装
├── rules_test.go              # ルール実装テスト
├── rules_detailed_test.go     # 詳細ルールテスト
├── integrated_engine.go       # 統合エンジン
├── performance_test.go        # パフォーマンステスト
└── boundary_test.go           # 境界値テスト
```

#### `/internal/validation` - 検証システム
```
validation/
├── parser.go                  # コマンドライン解析
├── parser_test.go             # 解析テスト
├── main_command_validator.go  # メインコマンド検証
├── main_command_validator_test.go
├── subcommand_validator.go    # サブコマンド検証
├── subcommand_validator_test.go
├── similar_command_suggester.go # 類似コマンド提案
├── similar_command_suggester_test.go
├── [各種]_commands.go         # コマンド別検証ロジック
├── [各種]_commands_test.go    # 対応テスト
├── error_message_generator.go # エラーメッセージ生成
├── user_friendly_help_system.go # ヘルプシステム
└── testdata/                  # テストデータ
    ├── commands/              # コマンドデータ
    └── errors/                # エラーパターン
```

#### `/internal/config` - 設定管理
```
config/
├── file.go                    # ファイルベース設定
├── file_test.go
├── env.go                     # 環境変数設定
├── env_test.go
├── interactive.go             # 対話的設定
├── interactive_test.go
├── integrated_config.go       # 統合設定管理
├── integrated_config_test.go
├── config_migrator.go         # 設定移行
├── interactive_config.go      # 対話的設定作成
├── profile_manager.go         # プロファイル管理
└── profile/                   # プロファイル詳細実装
    ├── types.go               # データ型定義
    ├── manager.go             # プロファイルマネージャー
    ├── manager_test.go
    ├── manager_race_test.go   # 競合条件テスト
    ├── storage.go             # ストレージ実装
    ├── storage_test.go
    ├── storage_atomic_test.go # アトミック操作テスト
    ├── template.go            # テンプレート機能
    ├── template_test.go
    ├── commands.go            # コマンド実装
    └── commands_test.go
```

#### `/internal/tui` - ターミナルUI
```
tui/
├── app.go                     # メインアプリケーション
├── app_test.go
├── app_edge_test.go           # エッジケース
├── file_selector.go           # ファイル選択
├── file_selector_test.go
├── filter/                    # フィルタリングシステム
│   ├── types.go               # 型定義
│   ├── system.go              # システム本体
│   ├── panel.go               # UIパネル
│   ├── preset.go              # プリセット管理
│   ├── text_filter.go         # テキストフィルタ
│   ├── category_filter.go     # カテゴリフィルタ
│   ├── status_filter.go       # ステータスフィルタ
│   └── [各種]_test.go         # 対応テスト
└── preview/                   # プレビュー機能
    ├── types.go               # 型定義
    ├── generator.go           # プレビュー生成
    ├── widget.go              # UIウィジェット
    └── [各種]_test.go         # 対応テスト
```

### `/testdata` - テストデータディレクトリ

**構造**:
```
testdata/
├── sample_v0_v1_mixed.sh      # 混合版サンプル
├── expected_v1_1.sh           # 期待される出力
├── mixed_with_non_usacloud.sh # 非usacloud行混在
├── expected_mixed_non_usacloud.sh # 期待される混在出力
└── to-be-fail01.sh            # 失敗テスト用
```

### `/features` - BDD仕様ディレクトリ

**構造**:
```
features/
└── sandbox.feature            # サンドボックス機能仕様
```

## 命名規則

### ファイル命名規則

#### Go ソースファイル
- **通常ファイル**: `snake_case.go`
  - 例: `main_command_validator.go`, `error_message_generator.go`
- **テストファイル**: `[対象ファイル名]_test.go`
  - 例: `parser_test.go`, `engine_test.go`
- **特殊テスト**: `[対象]_[種類]_test.go`
  - 例: `engine_edge_test.go`, `manager_race_test.go`

#### 設定・データファイル
- **設定ファイル**: `.conf`, `.ini`
- **テストデータ**: `[目的]_[内容].sh`, `[目的].json`
- **ドキュメント**: `[内容]-[詳細].md`

### パッケージ命名規則

#### パッケージ名
- **単語**: 小文字単数形
  - ✅ `transform`, `validation`, `config`
  - ❌ `transforms`, `validations`, `configs`

#### サブパッケージ
- **機能別**: 親パッケージ + 機能名
  - 例: `config/profile`, `tui/filter`, `tui/preview`

### 型・関数命名規則

#### 型名（Type Names）
- **PascalCase**: `CommandLine`, `ValidationResult`
- **Interface**: 動詞 + er形式推奨
  - 例: `Rule`, `Parser`, `Validator`

#### 関数名（Function Names）
- **Public**: PascalCase
  - 例: `NewDefaultEngine()`, `ParseCommand()`
- **Private**: camelCase
  - 例: `validateMainCommand()`, `generateSuggestions()`

#### 定数名（Constant Names）
- **Public**: PascalCase
  - 例: `DefaultTimeout`, `MaxRetryCount`
- **Private**: camelCase
- **列挙型**: 型名プレフィックス + 値
  - 例: `IssueParseError`, `IssueInvalidMainCommand`

### 変数命名規則

#### グローバル変数
- **Package Level**: camelCase（private）または PascalCase（public）
- **Error Variables**: `Err` プレフィックス
  - 例: `ErrEmptyCommand`, `ErrNotUsacloudCommand`

#### ローカル変数
- **短縮形推奨**: `cmd`, `cfg`, `res`, `err`
- **明確性重視**: 複雑な処理では完全な名前

## コーディング規約

### インポート順序

```go
package main

import (
    // 1. 標準ライブラリ
    "bufio"
    "flag"
    "fmt"
    "os"
    
    // 2. 外部ライブラリ
    "github.com/fatih/color"
    "github.com/rivo/tview"
    
    // 3. 内部パッケージ
    "github.com/armaniacs/usacloud-update/internal/config"
    "github.com/armaniacs/usacloud-update/internal/transform"
)
```

### エラーハンドリング

#### エラー定義
```go
// パッケージレベルエラー
var (
    ErrEmptyCommand       = errors.New("empty command")
    ErrNotUsacloudCommand = errors.New("not a usacloud command")
)

// カスタムエラー型
type ParseError struct {
    Message  string
    Position int
    Input    string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse error at position %d: %s", e.Position, e.Message)
}
```

#### エラー処理パターン
```go
// 基本パターン
result, err := someOperation()
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}

// 複数戻り値での成功・失敗判定
line, changed, before, after := rule.Apply(input)
if !changed {
    return Result{Line: line, Changed: false}
}
```

### テスト構造

#### テストファイル構造
```go
package transform

import (
    "testing"
)

func TestEngineApply(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid transformation",
            input:    "usacloud server list --output-type=csv",
            expected: "usacloud server list --output-type=json",
            wantErr:  false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := NewDefaultEngine()
            result := engine.Apply(tt.input)
            
            if result.Line != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result.Line)
            }
        })
    }
}
```

#### テストヘルパー関数
```go
// testing_utils.go
func setupTestEngine() *Engine {
    return NewDefaultEngine()
}

func assertTransform(t *testing.T, engine *Engine, input, expected string) {
    t.Helper()
    result := engine.Apply(input)
    if result.Line != expected {
        t.Errorf("Transform failed: expected %q, got %q", expected, result.Line)
    }
}
```

### ドキュメンテーション

#### パッケージコメント
```go
// Package validation provides command validation functionality for usacloud-update.
//
// This package includes parsers for usacloud command lines, validators for
// command syntax, and suggestion generators for error recovery.
package validation
```

#### 関数コメント
```go
// ParseCommand parses a usacloud command line and returns structured information.
// It returns an error if the command is not a valid usacloud command or if
// parsing fails due to syntax errors.
//
// Example:
//   cmd, err := ParseCommand("usacloud server list --zone=tk1v")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Main command: %s\n", cmd.MainCommand)
func ParseCommand(line string) (*CommandLine, error) {
    // implementation
}
```

## ビルドとデプロイメント

### Makefile ターゲット
```makefile
# 主要ターゲット
build:      # バイナリビルド
test:       # テスト実行
bdd:        # BDDテスト
golden:     # ゴールデンファイル更新
install:    # インストール
clean:      # クリーンアップ
```

### バージョン管理
- **セマンティックバージョニング**: `major.minor.patch`
- **開発版識別**: 奇数マイナーバージョン（1.9.0）
- **リリース版**: 偶数マイナーバージョン（2.0.0）

---

**最終更新**: 2025年1月
**バージョン**: v1.9.0対応
**メンテナー**: 開発チーム