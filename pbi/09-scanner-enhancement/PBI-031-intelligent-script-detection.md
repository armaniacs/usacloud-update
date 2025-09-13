# PBI-031: インテリジェントスクリプト検出機能

## 概要
ディレクトリスキャン時に、usacloudコマンドを含むスクリプトファイルをより正確に検出し、ファイルの重要度や変換優先度を自動判定する機能を実装します。機械学習的アプローチとルールベースの組み合わせによる高精度な検出システムを提供します。

## 受け入れ条件
- [x] usacloudコマンドを含むファイルを高精度で検出できる
- [x] ファイル内のコマンド密度と重要度を分析できる
- [x] 変換優先度を自動的にランク付けできる
- [x] 誤検出（false positive）を最小限に抑えられる
- [x] バイナリファイルや非テキストファイルを適切に除外できる

## 技術仕様

### 1. 検出エンジン設計
```go
type ScriptDetector struct {
    patterns        []DetectionPattern
    contentAnalyzer *ContentAnalyzer
    scorer          *ImportanceScorer
    config          *DetectionConfig
}

type DetectionResult struct {
    FilePath        string                `json:"file_path"`
    IsScript        bool                  `json:"is_script"`
    Confidence      float64               `json:"confidence"`
    CommandCount    int                   `json:"command_count"`
    ImportanceScore float64               `json:"importance_score"`
    Priority        Priority              `json:"priority"`
    Commands        []DetectedCommand     `json:"commands"`
    Metadata        map[string]interface{} `json:"metadata"`
}

type DetectedCommand struct {
    Line        int     `json:"line"`
    Content     string  `json:"content"`
    CommandType string  `json:"command_type"`
    Confidence  float64 `json:"confidence"`
    Deprecated  bool    `json:"deprecated"`
}

type Priority int

const (
    PriorityLow Priority = iota
    PriorityMedium
    PriorityHigh
    PriorityCritical
)
```

### 2. パターンベース検出
```go
type DetectionPattern struct {
    Name        string
    Pattern     *regexp.Regexp
    Weight      float64
    MinMatches  int
    Description string
}

func (sd *ScriptDetector) initializePatterns() {
    sd.patterns = []DetectionPattern{
        {
            Name:        "usacloud-command",
            Pattern:     regexp.MustCompile(`(?m)^\s*usacloud\s+\w+`),
            Weight:      1.0,
            MinMatches:  1,
            Description: "直接的なusacloudコマンド呼び出し",
        },
        {
            Name:        "usacloud-variable",
            Pattern:     regexp.MustCompile(`(?m)\busacloud\s*=`),
            Weight:      0.8,
            MinMatches:  1,
            Description: "usacloudへの変数代入",
        },
        {
            Name:        "usacloud-in-pipe",
            Pattern:     regexp.MustCompile(`(?m)\|\s*usacloud\s+`),
            Weight:      0.9,
            MinMatches:  1,
            Description: "パイプ経由でのusacloud実行",
        },
        {
            Name:        "sakura-cloud-reference",
            Pattern:     regexp.MustCompile(`(?i)\b(sakura\s*cloud|さくらクラウド)\b`),
            Weight:      0.3,
            MinMatches:  2,
            Description: "Sakura Cloudへの言及",
        },
        {
            Name:        "infrastructure-commands",
            Pattern:     regexp.MustCompile(`(?m)^\s*usacloud\s+(server|disk|switch|router|database)\s+`),
            Weight:      1.2,
            MinMatches:  1,
            Description: "インフラ管理コマンド",
        },
    }
}

func (sd *ScriptDetector) detectByPatterns(content string) (float64, []DetectedCommand) {
    var totalScore float64
    var commands []DetectedCommand
    
    lines := strings.Split(content, "\n")
    
    for _, pattern := range sd.patterns {
        matches := pattern.Pattern.FindAllStringSubmatch(content, -1)
        matchCount := len(matches)
        
        if matchCount >= pattern.MinMatches {
            score := pattern.Weight * float64(matchCount)
            totalScore += score
            
            // マッチした行を特定
            for i, line := range lines {
                if pattern.Pattern.MatchString(line) {
                    commands = append(commands, DetectedCommand{
                        Line:        i + 1,
                        Content:     strings.TrimSpace(line),
                        CommandType: pattern.Name,
                        Confidence:  pattern.Weight,
                        Deprecated:  sd.isDeprecatedCommand(line),
                    })
                }
            }
        }
    }
    
    return totalScore, commands
}
```

### 3. コンテンツ分析
```go
type ContentAnalyzer struct {
    fileTypeDetector *FileTypeDetector
    encodingDetector *EncodingDetector
}

func (ca *ContentAnalyzer) AnalyzeFile(filePath string) (*FileAnalysis, error) {
    content, err := ca.readFile(filePath)
    if err != nil {
        return nil, err
    }
    
    analysis := &FileAnalysis{
        FilePath:     filePath,
        FileSize:     len(content),
        LineCount:    strings.Count(content, "\n") + 1,
        FileType:     ca.fileTypeDetector.DetectType(filePath, content),
        Encoding:     ca.encodingDetector.DetectEncoding(content),
        IsText:       ca.isTextFile(content),
        IsBinary:     ca.isBinaryFile(content),
        Language:     ca.detectScriptLanguage(content),
    }
    
    if analysis.IsText {
        analysis.TextMetrics = ca.calculateTextMetrics(content)
    }
    
    return analysis, nil
}

type FileAnalysis struct {
    FilePath    string                 `json:"file_path"`
    FileSize    int                    `json:"file_size"`
    LineCount   int                    `json:"line_count"`
    FileType    string                 `json:"file_type"`
    Encoding    string                 `json:"encoding"`
    IsText      bool                   `json:"is_text"`
    IsBinary    bool                   `json:"is_binary"`
    Language    string                 `json:"language"`
    TextMetrics *TextMetrics           `json:"text_metrics,omitempty"`
}

type TextMetrics struct {
    CommentLines    int     `json:"comment_lines"`
    BlankLines      int     `json:"blank_lines"`
    CodeLines       int     `json:"code_lines"`
    CommentRatio    float64 `json:"comment_ratio"`
    Complexity      int     `json:"complexity"`
}

func (ca *ContentAnalyzer) detectScriptLanguage(content string) string {
    // シェバン行の検査
    if strings.HasPrefix(content, "#!/") {
        shebang := strings.Split(content, "\n")[0]
        if strings.Contains(shebang, "bash") {
            return "bash"
        }
        if strings.Contains(shebang, "sh") {
            return "sh"
        }
        if strings.Contains(shebang, "python") {
            return "python"
        }
    }
    
    // ファイル内容からの推測
    if regexp.MustCompile(`(?m)^\s*(function\s+\w+\s*\(|if\s*\[|for\s+\w+\s+in)`).MatchString(content) {
        return "bash"
    }
    
    if regexp.MustCompile(`(?m)^\s*(def\s+\w+\(|import\s+\w+|if\s+__name__\s*==)`).MatchString(content) {
        return "python"
    }
    
    return "unknown"
}
```

### 4. 重要度スコアリング
```go
type ImportanceScorer struct {
    weights ScoreWeights
}

type ScoreWeights struct {
    CommandCount     float64 `json:"command_count"`
    FileSize         float64 `json:"file_size"`
    InfraCommands    float64 `json:"infra_commands"`
    DeprecatedUsage  float64 `json:"deprecated_usage"`
    ComplexityFactor float64 `json:"complexity_factor"`
}

func (is *ImportanceScorer) CalculateScore(result *DetectionResult, analysis *FileAnalysis) float64 {
    var score float64
    
    // コマンド数による基本スコア
    score += float64(result.CommandCount) * is.weights.CommandCount
    
    // ファイルサイズによる調整
    if analysis.FileSize > 1000 { // 1KB以上
        score += math.Log(float64(analysis.FileSize)) * is.weights.FileSize
    }
    
    // インフラコマンドの重み付け
    infraCount := 0
    for _, cmd := range result.Commands {
        if is.isInfrastructureCommand(cmd.CommandType) {
            infraCount++
        }
    }
    score += float64(infraCount) * is.weights.InfraCommands
    
    // 廃止コマンドの使用による重み付け
    deprecatedCount := 0
    for _, cmd := range result.Commands {
        if cmd.Deprecated {
            deprecatedCount++
        }
    }
    score += float64(deprecatedCount) * is.weights.DeprecatedUsage
    
    // 複雑さによる調整
    if analysis.TextMetrics != nil {
        complexityBonus := float64(analysis.TextMetrics.Complexity) * is.weights.ComplexityFactor
        score += complexityBonus
    }
    
    return score
}

func (is *ImportanceScorer) AssignPriority(score float64) Priority {
    switch {
    case score >= 10.0:
        return PriorityCritical
    case score >= 5.0:
        return PriorityHigh
    case score >= 2.0:
        return PriorityMedium
    default:
        return PriorityLow
    }
}
```

### 5. 統合検出システム
```go
func (sd *ScriptDetector) ScanFile(filePath string) (*DetectionResult, error) {
    // ファイル分析
    analysis, err := sd.contentAnalyzer.AnalyzeFile(filePath)
    if err != nil {
        return nil, err
    }
    
    // バイナリファイルは除外
    if analysis.IsBinary {
        return &DetectionResult{
            FilePath:   filePath,
            IsScript:   false,
            Confidence: 0.0,
        }, nil
    }
    
    // コンテンツ読み取り
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    // パターンマッチング検出
    patternScore, commands := sd.detectByPatterns(string(content))
    
    // 信頼度計算
    confidence := sd.calculateConfidence(patternScore, analysis)
    
    result := &DetectionResult{
        FilePath:     filePath,
        IsScript:     confidence > sd.config.MinConfidence,
        Confidence:   confidence,
        CommandCount: len(commands),
        Commands:     commands,
        Metadata: map[string]interface{}{
            "file_analysis": analysis,
            "pattern_score": patternScore,
        },
    }
    
    // 重要度スコア計算
    result.ImportanceScore = sd.scorer.CalculateScore(result, analysis)
    result.Priority = sd.scorer.AssignPriority(result.ImportanceScore)
    
    return result, nil
}

func (sd *ScriptDetector) calculateConfidence(patternScore float64, analysis *FileAnalysis) float64 {
    baseConfidence := math.Tanh(patternScore / 3.0) // 0-1の範囲にノーマライズ
    
    // ファイル特性による調整
    if analysis.Language == "bash" || analysis.Language == "sh" {
        baseConfidence *= 1.2
    }
    
    if strings.HasSuffix(analysis.FilePath, ".sh") {
        baseConfidence *= 1.1
    }
    
    // 上限を1.0に制限
    if baseConfidence > 1.0 {
        baseConfidence = 1.0
    }
    
    return baseConfidence
}
```

## テスト戦略
- **検出精度テスト**: 既知のスクリプトファイルでの検出率確認
- **誤検出テスト**: 非スクリプトファイルでの誤検出率測定
- **パフォーマンステスト**: 大量ファイルでのスキャン速度確認
- **エッジケーステスト**: 特殊なファイル形式や文字エンコーディングでの動作確認

## 依存関係
- 前提PBI: なし（既存スキャナー機能を拡張）
- 関連PBI: PBI-032（スキャン結果キャッシュ）、PBI-033（除外ルール管理）
- 既存コード: internal/scanner/scanner.go

## 見積もり
- 開発工数: 12時間
  - パターンベース検出エンジン: 4時間
  - コンテンツ分析機能: 4時間
  - 重要度スコアリング: 3時間
  - 統合・テスト: 1時間

## 完了の定義
- [x] usacloudスクリプトの検出精度が95%以上になる
- [x] 誤検出率が5%以下になる
- [x] 重要度スコアが実際の重要性と相関する
- [x] 大量ファイル処理でも十分な性能を維持する
- [x] 多様なファイル形式に対応する

## 実装状況
✅ **PBI-031は完全に実装済み** (2025-09-11)

以下のファイルで完全に実装されています：
- `internal/scanner/detector.go` - 高精度なスクリプト検出エンジン
- `internal/scanner/intelligent_scanner.go` - インテリジェントスキャナー統合
- `internal/scanner/detector_test.go` - 包括的なテストスイート
- `internal/scanner/intelligent_scanner_test.go` - 統合テスト

実装内容：
- 5つの検出パターンによる高精度検出システム
- コンテンツ分析による重要度スコアリング
- Priority (Low/Medium/High/Critical) による優先度付け
- バイナリファイル・非テキストファイルの自動除外
- 包括的なテストカバレッジ

## 備考
- 検出パターンは設定ファイルで調整可能にする
- 機械学習的手法の将来導入も考慮した設計
- ファイル内容は必要最小限だけ読み込んで効率化