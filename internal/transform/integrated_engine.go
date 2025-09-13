// Package transform provides integrated transformation engine for usacloud-update
package transform

import (
	"fmt"
	"strings"
	"time"

	"github.com/armaniacs/usacloud-update/internal/validation"
)

// IntegratedEngine は統合変換エンジン
type IntegratedEngine struct {
	// 既存コンポーネント
	engine *Engine
	rules  []Rule

	// 新しい検証コンポーネント
	mainValidator      *validation.MainCommandValidator
	subValidator       *validation.SubcommandValidator
	deprecatedDetector *validation.DeprecatedCommandDetector
	similarSuggester   *validation.SimilarCommandSuggester
	errorFormatter     *validation.ComprehensiveErrorFormatter
	parser             *validation.Parser

	// 統合設定
	config *IntegrationConfig
	stats  *IntegratedStats

	// パフォーマンス用
	cache map[string]*IntegratedResult
}

// IntegrationConfig は統合設定
type IntegrationConfig struct {
	EnablePreValidation     bool
	EnablePostValidation    bool
	EnableRuleConflictCheck bool
	StrictMode              bool
	ValidationPriority      ValidationPriority
	PerformanceMode         bool
	ParallelMode            bool
	BatchSize               int
	CacheEnabled            bool
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
	OriginalLine    string
	TransformedLine string
	Changed         bool
	Changes         []Change

	// 新しい検証フィールド
	PreValidationIssues  []ValidationIssue
	PostValidationIssues []ValidationIssue
	Suggestions          []validation.SimilarityResult
	DeprecationInfo      *validation.DeprecationInfo

	// 統合メタデータ
	ProcessingStage ProcessingStage
	RuleConflicts   []RuleConflict
	Confidence      float64
	LineNumber      int
}

// ValidationIssue は検証問題
type ValidationIssue struct {
	Type       IssueType
	Severity   Severity
	Component  string
	Message    string
	Expected   []string
	Confidence float64
}

// IssueType は問題タイプ
type IssueType int

const (
	IssueInvalidMainCommand IssueType = iota
	IssueInvalidSubCommand
	IssueDeprecatedCommand
	IssueSyntaxError
	IssueQualityWarning
	IssueRuleConflict
)

// Severity は重要度
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

// ProcessingStage は処理段階
type ProcessingStage int

const (
	StagePreValidation ProcessingStage = iota
	StageTransformation
	StagePostValidation
	StageCompleted
)

// RuleConflict はルール競合情報
type RuleConflict struct {
	Rule1        string
	Rule2        string
	ConflictType ConflictType
	Severity     ConflictSeverity
	Resolution   string
}

// ConflictType は競合タイプ
type ConflictType int

const (
	ConflictOverlap ConflictType = iota
	ConflictContradiction
	ConflictRedundancy
)

// ConflictSeverity は競合の重要度
type ConflictSeverity int

const (
	ConflictCritical ConflictSeverity = iota
	ConflictMajor
	ConflictMinor
)

// IntegratedStats は統合統計
type IntegratedStats struct {
	// 処理統計
	TotalLines       int
	ProcessedLines   int
	TransformedLines int

	// 検証統計
	PreValidationIssues  int
	PostValidationIssues int
	DeprecatedCommands   int

	// パフォーマンス統計
	ProcessingTimeMs   int64
	AverageTimePerLine float64
	CacheHitRate       float64
	CacheHits          int
	CacheMisses        int

	// 品質統計
	AverageConfidence   float64
	HighConfidenceLines int
	LowConfidenceLines  int

	// ルール統計
	RuleConflicts     int
	ResolvedConflicts int
}

// NewDefaultIntegrationConfig はデフォルトの統合設定を作成
func NewDefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		EnablePreValidation:     true,
		EnablePostValidation:    true,
		EnableRuleConflictCheck: true,
		StrictMode:              false,
		ValidationPriority:      PriorityBalanced,
		PerformanceMode:         false,
		ParallelMode:            false,
		BatchSize:               100,
		CacheEnabled:            true,
	}
}

// NewIntegratedEngine は新しい統合エンジンを作成
func NewIntegratedEngine(config *IntegrationConfig) *IntegratedEngine {
	if config == nil {
		config = NewDefaultIntegrationConfig()
	}

	engine := &IntegratedEngine{
		engine:             NewDefaultEngine(),
		rules:              DefaultRules(),
		mainValidator:      validation.NewMainCommandValidator(),
		subValidator:       nil, // 初期化は後で行う
		deprecatedDetector: validation.NewDeprecatedCommandDetector(),
		similarSuggester:   validation.NewSimilarCommandSuggester(3, 5),
		errorFormatter:     validation.NewDefaultComprehensiveErrorFormatter(),
		parser:             validation.NewParser(),
		config:             config,
		stats:              NewIntegratedStats(),
		cache:              make(map[string]*IntegratedResult),
	}

	// SubcommandValidatorの初期化（MainCommandValidatorが必要）
	engine.subValidator = validation.NewSubcommandValidator(engine.mainValidator)

	return engine
}

// NewIntegratedStats は新しい統合統計を作成
func NewIntegratedStats() *IntegratedStats {
	return &IntegratedStats{
		TotalLines:           0,
		ProcessedLines:       0,
		TransformedLines:     0,
		PreValidationIssues:  0,
		PostValidationIssues: 0,
		DeprecatedCommands:   0,
		ProcessingTimeMs:     0,
		AverageTimePerLine:   0.0,
		CacheHitRate:         0.0,
		CacheHits:            0,
		CacheMisses:          0,
		AverageConfidence:    0.0,
		HighConfidenceLines:  0,
		LowConfidenceLines:   0,
		RuleConflicts:        0,
		ResolvedConflicts:    0,
	}
}

// Process は統合処理を実行
func (ie *IntegratedEngine) Process(line string, lineNumber int) *IntegratedResult {
	startTime := time.Now()

	// キャッシュチェック
	if ie.config.CacheEnabled && ie.config.PerformanceMode {
		if cached, exists := ie.cache[line]; exists {
			ie.stats.CacheHits++
			cached.LineNumber = lineNumber // 行番号のみ更新
			return cached
		}
		ie.stats.CacheMisses++
	}

	result := &IntegratedResult{
		OriginalLine:         line,
		TransformedLine:      line,
		Changed:              false,
		Changes:              make([]Change, 0),
		PreValidationIssues:  make([]ValidationIssue, 0),
		PostValidationIssues: make([]ValidationIssue, 0),
		Suggestions:          make([]validation.SimilarityResult, 0),
		RuleConflicts:        make([]RuleConflict, 0),
		Confidence:           1.0,
		LineNumber:           lineNumber,
		ProcessingStage:      StagePreValidation,
	}

	// Stage 1: 事前検証
	if ie.config.EnablePreValidation {
		ie.performPreValidation(result)
	}

	// Stage 2: 変換処理
	ie.performTransformation(result)

	// Stage 3: 事後検証
	if ie.config.EnablePostValidation {
		ie.performPostValidation(result)
	}

	// Stage 4: 結果統合
	result.ProcessingStage = StageCompleted
	ie.integrateResults(result)

	// 統計更新
	ie.updateStats(result, time.Since(startTime))

	// キャッシュに保存
	if ie.config.CacheEnabled && ie.config.PerformanceMode {
		ie.cache[line] = result
	}

	return result
}

// performPreValidation は事前検証を実行
func (ie *IntegratedEngine) performPreValidation(result *IntegratedResult) {
	result.ProcessingStage = StagePreValidation

	// usacloudコマンドでない場合はスキップ
	if !strings.Contains(result.OriginalLine, "usacloud") {
		return
	}

	// コメントや空行はスキップ
	trimmed := strings.TrimSpace(result.OriginalLine)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return
	}

	// コマンド解析
	parsed, err := ie.parser.Parse(result.OriginalLine)
	if err != nil {
		result.PreValidationIssues = append(result.PreValidationIssues, ValidationIssue{
			Type:       IssueSyntaxError,
			Severity:   SeverityError,
			Message:    fmt.Sprintf("コマンド解析エラー: %v", err),
			Confidence: 0.9,
		})
		return
	}

	// メインコマンド検証
	if !ie.mainValidator.IsValidCommand(parsed.MainCommand) {
		suggestions := ie.similarSuggester.SuggestMainCommands(parsed.MainCommand)
		result.Suggestions = append(result.Suggestions, suggestions...)

		result.PreValidationIssues = append(result.PreValidationIssues, ValidationIssue{
			Type:       IssueInvalidMainCommand,
			Severity:   SeverityError,
			Component:  parsed.MainCommand,
			Message:    fmt.Sprintf("'%s' は有効なusacloudコマンドではありません", parsed.MainCommand),
			Confidence: 0.8,
		})
	}

	// サブコマンド検証
	if parsed.SubCommand != "" && !ie.subValidator.IsValidSubcommand(parsed.MainCommand, parsed.SubCommand) {
		subSuggestions := ie.similarSuggester.SuggestSubcommands(parsed.MainCommand, parsed.SubCommand)
		result.Suggestions = append(result.Suggestions, subSuggestions...)

		result.PreValidationIssues = append(result.PreValidationIssues, ValidationIssue{
			Type:       IssueInvalidSubCommand,
			Severity:   SeverityError,
			Component:  parsed.SubCommand,
			Message:    fmt.Sprintf("'%s' は %s コマンドの有効なサブコマンドではありません", parsed.SubCommand, parsed.MainCommand),
			Confidence: 0.8,
		})
	}

	// 廃止コマンド検出
	if ie.deprecatedDetector.IsDeprecated(parsed.MainCommand) {
		deprecatedInfo := ie.deprecatedDetector.Detect(parsed.MainCommand)
		result.DeprecationInfo = deprecatedInfo

		result.PreValidationIssues = append(result.PreValidationIssues, ValidationIssue{
			Type:       IssueDeprecatedCommand,
			Severity:   SeverityWarning,
			Component:  parsed.MainCommand,
			Message:    fmt.Sprintf("'%s' は廃止されたコマンドです: %s", parsed.MainCommand, deprecatedInfo.Message),
			Confidence: 1.0,
		})
	}
}

// performTransformation は変換処理を実行
func (ie *IntegratedEngine) performTransformation(result *IntegratedResult) {
	result.ProcessingStage = StageTransformation

	// 事前検証で致命的エラーがある場合の処理
	if ie.hasCriticalPreValidationIssues(result) && ie.config.StrictMode {
		// 変換をスキップ
		return
	}

	// ルール競合検出
	if ie.config.EnableRuleConflictCheck {
		conflicts := ie.detectRuleConflicts(result.OriginalLine)
		result.RuleConflicts = conflicts
	}

	// 既存の変換エンジン実行
	transformResult := ie.engine.Apply(result.OriginalLine)

	// 変換結果を統合結果にマージ
	result.TransformedLine = transformResult.Line
	result.Changed = transformResult.Changed
	result.Changes = transformResult.Changes
}

// performPostValidation は事後検証を実行
func (ie *IntegratedEngine) performPostValidation(result *IntegratedResult) {
	result.ProcessingStage = StagePostValidation

	if !result.Changed {
		return // 変換されていない場合はスキップ
	}

	// 基本的な品質スコア計算
	qualityScore := ie.calculateQualityScore(result)
	result.Confidence = qualityScore

	if qualityScore < 0.7 { // 品質が低い場合
		result.PostValidationIssues = append(result.PostValidationIssues, ValidationIssue{
			Type:       IssueQualityWarning,
			Severity:   SeverityWarning,
			Message:    "変換結果の品質が低い可能性があります",
			Confidence: qualityScore,
		})
	}

	// 変換結果の基本的な構文チェック
	if ie.hasBasicSyntaxIssues(result.TransformedLine) {
		result.PostValidationIssues = append(result.PostValidationIssues, ValidationIssue{
			Type:       IssueSyntaxError,
			Severity:   SeverityWarning,
			Message:    "変換後の構文に問題がある可能性があります",
			Confidence: 0.6,
		})
	}
}

// integrateResults は結果を統合
func (ie *IntegratedEngine) integrateResults(result *IntegratedResult) {
	// 信頼度の調整
	if len(result.PreValidationIssues) > 0 {
		result.Confidence *= 0.9
	}
	if len(result.PostValidationIssues) > 0 {
		result.Confidence *= 0.8
	}
	if len(result.RuleConflicts) > 0 {
		result.Confidence *= 0.7
	}

	// 最小信頼度の設定
	if result.Confidence < 0.1 {
		result.Confidence = 0.1
	}
}

// updateStats は統計を更新
func (ie *IntegratedEngine) updateStats(result *IntegratedResult, duration time.Duration) {
	ie.stats.TotalLines++
	ie.stats.ProcessedLines++

	if result.Changed {
		ie.stats.TransformedLines++
	}

	ie.stats.PreValidationIssues += len(result.PreValidationIssues)
	ie.stats.PostValidationIssues += len(result.PostValidationIssues)

	if result.DeprecationInfo != nil {
		ie.stats.DeprecatedCommands++
	}

	// パフォーマンス統計
	ie.stats.ProcessingTimeMs += duration.Nanoseconds() / 1000000
	ie.stats.AverageTimePerLine = float64(ie.stats.ProcessingTimeMs) / float64(ie.stats.ProcessedLines)

	// キャッシュ効率
	if ie.stats.CacheHits+ie.stats.CacheMisses > 0 {
		ie.stats.CacheHitRate = float64(ie.stats.CacheHits) / float64(ie.stats.CacheHits+ie.stats.CacheMisses)
	}

	// 品質統計
	ie.stats.AverageConfidence = (ie.stats.AverageConfidence*float64(ie.stats.ProcessedLines-1) + result.Confidence) / float64(ie.stats.ProcessedLines)

	if result.Confidence >= 0.8 {
		ie.stats.HighConfidenceLines++
	} else if result.Confidence < 0.5 {
		ie.stats.LowConfidenceLines++
	}

	// ルール統計
	ie.stats.RuleConflicts += len(result.RuleConflicts)
}

// hasCriticalPreValidationIssues は致命的な事前検証問題があるかチェック
func (ie *IntegratedEngine) hasCriticalPreValidationIssues(result *IntegratedResult) bool {
	for _, issue := range result.PreValidationIssues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// detectRuleConflicts はルール競合を検出
func (ie *IntegratedEngine) detectRuleConflicts(line string) []RuleConflict {
	var conflicts []RuleConflict

	// 簡単な競合検出ロジック
	// 複数のルールが同じ行に適用される場合の検出
	applicableRules := ie.findApplicableRules(line)

	if len(applicableRules) > 1 {
		for i, rule1 := range applicableRules {
			for _, rule2 := range applicableRules[i+1:] {
				conflict := RuleConflict{
					Rule1:        rule1.Name(),
					Rule2:        rule2.Name(),
					ConflictType: ConflictOverlap,
					Severity:     ConflictMinor,
					Resolution:   "最初に適用されるルールが優先されます",
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// findApplicableRules は適用可能なルールを見つける
func (ie *IntegratedEngine) findApplicableRules(line string) []Rule {
	var applicable []Rule

	for _, rule := range ie.rules {
		_, changed, _, _ := rule.Apply(line)
		if changed {
			applicable = append(applicable, rule)
		}
	}

	return applicable
}

// calculateQualityScore は品質スコアを計算
func (ie *IntegratedEngine) calculateQualityScore(result *IntegratedResult) float64 {
	score := 1.0

	// 基本的な品質指標
	if len(result.Changes) == 0 {
		return 1.0 // 変更なしは完全
	}

	// 変更の数に基づく調整（多すぎる変更は品質を下げる）
	if len(result.Changes) > 3 {
		score *= 0.8
	}

	// ルール名の信頼性チェック
	for _, change := range result.Changes {
		if strings.Contains(change.RuleName, "experimental") {
			score *= 0.7
		}
	}

	return score
}

// hasBasicSyntaxIssues は基本的な構文問題があるかチェック
func (ie *IntegratedEngine) hasBasicSyntaxIssues(line string) bool {
	// 基本的な構文チェック
	trimmed := strings.TrimSpace(line)

	// 空行や異常に短い行
	if len(trimmed) < 3 {
		return false // 空行は問題ではない
	}

	// usacloudで始まる行の基本チェック
	if strings.HasPrefix(trimmed, "usacloud") {
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			return true // usacloudのみは構文エラー
		}
	}

	return false
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

	if stats.ProcessedLines > 0 {
		report.WriteString(fmt.Sprintf("  • 変換済み: %d行 (%.1f%%)\n",
			stats.TransformedLines,
			float64(stats.TransformedLines)/float64(stats.ProcessedLines)*100))
	}

	// 検証統計
	report.WriteString("\n🔍 検証統計:\n")
	report.WriteString(fmt.Sprintf("  • 事前検証問題: %d個\n", stats.PreValidationIssues))
	report.WriteString(fmt.Sprintf("  • 事後検証問題: %d個\n", stats.PostValidationIssues))
	report.WriteString(fmt.Sprintf("  • 廃止コマンド: %d個\n", stats.DeprecatedCommands))

	// パフォーマンス統計
	report.WriteString("\n⚡ パフォーマンス統計:\n")
	report.WriteString(fmt.Sprintf("  • 処理時間: %dms\n", stats.ProcessingTimeMs))

	if stats.ProcessedLines > 0 {
		report.WriteString(fmt.Sprintf("  • 行あたり平均: %.2fms\n", stats.AverageTimePerLine))
	}

	if stats.CacheHits+stats.CacheMisses > 0 {
		report.WriteString(fmt.Sprintf("  • キャッシュ効率: %.1f%% (%d/%d)\n",
			stats.CacheHitRate*100, stats.CacheHits, stats.CacheHits+stats.CacheMisses))
	}

	// 品質統計
	report.WriteString("\n📈 品質統計:\n")
	report.WriteString(fmt.Sprintf("  • 平均信頼度: %.1f%%\n", stats.AverageConfidence*100))
	report.WriteString(fmt.Sprintf("  • 高信頼度: %d行\n", stats.HighConfidenceLines))
	report.WriteString(fmt.Sprintf("  • 低信頼度: %d行\n", stats.LowConfidenceLines))

	// ルール統計
	if stats.RuleConflicts > 0 {
		report.WriteString("\n⚠️  ルール統計:\n")
		report.WriteString(fmt.Sprintf("  • 競合検出: %d個\n", stats.RuleConflicts))
		report.WriteString(fmt.Sprintf("  • 解決済み: %d個\n", stats.ResolvedConflicts))
	}

	return report.String()
}

// Reset は統計をリセット
func (stats *IntegratedStats) Reset() {
	stats.TotalLines = 0
	stats.ProcessedLines = 0
	stats.TransformedLines = 0
	stats.PreValidationIssues = 0
	stats.PostValidationIssues = 0
	stats.DeprecatedCommands = 0
	stats.ProcessingTimeMs = 0
	stats.AverageTimePerLine = 0.0
	stats.CacheHitRate = 0.0
	stats.CacheHits = 0
	stats.CacheMisses = 0
	stats.AverageConfidence = 0.0
	stats.HighConfidenceLines = 0
	stats.LowConfidenceLines = 0
	stats.RuleConflicts = 0
	stats.ResolvedConflicts = 0
}

// GetStats は統計情報を取得
func (ie *IntegratedEngine) GetStats() *IntegratedStats {
	return ie.stats
}

// ResetStats は統計をリセット
func (ie *IntegratedEngine) ResetStats() {
	ie.stats.Reset()
}

// ClearCache はキャッシュをクリア
func (ie *IntegratedEngine) ClearCache() {
	ie.cache = make(map[string]*IntegratedResult)
	ie.stats.CacheHits = 0
	ie.stats.CacheMisses = 0
	ie.stats.CacheHitRate = 0.0
}
