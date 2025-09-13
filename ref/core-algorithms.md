# Core Algorithms Reference

このドキュメントでは、usacloud-updateプロジェクトの中核を成すアルゴリズムの詳細実装について説明します。

## Transform Engine アルゴリズム

### 概要

Transform Engineは、usacloudコマンドをv0.x/v1.0からv1.1に変換するための中核アルゴリズムです。行単位での処理により、大容量ファイルにも対応しています。

### アーキテクチャ

```go
type Engine struct {
    rules []Rule
}

type Rule interface {
    Name() string
    Apply(line string) (string, bool, string, string)
}

type Result struct {
    Line    string
    Changed bool
    Changes []Change
}
```

### 変換フロー

#### 1. エンジン初期化
```go
func NewDefaultEngine() *Engine {
    return &Engine{rules: DefaultRules()}
}
```

**処理内容**:
- 9つのカテゴリの変換ルールを登録
- 正規表現パターンをコンパイル済み状態で保持

#### 2. 行単位変換処理
```go
func (e *Engine) Apply(line string) Result {
    result := Result{Line: line, Changed: false}
    
    for _, rule := range e.rules {
        newLine, changed, before, after := rule.Apply(result.Line)
        if changed {
            result.Line = newLine
            result.Changed = true
            result.Changes = append(result.Changes, Change{
                RuleName: rule.Name(),
                Before:   before,
                After:    after,
            })
        }
    }
    
    return result
}
```

**アルゴリズム特徴**:
1. **シーケンシャル適用**: ルールを順次適用
2. **累積変更**: 前のルールの結果に次のルールを適用
3. **変更追跡**: 各変更の詳細を記録

#### 3. ルール適用メカニズム

```go
type simpleRule struct {
    name   string
    re     *regexp.Regexp
    repl   func([]string) string
    reason string
    url    string
}

func (r *simpleRule) Apply(line string) (string, bool, string, string) {
    // 1. パターンマッチング
    m := r.re.FindStringSubmatch(line)
    if m == nil {
        return line, false, "", ""
    }
    
    // 2. 置換処理
    after := r.re.ReplaceAllString(line, r.repl(m))
    
    // 3. コメント追加
    comment := fmt.Sprintf(" # usacloud-update: %s (%s)", r.reason, r.url)
    if !strings.Contains(after, "# usacloud-update:") {
        after += comment
    }
    
    // 4. 変更フラグメント抽出
    beforeFrag := strings.TrimSpace(m[0])
    afterFrag := strings.TrimSpace(r.repl(m))
    
    return after, true, beforeFrag, afterFrag
}
```

### 変換ルールアルゴリズム

#### ルール1: 出力形式変換
```go
// パターン: usacloud文脈での --output-type=csv/tsv → json
Pattern: `(?i)\busacloud\s+[^\s]*\s+.*?(--output-type|\s-o)\s*=?\s*(csv|tsv)`
Algorithm:
  1. usacloudコマンドの識別
  2. --output-type または -o オプション検出
  3. csv/tsv → json への置換
  4. 説明コメント付加
```

#### ルール2: セレクタ廃止
```go
// パターン: --selector name=xxx → xxx
Pattern: `--selector\s+([^\\s]+)`
Algorithm:
  1. --selector オプション検出
  2. key=value 形式の解析
  3. value部分のみ抽出
  4. 引数位置への移動
```

#### ルール3-5: リソース名変更
```go
// 複数パターンの統一処理
Patterns:
  - iso-image → cdrom
  - startup-script → note
  - ipv4 → ipaddress
Algorithm:
  1. usacloudコマンド文脈の確認
  2. 対象リソース名の検出
  3. 新リソース名への置換
```

## Validation System アルゴリズム

### コマンドライン解析アルゴリズム

#### 1. 字句解析（Lexical Analysis）
```go
func ParseCommand(line string) (*CommandLine, error) {
    // 1. 前処理
    line = strings.TrimSpace(line)
    if line == "" {
        return nil, ErrEmptyCommand
    }
    
    // 2. usacloudコマンド確認
    if !strings.HasPrefix(line, "usacloud") {
        return nil, ErrNotUsacloudCommand
    }
    
    // 3. トークン分割
    tokens := tokenize(line)
    
    // 4. 構文解析
    return parseTokens(tokens)
}
```

#### 2. トークン化アルゴリズム
```go
func tokenize(line string) []string {
    var tokens []string
    var current strings.Builder
    inQuotes := false
    
    for _, r := range line {
        switch r {
        case '"', '\'':
            if inQuotes {
                tokens = append(tokens, current.String())
                current.Reset()
                inQuotes = false
            } else {
                inQuotes = true
            }
        case ' ', '\t':
            if !inQuotes && current.Len() > 0 {
                tokens = append(tokens, current.String())
                current.Reset()
            } else if inQuotes {
                current.WriteRune(r)
            }
        default:
            current.WriteRune(r)
        }
    }
    
    if current.Len() > 0 {
        tokens = append(tokens, current.String())
    }
    
    return tokens
}
```

#### 3. 構文解析アルゴリズム
```go
func parseTokens(tokens []string) (*CommandLine, error) {
    cmd := &CommandLine{
        Raw:     strings.Join(tokens, " "),
        Options: make(map[string]string),
        Flags:   []string{},
    }
    
    // usacloud スキップ
    i := 1
    
    // メインコマンド抽出
    if i < len(tokens) {
        cmd.MainCommand = tokens[i]
        i++
    }
    
    // サブコマンド抽出
    if i < len(tokens) && !strings.HasPrefix(tokens[i], "-") {
        cmd.SubCommand = tokens[i]
        i++
    }
    
    // オプション・引数解析
    for i < len(tokens) {
        token := tokens[i]
        if strings.HasPrefix(token, "--") {
            // ロングオプション処理
            parseOption(cmd, token, tokens, &i)
        } else if strings.HasPrefix(token, "-") {
            // ショートオプション処理
            parseFlag(cmd, token)
        } else {
            // 引数
            cmd.Arguments = append(cmd.Arguments, token)
        }
        i++
    }
    
    return cmd, nil
}
```

### 検証アルゴリズム

#### 1. メインコマンド検証
```go
func ValidateMainCommand(cmd string) error {
    validCommands := []string{
        "server", "disk", "archive", "cdrom", "bridge",
        "switch", "router", "load-balancer", "vpn-gateway",
        "database", "nfs", "simple-monitor", "license",
        "ipaddress", "subnet", "packet-filter", "note",
        "ssh-key", "certificate", "auth-status", "bill",
        "web-accel", "esme", "dns", "gslb", "proxylb",
        "mobile-gateway", "local-router", "config",
        "profile", "zone", "region",
    }
    
    for _, valid := range validCommands {
        if cmd == valid {
            return nil
        }
    }
    
    return fmt.Errorf("invalid main command: %s", cmd)
}
```

#### 2. 類似コマンド提案アルゴリズム
```go
func SuggestSimilarCommands(input string, candidates []string) []SimilarityResult {
    var results []SimilarityResult
    
    for _, candidate := range candidates {
        similarity := calculateSimilarity(input, candidate)
        if similarity > 0.3 { // 閾値
            results = append(results, SimilarityResult{
                Command:    candidate,
                Similarity: similarity,
            })
        }
    }
    
    // 類似度順ソート
    sort.Slice(results, func(i, j int) bool {
        return results[i].Similarity > results[j].Similarity
    })
    
    return results
}

func calculateSimilarity(s1, s2 string) float64 {
    // レーベンシュタイン距離ベースの類似度計算
    distance := levenshteinDistance(s1, s2)
    maxLen := max(len(s1), len(s2))
    if maxLen == 0 {
        return 1.0
    }
    return 1.0 - float64(distance)/float64(maxLen)
}
```

## TUI System アルゴリズム

### イベント処理アルゴリズム

#### 1. キー入力処理
```go
func (app *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
    switch event.Key() {
    case tcell.KeyRune:
        switch event.Rune() {
        case '?':
            // ヘルプ切り替え
            app.toggleHelp()
            return nil
        case 'q':
            // 終了
            app.Stop()
            return nil
        }
    case tcell.KeyEscape:
        // 前画面に戻る
        app.goBack()
        return nil
    }
    
    return event
}
```

#### 2. 動的レイアウト管理
```go
func (app *App) updateLayout() {
    if app.helpVisible {
        // ヘルプ表示時のレイアウト
        app.mainFlex.Clear()
        app.mainFlex.AddItem(app.contentView, 0, 2, true)
        app.mainFlex.AddItem(app.helpPanel, 0, 1, false)
    } else {
        // 通常時のレイアウト
        app.mainFlex.Clear()
        app.mainFlex.AddItem(app.contentView, 0, 1, true)
    }
    app.Draw()
}
```

## Sandbox Execution アルゴリズム

### 実行制御アルゴリズム

#### 1. コマンド実行
```go
func (e *Executor) Execute(cmd string) (*Result, error) {
    // 1. 前処理検証
    if err := e.validateCommand(cmd); err != nil {
        return nil, err
    }
    
    // 2. 環境準備
    ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
    defer cancel()
    
    // 3. コマンド実行
    result, err := e.executeWithRetry(ctx, cmd)
    if err != nil {
        return nil, err
    }
    
    // 4. 結果処理
    return e.processResult(result), nil
}
```

#### 2. リトライアルゴリズム
```go
func (e *Executor) executeWithRetry(ctx context.Context, cmd string) (*Result, error) {
    var lastErr error
    
    for attempt := 1; attempt <= e.maxRetries; attempt++ {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        result, err := e.executeOnce(ctx, cmd)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 指数バックオフ
        if attempt < e.maxRetries {
            backoff := time.Duration(attempt*attempt) * e.baseBackoff
            time.Sleep(backoff)
        }
    }
    
    return nil, fmt.Errorf("execution failed after %d attempts: %w", 
                          e.maxRetries, lastErr)
}
```

## パフォーマンス最適化

### 1. 正規表現最適化
- **コンパイル時最適化**: 全ルールの正規表現を事前コンパイル
- **マッチング効率**: 非マッチ時の早期終了
- **メモリ効率**: 文字列操作の最小化

### 2. 並列処理
```go
func (e *Engine) ProcessFiles(files []string) error {
    sem := make(chan struct{}, runtime.NumCPU())
    var wg sync.WaitGroup
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            e.ProcessFile(f)
        }(file)
    }
    
    wg.Wait()
    return nil
}
```

### 3. メモリ管理
- **バッファリング**: 大ファイル用の効率的バッファ管理
- **ガベージコレクション**: 不要オブジェクトの適切な解放
- **ストリーミング**: 行単位処理による定数メモリ使用量

---

**最終更新**: 2025年1月
**バージョン**: v1.9.0対応
**メンテナー**: 開発チーム