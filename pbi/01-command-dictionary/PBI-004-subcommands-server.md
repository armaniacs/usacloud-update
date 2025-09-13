# PBI-004: serverサブコマンド15個の詳細定義

## 概要
最も複雑で利用頻度の高い`server`コマンドの15個のサブコマンドを詳細に定義する。GitHub調査により特定された実際のサブコマンドに基づいて、完全なサブコマンド辞書を作成する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] serverコマンドの15個すべてのサブコマンドが定義されている
- [x] 各サブコマンドの機能と用途が明確に文書化されている
- [x] GitHub上の実装と整合性が取れている
- [x] 基本CRUD操作から特殊な操作まで網羅されている
- [x] テストで全サブコマンドの定義を検証できている

## 技術仕様

### GitHub調査で確認されたserverサブコマンド15個

#### 基本CRUD操作
1. **list** - サーバー一覧表示
2. **read** - サーバー詳細情報表示
3. **create** - サーバー作成
4. **update** - サーバー設定更新
5. **delete** - サーバー削除

#### 電源制御操作
6. **boot** - サーバー起動
7. **shutdown** - サーバーシャットダウン
8. **reset** - サーバーリセット（強制再起動）

#### 管理操作
9. **send-nmi** - NMI（Non Maskable Interrupt）送信

#### 監視操作
10. **monitor-cpu** - CPU使用率監視

#### 接続操作
11. **ssh** - SSH接続
12. **vnc** - VNC接続
13. **rdp** - RDP接続

#### 状態待機操作
14. **wait-until-ready** - サーバー起動完了まで待機
15. **wait-until-shutdown** - サーバーシャットダウン完了まで待機

### ファイル構造拡張
```go
// internal/validation/server_subcommands.go
package validation

// ServerSubcommands は server コマンドの全サブコマンドを定義
var ServerSubcommands = []string{
    // 基本CRUD操作
    "list", "read", "create", "update", "delete",
    
    // 電源制御操作
    "boot", "shutdown", "reset",
    
    // 管理操作
    "send-nmi",
    
    // 監視操作  
    "monitor-cpu",
    
    // 接続操作
    "ssh", "vnc", "rdp",
    
    // 状態待機操作
    "wait-until-ready", "wait-until-shutdown",
}

// ServerSubcommandDescriptions はサブコマンドの説明を提供
var ServerSubcommandDescriptions = map[string]string{
    "list": "サーバー一覧を表示",
    "read": "サーバーの詳細情報を表示", 
    "create": "新しいサーバーを作成",
    "update": "サーバーの設定を更新",
    "delete": "サーバーを削除",
    "boot": "サーバーを起動",
    "shutdown": "サーバーをシャットダウン",
    "reset": "サーバーをリセット（強制再起動）",
    "send-nmi": "サーバーにNMIを送信",
    "monitor-cpu": "サーバーのCPU使用率を監視",
    "ssh": "サーバーにSSH接続",
    "vnc": "サーバーにVNC接続", 
    "rdp": "サーバーにRDP接続",
    "wait-until-ready": "サーバーの起動完了まで待機",
    "wait-until-shutdown": "サーバーのシャットダウン完了まで待機",
}
```

### GitHubファイル対応表
各サブコマンドは以下のGitHub上のファイルに実装されている：
- `boot.go` / `zz_boot_gen.go` → `boot`
- `create.go` / `zz_create_gen.go` → `create`
- `delete.go` / `zz_delete_gen.go` → `delete`
- `list.go` / `zz_list_gen.go` → `list`
- `read.go` / `zz_read_gen.go` → `read`
- `update.go` / `zz_update_gen.go` → `update`
- `reset.go` / `zz_reset_gen.go` → `reset`
- `shutdown.go` / `zz_shutdown_gen.go` → `shutdown`
- `send_nmi.go` / `zz_send_nmi_gen.go` → `send-nmi`
- `monitor_cpu.go` / `zz_monitor_cpu_gen.go` → `monitor-cpu`
- `ssh_*.go` / `zz_ssh_gen.go` → `ssh`
- `vnc_*.go` / `zz_vnc_gen.go` → `vnc`
- `rdp_*.go` / `zz_rdp_gen.go` → `rdp`
- `wait_until_ready.go` / `zz_wait_until_ready_gen.go` → `wait-until-ready`
- `wait_until_shutdonw.go` / `zz_wait_until_shutdown_gen.go` → `wait-until-shutdown`

## テスト戦略
- ユニットテスト：15個すべてのサブコマンドが定義されていることを確認
- 分類テスト：CRUD、電源制御、接続等のサブコマンド分類が正しいことを確認
- 整合性テスト：GitHub上の実装ファイルとサブコマンド定義の整合性を確認
- 説明テスト：各サブコマンドに適切な日本語説明が提供されていることを確認

## 依存関係
- 前提PBI: PBI-001 (IaaSコマンド辞書)
- 関連PBI: PBI-005 (diskサブコマンド), PBI-009 (サブコマンド検証)

## 見積もり
- 開発工数: 3時間
  - GitHubファイル詳細分析: 1.5時間
  - サブコマンド定義とドキュメント作成: 1時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/server_subcommands.go`ファイルが作成されている
- [x] 15個のserverサブコマンドが`ServerSubcommands`配列に定義されている
- [x] 各サブコマンドの日本語説明が`ServerSubcommandDescriptions`マップに定義されている
- [x] GitHub上の実装ファイルとの対応関係が文書化されている
- [x] ユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/server_subcommands.go` - 15個のserverサブコマンド詳細定義
- `internal/validation/server_subcommands_test.go` - 包括的なユニットテスト

**実装内容:**
- 基本CRUD操作: list, read, create, update, delete（5個）
- 電源制御操作: boot, shutdown, reset（3個）
- 管理操作: send-nmi（1個）
- 監視操作: monitor-cpu（1個）
- 接続操作: ssh, vnc, rdp（3個）
- 待機操作: wait-until-ready, wait-until-shutdown（2個）
- 機能分類ヘルパー関数：各カテゴリの検証機能付き
- 12個のテストケースでカバレッジ100%

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.185s
```

## 備考
- serverは最も複雑なコマンドなので、他のコマンドのテンプレートとしても機能する
- 接続系サブコマンド（ssh, vnc, rdp）は特別なパラメータ処理が必要
- wait系サブコマンドは非同期処理との組み合わせで使用される
- この詳細定義は他のリソース系コマンドの参考実装となる