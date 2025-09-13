package scanner

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Priority represents the importance level of a detected script
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// String returns the string representation of Priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// DetectionResult represents the result of script detection
type DetectionResult struct {
	FilePath        string                 `json:"file_path"`
	IsScript        bool                   `json:"is_script"`
	Confidence      float64                `json:"confidence"`
	CommandCount    int                    `json:"command_count"`
	ImportanceScore float64                `json:"importance_score"`
	Priority        Priority               `json:"priority"`
	Commands        []DetectedCommand      `json:"commands"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DetectedCommand represents a detected usacloud command
type DetectedCommand struct {
	Line        int     `json:"line"`
	Content     string  `json:"content"`
	CommandType string  `json:"command_type"`
	Confidence  float64 `json:"confidence"`
	Deprecated  bool    `json:"deprecated"`
}

// DetectionPattern represents a pattern for detecting usacloud usage
type DetectionPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Weight      float64
	MinMatches  int
	Description string
}

// FileAnalysis represents the analysis of a file
type FileAnalysis struct {
	FilePath    string       `json:"file_path"`
	FileSize    int          `json:"file_size"`
	LineCount   int          `json:"line_count"`
	FileType    string       `json:"file_type"`
	Encoding    string       `json:"encoding"`
	IsText      bool         `json:"is_text"`
	IsBinary    bool         `json:"is_binary"`
	Language    string       `json:"language"`
	TextMetrics *TextMetrics `json:"text_metrics,omitempty"`
}

// TextMetrics represents metrics about text content
type TextMetrics struct {
	CommentLines int     `json:"comment_lines"`
	BlankLines   int     `json:"blank_lines"`
	CodeLines    int     `json:"code_lines"`
	CommentRatio float64 `json:"comment_ratio"`
	Complexity   int     `json:"complexity"`
}

// DetectionConfig contains configuration for detection
type DetectionConfig struct {
	MinConfidence     float64 `json:"min_confidence"`
	MaxFileSize       int64   `json:"max_file_size"`
	ScanBinaryFiles   bool    `json:"scan_binary_files"`
	EnableDetailedLog bool    `json:"enable_detailed_log"`
}

// ScoreWeights contains weights for importance scoring
type ScoreWeights struct {
	CommandCount     float64 `json:"command_count"`
	FileSize         float64 `json:"file_size"`
	InfraCommands    float64 `json:"infra_commands"`
	DeprecatedUsage  float64 `json:"deprecated_usage"`
	ComplexityFactor float64 `json:"complexity_factor"`
}

// ContentAnalyzer analyzes file content
type ContentAnalyzer struct {
	config *DetectionConfig
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(config *DetectionConfig) *ContentAnalyzer {
	return &ContentAnalyzer{
		config: config,
	}
}

// AnalyzeFile analyzes a file and returns its analysis
func (ca *ContentAnalyzer) AnalyzeFile(filePath string) (*FileAnalysis, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size limit
	if ca.config.MaxFileSize > 0 && fileInfo.Size() > ca.config.MaxFileSize {
		return &FileAnalysis{
			FilePath: filePath,
			FileSize: int(fileInfo.Size()),
			IsText:   false,
			IsBinary: true,
		}, nil
	}

	content, err := ca.readFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	analysis := &FileAnalysis{
		FilePath:  filePath,
		FileSize:  len(content),
		LineCount: strings.Count(content, "\n") + 1,
		FileType:  ca.detectFileType(filePath, content),
		Encoding:  ca.detectEncoding(content),
		IsText:    ca.isTextFile(content),
		IsBinary:  ca.isBinaryFile(content),
		Language:  ca.detectScriptLanguage(content),
	}

	if analysis.IsText {
		analysis.TextMetrics = ca.calculateTextMetrics(content)
	}

	return analysis, nil
}

// readFileContent reads file content with size limit
func (ca *ContentAnalyzer) readFileContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the close error but don't override the main return error
		}
	}()

	// Read with limit
	var content strings.Builder
	buffer := make([]byte, 4096)
	totalRead := 0

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			if ca.config.MaxFileSize > 0 && totalRead+n > int(ca.config.MaxFileSize) {
				// Limit reached
				remaining := int(ca.config.MaxFileSize) - totalRead
				content.Write(buffer[:remaining])
				break
			}
			content.Write(buffer[:n])
			totalRead += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return content.String(), nil
}

// detectFileType detects the type of file
func (ca *ContentAnalyzer) detectFileType(filePath, content string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".sh", ".bash":
		return "shell"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".go":
		return "go"
	case ".txt", ".md", ".readme":
		return "text"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		// Try to detect from content
		if strings.HasPrefix(content, "#!/") {
			return "script"
		}
		return "unknown"
	}
}

// detectEncoding detects the encoding of content
func (ca *ContentAnalyzer) detectEncoding(content string) string {
	if utf8.ValidString(content) {
		return "utf-8"
	}
	return "unknown"
}

// isTextFile determines if content is text
func (ca *ContentAnalyzer) isTextFile(content string) bool {
	// Check for null bytes (common in binary files)
	if strings.ContainsRune(content, 0) {
		return false
	}

	// Check UTF-8 validity
	if !utf8.ValidString(content) {
		return false
	}

	// Check ratio of printable characters
	printable := 0
	total := len(content)

	for _, r := range content {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 32 && r < 127) {
			printable++
		}
	}

	if total == 0 {
		return true
	}

	ratio := float64(printable) / float64(total)
	return ratio > 0.7 // 70% or more printable characters
}

// isBinaryFile determines if content is binary
func (ca *ContentAnalyzer) isBinaryFile(content string) bool {
	return !ca.isTextFile(content)
}

// detectScriptLanguage detects the scripting language
func (ca *ContentAnalyzer) detectScriptLanguage(content string) string {
	// Check shebang line
	if strings.HasPrefix(content, "#!/") {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			shebang := lines[0]
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
	}

	// Detect from content patterns
	if regexp.MustCompile(`(?m)^\s*(function\s+\w+\s*\(|if\s*\[|for\s+\w+\s+in)`).MatchString(content) {
		return "bash"
	}

	if regexp.MustCompile(`(?m)^\s*(def\s+\w+\(|import\s+\w+|if\s+__name__\s*==)`).MatchString(content) {
		return "python"
	}

	return "unknown"
}

// calculateTextMetrics calculates metrics for text content
func (ca *ContentAnalyzer) calculateTextMetrics(content string) *TextMetrics {
	lines := strings.Split(content, "\n")

	metrics := &TextMetrics{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			metrics.BlankLines++
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			metrics.CommentLines++
		} else {
			metrics.CodeLines++
		}

		// Simple complexity calculation based on control structures
		if regexp.MustCompile(`\b(if|for|while|case|function)\b`).MatchString(line) {
			metrics.Complexity++
		}
	}

	totalLines := len(lines)
	if totalLines > 0 {
		metrics.CommentRatio = float64(metrics.CommentLines) / float64(totalLines)
	}

	return metrics
}

// ImportanceScorer calculates importance scores
type ImportanceScorer struct {
	weights ScoreWeights
}

// NewImportanceScorer creates a new importance scorer
func NewImportanceScorer(weights ScoreWeights) *ImportanceScorer {
	return &ImportanceScorer{
		weights: weights,
	}
}

// CalculateScore calculates the importance score for a detection result
func (is *ImportanceScorer) CalculateScore(result *DetectionResult, analysis *FileAnalysis) float64 {
	var score float64

	// Command count base score
	score += float64(result.CommandCount) * is.weights.CommandCount

	// File size adjustment
	if analysis.FileSize > 1000 { // 1KB or more
		score += math.Log(float64(analysis.FileSize)) * is.weights.FileSize
	}

	// Infrastructure commands weight
	infraCount := 0
	for _, cmd := range result.Commands {
		if is.isInfrastructureCommand(cmd.CommandType) {
			infraCount++
		}
	}
	score += float64(infraCount) * is.weights.InfraCommands

	// Deprecated commands weight
	deprecatedCount := 0
	for _, cmd := range result.Commands {
		if cmd.Deprecated {
			deprecatedCount++
		}
	}
	score += float64(deprecatedCount) * is.weights.DeprecatedUsage

	// Complexity adjustment
	if analysis.TextMetrics != nil {
		complexityBonus := float64(analysis.TextMetrics.Complexity) * is.weights.ComplexityFactor
		score += complexityBonus
	}

	return score
}

// AssignPriority assigns priority based on score
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

// isInfrastructureCommand checks if command is infrastructure related
func (is *ImportanceScorer) isInfrastructureCommand(commandType string) bool {
	infraCommands := []string{
		"infrastructure-commands",
		"usacloud-command",
	}

	for _, infra := range infraCommands {
		if commandType == infra {
			return true
		}
	}
	return false
}

// ScriptDetector detects usacloud scripts
type ScriptDetector struct {
	patterns        []DetectionPattern
	contentAnalyzer *ContentAnalyzer
	scorer          *ImportanceScorer
	config          *DetectionConfig
}

// NewScriptDetector creates a new script detector
func NewScriptDetector(config *DetectionConfig) *ScriptDetector {
	if config == nil {
		config = &DetectionConfig{
			MinConfidence:   0.5,
			MaxFileSize:     1024 * 1024, // 1MB
			ScanBinaryFiles: false,
		}
	}

	weights := ScoreWeights{
		CommandCount:     1.0,
		FileSize:         0.1,
		InfraCommands:    1.5,
		DeprecatedUsage:  2.0,
		ComplexityFactor: 0.2,
	}

	sd := &ScriptDetector{
		contentAnalyzer: NewContentAnalyzer(config),
		scorer:          NewImportanceScorer(weights),
		config:          config,
	}

	sd.initializePatterns()
	return sd
}

// initializePatterns initializes detection patterns
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
			Pattern:     regexp.MustCompile(`(?im)\busacloud\s*=`),
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

// detectByPatterns detects commands using patterns
func (sd *ScriptDetector) detectByPatterns(content string) (float64, []DetectedCommand) {
	var totalScore float64
	var commands []DetectedCommand

	lines := strings.Split(content, "\n")

	for _, pattern := range sd.patterns {
		// Find all matches
		for i, line := range lines {
			if pattern.Pattern.MatchString(line) {
				totalScore += pattern.Weight
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

	return totalScore, commands
}

// isDeprecatedCommand checks if a command line contains deprecated usage
func (sd *ScriptDetector) isDeprecatedCommand(line string) bool {
	deprecated := []string{
		"summary",
		"object-storage",
		"iso-image",
		"startup-script",
		"ipv4",
		"product-",
	}

	for _, dep := range deprecated {
		if strings.Contains(line, dep) {
			return true
		}
	}
	return false
}

// calculateConfidence calculates confidence score
func (sd *ScriptDetector) calculateConfidence(patternScore float64, analysis *FileAnalysis) float64 {
	baseConfidence := math.Tanh(patternScore / 3.0) // Normalize to 0-1 range

	// Adjust based on file characteristics
	if analysis.Language == "bash" || analysis.Language == "sh" {
		baseConfidence *= 1.2
	}

	if strings.HasSuffix(analysis.FilePath, ".sh") {
		baseConfidence *= 1.1
	}

	// Cap at 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// ScanFile scans a single file for usacloud usage
func (sd *ScriptDetector) ScanFile(filePath string) (*DetectionResult, error) {
	// Analyze file
	analysis, err := sd.contentAnalyzer.AnalyzeFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze file: %w", err)
	}

	// Skip binary files unless configured to scan them
	if analysis.IsBinary && !sd.config.ScanBinaryFiles {
		return &DetectionResult{
			FilePath:   filePath,
			IsScript:   false,
			Confidence: 0.0,
		}, nil
	}

	// Read content for pattern matching
	content, err := sd.contentAnalyzer.readFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Pattern-based detection
	patternScore, commands := sd.detectByPatterns(content)

	// Calculate confidence
	confidence := sd.calculateConfidence(patternScore, analysis)

	result := &DetectionResult{
		FilePath:     filePath,
		IsScript:     confidence >= sd.config.MinConfidence,
		Confidence:   confidence,
		CommandCount: len(commands),
		Commands:     commands,
		Metadata: map[string]interface{}{
			"file_analysis": analysis,
			"pattern_score": patternScore,
		},
	}

	// Calculate importance score
	result.ImportanceScore = sd.scorer.CalculateScore(result, analysis)
	result.Priority = sd.scorer.AssignPriority(result.ImportanceScore)

	return result, nil
}

// ScanDirectory scans a directory for usacloud scripts
func (sd *ScriptDetector) ScanDirectory(dirPath string) ([]*DetectionResult, error) {
	var results []*DetectionResult

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Scan file
		result, scanErr := sd.ScanFile(path)
		if scanErr != nil {
			if sd.config.EnableDetailedLog {
				fmt.Printf("Warning: failed to scan file %s: %v\n", path, scanErr)
			}
			return nil // Continue scanning other files
		}

		if result.IsScript {
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return results, nil
}
