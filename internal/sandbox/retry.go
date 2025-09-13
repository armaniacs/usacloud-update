package sandbox

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"
)

// secureRandFloat64 はcrypto/randを使用し0-1の乱数を生成
func secureRandFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000000))
	if err != nil {
		// フォールバック: 時刻ベースの疑似乱数
		return float64(time.Now().UnixNano()%1000000000) / 1000000000.0
	}
	return float64(n.Int64()) / 1000000000.0
}

// BackoffType はバックオフ戦略の種類を定義する
type BackoffType int

const (
	BackoffFixed BackoffType = iota
	BackoffLinear
	BackoffExponential
	BackoffExponentialJitter
)

// String はバックオフタイプの文字列表現を返す
func (b BackoffType) String() string {
	switch b {
	case BackoffFixed:
		return "fixed"
	case BackoffLinear:
		return "linear"
	case BackoffExponential:
		return "exponential"
	case BackoffExponentialJitter:
		return "exponential_jitter"
	default:
		return "unknown"
	}
}

// RetryConfig はリトライ設定を定義する
type RetryConfig struct {
	MaxAttempts int                          `json:"max_attempts"`
	BaseDelay   time.Duration                `json:"base_delay"`
	MaxDelay    time.Duration                `json:"max_delay"`
	BackoffType BackoffType                  `json:"backoff_type"`
	Timeout     time.Duration                `json:"timeout,omitempty"`
	OnRetry     func(attempt int, err error) `json:"-"`
}

// DefaultRetryConfig はデフォルトのリトライ設定を返す
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		BackoffType: BackoffExponential,
		Timeout:     5 * time.Minute,
	}
}

// ExecutorInterface はExecutorのインターフェース
type ExecutorInterface interface {
	ExecuteCommand(command string) (*ExecutionResult, error)
}

// ErrorHandlerInterface はErrorHandlerのインターフェース
type ErrorHandlerInterface interface {
	Handle(err error, cmd string) *SandboxError
	GetRetryRecommendation(sandboxErr *SandboxError) *RetryConfig
}

// RetryableExecutor はリトライ機能付きでコマンドを実行する
type RetryableExecutor struct {
	executor     ExecutorInterface
	errorHandler ErrorHandlerInterface
	stats        *RetryStatistics
}

// NewRetryableExecutor は新しいRetryableExecutorを作成する
func NewRetryableExecutor(executor ExecutorInterface, errorHandler ErrorHandlerInterface) *RetryableExecutor {
	return &RetryableExecutor{
		executor:     executor,
		errorHandler: errorHandler,
		stats:        NewRetryStatistics(),
	}
}

// ExecuteWithRetry はリトライ機能付きでコマンドを実行する
func (r *RetryableExecutor) ExecuteWithRetry(ctx context.Context, cmd string, config *RetryConfig) (*ExecutionResult, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	// タイムアウト設定
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	var lastErr error
	var lastSandboxErr *SandboxError

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// 統計情報の更新
		r.stats.RecordAttempt(cmd, attempt)

		// コマンド実行
		result, err := r.executor.ExecuteCommand(cmd)
		if err == nil {
			r.stats.RecordSuccess(cmd, attempt)
			return result, nil
		}

		lastErr = err
		lastSandboxErr = r.errorHandler.Handle(err, cmd)

		// エラー統計の記録
		r.stats.RecordError(cmd, attempt, lastSandboxErr)

		// リトライ可能性の確認
		if !lastSandboxErr.Retryable {
			r.stats.RecordFinalFailure(cmd, attempt, "non-retryable error")
			return nil, fmt.Errorf("リトライ不可能なエラー: %w", err)
		}

		// 最後の試行の場合はリトライしない
		if attempt >= config.MaxAttempts {
			r.stats.RecordFinalFailure(cmd, attempt, "max attempts exceeded")
			break
		}

		// コンテキストのキャンセル確認
		if ctx.Err() != nil {
			r.stats.RecordFinalFailure(cmd, attempt, "context cancelled")
			return nil, ctx.Err()
		}

		// バックオフ待機
		delay := r.calculateBackoffDelay(attempt, config)
		r.stats.RecordRetryDelay(cmd, attempt, delay)

		// OnRetryコールバックの実行
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		select {
		case <-time.After(delay):
			// 正常に待機完了
		case <-ctx.Done():
			r.stats.RecordFinalFailure(cmd, attempt, "context cancelled during backoff")
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("最大試行回数(%d)を超過しました。最後のエラー: %w", config.MaxAttempts, lastErr)
}

// ExecuteWithAutoRetry は自動的にリトライ設定を決定してコマンドを実行する
func (r *RetryableExecutor) ExecuteWithAutoRetry(ctx context.Context, cmd string) (*ExecutionResult, error) {
	// まず1回実行してエラーの種類を判定
	result, err := r.executor.ExecuteCommand(cmd)
	if err == nil {
		return result, nil
	}

	// エラーを分析してリトライ設定を決定
	sandboxErr := r.errorHandler.Handle(err, cmd)
	if !sandboxErr.Retryable {
		return nil, err
	}

	config := r.errorHandler.GetRetryRecommendation(sandboxErr)
	if config == nil {
		return nil, err
	}

	// リトライ実行
	return r.ExecuteWithRetry(ctx, cmd, config)
}

// calculateBackoffDelay はバックオフ遅延を計算する
func (r *RetryableExecutor) calculateBackoffDelay(attempt int, config *RetryConfig) time.Duration {
	var delay time.Duration

	switch config.BackoffType {
	case BackoffFixed:
		delay = config.BaseDelay
	case BackoffLinear:
		delay = time.Duration(attempt) * config.BaseDelay
	case BackoffExponential:
		delay = time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
	case BackoffExponentialJitter:
		baseDelay := time.Duration(math.Pow(2, float64(attempt-1))) * config.BaseDelay
		// crypto/randを使用して安全なジッター生成
		jitterFloat := secureRandFloat64()
		jitter := time.Duration(jitterFloat * float64(baseDelay) * 0.1) // 10%のジッター
		delay = baseDelay + jitter
	default:
		delay = config.BaseDelay
	}

	// 最大遅延時間の制限
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}

// GetStatistics はリトライ統計情報を返す
func (r *RetryableExecutor) GetStatistics() *RetryStatistics {
	return r.stats
}

// ResetStatistics は統計情報をリセットする
func (r *RetryableExecutor) ResetStatistics() {
	r.stats = NewRetryStatistics()
}

// RetryStatistics はリトライ実行の統計情報を保持する
type RetryStatistics struct {
	StartTime       time.Time                `json:"start_time"`
	TotalCommands   int                      `json:"total_commands"`
	SuccessCount    int                      `json:"success_count"`
	FailureCount    int                      `json:"failure_count"`
	TotalAttempts   int                      `json:"total_attempts"`
	TotalRetries    int                      `json:"total_retries"`
	AverageAttempts float64                  `json:"average_attempts"`
	CommandStats    map[string]*CommandStats `json:"command_stats"`
	ErrorStats      map[SandboxErrorType]int `json:"error_stats"`
	DelayStats      *DelayStatistics         `json:"delay_stats"`
}

// CommandStats は個別コマンドの統計情報を保持する
type CommandStats struct {
	Command         string        `json:"command"`
	TotalAttempts   int           `json:"total_attempts"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	AverageAttempts float64       `json:"average_attempts"`
	TotalDelay      time.Duration `json:"total_delay"`
	LastError       *SandboxError `json:"last_error,omitempty"`
	FirstAttempt    time.Time     `json:"first_attempt"`
	LastAttempt     time.Time     `json:"last_attempt"`
}

// DelayStatistics は遅延統計情報を保持する
type DelayStatistics struct {
	TotalDelay   time.Duration `json:"total_delay"`
	AverageDelay time.Duration `json:"average_delay"`
	MinDelay     time.Duration `json:"min_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	DelayCount   int           `json:"delay_count"`
}

// NewRetryStatistics は新しいRetryStatisticsを作成する
func NewRetryStatistics() *RetryStatistics {
	return &RetryStatistics{
		StartTime:    time.Now(),
		CommandStats: make(map[string]*CommandStats),
		ErrorStats:   make(map[SandboxErrorType]int),
		DelayStats:   &DelayStatistics{},
	}
}

// RecordAttempt は試行を記録する
func (s *RetryStatistics) RecordAttempt(cmd string, attempt int) {
	s.TotalAttempts++
	if attempt > 1 {
		s.TotalRetries++
	}

	if stats, exists := s.CommandStats[cmd]; exists {
		stats.TotalAttempts++
		stats.LastAttempt = time.Now()
	} else {
		s.CommandStats[cmd] = &CommandStats{
			Command:       cmd,
			TotalAttempts: 1,
			FirstAttempt:  time.Now(),
			LastAttempt:   time.Now(),
		}
		s.TotalCommands++
	}
}

// RecordSuccess は成功を記録する
func (s *RetryStatistics) RecordSuccess(cmd string, finalAttempt int) {
	s.SuccessCount++

	if stats, exists := s.CommandStats[cmd]; exists {
		stats.SuccessCount++
		stats.AverageAttempts = float64(stats.TotalAttempts) / float64(stats.SuccessCount+stats.FailureCount)
	}

	s.AverageAttempts = float64(s.TotalAttempts) / float64(s.SuccessCount+s.FailureCount)
}

// RecordError はエラーを記録する
func (s *RetryStatistics) RecordError(cmd string, attempt int, sandboxErr *SandboxError) {
	if sandboxErr != nil {
		s.ErrorStats[sandboxErr.Type]++

		if stats, exists := s.CommandStats[cmd]; exists {
			stats.LastError = sandboxErr
		}
	}
}

// RecordFinalFailure は最終的な失敗を記録する
func (s *RetryStatistics) RecordFinalFailure(cmd string, finalAttempt int, reason string) {
	s.FailureCount++

	if stats, exists := s.CommandStats[cmd]; exists {
		stats.FailureCount++
		stats.AverageAttempts = float64(stats.TotalAttempts) / float64(stats.SuccessCount+stats.FailureCount)
	}

	s.AverageAttempts = float64(s.TotalAttempts) / float64(s.SuccessCount+s.FailureCount)
}

// RecordRetryDelay はリトライ遅延を記録する
func (s *RetryStatistics) RecordRetryDelay(cmd string, attempt int, delay time.Duration) {
	delayStats := s.DelayStats
	delayStats.TotalDelay += delay
	delayStats.DelayCount++
	delayStats.AverageDelay = delayStats.TotalDelay / time.Duration(delayStats.DelayCount)

	if delayStats.MinDelay == 0 || delay < delayStats.MinDelay {
		delayStats.MinDelay = delay
	}
	if delay > delayStats.MaxDelay {
		delayStats.MaxDelay = delay
	}

	if stats, exists := s.CommandStats[cmd]; exists {
		stats.TotalDelay += delay
	}
}

// GetSummary は統計のサマリーを返す
func (s *RetryStatistics) GetSummary() string {
	if s.TotalCommands == 0 {
		return "実行されたコマンドがありません"
	}

	successRate := float64(s.SuccessCount) / float64(s.TotalCommands) * 100

	summary := fmt.Sprintf(`リトライ統計サマリー:
  実行期間: %v
  総コマンド数: %d
  成功数: %d (%.1f%%)
  失敗数: %d
  総試行回数: %d
  総リトライ回数: %d
  平均試行回数: %.2f
  総遅延時間: %v
  平均遅延時間: %v`,
		time.Since(s.StartTime).Truncate(time.Second),
		s.TotalCommands,
		s.SuccessCount, successRate,
		s.FailureCount,
		s.TotalAttempts,
		s.TotalRetries,
		s.AverageAttempts,
		s.DelayStats.TotalDelay.Truncate(time.Millisecond),
		s.DelayStats.AverageDelay.Truncate(time.Millisecond),
	)

	if len(s.ErrorStats) > 0 {
		summary += "\n\nエラー種別内訳:"
		for errType, count := range s.ErrorStats {
			summary += fmt.Sprintf("\n  %s: %d", errType.String(), count)
		}
	}

	return summary
}

// GetDetailedReport は詳細なレポートを返す
func (s *RetryStatistics) GetDetailedReport() string {
	report := s.GetSummary()

	if len(s.CommandStats) > 0 {
		report += "\n\nコマンド別詳細:"
		for cmd, stats := range s.CommandStats {
			cmdSuccessRate := float64(stats.SuccessCount) / float64(stats.SuccessCount+stats.FailureCount) * 100
			report += fmt.Sprintf(`
  コマンド: %s
    総試行: %d, 成功: %d (%.1f%%), 失敗: %d
    平均試行回数: %.2f
    総遅延: %v
    実行期間: %v`,
				cmd,
				stats.TotalAttempts, stats.SuccessCount, cmdSuccessRate, stats.FailureCount,
				stats.AverageAttempts,
				stats.TotalDelay.Truncate(time.Millisecond),
				stats.LastAttempt.Sub(stats.FirstAttempt).Truncate(time.Second),
			)

			if stats.LastError != nil {
				report += fmt.Sprintf("\n    最新エラー: %s", stats.LastError.Message)
			}
		}
	}

	return report
}
