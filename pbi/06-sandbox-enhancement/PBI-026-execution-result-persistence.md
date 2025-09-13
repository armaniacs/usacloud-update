# PBI-026: サンドボックス実行結果の永続化

## 概要
サンドボックスでの実行結果を構造化して永続化し、後から結果の参照・分析・再利用を可能にします。実行履歴の検索、フィルタリング、エクスポート機能を含む包括的な結果管理システムを実装します。

## 受け入れ条件
- [ ] 実行結果を構造化データベースに永続化できる
- [ ] 実行履歴を時系列で参照・検索できる
- [ ] 結果をJSON、CSV形式でエクスポートできる
- [ ] 実行パフォーマンスデータを蓄積・分析できる
- [ ] 古い実行結果の自動クリーンアップ機能を提供する

## 技術仕様

### 1. データモデル設計
```go
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
```

### 2. データベース管理
```go
type ResultStore interface {
    SaveExecution(record *ExecutionRecord) error
    SaveSession(record *SessionRecord) error
    GetExecutions(filter ExecutionFilter) ([]*ExecutionRecord, error)
    GetSession(sessionID string) (*SessionRecord, error)
    ExportResults(format string, filter ExecutionFilter) ([]byte, error)
    CleanupOldRecords(olderThan time.Time) error
}

type SQLiteStore struct {
    db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    
    store := &SQLiteStore{db: db}
    if err := store.initSchema(); err != nil {
        return nil, err
    }
    
    return store, nil
}

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
```

### 3. 検索・フィルタリング機能
```go
type ExecutionFilter struct {
    SessionID    string
    Command      string
    Status       string
    File         string
    StartTime    *time.Time
    EndTime      *time.Time
    MinDuration  *int64
    MaxDuration  *int64
    Limit        int
    Offset       int
    OrderBy      string
    OrderDesc    bool
}

func (s *SQLiteStore) GetExecutions(filter ExecutionFilter) ([]*ExecutionRecord, error) {
    query := "SELECT * FROM executions WHERE 1=1"
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
    
    // 他の条件も同様に追加
    
    if filter.OrderBy != "" {
        direction := "ASC"
        if filter.OrderDesc {
            direction = "DESC"
        }
        query += fmt.Sprintf(" ORDER BY %s %s", filter.OrderBy, direction)
    }
    
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
    
    return records, nil
}
```

### 4. エクスポート機能
```go
type ExportFormat string

const (
    FormatJSON ExportFormat = "json"
    FormatCSV  ExportFormat = "csv"
    FormatHTML ExportFormat = "html"
)

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

func (s *SQLiteStore) exportCSV(records []*ExecutionRecord) ([]byte, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    
    // ヘッダー
    headers := []string{
        "ID", "SessionID", "Command", "File", "Status",
        "StartTime", "EndTime", "Duration(ms)", "ExitCode",
        "ErrorMessage", "CreatedAt",
    }
    writer.Write(headers)
    
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
        writer.Write(row)
    }
    
    writer.Flush()
    return buf.Bytes(), writer.Error()
}
```

## テスト戦略
- **データ永続化テスト**: CRUD操作の正確性検証
- **検索性能テスト**: 大量データでの検索速度確認
- **エクスポート機能テスト**: 各形式での正確な出力検証
- **並行アクセステスト**: 複数実行での整合性確認

## 依存関係
- 前提PBI: PBI-024（エラーハンドリング）、PBI-025（並行実行）
- 関連PBI: PBI-027（環境設定）、PBI-028（実行履歴管理）
- 既存コード: internal/sandbox/executor.go

## 見積もり
- 開発工数: 10時間
  - データモデル・スキーマ設計: 2時間
  - データベース操作実装: 4時間
  - 検索・フィルタリング機能: 2時間
  - エクスポート機能実装: 2時間

## 完了の定義
- [ ] 実行結果が正確に永続化される
- [ ] 検索・フィルタリング機能が正常に動作する
- [ ] 各形式でのエクスポートが正しく機能する
- [ ] 大量データでの性能が十分である
- [ ] 自動クリーンアップ機能が動作する

## 実装状況
❌ **PBI-026は未実装** (2025-09-11)

**現在の状況**:
- 実行結果永続化戦略とデータモデルが設計済み
- SQLiteベースのデータベースシステム、検索・フィルタリング機能の詳細設計完了
- エクスポート機能、自動クリーンアップ、性能最適化の仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `internal/sandbox/persistence.go` - 実行結果永続化システム
- SQLiteデータベーススキーマ、CRUD操作、インデックス設計
- 検索・フィルタリング機能、高度なクエリ機能
- エクスポート機能（JSON、CSV、HTMLレポート）
- 自動クリーンアップ、プライバシー配慮機能
- データ永続化テスト、性能テスト、並行アクセステスト

**次のステップ**:
1. データベーススキーマとCRUD操作の実装
2. 検索・フィルタリング機能の実装
3. エクスポート機能とレポート生成の実装
4. 自動クリーンアップとプライバシー機能の実装
5. 性能最適化と包括的テストの実行

## 備考
- データベースはSQLiteを使用（軽量で配布が容易）
- 大量データ対応のため、適切なインデックスを設定
- プライバシー考慮のため、機密データのマスキング機能も検討

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 実行結果永続化よりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存サンドボックス機能の安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のサンドボックス実行結果は引き続き保守・改善
- 実行結果永続化機能の実装は延期
- 現在のサンドボックス基盤の安定化を最優先