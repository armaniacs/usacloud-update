package preview

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/armaniacs/usacloud-update/internal/transform"
)

// Generator generates command previews with detailed analysis
type Generator struct {
	transformer *transform.Engine
	analyzer    *ImpactAnalyzer
	dictionary  *CommandDictionary
	options     *PreviewOptions
}

// NewGenerator creates a new preview generator
func NewGenerator(opts *PreviewOptions) *Generator {
	if opts == nil {
		opts = &PreviewOptions{
			IncludeDescription: true,
			IncludeImpact:      true,
			IncludeWarnings:    true,
			MaxDescLength:      500,
			Timeout:            5 * time.Second,
		}
	}

	return &Generator{
		transformer: transform.NewDefaultEngine(),
		analyzer:    NewImpactAnalyzer(),
		dictionary:  NewCommandDictionary(),
		options:     opts,
	}
}

// Generate creates a detailed preview for a command
func (g *Generator) Generate(original string, lineNumber int) (*CommandPreview, error) {
	startTime := time.Now()

	// Transform the command
	result := g.transformer.Apply(original)

	// Analyze changes
	changes := g.analyzeChanges(original, &result)

	// Get command description
	var description string
	if g.options.IncludeDescription {
		description = g.dictionary.GetDescription(result.Line)
		if len(description) > g.options.MaxDescLength {
			description = description[:g.options.MaxDescLength] + "..."
		}
	}

	// Analyze impact
	var impact *ImpactAnalysis
	if g.options.IncludeImpact {
		impact = g.analyzer.Analyze(result.Line)
	}

	// Generate warnings
	var warnings []string
	if g.options.IncludeWarnings {
		warnings = g.generateWarnings(original, result.Line, changes)
	}

	// Categorize command
	category := g.categorizeCommand(result.Line)

	preview := &CommandPreview{
		Original:    original,
		Transformed: result.Line,
		Changes:     changes,
		Description: description,
		Impact:      impact,
		Warnings:    warnings,
		Category:    category,
		Metadata: &PreviewMetadata{
			LineNumber:     lineNumber,
			GeneratedAt:    time.Now(),
			ProcessingTime: time.Since(startTime),
			Version:        "1.9.0",
		},
	}

	return preview, nil
}

// analyzeChanges analyzes the differences between original and transformed commands
func (g *Generator) analyzeChanges(original string, result *transform.Result) []ChangeHighlight {
	var changes []ChangeHighlight

	if !result.Changed {
		return changes
	}

	for _, change := range result.Changes {
		highlight := ChangeHighlight{
			Type:        g.mapChangeType(""),
			Original:    change.Before,
			Replacement: change.After,
			Reason:      g.generateChangeReason(&change),
			RuleName:    change.RuleName,
		}

		// Calculate position if possible
		if change.Before != "" {
			if pos := strings.Index(original, change.Before); pos >= 0 {
				highlight.Position = Range{
					Start: pos,
					End:   pos + len(change.Before),
				}
			}
		}

		changes = append(changes, highlight)
	}

	return changes
}

// mapChangeType maps transform change types to preview change types
func (g *Generator) mapChangeType(transformType string) ChangeType {
	switch transformType {
	case "option":
		return ChangeTypeOption
	case "argument":
		return ChangeTypeArgument
	case "command":
		return ChangeTypeCommand
	case "format":
		return ChangeTypeFormat
	case "removal":
		return ChangeTypeRemoval
	default:
		return ChangeTypeOption
	}
}

// generateChangeReason generates a human-readable reason for a change
func (g *Generator) generateChangeReason(change *transform.Change) string {
	switch {
	case strings.Contains(change.RuleName, "output"):
		return "出力フォーマットをJSONに変更"
	case strings.Contains(change.RuleName, "selector"):
		return "セレクタオプションを引数に変換"
	case strings.Contains(change.RuleName, "resource"):
		return "リソース名を新しい形式に変更"
	case strings.Contains(change.RuleName, "product"):
		return "プロダクト名を新しいエイリアスに変更"
	case strings.Contains(change.RuleName, "zone"):
		return "ゾーン指定の形式を正規化"
	case strings.Contains(change.RuleName, "deprecated"):
		return "廃止されたコマンド・オプションを削除"
	default:
		return "usacloud v1.1 互換性のための変更"
	}
}

// generateWarnings generates warnings for potential issues
func (g *Generator) generateWarnings(original, transformed string, changes []ChangeHighlight) []string {
	var warnings []string

	// Check for deprecated commands
	deprecatedCommands := []string{"summary", "object-storage"}
	for _, deprecated := range deprecatedCommands {
		if strings.Contains(transformed, deprecated) {
			warnings = append(warnings, fmt.Sprintf("コマンド '%s' は廃止されています", deprecated))
		}
	}

	// Check for manual intervention needed
	if strings.Contains(transformed, "# 手動での対応が必要") {
		warnings = append(warnings, "このコマンドは手動での対応が必要です")
	}

	// Check for output format changes
	for _, change := range changes {
		if change.Type == ChangeTypeFormat {
			warnings = append(warnings, "出力形式が変更されます。スクリプトの後続処理を確認してください")
			break
		}
	}

	// Check for high-risk operations
	riskPatterns := []string{
		"delete", "remove", "destroy", "shutdown", "reboot",
	}
	for _, pattern := range riskPatterns {
		if strings.Contains(strings.ToLower(transformed), pattern) {
			warnings = append(warnings, "このコマンドは破壊的な操作を含む可能性があります")
			break
		}
	}

	return warnings
}

// categorizeCommand determines the category of a command
func (g *Generator) categorizeCommand(command string) string {
	command = strings.ToLower(command)

	categories := map[string][]string{
		"server":     {"server", "instance", "vm"},
		"database":   {"database", "db", "mariadb", "postgres"},
		"network":    {"network", "switch", "router", "internet", "subnet", "interface"},
		"storage":    {"disk", "archive", "iso", "cdrom", "volume"},
		"security":   {"certificate", "cert", "ssh-key", "private-host", "security"},
		"monitoring": {"monitor", "activity", "metric", "log"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(command, keyword) {
				return category
			}
		}
	}

	return "other"
}

// ImpactAnalyzer analyzes the impact of executing commands
type ImpactAnalyzer struct {
	riskPatterns map[RiskLevel][]string
}

// NewImpactAnalyzer creates a new impact analyzer
func NewImpactAnalyzer() *ImpactAnalyzer {
	return &ImpactAnalyzer{
		riskPatterns: map[RiskLevel][]string{
			RiskCritical: {"delete", "destroy", "remove", "format"},
			RiskHigh:     {"shutdown", "reboot", "reset", "modify"},
			RiskMedium:   {"create", "update", "configure", "change"},
			RiskLow:      {"list", "show", "get", "describe", "info"},
		},
	}
}

// Analyze analyzes the impact and risk of a command
func (ia *ImpactAnalyzer) Analyze(command string) *ImpactAnalysis {
	command = strings.ToLower(command)

	// Determine risk level
	risk := RiskLow
	for level, patterns := range ia.riskPatterns {
		for _, pattern := range patterns {
			if strings.Contains(command, pattern) {
				if ia.isHigherRisk(level, risk) {
					risk = level
				}
			}
		}
	}

	// Extract resources
	resources := ia.extractResources(command)

	// Generate description
	description := ia.generateRiskDescription(risk, command)

	// Calculate complexity
	complexity := ia.calculateComplexity(command)

	return &ImpactAnalysis{
		Risk:         risk,
		Description:  description,
		Resources:    resources,
		Dependencies: ia.extractDependencies(command),
		Complexity:   complexity,
	}
}

// isHigherRisk checks if one risk level is higher than another
func (ia *ImpactAnalyzer) isHigherRisk(level1, level2 RiskLevel) bool {
	riskOrder := map[RiskLevel]int{
		RiskLow:      1,
		RiskMedium:   2,
		RiskHigh:     3,
		RiskCritical: 4,
	}
	return riskOrder[level1] > riskOrder[level2]
}

// extractResources extracts resource types from the command
func (ia *ImpactAnalyzer) extractResources(command string) []string {
	var resources []string

	resourcePatterns := map[string]*regexp.Regexp{
		"server":      regexp.MustCompile(`(?i)server|instance`),
		"database":    regexp.MustCompile(`(?i)database|mariadb|postgres`),
		"disk":        regexp.MustCompile(`(?i)disk|volume`),
		"network":     regexp.MustCompile(`(?i)switch|router|subnet`),
		"certificate": regexp.MustCompile(`(?i)certificate|cert`),
	}

	for resource, pattern := range resourcePatterns {
		if pattern.MatchString(command) {
			resources = append(resources, resource)
		}
	}

	return resources
}

// extractDependencies extracts potential dependencies from the command
func (ia *ImpactAnalyzer) extractDependencies(command string) []string {
	var dependencies []string

	// Common dependencies based on command patterns
	if strings.Contains(command, "server") {
		dependencies = append(dependencies, "network interface", "disk")
	}
	if strings.Contains(command, "database") {
		dependencies = append(dependencies, "server", "network")
	}
	if strings.Contains(command, "switch") {
		dependencies = append(dependencies, "subnet", "server")
	}

	return dependencies
}

// generateRiskDescription generates a description based on risk level
func (ia *ImpactAnalyzer) generateRiskDescription(risk RiskLevel, command string) string {
	switch risk {
	case RiskCritical:
		return "このコマンドは重要なリソースを削除・破壊する可能性があります。実行前に十分な確認が必要です。"
	case RiskHigh:
		return "このコマンドはシステムの状態を大きく変更します。影響範囲を事前に確認してください。"
	case RiskMedium:
		return "このコマンドは新しいリソースを作成または既存の設定を変更します。"
	case RiskLow:
		return "このコマンドは情報の取得・表示のみを行い、システムへの影響はありません。"
	default:
		return "影響度は不明です。"
	}
}

// calculateComplexity calculates the complexity score of a command
func (ia *ImpactAnalyzer) calculateComplexity(command string) int {
	complexity := 1

	// Count options
	optionCount := strings.Count(command, "--") + strings.Count(command, "-")
	complexity += optionCount

	// Add complexity for piping and chaining
	complexity += strings.Count(command, "|")
	complexity += strings.Count(command, "&&")
	complexity += strings.Count(command, "||")

	// Add complexity for complex operations
	if strings.Contains(command, "create") {
		complexity += 2
	}
	if strings.Contains(command, "delete") {
		complexity += 3
	}

	return complexity
}

// CommandDictionary provides descriptions for commands
type CommandDictionary struct {
	descriptions map[string]string
}

// NewCommandDictionary creates a new command dictionary
func NewCommandDictionary() *CommandDictionary {
	return &CommandDictionary{
		descriptions: map[string]string{
			"server":      "仮想サーバの管理コマンド。作成、削除、起動、停止などの操作が可能です。",
			"database":    "データベースアプライアンスの管理コマンド。MariaDB、PostgreSQLなどの操作が可能です。",
			"disk":        "ディスクの管理コマンド。作成、削除、接続、切断などの操作が可能です。",
			"switch":      "スイッチの管理コマンド。ネットワーク設定の変更が可能です。",
			"router":      "ルータの管理コマンド。ルーティング設定の変更が可能です。",
			"certificate": "SSL証明書の管理コマンド。証明書の作成、更新、削除が可能です。",
			"ssh-key":     "SSH公開鍵の管理コマンド。鍵の登録、削除が可能です。",
			"auth-status": "認証状態の確認コマンド。APIキーの有効性を確認できます。",
		},
	}
}

// GetDescription returns a description for a command
func (cd *CommandDictionary) GetDescription(command string) string {
	command = strings.TrimSpace(command)

	// Skip comments and empty lines
	if command == "" || strings.HasPrefix(command, "#") {
		return "コメントまたは空行です。"
	}

	// Extract the main command
	parts := strings.Fields(command)
	if len(parts) < 2 || parts[0] != "usacloud" {
		return "usacloudコマンドではありません。"
	}

	subcommand := parts[1]
	if desc, exists := cd.descriptions[subcommand]; exists {
		return desc
	}

	return fmt.Sprintf("usacloud %s コマンドです。詳細についてはusacloud %s --helpを参照してください。", subcommand, subcommand)
}
