Feature: Sakura Cloud サンドボックス環境でのusacloud変換コマンド検証

  Background: サンドボックス環境の設定
    Given サンドボックス環境の認証情報が設定されている
    And APIエンドポイントが "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/" に設定されている
    And 対象ゾーンが "tk1v" に設定されている

  Scenario: 基本的なサーバーリスト変換とサンドボックス実行
    Given 以下のusacloud v1.0スクリプトがある:
      """
      #!/bin/bash
      usacloud server list --output-type=csv
      usacloud server list --zone = all
      """
    When usacloud-update --sandbox を実行する
    Then スクリプトが以下のように変換される:
      """
      #!/bin/bash
      usacloud server list --output-type=json # usacloud-update: --output-type=csv は廃止されました
      usacloud server list --zone=all # usacloud-update: スペースを削除しました
      """
    And サンドボックス環境で変換後のコマンドが実行される
    And 実行結果がJSONフォーマットで返される
    And エラーが発生しない

  Scenario: セレクタ廃止コマンドの変換とサンドボックス検証
    Given 以下のusacloud v1.0スクリプトがある:
      """
      usacloud disk read --selector name=mydisk
      usacloud server list --selector name=myserver
      """
    When usacloud-update --sandbox を実行する
    Then スクリプトが以下のように変換される:
      """
      usacloud disk read mydisk # usacloud-update: --selector は廃止されました
      usacloud server list myserver # usacloud-update: --selector は廃止されました
      """
    And サンドボックス環境で "usacloud disk read mydisk" が実行される
    And サンドボックス環境で "usacloud server list myserver" が実行される
    And 実行結果にサンドボックス制約の警告が含まれる

  Scenario: 廃止コマンドのサンドボックス検証（手動対応必要）
    Given 以下のusacloud v0スクリプトがある:
      """
      usacloud summary
      usacloud object-storage list
      """
    When usacloud-update --sandbox を実行する
    Then スクリプトが以下のように変換される:
      """
      # usacloud summary # usacloud-update: summaryコマンドは廃止されました
      # usacloud object-storage list # usacloud-update: object-storageは廃止されました
      """
    And コメントアウトされたコマンドはサンドボックス実行がスキップされる
    And 手動対応が必要な旨がレポートに記載される

  Scenario: インタラクティブTUIでの個別コマンド実行選択
    Given 以下の複数のusacloudコマンドを含むスクリプトがある:
      """
      usacloud server list --output-type=csv
      usacloud disk list --output-type=tsv  
      usacloud iso-image list
      """
    When usacloud-update --sandbox をインタラクティブモードで実行する
    Then TUIインターフェースが表示される
    And 変換されたコマンドの一覧が表示される
    And 各コマンドに対して以下の選択肢が提供される:
      | 実行 | スキップ | 詳細表示 |
    When "usacloud server list --output-type=json" を選択して実行する
    Then サンドボックス環境でコマンドが実行される
    And 実行結果がTUI内に表示される
    And 実行ステータス（成功/失敗）が記録される

  Scenario: サンドボックス環境の制約エラーハンドリング
    Given サンドボックス環境への接続が失敗する状況
    When usacloud-update --sandbox を実行する
    Then 接続エラーが適切にハンドリングされる
    And ユーザーに分かりやすいエラーメッセージが表示される
    And 変換のみを実行するオプションが提示される

  Scenario: 認証情報の検証
    Given 不正な認証情報が設定されている
    When usacloud-update --sandbox を実行する
    Then 認証エラーが検出される
    And 環境変数の設定方法がガイドされる
    And 設定ファイルの作成方法がガイドされる

  Scenario: 大量のコマンドを含むスクリプトのバッチ実行
    Given 10個以上のusacloudコマンドを含むスクリプトがある
    When usacloud-update --sandbox --batch を実行する
    Then 全てのコマンドが自動変換される
    And サンドボックス環境で順次実行される
    And 実行進捗がプログレスバーで表示される
    And 実行結果サマリーが表示される
    And 失敗したコマンドがハイライトされる

  Scenario: ドライランモードでの安全確認
    Given 以下のusacloudスクリプトがある:
      """
      usacloud server create --name test-server --cpu 1 --memory 1
      usacloud disk create --name test-disk --size 20
      """
    When usacloud-update --sandbox --dry-run を実行する
    Then コマンドが変換される
    But サンドボックス環境での実際の実行は行われない
    And 実行予定のコマンドリストが表示される
    And リソース作成の影響が事前に説明される

  Scenario: 非usacloudコマンドの安全な処理
    Given 以下の混在スクリプトがある:
      """
      #!/bin/bash
      echo "Starting deployment"
      usacloud server list --output-type=csv
      docker run nginx
      usacloud disk read --selector name=mydisk
      kubectl apply -f deployment.yaml
      """
    When usacloud-update --sandbox を実行する
    Then usacloudコマンドのみが変換される
    And 非usacloudコマンドは変更されない
    And サンドボックス実行もusacloudコマンドのみが対象となる
    And 変換対象外のコマンドがレポートに記載される