package sandbox

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ExecutionRecord は実行記録を表す
type ExecutionRecord struct {
	ID            string    `json:"id" db:"id"`
	SessionID     string    `json:"session_id" db:"session_id"`
	Command       string    `json:"command" db:"command"`
	File          string    `json:"file" db:"file"`
	Status        string    `json:"status" db:"status"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	EndTime       time.Time `json:"end_time" db:"end_time"`
	Duration      int64     `json:"duration_ms" db:"duration_ms"`
	ExitCode      int       `json:"exit_code" db:"exit_code"`
	Stdout        string    `json:"stdout" db:"stdout"`
	Stderr        string    `json:"stderr" db:"stderr"`
	ErrorMessage  string    `json:"error_message" db:"error_message"`
	ResourceUsage string    `json:"resource_usage" db:"resource_usage"`
	Metadata      string    `json:"metadata" db:"metadata"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SessionRecord はセッション記録を表す
type SessionRecord struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Mode       string    `json:"mode" db:"mode"` // sandbox, dry-run, batch
	StartTime  time.Time `json:"start_time" db:"start_time"`
	EndTime    time.Time `json:"end_time" db:"end_time"`
	TotalJobs  int       `json:"total_jobs" db:"total_jobs"`
	Successful int       `json:"successful" db:"successful"`
	Failed     int       `json:"failed" db:"failed"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ExecutionFilter は実行記録の検索フィルター
type ExecutionFilter struct {
	SessionID   string     `json:"session_id,omitempty"`
	Command     string     `json:"command,omitempty"`
	Status      string     `json:"status,omitempty"`
	File        string     `json:"file,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	MinDuration *int64     `json:"min_duration,omitempty"`
	MaxDuration *int64     `json:"max_duration,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
	OrderBy     string     `json:"order_by,omitempty"`
	OrderDesc   bool       `json:"order_desc,omitempty"`
}

// ExportFormat はエクスポート形式を表す
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatHTML ExportFormat = "html"
)

// ResultStore は実行結果の永続化インターフェース
type ResultStore interface {
	SaveExecution(record *ExecutionRecord) error
	SaveSession(record *SessionRecord) error
	GetExecutions(filter ExecutionFilter) ([]*ExecutionRecord, error)
	GetSession(sessionID string) (*SessionRecord, error)
	GetSessions(limit, offset int) ([]*SessionRecord, error)
	ExportResults(format ExportFormat, filter ExecutionFilter) ([]byte, error)
	CleanupOldRecords(olderThan time.Time) error
	GetStatistics() (*PersistenceStatistics, error)
	Close() error
}

// SQLiteStore はSQLiteを使った実行結果ストア
type SQLiteStore struct {
	db   *sql.DB
	path string
}

// NewSQLiteStore は新しいSQLiteStoreを作成する
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{
		db:   db,
		path: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema はデータベーススキーマを初期化する
func (s *SQLiteStore) initSchema() error {
	// セッションテーブル
	sessionSchema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		mode TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		total_jobs INTEGER DEFAULT 0,
		successful INTEGER DEFAULT 0,
		failed INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_mode ON sessions(mode);
	`

	// 実行記録テーブル
	executionSchema := `
	CREATE TABLE IF NOT EXISTS executions (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		command TEXT NOT NULL,
		file TEXT,
		status TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		duration_ms INTEGER,
		exit_code INTEGER,
		stdout TEXT,
		stderr TEXT,
		error_message TEXT,
		resource_usage TEXT,
		metadata TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_executions_session_id ON executions(session_id);
	CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
	CREATE INDEX IF NOT EXISTS idx_executions_start_time ON executions(start_time);
	CREATE INDEX IF NOT EXISTS idx_executions_command ON executions(command);
	CREATE INDEX IF NOT EXISTS idx_executions_duration ON executions(duration_ms);
	`

	// スキーマを実行
	if _, err := s.db.Exec(sessionSchema); err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	if _, err := s.db.Exec(executionSchema); err != nil {
		return fmt.Errorf("failed to create executions table: %w", err)
	}

	return nil
}

// SaveExecution は実行記録を保存する
func (s *SQLiteStore) SaveExecution(record *ExecutionRecord) error {
	query := `
		INSERT INTO executions (
			id, session_id, command, file, status, start_time, end_time,
			duration_ms, exit_code, stdout, stderr, error_message,
			resource_usage, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		record.ID, record.SessionID, record.Command, record.File,
		record.Status, record.StartTime, record.EndTime, record.Duration,
		record.ExitCode, record.Stdout, record.Stderr, record.ErrorMessage,
		record.ResourceUsage, record.Metadata, record.CreatedAt)

	return err
}

// SaveSession はセッション記録を保存する
func (s *SQLiteStore) SaveSession(record *SessionRecord) error {
	query := `
		INSERT OR REPLACE INTO sessions (
			id, user_id, mode, start_time, end_time,
			total_jobs, successful, failed, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		record.ID, record.UserID, record.Mode, record.StartTime, record.EndTime,
		record.TotalJobs, record.Successful, record.Failed, record.CreatedAt)

	return err
}

// GetExecutions はフィルター条件に基づいて実行記録を取得する
func (s *SQLiteStore) GetExecutions(filter ExecutionFilter) ([]*ExecutionRecord, error) {
	query := "SELECT id, session_id, command, file, status, start_time, end_time, duration_ms, exit_code, stdout, stderr, error_message, resource_usage, metadata, created_at FROM executions WHERE 1=1"
	args := []interface{}{}

	if filter.SessionID != "" {
		query += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}

	if filter.Command != "" {
		query += " AND command LIKE ?"
		args = append(args, "%"+filter.Command+"%")
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.File != "" {
		query += " AND file LIKE ?"
		args = append(args, "%"+filter.File+"%")
	}

	if filter.StartTime != nil {
		query += " AND start_time >= ?"
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		query += " AND start_time <= ?"
		args = append(args, *filter.EndTime)
	}

	if filter.MinDuration != nil {
		query += " AND duration_ms >= ?"
		args = append(args, *filter.MinDuration)
	}

	if filter.MaxDuration != nil {
		query += " AND duration_ms <= ?"
		args = append(args, *filter.MaxDuration)
	}

	// ソート
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	direction := "ASC"
	if filter.OrderDesc {
		direction = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, direction)

	// ページング
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*ExecutionRecord
	for rows.Next() {
		record := &ExecutionRecord{}
		err := rows.Scan(
			&record.ID, &record.SessionID, &record.Command, &record.File,
			&record.Status, &record.StartTime, &record.EndTime,
			&record.Duration, &record.ExitCode, &record.Stdout,
			&record.Stderr, &record.ErrorMessage, &record.ResourceUsage,
			&record.Metadata, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

// GetSession は指定されたセッション記録を取得する
func (s *SQLiteStore) GetSession(sessionID string) (*SessionRecord, error) {
	query := `
		SELECT id, user_id, mode, start_time, end_time, total_jobs, successful, failed, created_at
		FROM sessions WHERE id = ?
	`

	row := s.db.QueryRow(query, sessionID)
	record := &SessionRecord{}

	err := row.Scan(
		&record.ID, &record.UserID, &record.Mode, &record.StartTime,
		&record.EndTime, &record.TotalJobs, &record.Successful,
		&record.Failed, &record.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

// GetSessions はセッション一覧を取得する
func (s *SQLiteStore) GetSessions(limit, offset int) ([]*SessionRecord, error) {
	query := `
		SELECT id, user_id, mode, start_time, end_time, total_jobs, successful, failed, created_at
		FROM sessions ORDER BY start_time DESC
	`

	if limit > 0 {
		query += " LIMIT ?"
		if offset > 0 {
			query += " OFFSET ?"
		}
	}

	var rows *sql.Rows
	var err error

	if limit > 0 {
		if offset > 0 {
			rows, err = s.db.Query(query, limit, offset)
		} else {
			rows, err = s.db.Query(query, limit)
		}
	} else {
		rows, err = s.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*SessionRecord
	for rows.Next() {
		session := &SessionRecord{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.Mode, &session.StartTime,
			&session.EndTime, &session.TotalJobs, &session.Successful,
			&session.Failed, &session.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// ExportResults は指定された形式で実行結果をエクスポートする
func (s *SQLiteStore) ExportResults(format ExportFormat, filter ExecutionFilter) ([]byte, error) {
	records, err := s.GetExecutions(filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatJSON:
		return json.MarshalIndent(records, "", "  ")
	case FormatCSV:
		return s.exportCSV(records)
	case FormatHTML:
		return s.exportHTML(records)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCSV はCSV形式でエクスポートする
func (s *SQLiteStore) exportCSV(records []*ExecutionRecord) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// ヘッダー
	headers := []string{
		"ID", "SessionID", "Command", "File", "Status",
		"StartTime", "EndTime", "Duration(ms)", "ExitCode",
		"ErrorMessage", "CreatedAt",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// データ行
	for _, record := range records {
		row := []string{
			record.ID,
			record.SessionID,
			record.Command,
			record.File,
			record.Status,
			record.StartTime.Format(time.RFC3339),
			record.EndTime.Format(time.RFC3339),
			fmt.Sprintf("%d", record.Duration),
			fmt.Sprintf("%d", record.ExitCode),
			record.ErrorMessage,
			record.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// exportHTML はHTML形式でエクスポートする
func (s *SQLiteStore) exportHTML(records []*ExecutionRecord) ([]byte, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Execution Results</title>
    <style>
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .success { color: green; }
        .failed { color: red; }
        .cancelled { color: orange; }
    </style>
</head>
<body>
    <h1>Execution Results ({{len .}} records)</h1>
    <table>
        <tr>
            <th>ID</th>
            <th>Session</th>
            <th>Command</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Start Time</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.SessionID}}</td>
            <td>{{.Command}}</td>
            <td class="{{.Status}}">{{.Status}}</td>
            <td>{{.Duration}}ms</td>
            <td>{{.StartTime.Format "2006-01-02 15:04:05"}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>`

	t, err := template.New("export").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, records); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CleanupOldRecords は古い記録を削除する
func (s *SQLiteStore) CleanupOldRecords(olderThan time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 古い実行記録を削除
	_, err = tx.Exec("DELETE FROM executions WHERE created_at < ?", olderThan)
	if err != nil {
		return err
	}

	// 関連する実行記録がないセッションも削除
	_, err = tx.Exec(`
		DELETE FROM sessions 
		WHERE id NOT IN (SELECT DISTINCT session_id FROM executions)
		AND created_at < ?
	`, olderThan)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// PersistenceStatistics は永続化の統計情報を保持する
type PersistenceStatistics struct {
	TotalExecutions int64            `json:"total_executions"`
	TotalSessions   int64            `json:"total_sessions"`
	DatabaseSize    int64            `json:"database_size_bytes"`
	OldestRecord    time.Time        `json:"oldest_record"`
	NewestRecord    time.Time        `json:"newest_record"`
	StatusCounts    map[string]int64 `json:"status_counts"`
}

// GetStatistics は永続化の統計情報を取得する
func (s *SQLiteStore) GetStatistics() (*PersistenceStatistics, error) {
	stats := &PersistenceStatistics{
		StatusCounts: make(map[string]int64),
	}

	// 実行記録数
	err := s.db.QueryRow("SELECT COUNT(*) FROM executions").Scan(&stats.TotalExecutions)
	if err != nil {
		return nil, err
	}

	// セッション数
	err = s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	if err != nil {
		return nil, err
	}

	// 日付範囲
	if stats.TotalExecutions > 0 {
		var oldestStr, newestStr string
		err = s.db.QueryRow("SELECT MIN(created_at), MAX(created_at) FROM executions").Scan(&oldestStr, &newestStr)
		if err != nil {
			return nil, err
		}

		if oldestStr != "" {
			if parsed, err := time.Parse("2006-01-02 15:04:05", oldestStr); err == nil {
				stats.OldestRecord = parsed
			} else if parsed, err := time.Parse(time.RFC3339, oldestStr); err == nil {
				stats.OldestRecord = parsed
			}
		}
		if newestStr != "" {
			if parsed, err := time.Parse("2006-01-02 15:04:05", newestStr); err == nil {
				stats.NewestRecord = parsed
			} else if parsed, err := time.Parse(time.RFC3339, newestStr); err == nil {
				stats.NewestRecord = parsed
			}
		}
	}

	// ステータス別カウント
	rows, err := s.db.Query("SELECT status, COUNT(*) FROM executions GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.StatusCounts[status] = count
	}

	return stats, nil
}

// Close はデータベース接続を閉じる
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ExecutionResultToPersistence は ExecutionResult を ExecutionRecord に変換する
func ExecutionResultToPersistence(result *ExecutionResult, sessionID string) *ExecutionRecord {
	record := &ExecutionRecord{
		ID:        generateID(),
		SessionID: sessionID,
		Command:   result.Command,
		Status:    "completed",
		StartTime: time.Now().Add(-result.Duration),
		EndTime:   time.Now(),
		Duration:  result.Duration.Milliseconds(),
		ExitCode:  0,
		Stdout:    result.Output,
		CreatedAt: time.Now(),
	}

	if !result.Success {
		record.Status = "failed"
		record.ErrorMessage = result.Error
		record.ExitCode = 1
	}

	if result.Skipped {
		record.Status = "skipped"
		record.ErrorMessage = result.SkipReason
	}

	return record
}

// JobToPersistence は Job を ExecutionRecord に変換する
func JobToPersistence(job *Job) *ExecutionRecord {
	record := &ExecutionRecord{
		ID:        job.ID,
		SessionID: "",
		Command:   job.Command,
		File:      job.File,
		Status:    strings.ToLower(job.Status.String()),
		StartTime: job.StartTime,
		EndTime:   job.EndTime,
		Duration:  job.Duration.Milliseconds(),
		CreatedAt: time.Now(),
	}

	if job.Result != nil {
		record.Stdout = job.Result.Output
		record.Stderr = job.Result.Error
		if job.Result.Success {
			record.ExitCode = 0
		} else {
			record.ExitCode = 1
		}
	}

	if job.Error != nil {
		record.ErrorMessage = job.Error.Error()
	}

	if job.Metadata != nil {
		if metadataJSON, err := json.Marshal(job.Metadata); err == nil {
			record.Metadata = string(metadataJSON)
		}
	}

	return record
}

// generateID は一意のIDを生成する
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
