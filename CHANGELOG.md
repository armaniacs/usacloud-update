# 変更履歴

このプロジェクトの重要な変更はすべてこのファイルに記録されます。

フォーマットは [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) に基づき、
このプロジェクトは [セマンティック バージョニング](https://semver.org/spec/v2.0.0.html) に準拠します。

## [1.2.0] - 2025-09-07

### 追加
- CLIヘルプメッセージの大幅改善:
  - ツールの概要と目的の説明を追加
  - 4つの実用的な使用例を追加
  - 日本語でのオプション説明に変更
  - README-Usage.mdへの案内を追加

### 修正
- **重要**: 非usacloud行の誤変換防止:
  - `--output-type=csv/tsv` ルールをusacloud文脈に限定
  - `--zone = all` ルールをusacloud文脈に限定
  - 他のツール（python, curl, docker等）のオプションが意図せず変更される問題を解決

### テスト
- 非usacloud行を含むテストケースを追加:
  - `testdata/mixed_with_non_usacloud.sh` - 混在した入力テストケース
  - `testdata/expected_mixed_non_usacloud.sh` - 期待される結果
  - `make verify-mixed` - 新しい検証コマンド
- 既存テストの回帰確認済み

### 技術的改善
- より安全な正規表現パターンの採用
- 包括的なテストカバレッジの実現
- ツールの安全性と信頼性の向上

## [1.1.0] - 2025-09-06

### 変更点
- **破壊的変更**: バイナリとディレクトリを `sacloud-update` から `usacloud-update` に名称変更
  - `cmd/sacloud-update/` → `cmd/usacloud-update/`
  - バイナリ名: `sacloud-update` → `usacloud-update`
  - Makefile の BINARY 変数を更新
  - インストールコマンドを変更: `go install github.com/armaniacs/usacloud-update/cmd/usacloud-update@latest`
- すべてのコメントマーカーを `# sacloud-update:` から `# usacloud-update:` に変更
- 生成ヘッダーで `usacloud-update` ツール名を使用するよう更新
- 新しい命名規則を反映するようにすべてのドキュメントファイルを更新

### 追加
- `/ref` ディレクトリに包括的なリファレンスドキュメントを追加:
  - `implementation-reference.md` - 詳細な技術実装ガイド
  - `project-dependencies.md` - 依存関係分析とプラットフォームサポート
  - `test-data-reference.md` - テストデータ構造と検証プロセス
  - `build-deployment.md` - ビルドシステムとデプロイメントガイド
- すべてのリファレンスドキュメントへのポインタでCLAUDE.mdを拡張

### 移行ガイド
既存のインストールやスクリプトで `sacloud-update` を参照している場合:

1. **バイナリのインストール**: 
   ```bash
   # 古いコマンド（もう動作しません）
   go install github.com/armaniacs/usacloud-update/cmd/sacloud-update@latest
   
   # 新しいコマンド
   go install github.com/armaniacs/usacloud-update/cmd/usacloud-update@latest
   ```

2. **バイナリの使用**:
   ```bash
   # 古いバイナリ名
   sacloud-update --help
   
   # 新しいバイナリ名
   usacloud-update --help
   ```

3. **生成されるコメント**: 
   - 出力コメントは `# sacloud-update:` の代わりに `# usacloud-update:` を表示します
   - この変更は新しく変換されるすべてのファイルに影響します

## [1.0.0] - 2025-09-06

### 追加
- usacloudコマンド変換ツールの初回リリース
- usacloud v0.x および v1.0 スクリプトを v1.1 互換に変換するサポート
- 9つのカテゴリの変換ルール:
  1. 出力フォーマット移行 (csv/tsv → json)
  2. セレクタ廃止 (--selector → 引数)
  3. リソース名変更 (iso-image, startup-script, ipv4)
  4. プロダクトエイリアス整理 (product-* → *-plan)
  5. コマンド廃止 (summary)
  6. サービス廃止 (object-storage)
  7. パラメータ正規化 (--zone スペース)
- --in, --out, --stats フラグを持つ CLI インターフェース
- ゴールデンファイルテストフレームワーク
- 包括的なドキュメントと使用ガイド
- 開発とテスト用の Makefile ターゲット

### 機能
- 行ごとの変換処理
- ドキュメントリンク付きの自動説明コメント
- 変換フィードバック用のカラー端末出力
- 大きなファイルのサポート（1MBバッファ）
- クロスプラットフォーム互換性（Linux、macOS、Windows）
- ランタイム依存関係なしの単一バイナリ配布

### 技術詳細
- Go 1.24.1 互換性
- 最小限の依存関係（端末出力用の fatih/color のみ）
- 正規表現ベースの変換エンジン
- 将来の拡張のための拡張可能なルールシステム
- 実世界の例による完全なテストカバレッジ