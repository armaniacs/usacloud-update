package sandbox

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/time/rate"
)

// JobStatus はジョブの実行状態を表す
type JobStatus int

const (
	JobPending JobStatus = iota
	JobRunning
	JobCompleted
	JobFailed
	JobCancelled
)

// String はJobStatusの文字列表現を返す
func (j JobStatus) String() string {
	switch j {
	case JobPending:
		return "pending"
	case JobRunning:
		return "running"
	case JobCompleted:
		return "completed"
	case JobFailed:
		return "failed"
	case JobCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Job は並行実行するジョブを表す
type Job struct {
	ID        string                 `json:"id"`
	Command   string                 `json:"command"`
	File      string                 `json:"file,omitempty"`
	Status    JobStatus              `json:"status"`
	Result    *ExecutionResult       `json:"result,omitempty"`
	Error     error                  `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorPolicy はエラー発生時の処理方針を表す
type ErrorPolicy int

const (
	ContinueOnError ErrorPolicy = iota
	StopOnError
	StopOnCriticalError
)

// String はErrorPolicyの文字列表現を返す
func (e ErrorPolicy) String() string {
	switch e {
	case ContinueOnError:
		return "continue_on_error"
	case StopOnError:
		return "stop_on_error"
	case StopOnCriticalError:
		return "stop_on_critical_error"
	default:
		return "unknown"
	}
}

// ParallelConfig は並行実行の設定を表す
type ParallelConfig struct {
	MaxConcurrency int           `json:"max_concurrency"`
	RateLimit      float64       `json:"rate_limit"` // requests per second
	ErrorPolicy    ErrorPolicy   `json:"error_policy"`
	Timeout        time.Duration `json:"timeout"`
	ShowProgress   bool          `json:"show_progress"`
	Debug          bool          `json:"debug"`
}

// DefaultParallelConfig はデフォルトの並行実行設定を返す
func DefaultParallelConfig() *ParallelConfig {
	return &ParallelConfig{
		MaxConcurrency: 5,
		RateLimit:      2.0, // 2 requests per second for Sakura Cloud API
		ErrorPolicy:    ContinueOnError,
		Timeout:        5 * time.Minute,
		ShowProgress:   true,
		Debug:          false,
	}
}

// ParallelExecutor は並行実行を管理する
type ParallelExecutor struct {
	executor    ExecutorInterface
	config      *ParallelConfig
	rateLimiter *rate.Limiter
	semaphore   chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	jobs        map[string]*Job
	monitor     *ProgressMonitor
}

// NewParallelExecutor は新しいParallelExecutorを作成する
func NewParallelExecutor(executor ExecutorInterface, config *ParallelConfig) *ParallelExecutor {
	if config == nil {
		config = DefaultParallelConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), config.Timeout)
	}

	pe := &ParallelExecutor{
		executor:    executor,
		config:      config,
		rateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit), 1),
		semaphore:   make(chan struct{}, config.MaxConcurrency),
		ctx:         ctx,
		cancel:      cancel,
		jobs:        make(map[string]*Job),
		monitor:     NewProgressMonitor(),
	}

	return pe
}

// SubmitJobs は複数のジョブを並行実行する
func (pe *ParallelExecutor) SubmitJobs(jobs []*Job) <-chan *Job {
	resultChan := make(chan *Job, len(jobs))

	// ジョブをマップに登録
	pe.mu.Lock()
	for _, job := range jobs {
		if job.ID == "" {
			job.ID = fmt.Sprintf("job_%d", time.Now().UnixNano())
		}
		job.Status = JobPending
		pe.jobs[job.ID] = job
	}
	pe.mu.Unlock()

	// プログレスモニター初期化
	if pe.config.ShowProgress {
		pe.monitor.Start(len(jobs))
	}

	// 各ジョブを並行実行
	for _, job := range jobs {
		pe.wg.Add(1)
		go pe.executeJob(job, resultChan)
	}

	// 完了待ちの goroutine
	go func() {
		pe.wg.Wait()
		close(resultChan)
		if pe.config.ShowProgress {
			pe.monitor.Finish()
		}
	}()

	return resultChan
}

// ExecuteJob は単一のジョブを実行する
func (pe *ParallelExecutor) ExecuteJob(job *Job) (*Job, error) {
	jobs := []*Job{job}
	resultChan := pe.SubmitJobs(jobs)

	// 結果を待機
	for result := range resultChan {
		if result.ID == job.ID {
			return result, result.Error
		}
	}

	return nil, fmt.Errorf("job execution failed")
}

// executeJob は内部でジョブを実行する
func (pe *ParallelExecutor) executeJob(job *Job, resultChan chan<- *Job) {
	defer pe.wg.Done()

	// セマフォによる並行数制限
	select {
	case pe.semaphore <- struct{}{}:
		defer func() { <-pe.semaphore }()
	case <-pe.ctx.Done():
		job.Status = JobCancelled
		job.Error = pe.ctx.Err()
		pe.sendResult(job, resultChan)
		return
	}

	// レート制限
	if err := pe.rateLimiter.Wait(pe.ctx); err != nil {
		job.Status = JobCancelled
		job.Error = fmt.Errorf("rate limit wait failed: %w", err)
		pe.sendResult(job, resultChan)
		return
	}

	// ジョブ実行開始
	job.Status = JobRunning
	job.StartTime = time.Now()

	if pe.config.Debug {
		fmt.Fprintf(os.Stderr, color.CyanString("[DEBUG] Starting job %s: %s\n"), job.ID, job.Command)
	}

	// コマンド実行
	result, err := pe.executor.ExecuteCommand(job.Command)
	job.EndTime = time.Now()
	job.Duration = job.EndTime.Sub(job.StartTime)

	if err != nil {
		job.Status = JobFailed
		job.Error = err
		pe.handleJobError(job)
	} else {
		job.Status = JobCompleted
		job.Result = result
	}

	if pe.config.Debug {
		fmt.Fprintf(os.Stderr, color.BlueString("[DEBUG] Completed job %s: status=%s, duration=%v\n"),
			job.ID, job.Status.String(), job.Duration)
	}

	pe.sendResult(job, resultChan)
}

// sendResult は結果を送信し、プログレスを更新する
func (pe *ParallelExecutor) sendResult(job *Job, resultChan chan<- *Job) {
	// ジョブ状態を更新
	pe.mu.Lock()
	pe.jobs[job.ID] = job
	pe.mu.Unlock()

	// プログレス更新
	if pe.config.ShowProgress {
		pe.monitor.Update(job)
	}

	// 結果送信
	select {
	case resultChan <- job:
	case <-pe.ctx.Done():
		return
	}
}

// handleJobError はジョブエラーを処理する
func (pe *ParallelExecutor) handleJobError(job *Job) {
	switch pe.config.ErrorPolicy {
	case StopOnError:
		if pe.config.Debug {
			fmt.Fprintf(os.Stderr, color.RedString("[ERROR] Stopping all jobs due to error in job %s: %v\n"),
				job.ID, job.Error)
		}
		pe.cancel() // すべてのジョブをキャンセル
	case StopOnCriticalError:
		if pe.isCriticalError(job.Error) {
			if pe.config.Debug {
				fmt.Fprintf(os.Stderr, color.RedString("[CRITICAL] Stopping all jobs due to critical error in job %s: %v\n"),
					job.ID, job.Error)
			}
			pe.cancel()
		}
	case ContinueOnError:
		// 続行（何もしない）
		if pe.config.Debug {
			fmt.Fprintf(os.Stderr, color.YellowString("[WARN] Job %s failed, continuing: %v\n"),
				job.ID, job.Error)
		}
	}
}

// isCriticalError はエラーが重大かどうかを判定する
func (pe *ParallelExecutor) isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 重大なエラーパターン
	criticalPatterns := []string{
		"authentication",
		"unauthorized",
		"forbidden",
		"api key",
		"token",
		"quota exceeded",
		"service unavailable",
	}

	for _, pattern := range criticalPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains は大文字小文字を区別せずに文字列に部分文字列が含まれるかチェックする
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Cancel はすべてのジョブをキャンセルする
func (pe *ParallelExecutor) Cancel() {
	pe.cancel()
}

// GetJobStatus は指定されたジョブのステータスを取得する
func (pe *ParallelExecutor) GetJobStatus(jobID string) (*Job, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	job, exists := pe.jobs[jobID]
	return job, exists
}

// GetAllJobs はすべてのジョブを取得する
func (pe *ParallelExecutor) GetAllJobs() map[string]*Job {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	jobs := make(map[string]*Job)
	for id, job := range pe.jobs {
		jobs[id] = job
	}
	return jobs
}

// GetStatistics は実行統計を取得する
func (pe *ParallelExecutor) GetStatistics() *ParallelStatistics {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	stats := &ParallelStatistics{
		TotalJobs:       len(pe.jobs),
		CompletedJobs:   0,
		FailedJobs:      0,
		CancelledJobs:   0,
		RunningJobs:     0,
		PendingJobs:     0,
		TotalDuration:   0,
		AverageDuration: 0,
	}

	var totalDuration time.Duration
	var completedCount int

	for _, job := range pe.jobs {
		switch job.Status {
		case JobCompleted:
			stats.CompletedJobs++
			completedCount++
			totalDuration += job.Duration
		case JobFailed:
			stats.FailedJobs++
		case JobCancelled:
			stats.CancelledJobs++
		case JobRunning:
			stats.RunningJobs++
		case JobPending:
			stats.PendingJobs++
		}
	}

	stats.TotalDuration = totalDuration
	if completedCount > 0 {
		stats.AverageDuration = totalDuration / time.Duration(completedCount)
	}

	return stats
}

// ParallelStatistics は並行実行の統計情報を保持する
type ParallelStatistics struct {
	TotalJobs       int           `json:"total_jobs"`
	CompletedJobs   int           `json:"completed_jobs"`
	FailedJobs      int           `json:"failed_jobs"`
	CancelledJobs   int           `json:"cancelled_jobs"`
	RunningJobs     int           `json:"running_jobs"`
	PendingJobs     int           `json:"pending_jobs"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
}

// ProgressMonitor は進捗表示を管理する
type ProgressMonitor struct {
	total      int
	completed  int
	failed     int
	cancelled  int
	mu         sync.RWMutex
	startTime  time.Time
	lastUpdate time.Time
}

// NewProgressMonitor は新しいProgressMonitorを作成する
func NewProgressMonitor() *ProgressMonitor {
	return &ProgressMonitor{}
}

// Start は進捗表示を開始する
func (pm *ProgressMonitor) Start(total int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.total = total
	pm.completed = 0
	pm.failed = 0
	pm.cancelled = 0
	pm.startTime = time.Now()
	pm.lastUpdate = pm.startTime

	fmt.Fprintf(os.Stderr, "\n%s\n", color.HiWhiteString("🚀 並行実行開始"))
	pm.displayProgress()
}

// Update は進捗を更新する
func (pm *ProgressMonitor) Update(job *Job) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch job.Status {
	case JobCompleted:
		pm.completed++
	case JobFailed:
		pm.failed++
	case JobCancelled:
		pm.cancelled++
	}

	// 表示更新（1秒に1回程度）
	now := time.Now()
	if now.Sub(pm.lastUpdate) >= time.Second {
		pm.displayProgress()
		pm.lastUpdate = now
	}
}

// displayProgress は進捗を表示する
func (pm *ProgressMonitor) displayProgress() {
	if pm.total == 0 {
		return
	}

	processed := pm.completed + pm.failed + pm.cancelled
	percentage := float64(processed) / float64(pm.total) * 100
	elapsed := time.Since(pm.startTime)

	var eta time.Duration
	if processed > 0 && processed < pm.total {
		avgTime := elapsed / time.Duration(processed)
		remaining := pm.total - processed
		eta = avgTime * time.Duration(remaining)
	}

	// プログレスバー
	barWidth := 30
	filledWidth := int(percentage * float64(barWidth) / 100)
	emptyWidth := barWidth - filledWidth

	progressBar := color.GreenString(string(make([]byte, filledWidth))) +
		color.WhiteString(string(make([]byte, emptyWidth)))

	status := fmt.Sprintf("\r[%s] %.1f%% (%d/%d) ",
		progressBar, percentage, processed, pm.total)

	if pm.completed > 0 {
		status += color.GreenString("✓%d ", pm.completed)
	}
	if pm.failed > 0 {
		status += color.RedString("✗%d ", pm.failed)
	}
	if pm.cancelled > 0 {
		status += color.YellowString("⚠%d ", pm.cancelled)
	}

	status += fmt.Sprintf("経過:%v", elapsed.Truncate(time.Second))

	if eta > 0 {
		status += fmt.Sprintf(" 残り:%v", eta.Truncate(time.Second))
	}

	fmt.Fprint(os.Stderr, status)
}

// Finish は進捗表示を終了する
func (pm *ProgressMonitor) Finish() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	elapsed := time.Since(pm.startTime)

	fmt.Fprintf(os.Stderr, "\n\n%s\n", color.HiWhiteString("✅ 並行実行完了"))
	fmt.Fprintf(os.Stderr, "総ジョブ数:   %d\n", pm.total)
	fmt.Fprintf(os.Stderr, "成功:         %s\n", color.GreenString("%d", pm.completed))
	fmt.Fprintf(os.Stderr, "失敗:         %s\n", color.RedString("%d", pm.failed))
	fmt.Fprintf(os.Stderr, "キャンセル:   %s\n", color.YellowString("%d", pm.cancelled))
	fmt.Fprintf(os.Stderr, "実行時間:     %v\n", elapsed.Truncate(time.Second))

	if pm.completed > 0 {
		avgTime := elapsed / time.Duration(pm.completed)
		fmt.Fprintf(os.Stderr, "平均実行時間: %v\n", avgTime.Truncate(time.Millisecond))
	}

	fmt.Fprintln(os.Stderr)
}
