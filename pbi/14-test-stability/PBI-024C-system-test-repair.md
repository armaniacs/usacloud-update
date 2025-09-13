# PBI-024C: System Test修復

## 概要
フィルターシステムテスト（system_test.go）のAPI整合性修復。sandbox.ExecutionResultの構造変更に対応し、フィルターシステム全体の統合テストを正常化する。

## 受け入れ条件
- [ ] system_test.goが新しいAPI仕様で正常にコンパイルできること
- [ ] FilterSystemの統合テストが正常に動作すること
- [ ] ExportConfig/ImportConfig機能のテストが成功すること
- [ ] 複数フィルターの連携テストが機能すること

## 現在の問題

### 1. ExecutionResult構造の不整合
```bash
# 確認されるコンパイルエラー
internal/tui/filter/system_test.go:77:4: unknown field Category in struct literal
internal/tui/filter/system_test.go:78:4: unknown field Status in struct literal
internal/tui/filter/system_test.go:78:22: undefined: sandbox.StatusSuccess
internal/tui/filter/system_test.go:209:28: unknown field Status in struct literal
```

### 2. 問題の詳細分析
```go
// 問題のあるコード（system_test.go:77-78）
&sandbox.ExecutionResult{
    Category: "server",        // ❌ 存在しないフィールド
    Status:   sandbox.StatusSuccess,  // ❌ 存在しないフィールド・定数
}

// 実際のExecutionResult構造
type ExecutionResult struct {
    Command    string        `json:"command"`
    Success    bool          `json:"success"`
    Output     string        `json:"output"`
    Error      string        `json:"error,omitempty"`
    Duration   time.Duration `json:"duration"`
    Skipped    bool          `json:"skipped"`
}
```

## 修復計画

### 1. ExecutionResult構造修正
#### 基本フィールドの置き換え
```go
// 修正前
items := []interface{}{
    &sandbox.ExecutionResult{
        Category: "server",              // ❌ 削除
        Status:   sandbox.StatusSuccess, // ❌ 削除
    },
    &sandbox.ExecutionResult{
        Category: "disk",
        Status:   sandbox.StatusFailed,  // ❌ 削除
    },
}

// 修正後
items := []interface{}{
    &sandbox.ExecutionResult{
        Command: "usacloud server list",  // ✅ 実際のフィールド
        Success: true,                    // ✅ 実際のフィールド
    },
    &sandbox.ExecutionResult{
        Command: "usacloud disk list",
        Success: false,                   // ✅ 失敗ケース
    },
}
```

### 2. システム統合テストの修正
#### FilterSystem_ExportImportConfigテストの修正
```go
// テスト対象機能の確認
func TestFilterSystem_ExportImportConfig(t *testing.T) {
    system := NewFilterSystem()
    
    // テストデータの修正
    items := []interface{}{
        &sandbox.ExecutionResult{
            Command: "usacloud server list",
            Success: true,
            Output:  "server1\nserver2",
        },
        &sandbox.ExecutionResult{
            Command: "usacloud disk list", 
            Success: false,
            Error:   "permission denied",
        },
    }
    
    // フィルター適用・設定エクスポート
    result := system.Apply(items)
    exported := system.ExportConfig()
    
    // 新しいシステムでインポート
    system2 := NewFilterSystem()
    err := system2.ImportConfig(exported)
    // ...
}
```

### 3. フィルター連携テストの修正
#### 複数フィルターの統合テスト
```go
func TestFilterSystem_MultipleFilters(t *testing.T) {
    system := NewFilterSystem()
    
    // テキストフィルター設定
    textFilter := system.GetFilter("text")
    textFilter.SetActive(true)
    textFilter.SetConfig(FilterConfig{"query": "server"})
    
    // ステータスフィルター設定  
    statusFilter := system.GetFilter("status")
    statusFilter.SetActive(true)
    statusFilter.SetConfig(FilterConfig{"statuses": []string{"success"}})
    
    // テストデータ
    items := []interface{}{
        &sandbox.ExecutionResult{
            Command: "usacloud server list",  // ✅ textフィルターにマッチ
            Success: true,                    // ✅ statusフィルターにマッチ
        },
        &sandbox.ExecutionResult{
            Command: "usacloud disk list",    // ❌ textフィルターにマッチしない
            Success: true,
        },
        &sandbox.ExecutionResult{
            Command: "usacloud server create", // ✅ textフィルターにマッチ
            Success: false,                    // ❌ statusフィルターにマッチしない
        },
    }
    
    result := system.Apply(items)
    
    // 期待結果: 両方のフィルターにマッチする1つのアイテム
    if len(result) != 1 {
        t.Errorf("Expected 1 item after filtering, got %d", len(result))
    }
}
```

## 実装タスク

### Phase 1: 基本構造修正
1. **ExecutionResult生成の修正**
   - Categoryフィールド削除
   - Statusフィールド → Successフィールド
   - 適切なCommandフィールド設定

2. **定数・enum削除**
   - sandbox.StatusSuccess削除
   - sandbox.StatusFailed削除
   - 文字列ベースのステータス判定に変更

### Phase 2: テストロジック修正
1. **期待値の再計算**
   - フィルター結果の正確な予測
   - 新しいステータス判定ロジックへの対応
   - エラーメッセージの整合性確認

2. **テストデータの充実**
   - 実際のコマンド例の使用
   - 出力・エラー情報の追加
   - 実行時間の設定

### Phase 3: 統合テスト強化
1. **ExportConfig/ImportConfig機能**
   - フィルター設定の完全な保存・復元
   - 設定の整合性検証
   - エラーハンドリングテスト

2. **パフォーマンステスト**
   - 大量データでのフィルタリング
   - 複数フィルター組み合わせ時の性能
   - メモリ使用量の監視

## 対象テスト関数

### 修正が必要なテスト
1. **TestFilterSystem_ExportImportConfig**
   - ExecutionResult構造修正
   - 設定インポート・エクスポートの検証

2. **TestFilterSystem_MultipleFilters** 
   - 複数フィルター連携のテスト
   - 期待結果の再計算

3. **TestFilterSystem_Performance**
   - パフォーマンステストの修正
   - ベンチマーク基準の更新

### 新規追加予定のテスト
1. **TestFilterSystem_ErrorHandling**
   - 不正な設定での動作確認
   - エラー状態からの回復

2. **TestFilterSystem_StateManagement**
   - フィルター状態の管理
   - アクティブ状態の切り替え

## 見積もり
- **作業時間**: 2.5時間
  - Phase 1（基本構造修正）: 1時間
  - Phase 2（テストロジック修正）: 1時間  
  - Phase 3（統合テスト強化）: 0.5時間

## 完了の定義
- [ ] system_test.goが正常にコンパイル
- [ ] 全既存テストが成功
- [ ] FilterSystem統合機能の完全なテストカバレッジ
- [ ] Export/Import機能の動作確認
- [ ] パフォーマンス基準の確立

## 備考
- **統合的重要性**: FilterSystemは全フィルター機能の中核
- **他テストへの影響**: 修正パターンがtext_filter_test.goにも適用可能
- **将来の拡張性**: 新しいフィルタータイプ追加時の基盤確立

---

**予定開始**: PBI-024B完了後
**ステータス**: 🔄 **準備完了**  
**依存**: PBI-024B（status_filter_test.go修復）
**次のステップ**: PBI-024D（text_filter_test.go修復）

---

## 実装状況 (2025-09-11)

🟠 **PBI-024Cは未実装** (2025-09-11)

### 現在の状況
- system_test.goファイルは存在するがAPI変更により動作不能
- sandbox.ExecutionResult構造変更に未対応
- FilterSystem統合テストが実行不可
- PBI-024シリーズの一部として設計済み

### 未実装要素
1. **ExecutionResult構造の修正**
   - CategoryフィールドからCommandフィールドへの変更
   - StatusフィールドからSuccessフィールドへの変更
   - sandbox.StatusSuccess定数の削除への対応

2. **テストケースの再実装**
   - TestFilterSystem_Integration
   - TestFilterSystem_ExportImport
   - TestFilterSystem_MultiFilter
   - TestFilterSystem_StateManagement

3. **統合テスト強化**
   - 複数フィルター連携の検証
   - Export/Import機能の動作確認
   - パフォーマンス基準の確立

### 次のステップ
1. PBI-024Bの完了待ち（status_filter_test.go修復）
2. sandbox.ExecutionResult新構造の確認
3. system_test.goのAPI整合性修正
4. 統合テストの実装と検証
5. FilterSystemの完全なテストカバレッジ確立

### 技術要件
- Go 1.24.1対応
- sandbox.ExecutionResult新API準拠
- FilterSystem統合機能の完全テスト
- 2.5時間の作業見積もり

### 受け入れ条件の進捗
- [ ] system_test.goが新しいAPI仕様で正常にコンパイルできること
- [ ] FilterSystemの統合テストが正常に動作すること
- [ ] ExportConfig/ImportConfig機能のテストが成功すること
- [ ] 複数フィルターの連携テストが機能すること