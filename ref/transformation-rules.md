# 変換ルール詳細

## 概要

sacloud-update は usacloud v0.x から v1.1 への移行を自動化するため、9つのカテゴリの変換ルールを適用します。各ルールは公式ドキュメントへの参照と共に、説明的なコメントを自動付与します。

## ルール一覧

### 1. 出力形式の変換

**ルール名**: `output-type-csv-tsv`

**変換内容**: 
- `--output-type=csv` → `--output-type=json`
- `--output-type=tsv` → `--output-type=json`
- `-o csv` → `-o json`
- `-o tsv` → `-o json`

**理由**: v1.0 で csv/tsv 出力は廃止され、JSON のみサポート

**例**:
```bash
# 変換前
usacloud server list --output-type=csv

# 変換後  
usacloud server list --output-type=json # sacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください
```

### 2. セレクタの引数化

**ルール名**: `selector-to-arg`

**変換内容**: `--selector` オプションをコマンド引数に変換

**処理ロジック**:
- `name=xxx` → `xxx`
- `id=xxx` → `xxx` 
- `tag=xxx` → `xxx`
- その他の形式もそのまま引数化

**例**:
```bash
# 変換前
usacloud disk read --selector name=mydisk

# 変換後
usacloud disk read mydisk # sacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行
```

### 3-5. リソース名の統一

#### 3. ISO イメージ → CD-ROM

**ルール名**: `iso-image-to-cdrom`
- `usacloud iso-image` → `usacloud cdrom`

#### 4. スタートアップスクリプト → ノート  

**ルール名**: `startup-script-to-note`
- `usacloud startup-script` → `usacloud note`

#### 5. IPv4 → IP アドレス

**ルール名**: `ipv4-to-ipaddress`
- `usacloud ipv4` → `usacloud ipaddress`

### 6. プロダクト エイリアスの整理

**ルール名**: `product-alias-*`

**変換対象**:
- `product-disk` → `disk-plan`
- `product-internet` → `internet-plan`  
- `product-server` → `server-plan`

**例**:
```bash
# 変換前
usacloud product-disk list

# 変換後
usacloud disk-plan list # sacloud-update: v1系では *-plan へ名称統一
```

### 7. 廃止コマンドの処理

**ルール名**: `summary-removed`

**処理**: `usacloud summary` コマンドをコメントアウト

**例**:
```bash
# 変換前
usacloud summary

# 変換後  
# usacloud summary # sacloud-update: summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください
```

### 8. オブジェクトストレージの非対応

**ルール名**: `object-storage-removed-*`

**対象エイリアス**:
- `object-storage`
- `ojs`

**処理**: 該当コマンドをコメントアウト

**メッセージ**: v1 ではオブジェクトストレージ操作は非対応。S3互換ツール等の検討を推奨

### 9. ゾーン指定の正規化

**ルール名**: `zone-all-normalize`

**変換内容**: `--zone = all` → `--zone=all` (空白の除去)

## ルール適用の仕組み

### 順次適用

ルールは `DefaultRules()` で定義された順序で適用されます。この順序は重要で、先に適用されたルールの結果が後続のルールの入力になります。

### コメント自動付与

各変換には以下の形式のコメントが自動付与されます:

```
# sacloud-update: [変更理由] ([公式ドキュメントURL])
```

### 重複処理の防止

既に `# sacloud-update:` コメントが含まれる行には追加コメントを付与しません。

## 変換例

### サンプル入力 (`testdata/sample_v0_v1_mixed.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

# v0風: csv/tsv
usacloud server list --output-type=csv

# v0風: selector  
usacloud disk read --selector name=mydisk
usacloud server delete --selector tag=to-be-removed

# v0風: リソース名
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 read --zone tk1a --ipaddress 203.0.113.10

# v0: product-*
usacloud product-disk list

# 廃止コマンド
usacloud summary
```

### 変換後出力 (`testdata/expected_v1_1.sh`)

```bash
# Updated for usacloud v1.1 by sacloud-update — DO NOT EDIT ABOVE THIS LINE
#!/usr/bin/env bash
set -euo pipefail

# v0風: csv/tsv
usacloud server list --output-type=json # sacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# v0風: selector
usacloud disk read mydisk # sacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
usacloud server delete to-be-removed # sacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# v0風: リソース名  
usacloud cdrom list # sacloud-update: v1ではリソース名がcdromに統一 (https://manual.sakura.ad.jp/cloud-api/1.1/cdrom/index.html)
usacloud note list # sacloud-update: v1ではstartup-scriptはnoteに統一 (https://docs.usacloud.jp/usacloud/)
usacloud ipaddress read --zone tk1a --ipaddress 203.0.113.10 # sacloud-update: v1ではIPv4関連はipaddressに整理 (https://docs.usacloud.jp/usacloud/references/ipaddress/)

# v0: product-*
usacloud disk-plan list # sacloud-update: v1系では *-plan へ名称統一 (https://docs.usacloud.jp/usacloud/)

# 廃止コマンド
# usacloud summary # sacloud-update: summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
```

## 新しいルールの追加方法

1. `ruledefs.go` の `DefaultRules()` 関数に以下を追加:

```go
rules = append(rules, mk(
    "ルール名",
    `正規表現パターン`,
    func(m []string) string { 
        // 置換ロジック
        return 置換後文字列
    },
    "変更理由の説明",
    "https://参考URL",
))
```

2. テストケースを `testdata/` に追加
3. `make golden` でゴールデンファイルを更新