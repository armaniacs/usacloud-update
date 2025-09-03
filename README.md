# sacloud-update

sacloudの v0, v1.0, v1.1 それぞれの記述が混じったbashスクリプトが与えられた場合に v1.1で動作するようにスクリプトを書き換えるプログラムです。

## Usage

cat input.sh | sacloud-update > output.sh
sacloud-update --in input.sh --out output.sh

## Notes
- 自動変換できない箇所は元行をコメントアウトし、**手動対応**を促すコメントを付与します。
- 引数化(`--selector`)は `name= / id= / tag=` の右辺をそのまま引数へ移します。
  - タグ指定の形式はv1で引数サポート。曖昧性がある場合は実行時エラーになります。
- CSV/TSVは`--output-type=json`へ統一。必要に応じて`--query`(JMESPath/jq)を併用してください。
