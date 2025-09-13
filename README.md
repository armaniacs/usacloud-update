# usacloud-update

[![Go](https://github.com/armaniacs/usacloud-update/workflows/Go/badge.svg)](https://github.com/armaniacs/usacloud-update/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**usacloud v0/v1.0/v1.1 混在スクリプトを v1.1 互換に自動変換 + サンドボックス実行**

usacloud-update は、異なるバージョンの usacloud コマンドが混在した bash スクリプトを v1.1 で動作するように自動変換するツールです。さらに、Sakura Cloud サンドボックス環境での実際のテスト実行まで一貫してサポートします。

## 🚀 v1.9.0 主要機能

### サンドボックス実行
変換したコマンドを Sakura Cloud tk1v ゾーンで料金なしで実際にテスト実行します。

### インタラクティブ TUI
直感的なターミナル UI でコマンドを選択・実行・結果確認します。
- ヘルプキー：ヘルプパネル表示/非表示切り替え（新機能）
- スペースキー：コマンド選択/解除
- エンターキー：選択したコマンド実行

### 完全な BDD テスト自動化
行動駆動開発による包括的なテスト仕様と実装です。
- 全シナリオ実装完了（品質保証済み）
- `make bdd` でテスト実行可能

### 柔軟な実行モード
- インタラクティブモード：TUI でコマンド個別選択・実行
- バッチモード：全コマンド自動実行
- ドライランモード：実行せずに結果プレビュー

## クイックスタート

### 基本的な変換
```bash
# パイプラインで使用
cat input.sh | usacloud-update > output.sh

# ファイル指定変換
usacloud-update --in script.sh --out updated_script.sh
```

### サンドボックス実行
```bash
# インタラクティブ TUI でサンドボックス実行
usacloud-update --sandbox --in script.sh

# ドライランモード（安全確認）
usacloud-update --sandbox --dry-run --in script.sh

# バッチモード（全自動実行）
usacloud-update --sandbox --batch --in script.sh
```

## インストール

### ユーザー向けインストール

```bash
# リポジトリをクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# ビルド
make build

# 動作確認
./bin/usacloud-update --version
```

### 開発者向けセットアップ

```bash
# 開発環境セットアップ
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# 依存関係インストール
go mod tidy

# 開発ツール実行
make build          # ビルド
make test           # テスト実行
make bdd            # BDDテスト
make vet            # 静的解析
make fmt            # コード整形

# ドキュメント確認
open ref/detailed-implementation-reference.md
```

## 設定

サンドボックス機能には Sakura Cloud の API 認証情報が必要です。

### 推奨: 設定ファイル方式
```bash
# 初回実行時に対話的に設定ファイルを作成
usacloud-update --sandbox --in your-script.sh

# または手動で設定ファイルを作成
cp usacloud-update.conf.sample ~/.config/usacloud-update/usacloud-update.conf
# APIキーを編集
```

### 従来: 環境変数方式
```bash
export SAKURACLOUD_ACCESS_TOKEN="your-access-token"
export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

## 主な変換ルール

- 出力形式：CSV/TSV → JSON（`--output-type=json`）
- セレクタ廃止：`--selector name=xxx` → `xxx`（引数化）
- リソース名変更：`iso-image` → `cdrom`、`startup-script` → `note`
- 廃止コマンド：自動コメントアウト + 手動対応指示

## 品質保証

### 包括的テスト体制（56.1% カバレッジ）
- 全ユニットテスト通過（`make test`）：17テストパッケージで100%成功
- BDD テスト完全実装（`make bdd`）：7つのプレースホルダ関数を完全実装
- 多層テスト実装：E2E/統合/BDD/パフォーマンス/回帰テスト
- エッジケーステスト：並行処理、エラー条件、境界値を網羅
- ゴールデンファイルテスト：出力比較による動作保証

### 実装完了度
- コア機能：Transform Engine（100%カバレッジ）、Validation System
- TUI システム：インタラクティブUI、ヘルプシステム（62.7%カバレッジ）
- サンドボックス：安全な実行環境（78.5%カバレッジ）
- 設定管理：多ソース優先度システム（85.7%カバレッジ）
- テストインフラ：5,175+行の包括的テストコード

### 開発者支援
- 包括的ドキュメント：2,383+行の詳細実装・テストガイド
- 完全なAPIリファレンス：全パッケージの使用方法とサンプル
- 段階的学習パス：初心者から上級者まで対応

## ドキュメント

### ユーザー向けドキュメント
- [詳細な使用ガイド](README-Usage.md)：完全な使用方法とサンドボックス機能
- [変更履歴](CHANGELOG.md)：バージョン管理と開発履歴

### 開発者向け包括的リファレンス
- [実装詳細ガイド](ref/detailed-implementation-reference.md)：1,313+行の包括的実装リファレンス（新規作成）
- [API リファレンス](ref/api-reference.md)：完全なAPIドキュメントと使用例
- [テストフレームワークガイド](ref/testing-framework-reference.md)：E2E/BDD/統合テストの完全ガイド（新規作成）

### 技術アーキテクチャ
- [コンポーネントアーキテクチャ](ref/component-architecture.md)：システム全体構造
- [コアアルゴリズム](ref/core-algorithms.md)：Transform/Validation エンジン詳細
- [テスト戦略ガイド](ref/testing-guide.md)：多層テスト戦略（56.1%カバレッジ）

### 開発プロセス
- [開発ワークフロー](ref/development-workflow.md)：開発プロセスとリリース準備
- [プロジェクト依存関係](ref/project-dependencies.md)：依存関係管理とプラットフォームサポート

## 開発者貢献

### 貢献の始め方

1. 理解：[実装詳細ガイド](ref/detailed-implementation-reference.md) で全体アーキテクチャを把握
2. 実装：[API リファレンス](ref/api-reference.md) で拡張ポイントを確認
3. テスト：[テストフレームワークガイド](ref/testing-framework-reference.md) でテスト戦略を学習

### 主な貢献分野

- Transform Engine：新しい変換ルールの追加（100%カバレッジ達成済み）
- Validation System：コマンド検証機能の拡張
- TUI Enhancement：ユーザーインターフェース改善
- Testing：テストカバレッジ向上（現在56.1% → 目標80%）

### 開発ロードマップ

- v2.0（2025 Q3-Q4）：安定化リリース、品質向上とバグ修正に専念
- v2.1以降：新機能の開発を再開（詳細は [TODO-target-2_0.md](TODO-target-2_0.md) を参照）

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。