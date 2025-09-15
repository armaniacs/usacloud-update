# usacloud-update 使用ガイド

## 概要

`usacloud-update` は、usacloud のバージョン v0、v1.0、v1.1 の記述が混在した bash スクリプトを、v1.1 で動作するように自動変換するツールです。

廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動で更新し、変換できない箇所は詳細なコメントと共に手動対応を促します。

## バージョン管理について

**v1.9.0より、開発版/リリース版を区別したバージョン管理を採用:**

- **奇数マイナーバージョン (v1.9.0, v2.1.0, ...) = 開発版**
  - 新機能開発・実験的機能を含む
  - フィードバック・テスト歓迎
- **偶数マイナーバージョン (v2.0.0, v2.2.0, ...) = リリース版** 
  - 安定版・本番利用推奨
  - 十分なテスト・検証済み

現在のバージョン: **v2.0.0 (安定版) 🎉 本番利用推奨**

### v2.0.0 安定版機能: スマート設定管理

環境変数を自動検出し、即座セットアップを実現する**企業利用対応機能**です。

- **環境変数自動検出** ✅: `usacloud-update config` でSAKURACLOUD_ACCESS_TOKEN等を自動識別
- **5分→30秒へ短縮** ✅: 初回設定時間の83%短縮を実現
- **19パッケージ100%テスト成功** ✅: `make test` で全モジュールの動作保証
- **厳密検証モード** ✅: `--strict-validation` フラグによる高精度検証

### v2.0.0 安定版機能: サンドボックス統合

- **サンドボックス環境** ✅: tk1v ゾーンでの料金なしテスト実行
- **インタラクティブTUI** ✅: ?キーヘルプ機能付き直感的インターフェース
- **バッチ実行** ✅: 全コマンドの自動実行モード
- **ドライランモード** ✅: 実行せずに結果プレビュー
- **BDD統合テスト** ✅: 行動駆動開発によるテスト自動化

## インストール

### ビルド方法

```bash
# リポジトリをクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# ビルド
make build

# バイナリの確認（3つの方法）
./bin/usacloud-update --help
./bin/usacloud-update -h
./bin/usacloud-update help    # 新機能：サブコマンド形式
```

### 動作確認

```bash
# サンプルファイルでの動作テスト
make run

# 期待結果との比較
make verify-sample
```

## 基本的な使い方

### コマンドライン形式

```bash
usacloud-update [オプション]
```

### オプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `--in` | `-` (stdin) | 入力ファイルパス |
| `--out` | `-` (stdout) | 出力ファイルパス |
| `--stats` | `true` | 変更統計を stderr に出力 |
| `--config` | (自動検出) | 設定ファイルパス（サンドボックス機能用） |
| `--sandbox` | `false` | サンドボックス環境での実際のコマンド実行 |
| `--interactive` | `true` | インタラクティブTUIモード (sandboxとの組み合わせで使用) |
| `--dry-run` | `false` | 実際の実行を行わず変換結果のみ表示 |
| `--batch` | `false` | バッチモード: 選択した全コマンドを自動実行 |
| `--strict-validation` | `false` | 厳密検証モード: より高精度な検証を実行 ✨**新機能** |

### 使用パターン

#### 1. 標準入出力を使用

```bash
# パイプライン
cat input.sh | usacloud-update > output.sh

# リダイレクト
usacloud-update < input.sh > output.sh
```

#### 2. ファイルを直接指定

```bash
# ファイル間変換
usacloud-update --in input.sh --out output.sh

# 統計出力を無効化
usacloud-update --in input.sh --out output.sh --stats=false
```

#### 3. 確認しながら実行

```bash
# 統計のみを確認（出力は破棄）
usacloud-update --in script.sh --out /dev/null

# 統計と出力の両方を確認
usacloud-update --in script.sh --out updated_script.sh
```

## 変換例

### 入力ファイル例 (`sample.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

# 古い出力形式
usacloud server list --output-type=csv

# 古いセレクタ構文
usacloud disk read --selector name=mydisk
usacloud server delete --selector tag=to-be-removed

# 変更されたリソース名
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 read --zone tk1a --ipaddress 203.0.113.10

# 古いプロダクト名
usacloud product-disk list

# 廃止されたコマンド
usacloud summary

# 非対応のオブジェクトストレージ
usacloud object-storage list

# ゾーン指定の記述ゆれ
usacloud server list --zone = all
```

### 変換後の出力

```bash
# Updated for usacloud v1.1 by usacloud-update — DO NOT EDIT ABOVE THIS LINE
#!/usr/bin/env bash
set -euo pipefail

# 古い出力形式
usacloud server list --output-type=json # usacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# 古いセレクタ構文
usacloud disk read mydisk # usacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
usacloud server delete to-be-removed # usacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# 変更されたリソース名
usacloud cdrom list # usacloud-update: v1ではリソース名がcdromに統一 (https://manual.sakura.ad.jp/cloud-api/1.1/cdrom/index.html)
usacloud note list # usacloud-update: v1ではstartup-scriptはnoteに統一 (https://docs.usacloud.jp/usacloud/)
usacloud ipaddress read --zone tk1a --ipaddress 203.0.113.10 # usacloud-update: v1ではIPv4関連はipaddressに整理 (https://docs.usacloud.jp/usacloud/references/ipaddress/)

# 古いプロダクト名
usacloud disk-plan list # usacloud-update: v1系では *-plan へ名称統一 (https://docs.usacloud.jp/usacloud/)

# 廃止されたコマンド
# usacloud summary # usacloud-update: summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# 非対応のオブジェクトストレージ
# usacloud object-storage list # usacloud-update: v1ではオブジェクトストレージ操作は非対応方針。S3互換ツール/他プロバイダやTerraformを検討 (https://github.com/sacloud/usacloud/issues/585)

# ゾーン指定の記述ゆれ
usacloud server list --zone=all # usacloud-update: 全ゾーン一括操作は --zone=all を推奨 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

### 統計出力例 (stderr)

```
#L5    --output-type=csv => --output-type=json [output-type-csv-tsv]
#L8    --selector name=mydisk => mydisk [selector-to-arg]
#L9    --selector tag=to-be-removed => to-be-removed [selector-to-arg]
#L12   iso-image => cdrom [iso-image-to-cdrom]
#L13   startup-script => note [startup-script-to-note]
#L14   ipv4 => ipaddress [ipv4-to-ipaddress]
#L17   product-disk => disk-plan [product-alias-product-disk]
#L20   usacloud summary => # usacloud summary [summary-removed]
#L23   usacloud object-storage list => # usacloud object-storage list [object-storage-removed-object-storage]
#L26   --zone = all => --zone=all [zone-all-normalize]
```

## サンドボックス機能

v2.0.0で追加されたサンドボックス機能により、変換したコマンドを実際のSakura Cloud環境でテスト実行できます。

### 概要

サンドボックス機能は以下の特徴があります。

- 安全なテスト環境：Sakura Cloud の tk1v ゾーン（サンドボックス専用）を使用
- 料金なし：サンドボックス環境では課金されません
- インタラクティブTUI：個別コマンドの選択・実行が可能
- バッチ実行：全コマンドの自動実行も可能
- ドライランモード：実行せずに結果をプレビュー

### 環境設定

#### 1. 設定ファイルの作成

サンドボックス機能を使用するには、Sakura CloudのAPIキーが必要です。

**【推奨】設定ファイル方式（環境変数自動検出対応 ✨新機能）**:

```bash
# スマート設定：環境変数自動検出・設定ファイル生成
usacloud-update config

# 手動で設定ファイルを作成
mkdir -p ~/.config/usacloud-update
cp usacloud-update.conf.sample ~/.config/usacloud-update/usacloud-update.conf

# 設定ファイルを編集してAPIキーを設定
vim ~/.config/usacloud-update/usacloud-update.conf
```

または、初回実行時に対話的に作成できます。
```bash
# 設定ファイルがない場合、自動的に対話的設定が開始されます
usacloud-update --sandbox
```

**カスタム設定ファイルパス**:
```bash
# --config フラグで特定の設定ファイルを指定
usacloud-update --sandbox --config /path/to/custom.conf --in script.sh
```

**カスタム設定ディレクトリ**（CI/CD環境など特別な場合）
```bash
export USACLOUD_UPDATE_CONFIG_DIR=/path/to/custom/config
usacloud-update --sandbox
```

**【レガシー】環境変数方式**:
```bash
# 環境変数を直接設定（廃止予定・設定ファイル移行推奨）
export SAKURACLOUD_ACCESS_TOKEN="your-access-token"
export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-secret"

# 注意：環境変数がある場合、`usacloud-update config` で移行提案されます
```

#### 2. APIキーの取得方法

1. [さくらのクラウド コントロールパネル](https://secure.sakura.ad.jp/cloud/)にログイン
2. 左メニューの「設定」→「APIキー」を選択
3. 「追加」ボタンをクリック
4. 名前を入力（例: `usacloud-update-sandbox`）
5. 「作成」ボタンをクリック
6. 表示されたアクセストークンとシークレットを設定ファイルに設定

#### 3. usacloud CLIのインストール

```bash
# macOS (Homebrew)
brew install sacloud/tap/usacloud

# Linux/Windows バイナリ
# https://docs.usacloud.jp/usacloud/installation/ を参照
```

### 使用パターン

#### 1. インタラクティブモード（デフォルト）

```bash
# 基本的なサンドボックス実行
usacloud-update --sandbox --in script.sh

# TUIが起動し、以下の操作が可能：
# - ↑↓: コマンド選択
# - Space: 個別コマンドの選択/解除
# - a: 全選択
# - n: 全解除
# - e: 選択したコマンドを実行
# - Enter: 現在のコマンドを個別実行
# - q: 終了
```

#### 2. ドライランモード

```bash
# 実行せずに結果をプレビュー
usacloud-update --sandbox --dry-run --in script.sh

# 出力例:
# [DRY RUN] Would execute: usacloud server list --zone=tk1v --output-type=json
```

#### 3. バッチモード

```bash
# 全コマンドを自動実行（TUIなし）
usacloud-update --sandbox --batch --in script.sh

# 非インタラクティブ + バッチ実行
usacloud-update --sandbox --interactive=false --batch --in script.sh
```

#### 4. 組み合わせ例

```bash
# ドライラン + インタラクティブ
usacloud-update --sandbox --dry-run --in script.sh

# バッチ + ドライラン
usacloud-update --sandbox --batch --dry-run --in script.sh
```

### TUI操作方法

インタラクティブモードでは、以下の画面構成で表示されます。

```
┌─ 📋 Converted Commands ─────────┐┌─ 🔍 Command Details ────────────┐
│✓ L1: usacloud server list...   ││Line 1:                          │
│  L2: usacloud disk list...     ││                                 │
│✓ L3: usacloud cdrom list...    ││Original:                        │
│...                             ││  usacloud server list --out... │
└─────────────────────────────────┘│Converted:                       │
                                   │  usacloud server list --out... │
┌─ 📊 Execution Results ──────────┐│Rule: output-type-csv-tsv        │
│Execution Summary:               ││                                 │
│                                 ││Execution Result:                │
│Total Executed: 2               ││Status: Success                  │
│Successful: 2                   ││Duration: 1.2s                   │
│Failed: 0                       ││                                 │
│Skipped: 1                      │└─────────────────────────────────┘
└─────────────────────────────────┘
Commands: 3  Selected: 2  Executed: 2  Successful: 2
Progress: [████████████████████] 100.0% (2/2) - Execution completed
```

**キーボードショートカット**:
- `↑↓`: コマンド選択
- `Space`: 選択状態の切り替え
- `Enter`: 現在のコマンドを実行
- `a`: 全usacloudコマンドを選択
- `n`: 全ての選択を解除
- `e`: 選択したコマンドを実行
- `?`: ヘルプパネルの表示/非表示切り替え（v1.9.0新機能）
- `q` または `Ctrl+C`: 終了

### BDDテスト（v1.9.0完全実装済み）

サンドボックス機能のBDD（行動駆動開発）テストを実行します。

```bash
make bdd
```

**BDD機能の実装状況**：
- **全シナリオ実装完了**：interactive TUI機能、コマンド変換・実行、結果表示
- **7つのBDDステップ関数**：包括的なサンドボックス機能テストを自動実行
- **Godog統合**：`features/sandbox.feature` による仕様駆動テスト
- **品質保証**：サンドボックス機能の動作を自動検証

### 高精度検証モード（v1.9.1新機能 ✨）

より厳密な検証を実行するモードです。

```bash
# 厳密検証モードでスクリプト変換
usacloud-update --strict-validation script.sh

# サンドボックスと組み合わせ
usacloud-update --sandbox --strict-validation script.sh
```

**厳密検証の特徴**：
- **高精度エラー検出**：通常の検証よりも詳細な問題を発見
- **詳細なフィードバック**：問題箇所の特定と修正提案を提供
- **品質向上**：変換前の入力スクリプトの品質を事前チェック

### 制限事項

サンドボックス環境では以下の制限があります。

- リソース機能制限：作成したリソースは正常に動作しません
- ネットワーク制限：インターネット接続はありません
- VNCコンソール：使用できません
- アップロード制限：アーカイブ/ISOイメージのアップロード不可

### セキュリティ機能

- ゾーン強制：全コマンドは `--zone=tk1v` で実行されます
- 危険操作禁止：`delete`, `shutdown`, `reset` 等は実行されません
- タイムアウト：30秒でコマンド実行をタイムアウト

## 変換ルール詳細

### 1. 出力形式の変換

**対象**: `--output-type=csv`, `--output-type=tsv`, `-o csv`, `-o tsv`

**変換後**: `--output-type=json`

**理由**: v1.0 で CSV/TSV 出力は廃止され、JSON のみサポートされます。

**対応方法**: CSV/TSV 形式が必要な場合は `--query` (JMESPath) や `jq` コマンドを使用して変換します。

### 2. セレクタの引数化

**対象**: `--selector name=xxx`, `--selector id=xxx`, `--selector tag=xxx`

**変換後**: `xxx` (引数化)

**理由**: v1 で `--selector` オプションは廃止され、引数で指定する形式に変更されました。

注意点：タグ指定に曖昧性があると実行時エラーが発生します。

### 3. リソース名の統一

| v0 リソース名 | v1.1 リソース名 | 変更理由 |
|-------------|--------------|---------|
| `iso-image` | `cdrom` | リソース名の統一 |
| `startup-script` | `note` | 名称の整理 |
| `ipv4` | `ipaddress` | IPv4 関連の整理 |

### 4. プロダクトエイリアスの統一

**対象**: `product-disk`, `product-internet`, `product-server`

**変換後**: `disk-plan`, `internet-plan`, `server-plan`

**理由**: v1 系では `-plan` 形式で名称が統一されています。

### 5. 廃止コマンドの処理

**対象**: `usacloud summary`

**処理**: コメントアウト + 代替手段の案内です。

**代替案**: `bill`, `self`, 各種 `list` コマンドまたは `rest` コマンドを使用します。

### 6. オブジェクトストレージの非対応

**対象**: `usacloud object-storage`, `usacloud ojs`

**処理**: コメントアウト + 代替手段の案内です。

**代替案**: S3 互換ツールや Terraform の使用を検討します。

### 7. ゾーン指定の正規化

**対象**: `--zone = all` (空白あり)  
**変換後**: `--zone=all` (空白なし)  
**理由**: 記述の統一化です。

## 注意事項

### 手動対応が必要な箇所

1. **コメントアウトされた行**
   - `# usacloud summary` → 代替コマンドへの置き換え
   - `# usacloud object-storage` → 別ツールへの移行

2. **セレクタの引数化**
   - タグ指定で複数のリソースがマッチする場合
   - 実行前にタグの一意性を確認

3. **出力形式の変更**
   - JSON から CSV/TSV への後処理が必要な場合
   - `--query` や `jq` コマンドの追加検討

### ファイルの取り扱い

1. **バックアップの作成**
   ```bash
   cp original_script.sh original_script.sh.backup
   usacloud-update --in original_script.sh --out updated_script.sh
   ```

2. **段階的な確認**
   ```bash
   # 統計のみ確認
   usacloud-update --in script.sh --out /dev/null
   
   # 変換実行
   usacloud-update --in script.sh --out script_v1.1.sh
   
   # 動作確認
   bash script_v1.1.sh
   ```

## トラブルシューティング

### よくある問題

#### 1. 「変換されない行がある」

**原因**: 想定外の記述形式やスペース
**対策**: 
```bash
# 統計出力で未変換行を確認
usacloud-update --in script.sh --out /dev/null --stats=true
```

#### 2. 「タグ指定でエラーが発生する」

原因：タグの曖昧性  
対策：タグが一意に識別できることを事前確認します。

#### 3. 「出力形式が期待と異なる」

原因：JSON 出力への変更  
対策：`--query` や `jq` を使用した後処理を追加します。

### ログの活用

```bash
# 詳細な変更ログを保存
usacloud-update --in script.sh --out updated.sh 2> changes.log

# ログの確認
cat changes.log
```

### サンドボックス機能のトラブルシューティング

#### 1. 「認証エラーが発生する」

**症状**:
```
Configuration validation failed:
SAKURACLOUD_ACCESS_TOKEN is required
```

**対策**:
```bash
# 【推奨】設定ファイルの確認
cat ~/.config/usacloud-update/usacloud-update.conf

# 設定ファイルの再作成
cp usacloud-update.conf.sample ~/.config/usacloud-update/usacloud-update.conf
vim ~/.config/usacloud-update/usacloud-update.conf  # APIキーを正しく設定

# または対話的に設定を作成
usacloud-update --sandbox  # 設定ファイルがない場合は対話的作成

# 【レガシー】環境変数の確認（廃止予定）
cat .env
echo $SAKURACLOUD_ACCESS_TOKEN
```

#### 2. 「usacloud CLI not found エラー」

**症状**:
```
Error: usacloud CLI not found
Please install usacloud CLI: https://docs.usacloud.jp/usacloud/installation/
```

**対策**:
```bash
# macOS (Homebrew)
brew install sacloud/tap/usacloud

# インストール確認
usacloud version
```

#### 3. 「接続エラー・タイムアウト」

**症状**:
```
command timed out after 30s
```

**対策**:
```bash
# ドライランモードで事前確認
usacloud-update --sandbox --dry-run --in script.sh

# ネットワーク接続確認
usacloud --zone=tk1v server list

# APIエンドポイント確認（設定ファイル）
SAKURACLOUD_API_URL=https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/
```

#### 4. 「TUIが正常に動作しない」

**症状**: 画面がちらつく、キーボード操作が効きません。

**対策**:
```bash
# 非インタラクティブモードで実行
usacloud-update --sandbox --interactive=false --batch --in script.sh

# ターミナルのサイズ確認（最小80x24推奨）
echo $COLUMNS x $LINES

# ヘルプパネル切り替え機能を利用（v1.9.0）
# TUI内で ? キーを押してヘルプの表示/非表示を切り替え
```

#### 5. 「コマンドが実行されない（スキップされる）」

症状：コマンドがSkippedになる。

原因と対策は以下の通りです。
- コメント行：`#` で始まる行は自動スキップ
- 非usacloudコマンド：usacloud以外のコマンドはスキップ
- 危険操作：delete, shutdownなどは安全のためスキップ

```bash
# デバッグモードで詳細確認
USACLOUD_UPDATE_DEBUG=true usacloud-update --sandbox --in script.sh
```

### デバッグ手順

1. **統計出力の確認**
   ```bash
   usacloud-update --in problematic.sh --out /dev/null
   ```

2. **部分的なテスト**
   ```bash
   echo "usacloud server list --output-type=csv" | usacloud-update
   ```

3. **バックアップからの復元**
   ```bash
   cp original.sh.backup original.sh
   ```

### 高度なデバッグ・開発

**実装レベルのデバッグが必要な場合**:
- **[実装詳細ガイド](ref/detailed-implementation-reference.md)** - コンポーネント内部動作の理解
- **[API リファレンス](ref/api-reference.md)** - プログラマティックな解析・テスト
- **[テストフレームワークガイド](ref/testing-framework-reference.md)** - カスタムテストケースの作成

## 開発者・上級者向けリファレンス

### 実装・アーキテクチャ詳細
- [実装詳細ガイド](ref/detailed-implementation-reference.md)：1,313+行の包括的実装ガイド（新規作成）
  - 全コンポーネントの詳細アーキテクチャ
  - 拡張ポイントと実装パターン
  - Transform Engine、Validation System、TUI、Sandbox の内部構造

### テスト・品質保証（v1.9.1安定化完了 ✨）
- [テストフレームワークガイド](ref/testing-framework-reference.md)：E2E/BDD/統合テスト完全ガイド（新規作成）
  - **完全テスト成功（`make test`）**：17テストパッケージで100%成功・安定化
  - **BDD/E2E/統合/パフォーマンステスト**の使用方法
  - **カスタムテストケース作成ガイド**

### API・統合開発
- [API リファレンス](ref/api-reference.md)：全パッケージの完全APIドキュメント
  - 全publicインターフェースとメソッド
  - 実用的なコード例とエラーハンドリング
  - システム統合・拡張のガイダンス

### その他技術ドキュメント
- [コンポーネントアーキテクチャ](ref/component-architecture.md)：システム全体構造
- [開発ワークフロー](ref/development-workflow.md)：開発プロセスとベストプラクティス
- [テスト戦略ガイド](ref/testing-guide.md)：多層テスト戦略

## 関連リンク

### 基本リンク
- [usacloud v1.0.0 アップグレードガイド](https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
- [usacloud v1.1 リファレンス](https://docs.usacloud.jp/usacloud/)
- [さくらのクラウド API v1.1](https://manual.sakura.ad.jp/cloud-api/1.1/)
- [usacloud GitHub リポジトリ](https://github.com/sacloud/usacloud/)

### サンドボックス機能関連
- [さくらのクラウド サンドボックス環境](https://manual.sakura.ad.jp/cloud/server/sandbox.html)
- [usacloud CLIインストールガイド](https://docs.usacloud.jp/usacloud/installation/)
- [さくらのクラウド APIキー管理](https://manual.sakura.ad.jp/cloud/api/apikey.html)
- [Godog BDD Framework](https://github.com/cucumber/godog)
- [tview TUI Library](https://github.com/rivo/tview)

## サポート

ツールに関する問題や改善要望は、プロジェクトの GitHub Issues で報告してください。

変換結果は必ず事前にテスト環境で確認し、本番環境での使用前に十分な検証することを推奨します。