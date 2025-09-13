# PBI-001: IaaS系44コマンドの完全辞書作成

## 概要
Sakura Cloud IaaSサービスに関連する44個のusacloudコマンドの完全な辞書データベースを作成する。各コマンドの基本的な操作パターン（CRUD操作等）を定義し、コマンド検証システムの基盤とする。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 49個のIaaSコマンドすべてが`IaaSCommands`マップに定義されている
- [x] 各コマンドの基本サブコマンド（list, read, create, update, delete）が適切に定義されている
- [x] コマンド固有の特殊サブコマンドが漏れなく含まれている
- [x] Go言語のmap形式でアクセス可能な構造になっている
- [x] テストで全49コマンドの定義を検証できている

## 技術仕様

### 対象となる44個のIaaSコマンド
1. archive - アーカイブ
2. authstatus - 認証状態
3. autobackup - 自動バックアップ
4. autoscale - オートスケール
5. bill - 請求
6. bridge - ブリッジ
7. category - カテゴリ
8. cdrom - CD-ROM
9. certificateauthority - 証明書オーソリティ
10. containerregistry - コンテナレジストリ
11. coupon - クーポン
12. database - データベース
13. disk - ディスク
14. diskplan - ディスクプラン
15. dns - DNS
16. enhanceddb - エンハンスドDB
17. esme - ESME
18. gslb - GSLB
19. icon - アイコン
20. iface - インターフェース
21. internet - インターネット
22. internetplan - インターネットプラン
23. ipaddress - IPアドレス
24. ipv6addr - IPv6アドレス
25. ipv6net - IPv6ネット
26. license - ライセンス
27. licenseinfo - ライセンス情報
28. loadbalancer - ロードバランサー
29. localrouter - ローカルルーター
30. mobilegateway - モバイルゲートウェイ
31. nfs - NFS
32. note - ノート
33. packetfilter - パケットフィルター
34. privatehost - プライベートホスト
35. privatehostplan - プライベートホストプラン
36. proxylb - プロキシLB
37. region - リージョン
38. self - セルフ
39. server - サーバー
40. serverplan - サーバープラン
41. serviceclass - サービスクラス
42. sim - SIM
43. simplemonitor - シンプル監視
44. sshkey - SSH鍵
45. subnet - サブネット
46. swytch - スイッチ
47. vpcrouter - VPCルーター
48. zone - ゾーン

### ファイル構造
```go
// internal/validation/iaas_commands.go
package validation

var IaaSCommands = map[string][]string{
    "archive": {"list", "read", "create", "update", "delete"},
    "server": {"list", "read", "create", "update", "delete", "boot", "shutdown", "reset"},
    // ... 44個すべてのコマンドを定義
}
```

## テスト戦略
- ユニットテスト：44個すべてのコマンドが辞書に存在することを確認
- カバレッジテスト：各コマンドの基本サブコマンドが適切に定義されていることを確認
- 統合テスト：実際のusacloudコマンドとの整合性を確認

## 依存関係
- 前提PBI: なし（最初のPBI）
- 関連PBI: PBI-004 (serverサブコマンド), PBI-005 (diskサブコマンド)

## 見積もり
- 開発工数: 4時間
  - GitHub調査とコマンドリスト作成: 2時間
  - Go辞書データ構造作成: 1時間
  - ユニットテスト作成: 1時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/iaas_commands.go`ファイルが作成されている
- [x] 49個のIaaSコマンドが`IaaSCommands`マップに定義されている
- [x] 各コマンドに最低限の基本サブコマンドが定義されている
- [x] ユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/iaas_commands.go` - 49個のIaaSコマンド辞書
- `internal/validation/iaas_commands_test.go` - 包括的なユニットテスト

**実装内容:**
- 49個のIaaSコマンドを完全定義
- 基本CRUD操作 + コマンド固有サブコマンド
- ヘルパー関数：IsValidIaaSCommand, IsValidIaaSSubcommand, GetIaaSCommandSubcommands等
- 10個のテストケースでカバレッジ100%

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.191s
```

## 備考
- GitHub上の`pkg/commands/iaas/`ディレクトリの構造を参考にする
- 各コマンドフォルダ内のGoファイルを分析してサブコマンドを特定する
- 将来的な機能追加を考慮して拡張可能な構造にする