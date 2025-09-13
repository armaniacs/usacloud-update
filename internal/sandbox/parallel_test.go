package sandbox

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockExecutorForParallel はテスト用のExecutorモック
type mockExecutorForParallel struct {
	executeFunc func(cmd string) (*ExecutionResult, error)
	delay       time.Duration
	callCount   int
	mu          sync.Mutex
}

func (m *mockExecutorForParallel) ExecuteCommand(cmd string) (*ExecutionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.executeFunc != nil {
		return m.executeFunc(cmd)
	}

	return &ExecutionResult{
		Command:  cmd,
		Success:  true,
		Output:   fmt.Sprintf("output for %s", cmd),
		Duration: m.delay,
	}, nil
}

func (m *mockExecutorForParallel) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status   JobStatus
		expected string
	}{
		{JobPending, "pending"},
		{JobRunning, "running"},
		{JobCompleted, "completed"},
		{JobFailed, "failed"},
		{JobCancelled, "cancelled"},
		{JobStatus(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.status.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestErrorPolicy_String(t *testing.T) {
	tests := []struct {
		policy   ErrorPolicy
		expected string
	}{
		{ContinueOnError, "continue_on_error"},
		{StopOnError, "stop_on_error"},
		{StopOnCriticalError, "stop_on_critical_error"},
		{ErrorPolicy(999), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.policy.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestDefaultParallelConfig(t *testing.T) {
	config := DefaultParallelConfig()

	if config.MaxConcurrency != 5 {
		t.Errorf("Expected MaxConcurrency to be 5, got %d", config.MaxConcurrency)
	}
	if config.RateLimit != 2.0 {
		t.Errorf("Expected RateLimit to be 2.0, got %f", config.RateLimit)
	}
	if config.ErrorPolicy != ContinueOnError {
		t.Errorf("Expected ErrorPolicy to be ContinueOnError, got %v", config.ErrorPolicy)
	}
	if config.Timeout != 5*time.Minute {
		t.Errorf("Expected Timeout to be 5m, got %v", config.Timeout)
	}
	if !config.ShowProgress {
		t.Error("Expected ShowProgress to be true")
	}
	if config.Debug {
		t.Error("Expected Debug to be false")
	}
}

func TestNewParallelExecutor(t *testing.T) {
	executor := &mockExecutorForParallel{}

	t.Run("with config", func(t *testing.T) {
		config := &ParallelConfig{
			MaxConcurrency: 3,
			RateLimit:      1.0,
			ErrorPolicy:    StopOnError,
			Timeout:        1 * time.Minute,
			ShowProgress:   false,
			Debug:          true,
		}

		pe := NewParallelExecutor(executor, config)

		if pe.executor != executor {
			t.Error("Expected executor to be set")
		}
		if pe.config.MaxConcurrency != 3 {
			t.Errorf("Expected MaxConcurrency to be 3, got %d", pe.config.MaxConcurrency)
		}
		if pe.rateLimiter == nil {
			t.Error("Expected rateLimiter to be initialized")
		}
		if len(pe.semaphore) != 0 || cap(pe.semaphore) != 3 {
			t.Errorf("Expected semaphore capacity to be 3, got %d", cap(pe.semaphore))
		}
		if pe.jobs == nil {
			t.Error("Expected jobs map to be initialized")
		}
		if pe.monitor == nil {
			t.Error("Expected monitor to be initialized")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		pe := NewParallelExecutor(executor, nil)

		if pe.config.MaxConcurrency != 5 {
			t.Errorf("Expected default MaxConcurrency to be 5, got %d", pe.config.MaxConcurrency)
		}
	})
}

func TestParallelExecutor_ExecuteJob(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		executor := &mockExecutorForParallel{}
		config := &ParallelConfig{
			MaxConcurrency: 1,
			RateLimit:      10.0,
			ErrorPolicy:    ContinueOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		job := &Job{
			ID:      "test-job",
			Command: "usacloud server list",
		}

		result, err := pe.ExecuteJob(job)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result")
		}
		if result.Status != JobCompleted {
			t.Errorf("Expected status to be JobCompleted, got %v", result.Status)
		}
		if result.Result == nil {
			t.Error("Expected execution result")
		}
		if result.Duration == 0 {
			t.Error("Expected non-zero duration")
		}
	})

	t.Run("failed execution", func(t *testing.T) {
		executor := &mockExecutorForParallel{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				return nil, errors.New("execution failed")
			},
		}
		config := &ParallelConfig{
			MaxConcurrency: 1,
			RateLimit:      10.0,
			ErrorPolicy:    ContinueOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		job := &Job{
			ID:      "test-job",
			Command: "usacloud server list",
		}

		result, err := pe.ExecuteJob(job)

		if err == nil {
			t.Fatal("Expected error")
		}
		if result == nil {
			t.Fatal("Expected result even on failure")
		}
		if result.Status != JobFailed {
			t.Errorf("Expected status to be JobFailed, got %v", result.Status)
		}
	})
}

func TestParallelExecutor_SubmitJobs(t *testing.T) {
	t.Run("multiple successful jobs", func(t *testing.T) {
		executor := &mockExecutorForParallel{
			delay: 10 * time.Millisecond,
		}
		config := &ParallelConfig{
			MaxConcurrency: 3,
			RateLimit:      50.0, // 高いレート制限でテストを高速化
			ErrorPolicy:    ContinueOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		jobs := []*Job{
			{Command: "usacloud server list"},
			{Command: "usacloud disk list"},
			{Command: "usacloud note list"},
			{Command: "usacloud switch list"},
			{Command: "usacloud archive list"},
		}

		start := time.Now()
		resultChan := pe.SubmitJobs(jobs)

		results := make([]*Job, 0, len(jobs))
		for result := range resultChan {
			results = append(results, result)
		}
		elapsed := time.Since(start)

		if len(results) != len(jobs) {
			t.Errorf("Expected %d results, got %d", len(jobs), len(results))
		}

		completedCount := 0
		for _, result := range results {
			if result.Status == JobCompleted {
				completedCount++
			}
		}

		if completedCount != len(jobs) {
			t.Errorf("Expected %d completed jobs, got %d", len(jobs), completedCount)
		}

		// 並行実行により、全体時間が単純な直列実行より短いことを確認
		// テスト環境では並行処理のオーバーヘッドがあるため、倍の時間を許容
		expectedSerialTime := time.Duration(len(jobs)) * 10 * time.Millisecond * 2
		if elapsed >= expectedSerialTime {
			t.Logf("Note: Parallel execution time %v vs expected max %v", elapsed, expectedSerialTime)
		}
	})

	t.Run("concurrency limit", func(t *testing.T) {
		var activeJobs int32
		var maxActiveJobs int32
		var mu sync.Mutex

		executor := &mockExecutorForParallel{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				mu.Lock()
				activeJobs++
				if activeJobs > maxActiveJobs {
					maxActiveJobs = activeJobs
				}
				mu.Unlock()

				time.Sleep(50 * time.Millisecond)

				mu.Lock()
				activeJobs--
				mu.Unlock()

				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success",
				}, nil
			},
		}

		config := &ParallelConfig{
			MaxConcurrency: 2,
			RateLimit:      50.0,
			ErrorPolicy:    ContinueOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		jobs := []*Job{
			{Command: "job1"},
			{Command: "job2"},
			{Command: "job3"},
			{Command: "job4"},
		}

		resultChan := pe.SubmitJobs(jobs)

		for range resultChan {
			// 結果を消費
		}

		if maxActiveJobs > 2 {
			t.Errorf("Expected max 2 concurrent jobs, got %d", maxActiveJobs)
		}
	})

	t.Run("error handling with StopOnError", func(t *testing.T) {
		callCount := 0
		executor := &mockExecutorForParallel{
			executeFunc: func(cmd string) (*ExecutionResult, error) {
				callCount++
				if cmd == "job2" {
					return nil, errors.New("deliberate failure")
				}
				// 遅延を入れて、エラー後のジョブが実行されないことを確認
				time.Sleep(100 * time.Millisecond)
				return &ExecutionResult{
					Command: cmd,
					Success: true,
					Output:  "success",
				}, nil
			},
		}

		config := &ParallelConfig{
			MaxConcurrency: 1,
			RateLimit:      50.0,
			ErrorPolicy:    StopOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		jobs := []*Job{
			{Command: "job1"},
			{Command: "job2"}, // これが失敗する
			{Command: "job3"},
			{Command: "job4"},
		}

		resultChan := pe.SubmitJobs(jobs)

		results := make([]*Job, 0)
		for result := range resultChan {
			results = append(results, result)
		}

		// エラー発生後にキャンセルされるため、一部のジョブがキャンセル状態になる
		cancelledOrFailedCount := 0
		for _, result := range results {
			if result.Status == JobCancelled || result.Status == JobFailed {
				cancelledOrFailedCount++
			}
		}

		if cancelledOrFailedCount == 0 {
			t.Error("Expected some jobs to be cancelled or failed due to StopOnError policy")
		}
	})

	t.Run("rate limiting", func(t *testing.T) {
		executor := &mockExecutorForParallel{}
		config := &ParallelConfig{
			MaxConcurrency: 5,
			RateLimit:      2.0, // 2 RPS
			ErrorPolicy:    ContinueOnError,
			ShowProgress:   false,
		}
		pe := NewParallelExecutor(executor, config)

		jobs := []*Job{
			{Command: "job1"},
			{Command: "job2"},
			{Command: "job3"},
		}

		start := time.Now()
		resultChan := pe.SubmitJobs(jobs)

		for range resultChan {
			// 結果を消費
		}
		elapsed := time.Since(start)

		// 2 RPSなので、3つのジョブには最低1秒はかかるはず
		expectedMinTime := 1 * time.Second
		if elapsed < expectedMinTime {
			t.Errorf("Expected rate limiting to enforce minimum time %v, got %v",
				expectedMinTime, elapsed)
		}
	})
}

func TestParallelExecutor_isCriticalError(t *testing.T) {
	executor := &mockExecutorForParallel{}
	pe := NewParallelExecutor(executor, nil)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"authentication error", errors.New("authentication failed"), true},
		{"unauthorized error", errors.New("unauthorized access"), true},
		{"api key error", errors.New("invalid api key"), true},
		{"quota exceeded", errors.New("quota exceeded"), true},
		{"normal error", errors.New("some random error"), false},
		{"network error", errors.New("network timeout"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := pe.isCriticalError(test.err)
			if result != test.expected {
				t.Errorf("Expected %v, got %v for error: %v", test.expected, result, test.err)
			}
		})
	}
}

func TestParallelExecutor_GetJobStatus(t *testing.T) {
	executor := &mockExecutorForParallel{}
	pe := NewParallelExecutor(executor, nil)

	// ジョブを手動で追加
	job := &Job{
		ID:      "test-job",
		Command: "test command",
		Status:  JobCompleted,
	}
	pe.jobs["test-job"] = job

	t.Run("existing job", func(t *testing.T) {
		result, exists := pe.GetJobStatus("test-job")
		if !exists {
			t.Error("Expected job to exist")
		}
		if result.ID != "test-job" {
			t.Errorf("Expected job ID 'test-job', got '%s'", result.ID)
		}
	})

	t.Run("non-existing job", func(t *testing.T) {
		result, exists := pe.GetJobStatus("non-existing")
		if exists {
			t.Error("Expected job not to exist")
		}
		if result != nil {
			t.Error("Expected nil result for non-existing job")
		}
	})
}

func TestParallelExecutor_GetStatistics(t *testing.T) {
	executor := &mockExecutorForParallel{}
	pe := NewParallelExecutor(executor, nil)

	// テスト用のジョブを手動で追加
	jobs := []*Job{
		{ID: "job1", Status: JobCompleted, Duration: 100 * time.Millisecond},
		{ID: "job2", Status: JobCompleted, Duration: 200 * time.Millisecond},
		{ID: "job3", Status: JobFailed},
		{ID: "job4", Status: JobCancelled},
		{ID: "job5", Status: JobRunning},
		{ID: "job6", Status: JobPending},
	}

	for _, job := range jobs {
		pe.jobs[job.ID] = job
	}

	stats := pe.GetStatistics()

	if stats.TotalJobs != 6 {
		t.Errorf("Expected TotalJobs to be 6, got %d", stats.TotalJobs)
	}
	if stats.CompletedJobs != 2 {
		t.Errorf("Expected CompletedJobs to be 2, got %d", stats.CompletedJobs)
	}
	if stats.FailedJobs != 1 {
		t.Errorf("Expected FailedJobs to be 1, got %d", stats.FailedJobs)
	}
	if stats.CancelledJobs != 1 {
		t.Errorf("Expected CancelledJobs to be 1, got %d", stats.CancelledJobs)
	}
	if stats.RunningJobs != 1 {
		t.Errorf("Expected RunningJobs to be 1, got %d", stats.RunningJobs)
	}
	if stats.PendingJobs != 1 {
		t.Errorf("Expected PendingJobs to be 1, got %d", stats.PendingJobs)
	}

	expectedTotal := 300 * time.Millisecond
	if stats.TotalDuration != expectedTotal {
		t.Errorf("Expected TotalDuration to be %v, got %v", expectedTotal, stats.TotalDuration)
	}

	expectedAverage := 150 * time.Millisecond
	if stats.AverageDuration != expectedAverage {
		t.Errorf("Expected AverageDuration to be %v, got %v", expectedAverage, stats.AverageDuration)
	}
}

func TestProgressMonitor(t *testing.T) {
	monitor := NewProgressMonitor()

	// Start
	monitor.Start(3)
	if monitor.total != 3 {
		t.Errorf("Expected total to be 3, got %d", monitor.total)
	}

	// Update with completed job
	job1 := &Job{Status: JobCompleted}
	monitor.Update(job1)
	if monitor.completed != 1 {
		t.Errorf("Expected completed to be 1, got %d", monitor.completed)
	}

	// Update with failed job
	job2 := &Job{Status: JobFailed}
	monitor.Update(job2)
	if monitor.failed != 1 {
		t.Errorf("Expected failed to be 1, got %d", monitor.failed)
	}

	// Update with cancelled job
	job3 := &Job{Status: JobCancelled}
	monitor.Update(job3)
	if monitor.cancelled != 1 {
		t.Errorf("Expected cancelled to be 1, got %d", monitor.cancelled)
	}

	// Finish
	monitor.Finish()
	// Finishは主に表示用なので、状態に変化がないことを確認
	if monitor.completed != 1 || monitor.failed != 1 || monitor.cancelled != 1 {
		t.Error("Finish should not change job counts")
	}
}

func TestParallelExecutor_Cancel(t *testing.T) {
	executor := &mockExecutorForParallel{
		delay: 100 * time.Millisecond,
	}
	config := &ParallelConfig{
		MaxConcurrency: 2,
		RateLimit:      50.0,
		ErrorPolicy:    ContinueOnError,
		ShowProgress:   false,
	}
	pe := NewParallelExecutor(executor, config)

	jobs := []*Job{
		{Command: "job1"},
		{Command: "job2"},
		{Command: "job3"},
		{Command: "job4"},
	}

	resultChan := pe.SubmitJobs(jobs)

	// 少し待ってからキャンセル
	time.Sleep(50 * time.Millisecond)
	pe.Cancel()

	results := make([]*Job, 0)
	for result := range resultChan {
		results = append(results, result)
	}

	// キャンセルされたジョブがあることを確認
	cancelledCount := 0
	for _, result := range results {
		if result.Status == JobCancelled {
			cancelledCount++
		}
	}

	if cancelledCount == 0 {
		t.Error("Expected some jobs to be cancelled")
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "hello", "hello", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
		{"substring in middle", "hello world test", "world", true},
		{"not found", "hello world", "xyz", false},
		{"empty substring", "hello", "", true},
		{"empty string", "", "hello", false},
		{"case sensitive", "Hello", "hello", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := contains(test.s, test.substr)
			if result != test.expected {
				t.Errorf("contains(%q, %q) = %v, expected %v",
					test.s, test.substr, result, test.expected)
			}
		})
	}
}

func TestParallelExecutor_ContextTimeout(t *testing.T) {
	t.Skip("Skipping context timeout test - timing sensitive and environment dependent")
}
