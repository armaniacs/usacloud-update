package errors

import (
	"fmt"

	"github.com/fatih/color"
)

// ErrorType represents different types of errors
type ErrorType int

const (
	ErrorFileNotFound ErrorType = iota
	ErrorFilePermission
	ErrorFileBinary
	ErrorFileRead
	ErrorFileWrite
	ErrorValidation
	ErrorConfig
	ErrorGeneral
)

// ErrorFormatter provides unified error message formatting
type ErrorFormatter struct {
	colorEnabled bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(colorEnabled bool) *ErrorFormatter {
	return &ErrorFormatter{
		colorEnabled: colorEnabled,
	}
}

// FormatError formats an error message with consistent styling
func (ef *ErrorFormatter) FormatError(errorType ErrorType, message string, details ...string) string {
	var icon string
	var prefix string

	switch errorType {
	case ErrorFileNotFound:
		icon = "❌"
		prefix = "ファイルエラー"
	case ErrorFilePermission:
		icon = "🔒"
		prefix = "権限エラー"
	case ErrorFileBinary:
		icon = "⚠️"
		prefix = "ファイル形式エラー"
	case ErrorFileRead:
		icon = "📖"
		prefix = "読み込みエラー"
	case ErrorFileWrite:
		icon = "💾"
		prefix = "書き込みエラー"
	case ErrorValidation:
		icon = "🔍"
		prefix = "検証エラー"
	case ErrorConfig:
		icon = "⚙️"
		prefix = "設定エラー"
	default:
		icon = "❌"
		prefix = "エラー"
	}

	// Format main error message
	errorMsg := fmt.Sprintf("%s %s: %s", icon, prefix, message)

	if ef.colorEnabled {
		errorMsg = color.RedString(errorMsg)
	}

	// Add details if provided
	if len(details) > 0 {
		for _, detail := range details {
			if ef.colorEnabled {
				errorMsg += "\n" + color.YellowString("   詳細: %s", detail)
			} else {
				errorMsg += fmt.Sprintf("\n   詳細: %s", detail)
			}
		}
	}

	return errorMsg
}

// FormatFileNotFound formats file not found errors
func (ef *ErrorFormatter) FormatFileNotFound(filePath string) string {
	return ef.FormatError(ErrorFileNotFound, fmt.Sprintf("ファイルが見つかりません: %s", filePath))
}

// FormatFilePermission formats file permission errors
func (ef *ErrorFormatter) FormatFilePermission(filePath string, operation string) string {
	message := fmt.Sprintf("%s権限がありません: %s", operation, filePath)
	return ef.FormatError(ErrorFilePermission, message)
}

// FormatBinaryFile formats binary file errors
func (ef *ErrorFormatter) FormatBinaryFile(filePath string) string {
	return ef.FormatError(ErrorFileBinary, fmt.Sprintf("バイナリファイルは処理できません: %s", filePath))
}

// FormatFileRead formats file read errors
func (ef *ErrorFormatter) FormatFileRead(filePath string, err error) string {
	return ef.FormatError(ErrorFileRead, fmt.Sprintf("ファイル読み込み失敗: %s", filePath), err.Error())
}

// FormatFileWrite formats file write errors
func (ef *ErrorFormatter) FormatFileWrite(filePath string, err error) string {
	return ef.FormatError(ErrorFileWrite, fmt.Sprintf("ファイル書き込み失敗: %s", filePath), err.Error())
}

// FormatValidation formats validation errors
func (ef *ErrorFormatter) FormatValidation(message string, details ...string) string {
	return ef.FormatError(ErrorValidation, message, details...)
}

// FormatConfig formats configuration errors
func (ef *ErrorFormatter) FormatConfig(message string, details ...string) string {
	return ef.FormatError(ErrorConfig, message, details...)
}
