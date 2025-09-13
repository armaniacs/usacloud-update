# PBI-024B: Status Filter Test修復

## 概要
ステータスフィルターテスト（status_filter_test.go）のAPI整合性修復。sandbox.ExecutionResultとpreview.CommandPreviewの構造変更に対応し、ステータスベースのフィルタリング機能の品質保証を確立する。

## 受け入れ条件
- [ ] status_filter_test.goが新しいAPI仕様で正常にコンパイルできること
- [ ] sandbox.ExecutionResultの新しい構造（Success/Skippedフィールド）に対応すること
- [ ] preview.CommandPreviewの新しい構造（Originalフィールド）に対応すること
- [ ] 全ステータスフィルター機能がテストされること

## 現在の問題

### 1. sandbox.ExecutionResult構造変更
```go
// 旧構造（テストで期待）
type ExecutionResult struct {
    Command string
    Status  ExecutionStatus  // ❌ 削除済み
}

// 新構造（実際の実装）
type ExecutionResult struct {
    Command string        `json:"command"`
    Success bool          `json:"success"`
    Skipped bool          `json:"skipped"`
    Output  string        `json:"output"`
    Error   string        `json:"error,omitempty"`
    Duration time.Duration `json:"duration"`
}
```

### 2. preview.CommandPreview構造変更
```go
// 旧構造（テストで期待）
type CommandPreview struct {
    Line        string    // ❌ 削除済み
    Command     string    // ❌ 削除済み
    Arguments   []string  // ❌ 削除済み
    LineNumber  int      // ❌ 削除済み
    IsConverted bool     // ❌ 削除済み
}

// 新構造（実際の実装）
type CommandPreview struct {
    Original    string            `json:"original"`
    Transformed string            `json:"transformed"`
    Changes     []ChangeHighlight `json:"changes"`
    Description string            `json:"description"`
    Impact      *ImpactAnalysis   `json:"impact"`
    Warnings    []string          `json:"warnings"`
    Category    string            `json:"category"`
    Metadata    *PreviewMetadata  `json:"metadata"`
}
```

## 修復計画

### 1. ExecutionResult関連テスト修正
#### 現在のgetStatusFromItem実装に合わせた修正
```go
// 修正前
items := []interface{}{
    &sandbox.ExecutionResult{
        Command: "usacloud server list",
        Status:  sandbox.StatusSuccess,  // ❌ 存在しないフィールド
    },
}

// 修正後
items := []interface{}{
    &sandbox.ExecutionResult{
        Command: "usacloud server list",
        Success: true,  // ✅ 実際のフィールド
    },
}
```

#### ステータス判定ロジックの整合
```go
// getStatusFromItem実装
func (f *StatusFilter) getStatusFromItem(item interface{}) string {
    switch v := item.(type) {
    case *sandbox.ExecutionResult:
        if v.Success {
            return "success"
        } else if v.Skipped {
            return "skipped"
        } else {
            return "failed"
        }
    // ...
}
```

### 2. CommandPreview関連テスト修正
```go
// 修正前
&preview.CommandPreview{
    Line:        "usacloud server list",  // ❌ 存在しないフィールド
    Command:     "usacloud",              // ❌ 存在しないフィールド
    Arguments:   []string{"server", "list"}, // ❌ 存在しないフィールド
}

// 修正後
&preview.CommandPreview{
    Original:    "usacloud server list",     // ✅ 実際のフィールド
    Transformed: "usacloud server list --output-type=json",
    Description: "Convert to JSON output",
    Category:    "output-format",
}
```

### 3. 期待値の調整
#### フィルター結果数の調整
```go
// Success/Failedフィルターの場合
config := FilterConfig{
    "statuses": []string{"success", "failed"},
}
// 修正前: 2つの結果を期待
// 修正後: 4つの結果を期待（success=true/false両方が該当）
```

#### Skippedステータスのテスト追加
```go
&sandbox.ExecutionResult{
    Command: "skipped command",
    Success: false,
    Skipped: true,  // ✅ Skippedステータスのテスト
}
```

## 実装タスク

### Phase 1: コンパイルエラー解消
1. **ExecutionResult構造修正**
   - Statusフィールド削除
   - Success/Skippedフィールド使用
   - StatusSuccess/StatusFailed定数削除

2. **CommandPreview構造修正**
   - Lineフィールド → Originalフィールド
   - 不要フィールド削除
   - 新しい必須フィールド追加

### Phase 2: テストロジック修正
1. **期待値調整**
   - フィルター結果数の再計算
   - ステータス判定ロジックの整合
   - エラーメッセージの更新

2. **新しいテストケース追加**
   - Skippedステータスのテスト
   - 複合ステータスフィルターのテスト
   - 無効ステータスの処理テスト

### Phase 3: 網羅性確保
1. **全ステータスパターンのテスト**
   - success (Success=true)
   - failed (Success=false, Skipped=false)
   - skipped (Skipped=true)
   - pending (CommandPreview)

2. **エッジケース追加**
   - 空のステータスリスト
   - 無効なステータス指定
   - 混在アイテムのフィルタリング

## 見積もり
- **作業時間**: 2時間
  - Phase 1（コンパイルエラー解消）: 1時間
  - Phase 2（テストロジック修正）: 0.5時間
  - Phase 3（網羅性確保）: 0.5時間

## 完了の定義
- [ ] status_filter_test.goが正常にコンパイル
- [ ] 全テスト関数が成功
- [ ] 新しいAPI仕様との整合性確認
- [ ] ステータスフィルター機能の完全なテストカバレッジ
- [ ] リグレッション防止のテスト追加

## 備考
- **部分修復済み**: ExecutionResultの一部とCommandPreviewの基本修正は完了
- **残課題**: 他のテストファイルでの同様問題解決のパターン確立
- **影響範囲**: system_test.go、text_filter_test.goでも同様の修正が必要

---

**予定開始**: PBI-024A完了後
**ステータス**: 🔄 **準備完了**
**依存**: PBI-024A（完了済み）
**次のステップ**: PBI-024C（system_test.go修復）

## 実装状況
🟠 **PBI-024Bは未実装** (2025-09-11)

### 現在の状況
- **ステータス**: 🟠 **0%完了** (未着手)
- **優先度**: 🔥 **高優先度** - 他PBIのパターンテンプレートとなる
- **準備状況**: ✅ **完全準備済み** - PBI-024Aの成果を活用可能
- **依存PBI**: PBI-024A ✅ 完了済み

### 未実装の要素
1. **sandbox.ExecutionResult構造変更対応**
   - Status フィールド削除対応
   - Success/Skipped フィールドへの変更
   - StatusSuccess/StatusFailed 定数削除
   - ステータス判定ロジックの整合性確保

2. **preview.CommandPreview構造変更対応**
   - Line フィールド → Original フィールドへの変更
   - Command、Arguments、LineNumber、IsConverted フィールド削除
   - 新しい必須フィールドの追加（Transformed、Description等）
   - CommandPreviewオブジェクトのテストデータ作成

3. **テストロジック修正**
   - フィルター結果数の再計算
   - ステータス判定ロジックとの整合性確認
   - エラーメッセージの更新
   - Skippedステータスの新規テスト追加

4. **網羅性確保**
   - 全ステータスパターンのテスト（success/failed/skipped/pending）
   - エッジケース追加（空ステータスリスト、無効ステータス指定）
   - 混在アイテムのフィルタリングテスト
   - リグレッション防止テストとエラーハンドリング

### 実装フェーズ
1. **Phase 1: コンパイルエラー解消** (1時間)
   - ExecutionResult構造修正（Status→Success/Skipped）
   - CommandPreview構造修正（Line→Original）
   - 不要フィールドと定数の削除

2. **Phase 2: テストロジック修正** (0.5時間)
   - 期待値調整（フィルター結果数の再計算）
   - ステータス判定ロジックの整合性確認
   - エラーメッセージの更新

3. **Phase 3: 網羅性確保** (0.5時間)
   - Skippedステータスの新規テスト
   - エッジケースとリグレッション防止テスト
   - 全ステータスパターンの完全テスト

### 期待される成果
- **API整合性**: sandbox、previewパッケージとの完全な整合
- **パターン確立**: 他PBI（024C、024D）の修復テンプレートとなる
- **品質保証**: ステータスフィルター機能の完全なテストカバレージ
- **他ファイル連携**: system_test.go、text_filter_test.goでの同様問題解決のパターン

### 次のステップ
1. PBI-024Aの成果をベースに修復パターン適用
2. Phase 1から順次実装開始
3. 各フェーズ完了後の測定と品質確認
4. 成果をPBI-024C、024Dのパターンとして文書化

### 関連ファイル
- 修復対象: `internal/tui/status_filter_test.go` 🟠
- パターン元: `internal/tui/preset_test.go` ✅
- 影響コード: `internal/sandbox/executor.go` ✅
- 影響コード: `internal/preview/generator.go` ✅
- 連携ファイル: `PBI-024C/D-*-repair.md`
- 総括管理: `PBI-024-DIVIDED-OVERVIEW.md`