# PBI-029: インタラクティブフィルタリングシステム

## 概要
TUIにおいて、大量のコマンドやファイルを効率的に絞り込むためのインタラクティブフィルタリング機能を実装します。リアルタイム検索、カテゴリフィルタ、複合条件フィルタリングを提供し、ユーザーが目的のコマンドを素早く見つけられるようにします。

## 受け入れ条件
- [ ] リアルタイムテキスト検索でコマンドを絞り込める
- [ ] コマンドカテゴリによるフィルタリングができる
- [ ] 実行ステータス（成功/失敗/未実行）で絞り込める
- [ ] 複数の条件を組み合わせてフィルタリングできる
- [ ] フィルタ条件の保存・呼び出しができる

## 技術仕様

### 1. フィルタシステム設計
```go
type FilterSystem struct {
    filters    []Filter
    activeSet  *FilterSet
    presets    map[string]*FilterSet
    callback   func([]interface{})
}

type Filter interface {
    Name() string
    Description() string
    Apply(items []interface{}) []interface{}
    IsActive() bool
    SetActive(bool)
    GetConfig() FilterConfig
    SetConfig(FilterConfig)
}

type FilterSet struct {
    ID      string
    Name    string
    Filters []FilterConfig
}

type FilterConfig struct {
    FilterID string
    Active   bool
    Value    interface{}
}
```

### 2. テキスト検索フィルタ
```go
type TextSearchFilter struct {
    active      bool
    searchTerm  string
    caseSensitive bool
    regex       bool
    fields      []string // 検索対象フィールド
}

func (f *TextSearchFilter) Apply(items []interface{}) []interface{} {
    if !f.active || f.searchTerm == "" {
        return items
    }
    
    var filtered []interface{}
    
    for _, item := range items {
        if f.matchesItem(item) {
            filtered = append(filtered, item)
        }
    }
    
    return filtered
}

func (f *TextSearchFilter) matchesItem(item interface{}) bool {
    switch v := item.(type) {
    case *CommandPreview:
        return f.searchInText(v.Original) || 
               f.searchInText(v.Transformed) ||
               f.searchInText(v.Description)
    case *ExecutionResult:
        return f.searchInText(v.Command) ||
               f.searchInText(v.Output) ||
               f.searchInText(v.Error)
    default:
        return false
    }
}

func (f *TextSearchFilter) searchInText(text string) bool {
    if !f.caseSensitive {
        text = strings.ToLower(text)
        term := strings.ToLower(f.searchTerm)
        
        if f.regex {
            matched, _ := regexp.MatchString(term, text)
            return matched
        }
        
        return strings.Contains(text, term)
    }
    
    if f.regex {
        matched, _ := regexp.MatchString(f.searchTerm, text)
        return matched
    }
    
    return strings.Contains(text, f.searchTerm)
}
```

### 3. カテゴリフィルタ
```go
type CategoryFilter struct {
    active           bool
    selectedCategories map[string]bool
    availableCategories []string
}

func (f *CategoryFilter) Apply(items []interface{}) []interface{} {
    if !f.active || len(f.selectedCategories) == 0 {
        return items
    }
    
    var filtered []interface{}
    
    for _, item := range items {
        category := f.getCategoryFromItem(item)
        if f.selectedCategories[category] {
            filtered = append(filtered, item)
        }
    }
    
    return filtered
}

func (f *CategoryFilter) getCategoryFromItem(item interface{}) string {
    switch v := item.(type) {
    case *CommandPreview:
        return v.Category
    case *ExecutionResult:
        return f.categorizeCommand(v.Command)
    default:
        return "unknown"
    }
}

func (f *CategoryFilter) categorizeCommand(command string) string {
    parts := strings.Fields(command)
    if len(parts) < 2 {
        return "unknown"
    }
    
    // usacloudコマンドのカテゴリ分類
    switch parts[1] {
    case "server", "disk", "switch", "router":
        return "infrastructure"
    case "archive", "cdrom", "note":
        return "storage"
    case "dns", "gslb", "proxylb":
        return "network"
    case "database", "nfs":
        return "managed-service"
    default:
        return "other"
    }
}
```

### 4. ステータスフィルタ
```go
type StatusFilter struct {
    active          bool
    allowedStatuses map[ExecutionStatus]bool
}

type ExecutionStatus string

const (
    StatusPending   ExecutionStatus = "pending"
    StatusRunning   ExecutionStatus = "running"
    StatusSuccess   ExecutionStatus = "success"
    StatusFailed    ExecutionStatus = "failed"
    StatusSkipped   ExecutionStatus = "skipped"
)

func (f *StatusFilter) Apply(items []interface{}) []interface{} {
    if !f.active {
        return items
    }
    
    var filtered []interface{}
    
    for _, item := range items {
        if result, ok := item.(*ExecutionResult); ok {
            if f.allowedStatuses[ExecutionStatus(result.Status)] {
                filtered = append(filtered, item)
            }
        } else {
            // プレビューなど実行前のアイテムは未実行として扱う
            if f.allowedStatuses[StatusPending] {
                filtered = append(filtered, item)
            }
        }
    }
    
    return filtered
}
```

### 5. TUIフィルタコンポーネント
```go
type FilterPanel struct {
    *tview.Flex
    searchInput     *tview.InputField
    categoryList    *tview.List
    statusCheckBox  *tview.Form
    presetDropdown  *tview.DropDown
    filterSystem    *FilterSystem
    onUpdate        func()
}

func NewFilterPanel(fs *FilterSystem) *FilterPanel {
    fp := &FilterPanel{
        Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
        filterSystem: fs,
    }
    
    fp.setupComponents()
    fp.layoutComponents()
    
    return fp
}

func (fp *FilterPanel) setupComponents() {
    // 検索入力
    fp.searchInput = tview.NewInputField().
        SetLabel("🔍 検索: ").
        SetChangedFunc(fp.onSearchChanged)
    
    // カテゴリリスト
    fp.categoryList = tview.NewList().
        SetTitle("📂 カテゴリ").
        SetBorder(true)
    
    categories := []string{
        "infrastructure", "storage", "network", 
        "managed-service", "other",
    }
    
    for _, category := range categories {
        fp.categoryList.AddItem(category, "", 0, fp.onCategoryToggle(category))
    }
    
    // ステータスチェックボックス
    fp.statusCheckBox = tview.NewForm().
        SetTitle("📊 ステータス").
        SetBorder(true)
    
    statuses := []ExecutionStatus{
        StatusPending, StatusRunning, StatusSuccess, 
        StatusFailed, StatusSkipped,
    }
    
    for _, status := range statuses {
        fp.statusCheckBox.AddCheckbox(string(status), true, fp.onStatusToggle(status))
    }
    
    // プリセットドロップダウン
    presetNames := fp.getPresetNames()
    fp.presetDropdown = tview.NewDropDown().
        SetLabel("💾 プリセット: ").
        SetOptions(presetNames, fp.onPresetSelect)
}

func (fp *FilterPanel) onSearchChanged(text string) {
    // リアルタイム検索の実装
    textFilter := fp.filterSystem.GetFilter("text-search").(*TextSearchFilter)
    textFilter.searchTerm = text
    textFilter.active = text != ""
    
    fp.triggerUpdate()
}

func (fp *FilterPanel) onCategoryToggle(category string) func() {
    return func() {
        categoryFilter := fp.filterSystem.GetFilter("category").(*CategoryFilter)
        if categoryFilter.selectedCategories[category] {
            delete(categoryFilter.selectedCategories, category)
        } else {
            categoryFilter.selectedCategories[category] = true
        }
        categoryFilter.active = len(categoryFilter.selectedCategories) > 0
        
        fp.triggerUpdate()
    }
}

func (fp *FilterPanel) triggerUpdate() {
    if fp.onUpdate != nil {
        fp.onUpdate()
    }
}
```

### 6. フィルタプリセット管理
```go
type PresetManager struct {
    presets map[string]*FilterSet
    storage PresetStorage
}

type PresetStorage interface {
    Save(preset *FilterSet) error
    Load(id string) (*FilterSet, error)
    List() []string
    Delete(id string) error
}

func (pm *PresetManager) SaveCurrentAsPreset(name string, fs *FilterSystem) error {
    preset := &FilterSet{
        ID:      generateID(),
        Name:    name,
        Filters: fs.ExportConfig(),
    }
    
    pm.presets[preset.ID] = preset
    return pm.storage.Save(preset)
}

func (pm *PresetManager) ApplyPreset(id string, fs *FilterSystem) error {
    preset, exists := pm.presets[id]
    if !exists {
        return fmt.Errorf("preset not found: %s", id)
    }
    
    return fs.ImportConfig(preset.Filters)
}
```

## テスト戦略
- **フィルタリング精度テスト**: 各フィルタの期待通りの絞り込み動作確認
- **パフォーマンステスト**: 大量データでのフィルタリング速度確認
- **UIレスポンステスト**: リアルタイム検索の応答性確認
- **プリセット機能テスト**: 保存・復元の正確性確認

## 依存関係
- 前提PBI: PBI-028（コマンドプレビュー機能）
- 関連PBI: PBI-030（キーボードショートカット）、PBI-031（カスタマイズ機能）
- 既存コード: internal/tui/app.go

## 見積もり
- 開発工数: 13時間
  - フィルタシステム設計・実装: 5時間
  - TUIコンポーネント実装: 5時間
  - プリセット管理機能: 2時間
  - 統合・テスト: 1時間

## 完了の定義
- [ ] リアルタイム検索が滑らかに動作する
- [ ] 複数フィルタの組み合わせが正しく機能する
- [ ] プリセット保存・復元が正確に動作する
- [ ] 大量データでも十分な性能を維持する
- [ ] UIが直感的で使いやすい

## 備考
- 検索性能最適化のためインデックス機能を検討
- 正規表現検索は上級者向けオプションとして提供
- フィルタ状態の永続化はローカルファイルベース

## 実装状況
❌ **PBI-029は未実装** (2025-09-11)

### 現在の状況
- インタラクティブフィルタリングシステムは未実装
- リアルタイムテキスト検索機能なし
- コマンドカテゴリフィルタ機能なし
- 実行ステータスでの絞り込み機能なし
- フィルタ条件の保存・呼び出し機能なし

### 実装すべき要素
1. **フィルタシステムコア**
   - FilterSystem クラスとFilter インターフェースの実装
   - FilterSet とFilterConfig 構造体の定義
   - フィルタチェーン処理エンジン
   - 動的フィルタ切り替え機能

2. **個別フィルタ実装**
   - TextSearchFilter: リアルタイムテキスト検索、正規表現サポート
   - CategoryFilter: コマンドカテゴリ別絞り込み
   - StatusFilter: 実行ステータス別フィルタリング
   - 組み合わせフィルタ機能

3. **TUIフィルタコンポーネント**
   - FilterPanel ウィジェットの完全実装
   - リアルタイム検索入力フィールド
   - カテゴリ選択リストコンポーネント
   - ステータスチェックボックスコンポーネント

4. **プリセット管理システム**
   - PresetManager クラスとPresetStorage インターフェース
   - フィルタ設定の保存・復元機能
   - プリセット選択ドロップダウン
   - ローカルファイルストレージ

5. **パフォーマンス最適化**
   - インデックシングシステム
   - 検索結果キャッシュ機能
   - 非同期フィルタリング処理
   - メモリ効率最適化

### 次のステップ
1. `internal/filter/` パッケージの作成
2. 基本的なFilterインターフェースとFilterSystemの実装
3. TextSearchFilterのリアルタイム検索機能実装
4. CategoryFilterとStatusFilterの実装
5. TUIFilterPanelコンポーネントの作成
6. プリセット管理システムの構築
7. 既存TUIアプリケーションへの統合
8. パフォーマンス最適化とテスト作成

### 実装状況

**📊 実装状況: 未実装**

#### 実装延期の判断理由
本機能は複雑なインタラクティブフィルタリングシステムの新規実装を含む大規模な機能拡張です。現在のプロジェクトの優先順位として、既存システムの安定性確保とバグ修正を最優先としており、新機能開発は一時的に延期します。

#### 延期期間
- **延期期間**: 次期メジャーリリース（v2.0以降）まで延期
- **再評価時期**: 現在の安定化作業完了後（推定：2025年Q2以降）

#### 現在の状況
- ✅ 仕様策定完了
- ❌ 実装未開始
- ❌ テスト未作成
- ❌ ドキュメント未作成

#### 実装時の考慮点
1. 既存TUIシステムとの統合複雑性
2. リアルタイム検索のパフォーマンス最適化が必要
3. 複数フィルタの組み合わせ処理の複雑性
4. プリセット管理システムの新規開発が必要
5. UIコンポーネントの大幅な変更が必要

### 関連ファイル
- 実装予定: `internal/filter/system.go`
- 実装予定: `internal/filter/text_search.go`
- 実装予定: `internal/filter/category.go`
- 実装予定: `internal/filter/status.go`
- 実装予定: `internal/filter/preset.go`
- 実装予定: `internal/tui/filter_panel.go`
- 統合対象: `internal/tui/app.go`
- 設定連携: `internal/tui/main_view.go`