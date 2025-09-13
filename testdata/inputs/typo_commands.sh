#!/bin/bash
# タイポコマンドテスト用サンプル
# 様々なタイポパターンを含むスクリプト

# 一般的なタイポパターン
usacloud serv list  # server のタイポ
usacloud databse list  # database のタイポ
usacloud dsk list  # disk のタイポ

# サブコマンドのタイポ
usacloud server lst  # list のタイポ
usacloud server crete --name test  # create のタイポ
usacloud disk updat 123456  # update のタイポ

# オプションのタイポ
usacloud server list --output-tyep json  # --output-type のタイポ
usacloud server list --zon is1a  # --zone のタイポ
usacloud server list --tgas web  # --tags のタイポ

# 文字順序の入れ替え
usacloud sevrer list  # server の文字入れ替え
usacloud sever list   # server の文字欠落
usacloud serverr list # server の文字重複

# 廃止コマンドのタイポ（二重のエラー）
usacloud iso-imag list  # iso-image のタイポ（かつ廃止）
usacloud startup-scritp list  # startup-script のタイポ（かつ廃止）

# 大文字小文字の混在
usacloud Server List
usacloud DATABASE list
usacloud Disk CREATE

# 余分なスペース
usacloud  server  list
usacloud server  list  --zone  is1a

# ハイフンとアンダースコアの混同
usacloud server_list
usacloud disk-create
usacloud database_read

# よくある英語のタイポ
usacloud databse list  # database
usacloud sever list    # server  
usacloud netwrok list  # network (存在しないが)

# 日本語キーボードでの典型的なタイプミス
usacloud serve list    # r と e の隣接
usacloud datavase list # b と v の隣接
usacloud lsit          # i と s の順序

# 省略形のタイポ
usacloud srv list      # server の省略形タイポ
usacloud db list       # database の省略形タイポ
usacloud vm list       # virtual machine（usacloud では存在しない）