package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScanResult represents the overall scan result
type ScanResult struct {
	StartTime     time.Time          `json:"start_time"`
	EndTime       time.Time          `json:"end_time"`
	Duration      time.Duration      `json:"duration"`
	TotalFiles    int                `json:"total_files"`
	ScannedFiles  int                `json:"scanned_files"`
	DetectedFiles int                `json:"detected_files"`
	Results       []*DetectionResult `json:"results"`
	Errors        []string           `json:"errors"`
	Statistics    *ScanStatistics    `json:"statistics"`
}

// ScanStatistics contains statistical information about the scan
type ScanStatistics struct {
	FileTypeDistribution   map[string]int   `json:"file_type_distribution"`
	LanguageDistribution   map[string]int   `json:"language_distribution"`
	PriorityDistribution   map[Priority]int `json:"priority_distribution"`
	ConfidenceDistribution map[string]int   `json:"confidence_distribution"`
	AverageCommands        float64          `json:"average_commands"`
	AverageConfidence      float64          `json:"average_confidence"`
	AverageImportance      float64          `json:"average_importance"`
}

// ScanOptions contains options for scanning
type ScanOptions struct {
	IncludePatterns  []string `json:"include_patterns"`
	ExcludePatterns  []string `json:"exclude_patterns"`
	MaxDepth         int      `json:"max_depth"`
	FollowSymlinks   bool     `json:"follow_symlinks"`
	Parallel         bool     `json:"parallel"`
	MaxWorkers       int      `json:"max_workers"`
	SortBy           string   `json:"sort_by"`    // "confidence", "importance", "commands", "path"
	SortOrder        string   `json:"sort_order"` // "asc", "desc"
	MinConfidence    float64  `json:"min_confidence"`
	MinImportance    float64  `json:"min_importance"`
	OnlyHighPriority bool     `json:"only_high_priority"`
}

// IntelligentScanner is the main scanner that combines detection with intelligent filtering
type IntelligentScanner struct {
	detector *ScriptDetector
	options  *ScanOptions
	mu       sync.Mutex
}

// NewIntelligentScanner creates a new intelligent scanner
func NewIntelligentScanner(config *DetectionConfig, options *ScanOptions) *IntelligentScanner {
	if options == nil {
		options = &ScanOptions{
			MaxDepth:       10,
			FollowSymlinks: false,
			Parallel:       true,
			MaxWorkers:     4,
			SortBy:         "importance",
			SortOrder:      "desc",
			MinConfidence:  0.3,
			MinImportance:  0.0,
		}
	}

	if options.MaxWorkers <= 0 {
		options.MaxWorkers = 4
	}

	return &IntelligentScanner{
		detector: NewScriptDetector(config),
		options:  options,
	}
}

// Scan scans the given path for usacloud scripts
func (is *IntelligentScanner) Scan(path string) (*ScanResult, error) {
	startTime := time.Now()

	result := &ScanResult{
		StartTime: startTime,
		Results:   make([]*DetectionResult, 0),
		Errors:    make([]string, 0),
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	// Determine if it's a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	var detectionResults []*DetectionResult
	var scanErrors []string

	if fileInfo.IsDir() {
		detectionResults, scanErrors = is.scanDirectory(path)
	} else {
		detectionResults, scanErrors = is.scanSingleFile(path)
	}

	// Filter results based on options
	filteredResults := is.filterResults(detectionResults)

	// Sort results
	is.sortResults(filteredResults)

	// Calculate statistics
	statistics := is.calculateStatistics(filteredResults, detectionResults)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TotalFiles = len(detectionResults)
	result.ScannedFiles = len(detectionResults)
	result.DetectedFiles = len(filteredResults)
	result.Results = filteredResults
	result.Errors = scanErrors
	result.Statistics = statistics

	return result, nil
}

// scanDirectory scans a directory for scripts
func (is *IntelligentScanner) scanDirectory(dirPath string) ([]*DetectionResult, []string) {
	var results []*DetectionResult
	var errors []string
	var mu sync.Mutex

	filePaths, walkErrors := is.collectFilePaths(dirPath)
	errors = append(errors, walkErrors...)

	if is.options.Parallel && len(filePaths) > 1 {
		// Parallel scanning
		resultChan := make(chan *DetectionResult, len(filePaths))
		errorChan := make(chan string, len(filePaths))

		// Create worker pool
		workers := is.options.MaxWorkers
		if workers > len(filePaths) {
			workers = len(filePaths)
		}

		workChan := make(chan string, len(filePaths))
		var wg sync.WaitGroup

		// Start workers
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for filePath := range workChan {
					result, err := is.detector.ScanFile(filePath)
					if err != nil {
						errorChan <- fmt.Sprintf("failed to scan %s: %v", filePath, err)
					} else {
						resultChan <- result
					}
				}
			}()
		}

		// Send work
		go func() {
			for _, filePath := range filePaths {
				workChan <- filePath
			}
			close(workChan)
		}()

		// Wait for completion
		go func() {
			wg.Wait()
			close(resultChan)
			close(errorChan)
		}()

		// Collect results
		for result := range resultChan {
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}

		// Collect errors
		for err := range errorChan {
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}
	} else {
		// Sequential scanning
		for _, filePath := range filePaths {
			result, err := is.detector.ScanFile(filePath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to scan %s: %v", filePath, err))
			} else {
				results = append(results, result)
			}
		}
	}

	return results, errors
}

// scanSingleFile scans a single file
func (is *IntelligentScanner) scanSingleFile(filePath string) ([]*DetectionResult, []string) {
	result, err := is.detector.ScanFile(filePath)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to scan %s: %v", filePath, err)}
	}
	return []*DetectionResult{result}, nil
}

// collectFilePaths collects file paths to scan
func (is *IntelligentScanner) collectFilePaths(dirPath string) ([]string, []string) {
	var filePaths []string
	var errors []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, fmt.Sprintf("error accessing %s: %v", path, err))
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Check depth limit
			relPath, _ := filepath.Rel(dirPath, path)
			depth := strings.Count(relPath, string(filepath.Separator))
			if depth >= is.options.MaxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 && !is.options.FollowSymlinks {
			return nil
		}

		// Apply include/exclude patterns
		if !is.shouldIncludeFile(path) {
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("failed to walk directory: %v", err))
	}

	return filePaths, errors
}

// shouldIncludeFile checks if a file should be included based on patterns
func (is *IntelligentScanner) shouldIncludeFile(filePath string) bool {
	fileName := filepath.Base(filePath)

	// Check exclude patterns first
	for _, pattern := range is.options.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return false
		}
	}

	// Check include patterns
	if len(is.options.IncludePatterns) == 0 {
		return true // Include all if no patterns specified
	}

	for _, pattern := range is.options.IncludePatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}

	return false
}

// filterResults filters results based on options
func (is *IntelligentScanner) filterResults(results []*DetectionResult) []*DetectionResult {
	var filtered []*DetectionResult

	for _, result := range results {
		// Skip non-scripts
		if !result.IsScript {
			continue
		}

		// Filter by confidence
		if result.Confidence < is.options.MinConfidence {
			continue
		}

		// Filter by importance
		if result.ImportanceScore < is.options.MinImportance {
			continue
		}

		// Filter by priority
		if is.options.OnlyHighPriority {
			if result.Priority != PriorityHigh && result.Priority != PriorityCritical {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// sortResults sorts results based on options
func (is *IntelligentScanner) sortResults(results []*DetectionResult) {
	sort.Slice(results, func(i, j int) bool {
		var less bool

		switch is.options.SortBy {
		case "confidence":
			less = results[i].Confidence < results[j].Confidence
		case "importance":
			less = results[i].ImportanceScore < results[j].ImportanceScore
		case "commands":
			less = results[i].CommandCount < results[j].CommandCount
		case "path":
			less = results[i].FilePath < results[j].FilePath
		default:
			// Default to importance
			less = results[i].ImportanceScore < results[j].ImportanceScore
		}

		if is.options.SortOrder == "desc" {
			return !less
		}
		return less
	})
}

// calculateStatistics calculates scan statistics
func (is *IntelligentScanner) calculateStatistics(filteredResults, allResults []*DetectionResult) *ScanStatistics {
	stats := &ScanStatistics{
		FileTypeDistribution:   make(map[string]int),
		LanguageDistribution:   make(map[string]int),
		PriorityDistribution:   make(map[Priority]int),
		ConfidenceDistribution: make(map[string]int),
	}

	var totalCommands, totalConfidence, totalImportance float64

	for _, result := range filteredResults {
		// File type distribution
		if analysis, ok := result.Metadata["file_analysis"].(*FileAnalysis); ok {
			stats.FileTypeDistribution[analysis.FileType]++
			stats.LanguageDistribution[analysis.Language]++
		}

		// Priority distribution
		stats.PriorityDistribution[result.Priority]++

		// Confidence distribution
		confRange := is.getConfidenceRange(result.Confidence)
		stats.ConfidenceDistribution[confRange]++

		// Averages
		totalCommands += float64(result.CommandCount)
		totalConfidence += result.Confidence
		totalImportance += result.ImportanceScore
	}

	if len(filteredResults) > 0 {
		stats.AverageCommands = totalCommands / float64(len(filteredResults))
		stats.AverageConfidence = totalConfidence / float64(len(filteredResults))
		stats.AverageImportance = totalImportance / float64(len(filteredResults))
	}

	return stats
}

// getConfidenceRange returns a range string for confidence values
func (is *IntelligentScanner) getConfidenceRange(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "0.9-1.0"
	case confidence >= 0.7:
		return "0.7-0.9"
	case confidence >= 0.5:
		return "0.5-0.7"
	case confidence >= 0.3:
		return "0.3-0.5"
	default:
		return "0.0-0.3"
	}
}

// GetTopResults returns the top N results by importance
func (is *IntelligentScanner) GetTopResults(results []*DetectionResult, n int) []*DetectionResult {
	// Sort by importance descending
	sorted := make([]*DetectionResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ImportanceScore > sorted[j].ImportanceScore
	})

	if n > len(sorted) {
		n = len(sorted)
	}

	return sorted[:n]
}

// GetResultsByPriority returns results filtered by priority
func (is *IntelligentScanner) GetResultsByPriority(results []*DetectionResult, priority Priority) []*DetectionResult {
	var filtered []*DetectionResult

	for _, result := range results {
		if result.Priority == priority {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// GetSummary returns a summary of scan results
func (is *IntelligentScanner) GetSummary(scanResult *ScanResult) string {
	summary := fmt.Sprintf("Scan Summary:\n")
	summary += fmt.Sprintf("  Duration: %v\n", scanResult.Duration)
	summary += fmt.Sprintf("  Total Files: %d\n", scanResult.TotalFiles)
	summary += fmt.Sprintf("  Detected Scripts: %d\n", scanResult.DetectedFiles)
	summary += fmt.Sprintf("  Detection Rate: %.1f%%\n",
		float64(scanResult.DetectedFiles)/float64(scanResult.TotalFiles)*100)

	if scanResult.Statistics != nil {
		summary += fmt.Sprintf("\nStatistics:\n")
		summary += fmt.Sprintf("  Average Commands per Script: %.1f\n", scanResult.Statistics.AverageCommands)
		summary += fmt.Sprintf("  Average Confidence: %.3f\n", scanResult.Statistics.AverageConfidence)
		summary += fmt.Sprintf("  Average Importance: %.3f\n", scanResult.Statistics.AverageImportance)

		summary += fmt.Sprintf("\nPriority Distribution:\n")
		for priority, count := range scanResult.Statistics.PriorityDistribution {
			summary += fmt.Sprintf("  %s: %d\n", priority.String(), count)
		}
	}

	if len(scanResult.Errors) > 0 {
		summary += fmt.Sprintf("\nErrors: %d\n", len(scanResult.Errors))
	}

	return summary
}
