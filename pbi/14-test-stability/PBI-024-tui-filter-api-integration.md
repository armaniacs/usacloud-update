# PBI-024: TUI Filter API整合性修復

## 概要
TUIフィルター機能のテストが完全にコンパイルエラーを起こしており、実装されたAPIと期待されるAPIに根本的な不一致がある問題を修復する。この問題により、TUIフィルター機能の品質保証と機能検証が不可能な状態となっている。

## 受け入れ条件
- [ ] TUIフィルター関連のテストが全てコンパイル・実行できること
- [ ] 実装されたAPIとテストコードのAPIが整合していること
- [ ] TUIフィルター機能の核心機能が正常にテストできること
- [ ] 既存のTUI機能に影響を与えることなく修復されること

## 技術仕様

### 現在の問題
```bash
# 現在の状況（preset_test.go.bakで無効化済み）
internal/tui/filter/preset_test.go:45:16: undefined: FilterPreset
internal/tui/filter/preset_test.go:46:16: undefined: FilterPreset
internal/tui/filter/preset_test.go:55:31: undefined: NewFilterManager
internal/tui/filter/preset_test.go:58:21: undefined: CreatePreset
```

### 1. API分析と設計修正
#### 問題の調査
```go
// 期待されるAPI（テストから推測）
type FilterPreset struct {
    Name        string
    Description string
    Rules       []FilterRule
}

type FilterManager interface {
    CreatePreset(name, description string, rules []FilterRule) error
    GetPreset(name string) (*FilterPreset, error)
    ListPresets() []FilterPreset
}

func NewFilterManager() *FilterManager
```

#### 実装すべき構造
```go
// internal/tui/filter/preset.go
type FilterPreset struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Rules       []FilterRule           `json:"rules"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type FilterRule struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`  // "contains", "equals", "regex"
    Value    interface{} `json:"value"`
    Enabled  bool        `json:"enabled"`
}

type FilterManager struct {
    presets []FilterPreset
    mutex   sync.RWMutex
}
```

### 2. テスト復活とAPIマッチング
#### テストファイル修正
```go
// internal/tui/filter/preset_test.go の復活
func TestFilterPreset_Creation(t *testing.T) {
    preset := FilterPreset{
        Name:        "Test Preset",
        Description: "Test description",
        Rules: []FilterRule{
            {Field: "name", Operator: "contains", Value: "test", Enabled: true},
        },
    }
    
    if preset.Name != "Test Preset" {
        t.Errorf("Expected name 'Test Preset', got %s", preset.Name)
    }
}

func TestFilterManager_Operations(t *testing.T) {
    manager := NewFilterManager()
    
    // プリセット作成テスト
    err := manager.CreatePreset("test", "Test preset", []FilterRule{
        {Field: "status", Operator: "equals", Value: "active", Enabled: true},
    })
    if err != nil {
        t.Fatalf("Failed to create preset: %v", err)
    }
    
    // プリセット取得テスト
    preset, err := manager.GetPreset("test")
    if err != nil {
        t.Fatalf("Failed to get preset: %v", err)
    }
    if preset.Name != "test" {
        t.Errorf("Expected preset name 'test', got %s", preset.Name)
    }
}
```

## テスト戦略
- **ユニットテスト**: FilterPreset、FilterRule、FilterManager の各機能テスト
- **統合テスト**: TUI内でのフィルター機能動作テスト
- **回帰テスト**: 既存TUI機能への影響確認テスト

## 依存関係
- 前提PBI: なし（独立した修復タスク）
- 関連PBI: PBI-025（Profile機能修復）、PBI-029（Test Coverage向上）
- 既存コード: internal/tui/ パッケージ全体

## 見積もり
- 開発工数: 8時間
  - API調査・設計修正: 3時間
  - FilterPreset/FilterManager実装: 3時間
  - テスト修正・作成: 2時間

## 完了の定義
- [ ] 全TUIフィルター関連テストがコンパイル・実行成功
- [ ] フィルター機能の核心動作が正常にテスト検証済み
- [ ] 既存TUI機能の回帰テスト通過
- [ ] コードレビュー完了
- [ ] ドキュメント更新（API仕様書）

## 備考
- この修復は、TUI機能の品質保証体制確立の基盤となる重要なタスク
- 無効化されたテストファイル（preset_test.go.bak）を有効化・修正する
- 将来的なTUIフィルター機能拡張の基盤APIを確立する

## 実装状況
⚠️ **PBI-024は分割済み（このファイルは廃止）** (2025-09-11)

### 重要な変更通知
❗ **このファイルは現在使用されていません**

**PBI-024は予想以上に複雑であったため、4つの小さなPBIに分割されました。**

### 新しいファイル構成
このファイルの代わりに、以下のファイル群を使用してください：

1. **`PBI-024-DIVIDED-OVERVIEW.md`** - 📊 **メインファイル**
   - 分割後の全体概要と進捗管理
   - 各PBIのステータスと優先度
   - 成果と品質指標

2. **`PBI-024A-preset-test-repair.md`** - ✅ **完了済み**
   - preset_test.goの完全修復完了
   - PresetManagerとFilePresetStorageの品質保証確立

3. **`PBI-024B-status-filter-test-repair.md`** - 🟠 **未実装**
   - status_filter_test.goの修復待ち
   - sandbox.ExecutionResult構造変更対応

4. **`PBI-024C-system-test-repair.md`** - 🟠 **未実装**
   - system_test.goの修復待ち
   - FilterSystem統合テストの修正

5. **`PBI-024D-text-filter-test-repair.md`** - 🟠 **未実装**
   - text_filter_test.goの修復待ち
   - FilterConfig構造変更対応

### 分割の利点
- **明確な進捗管理**: 25%完了（PBI-024A完了済み）
- **リスク軽減**: 段階的実装で中断・再開可能
- **品質保証**: PBI-024Aで確立した修復パターンを他に適用

### 現在の実装状況概要
- **完了率**: 25% (1/4 PBI完了)
- **成果**: preset_test.goの完全修復完了
- **次のタスク**: PBI-024B（Status Filter Test修復）が高優先度

### アクションアイテム
1. **このファイルを無視し、`PBI-024-DIVIDED-OVERVIEW.md`を参照**
2. **実装作業の続行は`PBI-024B-status-filter-test-repair.md`から**
3. **進捗管理は`PBI-024-DIVIDED-OVERVIEW.md`で確認**

### 関連ファイル
- メイン: `PBI-024-DIVIDED-OVERVIEW.md` 📊
- 完了: `PBI-024A-preset-test-repair.md` ✅
- 待機: `PBI-024B/C/D-*-repair.md` 🟠
- 非推奨: `PBI-024-tui-filter-api-integration.md` ⚠️