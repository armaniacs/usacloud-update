package sandbox

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// SandboxErrorType はサンドボックス実行エラーの種類を定義する
type SandboxErrorType int

const (
	ErrorTypeTimeout SandboxErrorType = iota
	ErrorTypeNetwork
	ErrorTypeAuth
	ErrorTypeCommand
	ErrorTypePermission
	ErrorTypeResource
	ErrorTypeUnknown
)

// String はエラータイプの文字列表現を返す
func (e SandboxErrorType) String() string {
	switch e {
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeAuth:
		return "auth"
	case ErrorTypeCommand:
		return "command"
	case ErrorTypePermission:
		return "permission"
	case ErrorTypeResource:
		return "resource"
	default:
		return "unknown"
	}
}

// SandboxError はサンドボックス実行エラーの詳細情報を保持する
type SandboxError struct {
	Type        SandboxErrorType       `json:"type"`
	Message     string                 `json:"message"`
	Command     string                 `json:"command"`
	Timestamp   time.Time              `json:"timestamp"`
	Retryable   bool                   `json:"retryable"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context,omitempty"`
	OriginalErr error                  `json:"-"`
}

// Error はerrorインターフェースを実装する
func (e *SandboxError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// ErrorHandler はサンドボックス実行エラーの処理を担当する
type ErrorHandler struct {
	logger    *log.Logger
	retryMax  int
	retryWait time.Duration
	debug     bool
}

// NewErrorHandler は新しいErrorHandlerを作成する
func NewErrorHandler(logger *log.Logger, retryMax int, retryWait time.Duration) *ErrorHandler {
	if logger == nil {
		logger = log.Default()
	}
	return &ErrorHandler{
		logger:    logger,
		retryMax:  retryMax,
		retryWait: retryWait,
		debug:     false,
	}
}

// SetDebug はデバッグモードを設定する
func (h *ErrorHandler) SetDebug(debug bool) {
	h.debug = debug
}

// Handle はエラーを分類し、適切なSandboxErrorを生成する
func (h *ErrorHandler) Handle(err error, cmd string) *SandboxError {
	if err == nil {
		return nil
	}

	sandboxErr := h.classifyError(err, cmd)
	h.logError(sandboxErr)
	return sandboxErr
}

// classifyError は入力エラーを分類し、適切なSandboxErrorを生成する
func (h *ErrorHandler) classifyError(err error, cmd string) *SandboxError {
	errMsg := strings.ToLower(err.Error())

	// タイムアウトエラー
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
		return &SandboxError{
			Type:      ErrorTypeTimeout,
			Message:   "コマンド実行がタイムアウトしました",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: true,
			Suggestions: []string{
				"--timeout オプションで実行時間を延長してください",
				"ネットワーク接続を確認してください",
				"コマンドの複雑さを確認し、より軽量な操作に分割してください",
			},
			Context:     map[string]interface{}{"retry_recommended": true},
			OriginalErr: err,
		}
	}

	// ネットワークエラー
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "dns") || strings.Contains(errMsg, "resolve") {
		return &SandboxError{
			Type:      ErrorTypeNetwork,
			Message:   "ネットワーク接続エラーが発生しました",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: true,
			Suggestions: []string{
				"インターネット接続を確認してください",
				"DNSの設定を確認してください",
				"プロキシ設定を確認してください",
				"しばらく時間をおいて再実行してください",
			},
			Context:     map[string]interface{}{"network_issue": true},
			OriginalErr: err,
		}
	}

	// 認証エラー
	if strings.Contains(errMsg, "auth") || strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "token") ||
		strings.Contains(errMsg, "api key") {
		return &SandboxError{
			Type:      ErrorTypeAuth,
			Message:   "認証エラーが発生しました",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: false,
			Suggestions: []string{
				"APIキーが正しく設定されているか確認してください",
				"アクセストークンの有効期限を確認してください",
				"設定ファイル（~/.config/usacloud-update/usacloud-update.conf）を確認してください",
				"usacloud auth コマンドで認証情報を再設定してください",
			},
			Context:     map[string]interface{}{"config_check_required": true},
			OriginalErr: err,
		}
	}

	// 権限エラー
	if strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "not allowed") {
		return &SandboxError{
			Type:      ErrorTypePermission,
			Message:   "権限エラーが発生しました",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: false,
			Suggestions: []string{
				"実行権限があることを確認してください",
				"必要なリソースへのアクセス権限を確認してください",
				"管理者権限が必要な操作の場合は、適切な権限で実行してください",
			},
			Context:     map[string]interface{}{"permission_issue": true},
			OriginalErr: err,
		}
	}

	// リソースエラー
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "no such") {
		return &SandboxError{
			Type:      ErrorTypeResource,
			Message:   "指定されたリソースが見つかりません",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: false,
			Suggestions: []string{
				"指定したリソース名が正しいか確認してください",
				"リソースが存在するゾーンを確認してください",
				"usacloud コマンドでリソースの一覧を確認してください",
			},
			Context:     map[string]interface{}{"resource_not_found": true},
			OriginalErr: err,
		}
	}

	// コマンドエラー
	if strings.Contains(errMsg, "command not found") || strings.Contains(errMsg, "unknown command") ||
		strings.Contains(errMsg, "invalid argument") {
		return &SandboxError{
			Type:      ErrorTypeCommand,
			Message:   "コマンドエラーが発生しました",
			Command:   cmd,
			Timestamp: time.Now(),
			Retryable: false,
			Suggestions: []string{
				"コマンドの構文を確認してください",
				"usacloud --help でヘルプを確認してください",
				"usacloudのバージョンを確認してください",
				"変換前のコマンドに問題がないか確認してください",
			},
			Context:     map[string]interface{}{"command_syntax_error": true},
			OriginalErr: err,
		}
	}

	// 未知のエラー
	return &SandboxError{
		Type:      ErrorTypeUnknown,
		Message:   fmt.Sprintf("予期しないエラーが発生しました: %s", err.Error()),
		Command:   cmd,
		Timestamp: time.Now(),
		Retryable: true,
		Suggestions: []string{
			"一時的な問題の可能性があります。しばらく待ってから再実行してください",
			"--debug フラグを使用して詳細なログを確認してください",
			"問題が継続する場合は、サポートにお問い合わせください",
		},
		Context:     map[string]interface{}{"unknown_error": true, "requires_investigation": true},
		OriginalErr: err,
	}
}

// logError はエラーをログに記録する
func (h *ErrorHandler) logError(sandboxErr *SandboxError) {
	if sandboxErr == nil {
		return
	}

	logLevel := "ERROR"
	if sandboxErr.Retryable {
		logLevel = "WARN"
	}

	logMsg := fmt.Sprintf("[%s] [%s] %s (Command: %s)",
		logLevel,
		sandboxErr.Type.String(),
		sandboxErr.Message,
		sandboxErr.Command,
	)

	if h.debug && sandboxErr.OriginalErr != nil {
		logMsg += fmt.Sprintf(" (Original: %s)", sandboxErr.OriginalErr.Error())
	}

	h.logger.Println(logMsg)

	// 重要なエラーの場合は追加情報をログ出力
	if sandboxErr.Type == ErrorTypeAuth || sandboxErr.Type == ErrorTypeUnknown {
		h.logger.Printf("Suggestions for %s: %v", sandboxErr.Type.String(), sandboxErr.Suggestions)
	}
}

// GetRetryRecommendation はリトライの推奨設定を返す
func (h *ErrorHandler) GetRetryRecommendation(sandboxErr *SandboxError) *RetryConfig {
	if sandboxErr == nil || !sandboxErr.Retryable {
		return nil
	}

	switch sandboxErr.Type {
	case ErrorTypeTimeout:
		return &RetryConfig{
			MaxAttempts: 2,
			BaseDelay:   5 * time.Second,
			MaxDelay:    30 * time.Second,
			BackoffType: BackoffExponential,
		}
	case ErrorTypeNetwork:
		return &RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   2 * time.Second,
			MaxDelay:    10 * time.Second,
			BackoffType: BackoffLinear,
		}
	case ErrorTypeUnknown:
		return &RetryConfig{
			MaxAttempts: 2,
			BaseDelay:   3 * time.Second,
			MaxDelay:    15 * time.Second,
			BackoffType: BackoffExponential,
		}
	default:
		return &RetryConfig{
			MaxAttempts: 1,
			BaseDelay:   1 * time.Second,
			MaxDelay:    5 * time.Second,
			BackoffType: BackoffFixed,
		}
	}
}

// FormatUserMessage はユーザー向けのエラーメッセージを整形する
func (h *ErrorHandler) FormatUserMessage(sandboxErr *SandboxError) string {
	if sandboxErr == nil {
		return "エラー情報がありません"
	}

	var msg strings.Builder

	// エラーメッセージ
	msg.WriteString(fmt.Sprintf("❌ %s\n", sandboxErr.Message))
	msg.WriteString(fmt.Sprintf("📝 コマンド: %s\n", sandboxErr.Command))
	msg.WriteString(fmt.Sprintf("⏰ 発生時刻: %s\n", sandboxErr.Timestamp.Format("2006-01-02 15:04:05")))

	// リトライ可能性
	if sandboxErr.Retryable {
		msg.WriteString("🔄 このエラーは再実行で解決する可能性があります\n")
	} else {
		msg.WriteString("⚠️  このエラーは設定や操作の修正が必要です\n")
	}

	// 提案
	if len(sandboxErr.Suggestions) > 0 {
		msg.WriteString("\n💡 対処方法:\n")
		for i, suggestion := range sandboxErr.Suggestions {
			msg.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
		}
	}

	return msg.String()
}

// ErrorStatistics はエラー統計情報を保持する
type ErrorStatistics struct {
	TotalErrors    int                      `json:"total_errors"`
	ErrorsByType   map[SandboxErrorType]int `json:"errors_by_type"`
	RetryableCount int                      `json:"retryable_count"`
	LastError      *SandboxError            `json:"last_error,omitempty"`
	StartTime      time.Time                `json:"start_time"`
}

// NewErrorStatistics は新しいErrorStatisticsを作成する
func NewErrorStatistics() *ErrorStatistics {
	return &ErrorStatistics{
		ErrorsByType: make(map[SandboxErrorType]int),
		StartTime:    time.Now(),
	}
}

// RecordError はエラーを統計に記録する
func (s *ErrorStatistics) RecordError(sandboxErr *SandboxError) {
	if sandboxErr == nil {
		return
	}

	s.TotalErrors++
	s.ErrorsByType[sandboxErr.Type]++
	if sandboxErr.Retryable {
		s.RetryableCount++
	}
	s.LastError = sandboxErr
}

// GetSummary は統計のサマリーを返す
func (s *ErrorStatistics) GetSummary() string {
	if s.TotalErrors == 0 {
		return "エラーは発生していません"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("総エラー数: %d\n", s.TotalErrors))
	summary.WriteString(fmt.Sprintf("再試行可能: %d\n", s.RetryableCount))
	summary.WriteString("エラー種別内訳:\n")

	for errType, count := range s.ErrorsByType {
		summary.WriteString(fmt.Sprintf("  %s: %d\n", errType.String(), count))
	}

	if s.LastError != nil {
		summary.WriteString(fmt.Sprintf("最新エラー: %s (%s)\n",
			s.LastError.Message, s.LastError.Timestamp.Format("15:04:05")))
	}

	return summary.String()
}
