# saclou-dupdate についての実装メモ

## usacloud-update という goのコマンドを作るにあたって使用するもの

https://github.com/sacloud/usacloud が現行のusacloud v1です。
https://github.com/sacloud/usacloud/tree/v0-backup/ が一つ前のv0です。

この間には大きな差異があります。
v1.0へのアップグレード - Usacloudドキュメント https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/
v1.1へのアップグレード - Usacloudドキュメント https://docs.usacloud.jp/usacloud/upgrade/v1_1_0/
よんで、v0, v1.0, v1.1 それぞれの記述が混じったbashスクリプトが与えられた場合に v1.1で動作するようにスクリプトを書き換えるプログラムをつくります。

書き換える理由はスクリプトの中にコメントの形で追加します。

usacloud-update は usacloud v1.1と同じく、go言語で実装されます。

## usacloud-update (Go CLI)

目的: `usacloud` を呼び出すBashスクリプト中の v0 / v1.0 / v1.1混在の記述を解析し、v1.1で動作する内容へ自動変換します。変換理由は行コメント(`# usacloud-update:`)として残します。

---


