package sandbox

import (
	"errors"
	"log"
	"strings"
	"testing"
	"time"
)

func TestNewErrorHandler(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := log.Default()
		handler := NewErrorHandler(logger, 3, 2*time.Second)

		if handler.logger != logger {
			t.Error("Expected logger to be set")
		}
		if handler.retryMax != 3 {
			t.Errorf("Expected retryMax to be 3, got %d", handler.retryMax)
		}
		if handler.retryWait != 2*time.Second {
			t.Errorf("Expected retryWait to be 2s, got %v", handler.retryWait)
		}
		if handler.debug {
			t.Error("Expected debug to be false by default")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		handler := NewErrorHandler(nil, 3, 2*time.Second)

		if handler.logger == nil {
			t.Error("Expected default logger to be set")
		}
	})
}

func TestErrorHandler_SetDebug(t *testing.T) {
	handler := NewErrorHandler(nil, 3, 2*time.Second)

	handler.SetDebug(true)
	if !handler.debug {
		t.Error("Expected debug to be true")
	}

	handler.SetDebug(false)
	if handler.debug {
		t.Error("Expected debug to be false")
	}
}

func TestErrorHandler_Handle(t *testing.T) {
	handler := NewErrorHandler(nil, 3, 2*time.Second)

	t.Run("nil error", func(t *testing.T) {
		result := handler.Handle(nil, "test")
		if result != nil {
			t.Error("Expected nil result for nil error")
		}
	})

	t.Run("timeout error", func(t *testing.T) {
		err := errors.New("command timeout occurred")
		result := handler.Handle(err, "usacloud server list")

		if result.Type != ErrorTypeTimeout {
			t.Errorf("Expected ErrorTypeTimeout, got %v", result.Type)
		}
		if !result.Retryable {
			t.Error("Expected timeout error to be retryable")
		}
		if result.Command != "usacloud server list" {
			t.Errorf("Expected command to be preserved, got %s", result.Command)
		}
		if len(result.Suggestions) == 0 {
			t.Error("Expected timeout error to have suggestions")
		}
	})

	t.Run("network error", func(t *testing.T) {
		err := errors.New("network connection failed")
		result := handler.Handle(err, "usacloud server list")

		if result.Type != ErrorTypeNetwork {
			t.Errorf("Expected ErrorTypeNetwork, got %v", result.Type)
		}
		if !result.Retryable {
			t.Error("Expected network error to be retryable")
		}
	})

	t.Run("auth error", func(t *testing.T) {
		err := errors.New("unauthorized access - invalid token")
		result := handler.Handle(err, "usacloud server list")

		if result.Type != ErrorTypeAuth {
			t.Errorf("Expected ErrorTypeAuth, got %v", result.Type)
		}
		if result.Retryable {
			t.Error("Expected auth error to be non-retryable")
		}
	})

	t.Run("permission error", func(t *testing.T) {
		err := errors.New("permission denied for this operation")
		result := handler.Handle(err, "usacloud server list")

		if result.Type != ErrorTypePermission {
			t.Errorf("Expected ErrorTypePermission, got %v", result.Type)
		}
		if result.Retryable {
			t.Error("Expected permission error to be non-retryable")
		}
	})

	t.Run("resource error", func(t *testing.T) {
		err := errors.New("resource not found")
		result := handler.Handle(err, "usacloud server show test")

		if result.Type != ErrorTypeResource {
			t.Errorf("Expected ErrorTypeResource, got %v", result.Type)
		}
		if result.Retryable {
			t.Error("Expected resource error to be non-retryable")
		}
	})

	t.Run("command error", func(t *testing.T) {
		err := errors.New("unknown command specified")
		result := handler.Handle(err, "usacloud invalid-command")

		if result.Type != ErrorTypeCommand {
			t.Errorf("Expected ErrorTypeCommand, got %v", result.Type)
		}
		if result.Retryable {
			t.Error("Expected command error to be non-retryable")
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		err := errors.New("something unexpected happened")
		result := handler.Handle(err, "usacloud server list")

		if result.Type != ErrorTypeUnknown {
			t.Errorf("Expected ErrorTypeUnknown, got %v", result.Type)
		}
		if !result.Retryable {
			t.Error("Expected unknown error to be retryable")
		}
	})
}

func TestErrorHandler_GetRetryRecommendation(t *testing.T) {
	handler := NewErrorHandler(nil, 3, 2*time.Second)

	t.Run("nil error", func(t *testing.T) {
		config := handler.GetRetryRecommendation(nil)
		if config != nil {
			t.Error("Expected nil config for nil error")
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:      ErrorTypeAuth,
			Retryable: false,
		}

		config := handler.GetRetryRecommendation(sandboxErr)
		if config != nil {
			t.Error("Expected nil config for non-retryable error")
		}
	})

	t.Run("timeout error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:      ErrorTypeTimeout,
			Retryable: true,
		}

		config := handler.GetRetryRecommendation(sandboxErr)
		if config == nil {
			t.Fatal("Expected retry config for timeout error")
		}

		if config.MaxAttempts != 2 {
			t.Errorf("Expected 2 attempts for timeout, got %d", config.MaxAttempts)
		}
		if config.BackoffType != BackoffExponential {
			t.Errorf("Expected exponential backoff for timeout, got %v", config.BackoffType)
		}
	})

	t.Run("network error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:      ErrorTypeNetwork,
			Retryable: true,
		}

		config := handler.GetRetryRecommendation(sandboxErr)
		if config == nil {
			t.Fatal("Expected retry config for network error")
		}

		if config.MaxAttempts != 3 {
			t.Errorf("Expected 3 attempts for network, got %d", config.MaxAttempts)
		}
		if config.BackoffType != BackoffLinear {
			t.Errorf("Expected linear backoff for network, got %v", config.BackoffType)
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:      ErrorTypeUnknown,
			Retryable: true,
		}

		config := handler.GetRetryRecommendation(sandboxErr)
		if config == nil {
			t.Fatal("Expected retry config for unknown error")
		}

		if config.MaxAttempts != 2 {
			t.Errorf("Expected 2 attempts for unknown, got %d", config.MaxAttempts)
		}
		if config.BackoffType != BackoffExponential {
			t.Errorf("Expected exponential backoff for unknown, got %v", config.BackoffType)
		}
	})
}

func TestErrorHandler_FormatUserMessage(t *testing.T) {
	handler := NewErrorHandler(nil, 3, 2*time.Second)

	t.Run("nil error", func(t *testing.T) {
		message := handler.FormatUserMessage(nil)
		if !strings.Contains(message, "エラー情報がありません") {
			t.Error("Expected message about no error info")
		}
	})

	t.Run("retryable error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:        ErrorTypeNetwork,
			Message:     "ネットワーク接続エラー",
			Command:     "usacloud server list",
			Timestamp:   time.Now(),
			Retryable:   true,
			Suggestions: []string{"ネットワーク接続を確認してください"},
		}

		message := handler.FormatUserMessage(sandboxErr)

		if !strings.Contains(message, "❌") {
			t.Error("Expected error emoji")
		}
		if !strings.Contains(message, "🔄") {
			t.Error("Expected retry emoji for retryable error")
		}
		if !strings.Contains(message, "ネットワーク接続エラー") {
			t.Error("Expected error message")
		}
		if !strings.Contains(message, "usacloud server list") {
			t.Error("Expected command")
		}
		if !strings.Contains(message, "💡 対処方法") {
			t.Error("Expected suggestions section")
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		sandboxErr := &SandboxError{
			Type:        ErrorTypeAuth,
			Message:     "認証エラー",
			Command:     "usacloud server list",
			Timestamp:   time.Now(),
			Retryable:   false,
			Suggestions: []string{"APIキーを確認してください"},
		}

		message := handler.FormatUserMessage(sandboxErr)

		if !strings.Contains(message, "⚠️") {
			t.Error("Expected warning emoji for non-retryable error")
		}
		if strings.Contains(message, "🔄") {
			t.Error("Did not expect retry emoji for non-retryable error")
		}
	})
}

func TestSandboxError_Error(t *testing.T) {
	sandboxErr := &SandboxError{
		Type:    ErrorTypeNetwork,
		Message: "ネットワーク接続エラー",
	}

	errorMsg := sandboxErr.Error()
	expected := "[network] ネットワーク接続エラー"

	if errorMsg != expected {
		t.Errorf("Expected '%s', got '%s'", expected, errorMsg)
	}
}

func TestSandboxErrorType_String(t *testing.T) {
	tests := []struct {
		errorType SandboxErrorType
		expected  string
	}{
		{ErrorTypeTimeout, "timeout"},
		{ErrorTypeNetwork, "network"},
		{ErrorTypeAuth, "auth"},
		{ErrorTypeCommand, "command"},
		{ErrorTypePermission, "permission"},
		{ErrorTypeResource, "resource"},
		{ErrorTypeUnknown, "unknown"},
		{SandboxErrorType(999), "unknown"}, // 未定義の値
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.errorType.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestErrorStatistics(t *testing.T) {
	t.Run("NewErrorStatistics", func(t *testing.T) {
		stats := NewErrorStatistics()

		if stats.TotalErrors != 0 {
			t.Error("Expected TotalErrors to be 0")
		}
		if stats.RetryableCount != 0 {
			t.Error("Expected RetryableCount to be 0")
		}
		if stats.ErrorsByType == nil {
			t.Error("Expected ErrorsByType to be initialized")
		}
		if stats.StartTime.IsZero() {
			t.Error("Expected StartTime to be set")
		}
	})

	t.Run("RecordError", func(t *testing.T) {
		stats := NewErrorStatistics()

		// nil errorの場合
		stats.RecordError(nil)
		if stats.TotalErrors != 0 {
			t.Error("Expected no change for nil error")
		}

		// 通常のエラー
		sandboxErr := &SandboxError{
			Type:      ErrorTypeNetwork,
			Retryable: true,
		}

		stats.RecordError(sandboxErr)

		if stats.TotalErrors != 1 {
			t.Errorf("Expected TotalErrors to be 1, got %d", stats.TotalErrors)
		}
		if stats.RetryableCount != 1 {
			t.Errorf("Expected RetryableCount to be 1, got %d", stats.RetryableCount)
		}
		if stats.ErrorsByType[ErrorTypeNetwork] != 1 {
			t.Errorf("Expected 1 network error, got %d", stats.ErrorsByType[ErrorTypeNetwork])
		}
		if stats.LastError != sandboxErr {
			t.Error("Expected LastError to be set")
		}

		// 非retryableエラー
		authErr := &SandboxError{
			Type:      ErrorTypeAuth,
			Retryable: false,
		}

		stats.RecordError(authErr)

		if stats.TotalErrors != 2 {
			t.Errorf("Expected TotalErrors to be 2, got %d", stats.TotalErrors)
		}
		if stats.RetryableCount != 1 {
			t.Errorf("Expected RetryableCount to be 1, got %d", stats.RetryableCount)
		}
	})

	t.Run("GetSummary", func(t *testing.T) {
		stats := NewErrorStatistics()

		// エラーがない場合
		summary := stats.GetSummary()
		if !strings.Contains(summary, "エラーは発生していません") {
			t.Error("Expected no error message")
		}

		// エラーがある場合
		sandboxErr := &SandboxError{
			Type:      ErrorTypeNetwork,
			Message:   "テストエラー",
			Timestamp: time.Now(),
			Retryable: true,
		}

		stats.RecordError(sandboxErr)
		summary = stats.GetSummary()

		if !strings.Contains(summary, "総エラー数: 1") {
			t.Error("Expected total error count")
		}
		if !strings.Contains(summary, "再試行可能: 1") {
			t.Error("Expected retryable count")
		}
		if !strings.Contains(summary, "network: 1") {
			t.Error("Expected error type breakdown")
		}
		if !strings.Contains(summary, "テストエラー") {
			t.Error("Expected last error message")
		}
	})
}
