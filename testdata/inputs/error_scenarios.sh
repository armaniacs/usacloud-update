#!/bin/bash
# エラーシナリオテスト用サンプル
# 様々なエラーパターンを含むスクリプト

# === 廃止コマンドエラー ===
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 list
usacloud product-disk list
usacloud summary
usacloud object-storage list

# === 非推奨オプション警告 ===
usacloud server list --output-type csv
usacloud disk list --output-type tsv
usacloud database list --output-type csv

# === 存在しないコマンドエラー ===
usacloud invalid-command list
usacloud server invalid-subcommand
usacloud xyz abc
usacloud nonexistent action

# === タイポエラー ===
usacloud serv list
usacloud databse list
usacloud dsk create
usacloud server lst

# === 無効なオプションエラー ===
usacloud server list --invalid-option
usacloud disk create --nonexistent-flag
usacloud database list --fake-parameter value

# === 構文エラー ===
usacloud server list --zone = invalid  # スペース付きの無効な構文
usacloud disk create --size=  # 値なしのオプション
usacloud database --  # 不完全なコマンド

# === 複数エラーの組み合わせ ===
usacloud serv lst --invalid-option  # タイポ + 無効オプション
usacloud iso-imag list --output-type csv  # 廃止コマンドタイポ + 非推奨オプション
usacloud invalid-cmd invalid-sub --fake-flag  # 存在しないコマンド + サブコマンド + オプション

# === パラメータエラー ===
usacloud server create  # 必須パラメータなし
usacloud disk create --size invalid-size  # 無効な値
usacloud database create --type unknown-type  # 無効なタイプ

# === ゾーン関連エラー ===
usacloud server list --zone invalid-zone
usacloud disk create --zone nonexistent
usacloud database list --zone fake-zone

# === 出力形式エラー ===
usacloud server list --output-type xml  # サポートされていない形式
usacloud disk list --output-type yaml   # サポートされていない形式

# === 長すぎるコマンドライン ===
usacloud server list --very-long-argument-name-that-exceeds-normal-limits-and-should-cause-validation-errors

# === 文字エンコーディング問題 ===
usacloud server list --tags "テスト用タグ"  # 日本語文字
usacloud database create --name "データベース名"  # 日本語名

# === 引用符の問題 ===
usacloud server create --name "unclosed quote
usacloud disk create --name 'mixed quotes"
usacloud database create --name unquoted space name

# === 特殊文字 ===
usacloud server list --tags "tag with spaces and !@#$%^&*()"
usacloud disk create --name "disk/with/slashes"
usacloud database create --name "db|with|pipes"