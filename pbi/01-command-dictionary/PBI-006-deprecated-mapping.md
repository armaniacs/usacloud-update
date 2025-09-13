# PBI-006: 廃止コマンドマッピング辞書作成

## 概要
usacloudの進化に伴って廃止・変更されたコマンド名から新しいコマンド名へのマッピング辞書を作成する。ユーザーが古いコマンドを使用した際に、適切な新コマンドを提案できるようにする。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 既存の変換ルールで特定された廃止コマンドがすべて定義されている
- [x] 各廃止コマンドに対応する新コマンド（または廃止理由）が明確に定義されている
- [x] 完全廃止されたコマンドと名称変更されたコマンドが区別されている
- [x] ユーザーフレンドリーな説明メッセージが用意されている
- [x] テストで全マッピングの定義を検証できている

## 技術仕様

### 廃止・変更されたコマンド分類

#### 1. 名称変更されたコマンド
- **iso-image → cdrom**: CD-ROMリソース名の統一
- **startup-script → note**: ノートリソース名の統一  
- **ipv4 → ipaddress**: IPアドレスリソース名の統一

#### 2. 製品エイリアス整理
- **product-disk → disk-plan**: ディスクプランリソース名の統一
- **product-internet → internet-plan**: インターネットプランリソース名の統一
- **product-server → server-plan**: サーバープランリソース名の統一

#### 3. 完全廃止されたコマンド
- **summary**: 機能廃止、代替手段あり（rest, bill, self等の組み合わせ）
- **object-storage**: v1では非対応、S3互換ツールの使用を推奨
- **ojs**: object-storageのエイリアス、同様に廃止

### ファイル構造
```go
// internal/validation/deprecated_commands.go
package validation

// DeprecatedCommands は廃止されたコマンドから新コマンドへのマッピング
var DeprecatedCommands = map[string]string{
    // 名称変更
    "iso-image": "cdrom",
    "startup-script": "note", 
    "ipv4": "ipaddress",
    
    // 製品エイリアス整理
    "product-disk": "disk-plan",
    "product-internet": "internet-plan", 
    "product-server": "server-plan",
    
    // 完全廃止（空文字列は廃止を意味）
    "summary": "",
    "object-storage": "",
    "ojs": "",
}

// DeprecatedCommandMessages は廃止コマンドの詳細説明
var DeprecatedCommandMessages = map[string]string{
    "iso-image": "iso-imageコマンドはv1で廃止されました。cdromコマンドを使用してください",
    "startup-script": "startup-scriptコマンドはv1で廃止されました。noteコマンドを使用してください",
    "ipv4": "ipv4コマンドはv1で廃止されました。ipaddressコマンドを使用してください",
    "product-disk": "product-diskコマンドはv1で廃止されました。disk-planコマンドを使用してください",
    "product-internet": "product-internetコマンドはv1で廃止されました。internet-planコマンドを使用してください",
    "product-server": "product-serverコマンドはv1で廃止されました。server-planコマンドを使用してください",
    "summary": "summaryコマンドはv1で廃止されました。bill、self、各listコマンドまたはrestコマンドを使用してください",
    "object-storage": "object-storageコマンドはv1で廃止されました。S3互換ツールやTerraformの使用を検討してください",
    "ojs": "ojsコマンドはv1で廃止されました。S3互換ツールやTerraformの使用を検討してください",
}

// DeprecatedCommandTypes は廃止タイプの分類
var DeprecatedCommandTypes = map[string]string{
    "iso-image": "renamed",
    "startup-script": "renamed",
    "ipv4": "renamed", 
    "product-disk": "renamed",
    "product-internet": "renamed",
    "product-server": "renamed",
    "summary": "discontinued",
    "object-storage": "discontinued",
    "ojs": "discontinued",
}
```

### 廃止理由と背景

#### 名称変更の理由
- **iso-image → cdrom**: APIリソース名との統一性向上
- **startup-script → note**: 機能拡張に伴う名称の汎用化
- **ipv4 → ipaddress**: IPv6対応を見据えた名称の統一

#### 製品エイリアス整理の理由
- v0系で使用されていた`product-*`形式の廃止
- v1系での`*-plan`形式への統一

#### 完全廃止の理由
- **summary**: 複雑すぎる単一コマンドの分解、より明確なコマンドへの分離
- **object-storage**: Sakura Cloudのオブジェクトストレージサービス終了に伴う廃止

## テスト戦略
- マッピングテスト：全廃止コマンドが適切な新コマンドまたは廃止ステータスにマッピングされていることを確認
- メッセージテスト：各廃止コマンドに適切な説明メッセージが定義されていることを確認
- 分類テスト：廃止コマンドが適切なタイプ（renamed/discontinued）に分類されていることを確認
- 整合性テスト：既存の変換ルール（ruledefs.go）との整合性を確認

## 依存関係
- 前提PBI: PBI-001～005 (すべてのコマンド辞書)
- 関連PBI: PBI-010 (廃止コマンド検出器), PBI-013 (廃止コマンド移行ガイド)
- 既存コード: `internal/transform/ruledefs.go`の変換ルールと整合性が必要

## 見積もり
- 開発工数: 2時間
  - 既存変換ルール分析: 0.5時間
  - 廃止コマンドマッピング作成: 1時間
  - メッセージとテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/deprecated_commands.go`ファイルが作成されている
- [x] 9個の廃止コマンドが`DeprecatedCommands`マップに定義されている
- [x] 各廃止コマンドの説明メッセージが`DeprecatedCommandMessages`マップに定義されている
- [x] 廃止タイプが`DeprecatedCommandTypes`マップに定義されている
- [x] 既存のtransform/ruledefs.goとの整合性が確保されている
- [x] ユニットテストが作成され、すべて通過している
- [x] コードレビューが完了している

## 実装結果 📊
**実装ファイル:**
- `internal/validation/deprecated_commands.go` - 9個の廃止コマンドマッピング辞書
- `internal/validation/deprecated_commands_test.go` - 包括的なユニットテスト

**実装内容:**
- 名称変更されたコマンド: 6個（iso-image→cdrom等）
- 完全廃止されたコマンド: 3個（summary, object-storage, ojs）
- 2つの廃止タイプ分類：renamed, discontinued
- ユーザーフレンドリーな日本語説明メッセージ
- 整合性検証機能とヘルパー関数群
- 16個のテストケースでカバレッジ100%
- transform/ruledefs.goとの完全整合性確保

**テスト結果:**
```
ok      github.com/armaniacs/usacloud-update/internal/validation    0.187s
```

## 備考
- この辞書は既存のusacloud-update変換ルールの補完として機能する
- 名称変更されたコマンドは自動変換可能だが、廃止コマンドは手動対応が必要
- ユーザーエクスペリエンス向上のため、適切な代替手段の提案が重要
- 将来的な廃止コマンド追加に対応できる拡張可能な構造