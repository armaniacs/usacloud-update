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

# 非サポート(object-storage)
usacloud object-storage list

# v1.0以降: allゾーン
usacloud server list --zone = all
