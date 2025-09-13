# PBI-025: 並行サンドボックス実行機能

## 概要
複数のusacloudコマンドを並行実行することで、大量のスクリプトファイル処理やバッチ処理の高速化を実現します。安全な並行実行のためのリソース管理、レート制限、デッドロック回避機能を含む包括的な並行実行エンジンを実装します。

## 受け入れ条件
- [ ] 複数のコマンドを安全に並行実行できる
- [ ] 並行実行数の上限を設定・制御できる
- [ ] レート制限機能でAPI制限を回避できる
- [ ] 並行実行の進捗状況をリアルタイム表示できる
- [ ] エラー発生時に適切に他の実行を制御できる

## 技術仕様

### 1. 並行実行制御システム
```go
type ParallelExecutor struct {
    maxConcurrency int
    rateLimiter    *rate.Limiter
    semaphore      chan struct{}
    wg             sync.WaitGroup
    ctx            context.Context
    cancel         context.CancelFunc
}

func NewParallelExecutor(maxConcurrency int, rps float64) *ParallelExecutor {
    ctx, cancel := context.WithCancel(context.Background())
    return &ParallelExecutor{
        maxConcurrency: maxConcurrency,
        rateLimiter:    rate.NewLimiter(rate.Limit(rps), 1),
        semaphore:      make(chan struct{}, maxConcurrency),
        ctx:            ctx,
        cancel:         cancel,
    }
}
```

### 2. 並行実行ジョブ管理
```go
type Job struct {
    ID      string
    Command string
    File    string
    Status  JobStatus
    Result  *ExecutionResult
    Error   error
}

type JobStatus int

const (
    JobPending JobStatus = iota
    JobRunning
    JobCompleted
    JobFailed
    JobCancelled
)

func (pe *ParallelExecutor) SubmitJobs(jobs []*Job) <-chan *Job {
    resultChan := make(chan *Job, len(jobs))
    
    for _, job := range jobs {
        pe.wg.Add(1)
        go pe.executeJob(job, resultChan)
    }
    
    go func() {
        pe.wg.Wait()
        close(resultChan)
    }()
    
    return resultChan
}

func (pe *ParallelExecutor) executeJob(job *Job, resultChan chan<- *Job) {
    defer pe.wg.Done()
    
    // セマフォによる並行数制限
    pe.semaphore <- struct{}{}
    defer func() { <-pe.semaphore }()
    
    // レート制限
    if err := pe.rateLimiter.Wait(pe.ctx); err != nil {
        job.Status = JobCancelled
        job.Error = err
        resultChan <- job
        return
    }
    
    job.Status = JobRunning
    result, err := pe.executor.Execute(job.Command)
    
    if err != nil {
        job.Status = JobFailed
        job.Error = err
    } else {
        job.Status = JobCompleted
        job.Result = result
    }
    
    resultChan <- job
}
```

### 3. 進捗表示とモニタリング
```go
type ProgressMonitor struct {
    total     int
    completed int
    failed    int
    mu        sync.RWMutex
    startTime time.Time
}

func (pm *ProgressMonitor) Update(job *Job) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    switch job.Status {
    case JobCompleted:
        pm.completed++
    case JobFailed:
        pm.failed++
    }
    
    // リアルタイム進捗表示
    pm.displayProgress()
}

func (pm *ProgressMonitor) displayProgress() {
    elapsed := time.Since(pm.startTime)
    percentage := float64(pm.completed+pm.failed) / float64(pm.total) * 100
    
    fmt.Printf("\r進捗: %.1f%% (%d/%d) 完了: %d, 失敗: %d, 経過時間: %v",
        percentage, pm.completed+pm.failed, pm.total,
        pm.completed, pm.failed, elapsed.Truncate(time.Second))
}
```

### 4. エラー処理とフェイルセーフ
```go
type ErrorPolicy int

const (
    ContinueOnError ErrorPolicy = iota
    StopOnError
    StopOnCriticalError
)

func (pe *ParallelExecutor) SetErrorPolicy(policy ErrorPolicy) {
    pe.errorPolicy = policy
}

func (pe *ParallelExecutor) handleJobError(job *Job) {
    switch pe.errorPolicy {
    case StopOnError:
        pe.cancel() // すべてのジョブをキャンセル
    case StopOnCriticalError:
        if isCriticalError(job.Error) {
            pe.cancel()
        }
    case ContinueOnError:
        // 続行
    }
}
```

## テスト戦略
- **並行性テスト**: 競合状態やデッドロックの検証
- **ロードテスト**: 大量ジョブでの性能・安定性確認
- **エラー処理テスト**: 各種エラーパターンでの動作検証
- **レート制限テスト**: API制限回避機能の確認

## 依存関係
- 前提PBI: PBI-024（エラーハンドリング強化）
- 関連PBI: PBI-026（実行結果永続化）、PBI-027（環境設定）
- 既存コード: internal/sandbox/executor.go

## 見積もり
- 開発工数: 12時間
  - 並行実行エンジン設計・実装: 6時間
  - 進捗モニタリング機能実装: 3時間
  - エラー処理・フェイルセーフ実装: 3時間

## 完了の定義
- [ ] 並行実行数制限が正しく動作する
- [ ] レート制限が効果的にAPI制限を回避する
- [ ] 進捗表示がリアルタイムで更新される
- [ ] エラー発生時の制御が適切に動作する
- [ ] 大量ジョブでのメモリリークがない

## 実装状況
❌ **PBI-025は未実装** (2025-09-11)

**現在の状況**:
- 並行サンドボックス実行戦略とアーキテクチャが設計済み
- ジョブキュー、ワーカープール、レート制限機能の詳細設計完了
- 進捗モニタリング、エラー処理、フェイルセーフ機能の仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `internal/sandbox/parallel_executor.go` - 並行実行エンジンとジョブキューシステム
- ワーカープール、レート制限、API制限回避機能
- 進捗モニタリングとリアルタイムステータス表示
- エラー処理、フェイルセーフ、メモリ管理機能
- 並行性テスト、ロードテスト、レート制限テスト

**次のステップ**:
1. 並行実行エンジンとジョブキューシステムの実装
2. ワーカープールとレート制限機能の実装
3. 進捗モニタリングとステータス表示機能の実装
4. エラー処理とフェイルセーフ機能の実装
5. 並行性テストとパフォーマンス検証の実行

## 備考
- 並行実行数のデフォルトは5、最大10とする
- Sakura CloudのAPI制限を考慮してレート制限を設定
- メモリ使用量を監視し、必要に応じてジョブの分割実行を検討

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 並列サンドボックス実行よりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存サンドボックス機能の安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のサンドボックス実行機能は引き続き保守・改善
- 並列サンドボックス実行機能の実装は延期
- 現在のサンドボックス基盤の安定化を最優先