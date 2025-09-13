# PBI-025: Profile Export/Import機能修復

## 概要
ProfileCommandのExportProfile関数で引数処理エラーが発生しており、テストが失敗している問題を修復する。現在、プロファイルエクスポート機能が期待される引数形式と実装が不一致となっており、ユーザビリティに重大な影響を与えている。

## 受け入れ条件
- [ ] ExportProfile関数が正しい引数形式で動作すること
- [ ] フラグベースの出力ファイル指定が正常に機能すること
- [ ] 既存のプロファイル管理機能に影響を与えないこと
- [ ] Export/Import機能の全テストが通過すること

## 技術仕様

### 現在の問題
```bash
# テスト失敗の詳細
=== RUN   TestProfileCommand_ExportProfile
    commands_test.go:334: ExportProfile() failed: プロファイルIDまたは名前と出力ファイルパスを指定してください
--- FAIL: TestProfileCommand_ExportProfile (0.00s)
```

### 1. 引数処理方式の修正
#### 現在の実装（問題のあるコード）
```go
// internal/config/profile/commands.go:371-374
func (pc *ProfileCommand) ExportProfile(cmd *cobra.Command, args []string) error {
    if len(args) < 2 {
        return fmt.Errorf("プロファイルIDまたは名前と出力ファイルパスを指定してください")
    }
    // ...
    outputFile := args[1]  // 2番目の引数を期待
}
```

#### 修正後の実装
```go
// フラグベースの出力ファイル指定に変更
func (pc *ProfileCommand) ExportProfile(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("プロファイルIDまたは名前を指定してください")
    }
    
    // --output フラグから出力ファイルを取得
    outputFile, _ := cmd.Flags().GetString("output")
    if outputFile == "" {
        return fmt.Errorf("出力ファイルパス（--output）を指定してください")
    }
    
    profile, err := pc.manager.GetProfile(args[0])
    if err != nil {
        return err
    }
    
    if err := pc.manager.ExportProfile(profile.ID, outputFile); err != nil {
        return fmt.Errorf("プロファイルをエクスポートできませんでした: %w", err)
    }
    
    fmt.Printf("プロファイル '%s' を '%s' にエクスポートしました。\n", profile.Name, outputFile)
    return nil
}
```

### 2. テスト修正
#### 現在のテスト（修正が必要）
```go
// commands_test.go:333
err = pc.ExportProfile(cmd, []string{profile.ID})  // 1つの引数のみ
```

#### 修正後のテスト
```go
func TestProfileCommand_ExportProfile(t *testing.T) {
    // ... セットアップコード ...
    
    // --output フラグを設定
    cmd := &cobra.Command{}
    cmd.Flags().String("output", exportFile, "Output file")
    
    // プロファイルIDのみを引数として渡す
    err = pc.ExportProfile(cmd, []string{profile.ID})
    if err != nil {
        t.Fatalf("ExportProfile() failed: %v", err)
    }
    
    // エクスポートファイルの存在確認
    if _, err := os.Stat(exportFile); os.IsNotExist(err) {
        t.Errorf("Export file was not created")
    }
}
```

### 3. CLI統合の修正
#### CLIコマンド定義の修正
```go
// 既存のCLIコマンド定義で--outputフラグの追加が必要
exportCmd := &cobra.Command{
    Use:   "export <profile-id>",
    Short: "プロファイルをエクスポート",
    RunE:  profileCmd.ExportProfile,
}
exportCmd.Flags().StringP("output", "o", "", "出力ファイルパス（必須）")
exportCmd.MarkFlagRequired("output")
```

## テスト戦略
- **ユニットテスト**: ExportProfile/ImportProfile関数の個別テスト
- **統合テスト**: CLIコマンドとしての実行テスト
- **エラーハンドリングテスト**: 不正な引数・フラグでのエラー処理確認

## 依存関係
- 前提PBI: なし（独立した修復タスク）
- 関連PBI: PBI-026（Profile FileStorage機能修復）
- 既存コード: internal/config/profile/ パッケージ

## 見積もり
- 開発工数: 4時間
  - 引数処理方式の修正: 1.5時間
  - テスト修正: 1.5時間
  - CLI統合修正: 1時間

## 完了の定義
- [ ] ExportProfile関数が--outputフラグで正常動作
- [ ] ImportProfile機能の整合性確認
- [ ] 全Profile Export/Importテストが通過
- [ ] CLIコマンドとしての動作確認完了
- [ ] 既存Profile機能の回帰テスト通過

## 備考
- ユーザビリティを考慮し、引数指定よりもフラグ指定の方が直感的
- 他のCLIコマンドとの一貫性を保つため、--outputフラグ方式を採用
- Export/Import機能は設定の共有・バックアップに重要な機能

---

## 実装状況 (2025-09-11)

🟠 **PBI-025は未実装** (2025-09-11)

### 現在の状況
- ProfileCommandのExportProfile関数で引数処理エラーが発生
- テストが失敗してプロファイルエクスポート機能が動作不能
- フラグベースの出力ファイル指定が実装されていない
- CLIコマンドとしての統合機能が不完全

### 未実装要素
1. **引数処理方式の修正**
   - 現在の2引数方式からフラグベース方式への変更
   - --outputフラグによる出力ファイル指定の実装
   - 引数検証ロジックの修正

2. **Export/Import機能の整合性確保**
   - ExportProfile関数の正常動作の確立
   - ImportProfile機能の整合性確認
   - 設定の共有・バックアップ機能の確立

3. **テスト修正と統合**
   - TestProfileCommand_ExportProfileの修正
   - CLIコマンドとしての実行テスト
   - エラーハンドリングテストの実装
   - 既存Profile機能の回帰テスト

### 次のステップ
1. internal/config/profile/commands.goの引数処理方式修正
2. --outputフラグによる出力ファイル指定の実装
3. テストケースの修正と検証
4. CLIコマンドとしての動作確認
5. 既存Profile機能の回帰テスト実行

### 技術要件
- Go 1.24.1対応
- Cobraフレームワークのフラグ処理機能活用
- Profile Export/Import機能の完全実装
- 4時間の作業見積もり

### 受け入れ条件の進捗
- [ ] ExportProfile関数が正しい引数形式で動作すること
- [ ] フラグベースの出力ファイル指定が正常に機能すること
- [ ] 既存のプロファイル管理機能に影響を与えないこと
- [ ] Export/Import機能の全テストが通過すること