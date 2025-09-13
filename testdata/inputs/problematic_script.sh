#!/bin/bash
# 問題のあるスクリプトサンプル
# 厳格モードでテストするためのスクリプト

# タイポエラー
usacloud serv list
usacloud databse list

# 廃止コマンドエラー
usacloud iso-image list
usacloud startup-script list

# 無効なオプション
usacloud server list --invalid-option
usacloud disk create --nonexistent-flag

# 非推奨オプション
usacloud server list --output-type csv
usacloud disk list --output-type tsv

# 構文エラー
usacloud server list --zone = invalid-zone
usacloud disk create --size=

# 複数エラーの組み合わせ
usacloud serv lst --invalid-option --output-type csv