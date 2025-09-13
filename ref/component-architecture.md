# Component Architecture Reference

このドキュメントでは、usacloud-updateプロジェクトの各内部コンポーネントの詳細アーキテクチャと相互関係について説明します。

## システム概要

usacloud-updateは以下の主要コンポーネントで構成される変換・検証・実行システムです：

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Frontend  │────│  Transform      │────│  Validation     │
│   (main.go)     │    │  Engine         │    │  System         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   TUI System    │────│  Sandbox        │────│  Config         │
│   (Interactive) │    │  Executor       │    │  Manager        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Scanner       │────│  Security       │────│  Testing        │
│   System        │    │  Components     │    │  Framework      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## コンポーネント詳細

### 1. Transform Engine (`internal/transform/`)

**目的**: usacloudコマンドの v0.x/v1.0 から v1.1 への変換

#### 主要ファイル
- `engine.go` - 変換エンジンのコアロジック
- `ruledefs.go` - 変換ルール定義（9つのカテゴリ）
- `rules.go` - ルール適用機構

#### アーキテクチャ
```go
type Engine struct {
    rules []Rule
}

type Rule interface {
    Name() string
    Apply(line string) (string, bool, string, string)
}

type Result struct {
    Line    string
    Changed bool
    Changes []Change
}
```

#### 変換フロー
1. **行単位処理**: 入力を1行ずつ解析
2. **ルール適用**: 9カテゴリのルールを順次適用
3. **結果生成**: 変更内容と説明コメントを生成

### 2. Validation System (`internal/validation/`)

**目的**: usacloudコマンドの構文検証と提案生成

#### 主要コンポーネント
- `parser.go` - コマンドライン解析
- `main_command_validator.go` - メインコマンド検証
- `subcommand_validator.go` - サブコマンド検証
- `similar_command_suggester.go` - 類似コマンド提案

#### データ構造
```go
type CommandLine struct {
    Raw         string
    MainCommand string
    SubCommand  string
    Arguments   []string
    Options     map[string]string
    Flags       []string
}
```

#### 検証フロー
1. **構文解析**: 入力コマンドの構造解析
2. **コマンド検証**: メイン・サブコマンドの妥当性確認
3. **提案生成**: エラー時の修正候補提示

### 3. TUI System (`internal/tui/`)

**目的**: インタラクティブなターミナルユーザーインターフェース

#### 主要機能
- `app.go` - メインアプリケーション
- `file_selector.go` - ファイル選択インターフェース
- `filter/` - フィルタリングシステム
- `preview/` - プレビュー機能

#### アーキテクチャ特徴
- **tview**ライブラリベースのUI
- **?**キーによるヘルプ切り替え
- **動的レイアウト管理**

### 4. Sandbox Executor (`internal/sandbox/`)

**目的**: 実際のSakura Cloud環境でのコマンド実行

#### 主要コンポーネント
- `executor.go` - メイン実行エンジン
- `error_handler.go` - エラーハンドリング
- `parallel.go` - 並列実行制御
- `persistence.go` - 結果永続化
- `retry.go` - リトライ機構
- `validation.go` - 実行前検証

#### 実行フロー
1. **前処理**: コマンド検証とサンドボックス環境準備
2. **実行**: tk1vゾーンでの実際のusacloudコマンド実行
3. **後処理**: 結果記録とクリーンアップ

### 5. Configuration Manager (`internal/config/`)

**目的**: 設定ファイルとプロファイル管理

#### 設定方式
- **新方式**: `~/.config/usacloud-update/usacloud-update.conf` (INI形式)
- **レガシー**: `.env`ファイル（後方互換性）

#### 主要機能
- `file.go` - ファイルベース設定
- `env.go` - 環境変数ベース設定
- `interactive.go` - 対話的設定作成
- `profile/` - マルチプロファイル管理

### 6. Security Components (`internal/security/`)

**目的**: セキュリティ機能とデータ保護

#### 主要機能
- `encryption.go` - AES-256-GCM暗号化
- `filter.go` - セキュリティフィルタリング
- `audit.go` - 監査ログ
- `monitor.go` - セキュリティ監視

### 7. Scanner System (`internal/scanner/`)

**目的**: スクリプトファイルのスキャンと解析

#### 機能
- `scanner.go` - ファイルスキャン
- `detector.go` - usacloudコマンド検出
- `intelligent_scanner.go` - インテリジェント解析

### 8. Testing Framework (`internal/testing/`)

**目的**: テスト支援とゴールデンファイル管理

#### コンポーネント
- `golden_test_framework.go` - ゴールデンファイルテスト
- `golden_generator.go` - テストデータ生成

### 9. BDD System (`internal/bdd/`)

**目的**: 行動駆動開発テスト

#### 主要ファイル
- `steps.go` - BDDステップ定義
- `extended_steps.go` - 拡張シナリオ
- `error_scenarios.go` - エラーシナリオ

## コンポーネント間の相互作用

### データフロー

```
Input Script
     │
     ▼
┌─────────┐    ┌─────────┐    ┌─────────┐
│ Scanner │───▶│Transform│───▶│Validation│
└─────────┘    └─────────┘    └─────────┘
     │              │              │
     ▼              ▼              ▼
┌─────────┐    ┌─────────┐    ┌─────────┐
│   TUI   │◀───┤ Config  │───▶│Sandbox  │
└─────────┘    └─────────┘    └─────────┘
     │              │              │
     ▼              ▼              ▼
Output Results ← Security ← Persistence
```

### 依存関係マトリックス

| Component | Transform | Validation | TUI | Sandbox | Config | Security |
|-----------|-----------|------------|-----|---------|---------|----------|
| CLI       | ✓         | ✓          | ✓   | ✓       | ✓       | -        |
| Transform | -         | -          | -   | -       | -       | -        |
| Validation| -         | -          | -   | -       | -       | -        |
| TUI       | ✓         | ✓          | -   | ✓       | ✓       | -        |
| Sandbox   | ✓         | ✓          | -   | -       | ✓       | ✓        |
| Config    | -         | -          | -   | -       | -       | ✓        |
| Security  | -         | -          | -   | -       | -       | -        |

## 設計原則

### 1. 単一責任原則
各コンポーネントは明確に定義された単一の責任を持つ

### 2. 依存性の逆転
上位レベルのモジュールは下位レベルの詳細に依存しない

### 3. インターフェース分離
クライアントは使用しないメソッドへの依存を強制されない

### 4. 開放閉鎖原則
拡張に対して開放的で、変更に対して閉鎖的

## パフォーマンス考慮事項

### 1. Transform Engine
- **行単位処理**による大ファイル対応
- **正規表現最適化**によるマッチング高速化

### 2. Validation System
- **キャッシュ機構**による重複検証回避
- **並列検証**による処理速度向上

### 3. Sandbox Executor
- **並列実行制御**によるスループット向上
- **リソース制限**による安定性確保

---

**最終更新**: 2025年1月
**バージョン**: v1.9.0対応
**メンテナー**: 開発チーム