# PBI-002: その他3コマンドの辞書作成

## 概要
IaaSサービス以外のusacloudユーティリティコマンド3個（config, rest, webaccelerator）の辞書データベースを作成する。これらのコマンドは特殊な目的を持つため、個別に定義する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 3個のその他コマンド（config, rest, webaccelerator）が`MiscCommands`マップに定義されている
- [x] 各コマンドの特殊なサブコマンドが適切に定義されている
- [x] IaaSコマンドとは異なる用途と構造が考慮されている
- [x] Go言語のmap形式でアクセス可能な構造になっている
- [x] テストで全3コマンドの定義を検証できている

## 技術仕様

### 対象となる3個のその他コマンド

#### 1. config - 設定管理
- **目的**: usacloudの設定ファイル管理
- **サブコマンド**: 
  - `list` - 設定一覧表示
  - `show` - 設定内容表示
  - `use` - 設定切り替え
  - `create` - 新規設定作成
  - `edit` - 設定編集
  - `delete` - 設定削除

#### 2. rest - REST API直接呼び出し
- **目的**: Sakura Cloud APIへの直接アクセス
- **サブコマンド**:
  - `get` - GET リクエスト
  - `post` - POST リクエスト
  - `put` - PUT リクエスト
  - `delete` - DELETE リクエスト

#### 3. webaccelerator - ウェブアクセラレーター
- **目的**: CDNサービスの管理
- **サブコマンド**:
  - `list` - ウェブアクセラレーター一覧
  - `read` - ウェブアクセラレーター詳細
  - `create` - ウェブアクセラレーター作成
  - `update` - ウェブアクセラレーター更新
  - `delete` - ウェブアクセラレーター削除
  - `purge` - キャッシュクリア

### ファイル構造
```go
// internal/validation/misc_commands.go
package validation

var MiscCommands = map[string][]string{
    "config": {"list", "show", "use", "create", "edit", "delete"},
    "rest": {"get", "post", "put", "delete"},
    "webaccelerator": {"list", "read", "create", "update", "delete", "purge"},
}
```

## テスト戦略
- ユニットテスト：3個すべてのコマンドが辞書に存在することを確認
- 機能テスト：各コマンドの特殊なサブコマンドが正しく定義されていることを確認
- 統合テスト：実際のusacloudコマンドとの整合性を確認

## 依存関係
- 前提PBI: PBI-001 (IaaSコマンド辞書)
- 関連PBI: PBI-003 (ルートコマンド辞書)

## 見積もり
- 開発工数: 2時間
  - GitHub調査とコマンド分析: 1時間
  - Go辞書データ構造作成: 0.5時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/misc_commands.go`ファイルが作成されている
- [x] 3個のその他コマンドが`MiscCommands`マップに定義されている
- [x] 各コマンドの特殊サブコマンドが適切に定義されている
- [x] ユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/misc_commands.go` - 3個のその他コマンド辞書
- `internal/validation/misc_commands_test.go` - 包括的なユニットテスト

**実装内容:**
- config: 6個のサブコマンド（設定管理）
- rest: 4個のHTTPメソッドサブコマンド（API直接アクセス）  
- webaccelerator: 6個のサブコマンド（CDN管理、purge含む）
- ヘルパー関数：IsValidMiscCommand, IsHTTPMethodSubcommand等
- 9個のテストケースでカバレッジ100%

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.183s
```

## 備考
- `pkg/resources.go`の`MiscResources`定義を参考にする
- configコマンドは設定管理に特化したサブコマンド構成
- restコマンドはHTTPメソッドベースのサブコマンド構成
- webacceleratorコマンドはCDN特有の`purge`サブコマンドを含む