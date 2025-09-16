# PBI-028: CLI共通ヘルパーパッケージの作成とヘルプメッセージ統一

## ユーザーストーリー
開発者として、ヘルプメッセージ生成ロジックを一元管理したい、なぜならmain.goとcobra_main.goで同じ内容を別々に管理するのは非効率だから

## ビジネス価値
- **保守性向上**: ヘルプメッセージの変更時に1箇所の修正で済む（現在2箇所必要）
- **一貫性確保**: バージョン表示や情報内容の一貫性を保証
- **開発効率**: 新しいCLIオプション追加時の作業効率向上

## BDD受け入れシナリオ

```gherkin
Scenario: 通常のヘルプメッセージ表示
  Given usacloud-updateコマンドが実行可能な状態
  When ユーザーが"usacloud-update --help"を実行する
  Then 統一されたヘルプメッセージが表示される
  And バージョン情報が含まれている
  And 使用方法が日本語で表示される

Scenario: Cobraコマンドでのヘルプ表示
  Given usacloud-updateコマンドが実行可能な状態
  When ユーザーが"usacloud-update help"を実行する
  Then 同じ内容のヘルプメッセージが表示される
  And 旧版と同じフォーマットが維持される

Scenario: 無効なオプション指定時のエラーヘルプ
  Given usacloud-updateコマンドが実行可能な状態
  When ユーザーが"usacloud-update --invalid-option"を実行する
  Then 「無効なオプションが指定されました」エラーが表示される
  And ヘルプメッセージが続けて表示される
```

## 受け入れ基準
- [ ] `internal/cli/helpers/help.go`パッケージが作成されている
- [ ] `GetHelpContent(version string) string`関数が実装されている
- [ ] `GetOptionsContent() string`関数が実装されている
- [ ] `GetFooterContent() string`関数が実装されている
- [ ] main.goから重複するヘルプ関数が削除されている
- [ ] cobra_main.goから重複するヘルプ関数が削除されている
- [ ] 両方のエントリポイントで同じヘルプが表示される
- [ ] 既存のテストが全て通る

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- CLI実行でのヘルプメッセージ表示テスト
- 旧版との出力比較テスト

統合テスト:
- ヘルパー関数の組み合わせテスト
- バージョン情報埋め込みテスト

単体テスト:
- 各ヘルパー関数の出力内容テスト
- 改行・フォーマット正確性テスト
- 文字エンコーディングテスト
```

## 実装アプローチ
- **Outside-In**: CLI実行テストから開始
- **Red-Green-Refactor**: 既存動作を保持しながらリファクタリング
- **段階的移行**: 関数作成→main.go移行→cobra_main.go移行

## 見積もり
2ストーリーポイント

## 技術的考慮事項
- **依存関係**: なし（Pure Go関数）
- **テスタビリティ**: 入出力明確な純粋関数として設計
- **パフォーマンス**: 文字列処理のため軽量

## Definition of Done
- [ ] 受け入れシナリオが全て通る
- [ ] 既存テストが全て通る（`make test`）
- [ ] ヘルプメッセージの内容に変更がない
- [ ] コードレビュー完了
- [ ] リファクタリング完了（重複削除）
- [ ] ドキュメント更新（CLAUDE.md）

## 実装詳細
1. `internal/cli/helpers/`ディレクトリ作成
2. `help.go`ファイル作成
3. 既存の`getHelpContent`, `getOptionsContent`, `getFooterContent`を移動
4. パッケージimportをmain.go, cobra_main.goに追加
5. 重複する関数定義を削除
6. テスト実行で動作確認

## 関連PBI
- PBI-029: エラーメッセージフォーマットの統一
- PBI-030: ファイルI/O処理の共通化