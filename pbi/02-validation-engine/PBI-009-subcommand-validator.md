# PBI-009: サブコマンド存在検証

## 概要
有効なメインコマンドに対して、指定されたサブコマンドが実在するかどうかを検証する機能を実装する。各コマンドに特有のサブコマンド辞書と照合し、存在しないサブコマンドを検出する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 各メインコマンドのサブコマンド辞書に基づいて正確に検証される
- [x] 存在しないサブコマンドが適切に検出される
- [x] コマンドタイプ別（IaaS/Misc/Root）の異なる検証ロジックが実装されている
- [x] サブコマンド不要なコマンドの特別処理が実装されている
- [x] 詳細な検証結果と候補提案が提供される

## 技術仕様

### サブコマンド検証のパターン

#### 1. IaaSコマンドの基本CRUD操作
```bash
usacloud server list     ✓ 有効
usacloud server create   ✓ 有効
usacloud server invalid  ✗ 無効
usacloud disk connect    ✓ 有効（disk特有）
usacloud disk invalid    ✗ 無効
```

#### 2. その他コマンドの特殊サブコマンド
```bash
usacloud config show     ✓ 有効
usacloud config invalid  ✗ 無効
usacloud rest get        ✓ 有効
usacloud rest invalid    ✗ 無効
```

#### 3. ルートコマンドの特別処理
```bash
usacloud completion bash  ✓ 有効
usacloud version list     ✗ version は単体コマンド
usacloud update-self run  ✗ update-self は単体コマンド
```

### 実装構造
```go
// internal/validation/subcommand_validator.go
package validation

import (
    "fmt"
    "strings"
)

// SubcommandValidationResult はサブコマンド検証結果
type SubcommandValidationResult struct {
    IsValid      bool     // サブコマンドが有効かどうか
    MainCommand  string   // メインコマンド
    SubCommand   string   // 検証対象サブコマンド
    ErrorType    string   // エラータイプ
    Message      string   // 詳細メッセージ
    Suggestions  []string // 候補サブコマンド
    Available    []string // 利用可能なサブコマンド一覧
}

// SubcommandValidator はサブコマンド検証器
type SubcommandValidator struct {
    commandSubcommands map[string][]string // command -> subcommands のマッピング
    standaloneCommands map[string]bool     // サブコマンドを持たないコマンド
    mainValidator      *MainCommandValidator // メインコマンド検証器
}

// NewSubcommandValidator は新しい検証器を作成
func NewSubcommandValidator(mainValidator *MainCommandValidator) *SubcommandValidator {
    validator := &SubcommandValidator{
        commandSubcommands: make(map[string][]string),
        standaloneCommands: map[string]bool{
            "version":     true,
            "update-self": true,
        },
        mainValidator: mainValidator,
    }
    
    // サブコマンド辞書の初期化
    validator.initializeSubcommands()
    
    return validator
}

// Validate はサブコマンドを検証
func (v *SubcommandValidator) Validate(mainCommand, subCommand string) *SubcommandValidationResult {
    // 実装詳細
}

// IsValidSubcommand はサブコマンドが有効かを判定
func (v *SubcommandValidator) IsValidSubcommand(mainCommand, subCommand string) bool {
    // 実装詳細
}

// GetAvailableSubcommands は利用可能なサブコマンドを取得
func (v *SubcommandValidator) GetAvailableSubcommands(mainCommand string) []string {
    return v.commandSubcommands[mainCommand]
}

// GetSimilarSubcommands は類似サブコマンドを取得
func (v *SubcommandValidator) GetSimilarSubcommands(mainCommand, subCommand string) []string {
    // 実装詳細（PBI-012で詳細実装）
}
```

### 検証ロジック詳細

#### 1. サブコマンド辞書の初期化
```go
func (v *SubcommandValidator) initializeSubcommands() {
    // IaaSコマンドのサブコマンド
    v.commandSubcommands["server"] = ServerSubcommands
    v.commandSubcommands["disk"] = DiskSubcommands
    // ... その他44個のIaaSコマンド
    
    // その他コマンドのサブコマンド
    v.commandSubcommands["config"] = []string{"list", "show", "use", "create", "edit", "delete"}
    v.commandSubcommands["rest"] = []string{"get", "post", "put", "delete"}
    v.commandSubcommands["webaccelerator"] = []string{"list", "read", "create", "update", "delete", "purge"}
    
    // ルートコマンドのサブコマンド
    v.commandSubcommands["completion"] = []string{"bash", "zsh", "fish", "powershell"}
    // version, update-self は standaloneCommands で処理
}
```

#### 2. 検証フロー
1. メインコマンドが単体コマンドかチェック
2. 単体コマンドにサブコマンドが指定されている場合はエラー
3. 通常コマンドの場合、サブコマンド辞書と照合
4. 存在しない場合は類似サブコマンドを検索
5. 詳細な検証結果を返す

#### 3. エラータイプの分類
```go
const (
    ErrorTypeUnexpectedSubcommand = "unexpected_subcommand" // 単体コマンドにサブコマンド
    ErrorTypeInvalidSubcommand   = "invalid_subcommand"    // 存在しないサブコマンド
    ErrorTypeMissingSubcommand   = "missing_subcommand"    // サブコマンドが必要だが未指定
)
```

## テスト戦略
- 全サブコマンドテスト：定義された全サブコマンドが正しく検証されることを確認
- 無効サブコマンドテスト：存在しないサブコマンドが適切に検出されることを確認
- 単体コマンドテスト：version, update-selfへの不正なサブコマンド指定が検出されることを確認
- 類似性テスト：typoのあるサブコマンドに対して適切な候補が提案されることを確認
- 網羅性テスト：各コマンドの特殊なサブコマンド（connect, purge等）が正しく処理されることを確認
- 統合テスト：メインコマンド検証器との連携が正しく動作することを確認

## 依存関係
- 前提PBI: PBI-004 (serverサブコマンド), PBI-005 (diskサブコマンド), PBI-007 (コマンドパーサー), PBI-008 (メインコマンド検証)
- 関連PBI: PBI-012 (類似コマンド提案) - サブコマンド候補提案で使用

## 見積もり
- 開発工数: 5時間
  - 基本検証ロジック実装: 2時間
  - サブコマンド辞書統合: 1.5時間
  - 特別処理（単体コマンド等）実装: 1時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/subcommand_validator.go`ファイルが作成されている
- [x] `SubcommandValidator`構造体と検証メソッドが実装されている
- [x] 全メインコマンドのサブコマンド辞書が正しく統合されている
- [x] 存在するサブコマンドが正しく検証される
- [x] 存在しないサブコマンドが適切に検出される
- [x] 単体コマンドの特別処理が正しく動作する
- [x] 詳細な検証結果と候補提案が提供される
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] メインコマンド検証器との統合テストが通過している
- [x] コードレビューが完了している

## 実装結果 📊

**実装ファイル:**
- `internal/validation/subcommand_validator.go` - SubcommandValidator構造体とメソッド群の完全実装
- `internal/validation/subcommand_validator_test.go` - 19テスト関数による包括的テスト

**実装内容:**
- 50+コマンドのサブコマンド辞書統合（server, disk, config, rest等）
- 単体コマンド特別処理（version, update-self）
- 3つのエラータイプ分類（unexpected, invalid, missing subcommand）
- 類似サブコマンド提案機能（基本実装）
- 大文字小文字正規化処理
- メインコマンド検証器との完全統合
- 詳細なSubcommandValidationResult構造による情報提供

**テスト結果:**
- 19のテスト関数すべて成功
- 有効/無効サブコマンドの完全検証
- 単体コマンド制約の検証
- 欠損サブコマンドの検出
- 類似コマンド提案機能の検証
- メインコマンド検証器との統合テスト

**技術的特徴:**
- 効率的なマップベースサブコマンド検索
- 既存辞書（server, disk等）との完全統合
- エラータイプ分類による詳細診断
- 建設的なエラーメッセージと候補提案
- 将来のサブコマンド追加に対応した拡張可能設計

## 備考
- この検証器はPBI-008のメインコマンド検証器と密接に連携する
- 各コマンドの特性（CRUD操作、特殊操作）を理解した検証が必要
- エラーメッセージは建設的で、ユーザーが次の行動を取りやすいものにする
- 将来的なサブコマンド追加に対応できる柔軟な設計