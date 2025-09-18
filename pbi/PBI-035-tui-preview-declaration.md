# PBI-035: TUI Preview機能宣言と表示実装

## ユーザーストーリー
**CLIユーザー**として、**TUI機能が現在Preview段階であることが明確に分かる表示**がほしい、なぜなら**機能の成熟度を正しく理解し、適切な期待値でツールを利用したい**から

## ビジネス価値
- **透明性向上**: 機能の開発段階をユーザーに明確に伝える（期待値管理）
- **フィードバック促進**: Preview表示によりユーザーからの改善提案を促す（品質向上に貢献）
- **段階的リリース戦略の明示**: v1.9.x Preview → v2.0.0 正式版のロードマップ共有

## BDD受け入れシナリオ

```gherkin
Feature: TUI Preview機能宣言表示
  CLIユーザーとして
  TUI機能がPreview段階であることを認識したい
  なぜなら適切な期待値で機能を利用したいから

Scenario: TUIモードでPreview表示を確認
  Given usacloud-updateがインストールされている
  And バージョンがv1.9.6である
  When ユーザーがTUIモードで起動する
  Then 画面最下行中央に「TUIはPreviewとして提供中」が表示される
  And 表示は黄色背景で視認性が高い
  And ウィンドウサイズに関わらず常に中央配置される

Scenario: ヘルプ表示時もPreview表示維持
  Given TUIモードが起動している
  When ユーザーが「?」キーでヘルプを表示する
  Then Preview表示が最下行に維持される
  And ヘルプパネルと重ならない

Scenario: ファイル選択後もPreview表示維持
  Given TUIモードでファイル選択画面が表示されている
  When ユーザーがファイルを選択して実行する
  Then Preview表示が消えずに維持される
```

## 受け入れ基準
- [ ] 画面最下行に「TUIはPreviewとして提供中」が常時表示される
- [ ] テキストは中央配置され、ウィンドウリサイズに追従する
- [ ] 黄色背景・黒文字で視認性を確保
- [ ] 既存のUI要素（ステータスバー、ヘルプ）と干渉しない
- [ ] バージョンがv1.9.6に更新される
- [ ] CHANGELOGにPreview機能であることが明記される

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- TUI起動時のPreview表示確認テスト
- ウィンドウリサイズ時の中央配置維持テスト

統合テスト:
- GridレイアウトへのPreview行追加テスト
- 既存UI要素との共存テスト
- テキスト配置計算の検証

単体テスト:
- Preview表示テキスト生成のテスト
- 中央配置計算ロジックのテスト
- カラー設定の適用テスト
```

## 実装アプローチ
- **Outside-In**: E2EテストでPreview表示の存在を確認するところから開始
- **Red-Green-Refactor**: テスト失敗→実装→リファクタリングのサイクル
- **リファクタリング**: マジックナンバーの定数化、表示ロジックの抽出

## 技術仕様

### 実装変更点
1. **レイアウト更新** (`internal/tui/file_selector.go`)
   ```go
   // setupLayout内でPreview表示行を追加
   fs.mainGrid.SetRows(0, 1, 1, 1) // 最下行にPreview用の行を追加

   // Preview表示用TextView作成
   fs.previewNotice = tview.NewTextView().
       SetText("[black:yellow:b] TUIはPreviewとして提供中 [::-]").
       SetTextAlign(tview.AlignCenter).
       SetDynamicColors(true)
   ```

2. **バージョン更新** (`cmd/usacloud-update/main.go`)
   ```go
   const version = "1.9.6"
   ```

3. **CHANGELOG更新**
   - v1.9.6リリースノートにTUI Preview機能宣言を明記
   - TUIはv1.9.x開発版でのPreview機能であることを明示
   - v2.0.0での正式版リリース予定を記載

## 見積もり
**1ストーリーポイント**（0.5日）
- レイアウト実装: 2時間
- テスト実装: 1時間
- ドキュメント更新: 1時間

## 技術的考慮事項

### UI/UX
- 黄色背景で注意喚起しつつ、作業の邪魔にならない
- 最下行配置で常に視界に入るが、主要操作領域と干渉しない

### テスタビリティ
- Preview表示のモック化対応
- E2Eテストでの表示確認方法の確立

### 後方互換性
- 既存のTUI機能に影響を与えない
- 環境変数による表示ON/OFF制御の検討

## Definition of Done
- [ ] BDD受け入れシナリオが全て通る
- [ ] E2E/統合/単体テストが全て通る
- [ ] make testが100%成功
- [ ] コードレビュー完了
- [ ] バージョンv1.9.6への更新完了
- [ ] CHANGELOGにPreview宣言記載
- [ ] リファクタリング完了

## 実装予定ファイル
- `internal/tui/file_selector.go` - Preview表示追加
- `internal/tui/preview_notice_test.go` - Preview表示のテスト
- `cmd/usacloud-update/main.go` - バージョン更新
- `CHANGELOG.md` - リリースノート更新

## 開発プロセス（BDD×TDD統合）
1. **BDDシナリオ確認**: Preview表示要件の理解
2. **E2Eテスト実装**: TUI起動時のPreview表示検証
3. **Outside-In TDD**: 表示から内部実装へ
4. **継続的リファクタリング**: 表示ロジックの最適化
5. **シナリオ検証**: 実際のTUIでのユーザー体験確認

## 品質保証メトリクス
- **ビヘイビアカバレッジ**: 3/3シナリオ実装
- **テストピラミッド比率**: E2E:統合:単体 = 2:3:5
- **ユーザビリティ**: Preview表示の視認性100%

---

**作成日**: 2025-09-18
**手法**: ryuzee × BDD × t_wada統合メソッド
**ステータス**: 📋 **Ready** - 実装準備完了