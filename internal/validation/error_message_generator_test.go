package validation

import (
	"strings"
	"testing"
)

func TestNewErrorMessageGenerator(t *testing.T) {
	// Test with color enabled
	generator := NewErrorMessageGenerator(true)
	if generator == nil {
		t.Error("Expected generator to be created, got nil")
	}
	if !generator.IsColorEnabled() {
		t.Error("Expected color to be enabled")
	}

	// Test with color disabled
	generator = NewErrorMessageGenerator(false)
	if generator.IsColorEnabled() {
		t.Error("Expected color to be disabled")
	}

	// Check that templates were initialized
	templates := generator.GetAllTemplates()
	if len(templates) == 0 {
		t.Error("Expected templates to be initialized")
	}
}

func TestInitializeTemplates(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	expectedTypes := []MessageType{
		TypeInvalidCommand,
		TypeInvalidSubcommand,
		TypeDeprecatedCommand,
		TypeDiscontinuedCommand,
		TypeSyntaxError,
		TypeMissingCommand,
		TypeMissingSubcommand,
		TypeSuggestion,
		TypeSuccess,
	}

	for _, msgType := range expectedTypes {
		template := generator.GetTemplate(msgType)
		if template == nil {
			t.Errorf("Expected template for type %s to be initialized", GetMessageTypeString(msgType))
		}
		if template.Template == "" {
			t.Errorf("Expected non-empty template for type %s", GetMessageTypeString(msgType))
		}
	}
}

func TestGenerateMessageInvalidCommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command": "invalidcommand",
	}

	message := generator.GenerateMessage(TypeInvalidCommand, params)

	if !strings.Contains(message, "invalidcommand") {
		t.Error("Message should contain the invalid command name")
	}
	if !strings.Contains(message, "エラー") {
		t.Error("Message should contain error indicator")
	}
	if !strings.Contains(message, "usacloud --help") {
		t.Error("Message should contain help suggestion")
	}
}

func TestGenerateMessageInvalidSubcommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command":              "invalidaction",
		"mainCommand":          "server",
		"availableSubcommands": []string{"list", "create", "boot", "shutdown"},
	}

	message := generator.GenerateMessage(TypeInvalidSubcommand, params)

	if !strings.Contains(message, "invalidaction") {
		t.Error("Message should contain the invalid subcommand name")
	}
	if !strings.Contains(message, "server") {
		t.Error("Message should contain the main command name")
	}
	if !strings.Contains(message, "list, create, boot, shutdown") {
		t.Error("Message should contain available subcommands")
	}
}

func TestGenerateMessageDeprecatedCommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command":            "iso-image",
		"replacementCommand": "cdrom",
	}

	message := generator.GenerateMessage(TypeDeprecatedCommand, params)

	if !strings.Contains(message, "iso-image") {
		t.Error("Message should contain the deprecated command name")
	}
	if !strings.Contains(message, "cdrom") {
		t.Error("Message should contain the replacement command name")
	}
	if !strings.Contains(message, "注意") {
		t.Error("Message should contain warning indicator")
	}
	if !strings.Contains(message, "廃止") {
		t.Error("Message should mention deprecation")
	}
}

func TestGenerateMessageDiscontinuedCommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	alternativeActions := "  • Use bill command\n  • Use self command\n"
	params := map[string]interface{}{
		"command":            "summary",
		"alternativeActions": alternativeActions,
	}

	message := generator.GenerateMessage(TypeDiscontinuedCommand, params)

	if !strings.Contains(message, "summary") {
		t.Error("Message should contain the discontinued command name")
	}
	if !strings.Contains(message, "代替手段") {
		t.Error("Message should mention alternatives")
	}
	if !strings.Contains(message, "bill command") {
		t.Error("Message should contain alternative actions")
	}
}

func TestGenerateMessageSyntaxError(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command": "version",
	}

	message := generator.GenerateMessage(TypeSyntaxError, params)

	if !strings.Contains(message, "version") {
		t.Error("Message should contain the command name")
	}
	if !strings.Contains(message, "サブコマンドを受け付けません") {
		t.Error("Message should mention subcommand restriction")
	}
	if !strings.Contains(message, "正しい使用法") {
		t.Error("Message should provide correct usage")
	}
}

func TestGenerateMessageMissingCommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{}
	message := generator.GenerateMessage(TypeMissingCommand, params)

	if !strings.Contains(message, "メインコマンドが指定されていません") {
		t.Error("Message should mention missing main command")
	}
	if !strings.Contains(message, "使用法") {
		t.Error("Message should provide usage information")
	}
}

func TestGenerateMessageMissingSubcommand(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command":              "server",
		"availableSubcommands": []string{"list", "create", "boot"},
	}

	message := generator.GenerateMessage(TypeMissingSubcommand, params)

	if !strings.Contains(message, "server") {
		t.Error("Message should contain the main command")
	}
	if !strings.Contains(message, "サブコマンドが必要") {
		t.Error("Message should mention missing subcommand")
	}
	if !strings.Contains(message, "list, create, boot") {
		t.Error("Message should list available subcommands")
	}
}

func TestGenerateMessageWithSuggestions(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	params := map[string]interface{}{
		"command":     "invalidcommand",
		"suggestions": []string{"server", "service", "status"},
	}

	message := generator.GenerateMessage(TypeInvalidCommand, params)

	if !strings.Contains(message, "もしかして") {
		t.Error("Message should contain suggestion intro")
	}
	if !strings.Contains(message, "server") {
		t.Error("Message should contain first suggestion")
	}
	if !strings.Contains(message, "•") {
		t.Error("Message should use bullet points for suggestions")
	}
}

func TestColorizeMessage(t *testing.T) {
	// Test with color enabled
	generator := NewErrorMessageGenerator(true)
	message := "Test message"

	colorizedError := generator.colorizeMessage(message, SeverityError)
	if !strings.Contains(colorizedError, string(ColorRed)) {
		t.Error("Error message should be colored red")
	}

	colorizedWarning := generator.colorizeMessage(message, SeverityWarning)
	if !strings.Contains(colorizedWarning, string(ColorYellow)) {
		t.Error("Warning message should be colored yellow")
	}

	colorizedInfo := generator.colorizeMessage(message, SeverityInfo)
	if !strings.Contains(colorizedInfo, string(ColorBlue)) {
		t.Error("Info message should be colored blue")
	}

	colorizedSuccess := generator.colorizeMessage(message, SeveritySuccess)
	if !strings.Contains(colorizedSuccess, string(ColorGreen)) {
		t.Error("Success message should be colored green")
	}

	// Test with color disabled
	generator.SetColorEnabled(false)
	plain := generator.colorizeMessage(message, SeverityError)
	if strings.Contains(plain, string(ColorRed)) {
		t.Error("Message should not be colored when color is disabled")
	}
}

func TestGenerateFromValidationResult(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

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
			name: "Invalid command",
			result: &ValidationResult{
				IsValid:   false,
				Command:   "invalid",
				ErrorType: "unknown_command",
			},
			expectContain: []string{"エラー", "invalid", "有効な"},
		},
		{
			name: "Empty command",
			result: &ValidationResult{
				IsValid:   false,
				Command:   "",
				ErrorType: "empty_command",
			},
			expectContain: []string{"エラー", "指定されていません"},
		},
		{
			name: "Deprecated command",
			result: &ValidationResult{
				IsValid:     false,
				Command:     "iso-image",
				ErrorType:   "deprecated_command",
				Suggestions: []string{"cdrom"},
			},
			expectContain: []string{"注意", "iso-image", "廃止", "cdrom"},
		},
		{
			name: "Unexpected subcommand",
			result: &ValidationResult{
				IsValid:   false,
				Command:   "version",
				ErrorType: "unexpected_subcommand",
			},
			expectContain: []string{"エラー", "version", "サブコマンドを受け付けません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := generator.GenerateFromValidationResult(tt.result)

			for _, expected := range tt.expectContain {
				if !strings.Contains(message, expected) {
					t.Errorf("Expected message to contain '%s', got: %s", expected, message)
				}
			}
		})
	}
}

func TestGenerateFromSubcommandResult(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

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
			name: "Invalid subcommand",
			result: &SubcommandValidationResult{
				IsValid:     false,
				MainCommand: "server",
				SubCommand:  "invalid",
				ErrorType:   ErrorTypeInvalidSubcommand,
				Available:   []string{"list", "create", "boot"},
			},
			expectContain: []string{"エラー", "server", "有効なサブコマンドではありません", "list, create, boot"},
		},
		{
			name: "Missing subcommand",
			result: &SubcommandValidationResult{
				IsValid:     false,
				MainCommand: "server",
				SubCommand:  "",
				ErrorType:   ErrorTypeMissingSubcommand,
				Available:   []string{"list", "create"},
			},
			expectContain: []string{"エラー", "server", "サブコマンドが必要", "list, create"},
		},
		{
			name: "Unexpected subcommand",
			result: &SubcommandValidationResult{
				IsValid:     false,
				MainCommand: "version",
				SubCommand:  "list",
				ErrorType:   ErrorTypeUnexpectedSubcommand,
			},
			expectContain: []string{"エラー", "version", "サブコマンドを受け付けません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := generator.GenerateFromSubcommandResult(tt.result)

			for _, expected := range tt.expectContain {
				if !strings.Contains(message, expected) {
					t.Errorf("Expected message to contain '%s', got: %s", expected, message)
				}
			}
		})
	}
}

func TestGetSeverityString(t *testing.T) {
	tests := []struct {
		severity MessageSeverity
		expected string
	}{
		{SeverityError, "Error"},
		{SeverityWarning, "Warning"},
		{SeverityInfo, "Info"},
		{SeveritySuccess, "Success"},
	}

	for _, tt := range tests {
		result := GetSeverityString(tt.severity)
		if result != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestGetMessageTypeString(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{TypeInvalidCommand, "InvalidCommand"},
		{TypeInvalidSubcommand, "InvalidSubcommand"},
		{TypeDeprecatedCommand, "DeprecatedCommand"},
		{TypeDiscontinuedCommand, "DiscontinuedCommand"},
		{TypeSyntaxError, "SyntaxError"},
		{TypeMissingCommand, "MissingCommand"},
		{TypeMissingSubcommand, "MissingSubcommand"},
		{TypeSuggestion, "Suggestion"},
		{TypeSuccess, "Success"},
	}

	for _, tt := range tests {
		result := GetMessageTypeString(tt.msgType)
		if result != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestSetColorEnabled(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Initially disabled
	if generator.IsColorEnabled() {
		t.Error("Expected color to be initially disabled")
	}

	// Enable color
	generator.SetColorEnabled(true)
	if !generator.IsColorEnabled() {
		t.Error("Expected color to be enabled after SetColorEnabled(true)")
	}

	// Disable color
	generator.SetColorEnabled(false)
	if generator.IsColorEnabled() {
		t.Error("Expected color to be disabled after SetColorEnabled(false)")
	}
}

func TestFormatSuggestions(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Test empty suggestions
	result := generator.formatSuggestions([]string{})
	if result != "" {
		t.Error("Expected empty result for empty suggestions")
	}

	// Test with suggestions
	suggestions := []string{"server", "service", "status"}
	result = generator.formatSuggestions(suggestions)

	if !strings.Contains(result, "もしかして") {
		t.Error("Result should contain suggestion intro")
	}
	if !strings.Contains(result, "server") {
		t.Error("Result should contain 'server'")
	}
	if !strings.Contains(result, "service") {
		t.Error("Result should contain 'service'")
	}
	if !strings.Contains(result, "status") {
		t.Error("Result should contain 'status'")
	}
	if !strings.Contains(result, "•") {
		t.Error("Result should use bullet points")
	}
}

func TestExtractFormatArgs(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Test with simple parameters
	params := map[string]interface{}{
		"command":              "server",
		"subCommand":           "invalid",
		"availableSubcommands": []string{"list", "create", "boot"},
	}

	template := "Error: '%s' is not valid for '%s'. Available: %s"
	args := generator.extractFormatArgs(template, params)

	if len(args) != 3 {
		t.Errorf("Expected 3 format args, got %d", len(args))
	}

	if args[0] != "server" {
		t.Errorf("Expected first arg to be 'server', got '%v'", args[0])
	}

	if args[1] != "invalid" {
		t.Errorf("Expected second arg to be 'invalid', got '%v'", args[1])
	}

	if args[2] != "list, create, boot" {
		t.Errorf("Expected third arg to be joined subcommands, got '%v'", args[2])
	}
}

func TestMessageTemplateFields(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Test that all templates have proper fields
	for msgType := TypeInvalidCommand; msgType <= TypeSuccess; msgType++ {
		template := generator.GetTemplate(msgType)
		if template == nil {
			t.Errorf("Template for type %s should not be nil", GetMessageTypeString(msgType))
			continue
		}

		if template.Template == "" {
			t.Errorf("Template string for type %s should not be empty", GetMessageTypeString(msgType))
		}

		if template.Type != msgType {
			t.Errorf("Template type field should match the requested type")
		}

		// Check severity is within valid range
		if template.Severity < SeverityError || template.Severity > SeveritySuccess {
			t.Errorf("Template severity for type %s is out of range", GetMessageTypeString(msgType))
		}
	}
}

func TestJapaneseMessageContent(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Test that messages contain proper Japanese
	params := map[string]interface{}{
		"command": "test",
	}

	message := generator.GenerateMessage(TypeInvalidCommand, params)

	// Check for proper Japanese error indicators
	japaneseKeywords := []string{"エラー", "コマンド", "使用"}
	foundKeywords := 0
	for _, keyword := range japaneseKeywords {
		if strings.Contains(message, keyword) {
			foundKeywords++
		}
	}

	if foundKeywords == 0 {
		t.Error("Message should contain proper Japanese keywords")
	}
}

func TestUnknownMessageType(t *testing.T) {
	generator := NewErrorMessageGenerator(false)

	// Test with invalid message type (beyond the defined range)
	invalidType := MessageType(999)
	params := map[string]interface{}{}

	message := generator.GenerateMessage(invalidType, params)

	if !strings.Contains(message, "不明なエラー") {
		t.Error("Should return default error message for unknown type")
	}
}
