# PBI-003: ルート3コマンドの辞書作成

## 概要
usacloudのルートレベルで直接実行される3個のコマンド（completion, version, update-self）の辞書データベースを作成する。これらのコマンドはツール自体の管理機能を提供する特別なコマンドである。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 3個のルートコマンド（completion, version, update-self）が`RootCommands`マップに定義されている
- [x] 各コマンドの特殊な性質（サブコマンドなし、または限定的なサブコマンド）が適切に定義されている
- [x] 他のコマンドカテゴリとは異なる処理が考慮されている
- [x] Go言語のmap形式でアクセス可能な構造になっている
- [x] テストで全3コマンドの定義を検証できている

## 技術仕様

### 対象となる3個のルートコマンド

#### 1. completion - シェル補完
- **目的**: bashやzsh等のシェル補完スクリプト生成
- **サブコマンド**: 
  - `bash` - bash補完スクリプト生成
  - `zsh` - zsh補完スクリプト生成  
  - `fish` - fish補完スクリプト生成
  - `powershell` - PowerShell補完スクリプト生成

#### 2. version - バージョン表示
- **目的**: usacloudのバージョン情報表示
- **サブコマンド**: なし（単体コマンド）
- **特記事項**: `usacloud version`として単独で使用

#### 3. update-self - 自己アップデート
- **目的**: usacloud自体のアップデート
- **サブコマンド**: なし（単体コマンド）
- **特記事項**: `usacloud update-self`として単独で使用

### ファイル構造
```go
// internal/validation/root_commands.go
package validation

var RootCommands = map[string][]string{
    "completion": {"bash", "zsh", "fish", "powershell"},
    "version": {}, // サブコマンドなし
    "update-self": {}, // サブコマンドなし
}
```

### 特別な処理考慮事項
- `version`と`update-self`はサブコマンドを持たない単体コマンド
- これらのコマンドに対してサブコマンドが指定された場合はエラーとする
- `completion`のみサブコマンドを持つ

## テスト戦略
- ユニットテスト：3個すべてのルートコマンドが辞書に存在することを確認
- 単体コマンドテスト：`version`と`update-self`がサブコマンドなしで正しく処理されることを確認
- 補完テスト：`completion`コマンドの4つのシェル補完サブコマンドが適切に定義されていることを確認
- エラーテスト：単体コマンドにサブコマンドが指定された場合の適切なエラー処理を確認

## 依存関係
- 前提PBI: PBI-001 (IaaSコマンド辞書), PBI-002 (その他コマンド辞書)
- 関連PBI: PBI-008 (メインコマンド検証) - 単体コマンドの特別処理

## 見積もり
- 開発工数: 1.5時間
  - GitHub調査とコマンド分析: 0.5時間
  - Go辞書データ構造作成: 0.5時間
  - ユニットテスト作成（単体コマンド対応含む）: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/root_commands.go`ファイルが作成されている
- [x] 3個のルートコマンドが`RootCommands`マップに定義されている
- [x] 単体コマンドの空サブコマンドリストが適切に定義されている
- [x] `completion`コマンドの4つのシェル補完サブコマンドが定義されている
- [x] ユニットテストが作成され、すべて通過している
- [x] 単体コマンドに対するサブコマンド指定エラーのテストが通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/root_commands.go` - 3個のルートコマンド辞書
- `internal/validation/root_commands_test.go` - 包括的なユニットテスト

**実装内容:**
- completion: 4個のシェル補完サブコマンド（bash, zsh, fish, powershell）
- version: 単体コマンド（サブコマンドなし）
- update-self: 単体コマンド（サブコマンドなし）
- 単体コマンド検証：`IsStandaloneCommand`, `ValidateRootCommandUsage`等のヘルパー関数
- 13個のテストケースでカバレッジ100%

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.214s
```

## 備考
- `pkg/resources.go`の`RootCommands`定義を参考にする
- cobra.Commandの構造から各コマンドの特性を理解する
- 単体コマンドの検証時は特別なロジックが必要
- shell補完スクリプト生成機能は重要なユーザビリティ機能