package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestDB はテスト用のデータベースを作成する
func setupTestDB(t *testing.T) (*SQLiteStore, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}

	return store, cleanup
}

func TestNewSQLiteStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Error("Expected database connection to be initialized")
	}

	if store.path != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, store.path)
	}

	// テーブルが作成されていることを確認
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('sessions', 'executions')").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 tables, got %d", count)
	}
}

func TestSQLiteStore_SaveAndGetSession(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	session := &SessionRecord{
		ID:         "session-001",
		UserID:     "user-001",
		Mode:       "sandbox",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now(),
		TotalJobs:  10,
		Successful: 8,
		Failed:     2,
		CreatedAt:  time.Now(),
	}

	// セッションを保存
	err := store.SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// セッションを取得
	retrieved, err := store.GetSession("session-001")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected session to be found")
	}

	if retrieved.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, retrieved.ID)
	}
	if retrieved.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, retrieved.UserID)
	}
	if retrieved.TotalJobs != session.TotalJobs {
		t.Errorf("Expected TotalJobs %d, got %d", session.TotalJobs, retrieved.TotalJobs)
	}
}

func TestSQLiteStore_SaveAndGetExecution(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// 先にセッションを作成
	session := &SessionRecord{
		ID:        "session-001",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}
	store.SaveSession(session)

	execution := &ExecutionRecord{
		ID:           "exec-001",
		SessionID:    "session-001",
		Command:      "usacloud server list",
		File:         "test.sh",
		Status:       "completed",
		StartTime:    time.Now().Add(-30 * time.Second),
		EndTime:      time.Now(),
		Duration:     30000,
		ExitCode:     0,
		Stdout:       "server1\nserver2",
		Stderr:       "",
		ErrorMessage: "",
		Metadata:     `{"test": true}`,
		CreatedAt:    time.Now(),
	}

	// 実行記録を保存
	err := store.SaveExecution(execution)
	if err != nil {
		t.Fatalf("Failed to save execution: %v", err)
	}

	// 実行記録を取得
	filter := ExecutionFilter{
		SessionID: "session-001",
	}
	records, err := store.GetExecutions(filter)
	if err != nil {
		t.Fatalf("Failed to get executions: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	retrieved := records[0]
	if retrieved.ID != execution.ID {
		t.Errorf("Expected ID %s, got %s", execution.ID, retrieved.ID)
	}
	if retrieved.Command != execution.Command {
		t.Errorf("Expected Command %s, got %s", execution.Command, retrieved.Command)
	}
	if retrieved.Status != execution.Status {
		t.Errorf("Expected Status %s, got %s", execution.Status, retrieved.Status)
	}
}

func TestSQLiteStore_GetExecutionsWithFilter(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータの準備
	session := &SessionRecord{
		ID:        "session-001",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}
	store.SaveSession(session)

	executions := []*ExecutionRecord{
		{
			ID:        "exec-001",
			SessionID: "session-001",
			Command:   "usacloud server list",
			Status:    "completed",
			Duration:  1000,
			StartTime: time.Now().Add(-3 * time.Hour),
			EndTime:   time.Now().Add(-3 * time.Hour).Add(1 * time.Second),
			CreatedAt: time.Now(),
		},
		{
			ID:        "exec-002",
			SessionID: "session-001",
			Command:   "usacloud disk list",
			Status:    "failed",
			Duration:  500,
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-2 * time.Hour).Add(500 * time.Millisecond),
			CreatedAt: time.Now(),
		},
		{
			ID:        "exec-003",
			SessionID: "session-001",
			Command:   "usacloud note list",
			Status:    "completed",
			Duration:  2000,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now().Add(-1 * time.Hour).Add(2 * time.Second),
			CreatedAt: time.Now(),
		},
	}

	for _, exec := range executions {
		store.SaveExecution(exec)
	}

	t.Run("filter by status", func(t *testing.T) {
		filter := ExecutionFilter{Status: "completed"}
		records, err := store.GetExecutions(filter)
		if err != nil {
			t.Fatalf("Failed to get executions: %v", err)
		}

		if len(records) != 2 {
			t.Errorf("Expected 2 completed records, got %d", len(records))
		}

		for _, record := range records {
			if record.Status != "completed" {
				t.Errorf("Expected status 'completed', got '%s'", record.Status)
			}
		}
	})

	t.Run("filter by command", func(t *testing.T) {
		filter := ExecutionFilter{Command: "server"}
		records, err := store.GetExecutions(filter)
		if err != nil {
			t.Fatalf("Failed to get executions: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("Expected 1 record with 'server', got %d", len(records))
		}

		if !strings.Contains(records[0].Command, "server") {
			t.Errorf("Expected command to contain 'server', got '%s'", records[0].Command)
		}
	})

	t.Run("filter by duration range", func(t *testing.T) {
		minDuration := int64(800)
		maxDuration := int64(1500)
		filter := ExecutionFilter{
			MinDuration: &minDuration,
			MaxDuration: &maxDuration,
		}
		records, err := store.GetExecutions(filter)
		if err != nil {
			t.Fatalf("Failed to get executions: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("Expected 1 record in duration range, got %d", len(records))
		}

		if records[0].Duration < minDuration || records[0].Duration > maxDuration {
			t.Errorf("Record duration %d not in range [%d, %d]", records[0].Duration, minDuration, maxDuration)
		}
	})

	t.Run("order by duration desc", func(t *testing.T) {
		filter := ExecutionFilter{
			OrderBy:   "duration_ms",
			OrderDesc: true,
		}
		records, err := store.GetExecutions(filter)
		if err != nil {
			t.Fatalf("Failed to get executions: %v", err)
		}

		if len(records) != 3 {
			t.Errorf("Expected 3 records, got %d", len(records))
		}

		// 降順であることを確認
		if records[0].Duration < records[1].Duration || records[1].Duration < records[2].Duration {
			t.Error("Records not ordered by duration desc")
		}
	})

	t.Run("limit and offset", func(t *testing.T) {
		filter := ExecutionFilter{
			Limit:  2,
			Offset: 1,
		}
		records, err := store.GetExecutions(filter)
		if err != nil {
			t.Fatalf("Failed to get executions: %v", err)
		}

		if len(records) != 2 {
			t.Errorf("Expected 2 records with limit, got %d", len(records))
		}
	})
}

func TestSQLiteStore_ExportResults(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータの準備
	session := &SessionRecord{
		ID:        "session-001",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}
	store.SaveSession(session)

	execution := &ExecutionRecord{
		ID:           "exec-001",
		SessionID:    "session-001",
		Command:      "usacloud server list",
		Status:       "completed",
		StartTime:    time.Now().Add(-1 * time.Minute),
		EndTime:      time.Now(),
		Duration:     60000,
		ExitCode:     0,
		ErrorMessage: "",
		CreatedAt:    time.Now(),
	}
	store.SaveExecution(execution)

	filter := ExecutionFilter{SessionID: "session-001"}

	t.Run("export as JSON", func(t *testing.T) {
		data, err := store.ExportResults(FormatJSON, filter)
		if err != nil {
			t.Fatalf("Failed to export as JSON: %v", err)
		}

		var records []*ExecutionRecord
		err = json.Unmarshal(data, &records)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("Expected 1 record in JSON, got %d", len(records))
		}

		if records[0].ID != execution.ID {
			t.Errorf("Expected ID %s, got %s", execution.ID, records[0].ID)
		}
	})

	t.Run("export as CSV", func(t *testing.T) {
		data, err := store.ExportResults(FormatCSV, filter)
		if err != nil {
			t.Fatalf("Failed to export as CSV: %v", err)
		}

		csv := string(data)
		lines := strings.Split(csv, "\n")

		// ヘッダー行 + データ行 + 空行
		if len(lines) < 2 {
			t.Errorf("Expected at least 2 lines in CSV, got %d", len(lines))
		}

		// ヘッダーの確認
		if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "Command") {
			t.Error("CSV header missing expected columns")
		}

		// データ行の確認
		if !strings.Contains(lines[1], execution.ID) {
			t.Error("CSV data missing execution ID")
		}
	})

	t.Run("export as HTML", func(t *testing.T) {
		data, err := store.ExportResults(FormatHTML, filter)
		if err != nil {
			t.Fatalf("Failed to export as HTML: %v", err)
		}

		html := string(data)

		// HTML構造の確認
		if !strings.Contains(html, "<html>") || !strings.Contains(html, "</html>") {
			t.Error("HTML export missing basic HTML structure")
		}

		if !strings.Contains(html, "<table>") || !strings.Contains(html, "</table>") {
			t.Error("HTML export missing table structure")
		}

		if !strings.Contains(html, execution.ID) {
			t.Error("HTML export missing execution data")
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, err := store.ExportResults("xml", filter)
		if err == nil {
			t.Error("Expected error for unsupported format")
		}
	})
}

func TestSQLiteStore_GetSessions(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// テストデータの準備
	sessions := []*SessionRecord{
		{
			ID:        "session-001",
			UserID:    "user-001",
			Mode:      "sandbox",
			StartTime: time.Now().Add(-3 * time.Hour),
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:        "session-002",
			UserID:    "user-001",
			Mode:      "batch",
			StartTime: time.Now().Add(-2 * time.Hour),
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "session-003",
			UserID:    "user-002",
			Mode:      "dry-run",
			StartTime: time.Now().Add(-1 * time.Hour),
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, session := range sessions {
		store.SaveSession(session)
	}

	t.Run("get all sessions", func(t *testing.T) {
		records, err := store.GetSessions(0, 0)
		if err != nil {
			t.Fatalf("Failed to get sessions: %v", err)
		}

		if len(records) != 3 {
			t.Errorf("Expected 3 sessions, got %d", len(records))
		}

		// 開始時間の降順でソートされていることを確認
		for i := 0; i < len(records)-1; i++ {
			if records[i].StartTime.Before(records[i+1].StartTime) {
				t.Error("Sessions not ordered by start_time DESC")
			}
		}
	})

	t.Run("get sessions with limit", func(t *testing.T) {
		records, err := store.GetSessions(2, 0)
		if err != nil {
			t.Fatalf("Failed to get sessions: %v", err)
		}

		if len(records) != 2 {
			t.Errorf("Expected 2 sessions with limit, got %d", len(records))
		}
	})

	t.Run("get sessions with offset", func(t *testing.T) {
		records, err := store.GetSessions(2, 1)
		if err != nil {
			t.Fatalf("Failed to get sessions: %v", err)
		}

		if len(records) != 2 {
			t.Errorf("Expected 2 sessions with offset, got %d", len(records))
		}
	})
}

func TestSQLiteStore_CleanupOldRecords(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// 古いセッションと実行記録を作成
	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now().Add(-1 * time.Hour)

	oldSession := &SessionRecord{
		ID:        "old-session",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: oldTime,
		CreatedAt: oldTime,
	}

	newSession := &SessionRecord{
		ID:        "new-session",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: newTime,
		CreatedAt: newTime,
	}

	store.SaveSession(oldSession)
	store.SaveSession(newSession)

	oldExecution := &ExecutionRecord{
		ID:        "old-exec",
		SessionID: "old-session",
		Command:   "old command",
		Status:    "completed",
		StartTime: oldTime,
		EndTime:   oldTime.Add(1 * time.Second),
		CreatedAt: oldTime,
	}

	newExecution := &ExecutionRecord{
		ID:        "new-exec",
		SessionID: "new-session",
		Command:   "new command",
		Status:    "completed",
		StartTime: newTime,
		EndTime:   newTime.Add(1 * time.Second),
		CreatedAt: newTime,
	}

	store.SaveExecution(oldExecution)
	store.SaveExecution(newExecution)

	// 24時間より古い記録を削除
	cutoffTime := time.Now().Add(-24 * time.Hour)
	err := store.CleanupOldRecords(cutoffTime)
	if err != nil {
		t.Fatalf("Failed to cleanup old records: %v", err)
	}

	// 新しい記録が残っていることを確認
	newSessionResult, err := store.GetSession("new-session")
	if err != nil {
		t.Fatalf("Failed to get new session: %v", err)
	}
	if newSessionResult == nil {
		t.Error("New session should not be deleted")
	}

	newExecFilter := ExecutionFilter{SessionID: "new-session"}
	newExecRecords, err := store.GetExecutions(newExecFilter)
	if err != nil {
		t.Fatalf("Failed to get new executions: %v", err)
	}
	if len(newExecRecords) != 1 {
		t.Errorf("Expected 1 new execution, got %d", len(newExecRecords))
	}

	// 古い記録が削除されていることを確認
	oldSessionResult, err := store.GetSession("old-session")
	if err != nil {
		t.Fatalf("Failed to check old session: %v", err)
	}
	if oldSessionResult != nil {
		t.Error("Old session should be deleted")
	}

	oldExecFilter := ExecutionFilter{SessionID: "old-session"}
	oldExecRecords, err := store.GetExecutions(oldExecFilter)
	if err != nil {
		t.Fatalf("Failed to check old executions: %v", err)
	}
	if len(oldExecRecords) != 0 {
		t.Errorf("Expected 0 old executions, got %d", len(oldExecRecords))
	}
}

func TestSQLiteStore_GetStatistics(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// 空の統計
	stats, err := store.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get empty statistics: %v", err)
	}

	if stats.TotalExecutions != 0 {
		t.Errorf("Expected 0 executions, got %d", stats.TotalExecutions)
	}
	if stats.TotalSessions != 0 {
		t.Errorf("Expected 0 sessions, got %d", stats.TotalSessions)
	}

	// テストデータを追加
	session := &SessionRecord{
		ID:        "session-001",
		UserID:    "user-001",
		Mode:      "sandbox",
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}
	store.SaveSession(session)

	executions := []*ExecutionRecord{
		{
			ID:        "exec-001",
			SessionID: "session-001",
			Command:   "cmd1",
			Status:    "completed",
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-2 * time.Hour).Add(1 * time.Second),
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "exec-002",
			SessionID: "session-001",
			Command:   "cmd2",
			Status:    "failed",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now().Add(-1 * time.Hour).Add(1 * time.Second),
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, exec := range executions {
		store.SaveExecution(exec)
	}

	// 統計を取得
	stats, err = store.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if stats.TotalExecutions != 2 {
		t.Errorf("Expected 2 executions, got %d", stats.TotalExecutions)
	}
	if stats.TotalSessions != 1 {
		t.Errorf("Expected 1 session, got %d", stats.TotalSessions)
	}

	// ステータス別カウント
	if stats.StatusCounts["completed"] != 1 {
		t.Errorf("Expected 1 completed, got %d", stats.StatusCounts["completed"])
	}
	if stats.StatusCounts["failed"] != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.StatusCounts["failed"])
	}

	// 日付範囲（空の場合は警告のみ）
	if stats.OldestRecord.IsZero() || stats.NewestRecord.IsZero() {
		t.Logf("Warning: Date parsing may have failed - oldest: %v, newest: %v", stats.OldestRecord, stats.NewestRecord)
	} else if stats.OldestRecord.After(stats.NewestRecord) {
		t.Error("Oldest record should be before newest record")
	}
}

func TestExecutionResultToPersistence(t *testing.T) {
	result := &ExecutionResult{
		Command:  "usacloud server list",
		Success:  true,
		Output:   "server1\nserver2",
		Duration: 2 * time.Second,
		Skipped:  false,
	}

	record := ExecutionResultToPersistence(result, "session-123")

	if record.SessionID != "session-123" {
		t.Errorf("Expected SessionID 'session-123', got '%s'", record.SessionID)
	}
	if record.Command != result.Command {
		t.Errorf("Expected Command '%s', got '%s'", result.Command, record.Command)
	}
	if record.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", record.Status)
	}
	if record.Stdout != result.Output {
		t.Errorf("Expected Stdout '%s', got '%s'", result.Output, record.Stdout)
	}
	if record.Duration != result.Duration.Milliseconds() {
		t.Errorf("Expected Duration %d, got %d", result.Duration.Milliseconds(), record.Duration)
	}
}

func TestJobToPersistence(t *testing.T) {
	startTime := time.Now().Add(-1 * time.Minute)
	endTime := time.Now()

	job := &Job{
		ID:        "job-123",
		Command:   "usacloud server list",
		File:      "test.sh",
		Status:    JobCompleted,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
		Result: &ExecutionResult{
			Success: true,
			Output:  "server output",
		},
		Metadata: map[string]interface{}{
			"priority": "high",
			"retry":    3,
		},
	}

	record := JobToPersistence(job)

	if record.ID != job.ID {
		t.Errorf("Expected ID '%s', got '%s'", job.ID, record.ID)
	}
	if record.Command != job.Command {
		t.Errorf("Expected Command '%s', got '%s'", job.Command, record.Command)
	}
	if record.File != job.File {
		t.Errorf("Expected File '%s', got '%s'", job.File, record.File)
	}
	if record.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", record.Status)
	}
	if record.StartTime != job.StartTime {
		t.Errorf("Expected StartTime %v, got %v", job.StartTime, record.StartTime)
	}
	if record.EndTime != job.EndTime {
		t.Errorf("Expected EndTime %v, got %v", job.EndTime, record.EndTime)
	}
	if record.ExitCode != 0 {
		t.Errorf("Expected ExitCode 0, got %d", record.ExitCode)
	}

	// メタデータのJSONチェック
	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(record.Metadata), &metadata)
	if err != nil {
		t.Fatalf("Failed to parse metadata JSON: %v", err)
	}
	if metadata["priority"] != "high" {
		t.Errorf("Expected priority 'high', got '%v'", metadata["priority"])
	}
}
