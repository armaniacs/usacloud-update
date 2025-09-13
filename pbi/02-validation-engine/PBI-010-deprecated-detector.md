# PBI-010: 廃止コマンド自動検出

## 概要
usacloudの進化に伴って廃止されたコマンドを自動的に検出し、適切な代替コマンドや移行方法を提案する機能を実装する。PBI-006で定義された廃止コマンドマッピングを活用して、ユーザーの移行を支援する。

## 受け入れ条件 ✅ **完了 2025-09-09**
- [x] 全ての廃止コマンド（9個）が適切に検出される
- [x] 名称変更コマンドには新コマンド名が提案される
- [x] 完全廃止コマンドには代替手段が提案される
- [x] 廃止の種類に応じて適切なメッセージが提供される
- [x] 既存の変換ルールとの整合性が保たれている

## 技術仕様

### 検出対象の廃止コマンド

#### 1. 名称変更されたコマンド（6個）
```bash
iso-image → cdrom
startup-script → note
ipv4 → ipaddress
product-disk → disk-plan
product-internet → internet-plan
product-server → server-plan
```

#### 2. 完全廃止されたコマンド（3個）
```bash
summary → (代替手段: bill, self, rest等の組み合わせ)
object-storage → (代替手段: S3互換ツール、Terraform)
ojs → (object-storageのエイリアス、同様に廃止)
```

### 実装構造
```go
// internal/validation/deprecated_detector.go
package validation

import (
    "fmt"
    "strings"
)

// DeprecationInfo は廃止コマンドの情報
type DeprecationInfo struct {
    Command         string   // 廃止されたコマンド
    ReplacementCommand string // 代替コマンド（空の場合は完全廃止）
    DeprecationType   string // "renamed" または "discontinued"
    Message          string // 詳細な説明メッセージ
    AlternativeActions []string // 代替手段（完全廃止の場合）
    DocumentationURL  string // 関連ドキュメントのURL
}

// DeprecatedCommandDetector は廃止コマンド検出器
type DeprecatedCommandDetector struct {
    deprecatedCommands map[string]*DeprecationInfo
}

// NewDeprecatedCommandDetector は新しい検出器を作成
func NewDeprecatedCommandDetector() *DeprecatedCommandDetector {
    detector := &DeprecatedCommandDetector{
        deprecatedCommands: make(map[string]*DeprecationInfo),
    }
    
    // 廃止コマンド情報の初期化
    detector.initializeDeprecatedCommands()
    
    return detector
}

// Detect は廃止コマンドを検出
func (d *DeprecatedCommandDetector) Detect(command string) *DeprecationInfo {
    return d.deprecatedCommands[command]
}

// IsDeprecated はコマンドが廃止されているかを判定
func (d *DeprecatedCommandDetector) IsDeprecated(command string) bool {
    return d.deprecatedCommands[command] != nil
}

// GetReplacementCommand は代替コマンドを取得
func (d *DeprecatedCommandDetector) GetReplacementCommand(command string) string {
    if info := d.deprecatedCommands[command]; info != nil {
        return info.ReplacementCommand
    }
    return ""
}
```

### 廃止コマンド情報の初期化
```go
func (d *DeprecatedCommandDetector) initializeDeprecatedCommands() {
    // 名称変更されたコマンド
    d.deprecatedCommands["iso-image"] = &DeprecationInfo{
        Command:            "iso-image",
        ReplacementCommand: "cdrom", 
        DeprecationType:    "renamed",
        Message:           "iso-imageコマンドはv1で廃止されました。cdromコマンドを使用してください。",
        DocumentationURL:  "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
    }
    
    d.deprecatedCommands["startup-script"] = &DeprecationInfo{
        Command:            "startup-script",
        ReplacementCommand: "note",
        DeprecationType:    "renamed", 
        Message:           "startup-scriptコマンドはv1で廃止されました。noteコマンドを使用してください。",
        DocumentationURL:  "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
    }
    
    // ... その他の名称変更コマンド
    
    // 完全廃止されたコマンド
    d.deprecatedCommands["summary"] = &DeprecationInfo{
        Command:         "summary",
        ReplacementCommand: "",
        DeprecationType: "discontinued",
        Message:        "summaryコマンドはv1で廃止されました。",
        AlternativeActions: []string{
            "請求情報は 'usacloud bill list' を使用してください",
            "アカウント情報は 'usacloud self read' を使用してください", 
            "個別リソース情報は各リソースの 'list' コマンドを使用してください",
            "詳細な情報が必要な場合は 'usacloud rest' コマンドを使用してください",
        },
        DocumentationURL: "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
    }
    
    // ... その他の完全廃止コマンド
}
```

### 検出・提案ロジック

#### 1. 基本検出フロー
1. 入力されたメインコマンドが廃止コマンド辞書に存在するかチェック
2. 存在する場合、廃止タイプに応じた処理を実行
3. 適切なメッセージと代替手段を提案

#### 2. 名称変更コマンドの処理
```go
func (d *DeprecatedCommandDetector) handleRenamedCommand(info *DeprecationInfo) string {
    return fmt.Sprintf(
        "%s は %s に名称変更されました。%s を使用してください。\n詳細: %s",
        info.Command,
        info.ReplacementCommand, 
        info.ReplacementCommand,
        info.DocumentationURL,
    )
}
```

#### 3. 完全廃止コマンドの処理
```go
func (d *DeprecatedCommandDetector) handleDiscontinuedCommand(info *DeprecationInfo) string {
    message := fmt.Sprintf("%s\n\n代替手段:\n", info.Message)
    for _, action := range info.AlternativeActions {
        message += fmt.Sprintf("  - %s\n", action)
    }
    message += fmt.Sprintf("\n詳細: %s", info.DocumentationURL)
    return message
}
```

## テスト戦略
- 廃止コマンド検出テスト：9個すべての廃止コマンドが正しく検出されることを確認
- 名称変更テスト：名称変更されたコマンドに対して適切な新コマンドが提案されることを確認
- 完全廃止テスト：完全廃止されたコマンドに対して適切な代替手段が提案されることを確認
- メッセージテスト：各廃止タイプに応じて適切なメッセージが生成されることを確認
- 整合性テスト：既存のtransform/ruledefs.goとの整合性を確認
- ドキュメントリンクテスト：提案されるドキュメントURLが有効であることを確認

## 依存関係
- 前提PBI: PBI-006 (廃止コマンドマッピング), PBI-007 (コマンドパーサー)
- 関連PBI: PBI-013 (廃止コマンド移行ガイド) - より詳細な移行支援
- 既存コード: `internal/transform/ruledefs.go`との整合性が必要

## 見積もり
- 開発工数: 3時間
  - 基本検出ロジック実装: 1.5時間
  - メッセージ生成とドキュメント統合: 1時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-09-09**
- [x] `internal/validation/deprecated_detector.go`ファイルが作成されている
- [x] `DeprecatedCommandDetector`構造体と検出メソッドが実装されている
- [x] 9個すべての廃止コマンドが正しく検出される
- [x] 名称変更コマンドに対して適切な新コマンドが提案される
- [x] 完全廃止コマンドに対して適切な代替手段が提案される
- [x] わかりやすい日本語メッセージが提供される
- [x] ドキュメントURLが正しく提供される
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] 既存の変換ルールとの整合性が確認されている
- [x] コードレビューが完了している

## 実装結果 📊

**実装ファイル:**
- `internal/validation/deprecated_detector.go` - DeprecatedCommandDetector構造体とメソッド群の完全実装
- `internal/validation/deprecated_detector_test.go` - 18テスト関数による包括的テスト

**実装内容:**
- 9個の廃止コマンド完全検出（名称変更6個、完全廃止3個）
- DeprecationInfo構造による詳細情報管理
- 2つの廃止タイプ分類（renamed, discontinued）
- 建設的な移行メッセージ生成機能
- 完全廃止コマンドの代替手段提案機能
- 大文字小文字正規化処理
- PBI-006廃止コマンドマッピングとの整合性検証

**テスト結果:**
- 18のテスト関数すべて成功
- 全廃止コマンドの検出確認
- 名称変更/完全廃止の適切な分類
- 代替手段提案の検証
- メッセージ生成機能の検証
- 既存マッピングとの整合性確認

**技術的特徴:**
- 効率的なマップベース検索
- ユーザーフレンドリーな日本語メッセージ
- ドキュメントURL統合による詳細サポート
- 拡張可能な廃止コマンド管理構造
- 既存システムとの完全整合性保証

## 備考
- この検出器は既存のusacloud-update変換機能を補完する
- ユーザーエクスペリエンスの向上に直結する重要な機能
- 新しい廃止コマンドが発生した場合の拡張性を考慮した設計
- エラーメッセージではなく、建設的な提案として提供することが重要