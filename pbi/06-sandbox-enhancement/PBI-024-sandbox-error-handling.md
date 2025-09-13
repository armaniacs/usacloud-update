# PBI-024: サンドボックス実行エラーハンドリング強化

## 概要
現在のサンドボックス実行機能のエラーハンドリングを強化し、より詳細なエラー情報の提供と自動回復機能を実装します。usacloudコマンド実行時のタイムアウト、ネットワークエラー、認証エラーなどの様々な障害パターンに対して、適切な対処方法をユーザーに提示します。

## 受け入れ条件
- [ ] タイムアウトエラー時に明確なエラーメッセージと対処方法を表示する
- [ ] ネットワークエラー時に再実行オプションを提供する
- [ ] 認証エラー時に設定確認を促すメッセージを表示する
- [ ] 予期しないエラー時にログ出力と問題報告方法を提示する
- [ ] エラー発生時のコマンド実行状態を適切に管理する

## 技術仕様

### 1. エラー分類システム
```go
// エラータイプ定義
type SandboxErrorType int

const (
    ErrorTypeTimeout SandboxErrorType = iota
    ErrorTypeNetwork
    ErrorTypeAuth
    ErrorTypeCommand
    ErrorTypeUnknown
)

type SandboxError struct {
    Type        SandboxErrorType
    Message     string
    Command     string
    Timestamp   time.Time
    Retryable   bool
    Suggestions []string
}
```

### 2. エラーハンドラー実装
```go
type ErrorHandler struct {
    logger    *log.Logger
    retryMax  int
    retryWait time.Duration
}

func (h *ErrorHandler) Handle(err error, cmd string) *SandboxError {
    sandboxErr := h.classifyError(err, cmd)
    h.logError(sandboxErr)
    return sandboxErr
}

func (h *ErrorHandler) classifyError(err error, cmd string) *SandboxError {
    // エラー分類ロジック
    if strings.Contains(err.Error(), "timeout") {
        return &SandboxError{
            Type:        ErrorTypeTimeout,
            Message:     "コマンド実行がタイムアウトしました",
            Command:     cmd,
            Timestamp:   time.Now(),
            Retryable:   true,
            Suggestions: []string{"--timeout オプションで時間を延長してください", "ネットワーク接続を確認してください"},
        }
    }
    // その他のエラー分類
    return nil
}
```

### 3. 自動回復機能
```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

func (e *Executor) ExecuteWithRetry(cmd string, config RetryConfig) (*Result, error) {
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        result, err := e.Execute(cmd)
        if err == nil {
            return result, nil
        }
        
        sandboxErr := e.errorHandler.Handle(err, cmd)
        if !sandboxErr.Retryable {
            return nil, err
        }
        
        if attempt < config.MaxAttempts {
            delay := calculateBackoffDelay(attempt, config)
            time.Sleep(delay)
        }
    }
    return nil, fmt.Errorf("最大試行回数を超過しました")
}
```

## テスト戦略
- **ユニットテスト**: エラー分類ロジックの検証
- **統合テスト**: 各種エラーパターンでの動作確認
- **エラー注入テスト**: 意図的にエラーを発生させての挙動テスト
- **ユーザビリティテスト**: エラーメッセージの分かりやすさ検証

## 依存関係
- 前提PBI: なし（既存サンドボックス機能を拡張）
- 関連PBI: PBI-025（並行実行でのエラー処理）、PBI-026（エラー履歴の永続化）
- 既存コード: internal/sandbox/executor.go

## 見積もり
- 開発工数: 8時間
  - エラー分類システム設計・実装: 3時間
  - 自動回復機能実装: 3時間
  - テスト実装: 2時間

## 完了の定義
- [ ] 全エラーパターンで適切なメッセージが表示される
- [ ] 自動リトライ機能が正常に動作する
- [ ] テストカバレッジが90%以上になる
- [ ] エラーハンドリングのドキュメントが整備される

## 実装状況
❌ **PBI-024は未実装** (2025-09-11)

**現在の状況**:
- 包括的なサンドボックスエラーハンドリング戦略が設計済み
- エラー分類システム、自動回復、リトライ機能の詳細設計完了
- ユーザーフレンドリーなエラーメッセージシステムの仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `internal/sandbox/error_handler.go` - 包括的なエラーハンドリングシステム
- エラー分類、ユーザーフレンドリーメッセージ生成機能
- 自動回復、リトライ、バックオフ機能
- ログシステムとエラーレポート機能
- 包括的なテストスイートとエラー注入テスト

**次のステップ**:
1. エラーハンドリングフレームワークの基盤実装
2. エラー分類とメッセージ生成機能の実装
3. 自動回復とリトライ機能の実装
4. ログシステムとテストスイートの実装
5. ドキュメント作成と統合テストの実行

## 備考
- エラーメッセージは日本語で分かりやすく表示する
- ログレベルはエラーの重要度に応じて適切に設定する
- 将来的にはメトリクス収集との連携も考慮する

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- サンドボックス機能拡張よりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存サンドボックス機能の安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のサンドボックスエラーハンドリングは引き続き保守・改善
- 拡張エラーハンドリング機能の実装は延期
- 現在のサンドボックス基盤の安定化を最優先