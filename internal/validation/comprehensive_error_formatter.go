// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"fmt"
	"strings"
)

// ErrorContext represents error context information
type ErrorContext struct {
	InputCommand    string             // Original input command
	CommandParts    []string           // Parsed command parts
	DetectedIssues  []ValidationIssue  // Detected validation issues
	Suggestions     []SimilarityResult // Similar command suggestions
	DeprecationInfo *DeprecationInfo   // Deprecation information
	HelpURL         string             // Help URL
}

// ValidationIssue represents a validation issue found
type ValidationIssue struct {
	Type      IssueType       // Issue type
	Severity  MessageSeverity // Issue severity
	Component string          // Component with issue
	Message   string          // Issue description
	Expected  []string        // Expected values
}

// IssueType represents the type of validation issue
type IssueType int

const (
	IssueInvalidMainCommand IssueType = iota
	IssueInvalidSubCommand
	IssueDeprecatedCommand
	IssueSyntaxError
	IssueAmbiguousCommand
)

// UserIntent represents inferred user intent
type UserIntent int

const (
	IntentTypo      UserIntent = iota // User made a typo
	IntentExploring                   // User is exploring commands
	IntentMigrating                   // User is migrating from old version
	IntentLearning                    // User is learning the tool
)

// ErrorAnalysis represents the result of error context analysis
type ErrorAnalysis struct {
	PrimaryIssue      *ValidationIssue  // Most important issue
	SecondaryIssues   []ValidationIssue // Additional issues
	UserIntent        UserIntent        // Inferred user intent
	RecommendedAction string            // Recommended action
}

// VisualElements defines visual elements for formatted output
type VisualElements struct {
	ErrorIcon      string // ❌
	WarningIcon    string // ⚠️
	InfoIcon       string // ℹ️
	SuggestionIcon string // 💡
	SuccessIcon    string // ✅
	MigrationIcon  string // 🔄
	ListIcon       string // 📋
}

// Messages holds localized messages
type Messages struct {
	InvalidCommand     string
	InvalidSubcommand  string
	DeprecatedCommand  string
	SuggestionsHeader  string
	AlternativesHeader string
	MigrationHeader    string
	AvailableCommands  string
	SeeAlso            string
	MultipleIssues     string
	FixedExample       string
}

// ComprehensiveErrorFormatter provides comprehensive error formatting
type ComprehensiveErrorFormatter struct {
	messageGenerator   *ErrorMessageGenerator
	commandSuggester   *SimilarCommandSuggester
	deprecatedDetector *DeprecatedCommandDetector
	colorEnabled       bool
	language           string // "ja" or "en"
}

// NewComprehensiveErrorFormatter creates a new comprehensive error formatter
func NewComprehensiveErrorFormatter(
	msgGen *ErrorMessageGenerator,
	suggester *SimilarCommandSuggester,
	detector *DeprecatedCommandDetector,
	colorEnabled bool,
	language string,
) *ComprehensiveErrorFormatter {
	if language == "" {
		language = "ja" // Default to Japanese
	}

	return &ComprehensiveErrorFormatter{
		messageGenerator:   msgGen,
		commandSuggester:   suggester,
		deprecatedDetector: detector,
		colorEnabled:       colorEnabled,
		language:           language,
	}
}

// NewDefaultComprehensiveErrorFormatter creates a formatter with default settings
func NewDefaultComprehensiveErrorFormatter() *ComprehensiveErrorFormatter {
	return NewComprehensiveErrorFormatter(
		NewErrorMessageGenerator(true),
		NewDefaultSimilarCommandSuggester(),
		NewDeprecatedCommandDetector(),
		true,
		"ja",
	)
}

// FormatError formats a comprehensive error message
func (f *ComprehensiveErrorFormatter) FormatError(context *ErrorContext) string {
	if context == nil || len(context.DetectedIssues) == 0 {
		return f.formatUnknownError()
	}

	analysis := f.analyzeErrorContext(context)
	visual := f.getVisualElements()
	messages := f.getMessages()

	var sections []string

	// Header section
	headerSection := f.formatHeader(analysis, visual, messages)
	if headerSection != "" {
		sections = append(sections, headerSection)
	}

	// Problem description
	problemSection := f.formatProblemDescription(analysis, context, messages)
	if problemSection != "" {
		sections = append(sections, problemSection)
	}

	// Suggestions section
	suggestionSection := f.formatSuggestions(context, visual, messages)
	if suggestionSection != "" {
		sections = append(sections, suggestionSection)
	}

	// Migration/alternatives section
	migrationSection := f.formatMigrationInfo(context, visual, messages)
	if migrationSection != "" {
		sections = append(sections, migrationSection)
	}

	// Help section
	helpSection := f.formatHelpInfo(context, visual, messages)
	if helpSection != "" {
		sections = append(sections, helpSection)
	}

	result := strings.Join(sections, "\n\n")

	// Apply color if enabled
	if f.colorEnabled && analysis.PrimaryIssue != nil {
		result = f.applyColor(result, analysis.PrimaryIssue.Severity)
	}

	return result
}

// FormatValidationResult formats error from validation result
func (f *ComprehensiveErrorFormatter) FormatValidationResult(result *ValidationResult) string {
	if result.IsValid {
		return f.formatSuccess()
	}

	context := f.buildContextFromValidationResult(result)
	return f.FormatError(context)
}

// FormatSubcommandResult formats error from subcommand validation result
func (f *ComprehensiveErrorFormatter) FormatSubcommandResult(result *SubcommandValidationResult) string {
	if result.IsValid {
		// Return specific success message for subcommands
		visual := f.getVisualElements()
		if f.language == "en" {
			return fmt.Sprintf("%s Subcommand is valid", visual.SuccessIcon)
		}
		return fmt.Sprintf("%s サブコマンドは正常です", visual.SuccessIcon)
	}

	context := f.buildContextFromSubcommandResult(result)
	return f.FormatError(context)
}

// analyzeErrorContext analyzes the error context and returns analysis
func (f *ComprehensiveErrorFormatter) analyzeErrorContext(context *ErrorContext) *ErrorAnalysis {
	analysis := &ErrorAnalysis{
		SecondaryIssues: make([]ValidationIssue, 0),
	}

	// Find primary issue (highest severity)
	var primaryIssue *ValidationIssue
	for i, issue := range context.DetectedIssues {
		if primaryIssue == nil || issue.Severity < primaryIssue.Severity {
			if primaryIssue != nil {
				analysis.SecondaryIssues = append(analysis.SecondaryIssues, *primaryIssue)
			}
			primaryIssue = &context.DetectedIssues[i]
		} else {
			analysis.SecondaryIssues = append(analysis.SecondaryIssues, issue)
		}
	}

	analysis.PrimaryIssue = primaryIssue
	analysis.UserIntent = f.inferUserIntent(context)
	analysis.RecommendedAction = f.recommendAction(context)

	return analysis
}

// inferUserIntent infers the user's intent from context
func (f *ComprehensiveErrorFormatter) inferUserIntent(context *ErrorContext) UserIntent {
	// Check for deprecation (migration intent)
	if context.DeprecationInfo != nil {
		return IntentMigrating
	}

	// Check for close suggestions (typo intent)
	for _, suggestion := range context.Suggestions {
		if suggestion.Score > 0.8 {
			return IntentTypo
		}
	}

	// Check if exploring (short command, many options)
	if len(context.CommandParts) <= 2 && len(context.Suggestions) > 3 {
		return IntentExploring
	}

	return IntentLearning
}

// recommendAction recommends action based on context
func (f *ComprehensiveErrorFormatter) recommendAction(context *ErrorContext) string {
	if len(context.Suggestions) > 0 {
		topSuggestion := context.Suggestions[0]
		if len(context.CommandParts) > 1 {
			return fmt.Sprintf("usacloud %s", topSuggestion.Command)
		}
		return fmt.Sprintf("usacloud %s", topSuggestion.Command)
	}

	return "usacloud --help"
}

// formatHeader formats the error header
func (f *ComprehensiveErrorFormatter) formatHeader(analysis *ErrorAnalysis, visual *VisualElements, messages *Messages) string {
	if analysis.PrimaryIssue == nil {
		return ""
	}

	icon := visual.ErrorIcon
	switch analysis.PrimaryIssue.Severity {
	case SeverityWarning:
		icon = visual.WarningIcon
	case SeverityInfo:
		icon = visual.InfoIcon
	case SeveritySuccess:
		icon = visual.SuccessIcon
	}

	// Handle multiple issues
	if len(analysis.SecondaryIssues) > 0 {
		return fmt.Sprintf("%s %s", icon, messages.MultipleIssues)
	}

	// Generate localized message based on issue type
	message := f.generateLocalizedMessage(analysis.PrimaryIssue, messages)
	return fmt.Sprintf("%s %s", icon, message)
}

// formatProblemDescription formats the problem description
func (f *ComprehensiveErrorFormatter) formatProblemDescription(analysis *ErrorAnalysis, context *ErrorContext, messages *Messages) string {
	if len(analysis.SecondaryIssues) == 0 {
		return ""
	}

	var descriptions []string
	for _, issue := range analysis.SecondaryIssues {
		desc := fmt.Sprintf("   %s", issue.Message)
		descriptions = append(descriptions, desc)
	}

	return strings.Join(descriptions, "\n")
}

// formatSuggestions formats command suggestions
func (f *ComprehensiveErrorFormatter) formatSuggestions(context *ErrorContext, visual *VisualElements, messages *Messages) string {
	if len(context.Suggestions) == 0 {
		return ""
	}

	var sections []string

	// Suggestions header
	header := fmt.Sprintf("%s %s", visual.SuggestionIcon, messages.SuggestionsHeader)
	sections = append(sections, header)

	// List suggestions with scores
	for i, suggestion := range context.Suggestions {
		if i >= 3 { // Limit to top 3 suggestions
			break
		}
		scorePercent := int(suggestion.Score * 100)
		suggestionLine := fmt.Sprintf("   • %s (%s: %d%%)",
			suggestion.Command,
			f.getScoreLabel(),
			scorePercent)
		sections = append(sections, suggestionLine)
	}

	return strings.Join(sections, "\n")
}

// formatMigrationInfo formats migration/deprecation information
func (f *ComprehensiveErrorFormatter) formatMigrationInfo(context *ErrorContext, visual *VisualElements, messages *Messages) string {
	if context.DeprecationInfo == nil {
		return ""
	}

	var sections []string

	// Migration header
	if context.DeprecationInfo.ReplacementCommand != "" {
		header := fmt.Sprintf("%s %s", visual.MigrationIcon, messages.AlternativesHeader)
		sections = append(sections, header)

		replacement := fmt.Sprintf("   usacloud %s", context.DeprecationInfo.ReplacementCommand)
		sections = append(sections, replacement)
	}

	// Migration guide
	if len(context.DeprecationInfo.AlternativeActions) > 0 {
		guideHeader := fmt.Sprintf("%s %s", visual.ListIcon, messages.MigrationHeader)
		sections = append(sections, guideHeader)

		for _, action := range context.DeprecationInfo.AlternativeActions {
			sections = append(sections, fmt.Sprintf("   • %s", action))
		}
	}

	return strings.Join(sections, "\n")
}

// formatHelpInfo formats help information
func (f *ComprehensiveErrorFormatter) formatHelpInfo(context *ErrorContext, visual *VisualElements, messages *Messages) string {
	helpInfo := f.generateDynamicHelp(context)
	if helpInfo == "" {
		return ""
	}

	return fmt.Sprintf("%s %s", visual.InfoIcon, fmt.Sprintf(messages.SeeAlso, helpInfo))
}

// generateDynamicHelp generates dynamic help information
func (f *ComprehensiveErrorFormatter) generateDynamicHelp(context *ErrorContext) string {
	var helpItems []string

	// Basic help command
	if len(context.CommandParts) <= 1 {
		helpItems = append(helpItems, "usacloud --help")
	} else if len(context.CommandParts) >= 2 {
		helpItems = append(helpItems, fmt.Sprintf("usacloud %s --help", context.CommandParts[0]))
	}

	// Documentation URL
	if context.HelpURL != "" {
		helpItems = append(helpItems, context.HelpURL)
	} else {
		// Default documentation
		helpItems = append(helpItems, "https://docs.usacloud.jp/usacloud/")
	}

	return strings.Join(helpItems, "\n   ")
}

// buildContextFromValidationResult builds context from validation result
func (f *ComprehensiveErrorFormatter) buildContextFromValidationResult(result *ValidationResult) *ErrorContext {
	context := &ErrorContext{
		InputCommand:   result.Command,
		CommandParts:   []string{result.Command},
		DetectedIssues: make([]ValidationIssue, 0),
	}

	// Add validation issue
	severity := SeverityError
	if result.ErrorType == "deprecated_command" || result.ErrorType == "discontinued_command" {
		severity = SeverityWarning
	}

	issue := ValidationIssue{
		Type:      f.mapErrorTypeToIssueType(result.ErrorType),
		Severity:  severity,
		Component: result.Command,
		Message:   f.messageGenerator.GenerateFromValidationResult(result),
	}
	context.DetectedIssues = append(context.DetectedIssues, issue)

	// Add suggestions
	if len(result.Suggestions) > 0 {
		suggestions := f.commandSuggester.SuggestMainCommands(result.Command)
		context.Suggestions = suggestions
	}

	// Add deprecation info
	if f.deprecatedDetector.IsDeprecated(result.Command) {
		context.DeprecationInfo = f.deprecatedDetector.Detect(result.Command)
	}

	return context
}

// buildContextFromSubcommandResult builds context from subcommand validation result
func (f *ComprehensiveErrorFormatter) buildContextFromSubcommandResult(result *SubcommandValidationResult) *ErrorContext {
	context := &ErrorContext{
		InputCommand:   fmt.Sprintf("%s %s", result.MainCommand, result.SubCommand),
		CommandParts:   []string{result.MainCommand, result.SubCommand},
		DetectedIssues: make([]ValidationIssue, 0),
	}

	// Add validation issue
	issue := ValidationIssue{
		Type:      IssueInvalidSubCommand,
		Severity:  SeverityError,
		Component: result.SubCommand,
		Message:   f.messageGenerator.GenerateFromSubcommandResult(result),
		Expected:  result.Available,
	}
	context.DetectedIssues = append(context.DetectedIssues, issue)

	// Add suggestions
	if len(result.Suggestions) > 0 {
		suggestions := f.commandSuggester.SuggestSubcommands(result.MainCommand, result.SubCommand)
		context.Suggestions = suggestions
	}

	return context
}

// mapErrorTypeToIssueType maps error type string to IssueType
func (f *ComprehensiveErrorFormatter) mapErrorTypeToIssueType(errorType string) IssueType {
	switch errorType {
	case "unknown_command", "empty_command":
		return IssueInvalidMainCommand
	case "deprecated_command", "discontinued_command":
		return IssueDeprecatedCommand
	case "unexpected_subcommand":
		return IssueSyntaxError
	default:
		return IssueInvalidMainCommand
	}
}

// getVisualElements returns visual elements (icons)
func (f *ComprehensiveErrorFormatter) getVisualElements() *VisualElements {
	return &VisualElements{
		ErrorIcon:      "❌",
		WarningIcon:    "⚠️",
		InfoIcon:       "ℹ️",
		SuggestionIcon: "💡",
		SuccessIcon:    "✅",
		MigrationIcon:  "🔄",
		ListIcon:       "📋",
	}
}

// getMessages returns localized messages
func (f *ComprehensiveErrorFormatter) getMessages() *Messages {
	if f.language == "en" {
		return &Messages{
			InvalidCommand:     "Error: '%s' is not a valid usacloud command",
			InvalidSubcommand:  "Error: '%s' is not a valid subcommand",
			DeprecatedCommand:  "Warning: '%s' command was deprecated in v1",
			SuggestionsHeader:  "Did you mean one of these?",
			AlternativesHeader: "Use this instead:",
			MigrationHeader:    "Migration guide:",
			AvailableCommands:  "Available commands for %s:",
			SeeAlso:            "See also: %s",
			MultipleIssues:     "Multiple issues detected:",
			FixedExample:       "Fixed example:",
		}
	}

	// Default Japanese
	return &Messages{
		InvalidCommand:     "エラー: '%s' は有効なusacloudコマンドではありません",
		InvalidSubcommand:  "エラー: '%s' は有効なサブコマンドではありません",
		DeprecatedCommand:  "注意: '%s' コマンドはv1で廃止されました",
		SuggestionsHeader:  "もしかして以下のコマンドですか？",
		AlternativesHeader: "代わりに以下を使用してください:",
		MigrationHeader:    "移行方法:",
		AvailableCommands:  "%s で利用可能なサブコマンド:",
		SeeAlso:            "詳細情報: %s",
		MultipleIssues:     "複数の問題が検出されました:",
		FixedExample:       "修正例:",
	}
}

// getScoreLabel returns score label in appropriate language
func (f *ComprehensiveErrorFormatter) getScoreLabel() string {
	if f.language == "en" {
		return "similarity"
	}
	return "類似度"
}

// formatSuccess formats success message
func (f *ComprehensiveErrorFormatter) formatSuccess() string {
	visual := f.getVisualElements()
	if f.language == "en" {
		return fmt.Sprintf("%s Command is valid", visual.SuccessIcon)
	}
	return fmt.Sprintf("%s コマンドは正常です", visual.SuccessIcon)
}

// formatUnknownError formats unknown error
func (f *ComprehensiveErrorFormatter) formatUnknownError() string {
	visual := f.getVisualElements()
	if f.language == "en" {
		return fmt.Sprintf("%s An unknown error occurred", visual.ErrorIcon)
	}
	return fmt.Sprintf("%s 不明なエラーが発生しました", visual.ErrorIcon)
}

// applyColor applies color to the message based on severity
func (f *ComprehensiveErrorFormatter) applyColor(message string, severity MessageSeverity) string {
	if !f.colorEnabled {
		return message
	}

	var color ColorCode
	switch severity {
	case SeverityError:
		color = ColorRed
	case SeverityWarning:
		color = ColorYellow
	case SeverityInfo:
		color = ColorBlue
	case SeveritySuccess:
		color = ColorGreen
	default:
		return message
	}

	return fmt.Sprintf("%s%s%s", color, message, ColorReset)
}

// SetColorEnabled enables or disables color output
func (f *ComprehensiveErrorFormatter) SetColorEnabled(enabled bool) {
	f.colorEnabled = enabled
	if f.messageGenerator != nil {
		f.messageGenerator.SetColorEnabled(enabled)
	}
}

// IsColorEnabled returns whether color output is enabled
func (f *ComprehensiveErrorFormatter) IsColorEnabled() bool {
	return f.colorEnabled
}

// SetLanguage sets the output language
func (f *ComprehensiveErrorFormatter) SetLanguage(language string) {
	if language == "en" || language == "ja" {
		f.language = language
	}
}

// GetLanguage returns the current output language
func (f *ComprehensiveErrorFormatter) GetLanguage() string {
	return f.language
}

// GetIssueTypeString returns the issue type as a string
func GetIssueTypeString(issueType IssueType) string {
	switch issueType {
	case IssueInvalidMainCommand:
		return "InvalidMainCommand"
	case IssueInvalidSubCommand:
		return "InvalidSubCommand"
	case IssueDeprecatedCommand:
		return "DeprecatedCommand"
	case IssueSyntaxError:
		return "SyntaxError"
	case IssueAmbiguousCommand:
		return "AmbiguousCommand"
	default:
		return "Unknown"
	}
}

// generateLocalizedMessage generates a localized message based on issue type
func (f *ComprehensiveErrorFormatter) generateLocalizedMessage(issue *ValidationIssue, messages *Messages) string {
	switch issue.Type {
	case IssueInvalidMainCommand:
		return fmt.Sprintf(messages.InvalidCommand, issue.Component)
	case IssueInvalidSubCommand:
		return fmt.Sprintf(messages.InvalidSubcommand, issue.Component)
	case IssueDeprecatedCommand:
		// For deprecated commands, use the detailed message that includes replacement info
		return issue.Message
	default:
		return issue.Message
	}
}

// GetUserIntentString returns the user intent as a string
func GetUserIntentString(intent UserIntent) string {
	switch intent {
	case IntentTypo:
		return "Typo"
	case IntentExploring:
		return "Exploring"
	case IntentMigrating:
		return "Migrating"
	case IntentLearning:
		return "Learning"
	default:
		return "Unknown"
	}
}
