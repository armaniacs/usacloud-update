# usacloud-update 使用ガイド

## 概要

`usacloud-update` は、usacloud のバージョン v0、v1.0、v1.1 の記述が混在した bash スクリプトを、v1.1 で動作するように自動変換するツールです。

廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動で更新し、変換できない箇所は適切なコメントと共に手動対応を促します。

## インストール

### ビルド方法

```bash
# リポジトリをクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# ビルド
make build

# バイナリの確認
./bin/usacloud-update --help
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

## 変換ルール詳細

### 1. 出力形式の変換

**対象**: `--output-type=csv`, `--output-type=tsv`, `-o csv`, `-o tsv`
**変換後**: `--output-type=json`
**理由**: v1.0 で CSV/TSV 出力は廃止され、JSON のみサポート

**対応方法**: 必要に応じて `--query` (JMESPath) や `jq` コマンドを使用して CSV/TSV 形式に変換

### 2. セレクタの引数化

**対象**: `--selector name=xxx`, `--selector id=xxx`, `--selector tag=xxx`
**変換後**: `xxx` (引数化)
**理由**: v1 で `--selector` オプションは廃止され、引数で指定する形式に変更

**注意**: タグ指定で曖昧性がある場合は実行時エラーが発生する可能性があります

### 3. リソース名の統一

| v0 リソース名 | v1.1 リソース名 | 変更理由 |
|-------------|--------------|---------|
| `iso-image` | `cdrom` | リソース名の統一 |
| `startup-script` | `note` | 名称の整理 |
| `ipv4` | `ipaddress` | IPv4 関連の整理 |

### 4. プロダクトエイリアスの統一

**対象**: `product-disk`, `product-internet`, `product-server`
**変換後**: `disk-plan`, `internet-plan`, `server-plan`
**理由**: v1 系では `-plan` 形式で名称を統一

### 5. 廃止コマンドの処理

**対象**: `usacloud summary`
**処理**: コメントアウト + 代替手段の案内
**代替案**: `bill`, `self`, 各種 `list` コマンドまたは `rest` コマンドの使用

### 6. オブジェクトストレージの非対応

**対象**: `usacloud object-storage`, `usacloud ojs`
**処理**: コメントアウト + 代替手段の案内
**代替案**: S3 互換ツールや Terraform の使用を検討

### 7. ゾーン指定の正規化

**対象**: `--zone = all` (空白あり)
**変換後**: `--zone=all` (空白なし)
**理由**: 記述の統一化

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

**原因**: タグの曖昧性
**対策**: タグが一意に識別できることを事前確認

#### 3. 「出力形式が期待と異なる」

**原因**: JSON 出力への変更
**対策**: `--query` や `jq` を使用した後処理の追加

### ログの活用

```bash
# 詳細な変更ログを保存
usacloud-update --in script.sh --out updated.sh 2> changes.log

# ログの確認
cat changes.log
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

## 関連リンク

- [usacloud v1.0.0 アップグレードガイド](https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
- [usacloud v1.1 リファレンス](https://docs.usacloud.jp/usacloud/)
- [さくらのクラウド API v1.1](https://manual.sakura.ad.jp/cloud-api/1.1/)
- [usacloud GitHub リポジトリ](https://github.com/sacloud/usacloud/)

## サポート

ツールに関する問題や改善要望は、プロジェクトの GitHub Issues で報告してください。

変換結果は必ず事前にテスト環境で確認し、本番環境での使用前に十分な検証を行うことを推奨します。