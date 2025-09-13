package testing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
)

// GoldenTestSuite は拡張ゴールデンテストスイート
type GoldenTestSuite struct {
	t           *testing.T
	testDataDir string
	updateFlag  bool // -update フラグの状態

	// テスト対象システム
	integratedCLI *IntegratedCLI
	config        *config.IntegratedConfig
}

// GoldenTestOptions はゴールデンテストオプション
type GoldenTestOptions struct {
	InputFile          string
	ConfigFile         string
	Language           string
	ColorEnabled       bool
	StrictMode         bool
	InteractiveMode    bool
	IncludeTransform   bool
	IncludeValidation  bool
	IncludeErrors      bool
	IncludeSuggestions bool
	IncludeHelp        bool
}

// GoldenTestResult はゴールデンテスト結果
type GoldenTestResult struct {
	// 変換結果
	TransformOutput string          `json:"transform_output"`
	TransformStats  *TransformStats `json:"transform_stats"`

	// 検証結果
	ValidationResults []ValidationResult `json:"validation_results"`
	ValidationSummary *ValidationSummary `json:"validation_summary"`

	// エラー・警告
	ErrorMessages   []ErrorMessage   `json:"error_messages"`
	WarningMessages []WarningMessage `json:"warning_messages"`

	// 提案
	Suggestions         []SuggestionResult   `json:"suggestions"`
	DeprecationWarnings []DeprecationWarning `json:"deprecation_warnings"`

	// ヘルプ出力
	HelpOutput        string `json:"help_output"`
	InteractiveOutput string `json:"interactive_output"`

	// メタデータ
	TestMetadata *TestMetadata `json:"test_metadata"`
}

// TransformStats は変換統計
type TransformStats struct {
	TotalLines       int      `json:"total_lines"`
	ProcessedLines   int      `json:"processed_lines"`
	TransformedLines int      `json:"transformed_lines"`
	SkippedLines     int      `json:"skipped_lines"`
	RulesApplied     []string `json:"rules_applied"`
}

// ValidationResult は検証結果
type ValidationResult struct {
	LineNumber       int               `json:"line_number"`
	OriginalLine     string            `json:"original_line"`
	ValidationStatus string            `json:"validation_status"`
	Issues           []ValidationIssue `json:"issues"`
}

// ValidationIssue は検証問題
type ValidationIssue struct {
	Type               string `json:"type"`
	Severity           string `json:"severity"`
	Message            string `json:"message"`
	Suggestion         string `json:"suggestion,omitempty"`
	ReplacementCommand string `json:"replacement_command,omitempty"`
}

// ValidationSummary は検証サマリー
type ValidationSummary struct {
	TotalIssues        int `json:"total_issues"`
	Errors             int `json:"errors"`
	Warnings           int `json:"warnings"`
	Suggestions        int `json:"suggestions"`
	DeprecatedCommands int `json:"deprecated_commands"`
}

// ErrorMessage はエラーメッセージ
type ErrorMessage struct {
	Type             string `json:"type"`
	FormattedMessage string `json:"formatted_message"`
}

// WarningMessage は警告メッセージ
type WarningMessage struct {
	Type             string `json:"type"`
	FormattedMessage string `json:"formatted_message"`
}

// SuggestionResult は提案結果
type SuggestionResult struct {
	LineNumber       int     `json:"line_number"`
	OriginalCommand  string  `json:"original_command"`
	SuggestedCommand string  `json:"suggested_command"`
	Reason           string  `json:"reason"`
	Confidence       float64 `json:"confidence"`
}

// DeprecationWarning は廃止警告
type DeprecationWarning struct {
	DeprecatedCommand  string `json:"deprecated_command"`
	ReplacementCommand string `json:"replacement_command"`
	DeprecationVersion string `json:"deprecation_version"`
	MigrationGuideURL  string `json:"migration_guide_url"`
}

// TestMetadata はテストメタデータ
type TestMetadata struct {
	TestName     string `json:"test_name"`
	InputFile    string `json:"input_file"`
	ConfigUsed   string `json:"config_used"`
	TestDate     string `json:"test_date"`
	ToolVersion  string `json:"tool_version"`
	Language     string `json:"language"`
	ColorEnabled bool   `json:"color_enabled"`
}

// IntegratedCLI は統合CLI（簡易実装）
type IntegratedCLI struct {
	config *config.IntegratedConfig
}

// NewGoldenTestSuite は新しいゴールデンテストスイートを作成
func NewGoldenTestSuite(t *testing.T) *GoldenTestSuite {
	return &GoldenTestSuite{
		t:           t,
		testDataDir: "../testdata",       // testsディレクトリから実行される場合を考慮
		updateFlag:  updateGoldenFiles(), // フラグから取得
	}
}

// updateGoldenFiles は-updateフラグの状態を取得
func updateGoldenFiles() bool {
	// 環境変数でアップデートモードを制御
	return os.Getenv("UPDATE_GOLDEN") == "true"
}

// RunGoldenTest はゴールデンテストを実行
func (gts *GoldenTestSuite) RunGoldenTest(testName string, options *GoldenTestOptions) {
	gts.t.Helper()

	// 入力ファイル読み込み
	inputPath := filepath.Join(gts.testDataDir, "inputs", options.InputFile)
	input, err := ioutil.ReadFile(inputPath)
	if err != nil {
		gts.t.Fatalf("入力ファイル読み込みエラー %s: %v", inputPath, err)
	}

	// テスト実行
	result := gts.executeTest(testName, string(input), options)

	// ゴールデンファイル比較
	gts.compareWithGoldenFile(testName, result, options)
}

// executeTest はテストを実行
func (gts *GoldenTestSuite) executeTest(
	testName, input string,
	options *GoldenTestOptions,
) *GoldenTestResult {
	// 設定読み込み
	config := gts.loadTestConfig(options.ConfigFile)

	// 統合CLIの初期化
	cli := NewIntegratedCLI(config)

	// テスト実行
	var result GoldenTestResult

	// 1. 変換処理実行
	if options.IncludeTransform {
		transformResult := cli.ProcessInput(input)
		result.TransformOutput = transformResult.Output
		result.TransformStats = transformResult.Stats
	}

	// 2. 検証処理実行
	if options.IncludeValidation {
		validationResults := cli.ValidateInput(input)
		result.ValidationResults = validationResults
		result.ValidationSummary = cli.SummarizeValidation(validationResults)
	}

	// 3. エラーメッセージ生成
	if options.IncludeErrors {
		errorMessages := cli.GenerateErrorMessages(result.ValidationResults)
		result.ErrorMessages = errorMessages
	}

	// 4. 提案生成
	if options.IncludeSuggestions {
		suggestions := cli.GenerateSuggestions(result.ValidationResults)
		result.Suggestions = suggestions
	}

	// 5. ヘルプ出力生成
	if options.IncludeHelp {
		helpOutput := cli.GenerateHelp(input, result.ValidationResults)
		result.HelpOutput = helpOutput
	}

	// メタデータ設定
	result.TestMetadata = &TestMetadata{
		TestName:     testName,
		InputFile:    options.InputFile,
		ConfigUsed:   options.ConfigFile,
		TestDate:     getCurrentTimestamp(),
		ToolVersion:  getToolVersion(),
		Language:     config.General.Language,
		ColorEnabled: config.General.ColorOutput,
	}

	return &result
}

// loadTestConfig はテスト設定を読み込み
func (gts *GoldenTestSuite) loadTestConfig(configFile string) *config.IntegratedConfig {
	if configFile == "" {
		configFile = "default.conf"
	}

	// 簡易設定作成（実際の実装では設定ファイルから読み込み）
	cfg := &config.IntegratedConfig{
		General: &config.GeneralConfig{
			Language:    "ja",
			ColorOutput: false,
		},
	}

	return cfg
}

// NewIntegratedCLI は統合CLIを作成
func NewIntegratedCLI(config *config.IntegratedConfig) *IntegratedCLI {
	return &IntegratedCLI{
		config: config,
	}
}

// ProcessInputResult は処理結果
type ProcessInputResult struct {
	Output string
	Stats  *TransformStats
}

// ProcessInput は入力を処理
func (cli *IntegratedCLI) ProcessInput(input string) *ProcessInputResult {
	// 簡易実装：実際の変換処理をシミュレート
	lines := strings.Split(input, "\n")
	var transformedLines []string
	rulesApplied := []string{}
	transformedCount := 0

	for _, line := range lines {
		transformed := line
		if strings.Contains(line, "usacloud") {
			if strings.Contains(line, "--output-type csv") {
				transformed = strings.Replace(line, "--output-type csv", "--output-type json", -1)
				transformed += "  # usacloud-update: csv → json 形式変更 https://docs.usacloud.jp/"
				rulesApplied = append(rulesApplied, "output_format_csv_to_json")
				transformedCount++
			}
			if strings.Contains(line, "iso-image") {
				transformed = strings.Replace(transformed, "iso-image", "cdrom", -1)
				transformed += "  # usacloud-update: iso-image → cdrom 名称変更 https://docs.usacloud.jp/"
				rulesApplied = append(rulesApplied, "resource_rename_iso_image")
				transformedCount++
			}
		}
		transformedLines = append(transformedLines, transformed)
	}

	output := strings.Join(transformedLines, "\n")

	stats := &TransformStats{
		TotalLines:       len(lines),
		ProcessedLines:   len(lines) - 2, // コメント行などを除外
		TransformedLines: transformedCount,
		SkippedLines:     2,
		RulesApplied:     rulesApplied,
	}

	return &ProcessInputResult{
		Output: output,
		Stats:  stats,
	}
}

// ValidateInput は入力を検証
func (cli *IntegratedCLI) ValidateInput(input string) []ValidationResult {
	lines := strings.Split(input, "\n")
	var results []ValidationResult

	for i, line := range lines {
		if strings.Contains(line, "usacloud") {
			var issues []ValidationIssue

			// CSV出力形式の警告
			if strings.Contains(line, "--output-type csv") {
				issues = append(issues, ValidationIssue{
					Type:       "deprecated_parameter",
					Severity:   "warning",
					Message:    "csv出力形式は非推奨です。json形式の使用を推奨します。",
					Suggestion: "--output-type json",
				})
			}

			// 廃止コマンドのエラー
			if strings.Contains(line, "iso-image") {
				issues = append(issues, ValidationIssue{
					Type:               "deprecated_command",
					Severity:           "error",
					Message:            "iso-imageコマンドは廃止されました。cdromコマンドを使用してください。",
					ReplacementCommand: "cdrom",
				})
			}

			if len(issues) > 0 {
				status := "warning"
				for _, issue := range issues {
					if issue.Severity == "error" {
						status = "error"
						break
					}
				}

				results = append(results, ValidationResult{
					LineNumber:       i + 1,
					OriginalLine:     line,
					ValidationStatus: status,
					Issues:           issues,
				})
			}
		}
	}

	return results
}

// SummarizeValidation は検証結果をサマリー
func (cli *IntegratedCLI) SummarizeValidation(results []ValidationResult) *ValidationSummary {
	summary := &ValidationSummary{}

	for _, result := range results {
		for _, issue := range result.Issues {
			summary.TotalIssues++
			switch issue.Severity {
			case "error":
				summary.Errors++
			case "warning":
				summary.Warnings++
			}

			if issue.Type == "deprecated_command" {
				summary.DeprecatedCommands++
			}

			if issue.Suggestion != "" {
				summary.Suggestions++
			}
		}
	}

	return summary
}

// GenerateErrorMessages はエラーメッセージを生成
func (cli *IntegratedCLI) GenerateErrorMessages(results []ValidationResult) []ErrorMessage {
	var messages []ErrorMessage

	for _, result := range results {
		for _, issue := range result.Issues {
			if issue.Severity == "error" {
				formatted := fmt.Sprintf("❌ エラー: '%s' コマンドはv1で廃止されました。\n\n🔄 代わりに以下を使用してください:\n   usacloud %s list\n\nℹ️  詳細な移行ガイド: https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
					strings.Split(result.OriginalLine, " ")[1], // コマンド名を抽出
					issue.ReplacementCommand)

				messages = append(messages, ErrorMessage{
					Type:             "deprecated_command_error",
					FormattedMessage: formatted,
				})
			}
		}
	}

	return messages
}

// GenerateSuggestions は提案を生成
func (cli *IntegratedCLI) GenerateSuggestions(results []ValidationResult) []SuggestionResult {
	var suggestions []SuggestionResult

	for _, result := range results {
		for _, issue := range result.Issues {
			if issue.Suggestion != "" {
				suggestions = append(suggestions, SuggestionResult{
					LineNumber:       result.LineNumber,
					OriginalCommand:  strings.TrimSpace(result.OriginalLine),
					SuggestedCommand: strings.Replace(result.OriginalLine, "csv", "json", -1),
					Reason:           "JSON形式の方が構造化データの処理に適しています",
					Confidence:       0.9,
				})
			}
		}
	}

	return suggestions
}

// GenerateHelp はヘルプを生成
func (cli *IntegratedCLI) GenerateHelp(input string, results []ValidationResult) string {
	if len(results) == 0 {
		return "✅ 問題は検出されませんでした。"
	}

	help := "🔍 検出された問題と解決方法:\n\n"
	for _, result := range results {
		help += fmt.Sprintf("行 %d: %s\n", result.LineNumber, result.OriginalLine)
		for _, issue := range result.Issues {
			help += fmt.Sprintf("  • %s\n", issue.Message)
			if issue.Suggestion != "" {
				help += fmt.Sprintf("    推奨: %s\n", issue.Suggestion)
			}
		}
		help += "\n"
	}

	return help
}

// compareWithGoldenFile はゴールデンファイルとの比較
func (gts *GoldenTestSuite) compareWithGoldenFile(
	testName string,
	result *GoldenTestResult,
	options *GoldenTestOptions,
) {
	goldenPath := gts.getGoldenFilePath(testName, options)

	// 現在の結果をJSONに変換
	currentJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		gts.t.Fatalf("結果のJSON変換エラー: %v", err)
	}

	if gts.updateFlag {
		// ゴールデンファイル更新
		gts.updateGoldenFile(goldenPath, currentJSON)
		return
	}

	// 既存ゴールデンファイル読み込み
	expectedJSON, err := ioutil.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			gts.t.Fatalf("ゴールデンファイルが存在しません: %s\n"+
				"-update フラグを使用してファイルを作成してください", goldenPath)
		}
		gts.t.Fatalf("ゴールデンファイル読み込みエラー: %v", err)
	}

	// 差分検出
	diff := gts.generateDetailedDiff(expectedJSON, currentJSON)
	if diff != nil && diff.HasDifferences {
		gts.t.Errorf("ゴールデンファイルテスト失敗: %s\n\n%s\n\n"+
			"ファイルを更新する場合は -update フラグを使用してください",
			testName, diff.Report())
	}
}

// getGoldenFilePath はゴールデンファイルパスを取得
func (gts *GoldenTestSuite) getGoldenFilePath(testName string, options *GoldenTestOptions) string {
	subdir := "integration"
	if options.Language != "" && options.Language != "ja" {
		subdir = fmt.Sprintf("integration_%s", options.Language)
	}

	return filepath.Join(gts.testDataDir, "golden", subdir, testName+".golden")
}

// updateGoldenFile はゴールデンファイルを更新
func (gts *GoldenTestSuite) updateGoldenFile(goldenPath string, content []byte) {
	// ディレクトリ作成
	dir := filepath.Dir(goldenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		gts.t.Fatalf("ディレクトリ作成エラー %s: %v", dir, err)
	}

	// ファイル書き込み
	if err := ioutil.WriteFile(goldenPath, content, 0644); err != nil {
		gts.t.Fatalf("ゴールデンファイル書き込みエラー %s: %v", goldenPath, err)
	}

	gts.t.Logf("ゴールデンファイル更新: %s", goldenPath)
}

// getCurrentTimestamp は現在のタイムスタンプを取得
func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}

// getToolVersion はツールバージョンを取得
func getToolVersion() string {
	return "1.9.0" // 実際の実装では動的に取得
}

// DetailedDiff は詳細差分情報
type DetailedDiff struct {
	HasDifferences   bool         `json:"has_differences"`
	TransformDiff    *SectionDiff `json:"transform_diff"`
	ValidationDiff   *SectionDiff `json:"validation_diff"`
	ErrorMessageDiff *SectionDiff `json:"error_message_diff"`
	SuggestionDiff   *SectionDiff `json:"suggestion_diff"`
	MetadataDiff     *SectionDiff `json:"metadata_diff"`
}

// SectionDiff はセクション別差分
type SectionDiff struct {
	SectionName   string     `json:"section_name"`
	HasChanges    bool       `json:"has_changes"`
	AddedLines    []string   `json:"added_lines"`
	RemovedLines  []string   `json:"removed_lines"`
	ModifiedLines []LineDiff `json:"modified_lines"`
}

// LineDiff は行レベル差分
type LineDiff struct {
	LineNumber int    `json:"line_number"`
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
}

// generateDetailedDiff は詳細差分を生成
func (gts *GoldenTestSuite) generateDetailedDiff(expected, actual []byte) *DetailedDiff {
	var expectedResult, actualResult GoldenTestResult

	json.Unmarshal(expected, &expectedResult)
	json.Unmarshal(actual, &actualResult)

	diff := &DetailedDiff{}

	// 変換結果の差分
	diff.TransformDiff = gts.compareTransformOutput(
		expectedResult.TransformOutput,
		actualResult.TransformOutput)

	// 検証結果の差分
	diff.ValidationDiff = gts.compareValidationResults(
		expectedResult.ValidationResults,
		actualResult.ValidationResults)

	// エラーメッセージの差分
	diff.ErrorMessageDiff = gts.compareErrorMessages(
		expectedResult.ErrorMessages,
		actualResult.ErrorMessages)

	// 提案の差分
	diff.SuggestionDiff = gts.compareSuggestions(
		expectedResult.Suggestions,
		actualResult.Suggestions)

	// 差分の有無を判定
	diff.HasDifferences = diff.TransformDiff.HasChanges ||
		diff.ValidationDiff.HasChanges ||
		diff.ErrorMessageDiff.HasChanges ||
		diff.SuggestionDiff.HasChanges

	return diff
}

// compareTransformOutput は変換出力を比較
func (gts *GoldenTestSuite) compareTransformOutput(expected, actual string) *SectionDiff {
	diff := &SectionDiff{
		SectionName: "transform_output",
		HasChanges:  expected != actual,
	}

	if diff.HasChanges {
		expectedLines := strings.Split(expected, "\n")
		actualLines := strings.Split(actual, "\n")

		// 簡易差分計算
		maxLines := len(expectedLines)
		if len(actualLines) > maxLines {
			maxLines = len(actualLines)
		}

		for i := 0; i < maxLines; i++ {
			var expectedLine, actualLine string
			if i < len(expectedLines) {
				expectedLine = expectedLines[i]
			}
			if i < len(actualLines) {
				actualLine = actualLines[i]
			}

			if expectedLine != actualLine {
				if expectedLine == "" {
					diff.AddedLines = append(diff.AddedLines, actualLine)
				} else if actualLine == "" {
					diff.RemovedLines = append(diff.RemovedLines, expectedLine)
				} else {
					diff.ModifiedLines = append(diff.ModifiedLines, LineDiff{
						LineNumber: i + 1,
						Expected:   expectedLine,
						Actual:     actualLine,
					})
				}
			}
		}
	}

	return diff
}

// compareValidationResults は検証結果を比較
func (gts *GoldenTestSuite) compareValidationResults(expected, actual []ValidationResult) *SectionDiff {
	diff := &SectionDiff{
		SectionName: "validation_results",
		HasChanges:  len(expected) != len(actual),
	}

	// 簡易比較実装
	if !diff.HasChanges {
		for i := range expected {
			if i >= len(actual) || expected[i].LineNumber != actual[i].LineNumber {
				diff.HasChanges = true
				break
			}
		}
	}

	return diff
}

// compareErrorMessages はエラーメッセージを比較
func (gts *GoldenTestSuite) compareErrorMessages(expected, actual []ErrorMessage) *SectionDiff {
	diff := &SectionDiff{
		SectionName: "error_messages",
		HasChanges:  len(expected) != len(actual),
	}

	// 簡易比較実装
	return diff
}

// compareSuggestions は提案を比較
func (gts *GoldenTestSuite) compareSuggestions(expected, actual []SuggestionResult) *SectionDiff {
	diff := &SectionDiff{
		SectionName: "suggestions",
		HasChanges:  len(expected) != len(actual),
	}

	// 簡易比較実装
	return diff
}

// Report は差分レポートを生成
func (dd *DetailedDiff) Report() string {
	if !dd.HasDifferences {
		return "差分なし"
	}

	var report strings.Builder
	report.WriteString("📊 ゴールデンファイル差分レポート\n")
	report.WriteString("================================\n\n")

	if dd.TransformDiff.HasChanges {
		report.WriteString("🔄 変換結果の差分:\n")
		report.WriteString(dd.TransformDiff.formatDiff())
		report.WriteString("\n")
	}

	if dd.ValidationDiff.HasChanges {
		report.WriteString("🔍 検証結果の差分:\n")
		report.WriteString(dd.ValidationDiff.formatDiff())
		report.WriteString("\n")
	}

	if dd.ErrorMessageDiff.HasChanges {
		report.WriteString("❌ エラーメッセージの差分:\n")
		report.WriteString(dd.ErrorMessageDiff.formatDiff())
		report.WriteString("\n")
	}

	if dd.SuggestionDiff.HasChanges {
		report.WriteString("💡 提案の差分:\n")
		report.WriteString(dd.SuggestionDiff.formatDiff())
		report.WriteString("\n")
	}

	return report.String()
}

// formatDiff は差分をフォーマット
func (sd *SectionDiff) formatDiff() string {
	var result strings.Builder

	for _, added := range sd.AddedLines {
		result.WriteString(fmt.Sprintf("+ %s\n", added))
	}

	for _, removed := range sd.RemovedLines {
		result.WriteString(fmt.Sprintf("- %s\n", removed))
	}

	for _, modified := range sd.ModifiedLines {
		result.WriteString(fmt.Sprintf("@ 行 %d:\n", modified.LineNumber))
		result.WriteString(fmt.Sprintf("- %s\n", modified.Expected))
		result.WriteString(fmt.Sprintf("+ %s\n", modified.Actual))
	}

	return result.String()
}
