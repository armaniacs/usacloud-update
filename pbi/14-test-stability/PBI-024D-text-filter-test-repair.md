# PBI-024D: Text Filter Test修復

## 概要
テキストフィルターテスト（text_filter_test.go）のAPI整合性修復。FilterConfigの構造変更に対応し、テキストベースのフィルタリング機能の品質保証を確立する。

## 受け入れ条件
- [ ] text_filter_test.goが新しいAPI仕様で正常にコンパイルできること
- [ ] FilterConfigの新しい構造に対応すること
- [ ] テキスト検索・フィルタリング機能が完全にテストされること
- [ ] 正規表現・大文字小文字処理のテストが機能すること

## 現在の問題

### 1. FilterConfig構造の変更
```bash
# 確認されるコンパイルエラー
internal/tui/filter/text_filter_test.go:257:12: config.FilterID undefined
internal/tui/filter/text_filter_test.go:258:64: config.FilterID undefined  
internal/tui/filter/text_filter_test.go:261:13: config.Active undefined
```

### 2. 問題の詳細分析
```go
// 旧FilterConfig構造（テストで期待）
type FilterConfig struct {
    FilterID string `json:"filter_id"`  // ❌ 削除済み
    Active   bool   `json:"active"`     // ❌ 削除済み
    // ...その他のフィールド
}

// 新FilterConfig構造（実際の実装）
type FilterConfig map[string]interface{}
```

### 3. 影響を受けるテストコード
```go
// 問題のあるコード例
config := filter.GetConfig()
if config.FilterID != expectedID {              // ❌ 存在しないフィールド
    t.Errorf("Expected FilterID %s", expectedID)
}

if !config.Active {                             // ❌ 存在しないフィールド  
    t.Error("Filter should be active")
}
```

## 修復計画

### 1. FilterConfig使用パターンの修正
#### map[string]interface{}形式への変更
```go
// 修正前
config := filter.GetConfig()
if config.FilterID != "text-filter" {
    t.Error("Wrong filter ID")
}

// 修正後
config := filter.GetConfig()
if query, ok := config["query"].(string); !ok || query != expectedQuery {
    t.Errorf("Expected query %s, got %v", expectedQuery, query)
}
```

#### アクティブ状態の確認方法変更
```go
// 修正前
if !config.Active {
    t.Error("Filter should be active")
}

// 修正後
if !filter.IsActive() {  // ✅ Filter interfaceのメソッド使用
    t.Error("Filter should be active")
}
```

### 2. テキストフィルター特有の設定
#### query設定のテスト
```go
func TestTextFilter_QueryConfiguration(t *testing.T) {
    filter := NewTextFilter()
    
    // クエリ設定
    config := FilterConfig{
        "query": "server",
        "case_sensitive": false,
        "regex_mode": false,
    }
    
    err := filter.SetConfig(config)
    if err != nil {
        t.Fatalf("SetConfig failed: %v", err)
    }
    
    // 設定取得・確認
    retrievedConfig := filter.GetConfig()
    
    if query, ok := retrievedConfig["query"].(string); !ok || query != "server" {
        t.Errorf("Expected query 'server', got %v", query)
    }
    
    if caseSensitive, ok := retrievedConfig["case_sensitive"].(bool); !ok || caseSensitive {
        t.Errorf("Expected case_sensitive false, got %v", caseSensitive)
    }
}
```

### 3. テキスト検索機能のテスト
#### 大文字小文字処理
```go
func TestTextFilter_CaseSensitivity(t *testing.T) {
    filter := NewTextFilter()
    filter.SetActive(true)
    
    tests := []struct {
        name          string
        caseSensitive bool
        query         string
        text          string
        shouldMatch   bool
    }{
        {"Case insensitive match", false, "Server", "server list", true},
        {"Case insensitive no match", false, "Server", "disk list", false},
        {"Case sensitive match", true, "Server", "Server list", true},
        {"Case sensitive no match", true, "Server", "server list", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := FilterConfig{
                "query": tt.query,
                "case_sensitive": tt.caseSensitive,
            }
            filter.SetConfig(config)
            
            items := []interface{}{
                &sandbox.ExecutionResult{
                    Command: tt.text,
                    Success: true,
                },
            }
            
            result := filter.Apply(items)
            
            if tt.shouldMatch && len(result) != 1 {
                t.Errorf("Expected match but got %d items", len(result))
            }
            if !tt.shouldMatch && len(result) != 0 {
                t.Errorf("Expected no match but got %d items", len(result))
            }
        })
    }
}
```

#### 正規表現モード
```go
func TestTextFilter_RegexMode(t *testing.T) {
    filter := NewTextFilter()
    filter.SetActive(true)
    
    config := FilterConfig{
        "query": "server.*list",
        "regex_mode": true,
    }
    filter.SetConfig(config)
    
    items := []interface{}{
        &sandbox.ExecutionResult{
            Command: "usacloud server list",     // ✅ マッチ
            Success: true,
        },
        &sandbox.ExecutionResult{
            Command: "usacloud server create",   // ❌ マッチしない
            Success: true,
        },
    }
    
    result := filter.Apply(items)
    
    if len(result) != 1 {
        t.Errorf("Expected 1 match for regex pattern, got %d", len(result))
    }
}
```

## 実装タスク

### Phase 1: FilterConfig修正
1. **古いフィールドアクセス削除**
   - config.FilterID削除
   - config.Active削除
   - map形式でのアクセスに変更

2. **Filter interfaceメソッド使用**
   - filter.IsActive()でアクティブ状態確認
   - filter.Name()でフィルター名取得
   - filter.SetActive()で状態変更

### Phase 2: テスト機能の拡充
1. **基本テキスト検索**
   - 部分一致検索
   - 完全一致検索
   - 空クエリの処理

2. **高度な検索機能**
   - 大文字小文字区別設定
   - 正規表現モード
   - 複数語検索

### Phase 3: エッジケースとエラーハンドリング
1. **不正入力の処理**
   - 不正な正規表現
   - null/undefined値
   - 型不整合

2. **パフォーマンステスト**
   - 大量テキストでの検索
   - 複雑な正規表現パターン
   - メモリ効率の確認

## 対象テスト関数

### 修正が必要なテスト
1. **TestTextFilter_Configuration**
   - FilterConfig構造の修正
   - 設定取得・確認方法の変更

2. **TestTextFilter_Apply**
   - ExecutionResult構造の確認
   - フィルタリング結果の検証

3. **TestTextFilter_RegexSearch**
   - 正規表現機能のテスト
   - エラーハンドリングの確認

### 新規追加予定のテスト
1. **TestTextFilter_PerformanceSearch**
   - 大量データでの検索性能
   - 複雑パターンでの処理時間

2. **TestTextFilter_ErrorHandling**
   - 不正な正規表現の処理
   - 設定エラーからの回復

## 検索対象の拡張

### FilterableItemインターフェース対応
```go
// getSearchableTextの実装確認
func (f *TextFilter) getSearchableText(item interface{}) []string {
    switch v := item.(type) {
    case *sandbox.ExecutionResult:
        return []string{v.Command, v.Output, v.Error}
    case *preview.CommandPreview:
        return []string{v.Original, v.Transformed, v.Description}
    case FilterableItem:
        return v.GetSearchableText()
    default:
        return []string{}
    }
}
```

## 見積もり
- **作業時間**: 2時間
  - Phase 1（FilterConfig修正）: 0.5時間
  - Phase 2（テスト機能拡充）: 1時間
  - Phase 3（エッジケース）: 0.5時間

## 完了の定義
- [ ] text_filter_test.goが正常にコンパイル
- [ ] 全テキストフィルター機能がテスト済み
- [ ] 大文字小文字・正規表現処理の確認完了
- [ ] パフォーマンス基準の確立
- [ ] エラーハンドリングの動作確認

## 備考
- **最終修復**: PBI-024の最後の修復作業
- **パターン完成**: 他パッケージでのFilterConfig使用パターンの確立
- **品質保証**: テキスト検索機能の信頼性確保

---

**予定開始**: PBI-024C完了後
**ステータス**: 🔄 **準備完了**
**依存**: PBI-024C（system_test.go修復）  
**完了後**: PBI-024全体完了、PBI-025開始可能

---

## 実装状況 (2025-09-11)

🟠 **PBI-024Dは未実装** (2025-09-11)

### 現在の状況
- text_filter_test.goファイルは存在するがAPI変更により動作不能
- FilterConfig構造変更に未対応
- テキストフィルター機能のテストが実行不可
- PBI-024シリーズの最終修復作業として設計済み

### 未実装要素
1. **FilterConfig構造の修正**
   - FilterIDフィールドの削除への対応
   - Activeフィールドの削除への対応
   - map[string]interface{}型への変更対応

2. **テストケースの再実装**
   - TestTextFilter_BasicFiltering
   - TestTextFilter_RegexSupport
   - TestTextFilter_CaseInsensitive
   - TestTextFilter_Performance
   - TestTextFilter_ErrorHandling

3. **機能テストの強化**
   - 大文字小文字処理の確認
   - 正規表現機能の動作確認
   - パフォーマンス基準の確立
   - エラーハンドリングの検証

### 次のステップ
1. PBI-024Cの完了待ち（system_test.go修復）
2. FilterConfig新構造の確認
3. text_filter_test.goのAPI整合性修正
4. テキストフィルター機能の包括的テスト実装
5. パフォーマンス基準とエラーハンドリングの確立

### 技術要件
- Go 1.24.1対応
- FilterConfig新API準拠（map[string]interface{}型）
- テキスト検索・フィルタリング機能の完全テスト
- 2時間の作業見積もり

### 受け入れ条件の進捗
- [ ] text_filter_test.goが新しいAPI仕様で正常にコンパイルできること
- [ ] FilterConfigの新しい構造に対応すること
- [ ] テキスト検索・フィルタリング機能が完全にテストされること
- [ ] 正規表現・大文字小文字処理のテストが機能すること