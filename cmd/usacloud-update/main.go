package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/transform"
	"github.com/armaniacs/usacloud-update/internal/tui"
	"github.com/armaniacs/usacloud-update/internal/validation"
	"github.com/fatih/color"
)

const version = "1.9.0"

// ProcessResult は統合された処理結果
type ProcessResult struct {
	LineNumber       int
	OriginalLine     string
	TransformResult  *transform.Result
	ValidationResult *ValidationResult
}

// ValidationResult は検証結果
type ValidationResult struct {
	LineNumber  int
	Line        string
	Issues      []ValidationIssue
	Suggestions []validation.SimilarityResult
}

// ValidationIssue は検証で発見された問題
type ValidationIssue struct {
	Type      IssueType
	Message   string
	Component string // 問題のあるコマンド・サブコマンド名
}

// IssueType は問題タイプ
type IssueType int

const (
	IssueParseError IssueType = iota
	IssueInvalidMainCommand
	IssueInvalidSubCommand
	IssueDeprecatedCommand
	IssueSyntaxError
)

// HasErrors は ValidationResult がエラーを持つかチェック
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Issues) > 0
}

// GetErrorSummary は ValidationResult のエラー要約を取得
func (vr *ValidationResult) GetErrorSummary() string {
	if len(vr.Issues) == 0 {
		return ""
	}
	return vr.Issues[0].Message
}

// FileAnalysis はファイル分析結果
type FileAnalysis struct {
	TotalLines    int
	UsacloudLines int
	Issues        []ValidationResult
}

// InteractiveIssue はインタラクティブ修正用の問題情報
type InteractiveIssue struct {
	LineNumber    int
	Description   string
	CurrentCode   string
	SuggestedCode string
	Reason        string
}

// Config は統合された設定
type Config struct {
	// 既存設定
	InputPath  string
	OutputPath string
	ShowStats  bool

	// 新しい検証設定
	ValidateOnly     bool
	StrictValidation bool
	InteractiveMode  bool
	HelpMode         string
	SuggestionLevel  int
	SkipDeprecated   bool
	ColorEnabled     bool
	LanguageCode     string

	// サンドボックス設定
	SandboxMode        bool
	DryRun             bool
	BatchMode          bool
	SandboxInteractive bool

	// 設定ファイル
	ConfigFile string
}

// ValidationConfig は検証システム設定
type ValidationConfig struct {
	MaxSuggestions        int
	MaxDistance           int
	EnableTypoDetection   bool
	EnableInteractiveHelp bool
	ErrorFormat           string
	LogLevel              string
}

// IntegratedCLI は統合CLIインターフェース
type IntegratedCLI struct {
	config             *Config
	validationConfig   *ValidationConfig
	transformEngine    *transform.Engine
	mainValidator      *validation.MainCommandValidator
	subValidator       *validation.SubcommandValidator
	deprecatedDetector *validation.DeprecatedCommandDetector
	similarSuggester   *validation.SimilarCommandSuggester
	errorFormatter     *validation.ComprehensiveErrorFormatter
	helpSystem         *validation.UserFriendlyHelpSystem
}

// NewIntegratedCLI は新しい統合CLIを作成
func NewIntegratedCLI() *IntegratedCLI {
	cfg := parseFlags()
	valCfg := loadValidationConfig()

	// 検証システムの初期化
	mainValidator := validation.NewMainCommandValidator()
	subValidator := validation.NewSubcommandValidator(mainValidator)
	deprecatedDetector := validation.NewDeprecatedCommandDetector()
	similarSuggester := validation.NewSimilarCommandSuggester(valCfg.MaxDistance, valCfg.MaxSuggestions)
	errorFormatter := validation.NewDefaultComprehensiveErrorFormatter()
	helpSystem := validation.NewDefaultUserFriendlyHelpSystem()

	cli := &IntegratedCLI{
		config:             cfg,
		validationConfig:   valCfg,
		transformEngine:    transform.NewDefaultEngine(),
		mainValidator:      mainValidator,
		subValidator:       subValidator,
		deprecatedDetector: deprecatedDetector,
		similarSuggester:   similarSuggester,
		errorFormatter:     errorFormatter,
		helpSystem:         helpSystem,
	}

	return cli
}

// runValidationMode は検証のみまたはインタラクティブモードを実行
func (cli *IntegratedCLI) runValidationMode() error {
	// 入力ファイル読み込み
	content, err := cli.readInputFile()
	if err != nil {
		return fmt.Errorf("入力ファイル読み込みエラー: %w", err)
	}

	if cli.config.InteractiveMode {
		return cli.runInteractiveValidation(content)
	}

	// 検証のみモード
	return cli.performValidationOnly(content)
}

// runIntegratedMode は変換と検証を統合したモードを実行
func (cli *IntegratedCLI) runIntegratedMode() error {
	// 入力ファイル読み込み
	content, err := cli.readInputFile()
	if err != nil {
		return fmt.Errorf("入力ファイル読み込みエラー: %w", err)
	}

	// バッチモード処理
	results, err := cli.processLines(content)
	if err != nil {
		return fmt.Errorf("処理エラー: %w", err)
	}

	// 出力生成
	err = cli.generateOutput(results)
	if err != nil {
		return err
	}

	// 変換完了メッセージを標準出力に出力
	fmt.Println("✅ 変換完了")

	return nil
}

// readInputFile は入力ファイルを読み込み
func (cli *IntegratedCLI) readInputFile() ([]string, error) {
	var r io.Reader = os.Stdin
	if cli.config.InputPath != "-" {
		f, err := os.Open(cli.config.InputPath)
		if err != nil {
			// Localize error messages based on error type
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("ファイルが見つかりません: %s", cli.config.InputPath)
			}
			if os.IsPermission(err) {
				return nil, fmt.Errorf("読み取り権限がありません: %s", cli.config.InputPath)
			}
			return nil, fmt.Errorf("入力ファイルを開けません: %w", err)
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				// Log close error but don't override main error
			}
		}()
		r = f

		// Check for binary content by reading first few bytes
		firstBytes := make([]byte, 512)
		n, _ := f.Read(firstBytes)
		if n > 0 {
			// Check if content contains null bytes (binary indicator)
			for i := 0; i < n; i++ {
				if firstBytes[i] == 0 {
					return nil, fmt.Errorf("バイナリファイルは処理できません: %s", cli.config.InputPath)
				}
			}
		}

		// Reset file position to beginning
		if _, err := f.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("ファイル位置のリセットに失敗: %w", err)
		}
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Check for empty file
	if cli.config.InputPath != "-" && len(lines) == 0 {
		return nil, fmt.Errorf("空のファイルは処理できません: %s", cli.config.InputPath)
	}

	return lines, nil
}

// processLines は行ごとの処理を実行（変換と検証の統合）
func (cli *IntegratedCLI) processLines(lines []string) ([]*ProcessResult, error) {
	var results []*ProcessResult

	for lineNumber, line := range lines {
		lineNum := lineNumber + 1

		// 既存の変換処理
		transformResult := cli.transformEngine.Apply(line)

		// 新しい検証処理（変換前）
		var validationResult *ValidationResult
		if !cli.config.SkipDeprecated {
			validationResult = cli.validateLine(line, lineNum)

			// 厳格検証モードでエラーがあれば停止
			if cli.config.StrictValidation && validationResult != nil && validationResult.HasErrors() {
				return nil, fmt.Errorf("行 %d で検証エラー: %s", lineNum, validationResult.GetErrorSummary())
			}
		}

		// 統合結果の作成
		result := &ProcessResult{
			LineNumber:       lineNum,
			OriginalLine:     line,
			TransformResult:  &transformResult,
			ValidationResult: validationResult,
		}

		results = append(results, result)

		// リアルタイム出力（既存機能）
		if transformResult.Changed && cli.config.ShowStats {
			cli.outputColorizedChange(result.TransformResult, lineNum)
		}
	}

	return results, nil
}

// validateLine は単一行の検証を実行
func (cli *IntegratedCLI) validateLine(line string, lineNumber int) *ValidationResult {
	// usacloudコマンドでない行はスキップ
	if !strings.Contains(line, "usacloud") {
		return nil
	}

	// コマンド解析
	parser := validation.NewParser()
	parsed, err := parser.Parse(line)
	if err != nil {
		return &ValidationResult{
			LineNumber: lineNumber,
			Line:       line,
			Issues:     []ValidationIssue{{Type: IssueParseError, Message: err.Error(), Component: ""}},
		}
	}

	// 空のメインコマンドをチェック
	if parsed.MainCommand == "" {
		return nil // メインコマンドがない場合はスキップ
	}

	var issues []ValidationIssue
	var suggestions []validation.SimilarityResult

	// 廃止コマンド検証を最初に行う
	if cli.deprecatedDetector.IsDeprecated(parsed.MainCommand) {
		deprecatedInfo := cli.deprecatedDetector.Detect(parsed.MainCommand)
		var message string

		// 代替コマンドがある場合はそれを含めてメッセージを作成
		if deprecatedInfo.ReplacementCommand != "" {
			message = fmt.Sprintf("'%s' は廃止されました。代わりに '%s' を使用してください", parsed.MainCommand, deprecatedInfo.ReplacementCommand)
			// サジェスションも追加
			suggestions = append(suggestions, validation.SimilarityResult{
				Command: deprecatedInfo.ReplacementCommand,
				Score:   1.0,
			})
		} else {
			// 代替コマンドがない場合は元のメッセージを使用
			message = fmt.Sprintf("'%s' は廃止されたコマンドです: %s", parsed.MainCommand, deprecatedInfo.Message)
		}

		issues = append(issues, ValidationIssue{
			Type:      IssueDeprecatedCommand,
			Message:   message,
			Component: parsed.MainCommand,
		})

		// 廃止コマンドの場合でもサブコマンドを検証する（元のコマンド名で報告）
		if parsed.SubCommand != "" {
			// 廃止コマンドのサブコマンドは代替コマンドに対して検証
			replacementCommand := deprecatedInfo.ReplacementCommand
			if replacementCommand != "" && !cli.subValidator.IsValidSubcommand(replacementCommand, parsed.SubCommand) {
				issues = append(issues, ValidationIssue{
					Type:      IssueInvalidSubCommand,
					Message:   fmt.Sprintf("'%s' は %s コマンドの有効なサブコマンドではありません", parsed.SubCommand, parsed.MainCommand),
					Component: parsed.SubCommand,
				})

				// サブコマンド提案を取得（代替コマンド用）
				subSuggestions := cli.similarSuggester.SuggestSubcommands(replacementCommand, parsed.SubCommand)
				suggestions = append(suggestions, subSuggestions...)
			} else if replacementCommand == "" {
				// 代替コマンドがない場合、サブコマンドも無効として扱う
				issues = append(issues, ValidationIssue{
					Type:      IssueInvalidSubCommand,
					Message:   fmt.Sprintf("'%s' は無効なサブコマンドです（メインコマンド '%s' が廃止されています）", parsed.SubCommand, parsed.MainCommand),
					Component: parsed.SubCommand,
				})
			}
		}
	} else {
		// 廃止されていない場合のみメインコマンドの有効性を検証
		mainValidationResult := cli.mainValidator.Validate(parsed.MainCommand)
		if !mainValidationResult.IsValid {
			issues = append(issues, ValidationIssue{
				Type:      IssueInvalidMainCommand,
				Message:   fmt.Sprintf("'%s' は有効なusacloudコマンドではありません", parsed.MainCommand),
				Component: parsed.MainCommand,
			})

			// 類似提案を取得
			suggestions = cli.similarSuggester.SuggestMainCommands(parsed.MainCommand)
		} else if mainValidationResult.Message != "" {
			// Case sensitivity issue - treat as invalid for strict validation
			issues = append(issues, ValidationIssue{
				Type:      IssueInvalidMainCommand,
				Message:   fmt.Sprintf("'%s' は有効なusacloudコマンドではありません", parsed.MainCommand),
				Component: parsed.MainCommand,
			})

			// Add suggestions from validation result (correct case)
			for _, suggestion := range mainValidationResult.Suggestions {
				suggestions = append(suggestions, validation.SimilarityResult{
					Command: suggestion,
					Score:   1.0,
				})
			}
		} else {
			// メインコマンドが有効な場合のみサブコマンド検証を行う
			if parsed.SubCommand != "" && !cli.subValidator.IsValidSubcommand(parsed.MainCommand, parsed.SubCommand) {
				issues = append(issues, ValidationIssue{
					Type:      IssueInvalidSubCommand,
					Message:   fmt.Sprintf("'%s' は %s コマンドの有効なサブコマンドではありません", parsed.SubCommand, parsed.MainCommand),
					Component: parsed.SubCommand,
				})

				// サブコマンド提案を取得
				subSuggestions := cli.similarSuggester.SuggestSubcommands(parsed.MainCommand, parsed.SubCommand)
				suggestions = append(suggestions, subSuggestions...)
			}
		}
	}

	if len(issues) == 0 {
		return nil
	}

	return &ValidationResult{
		LineNumber:  lineNumber,
		Line:        line,
		Issues:      issues,
		Suggestions: suggestions,
	}
}

// outputColorizedChange は変更をカラー出力
func (cli *IntegratedCLI) outputColorizedChange(result *transform.Result, lineNumber int) {
	for _, change := range result.Changes {
		fmt.Fprintf(os.Stderr, color.YellowString("#L%-5d %s => %s [%s]\n"),
			lineNumber, change.Before, change.After, change.RuleName)
	}
}

// generateOutput は出力を生成
func (cli *IntegratedCLI) generateOutput(results []*ProcessResult) error {
	var outLines []string

	for _, result := range results {
		outLines = append(outLines, result.TransformResult.Line)
	}

	output := strings.Join(append([]string{transform.GeneratedHeader()}, outLines...), "\n") + "\n"

	var w io.Writer = os.Stdout
	if cli.config.OutputPath != "-" {
		f, err := os.Create(cli.config.OutputPath)
		if err != nil {
			// Localize error messages based on error type
			if os.IsPermission(err) {
				return fmt.Errorf("出力ファイル書き込み失敗: 権限が不足しています: %s", cli.config.OutputPath)
			}
			if strings.Contains(err.Error(), "is a directory") {
				return fmt.Errorf("出力先がディレクトリです: %s", cli.config.OutputPath)
			}
			return fmt.Errorf("出力ファイル作成エラー: %w", err)
		}
		defer f.Close()
		w = f
	}

	_, err := io.WriteString(w, output)
	return err
}

// performValidationOnly は検証のみを実行
func (cli *IntegratedCLI) performValidationOnly(lines []string) error {
	fmt.Fprint(os.Stderr, color.CyanString("🔍 検証を実行中...\n\n"))

	var allIssues []ValidationResult

	for lineNumber, line := range lines {
		result := cli.validateLine(line, lineNumber+1)
		if result != nil {
			allIssues = append(allIssues, *result)
		}
	}

	// 結果表示
	if len(allIssues) == 0 {
		// 成功時は標準出力に出力
		fmt.Print(color.GreenString("✅ 検証完了: 問題は見つかりませんでした\n"))
		return nil
	}

	// 構造化されたエラーレポートを出力
	fmt.Fprint(os.Stderr, color.CyanString("📋 検証結果\n"))
	fmt.Fprintf(os.Stderr, color.YellowString("⚠️  %d個の問題が見つかりました:\n\n"), len(allIssues))

	// エラーと警告を分類
	var errorCount, warningCount int
	for _, issue := range allIssues {
		for _, issueDetail := range issue.Issues {
			switch issueDetail.Type {
			case IssueInvalidMainCommand, IssueInvalidSubCommand:
				errorCount++
			case IssueDeprecatedCommand:
				warningCount++
			default:
				errorCount++
			}
		}
	}

	// セクション別レポート
	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, color.RedString("🔴 エラー (%d件) - 重要度: 高\n"), errorCount)
	}
	if warningCount > 0 {
		fmt.Fprintf(os.Stderr, color.YellowString("🟡 警告 (%d件) - 重要度: 中\n"), warningCount)
	}
	fmt.Fprint(os.Stderr, "\n")

	// 詳細なエラー情報を表示
	for _, issue := range allIssues {
		context := &validation.ErrorContext{
			InputCommand:   issue.Line,
			DetectedIssues: convertToValidationIssues(issue.Issues),
			Suggestions:    issue.Suggestions,
		}

		errorMessage := cli.errorFormatter.FormatError(context)
		fmt.Fprint(os.Stderr, errorMessage)
		fmt.Fprint(os.Stderr, "\n")
	}

	return fmt.Errorf("%d個の検証エラーが見つかりました", len(allIssues))
}

// convertToValidationIssues は内部のValidationIssueを検証システムの型に変換
func convertToValidationIssues(issues []ValidationIssue) []validation.ValidationIssue {
	var result []validation.ValidationIssue

	for _, issue := range issues {
		// 対応する変換処理
		validationIssue := validation.ValidationIssue{
			Type:      convertIssueType(issue.Type),
			Severity:  validation.SeverityError, // デフォルト
			Component: issue.Component,          // コマンドやサブコマンド名を設定
			Message:   issue.Message,
			Expected:  []string{},
		}
		result = append(result, validationIssue)
	}

	return result
}

// convertIssueType は内部IssueTypeを検証システムの型に変換
func convertIssueType(issueType IssueType) validation.IssueType {
	switch issueType {
	case IssueInvalidMainCommand:
		return validation.IssueInvalidMainCommand
	case IssueInvalidSubCommand:
		return validation.IssueInvalidSubCommand
	case IssueDeprecatedCommand:
		return validation.IssueDeprecatedCommand
	case IssueSyntaxError:
		return validation.IssueSyntaxError
	default:
		return validation.IssueInvalidMainCommand
	}
}

// runInteractiveValidation はインタラクティブ検証を実行
func (cli *IntegratedCLI) runInteractiveValidation(lines []string) error {
	fmt.Fprint(os.Stderr, color.CyanString("🚀 インタラクティブ検証モードを開始します\n\n"))

	// ファイル分析
	analysis := cli.analyzeFile(lines)

	// 問題点の表示と選択
	issues := cli.identifyIssues(analysis)
	if len(issues) == 0 {
		fmt.Fprint(os.Stderr, color.GreenString("✅ 問題は見つかりませんでした\n"))
		return nil
	}

	selectedIssues := cli.selectIssuesInteractively(issues)

	// 推奨変更の適用
	return cli.applySelectedChanges(selectedIssues)
}

// analyzeFile はファイル全体を分析
func (cli *IntegratedCLI) analyzeFile(lines []string) *FileAnalysis {
	analysis := &FileAnalysis{
		TotalLines:    len(lines),
		UsacloudLines: 0,
		Issues:        []ValidationResult{},
	}

	for lineNumber, line := range lines {
		result := cli.validateLine(line, lineNumber+1)
		if result != nil {
			analysis.Issues = append(analysis.Issues, *result)
		}

		if strings.Contains(line, "usacloud") {
			analysis.UsacloudLines++
		}
	}

	return analysis
}

// identifyIssues は問題を特定
func (cli *IntegratedCLI) identifyIssues(analysis *FileAnalysis) []InteractiveIssue {
	var issues []InteractiveIssue

	for _, validationResult := range analysis.Issues {
		for _, issue := range validationResult.Issues {
			interactiveIssue := InteractiveIssue{
				LineNumber:    validationResult.LineNumber,
				Description:   issue.Message,
				CurrentCode:   validationResult.Line,
				SuggestedCode: cli.generateSuggestedFix(validationResult),
				Reason:        cli.generateReason(issue),
			}
			issues = append(issues, interactiveIssue)
		}
	}

	return issues
}

// generateSuggestedFix は修正提案を生成
func (cli *IntegratedCLI) generateSuggestedFix(result ValidationResult) string {
	// 簡単な修正提案生成
	if len(result.Suggestions) > 0 {
		suggestion := result.Suggestions[0]
		// 元のコマンドを提案で置換
		return strings.Replace(result.Line, extractCommand(result.Line), suggestion.Command, 1)
	}

	return result.Line // 提案がない場合は元のまま
}

// extractCommand は行からコマンド部分を抽出
func extractCommand(line string) string {
	parts := strings.Fields(line)
	for i, part := range parts {
		if part == "usacloud" && i+1 < len(parts) {
			if i+2 < len(parts) {
				return parts[i+1] + " " + parts[i+2] // main + sub
			}
			return parts[i+1] // main のみ
		}
	}
	return ""
}

// generateReason は理由を生成
func (cli *IntegratedCLI) generateReason(issue ValidationIssue) string {
	switch issue.Type {
	case IssueInvalidMainCommand:
		return "指定されたメインコマンドがusacloudでサポートされていません"
	case IssueInvalidSubCommand:
		return "指定されたサブコマンドがこのメインコマンドでサポートされていません"
	case IssueDeprecatedCommand:
		return "このコマンドは廃止されており、新しい代替コマンドの使用が推奨されます"
	default:
		return "構文エラーが検出されました"
	}
}

// selectIssuesInteractively はインタラクティブに問題を選択
func (cli *IntegratedCLI) selectIssuesInteractively(issues []InteractiveIssue) []InteractiveIssue {
	var selected []InteractiveIssue

	fmt.Printf("\n📋 %d個の問題が検出されました:\n\n", len(issues))

	for i, issue := range issues {
		fmt.Printf("  %d. %s (行: %d)\n", i+1, issue.Description, issue.LineNumber)
		fmt.Printf("     現在: %s\n", issue.CurrentCode)
		fmt.Printf("     推奨: %s\n", issue.SuggestedCode)
		fmt.Printf("     理由: %s\n", issue.Reason)

		fmt.Printf("\n     この変更を適用しますか？ [y/N/s(skip)/q(quit)]: ")

		response := cli.readUserInput()
		switch strings.ToLower(response) {
		case "y", "yes":
			selected = append(selected, issue)
			fmt.Printf("     ✅ 適用予定に追加しました\n\n")
		case "s", "skip":
			fmt.Printf("     ⏭️  スキップしました\n\n")
		case "q", "quit":
			fmt.Printf("     🚪 インタラクティブモードを終了します\n")
			return selected
		default:
			fmt.Printf("     ❌ 適用しませんでした\n\n")
		}
	}

	return selected
}

// readUserInput はユーザー入力を読み取り
func (cli *IntegratedCLI) readUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// applySelectedChanges は選択された変更を適用
func (cli *IntegratedCLI) applySelectedChanges(issues []InteractiveIssue) error {
	if len(issues) == 0 {
		fmt.Fprint(os.Stderr, color.YellowString("適用する変更がありません\n"))
		return nil
	}

	fmt.Fprintf(os.Stderr, color.CyanString("🔧 %d個の変更を適用中...\n\n"), len(issues))

	for _, issue := range issues {
		fmt.Printf("行 %d: %s\n", issue.LineNumber, issue.SuggestedCode)
	}

	fmt.Fprint(os.Stderr, color.GreenString("✅ 変更適用完了\n"))
	return nil
}

// parseFlags はフラグから設定を解析
func parseFlags() *Config {
	return &Config{
		InputPath:          *inFile,
		OutputPath:         *outFile,
		ShowStats:          *stats,
		ValidateOnly:       *validateOnly,
		StrictValidation:   *strictValidation,
		InteractiveMode:    *interactiveMode,
		HelpMode:           *helpMode,
		SuggestionLevel:    *suggestionLevel,
		SkipDeprecated:     *skipDeprecated,
		ColorEnabled:       *colorEnabled,
		LanguageCode:       *languageCode,
		SandboxMode:        *sandboxMode,
		DryRun:             *dryRun,
		BatchMode:          *batch,
		SandboxInteractive: *interactive,
		ConfigFile:         *configFile,
	}
}

// loadValidationConfig は検証設定を読み込み
func loadValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxSuggestions:        5,
		MaxDistance:           3,
		EnableTypoDetection:   true,
		EnableInteractiveHelp: true,
		ErrorFormat:           "comprehensive",
		LogLevel:              "info",
	}
}

var (
	inFile      = flag.String("in", "-", "入力ファイルパス ('-'で標準入力)")
	outFile     = flag.String("out", "-", "出力ファイルパス ('-'で標準出力)")
	stats       = flag.Bool("stats", true, "変更の統計情報を標準エラー出力に表示")
	showVersion = flag.Bool("version", false, "バージョン情報を表示")
	showHelp    = flag.Bool("help", false, "ヘルプメッセージを表示")

	// Sandbox functionality flags
	sandboxMode = flag.Bool("sandbox", false, "サンドボックス環境での実際のコマンド実行")
	interactive = flag.Bool("interactive", true, "インタラクティブTUIモード (sandboxとの組み合わせで使用)")
	dryRun      = flag.Bool("dry-run", false, "実際の実行を行わず変換結果のみ表示")
	batch       = flag.Bool("batch", false, "バッチモード: 選択した全コマンドを自動実行")

	// New validation functionality flags
	validateOnly     = flag.Bool("validate-only", false, "検証のみ実行（変換は行わない）")
	strictValidation = flag.Bool("strict-validation", false, "厳格検証モード（エラー発生時に処理を停止）")
	interactiveMode  = flag.Bool("interactive-mode", false, "インタラクティブ検証・修正モード")
	helpMode         = flag.String("help-mode", "enhanced", "ヘルプモード (basic/enhanced/interactive)")
	suggestionLevel  = flag.Int("suggestion-level", 3, "提案レベル設定 (1-5)")
	skipDeprecated   = flag.Bool("skip-deprecated", false, "廃止コマンド警告をスキップ")
	colorEnabled     = flag.Bool("color", true, "カラー出力を有効にする")
	languageCode     = flag.String("language", "ja", "言語設定 (ja/en)")
	configFile       = flag.String("config", "", "設定ファイルパス（指定しない場合はデフォルト設定を使用）")
)

// printHelpMessage prints help message to stdout
func printHelpMessage() {
	fmt.Printf(`usacloud-update v%s

概要:
  usacloud v0、v1.0、v1.1の記述が混在したbashスクリプトを、v1.1で動作するように自動変換します。
  廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動更新し、
  変換できない箇所は適切なコメントと共に手動対応を促します。
  
  --sandboxオプションでSakura Cloudサンドボックス環境での実際のコマンド実行が可能です。

使用方法:
  usacloud-update [オプション]

基本的な使用例:
  # パイプラインで使用
  cat input.sh | usacloud-update > output.sh

  # ファイルを指定して変換
  usacloud-update --in script.sh --out updated_script.sh

  # 変更統計のみ確認（出力は破棄）
  usacloud-update --in script.sh --out /dev/null

  # 統計出力を無効にして変換
  usacloud-update --in script.sh --out updated.sh --stats=false

サンドボックス機能の使用例:
  # インタラクティブTUIでサンドボックス実行
  usacloud-update --sandbox --in script.sh

  # ドライランモード（実行せずに結果確認）
  usacloud-update --sandbox --dry-run --in script.sh

  # バッチモード（全コマンド自動実行）
  usacloud-update --sandbox --batch --in script.sh

  # TUIなしで直接バッチ実行
  usacloud-update --sandbox --interactive=false --batch --in script.sh

環境設定:
  サンドボックス機能を使用するには設定ファイルまたは環境変数が必要です:
  
  【推奨】設定ファイル方式:
    usacloud-update.conf.sample を参考に ~/.config/usacloud-update/usacloud-update.conf を作成
    初回実行時に対話的に作成することも可能
    
    設定ファイルディレクトリのカスタマイズ:
      USACLOUD_UPDATE_CONFIG_DIR=/path/to/config - カスタム設定ディレクトリを指定
  
  環境変数方式（レガシー）:
    SAKURACLOUD_ACCESS_TOKEN、SAKURACLOUD_ACCESS_TOKEN_SECRET

オプション:
`, version)
	flag.PrintDefaults()
	fmt.Printf(`
詳細な使用方法とルールについては README-Usage.md を参照してください。
サンドボックス機能はSakura Cloudのtk1vゾーンを使用します（料金は発生しません）。

困ったときは: usacloud-update --help
`)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `無効なオプションが指定されました。正しい使用方法については --help オプションを参照してください。

usacloud-update v%s

概要:
  usacloud v0、v1.0、v1.1の記述が混在したbashスクリプトを、v1.1で動作するように自動変換します。
  廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動更新し、
  変換できない箇所は適切なコメントと共に手動対応を促します。
  
  --sandboxオプションでSakura Cloudサンドボックス環境での実際のコマンド実行が可能です。

使用方法:
  usacloud-update [オプション]

基本的な使用例:
  # パイプラインで使用
  cat input.sh | usacloud-update > output.sh

  # ファイルを指定して変換
  usacloud-update --in script.sh --out updated_script.sh

  # 変更統計のみ確認（出力は破棄）
  usacloud-update --in script.sh --out /dev/null

  # 統計出力を無効にして変換
  usacloud-update --in script.sh --out updated.sh --stats=false

サンドボックス機能の使用例:
  # インタラクティブTUIでサンドボックス実行
  usacloud-update --sandbox --in script.sh

  # ドライランモード（実行せずに結果確認）
  usacloud-update --sandbox --dry-run --in script.sh

  # バッチモード（全コマンド自動実行）
  usacloud-update --sandbox --batch --in script.sh

  # TUIなしで直接バッチ実行
  usacloud-update --sandbox --interactive=false --batch --in script.sh

環境設定:
  サンドボックス機能を使用するには設定ファイルまたは環境変数が必要です:
  
  【推奨】設定ファイル方式:
    usacloud-update.conf.sample を参考に ~/.config/usacloud-update/usacloud-update.conf を作成
    初回実行時に対話的に作成することも可能
    
    設定ファイルディレクトリのカスタマイズ:
      USACLOUD_UPDATE_CONFIG_DIR=/path/to/config - カスタム設定ディレクトリを指定
  
  環境変数方式（レガシー）:
    SAKURACLOUD_ACCESS_TOKEN、SAKURACLOUD_ACCESS_TOKEN_SECRET

オプション:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
詳細な使用方法とルールについては README-Usage.md を参照してください。
サンドボックス機能はSakura Cloudのtk1vゾーンを使用します（料金は発生しません）。

困ったときは: usacloud-update --help
`)
	}
}

func main() {
	flag.Parse()

	if *showHelp {
		printHelpMessage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("usacloud-update v%s\n", version)
		os.Exit(0)
	}

	// Handle positional arguments for input file
	args := flag.Args()
	if len(args) > 0 && *inFile == "-" {
		*inFile = args[0]
	}

	// Load and validate configuration if --config flag is provided
	if *configFile != "" {
		_, err := config.LoadConfig(*configFile)
		if err != nil {
			if config.IsConfigNotFound(err) {
				fmt.Fprintf(os.Stderr, color.RedString("設定ファイルが見つかりません: %s\n"), *configFile)
				fmt.Fprint(os.Stderr, color.YellowString("デフォルト設定を使用します。\n"))
				fmt.Fprintf(os.Stderr, "修正方法: 設定ファイルのパスを確認してください。\n")
				fmt.Fprintf(os.Stderr, "設定例については README-Usage.md を確認してください。\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, color.RedString("設定ファイルエラー: %v\n"), err)
			fmt.Fprint(os.Stderr, color.YellowString("フォールバック: デフォルト値を使用します。\n"))
			fmt.Fprintf(os.Stderr, "修正方法: 設定ファイルの形式を確認してください。\n")
			fmt.Fprintf(os.Stderr, "設定例については usacloud-update.conf.sample を参照してください。\n")
			os.Exit(1)
		}
	}

	// Create integrated CLI
	cli := NewIntegratedCLI()

	// Handle different modes
	if cli.config.SandboxMode {
		runSandboxMode()
		return
	}

	// Check if validation-only or interactive mode is requested
	if cli.config.ValidateOnly || cli.config.InteractiveMode {
		if err := cli.runValidationMode(); err != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Validation error: %v\n"), err)
			os.Exit(1)
		}
		return
	}

	// Traditional conversion mode with optional validation
	if err := cli.runIntegratedMode(); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error: %v\n"), err)
		os.Exit(1)
	}
}

// runTraditionalMode executes the original conversion logic
func runTraditionalMode() {
	var r io.Reader = os.Stdin
	if *inFile != "-" {
		f, e := os.Open(*inFile)
		if e != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error opening input file: %v\n"), e)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	eng := transform.NewDefaultEngine()
	var outLines []string
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()
		res := eng.Apply(line)
		if res.Changed {
			for _, c := range res.Changes {
				if *stats {
					fmt.Fprintf(os.Stderr, color.YellowString("#L%-5d %s => %s [%s]\n"), lineNum, c.Before, c.After, c.RuleName)
				}
			}
		}
		outLines = append(outLines, res.Line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error reading input: %v\n"), err)
		os.Exit(1)
	}

	output := strings.Join(append([]string{transform.GeneratedHeader()}, outLines...), "\n") + "\n"

	var w io.Writer = os.Stdout
	if *outFile != "-" {
		f, e := os.Create(*outFile)
		if e != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error creating output file: %v\n"), e)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	if _, err := io.WriteString(w, output); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error writing output: %v\n"), err)
		os.Exit(1)
	}
}

// runSandboxMode executes the new sandbox functionality
func runSandboxMode() {
	// Load configuration with new file-based system
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("Error loading configuration: %v\n"), err)
		os.Exit(1)
	}

	// Override config with command line flags
	cfg.Enabled = *sandboxMode
	cfg.DryRun = *dryRun
	cfg.Interactive = *interactive && !*batch

	// Validate configuration if sandbox is enabled
	if cfg.Enabled {
		if err := cfg.Validate(); err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Configuration validation failed:\n"))
			cfg.PrintGuide()
			os.Exit(1)
		}

		// Check if usacloud CLI is installed
		if !sandbox.IsUsacloudInstalled() {
			fmt.Fprint(os.Stderr, color.RedString("Error: usacloud CLI not found\n"))
			fmt.Fprintf(os.Stderr, "Please install usacloud CLI: https://docs.usacloud.jp/usacloud/installation/\n")
			os.Exit(1)
		}
	}

	// Handle input source
	var lines []string
	var inputSource string

	if *inFile != "-" {
		// Explicit file input
		f, err := os.Open(*inFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error opening input file: %v\n"), err)
			os.Exit(1)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error reading input: %v\n"), err)
			os.Exit(1)
		}
		inputSource = *inFile
	} else {
		// No explicit input file - check if stdin has data or use file selector
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, color.RedString("Error reading stdin: %v\n"), err)
				os.Exit(1)
			}
			inputSource = "<stdin>"
		} else {
			// No stdin data - use file selector
			selectedFiles, err := runFileSelector(cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, color.RedString("Error in file selection: %v\n"), err)
				os.Exit(1)
			}

			if len(selectedFiles) == 0 {
				fmt.Fprint(os.Stderr, color.YellowString("No files selected. Exiting.\n"))
				os.Exit(0)
			}

			// Process multiple files
			if len(selectedFiles) > 1 {
				runMultiFileMode(cfg, selectedFiles)
				return
			}

			// Single file selected
			f, err := os.Open(selectedFiles[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, color.RedString("Error opening selected file: %v\n"), err)
				os.Exit(1)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, color.RedString("Error reading selected file: %v\n"), err)
				os.Exit(1)
			}
			inputSource = selectedFiles[0]
		}
	}

	// Log input source for debugging
	if cfg.Debug {
		fmt.Fprintf(os.Stderr, color.CyanString("[DEBUG] Input source: %s\n"), inputSource)
	}

	// Handle different execution modes
	if cfg.Interactive && !*batch {
		// Interactive TUI mode
		runInteractiveMode(cfg, lines)
	} else {
		// Batch mode or non-interactive mode
		runBatchMode(cfg, lines)
	}
}

// runFileSelector shows the file selector TUI and returns selected files
func runFileSelector(cfg *config.SandboxConfig) ([]string, error) {
	var selectedFiles []string
	var selectorError error

	fileSelector := tui.NewFileSelector(cfg)

	fileSelector.SetOnFilesSelected(func(files []string) {
		selectedFiles = files
	})

	fileSelector.SetOnCancel(func() {
		selectedFiles = nil
	})

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Show informative message
	fmt.Fprint(os.Stderr, color.CyanString("🔍 Scanning current directory for script files...\n"))
	fmt.Fprintf(os.Stderr, "Directory: %s\n\n", currentDir)

	// Run file selector
	if err := fileSelector.Run(currentDir); err != nil {
		selectorError = err
	}

	return selectedFiles, selectorError
}

// runMultiFileMode processes multiple files sequentially
func runMultiFileMode(cfg *config.SandboxConfig, filePaths []string) {
	fmt.Fprintf(os.Stderr, "🔄 Processing %d files in batch mode...\n\n", len(filePaths))

	var allResults []*sandbox.ExecutionResult
	executor := sandbox.NewExecutor(cfg)

	for i, filePath := range filePaths {
		fmt.Fprintf(os.Stderr, color.BlueString("📄 Processing file %d/%d: %s\n"), i+1, len(filePaths), filePath)

		// Read file
		lines, err := readFileLines(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error reading file %s: %v\n"), filePath, err)
			continue
		}

		// Execute
		results, err := executor.ExecuteScript(lines)
		if err != nil {
			fmt.Fprintf(os.Stderr, color.RedString("Error executing file %s: %v\n"), filePath, err)
			continue
		}

		allResults = append(allResults, results...)

		// Print individual file summary
		succeeded := 0
		failed := 0
		skipped := 0

		for _, result := range results {
			if result.Skipped {
				skipped++
			} else if result.Success {
				succeeded++
			} else {
				failed++
			}
		}

		fmt.Fprintf(os.Stderr, "  ✅ %d successful, ❌ %d failed, ⏭️  %d skipped\n\n", succeeded, failed, skipped)
	}

	// Print overall summary
	if len(allResults) > 0 {
		fmt.Fprint(os.Stderr, color.HiWhiteString("📊 Overall Summary:\n"))
		executor.PrintSummary(allResults)
	}

	// Exit with error code if any commands failed
	for _, result := range allResults {
		if !result.Success && !result.Skipped {
			os.Exit(1)
		}
	}
}

// readFileLines reads a file and returns its lines
func readFileLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// runInteractiveMode runs the TUI for interactive command selection and execution
func runInteractiveMode(cfg *config.SandboxConfig, lines []string) {
	app := tui.NewApp(cfg)

	if err := app.LoadScript(lines); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error loading script: %v\n"), err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, color.CyanString("🚀 Starting interactive sandbox mode...\n"))
	fmt.Fprintf(os.Stderr, "Use arrow keys to navigate, Space to select, Enter to execute, 'q' to quit\n\n")

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error running TUI: %v\n"), err)
		os.Exit(1)
	}
}

// runBatchMode runs all commands automatically without user interaction
func runBatchMode(cfg *config.SandboxConfig, lines []string) {
	executor := sandbox.NewExecutor(cfg)

	fmt.Fprint(os.Stderr, color.CyanString("🔄 Starting batch sandbox execution...\n\n"))

	results, err := executor.ExecuteScript(lines)
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Error executing script: %v\n"), err)
		os.Exit(1)
	}

	// Print results to stdout (for potential piping/redirection)
	for _, result := range results {
		if !result.Skipped && result.Success && result.Output != "" {
			fmt.Println(result.Output)
		}
	}

	// Print summary to stderr
	executor.PrintSummary(results)

	// Exit with error code if any commands failed
	for _, result := range results {
		if !result.Success && !result.Skipped {
			os.Exit(1)
		}
	}
}
