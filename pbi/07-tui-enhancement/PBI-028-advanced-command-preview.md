# PBI-028: 高度なコマンドプレビュー機能

## 概要
TUIにおいて、変換されたコマンドの詳細なプレビュー機能を実装します。変更点のハイライト表示、実行前の影響分析、コマンドの説明表示などを含む包括的なプレビューシステムを提供します。

## 受け入れ条件
- [ ] 変換前後のコマンドを並列表示できる
- [ ] 変更箇所をハイライト表示できる
- [ ] 各コマンドの機能説明を表示できる
- [ ] 実行時の予想される影響を表示できる
- [ ] プレビュー内容をフィルタリング・検索できる

## 技術仕様

### 1. プレビューデータ構造
```go
type CommandPreview struct {
    Original    string                 `json:"original"`
    Transformed string                 `json:"transformed"`
    Changes     []ChangeHighlight      `json:"changes"`
    Description string                 `json:"description"`
    Impact      *ImpactAnalysis        `json:"impact"`
    Warnings    []string               `json:"warnings"`
    Category    string                 `json:"category"`
}

type ChangeHighlight struct {
    Type        ChangeType `json:"type"`
    Position    Range      `json:"position"`
    Original    string     `json:"original"`
    Replacement string     `json:"replacement"`
    Reason      string     `json:"reason"`
}

type ChangeType string

const (
    ChangeTypeOption    ChangeType = "option"
    ChangeTypeArgument  ChangeType = "argument"
    ChangeTypeCommand   ChangeType = "command"
    ChangeTypeFormat    ChangeType = "format"
    ChangeTypeRemoval   ChangeType = "removal"
)

type Range struct {
    Start int `json:"start"`
    End   int `json:"end"`
}

type ImpactAnalysis struct {
    Risk         RiskLevel `json:"risk"`
    Description  string    `json:"description"`
    Resources    []string  `json:"resources"`
    Dependencies []string  `json:"dependencies"`
}

type RiskLevel string

const (
    RiskLow    RiskLevel = "low"
    RiskMedium RiskLevel = "medium" 
    RiskHigh   RiskLevel = "high"
)
```

### 2. プレビュー生成エンジン
```go
type PreviewGenerator struct {
    transformer *transform.Engine
    analyzer    *ImpactAnalyzer
    dictionary  *CommandDictionary
}

func (pg *PreviewGenerator) Generate(original string) (*CommandPreview, error) {
    // 変換実行
    result := pg.transformer.Transform(original)
    
    // 変更点分析
    changes := pg.analyzeChanges(original, result.Line)
    
    // コマンド説明取得
    description := pg.dictionary.GetDescription(result.Line)
    
    // 影響分析
    impact := pg.analyzer.Analyze(result.Line)
    
    // 警告生成
    warnings := pg.generateWarnings(original, result.Line)
    
    return &CommandPreview{
        Original:    original,
        Transformed: result.Line,
        Changes:     changes,
        Description: description,
        Impact:      impact,
        Warnings:    warnings,
        Category:    pg.categorizeCommand(result.Line),
    }, nil
}

func (pg *PreviewGenerator) analyzeChanges(original, transformed string) []ChangeHighlight {
    var changes []ChangeHighlight
    
    // diff アルゴリズムを使用して変更点を特定
    diffs := difflib.UnifiedDiff{
        A:       difflib.SplitLines(original),
        B:       difflib.SplitLines(transformed),
        Context: 0,
    }
    
    for _, diff := range diffs {
        if strings.HasPrefix(diff, "-") && !strings.HasPrefix(diff, "---") {
            // 削除された部分
            changes = append(changes, ChangeHighlight{
                Type:     ChangeTypeRemoval,
                Original: strings.TrimPrefix(diff, "-"),
                Reason:   "このオプション・引数は廃止されました",
            })
        } else if strings.HasPrefix(diff, "+") && !strings.HasPrefix(diff, "+++") {
            // 追加された部分
            changes = append(changes, ChangeHighlight{
                Type:        ChangeTypeOption,
                Replacement: strings.TrimPrefix(diff, "+"),
                Reason:      "新しい形式に変換されました",
            })
        }
    }
    
    return changes
}
```

### 3. TUIプレビューウィジェット
```go
type PreviewWidget struct {
    *tview.Flex
    originalView    *tview.TextView
    transformedView *tview.TextView
    changesView     *tview.TextView
    impactView      *tview.TextView
    descriptionView *tview.TextView
    currentPreview  *CommandPreview
    app            *tview.Application
}

func NewPreviewWidget() *PreviewWidget {
    pw := &PreviewWidget{
        Flex: tview.NewFlex(),
    }
    
    pw.setupViews()
    pw.layoutViews()
    
    return pw
}

func (pw *PreviewWidget) setupViews() {
    // オリジナルコマンドビュー
    pw.originalView = tview.NewTextView().
        SetDynamicColors(true).
        SetTitle("🔍 変換前").
        SetBorder(true).
        SetBorderColor(tcell.ColorGray)
    
    // 変換後コマンドビュー  
    pw.transformedView = tview.NewTextView().
        SetDynamicColors(true).
        SetTitle("✨ 変換後").
        SetBorder(true).
        SetBorderColor(tcell.ColorGreen)
    
    // 変更点詳細ビュー
    pw.changesView = tview.NewTextView().
        SetDynamicColors(true).
        SetTitle("📋 変更詳細").
        SetBorder(true).
        SetScrollable(true)
    
    // 影響分析ビュー
    pw.impactView = tview.NewTextView().
        SetDynamicColors(true).
        SetTitle("⚠️ 影響分析").
        SetBorder(true).
        SetScrollable(true)
    
    // 説明ビュー
    pw.descriptionView = tview.NewTextView().
        SetDynamicColors(true).
        SetTitle("📖 コマンド説明").
        SetBorder(true).
        SetScrollable(true).
        SetWrap(true)
}

func (pw *PreviewWidget) layoutViews() {
    // 上段: オリジナル | 変換後
    topFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
        AddItem(pw.originalView, 0, 1, false).
        AddItem(pw.transformedView, 0, 1, false)
    
    // 下段: 変更詳細 | 影響分析 | 説明
    bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
        AddItem(pw.changesView, 0, 1, false).
        AddItem(pw.impactView, 0, 1, false).
        AddItem(pw.descriptionView, 0, 1, false)
    
    // 全体レイアウト
    pw.Flex.SetDirection(tview.FlexRow).
        AddItem(topFlex, 0, 2, false).
        AddItem(bottomFlex, 0, 3, false)
}

func (pw *PreviewWidget) UpdatePreview(preview *CommandPreview) {
    pw.currentPreview = preview
    
    // オリジナル表示
    pw.originalView.Clear()
    fmt.Fprintf(pw.originalView, "[white]%s[white]", preview.Original)
    
    // 変換後表示（変更箇所をハイライト）
    pw.transformedView.Clear()
    highlighted := pw.highlightChanges(preview.Transformed, preview.Changes)
    fmt.Fprintf(pw.transformedView, "%s", highlighted)
    
    // 変更詳細表示
    pw.changesView.Clear()
    for i, change := range preview.Changes {
        color := pw.getChangeColor(change.Type)
        fmt.Fprintf(pw.changesView, "[%s]%d. %s[white]\n", color, i+1, change.Reason)
        if change.Original != "" {
            fmt.Fprintf(pw.changesView, "   削除: [red]%s[white]\n", change.Original)
        }
        if change.Replacement != "" {
            fmt.Fprintf(pw.changesView, "   追加: [green]%s[white]\n", change.Replacement)
        }
        fmt.Fprintf(pw.changesView, "\n")
    }
    
    // 影響分析表示
    pw.updateImpactView()
    
    // 説明表示
    pw.descriptionView.Clear()
    fmt.Fprintf(pw.descriptionView, "[white]%s[white]", preview.Description)
}

func (pw *PreviewWidget) highlightChanges(text string, changes []ChangeHighlight) string {
    // 変更箇所をカラーコードでハイライト
    highlighted := text
    for _, change := range changes {
        color := pw.getChangeColor(change.Type)
        if change.Replacement != "" {
            highlighted = strings.ReplaceAll(highlighted, change.Replacement, 
                fmt.Sprintf("[%s]%s[white]", color, change.Replacement))
        }
    }
    return highlighted
}

func (pw *PreviewWidget) getChangeColor(changeType ChangeType) string {
    switch changeType {
    case ChangeTypeOption:
        return "green"
    case ChangeTypeArgument:
        return "blue" 
    case ChangeTypeCommand:
        return "yellow"
    case ChangeTypeFormat:
        return "cyan"
    case ChangeTypeRemoval:
        return "red"
    default:
        return "white"
    }
}
```

## テスト戦略
- **UIテスト**: プレビュー表示の正確性確認
- **レスポンシブテスト**: 異なるターミナルサイズでの動作確認
- **パフォーマンステスト**: 大量コマンドでの応答速度確認
- **ユーザビリティテスト**: 情報の見やすさと理解しやすさ検証

## 依存関係
- 前提PBI: なし（既存TUI機能を拡張）
- 関連PBI: PBI-029（フィルタリング機能）、PBI-030（キーボードショートカット）
- 既存コード: internal/tui/app.go

## 見積もり
- 開発工数: 11時間
  - プレビューデータ構造設計: 2時間
  - プレビュー生成エンジン実装: 4時間
  - TUIウィジェット実装: 4時間
  - 統合・テスト: 1時間

## 完了の定義
- [ ] 変換前後のコマンドが分かりやすく表示される
- [ ] 変更箇所が視覚的にハイライトされる
- [ ] 影響分析情報が適切に表示される
- [ ] プレビューが高速で応答する
- [ ] キーボード操作で快適にナビゲートできる

## 備考
- カラーテーマは端末の背景色に対応
- 長いコマンドでも適切に表示されるよう水平スクロール対応
- アクセシビリティを考慮した色選択

## 実装状況
❌ **PBI-028は未実装** (2025-09-11)

### 現在の状況
- 高度なコマンドプレビュー機能は未実装
- 変換前後のコマンド並列表示機能なし
- 変更箇所のハイライト表示機能なし
- 影響分析・リスク評価システムなし
- コマンド説明辞書システムなし

### 実装すべき要素
1. **プレビューデータ構造**
   - CommandPreview 構造体の実装
   - ChangeHighlight とChangeType の定義
   - ImpactAnalysis とRiskLevel システム
   - Range 構造体による位置情報管理

2. **プレビュー生成エンジン**
   - PreviewGenerator クラスの実装
   - diffアルゴリズムによる変更点解析
   - コマンド辞書との連携機能
   - 自動警告生成システム

3. **TUIプレビューウィジェット**
   - PreviewWidget の完全実装
   - 6つのビューエリア（オリジナル・変換後・変更詳細・影響分析・説明・警告）
   - 動的カラーハイライト機能
   - レスポンシブレイアウト管理

4. **視覚化機能**
   - 変更タイプ別カラーコーディング
   - リスクレベル表示
   - スクロール・検索機能
   - キーボードナビゲーション

### 次のステップ
1. `internal/preview/` パッケージの作成
2. CommandPreview データ構造の定義と実装
3. diffアルゴリズムベースの変更解析エンジン実装
4. コマンド辞書システムの構築
5. TUIプレビューウィジェットの作成
6. 既存TUIアプリケーションへの統合
7. カラーテーマとアクセシビリティ対応
8. 包括的なUIテスト作成

### 実装状況

**📊 実装状況: 未実装**

#### 実装延期の判断理由
本機能は高度なコマンドプレビューシステムの新規実装を含む大規模な機能拡張です。現在のプロジェクトの優先順位として、既存システムの安定性確保とバグ修正を最優先としており、新機能開発は一時的に延期します。

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
2. パフォーマンスへの影響評価が必要
3. コマンド辞書システムの新規開発が必要
4. UIコンポーネントの大幅な変更が必要

### 関連ファイル
- 実装予定: `internal/preview/generator.go`
- 実装予定: `internal/preview/widget.go`
- 実装予定: `internal/preview/analyzer.go`
- 実装予定: `internal/dictionary/commands.go`
- 統合対象: `internal/tui/app.go`
- 設定連携: `internal/tui/main_view.go`