# PBI-016: 変換エンジン統合

## 概要
新しいコマンド検証システムを既存のusacloud-update変換エンジンと統合し、変換前の検証、変換中の追加検証、変換結果の検証を一元的に管理する統合システムを実装する。既存のルールベース変換と新しい検証機能が効率的に連携する設計を実現する。

## 受け入れ条件
- [ ] 既存の変換ルールシステムとの完全な互換性が保たれている
- [ ] 変換前・変換中・変換後の各段階で適切な検証が実行される
- [ ] 変換処理のパフォーマンスが大幅に低下していない
- [ ] 検証結果が変換結果に適切に統合されている
- [ ] ルールの競合やオーバーラップが適切に処理されている

## 技術仕様

### 統合アーキテクチャ

#### 1. 拡張された変換エンジン
```go
// internal/transform/integrated_engine.go
package transform

import (
    "fmt"
    "strings"
    
    "github.com/armaniacs/usacloud-update/internal/validation"
)

// IntegratedEngine は統合変換エンジン
type IntegratedEngine struct {
    // 既存コンポーネント
    engine         *Engine
    rules          []Rule
    
    // 新しい検証コンポーネント
    validator      *validation.ValidationSystem
    preValidator   *validation.PreTransformValidator
    postValidator  *validation.PostTransformValidator
    
    // 統合設定
    config         *IntegrationConfig
    stats          *IntegratedStats
}

// IntegrationConfig は統合設定
type IntegrationConfig struct {
    EnablePreValidation     bool
    EnablePostValidation    bool
    EnableRuleConflictCheck bool
    StrictMode             bool
    ValidationPriority     ValidationPriority
    PerformanceMode        bool
}

// ValidationPriority は検証優先度
type ValidationPriority int

const (
    PriorityValidationFirst ValidationPriority = iota // 検証優先
    PriorityTransformFirst                            // 変換優先
    PriorityBalanced                                  // バランス
)

// IntegratedResult は統合処理結果
type IntegratedResult struct {
    // 既存フィールド
    OriginalLine  string
    TransformedLine string
    Changed       bool
    RuleName      string
    BeforeFragment string
    AfterFragment  string
    
    // 新しい検証フィールド
    ValidationResults []validation.ValidationResult
    PreValidationIssues []validation.ValidationIssue
    PostValidationIssues []validation.ValidationIssue
    Suggestions []validation.SuggestionResult
    DeprecationInfo *validation.DeprecationInfo
    
    // 統合メタデータ
    ProcessingStage ProcessingStage
    RuleConflicts  []RuleConflict
    Confidence     float64
}

// ProcessingStage は処理段階
type ProcessingStage int

const (
    StagePreValidation  ProcessingStage = iota
    StageTransformation
    StagePostValidation
    StageCompleted
)

// NewIntegratedEngine は新しい統合エンジンを作成
func NewIntegratedEngine(config *IntegrationConfig) *IntegratedEngine {
    return &IntegratedEngine{
        engine:        NewEngine(),
        validator:     validation.NewValidationSystem(),
        preValidator:  validation.NewPreTransformValidator(),
        postValidator: validation.NewPostTransformValidator(),
        config:       config,
        stats:        NewIntegratedStats(),
    }
}

// Process は統合処理を実行
func (ie *IntegratedEngine) Process(line string, lineNumber int) *IntegratedResult {
    result := &IntegratedResult{
        OriginalLine: line,
        ValidationResults: make([]validation.ValidationResult, 0),
    }
    
    // Stage 1: 事前検証
    if ie.config.EnablePreValidation {
        ie.performPreValidation(result, lineNumber)
    }
    
    // Stage 2: 変換処理
    ie.performTransformation(result)
    
    // Stage 3: 事後検証
    if ie.config.EnablePostValidation {
        ie.performPostValidation(result)
    }
    
    // Stage 4: 結果統合
    ie.integrateResults(result)
    
    return result
}
```

### 検証段階の詳細実装

#### 1. 事前検証（Pre-Validation）
```go
// PreTransformValidator は変換前検証器
type PreTransformValidator struct {
    commandValidator    *validation.MainCommandValidator
    subcommandValidator *validation.SubcommandValidator
    deprecatedDetector  *validation.DeprecatedCommandDetector
    syntaxAnalyzer     *validation.SyntaxAnalyzer
}

// performPreValidation は事前検証を実行
func (ie *IntegratedEngine) performPreValidation(result *IntegratedResult, lineNumber int) {
    result.ProcessingStage = StagePreValidation
    
    // コマンド解析
    cmdLine := ie.preValidator.ParseCommandLine(result.OriginalLine)
    if cmdLine == nil {
        return // コメントや空行などをスキップ
    }
    
    // メインコマンド検証
    if mainResult := ie.preValidator.ValidateMainCommand(cmdLine); mainResult != nil {
        result.PreValidationIssues = append(result.PreValidationIssues, 
            mainResult.ToValidationIssue())
    }
    
    // サブコマンド検証
    if subResult := ie.preValidator.ValidateSubCommand(cmdLine); subResult != nil {
        result.PreValidationIssues = append(result.PreValidationIssues, 
            subResult.ToValidationIssue())
    }
    
    // 廃止コマンド検出
    if deprecatedInfo := ie.preValidator.DetectDeprecated(cmdLine); deprecatedInfo != nil {
        result.DeprecationInfo = deprecatedInfo
        result.PreValidationIssues = append(result.PreValidationIssues,
            deprecatedInfo.ToValidationIssue())
    }
    
    // 構文解析
    if syntaxIssues := ie.preValidator.AnalyzeSyntax(cmdLine); len(syntaxIssues) > 0 {
        result.PreValidationIssues = append(result.PreValidationIssues, syntaxIssues...)
    }
    
    ie.stats.RecordPreValidation(len(result.PreValidationIssues))
}

// performTransformation は変換処理を実行
func (ie *IntegratedEngine) performTransformation(result *IntegratedResult) {
    result.ProcessingStage = StageTransformation
    
    // 事前検証で致命的エラーがある場合の処理
    if ie.hasCriticalPreValidationIssues(result) && ie.config.StrictMode {
        result.TransformedLine = result.OriginalLine // 変換をスキップ
        result.Changed = false
        return
    }
    
    // 既存の変換エンジン実行
    transformResult := ie.engine.ApplyRules(result.OriginalLine)
    
    // 変換結果を統合結果にマージ
    result.TransformedLine = transformResult.Line
    result.Changed = transformResult.Changed
    result.RuleName = transformResult.RuleName
    result.BeforeFragment = transformResult.BeforeFragment
    result.AfterFragment = transformResult.AfterFragment
    
    // ルール競合検出
    if ie.config.EnableRuleConflictCheck {
        conflicts := ie.detectRuleConflicts(result.OriginalLine)
        result.RuleConflicts = conflicts
    }
    
    ie.stats.RecordTransformation(result.Changed)
}
```

#### 2. 事後検証（Post-Validation）
```go
// PostTransformValidator は変換後検証器
type PostTransformValidator struct {
    consistencyChecker *validation.ConsistencyChecker
    qualityAnalyzer   *validation.QualityAnalyzer
    syntaxValidator   *validation.SyntaxValidator
}

// performPostValidation は事後検証を実行
func (ie *IntegratedEngine) performPostValidation(result *IntegratedResult) {
    result.ProcessingStage = StagePostValidation
    
    if !result.Changed {
        return // 変換されていない場合はスキップ
    }
    
    // 構文一貫性チェック
    if consistencyIssues := ie.postValidator.CheckConsistency(
        result.OriginalLine, result.TransformedLine); len(consistencyIssues) > 0 {
        result.PostValidationIssues = append(result.PostValidationIssues, consistencyIssues...)
    }
    
    // 品質分析
    qualityScore := ie.postValidator.AnalyzeQuality(result.TransformedLine)
    result.Confidence = qualityScore
    
    if qualityScore < 0.7 { // 品質が低い場合
        result.PostValidationIssues = append(result.PostValidationIssues,
            validation.ValidationIssue{
                Type:        validation.IssueQualityWarning,
                Severity:    validation.SeverityWarning,
                Message:     "変換結果の品質が低い可能性があります",
                Confidence:  qualityScore,
            })
    }
    
    // 変換結果の構文チェック
    if syntaxIssues := ie.postValidator.ValidateSyntax(result.TransformedLine); len(syntaxIssues) > 0 {
        result.PostValidationIssues = append(result.PostValidationIssues, syntaxIssues...)
    }
    
    ie.stats.RecordPostValidation(len(result.PostValidationIssues), result.Confidence)
}
```

### ルール競合処理

#### 1. ルール競合検出
```go
// RuleConflict はルール競合情報
type RuleConflict struct {
    Rule1        string    // 競合するルール1
    Rule2        string    // 競合するルール2
    ConflictType ConflictType // 競合タイプ
    Severity     ConflictSeverity // 競合の重要度
    Resolution   string    // 推奨される解決方法
}

// ConflictType は競合タイプ
type ConflictType int

const (
    ConflictOverlap       ConflictType = iota // パターンのオーバーラップ
    ConflictContradiction                     // 矛盾する変換
    ConflictRedundancy                        // 冗長な変換
)

// detectRuleConflicts はルール競合を検出
func (ie *IntegratedEngine) detectRuleConflicts(line string) []RuleConflict {
    var conflicts []RuleConflict
    applicableRules := ie.findApplicableRules(line)
    
    // 適用可能なルールが複数ある場合の競合チェック
    for i, rule1 := range applicableRules {
        for j, rule2 := range applicableRules[i+1:] {
            if conflict := ie.analyzeRuleConflict(rule1, rule2, line); conflict != nil {
                conflicts = append(conflicts, *conflict)
            }
        }
    }
    
    return conflicts
}

// resolveRuleConflict はルール競合を解決
func (ie *IntegratedEngine) resolveRuleConflict(conflicts []RuleConflict, line string) string {
    if len(conflicts) == 0 {
        return ie.engine.ApplyRules(line).Line
    }
    
    // 優先度ベースの解決
    prioritizedRules := ie.prioritizeRules(conflicts, line)
    
    // 最優先ルールを適用
    if len(prioritizedRules) > 0 {
        return ie.applySpecificRule(prioritizedRules[0], line)
    }
    
    return line
}
```

### パフォーマンス最適化

#### 1. 効率的な処理パイプライン
```go
// ProcessingPipeline は処理パイプライン
type ProcessingPipeline struct {
    stages        []ProcessingStage
    parallelMode  bool
    batchSize     int
    cacheEnabled  bool
    cache         map[string]*IntegratedResult
}

// ProcessBatch はバッチ処理を実行
func (pp *ProcessingPipeline) ProcessBatch(lines []string, startLineNumber int) []*IntegratedResult {
    if pp.parallelMode {
        return pp.processParallel(lines, startLineNumber)
    }
    return pp.processSequential(lines, startLineNumber)
}

func (pp *ProcessingPipeline) processParallel(lines []string, startLineNumber int) []*IntegratedResult {
    numWorkers := min(len(lines), 8) // 最大8並列
    jobs := make(chan ProcessJob, len(lines))
    results := make(chan *IntegratedResult, len(lines))
    
    // ワーカー起動
    for w := 0; w < numWorkers; w++ {
        go pp.worker(jobs, results)
    }
    
    // ジョブ投入
    for i, line := range lines {
        jobs <- ProcessJob{
            Line:       line,
            LineNumber: startLineNumber + i,
        }
    }
    close(jobs)
    
    // 結果収集
    var allResults []*IntegratedResult
    for i := 0; i < len(lines); i++ {
        allResults = append(allResults, <-results)
    }
    
    return allResults
}

// キャッシュ機能
func (ie *IntegratedEngine) processWithCache(line string, lineNumber int) *IntegratedResult {
    if ie.config.PerformanceMode {
        if cached, exists := ie.cache[line]; exists {
            return cached
        }
    }
    
    result := ie.Process(line, lineNumber)
    
    if ie.config.PerformanceMode {
        ie.cache[line] = result
    }
    
    return result
}
```

### 統計とメトリクス

#### 1. 統合統計情報
```go
// IntegratedStats は統合統計
type IntegratedStats struct {
    // 処理統計
    TotalLines          int
    ProcessedLines      int
    TransformedLines    int
    
    // 検証統計
    PreValidationIssues  int
    PostValidationIssues int
    DeprecatedCommands   int
    
    // パフォーマンス統計
    ProcessingTimeMs     int64
    AverageTimePerLine   float64
    CacheHitRate        float64
    
    // 品質統計
    AverageConfidence   float64
    HighConfidenceLines int
    LowConfidenceLines  int
}

// GenerateReport は統合レポートを生成
func (stats *IntegratedStats) GenerateReport() string {
    report := strings.Builder{}
    
    report.WriteString("📊 usacloud-update 統合処理レポート\n")
    report.WriteString("================================\n\n")
    
    // 処理統計
    report.WriteString("🔄 処理統計:\n")
    report.WriteString(fmt.Sprintf("  • 総行数: %d行\n", stats.TotalLines))
    report.WriteString(fmt.Sprintf("  • 処理済み: %d行\n", stats.ProcessedLines))
    report.WriteString(fmt.Sprintf("  • 変換済み: %d行 (%.1f%%)\n", 
        stats.TransformedLines, 
        float64(stats.TransformedLines)/float64(stats.ProcessedLines)*100))
    
    // 検証統計
    report.WriteString("\n🔍 検証統計:\n")
    report.WriteString(fmt.Sprintf("  • 事前検証問題: %d個\n", stats.PreValidationIssues))
    report.WriteString(fmt.Sprintf("  • 事後検証問題: %d個\n", stats.PostValidationIssues))
    report.WriteString(fmt.Sprintf("  • 廃止コマンド: %d個\n", stats.DeprecatedCommands))
    
    // パフォーマンス統計
    report.WriteString("\n⚡ パフォーマンス統計:\n")
    report.WriteString(fmt.Sprintf("  • 処理時間: %dms\n", stats.ProcessingTimeMs))
    report.WriteString(fmt.Sprintf("  • 行あたり平均: %.2fms\n", stats.AverageTimePerLine))
    
    if stats.CacheHitRate > 0 {
        report.WriteString(fmt.Sprintf("  • キャッシュ効率: %.1f%%\n", stats.CacheHitRate*100))
    }
    
    // 品質統計
    report.WriteString("\n📈 品質統計:\n")
    report.WriteString(fmt.Sprintf("  • 平均信頼度: %.1f%%\n", stats.AverageConfidence*100))
    report.WriteString(fmt.Sprintf("  • 高信頼度: %d行\n", stats.HighConfidenceLines))
    report.WriteString(fmt.Sprintf("  • 低信頼度: %d行\n", stats.LowConfidenceLines))
    
    return report.String()
}
```

## テスト戦略
- 回帰テスト：既存のゴールデンファイルテストが全て通過することを確認
- 統合テスト：各処理段階の連携が正しく動作することを確認
- パフォーマンステスト：統合後の処理速度が許容範囲内であることを確認
- 競合テスト：ルール競合が適切に検出・解決されることを確認
- 品質テスト：変換品質が既存レベル以上であることを確認
- 並列処理テスト：並列処理モードが正しく動作することを確認

## 依存関係
- 前提PBI: PBI-001～014 (全検証・エラーフィードバック), PBI-015 (統合CLI)
- 既存コード: `internal/transform/engine.go`, `internal/transform/rules.go`
- 関連PBI: PBI-017 (設定統合)

## 見積もり
- 開発工数: 12時間
  - 統合エンジン実装: 4時間
  - 事前・事後検証実装: 3時間
  - ルール競合処理実装: 2時間
  - パフォーマンス最適化: 2時間
  - 統計・レポート機能: 1時間

## 完了の定義
- [ ] `internal/transform/integrated_engine.go`ファイルが作成されている
- [ ] `IntegratedEngine`が既存エンジンとの完全互換性を保っている
- [ ] 事前検証・変換・事後検証の3段階処理が実装されている
- [ ] ルール競合検出と解決機能が実装されている
- [ ] パフォーマンス最適化（並列処理、キャッシュ）が実装されている
- [ ] 統合統計とレポート機能が実装されている
- [ ] 全ての既存テストが継続して通過している
- [ ] 新機能の包括的テストが作成され、すべて通過している
- [ ] パフォーマンス回帰がないことが確認されている
- [ ] コードレビューが完了している

## 備考
- 既存の変換エンジンとの後方互換性を絶対に保つことが最重要
- パフォーマンスの劣化は5%以内に抑える必要がある
- 複雑な統合により保守性が下がらないよう注意深い設計が必要
- 将来的な新ルール追加時の拡張性を考慮した実装が重要

## 実装状況
🟠 **PBI-016は部分実装** (2025-09-11)

### 現在の状況
- 基本的な変換エンジンは実装済み（`internal/transform/engine.go`）
- ルールベースの変換システムは動作中
- 既存のDefaultRules()とApplyRules()機能は完全実装済み
- 基本的な結果構造体（Result）は存在している

### 未実装の要素
1. **IntegratedEngine コアシステム**
   - IntegratedEngine 構造体と統合アーキテクチャ
   - IntegrationConfig とValidationPriority 設定管理
   - IntegratedResult 拡張結果構造体
   - ProcessingStage と段階管理システム

2. **事前・事後検証統合**
   - PreTransformValidator: 変換前検証システム
   - PostTransformValidator: 変換後検証システム
   - performPreValidation() とperformPostValidation() メソッド
   - ValidationSystem との連携インターフェース

3. **ルール競合処理**
   - RuleConflict 構造体と競合検出システム
   - detectRuleConflicts() とresolveRuleConflict() 機能
   - ConflictType とConflictSeverity 管理
   - ルール優先度ベースの競合解決

4. **パフォーマンス最適化**
   - ProcessingPipeline と並列処理システム
   - processParallel() とprocessSequential() 機能
   - キャッシュ機能とprocessWithCache() メソッド
   - バッチ処理とワーカープールシステム

5. **統合統計システム**
   - IntegratedStats 構造体とメトリクス収集
   - GenerateReport() 統合レポート生成
   - 処理・検証・パフォーマンス・品質統計
   - リアルタイムメトリクス監視

6. **高度な結果統合**
   - ValidationResults とPreValidationIssues 管理
   - DeprecationInfo とSuggestions 統合
   - Confidence スコアと品質管理
   - integrateResults() 結果統合メソッド

### 部分実装済みの要素
✅ **基本変換エンジン**: Engine, Ruleインターフェース, ApplyRules()
✅ **変換ルールシステム**: DefaultRules(), simpleRule 実装
✅ **基本結果構造**: Result 構造体と基本フィールド
✅ **ゴールデンファイルテスト**: 既存変換ロジックのテストカバレッジ

### 次のステップ
1. `internal/transform/integrated_engine.go` ファイルの作成
2. IntegratedEngine 構造体と基本インターフェースの実装
3. PreTransformValidator とPostTransformValidator の実装
4. 事前・変換・事後の3段階処理パイプライン構築
5. ルール競合検出と解決システムの実装
6. パフォーマンス最適化（並列処理・キャッシュ）の実装
7. IntegratedStats とレポート生成機能の作成
8. 既存エンジンとの後方互換性テスト
9. パフォーマンスリグレッションテスト

### 関連ファイル
- 拡張対象: `internal/transform/engine.go` ✅
- 拡張対象: `internal/transform/rules.go` ✅
- 実装予定: `internal/transform/integrated_engine.go`
- 実装予定: `internal/transform/pre_validator.go`
- 実装予定: `internal/transform/post_validator.go`
- 実装予定: `internal/transform/conflict_resolver.go`
- 実装予定: `internal/transform/pipeline.go`
- 統合対象: `internal/validation/` パッケージ
- テスト連携: `internal/transform/engine_test.go` ✅