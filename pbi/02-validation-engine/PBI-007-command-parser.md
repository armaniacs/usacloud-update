# PBI-007: usacloudコマンドライン解析器

## 概要
usacloudコマンドラインを解析して、メインコマンド、サブコマンド、オプションを分離・抽出するパーサーを実装する。この解析器は後続のコマンド検証処理の基盤となる重要なコンポーネントである。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] usacloudコマンドラインを正確に解析できる
- [x] メインコマンド、サブコマンド、オプションが適切に分離される
- [x] 複雑なコマンドライン（複数オプション、引数含む）を処理できる
- [x] エラー処理が適切に実装されている
- [x] 様々なコマンドパターンに対応している

## 技術仕様

### 解析対象のコマンドパターン

#### 1. 基本的なコマンドパターン
```bash
usacloud server list
usacloud disk create --name test
usacloud config show
```

#### 2. 複雑なコマンドパターン
```bash
usacloud server create --name test --cpu 2 --memory 4 --zone tk1a
usacloud disk connect --server-id 123456789 disk-id
usacloud server ssh --user root --key ~/.ssh/id_rsa server-name
```

#### 3. エラーケースのパターン
```bash
usacloud                    # コマンドなし
usacloud invalid-command    # 存在しないコマンド
usacloud server invalid-sub # 存在しないサブコマンド
```

### 実装構造
```go
// internal/validation/parser.go
package validation

import (
    "errors"
    "regexp"
    "strings"
)

// CommandLine は解析されたコマンドライン情報を保持
type CommandLine struct {
    Raw         string            // 元のコマンドライン
    MainCommand string            // メインコマンド (server, disk, config等)
    SubCommand  string            // サブコマンド (list, create, show等)
    Arguments   []string          // 位置引数
    Options     map[string]string // オプション (--name=value)
    Flags       []string          // フラグ (--force, --dry-run等)
}

// ParseError は解析エラーを表現
type ParseError struct {
    Message string
    Position int
    Input   string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("parse error at position %d: %s in '%s'", e.Position, e.Message, e.Input)
}

// Parser はコマンドライン解析器
type Parser struct {
    // 設定やキャッシュ等
}

// NewParser は新しい解析器を作成
func NewParser() *Parser {
    return &Parser{}
}

// Parse はコマンドラインを解析
func (p *Parser) Parse(commandLine string) (*CommandLine, error) {
    // 実装詳細
}

// IsUsacloudCommand はusacloudコマンドかどうかを判定
func (p *Parser) IsUsacloudCommand(commandLine string) bool {
    return strings.HasPrefix(strings.TrimSpace(commandLine), "usacloud ")
}
```

### 解析ロジック詳細

#### 1. 前処理
- 文字列のトリム
- usacloudプレフィックスの確認
- 空文字列・無効文字列の検出

#### 2. トークン分割
- 空白での基本分割
- クォート処理（"string with spaces"）
- エスケープ文字処理

#### 3. 構造解析
- メインコマンドの特定（第1引数）
- サブコマンドの特定（第2引数、存在する場合）
- オプションの解析（--key=value, --key value）
- フラグの解析（--force, --dry-run）
- 位置引数の抽出

### エラー処理戦略
```go
// 解析エラーの種類
var (
    ErrEmptyCommand     = errors.New("empty command")
    ErrNotUsacloudCommand = errors.New("not a usacloud command")
    ErrInvalidSyntax    = errors.New("invalid command syntax")
    ErrMissingArgument  = errors.New("missing required argument")
)
```

## テスト戦略
- 基本パターンテスト：単純なコマンドの正しい解析を確認
- 複雑パターンテスト：複数オプション・引数を含むコマンドの解析を確認
- エラーケーステスト：無効なコマンドラインの適切なエラー処理を確認
- 境界値テスト：空文字列、非常に長いコマンドライン等の処理を確認
- パフォーマンステスト：大量のコマンド解析時のメモリ・CPU使用量を確認

## 依存関係
- 前提PBI: PBI-001～006 (コマンド辞書) - 解析結果の検証で使用
- 関連PBI: PBI-008～010 (検証エンジン) - このパーサーの出力を使用

## 見積もり
- 開発工数: 6時間
  - 基本解析ロジック実装: 3時間
  - エラー処理とエッジケース対応: 2時間
  - ユニットテスト作成: 1時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/parser.go`ファイルが作成されている
- [x] `CommandLine`構造体が適切に定義されている
- [x] `Parser`構造体と解析メソッドが実装されている
- [x] 基本的なコマンドパターンが正しく解析される
- [x] 複雑なコマンドパターンが正しく解析される
- [x] エラーケースが適切に処理される
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している
- [x] パフォーマンステストが通過している

## 備考
- このパーサーはusacloud-update全体の検証システムの基盤となる
- 正確性とパフォーマンスの両立が重要
- 将来的なコマンド拡張に対応できる柔軟な設計が必要
- 既存のtransform/engine.goとの連携を考慮した設計