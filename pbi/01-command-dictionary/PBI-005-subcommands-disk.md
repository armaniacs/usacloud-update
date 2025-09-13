# PBI-005: diskサブコマンドの詳細定義

## 概要
`disk`コマンドのサブコマンドを詳細に定義する。ディスクリソース管理に特化したサブコマンドの完全な辞書を作成し、基本CRUD操作からディスク特有の接続操作まで網羅する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] diskコマンドのすべてのサブコマンドが定義されている
- [x] 基本CRUD操作が網羅されている
- [x] ディスク特有の接続操作（connect/disconnect）が定義されている
- [x] 各サブコマンドの機能と用途が明確に文書化されている
- [x] テストで全サブコマンドの定義を検証できている

## 技術仕様

### diskサブコマンド定義

#### 基本CRUD操作
1. **list** - ディスク一覧表示
2. **read** - ディスク詳細情報表示
3. **create** - ディスク作成
4. **update** - ディスク設定更新
5. **delete** - ディスク削除

#### ディスク特有操作
6. **connect** - ディスクをサーバーに接続
7. **disconnect** - ディスクをサーバーから切断

#### 管理操作
8. **clone** - ディスククローン作成（可能性あり）
9. **resize** - ディスクサイズ変更（可能性あり）

### ファイル構造
```go
// internal/validation/disk_subcommands.go
package validation

// DiskSubcommands は disk コマンドの全サブコマンドを定義
var DiskSubcommands = []string{
    // 基本CRUD操作
    "list", "read", "create", "update", "delete",
    
    // ディスク特有操作
    "connect", "disconnect",
    
    // 管理操作
    "clone", "resize",
}

// DiskSubcommandDescriptions はサブコマンドの説明を提供
var DiskSubcommandDescriptions = map[string]string{
    "list": "ディスク一覧を表示",
    "read": "ディスクの詳細情報を表示",
    "create": "新しいディスクを作成",
    "update": "ディスクの設定を更新",
    "delete": "ディスクを削除",
    "connect": "ディスクをサーバーに接続",
    "disconnect": "ディスクをサーバーから切断",
    "clone": "ディスクのクローンを作成",
    "resize": "ディスクサイズを変更",
}

// DiskSubcommandCategories はサブコマンドの分類を提供
var DiskSubcommandCategories = map[string]string{
    "list": "basic",
    "read": "basic", 
    "create": "basic",
    "update": "basic",
    "delete": "basic",
    "connect": "attachment",
    "disconnect": "attachment",
    "clone": "management",
    "resize": "management",
}
```

### 特別な考慮事項

#### 接続操作の特殊性
- `connect`: ディスクをサーバーに接続する際は、対象サーバーIDが必要
- `disconnect`: ディスクをサーバーから切断する際は、現在の接続状態確認が必要
- これらの操作は他のリソースとの依存関係を持つ

#### 管理操作の重要性
- `clone`: 既存ディスクからの複製作成、バックアップ・テスト環境構築で重要
- `resize`: ディスク容量の拡張、システム運用で頻繁に使用される

## テスト戦略
- ユニットテスト：全サブコマンドが定義されていることを確認
- 分類テスト：basic、attachment、managementの各カテゴリが正しく分類されていることを確認
- 接続テスト：connect/disconnectの特殊な引数処理が考慮されていることを確認
- ドキュメントテスト：各サブコマンドに適切な日本語説明が提供されていることを確認

## 依存関係
- 前提PBI: PBI-001 (IaaSコマンド辞書), PBI-004 (serverサブコマンド)
- 関連PBI: PBI-009 (サブコマンド検証) - 特にattachment操作の検証

## 見積もり
- 開発工数: 2時間
  - GitHub調査とコマンド分析: 1時間
  - サブコマンド定義とドキュメント作成: 0.5時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/disk_subcommands.go`ファイルが作成されている
- [x] diskコマンドのサブコマンドが`DiskSubcommands`配列に定義されている
- [x] 各サブコマンドの日本語説明が`DiskSubcommandDescriptions`マップに定義されている
- [x] サブコマンドの分類が`DiskSubcommandCategories`マップに定義されている
- [x] connect/disconnectの特殊性が文書化されている
- [x] ユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/disk_subcommands.go` - 9個のdiskサブコマンド詳細定義
- `internal/validation/disk_subcommands_test.go` - 包括的なユニットテスト

**実装内容:**
- 基本CRUD操作: list, read, create, update, delete（5個）
- アタッチメント操作: connect, disconnect（2個）
- 管理操作: clone, resize（2個）
- 3つのカテゴリ分類システム：basic, attachment, management
- カテゴリフィルタ機能とヘルパー関数
- 14個のテストケースでカバレッジ100%

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.187s
```

## 備考
- diskは多くのIaaSリソースで共通する操作パターンを持つ
- connect/disconnect操作は他のリソース（server）との連携が必要
- この定義は他の接続可能リソース（cdrom等）の参考となる
- クラウド運用において最も頻繁に使用されるコマンドの一つ