#!/bin/bash

# 有効なコマンド
usacloud server list

# 無効なメインコマンド 
usacloud invalidcommand list

# 無効なサブコマンド
usacloud server invalidaction

# 廃止コマンド
usacloud iso-image list

# 通常のシェルコマンド（検証対象外）
echo "Hello World"
ls -la

# 複数の問題を持つ行
usacloud iso-imag lst