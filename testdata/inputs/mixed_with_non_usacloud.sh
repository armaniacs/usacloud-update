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
usacloud server list --output-type=csv
usacloud disk read --selector name=mydisk
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 read --zone tk1a --ipaddress 203.0.113.10
usacloud product-disk list
usacloud summary
usacloud object-storage list
usacloud server list --zone = all

# 混在行のテスト
echo "Before usacloud command"
usacloud server list --output-type=tsv
echo "After usacloud command"

# 複雑なケース
other-tool --output-type=csv && echo "done"
usacloud server list --output-type=csv && echo "usacloud done"