package sandbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockExecutor はテスト用のExecutorモック
type mockExecutor struct {
	executeFunc func(cmd string) (*ExecutionResult, error)
	callCount   int
}

func (m *mockExecutor) ExecuteCommand(cmd string) (*ExecutionResult, error) {
	m.callCount++
	if m.executeFunc != nil {
		return m.executeFunc(cmd)
	}
	return &ExecutionResult{
		Command: cmd,
		Success: true,
		Output:  "mock output",
	}, nil
}

// mockErrorHandler はテスト用のErrorHandlerモック
type mockErrorHandler struct {
	handleFunc                 func(err error, cmd string) *SandboxError
	getRetryRecommendationFunc func(sandboxErr *SandboxError) *RetryConfig
}

func (m *mockErrorHandler) Handle(err error, cmd string) *SandboxError {
	if m.handleFunc != nil {
		return m.handleFunc(err, cmd)
	}
	return &SandboxError{
		Type:      ErrorTypeUnknown,
		Message:   err.Error(),
		Command:   cmd,
		Timestamp: time.Now(),
		Retryable: true,
	}
}

func (m *mockErrorHandler) GetRetryRecommendation(sandboxErr *SandboxError) *RetryConfig {
	if m.getRetryRecommendationFunc != nil {
		return m.getRetryRecommendationFunc(sandboxErr)
	}
	return DefaultRetryConfig()
}

func TestBackoffType_String(t *testing.T) {
	tests := []struct {
		backoffType BackoffType
		expected    string
	}{
		{BackoffFixed, "fixed"},
		{BackoffLinear, "linear"},
		{BackoffExponential, "exponential"},
		{BackoffExponentialJitter, "exponential_jitter"},
		{BackoffType(999), "unknown"}, // 未定義の値
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.backoffType.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay to be 1s, got %v", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay to be 30s, got %v", config.MaxDelay)
	}
	if config.BackoffType != BackoffExponential {
		t.Errorf("Expected BackoffType to be BackoffExponential, got %v", config.BackoffType)
	}
	if config.Timeout != 5*time.Minute {
		t.Errorf("Expected Timeout to be 5m, got %v", config.Timeout)
	}
}

func TestNewRetryableExecutor(t *testing.T) {
	executor := &mockExecutor{}
	errorHandler := &mockErrorHandler{}

	retryExecutor := NewRetryableExecutor(executor, errorHandler)

	if retryExecutor.executor == nil {
		t.Error("Expected executor to be set")
	}
	if retryExecutor.errorHandler == nil {
		t.Error("Expected errorHandler to be set")
	}
	if retryExecutor.stats == nil {
		t.Error("Expected stats to be initialized")
	}
}

func TestRetryableExecutor_ExecuteWithRetry(t *testing.T) {
	t.Run("successful execution on first attempt", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success",
				}, nil
			},
		}
		errorHandler := &mockErrorHandler{}
		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			BackoffType: BackoffFixed,
		}

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", config)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
		if !result.Success {
			t.Error("Expected successful result")
		}
		if executor.callCount != 1 {
			t.Errorf("Expected 1 execution attempt, got %d", executor.callCount)
		}
	})

	t.Run("retry on retryable error", func(t *testing.T) {
		attemptCount := 0
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				attemptCount++
				if attemptCount < 3 {
					return nil, errors.New("temporary failure")
				}
				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success after retry",
				}, nil
			},
		}

		errorHandler := &mockErrorHandler{
			handleFunc: func(err error, cmd string) *SandboxError {
				return &SandboxError{
					Type:      ErrorTypeNetwork,
					Message:   err.Error(),
					Command:   cmd,
					Timestamp: time.Now(),
					Retryable: true,
				}
			},
		}

		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			BackoffType: BackoffFixed,
		}

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", config)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
		if !result.Success {
			t.Error("Expected successful result")
		}
		if executor.callCount != 3 {
			t.Errorf("Expected 3 execution attempts, got %d", executor.callCount)
		}
	})

	t.Run("fail on non-retryable error", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return nil, errors.New("authentication failed")
			},
		}

		errorHandler := &mockErrorHandler{
			handleFunc: func(err error, cmd string) *SandboxError {
				return &SandboxError{
					Type:      ErrorTypeAuth,
					Message:   err.Error(),
					Command:   cmd,
					Timestamp: time.Now(),
					Retryable: false,
				}
			},
		}

		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			BackoffType: BackoffFixed,
		}

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", config)

		if err == nil {
			t.Fatal("Expected error for non-retryable failure")
		}
		if result != nil {
			t.Error("Expected no result for non-retryable failure")
		}
		if executor.callCount != 1 {
			t.Errorf("Expected 1 execution attempt, got %d", executor.callCount)
		}
	})

	t.Run("exceed max attempts", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return nil, errors.New("persistent failure")
			},
		}

		errorHandler := &mockErrorHandler{
			handleFunc: func(err error, cmd string) *SandboxError {
				return &SandboxError{
					Type:      ErrorTypeNetwork,
					Message:   err.Error(),
					Command:   cmd,
					Timestamp: time.Now(),
					Retryable: true,
				}
			},
		}

		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()
		config := &RetryConfig{
			MaxAttempts: 2,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			BackoffType: BackoffFixed,
		}

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", config)

		if err == nil {
			t.Fatal("Expected error when max attempts exceeded")
		}
		if result != nil {
			t.Error("Expected no result when max attempts exceeded")
		}
		if executor.callCount != 2 {
			t.Errorf("Expected 2 execution attempts, got %d", executor.callCount)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				time.Sleep(100 * time.Millisecond)
				return nil, errors.New("should not reach here")
			},
		}

		errorHandler := &mockErrorHandler{}
		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		config := &RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			BackoffType: BackoffFixed,
		}

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", config)

		if err == nil {
			t.Fatal("Expected context timeout error")
		}
		if result != nil {
			t.Error("Expected no result on timeout")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("nil config uses default", func(t *testing.T) {
		executor := &mockExecutor{}
		errorHandler := &mockErrorHandler{}
		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()

		result, err := retryExecutor.ExecuteWithRetry(ctx, "test command", nil)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
	})
}

func TestRetryableExecutor_ExecuteWithAutoRetry(t *testing.T) {
	t.Run("successful execution on first attempt", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success",
				}, nil
			},
		}
		errorHandler := &mockErrorHandler{}
		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()

		result, err := retryExecutor.ExecuteWithAutoRetry(ctx, "test command")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
		if !result.Success {
			t.Error("Expected successful result")
		}
		if executor.callCount != 1 {
			t.Errorf("Expected 1 execution attempt, got %d", executor.callCount)
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return nil, errors.New("authentication failed")
			},
		}

		errorHandler := &mockErrorHandler{
			handleFunc: func(err error, cmd string) *SandboxError {
				return &SandboxError{
					Type:      ErrorTypeAuth,
					Message:   err.Error(),
					Command:   cmd,
					Timestamp: time.Now(),
					Retryable: false,
				}
			},
		}

		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()

		result, err := retryExecutor.ExecuteWithAutoRetry(ctx, "test command")

		if err == nil {
			t.Fatal("Expected error for non-retryable failure")
		}
		if result != nil {
			t.Error("Expected no result for non-retryable failure")
		}
		if executor.callCount != 1 {
			t.Errorf("Expected 1 execution attempt, got %d", executor.callCount)
		}
	})

	t.Run("retryable error with auto retry", func(t *testing.T) {
		attemptCount := 0
		executor := &mockExecutor{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				attemptCount++
				if attemptCount == 1 {
					return nil, errors.New("network error")
				}
				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success after retry",
				}, nil
			},
		}

		errorHandler := &mockErrorHandler{
			handleFunc: func(err error, cmd string) *SandboxError {
				return &SandboxError{
					Type:      ErrorTypeNetwork,
					Message:   err.Error(),
					Command:   cmd,
					Timestamp: time.Now(),
					Retryable: true,
				}
			},
			getRetryRecommendationFunc: func(sandboxErr *SandboxError) *RetryConfig {
				return &RetryConfig{
					MaxAttempts: 2,
					BaseDelay:   10 * time.Millisecond,
					MaxDelay:    100 * time.Millisecond,
					BackoffType: BackoffFixed,
				}
			},
		}

		retryExecutor := NewRetryableExecutor(executor, errorHandler)

		ctx := context.Background()

		result, err := retryExecutor.ExecuteWithAutoRetry(ctx, "test command")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
		if !result.Success {
			t.Error("Expected successful result")
		}
		// 最初の実行 + リトライでの実行 = 2回
		if executor.callCount != 2 {
			t.Errorf("Expected 2 execution attempts, got %d", executor.callCount)
		}
	})
}

func TestRetryableExecutor_calculateBackoffDelay(t *testing.T) {
	executor := &mockExecutor{}
	errorHandler := &mockErrorHandler{}
	retryExecutor := NewRetryableExecutor(executor, errorHandler)

	t.Run("fixed backoff", func(t *testing.T) {
		config := &RetryConfig{
			BackoffType: BackoffFixed,
			BaseDelay:   1 * time.Second,
			MaxDelay:    10 * time.Second,
		}

		delay1 := retryExecutor.calculateBackoffDelay(1, config)
		delay2 := retryExecutor.calculateBackoffDelay(2, config)
		delay3 := retryExecutor.calculateBackoffDelay(3, config)

		if delay1 != 1*time.Second || delay2 != 1*time.Second || delay3 != 1*time.Second {
			t.Errorf("Expected fixed delay of 1s, got %v, %v, %v", delay1, delay2, delay3)
		}
	})

	t.Run("linear backoff", func(t *testing.T) {
		config := &RetryConfig{
			BackoffType: BackoffLinear,
			BaseDelay:   1 * time.Second,
			MaxDelay:    10 * time.Second,
		}

		delay1 := retryExecutor.calculateBackoffDelay(1, config)
		delay2 := retryExecutor.calculateBackoffDelay(2, config)
		delay3 := retryExecutor.calculateBackoffDelay(3, config)

		expectedDelay1 := 1 * time.Second
		expectedDelay2 := 2 * time.Second
		expectedDelay3 := 3 * time.Second

		if delay1 != expectedDelay1 || delay2 != expectedDelay2 || delay3 != expectedDelay3 {
			t.Errorf("Expected linear delays %v, %v, %v, got %v, %v, %v",
				expectedDelay1, expectedDelay2, expectedDelay3, delay1, delay2, delay3)
		}
	})

	t.Run("exponential backoff", func(t *testing.T) {
		config := &RetryConfig{
			BackoffType: BackoffExponential,
			BaseDelay:   1 * time.Second,
			MaxDelay:    10 * time.Second,
		}

		delay1 := retryExecutor.calculateBackoffDelay(1, config)
		delay2 := retryExecutor.calculateBackoffDelay(2, config)
		delay3 := retryExecutor.calculateBackoffDelay(3, config)

		expectedDelay1 := 1 * time.Second
		expectedDelay2 := 2 * time.Second
		expectedDelay3 := 4 * time.Second

		if delay1 != expectedDelay1 || delay2 != expectedDelay2 || delay3 != expectedDelay3 {
			t.Errorf("Expected exponential delays %v, %v, %v, got %v, %v, %v",
				expectedDelay1, expectedDelay2, expectedDelay3, delay1, delay2, delay3)
		}
	})

	t.Run("max delay limit", func(t *testing.T) {
		config := &RetryConfig{
			BackoffType: BackoffExponential,
			BaseDelay:   1 * time.Second,
			MaxDelay:    3 * time.Second,
		}

		delay5 := retryExecutor.calculateBackoffDelay(5, config)

		if delay5 != 3*time.Second {
			t.Errorf("Expected delay to be capped at MaxDelay (3s), got %v", delay5)
		}
	})

	t.Run("exponential with jitter", func(t *testing.T) {
		config := &RetryConfig{
			BackoffType: BackoffExponentialJitter,
			BaseDelay:   1 * time.Second,
			MaxDelay:    10 * time.Second,
		}

		delay2 := retryExecutor.calculateBackoffDelay(2, config)

		// ジッターありなので、ベース遅延(2s)以上になる
		expectedMin := 2 * time.Second
		expectedMax := 2*time.Second + time.Duration(float64(2*time.Second)*0.1)

		if delay2 < expectedMin || delay2 > expectedMax {
			t.Errorf("Expected jittered delay between %v and %v, got %v", expectedMin, expectedMax, delay2)
		}
	})
}

func TestRetryStatistics(t *testing.T) {
	t.Run("NewRetryStatistics", func(t *testing.T) {
		stats := NewRetryStatistics()

		if stats.TotalCommands != 0 {
			t.Error("Expected TotalCommands to be 0")
		}
		if stats.CommandStats == nil {
			t.Error("Expected CommandStats to be initialized")
		}
		if stats.ErrorStats == nil {
			t.Error("Expected ErrorStats to be initialized")
		}
		if stats.DelayStats == nil {
			t.Error("Expected DelayStats to be initialized")
		}
	})

	t.Run("RecordAttempt", func(t *testing.T) {
		stats := NewRetryStatistics()

		stats.RecordAttempt("test command", 1)

		if stats.TotalAttempts != 1 {
			t.Errorf("Expected TotalAttempts to be 1, got %d", stats.TotalAttempts)
		}
		if stats.TotalRetries != 0 {
			t.Errorf("Expected TotalRetries to be 0 for first attempt, got %d", stats.TotalRetries)
		}
		if stats.TotalCommands != 1 {
			t.Errorf("Expected TotalCommands to be 1, got %d", stats.TotalCommands)
		}

		cmdStats, exists := stats.CommandStats["test command"]
		if !exists {
			t.Fatal("Expected command stats to be created")
		}
		if cmdStats.TotalAttempts != 1 {
			t.Errorf("Expected command TotalAttempts to be 1, got %d", cmdStats.TotalAttempts)
		}

		// 2回目の試行（リトライ）
		stats.RecordAttempt("test command", 2)

		if stats.TotalAttempts != 2 {
			t.Errorf("Expected TotalAttempts to be 2, got %d", stats.TotalAttempts)
		}
		if stats.TotalRetries != 1 {
			t.Errorf("Expected TotalRetries to be 1, got %d", stats.TotalRetries)
		}
		if stats.TotalCommands != 1 {
			t.Errorf("Expected TotalCommands to still be 1, got %d", stats.TotalCommands)
		}
	})

	t.Run("RecordSuccess", func(t *testing.T) {
		stats := NewRetryStatistics()
		stats.RecordAttempt("test command", 1)

		stats.RecordSuccess("test command", 1)

		if stats.SuccessCount != 1 {
			t.Errorf("Expected SuccessCount to be 1, got %d", stats.SuccessCount)
		}

		cmdStats := stats.CommandStats["test command"]
		if cmdStats.SuccessCount != 1 {
			t.Errorf("Expected command SuccessCount to be 1, got %d", cmdStats.SuccessCount)
		}
	})

	t.Run("GetSummary", func(t *testing.T) {
		stats := NewRetryStatistics()

		// コマンドがない場合
		summary := stats.GetSummary()
		if summary == "" {
			t.Error("Expected non-empty summary")
		}

		// コマンドがある場合
		stats.RecordAttempt("test command", 1)
		stats.RecordSuccess("test command", 1)

		summary = stats.GetSummary()
		if summary == "" {
			t.Error("Expected non-empty summary")
		}
	})
}
