// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"fmt"
	"strings"
)

// MessageSeverity represents message severity levels
type MessageSeverity int

const (
	SeverityError   MessageSeverity = iota // Error (red)
	SeverityWarning                        // Warning (yellow)
	SeverityInfo                           // Information (blue)
	SeveritySuccess                        // Success (green)
)

// MessageType represents message types
type MessageType int

const (
	TypeInvalidCommand MessageType = iota
	TypeInvalidSubcommand
	TypeDeprecatedCommand
	TypeDiscontinuedCommand
	TypeSyntaxError
	TypeMissingCommand
	TypeMissingSubcommand
	TypeSuggestion
	TypeSuccess
)

// ColorCode represents terminal color codes
type ColorCode string

const (
	ColorRed    ColorCode = "\033[31m"
	ColorYellow ColorCode = "\033[33m"
	ColorBlue   ColorCode = "\033[34m"
	ColorGreen  ColorCode = "\033[32m"
	ColorReset  ColorCode = "\033[0m"
)

// MessageTemplate represents a message template
type MessageTemplate struct {
	Template    string          // Message template
	Severity    MessageSeverity // Message severity
	Type        MessageType     // Message type
	Suggestions bool            // Whether to include suggestions
}

// ErrorMessageGenerator represents an error message generator
type ErrorMessageGenerator struct {
	templates    map[MessageType]*MessageTemplate
	colorEnabled bool
}

// NewErrorMessageGenerator creates a new error message generator
func NewErrorMessageGenerator(colorEnabled bool) *ErrorMessageGenerator {
	generator := &ErrorMessageGenerator{
		templates:    make(map[MessageType]*MessageTemplate),
		colorEnabled: colorEnabled,
	}

	// Initialize message templates
	generator.initializeTemplates()

	return generator
}

// initializeTemplates initializes all message templates
func (g *ErrorMessageGenerator) initializeTemplates() {
	g.templates[TypeInvalidCommand] = &MessageTemplate{
		Template: "エラー: '%s' は有効なusacloudコマンドではありません。\n" +
			"利用可能なコマンドを確認するには 'usacloud --help' を実行してください。",
		Severity:    SeverityError,
		Type:        TypeInvalidCommand,
		Suggestions: true,
	}

	g.templates[TypeInvalidSubcommand] = &MessageTemplate{
		Template: "エラー: '%s' は %s コマンドの有効なサブコマンドではありません。\n" +
			"%s コマンドで利用可能なサブコマンド: %s",
		Severity:    SeverityError,
		Type:        TypeInvalidSubcommand,
		Suggestions: true,
	}

	g.templates[TypeDeprecatedCommand] = &MessageTemplate{
		Template: "注意: '%s' コマンドはv1で廃止されました。\n" +
			"代わりに '%s' を使用してください。",
		Severity:    SeverityWarning,
		Type:        TypeDeprecatedCommand,
		Suggestions: true,
	}

	g.templates[TypeDiscontinuedCommand] = &MessageTemplate{
		Template: "注意: '%s' コマンドはv1で完全に廃止されました。\n" +
			"代替手段:\n%s",
		Severity:    SeverityWarning,
		Type:        TypeDiscontinuedCommand,
		Suggestions: true,
	}

	g.templates[TypeSyntaxError] = &MessageTemplate{
		Template: "エラー: '%s' コマンドはサブコマンドを受け付けません。\n" +
			"正しい使用法: usacloud %s",
		Severity:    SeverityError,
		Type:        TypeSyntaxError,
		Suggestions: true,
	}

	g.templates[TypeMissingCommand] = &MessageTemplate{
		Template: "エラー: メインコマンドが指定されていません。\n" +
			"使用法: usacloud <command> [subcommand] [options]",
		Severity:    SeverityError,
		Type:        TypeMissingCommand,
		Suggestions: true,
	}

	g.templates[TypeMissingSubcommand] = &MessageTemplate{
		Template: "エラー: '%s' コマンドにはサブコマンドが必要です。\n" +
			"利用可能なサブコマンド: %s",
		Severity:    SeverityError,
		Type:        TypeMissingSubcommand,
		Suggestions: true,
	}

	g.templates[TypeSuggestion] = &MessageTemplate{
		Template:    "ヒント: 次のコマンドをお試しください: %s",
		Severity:    SeverityInfo,
		Type:        TypeSuggestion,
		Suggestions: true,
	}

	g.templates[TypeSuccess] = &MessageTemplate{
		Template:    "✅ %s",
		Severity:    SeveritySuccess,
		Type:        TypeSuccess,
		Suggestions: false,
	}
}

// GenerateMessage generates an error message based on type and parameters
func (g *ErrorMessageGenerator) GenerateMessage(msgType MessageType, params map[string]interface{}) string {
	template, exists := g.templates[msgType]
	if !exists {
		return "不明なエラーが発生しました。"
	}

	message := g.formatMessage(template.Template, params)

	// Add suggestions if available and template supports them
	if template.Suggestions {
		if suggestions, ok := params["suggestions"].([]string); ok && len(suggestions) > 0 {
			message += g.formatSuggestions(suggestions)
		}
	}

	// Colorize the message if color is enabled
	if g.colorEnabled {
		message = g.colorizeMessage(message, template.Severity)
	}

	return message
}

// formatMessage formats the message template with parameters
func (g *ErrorMessageGenerator) formatMessage(template string, params map[string]interface{}) string {
	switch {
	case strings.Contains(template, "%s"):
		// Handle format string templates
		args := g.extractFormatArgs(template, params)
		return fmt.Sprintf(template, args...)
	default:
		// Handle simple replacement templates
		result := template
		for key, value := range params {
			placeholder := fmt.Sprintf("{%s}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
		return result
	}
}

// extractFormatArgs extracts format arguments based on template and parameters
func (g *ErrorMessageGenerator) extractFormatArgs(template string, params map[string]interface{}) []interface{} {
	var args []interface{}

	// Count format specifiers
	formatCount := strings.Count(template, "%s")

	// Template-specific parameter extraction
	switch {
	case strings.Contains(template, "は有効なusacloudコマンドではありません"):
		// TypeInvalidCommand: '%s' は有効なusacloudコマンドではありません
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}

	case strings.Contains(template, "コマンドの有効なサブコマンドではありません"):
		// TypeInvalidSubcommand: '%s' は %s コマンドの有効なサブコマンドではありません。%s コマンドで利用可能なサブコマンド: %s
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}
		if formatCount >= 2 {
			args = append(args, getParamString(params, "mainCommand"))
		}
		if formatCount >= 3 {
			args = append(args, getParamString(params, "mainCommand"))
		}
		if formatCount >= 4 {
			if slice, ok := params["availableSubcommands"].([]string); ok {
				args = append(args, strings.Join(slice, ", "))
			} else {
				args = append(args, "")
			}
		}

	case strings.Contains(template, "代わりに") && strings.Contains(template, "を使用してください"):
		// TypeDeprecatedCommand: '%s' は '%s' に名称変更されました
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}
		if formatCount >= 2 {
			args = append(args, getParamString(params, "replacementCommand"))
		}

	case strings.Contains(template, "完全に廃止されました"):
		// TypeDiscontinuedCommand: '%s' コマンドは完全に廃止されました。代替手段: %s
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}
		if formatCount >= 2 {
			args = append(args, getParamString(params, "alternativeActions"))
		}

	case strings.Contains(template, "サブコマンドを受け付けません"):
		// TypeSyntaxError: '%s' コマンドはサブコマンドを受け付けません。正しい使用法: usacloud %s
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}
		if formatCount >= 2 {
			args = append(args, getParamString(params, "command"))
		}

	case strings.Contains(template, "サブコマンドが必要"):
		// TypeMissingSubcommand: '%s' コマンドにはサブコマンドが必要です。利用可能なサブコマンド: %s
		if formatCount >= 1 {
			args = append(args, getParamString(params, "command"))
		}
		if formatCount >= 2 {
			if slice, ok := params["availableSubcommands"].([]string); ok {
				args = append(args, strings.Join(slice, ", "))
			} else {
				args = append(args, "")
			}
		}

	case strings.Contains(template, "✅"):
		// TypeSuccess: ✅ %s
		if formatCount >= 1 {
			if message, ok := params["message"]; ok {
				args = append(args, fmt.Sprintf("%v", message))
			} else {
				args = append(args, "")
			}
		}

	default:
		// Generic parameter order for other templates - reordered to put availableSubcommands before mainCommand
		paramOrder := []string{
			"command", "subCommand", "availableSubcommands", "mainCommand",
			"replacementCommand", "alternativeActions", "message",
		}

		for i := 0; i < formatCount; i++ {
			if i < len(paramOrder) {
				key := paramOrder[i]
				if value, exists := params[key]; exists {
					// Special handling for string slices
					if slice, ok := value.([]string); ok {
						args = append(args, strings.Join(slice, ", "))
					} else {
						args = append(args, fmt.Sprintf("%v", value))
					}
				} else {
					args = append(args, "")
				}
			} else {
				args = append(args, "")
			}
		}
	}

	return args
}

// getParamString gets a parameter as a string, with empty string as default
func getParamString(params map[string]interface{}, key string) string {
	if value, exists := params[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// formatSuggestions formats suggestion messages
func (g *ErrorMessageGenerator) formatSuggestions(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	result := "\n\nもしかして:"
	for _, suggestion := range suggestions {
		result += fmt.Sprintf("\n  • %s", suggestion)
	}

	return result
}

// colorizeMessage adds color to the message based on severity
func (g *ErrorMessageGenerator) colorizeMessage(message string, severity MessageSeverity) string {
	if !g.colorEnabled {
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

// GenerateFromValidationResult generates a message from validation results
func (g *ErrorMessageGenerator) GenerateFromValidationResult(result *ValidationResult) string {
	if result.IsValid {
		return g.GenerateMessage(TypeSuccess, map[string]interface{}{
			"message": "コマンドは正常です",
		})
	}

	params := map[string]interface{}{
		"command": result.Command,
	}

	// Add suggestions if available
	if len(result.Suggestions) > 0 {
		params["suggestions"] = result.Suggestions
	}

	switch result.ErrorType {
	case "empty_command":
		return g.GenerateMessage(TypeMissingCommand, params)
	case "unknown_command":
		return g.GenerateMessage(TypeInvalidCommand, params)
	case "deprecated_command":
		params["replacementCommand"] = result.Suggestions[0] // First suggestion is replacement
		return g.GenerateMessage(TypeDeprecatedCommand, params)
	case "discontinued_command":
		// Get alternative actions from deprecated detector
		detector := NewDeprecatedCommandDetector()
		actions := detector.GetAlternativeActions(result.Command)
		if len(actions) > 0 {
			actionList := ""
			for _, action := range actions {
				actionList += fmt.Sprintf("  • %s\n", action)
			}
			params["alternativeActions"] = actionList
			return g.GenerateMessage(TypeDiscontinuedCommand, params)
		}
		return g.GenerateMessage(TypeDeprecatedCommand, params)
	case "unexpected_subcommand":
		return g.GenerateMessage(TypeSyntaxError, params)
	default:
		return g.GenerateMessage(TypeInvalidCommand, params)
	}
}

// GenerateFromSubcommandResult generates a message from subcommand validation results
func (g *ErrorMessageGenerator) GenerateFromSubcommandResult(result *SubcommandValidationResult) string {
	if result.IsValid {
		return g.GenerateMessage(TypeSuccess, map[string]interface{}{
			"message": "サブコマンドは正常です",
		})
	}

	params := map[string]interface{}{
		"command":     result.MainCommand,
		"subCommand":  result.SubCommand,
		"mainCommand": result.MainCommand,
	}

	// Add available subcommands
	if len(result.Available) > 0 {
		params["availableSubcommands"] = result.Available
	}

	// Add suggestions if available
	if len(result.Suggestions) > 0 {
		params["suggestions"] = result.Suggestions
	}

	switch result.ErrorType {
	case ErrorTypeMissingSubcommand:
		return g.GenerateMessage(TypeMissingSubcommand, params)
	case ErrorTypeInvalidSubcommand:
		return g.GenerateMessage(TypeInvalidSubcommand, params)
	case ErrorTypeUnexpectedSubcommand:
		return g.GenerateMessage(TypeSyntaxError, params)
	default:
		return g.GenerateMessage(TypeInvalidSubcommand, params)
	}
}

// GetSeverityString returns the severity as a string
func GetSeverityString(severity MessageSeverity) string {
	switch severity {
	case SeverityError:
		return "Error"
	case SeverityWarning:
		return "Warning"
	case SeverityInfo:
		return "Info"
	case SeveritySuccess:
		return "Success"
	default:
		return "Unknown"
	}
}

// GetMessageTypeString returns the message type as a string
func GetMessageTypeString(msgType MessageType) string {
	switch msgType {
	case TypeInvalidCommand:
		return "InvalidCommand"
	case TypeInvalidSubcommand:
		return "InvalidSubcommand"
	case TypeDeprecatedCommand:
		return "DeprecatedCommand"
	case TypeDiscontinuedCommand:
		return "DiscontinuedCommand"
	case TypeSyntaxError:
		return "SyntaxError"
	case TypeMissingCommand:
		return "MissingCommand"
	case TypeMissingSubcommand:
		return "MissingSubcommand"
	case TypeSuggestion:
		return "Suggestion"
	case TypeSuccess:
		return "Success"
	default:
		return "Unknown"
	}
}

// SetColorEnabled enables or disables color output
func (g *ErrorMessageGenerator) SetColorEnabled(enabled bool) {
	g.colorEnabled = enabled
}

// IsColorEnabled returns whether color output is enabled
func (g *ErrorMessageGenerator) IsColorEnabled() bool {
	return g.colorEnabled
}

// GetTemplate returns the template for a given message type
func (g *ErrorMessageGenerator) GetTemplate(msgType MessageType) *MessageTemplate {
	return g.templates[msgType]
}

// GetAllTemplates returns all templates
func (g *ErrorMessageGenerator) GetAllTemplates() map[MessageType]*MessageTemplate {
	result := make(map[MessageType]*MessageTemplate)
	for msgType, template := range g.templates {
		result[msgType] = &MessageTemplate{
			Template:    template.Template,
			Severity:    template.Severity,
			Type:        template.Type,
			Suggestions: template.Suggestions,
		}
	}
	return result
}
