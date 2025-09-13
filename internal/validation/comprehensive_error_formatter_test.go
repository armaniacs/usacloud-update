package validation

import (
	"strings"
	"testing"
)

func TestNewComprehensiveErrorFormatter(t *testing.T) {
	msgGen := NewErrorMessageGenerator(true)
	suggester := NewDefaultSimilarCommandSuggester()
	detector := NewDeprecatedCommandDetector()

	formatter := NewComprehensiveErrorFormatter(msgGen, suggester, detector, true, "ja")
	if formatter == nil {
		t.Error("Expected formatter to be created, got nil")
	}
	if !formatter.IsColorEnabled() {
		t.Error("Expected color to be enabled")
	}
	if formatter.GetLanguage() != "ja" {
		t.Errorf("Expected language to be 'ja', got '%s'", formatter.GetLanguage())
	}
}

func TestNewDefaultComprehensiveErrorFormatter(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	if formatter == nil {
		t.Error("Expected formatter to be created, got nil")
	}
	if !formatter.IsColorEnabled() {
		t.Error("Expected color to be enabled by default")
	}
	if formatter.GetLanguage() != "ja" {
		t.Error("Expected default language to be Japanese")
	}
}

func TestFormatValidationResult(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	formatter.SetColorEnabled(false) // Disable for easier testing

	tests := []struct {
		name          string
		result        *ValidationResult
		expectContain []string
	}{
		{
			name: "Valid command",
			result: &ValidationResult{
				IsValid: true,
				Command: "server",
			},
			expectContain: []string{"✅", "コマンドは正常です"},
		},
		{
			name: "Invalid command with suggestions",
			result: &ValidationResult{
				IsValid:     false,
				Command:     "sever",
				ErrorType:   "unknown_command",
				Suggestions: []string{"server"},
			},
			expectContain: []string{"❌", "sever", "有効な", "💡"},
		},
		{
			name: "Deprecated command",
			result: &ValidationResult{
				IsValid:     false,
				Command:     "iso-image",
				ErrorType:   "deprecated_command",
				Suggestions: []string{"cdrom"},
			},
			expectContain: []string{"⚠️", "iso-image", "廃止", "🔄", "cdrom"},
		},
		{
			name: "Empty command",
			result: &ValidationResult{
				IsValid:   false,
				Command:   "",
				ErrorType: "empty_command",
			},
			expectContain: []string{"❌", "エラー", "有効なusacloudコマンドではありません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatValidationResult(tt.result)

			for _, expected := range tt.expectContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestFormatSubcommandResult(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	formatter.SetColorEnabled(false) // Disable for easier testing

	tests := []struct {
		name          string
		result        *SubcommandValidationResult
		expectContain []string
	}{
		{
			name: "Valid subcommand",
			result: &SubcommandValidationResult{
				IsValid:     true,
				MainCommand: "server",
				SubCommand:  "list",
			},
			expectContain: []string{"✅", "サブコマンドは正常です"},
		},
		{
			name: "Invalid subcommand with suggestions",
			result: &SubcommandValidationResult{
				IsValid:     false,
				MainCommand: "server",
				SubCommand:  "lst",
				ErrorType:   ErrorTypeInvalidSubcommand,
				Available:   []string{"list", "create", "delete"},
				Suggestions: []string{"list"},
			},
			expectContain: []string{"❌", "lst", "server", "有効な", "💡"},
		},
		{
			name: "Missing subcommand",
			result: &SubcommandValidationResult{
				IsValid:     false,
				MainCommand: "server",
				SubCommand:  "",
				ErrorType:   ErrorTypeMissingSubcommand,
				Available:   []string{"list", "create", "delete"},
			},
			expectContain: []string{"❌", "エラー", "有効なサブコマンドではありません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatSubcommandResult(tt.result)

			for _, expected := range tt.expectContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	formatter.SetColorEnabled(false)

	// Test with comprehensive error context
	context := &ErrorContext{
		InputCommand: "sever lst",
		CommandParts: []string{"sever", "lst"},
		DetectedIssues: []ValidationIssue{
			{
				Type:      IssueInvalidMainCommand,
				Severity:  SeverityError,
				Component: "sever",
				Message:   "エラー: 'sever' は有効なusacloudコマンドではありません",
			},
		},
		Suggestions: []SimilarityResult{
			{
				Command:  "server",
				Distance: 1,
				Score:    0.83,
			},
		},
		HelpURL: "https://docs.usacloud.jp/",
	}

	result := formatter.FormatError(context)

	expectedElements := []string{
		"❌",      // Error icon
		"sever",  // Invalid command
		"💡",      // Suggestion icon
		"server", // Suggested command
		"83%",    // Similarity score
		"ℹ️",     // Info icon
		"詳細情報",   // Help info
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	}
}

func TestFormatErrorWithDeprecation(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	formatter.SetColorEnabled(false)

	deprecationInfo := &DeprecationInfo{
		Command:            "iso-image",
		ReplacementCommand: "cdrom",
		DeprecationType:    "renamed",
		Message:            "iso-imageコマンドはv1で廃止されました",
		AlternativeActions: []string{"cdrom list を使用してください"},
	}

	context := &ErrorContext{
		InputCommand: "iso-image",
		CommandParts: []string{"iso-image"},
		DetectedIssues: []ValidationIssue{
			{
				Type:      IssueDeprecatedCommand,
				Severity:  SeverityWarning,
				Component: "iso-image",
				Message:   "注意: 'iso-image' コマンドはv1で廃止されました",
			},
		},
		DeprecationInfo: deprecationInfo,
		HelpURL:         "https://docs.usacloud.jp/upgrade/",
	}

	result := formatter.FormatError(context)

	expectedElements := []string{
		"⚠️",        // Warning icon
		"iso-image", // Deprecated command
		"🔄",         // Migration icon
		"cdrom",     // Replacement command
		"📋",         // List icon
		"移行方法",      // Migration guide
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	}
}

func TestLanguageSupport(t *testing.T) {
	tests := []struct {
		language      string
		expectContain []string
	}{
		{
			language:      "ja",
			expectContain: []string{"エラー", "もしかして", "詳細情報"},
		},
		{
			language:      "en",
			expectContain: []string{"Error", "Did you mean", "See also"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			formatter := NewDefaultComprehensiveErrorFormatter()
			formatter.SetLanguage(tt.language)
			formatter.SetColorEnabled(false)

			context := &ErrorContext{
				InputCommand: "invalidcommand",
				CommandParts: []string{"invalidcommand"},
				DetectedIssues: []ValidationIssue{
					{
						Type:      IssueInvalidMainCommand,
						Severity:  SeverityError,
						Component: "invalidcommand",
						Message:   "Invalid command message",
					},
				},
				Suggestions: []SimilarityResult{
					{
						Command: "server",
						Score:   0.7,
					},
				},
				HelpURL: "https://docs.usacloud.jp/",
			}

			result := formatter.FormatError(context)

			for _, expected := range tt.expectContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result in %s to contain '%s', got: %s",
						tt.language, expected, result)
				}
			}
		})
	}
}

func TestAnalyzeErrorContext(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	context := &ErrorContext{
		InputCommand: "sever lst",
		CommandParts: []string{"sever", "lst"},
		DetectedIssues: []ValidationIssue{
			{
				Type:     IssueInvalidMainCommand,
				Severity: SeverityError,
				Message:  "Invalid main command",
			},
			{
				Type:     IssueInvalidSubCommand,
				Severity: SeverityWarning,
				Message:  "Invalid subcommand",
			},
		},
		Suggestions: []SimilarityResult{
			{Command: "server", Score: 0.85},
		},
	}

	analysis := formatter.analyzeErrorContext(context)

	if analysis.PrimaryIssue == nil {
		t.Error("Expected primary issue to be identified")
	}

	if analysis.PrimaryIssue.Type != IssueInvalidMainCommand {
		t.Error("Expected primary issue to be invalid main command (highest severity)")
	}

	if len(analysis.SecondaryIssues) != 1 {
		t.Errorf("Expected 1 secondary issue, got %d", len(analysis.SecondaryIssues))
	}

	if analysis.UserIntent != IntentTypo {
		t.Errorf("Expected user intent to be typo (high score suggestion), got %s",
			GetUserIntentString(analysis.UserIntent))
	}
}

func TestInferUserIntent(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	tests := []struct {
		name           string
		context        *ErrorContext
		expectedIntent UserIntent
	}{
		{
			name: "Migration intent (deprecated command)",
			context: &ErrorContext{
				DeprecationInfo: &DeprecationInfo{
					Command: "iso-image",
				},
			},
			expectedIntent: IntentMigrating,
		},
		{
			name: "Typo intent (high score suggestion)",
			context: &ErrorContext{
				Suggestions: []SimilarityResult{
					{Command: "server", Score: 0.9},
				},
			},
			expectedIntent: IntentTypo,
		},
		{
			name: "Exploring intent (short command, many suggestions)",
			context: &ErrorContext{
				CommandParts: []string{"se"},
				Suggestions: []SimilarityResult{
					{Command: "server", Score: 0.6},
					{Command: "service", Score: 0.5},
					{Command: "self", Score: 0.4},
					{Command: "session", Score: 0.3},
				},
			},
			expectedIntent: IntentExploring,
		},
		{
			name: "Learning intent (default)",
			context: &ErrorContext{
				CommandParts: []string{"somecommand"},
				Suggestions: []SimilarityResult{
					{Command: "server", Score: 0.3},
				},
			},
			expectedIntent: IntentLearning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := formatter.inferUserIntent(tt.context)
			if intent != tt.expectedIntent {
				t.Errorf("Expected intent %s, got %s",
					GetUserIntentString(tt.expectedIntent),
					GetUserIntentString(intent))
			}
		})
	}
}

func TestColorHandling(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	// Test color enabled
	formatter.SetColorEnabled(true)
	if !formatter.IsColorEnabled() {
		t.Error("Expected color to be enabled")
	}

	// Test color disabled
	formatter.SetColorEnabled(false)
	if formatter.IsColorEnabled() {
		t.Error("Expected color to be disabled")
	}

	// Test color application
	message := "Test message"
	coloredMessage := formatter.applyColor(message, SeverityError)

	// When color is disabled, should return original message
	if coloredMessage != message {
		t.Error("Expected original message when color is disabled")
	}

	// When color is enabled, should add color codes
	formatter.SetColorEnabled(true)
	coloredMessage = formatter.applyColor(message, SeverityError)
	if !strings.Contains(coloredMessage, string(ColorRed)) {
		t.Error("Expected colored message to contain red color code")
	}
}

func TestGetVisualElements(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	visual := formatter.getVisualElements()

	expectedIcons := map[string]string{
		"ErrorIcon":      "❌",
		"WarningIcon":    "⚠️",
		"InfoIcon":       "ℹ️",
		"SuggestionIcon": "💡",
		"SuccessIcon":    "✅",
		"MigrationIcon":  "🔄",
		"ListIcon":       "📋",
	}

	if visual.ErrorIcon != expectedIcons["ErrorIcon"] {
		t.Errorf("Expected ErrorIcon to be %s, got %s", expectedIcons["ErrorIcon"], visual.ErrorIcon)
	}
	if visual.WarningIcon != expectedIcons["WarningIcon"] {
		t.Errorf("Expected WarningIcon to be %s, got %s", expectedIcons["WarningIcon"], visual.WarningIcon)
	}
	if visual.InfoIcon != expectedIcons["InfoIcon"] {
		t.Errorf("Expected InfoIcon to be %s, got %s", expectedIcons["InfoIcon"], visual.InfoIcon)
	}
}

func TestGetMessages(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	// Test Japanese messages
	formatter.SetLanguage("ja")
	messages := formatter.getMessages()
	if !strings.Contains(messages.InvalidCommand, "エラー") {
		t.Error("Expected Japanese error message to contain 'エラー'")
	}

	// Test English messages
	formatter.SetLanguage("en")
	messages = formatter.getMessages()
	if !strings.Contains(messages.InvalidCommand, "Error") {
		t.Error("Expected English error message to contain 'Error'")
	}
}

func TestBuildContextFromValidationResult(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	result := &ValidationResult{
		IsValid:     false,
		Command:     "sever",
		ErrorType:   "unknown_command",
		Suggestions: []string{"server"},
	}

	context := formatter.buildContextFromValidationResult(result)

	if context.InputCommand != "sever" {
		t.Errorf("Expected InputCommand to be 'sever', got '%s'", context.InputCommand)
	}

	if len(context.CommandParts) != 1 || context.CommandParts[0] != "sever" {
		t.Errorf("Expected CommandParts to be ['sever'], got %v", context.CommandParts)
	}

	if len(context.DetectedIssues) != 1 {
		t.Errorf("Expected 1 detected issue, got %d", len(context.DetectedIssues))
	}

	if context.DetectedIssues[0].Type != IssueInvalidMainCommand {
		t.Errorf("Expected issue type to be IssueInvalidMainCommand, got %s",
			GetIssueTypeString(context.DetectedIssues[0].Type))
	}
}

func TestBuildContextFromSubcommandResult(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	result := &SubcommandValidationResult{
		IsValid:     false,
		MainCommand: "server",
		SubCommand:  "lst",
		ErrorType:   ErrorTypeInvalidSubcommand,
		Available:   []string{"list", "create", "delete"},
		Suggestions: []string{"list"},
	}

	context := formatter.buildContextFromSubcommandResult(result)

	if context.InputCommand != "server lst" {
		t.Errorf("Expected InputCommand to be 'server lst', got '%s'", context.InputCommand)
	}

	if len(context.CommandParts) != 2 {
		t.Errorf("Expected 2 command parts, got %d", len(context.CommandParts))
	}

	if context.DetectedIssues[0].Type != IssueInvalidSubCommand {
		t.Errorf("Expected issue type to be IssueInvalidSubCommand, got %s",
			GetIssueTypeString(context.DetectedIssues[0].Type))
	}
}

func TestMapErrorTypeToIssueType(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	tests := []struct {
		errorType    string
		expectedType IssueType
	}{
		{"unknown_command", IssueInvalidMainCommand},
		{"empty_command", IssueInvalidMainCommand},
		{"deprecated_command", IssueDeprecatedCommand},
		{"discontinued_command", IssueDeprecatedCommand},
		{"unexpected_subcommand", IssueSyntaxError},
		{"unknown_type", IssueInvalidMainCommand}, // Default case
	}

	for _, tt := range tests {
		result := formatter.mapErrorTypeToIssueType(tt.errorType)
		if result != tt.expectedType {
			t.Errorf("mapErrorTypeToIssueType(%s) = %s, expected %s",
				tt.errorType, GetIssueTypeString(result), GetIssueTypeString(tt.expectedType))
		}
	}
}

func TestGetIssueTypeString(t *testing.T) {
	tests := []struct {
		issueType IssueType
		expected  string
	}{
		{IssueInvalidMainCommand, "InvalidMainCommand"},
		{IssueInvalidSubCommand, "InvalidSubCommand"},
		{IssueDeprecatedCommand, "DeprecatedCommand"},
		{IssueSyntaxError, "SyntaxError"},
		{IssueAmbiguousCommand, "AmbiguousCommand"},
	}

	for _, tt := range tests {
		result := GetIssueTypeString(tt.issueType)
		if result != tt.expected {
			t.Errorf("GetIssueTypeString(%v) = %s, expected %s", tt.issueType, result, tt.expected)
		}
	}
}

func TestGetUserIntentString(t *testing.T) {
	tests := []struct {
		intent   UserIntent
		expected string
	}{
		{IntentTypo, "Typo"},
		{IntentExploring, "Exploring"},
		{IntentMigrating, "Migrating"},
		{IntentLearning, "Learning"},
	}

	for _, tt := range tests {
		result := GetUserIntentString(tt.intent)
		if result != tt.expected {
			t.Errorf("GetUserIntentString(%v) = %s, expected %s", tt.intent, result, tt.expected)
		}
	}
}

func TestFormatterEdgeCases(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()
	formatter.SetColorEnabled(false)

	// Test nil context
	result := formatter.FormatError(nil)
	if !strings.Contains(result, "不明なエラー") {
		t.Error("Expected unknown error message for nil context")
	}

	// Test empty issues
	context := &ErrorContext{
		InputCommand:   "test",
		DetectedIssues: []ValidationIssue{},
	}
	result = formatter.FormatError(context)
	if !strings.Contains(result, "不明なエラー") {
		t.Error("Expected unknown error message for empty issues")
	}

	// Test invalid language (should default to Japanese)
	formatter.SetLanguage("invalid")
	if formatter.GetLanguage() != "ja" {
		t.Error("Expected language to remain 'ja' for invalid language")
	}
}

func TestGenerateDynamicHelp(t *testing.T) {
	formatter := NewDefaultComprehensiveErrorFormatter()

	tests := []struct {
		name          string
		context       *ErrorContext
		expectContain []string
	}{
		{
			name: "Single command part",
			context: &ErrorContext{
				CommandParts: []string{"test"},
			},
			expectContain: []string{"usacloud --help", "docs.usacloud.jp"},
		},
		{
			name: "Multiple command parts",
			context: &ErrorContext{
				CommandParts: []string{"server", "list"},
			},
			expectContain: []string{"usacloud server --help", "docs.usacloud.jp"},
		},
		{
			name: "With custom help URL",
			context: &ErrorContext{
				CommandParts: []string{"test"},
				HelpURL:      "https://custom.help.url/",
			},
			expectContain: []string{"usacloud --help", "https://custom.help.url/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := formatter.generateDynamicHelp(tt.context)

			for _, expected := range tt.expectContain {
				if !strings.Contains(help, expected) {
					t.Errorf("Expected help to contain '%s', got: %s", expected, help)
				}
			}
		})
	}
}
