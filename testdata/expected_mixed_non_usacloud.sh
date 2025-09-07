# Updated for usacloud v1.1 by usacloud-update — DO NOT EDIT ABOVE THIS LINE
#!/usr/bin/env bash
set -euo pipefail

# 非usacloud行（変更されないべき）
echo "Hello World"
ls -la /tmp
MY_VAR="some value"
python script.py --output-type=csv
other-tool --output-type=csv
curl https://example.com --zone = all
docker run --name test --zone=all image

# 通常のコメント
# This is just a regular comment

# usacloud行（変更されるべき）
usacloud server list --output-type=json # usacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
usacloud disk read mydisk # usacloud-update: --selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
usacloud cdrom list # usacloud-update: v1ではリソース名がcdromに統一 (https://manual.sakura.ad.jp/cloud-api/1.1/cdrom/index.html)
usacloud note list # usacloud-update: v1ではstartup-scriptはnoteに統一 (https://docs.usacloud.jp/usacloud/)
usacloud ipaddress read --zone tk1a --ipaddress 203.0.113.10 # usacloud-update: v1ではIPv4関連はipaddressに整理 (https://docs.usacloud.jp/usacloud/references/ipaddress/)
usacloud disk-plan list # usacloud-update: v1系では *-plan へ名称統一 (https://docs.usacloud.jp/usacloud/)
# usacloud summary # usacloud-update: summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
# usacloud object-storage list # usacloud-update: v1ではオブジェクトストレージ操作は非対応方針。S3互換ツール/他プロバイダやTerraformを検討 (https://github.com/sacloud/usacloud/issues/585)
usacloud server list --zone=all # usacloud-update: 全ゾーン一括操作は --zone=all を推奨 (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)

# 混在行のテスト
echo "Before usacloud command"
usacloud server list --output-type=json # usacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
echo "After usacloud command"

# 複雑なケース
other-tool --output-type=csv && echo "done"
usacloud server list --output-type=json && echo "usacloud done" # usacloud-update: v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください (https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/)
