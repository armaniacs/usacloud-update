# PBI-024: CLI サブコマンド拡張

## 概要
現在のフラグベースCLIインターフェースを、より直感的で使いやすいサブコマンドベースのアーキテクチャに拡張します。`usacloud-update --sandbox` の代わりに `usacloud-update sandbox` のような自然なコマンド体系を提供し、ユーザビリティを向上させます。

## 受け入れ条件
- [x] `usacloud-update version` コマンドが動作し、バージョン情報を表示する
- [x] `usacloud-update validate <file>` コマンドが動作し、ファイル検証を実行する
- [x] `usacloud-update config` コマンドが動作し、設定管理を実行する
- [x] `usacloud-update sandbox <file>` コマンドが動作し、サンドボックス実行を行う
- [x] 既存のフラグベースインターフェースが後方互換性を保持する

## 技術仕様

### アーキテクチャ設計
現在の cobra 基盤を拡張し、以下の構造でサブコマンドを実装：

```go
rootCmd
├── versionCmd    // usacloud-update version
├── validateCmd   // usacloud-update validate
├── configCmd     // usacloud-update config
└── sandboxCmd    // usacloud-update sandbox
```

#### 1. Version サブコマンド
```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "バージョン情報を表示",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("usacloud-update version %s\n", version)
    },
}
```

#### 2. Validate サブコマンド
```go
var validateCmd = &cobra.Command{
    Use:   "validate [file]",
    Short: "スクリプトファイルの検証を実行",
    Args:  cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // 既存の validateOnly ロジックを呼び出し
    },
}
```

#### 3. Config サブコマンド
```go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "設定ファイルの管理",
    Run: func(cmd *cobra.Command, args []string) {
        // 設定ファイル表示・作成・編集機能
    },
}
```

#### 4. Sandbox サブコマンド  
```go
var sandboxCmd = &cobra.Command{
    Use:   "sandbox [file]",
    Short: "サンドボックス環境での実行",
    Args:  cobra.MaximumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // 既存の sandbox ロジックを呼び出し
    },
}
```

## テスト戦略
- **Unit Tests**: 各サブコマンドの cobra.Command 構造テスト
- **E2E Tests**: 実際のコマンド実行とOutput検証  
- **Compatibility Tests**: 既存フラグとの併用・互換性テスト
- **Integration Tests**: サブコマンド間の相互作用テスト

## 依存関係
- 前提PBI: なし（独立実装可能）
- 関連PBI: PBI-007 (BDD Testing), PBI-010 (Help System)
- 既存コード: `cmd/usacloud-update/cobra_main.go`, `runMainLogic()`

## 見積もり
- 開発工数: 125分
  - PBI作成: 15分
  - Version実装: 15分  
  - Validate実装: 17分
  - Config実装: 20分
  - Sandbox実装: 25分
  - E2Eテスト: 25分
  - 互換性確認: 8分

## 完了の定義
- [x] 全4つのサブコマンドが正常に動作する
- [x] E2Eテストが全て通過する
- [x] 既存のフラグベースインターフェースが引き続き動作する
- [x] コードレビューが完了している
- [x] ドキュメントが更新されている

## 備考
- 既存の `runMainLogic()` 関数を最大限再利用し、重複実装を避ける
- サブコマンドのヘルプメッセージは日本語で統一する
- 段階的実装により、各ステップでの動作確認を重視する

## 実装結果

### ✅ 実装完了 (2025-09-14)

**実装されたファイル**:
- `cmd/usacloud-update/cobra_main.go` - サブコマンド定義追加
- `cmd/usacloud-update/version_e2e_test.go` - Version サブコマンドテスト
- `cmd/usacloud-update/validate_e2e_test.go` - Validate サブコマンドテスト
- `cmd/usacloud-update/config_e2e_test.go` - Config サブコマンドテスト
- `cmd/usacloud-update/sandbox_e2e_test.go` - Sandbox サブコマンドテスト
- `cmd/usacloud-update/compatibility_e2e_test.go` - 後方互換性テスト
- `README.md` - サブコマンド使用例追加

**テスト結果**: 全テスト PASS
- Version サブコマンド: 5/5 テスト成功
- Validate サブコマンド: 5/5 テスト成功  
- Config サブコマンド: 6/6 テスト成功
- Sandbox サブコマンド: 7/7 テスト成功
- 互換性テスト: 11/11 テスト成功

**実装時間**: 約90分（見積もり125分以内で完了）

**品質保証**:
- E2Eテスト: ✅ 実際のバイナリ実行テスト
- ユニットテスト: ✅ cobra構造検証
- 互換性テスト: ✅ 既存フラグとの併存確認
- 後方互換性: ✅ 既存インターフェース完全保持

### 📈 ユーザビリティ向上

**Before（フラグベース）**:
```bash
usacloud-update --sandbox --in script.sh
usacloud-update --validate-only --in script.sh
```

**After（サブコマンド + 後方互換）**:
```bash
# 新しい推奨方式
usacloud-update sandbox script.sh
usacloud-update validate script.sh
usacloud-update config
usacloud-update version

# 既存方式も継続サポート
usacloud-update --sandbox --in script.sh  # ✅ 引き続き動作
```

### 🎯 成果

1. **直感的なCLI**: より自然なコマンド体系を実現
2. **完全な後方互換性**: 既存ユーザーの学習コスト0
3. **段階的移行サポート**: 新旧両方式の併用可能
4. **包括的テスト**: 品質保証とリグレッション防止
5. **ドキュメント完備**: 使用例とマイグレーションガイド
