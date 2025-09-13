package performance

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// safeUint64ToTimeDuration はuint64をtime.Durationに安全に変換
func (pts *PerformanceTestSuite) safeUint64ToTimeDuration(ns uint64) time.Duration {
	// int64の最大値を超えないようにチェック
	if ns > math.MaxInt64 {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(ns)
}

type PerformanceMetrics struct {
	TotalTime          time.Duration `json:"total_time"`
	AverageTimePerLine time.Duration `json:"average_time_per_line"`
	ProcessingRate     float64       `json:"processing_rate"`

	InitialMemoryMB  float64 `json:"initial_memory_mb"`
	PeakMemoryMB     float64 `json:"peak_memory_mb"`
	FinalMemoryMB    float64 `json:"final_memory_mb"`
	MemoryEfficiency float64 `json:"memory_efficiency"`

	TotalLines       int `json:"total_lines"`
	ProcessedLines   int `json:"processed_lines"`
	TransformedLines int `json:"transformed_lines"`
	ErrorsDetected   int `json:"errors_detected"`

	CPUUsage       float64       `json:"cpu_usage"`
	GoroutineCount int           `json:"goroutine_count"`
	GCCount        uint32        `json:"gc_count"`
	GCPauseTime    time.Duration `json:"gc_pause_time"`
}

type PerformanceTolerance struct {
	TimeRegression    float64
	MemoryRegression  float64
	MinProcessingRate float64
	MaxMemoryPerLine  float64
}

type PerformanceTestSuite struct {
	t               *testing.T
	tempDir         string
	warmupRuns      int
	measurementRuns int
	timeout         time.Duration
	tolerance       PerformanceTolerance
}

func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		t:               t,
		tempDir:         t.TempDir(),
		warmupRuns:      3,
		measurementRuns: 5,
		timeout:         5 * time.Minute,

		tolerance: PerformanceTolerance{
			TimeRegression:    0.05,
			MemoryRegression:  0.10,
			MinProcessingRate: 1000,
			MaxMemoryPerLine:  1.0,
		},
	}
}

type ProcessingResult struct {
	ProcessedLines   int
	TransformedLines int
	ErrorsDetected   int
	Duration         time.Duration
}

func (pts *PerformanceTestSuite) MeasurePerformance(
	testName string,
	inputFile string,
	testFunc func() *ProcessingResult,
) *PerformanceMetrics {
	pts.t.Helper()

	for i := 0; i < pts.warmupRuns; i++ {
		testFunc()
		runtime.GC()
	}

	var measurements []*PerformanceMetrics
	for i := 0; i < pts.measurementRuns; i++ {
		metrics := pts.measureSingleRun(testFunc, inputFile)
		measurements = append(measurements, metrics)

		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	avgMetrics := pts.calculateAverageMetrics(measurements)
	pts.validatePerformanceRequirements(testName, avgMetrics)

	return avgMetrics
}

func (pts *PerformanceTestSuite) measureSingleRun(
	testFunc func() *ProcessingResult,
	inputFile string,
) *PerformanceMetrics {
	var initialMemStats, finalMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)

	startTime := time.Now()
	result := testFunc()
	endTime := time.Now()

	runtime.ReadMemStats(&finalMemStats)

	totalLines := pts.countLines(inputFile)

	executionTime := endTime.Sub(startTime)
	if result.Duration > 0 {
		executionTime = result.Duration
	}

	return &PerformanceMetrics{
		TotalTime:          executionTime,
		AverageTimePerLine: executionTime / time.Duration(totalLines),
		ProcessingRate:     float64(totalLines) / executionTime.Seconds(),

		InitialMemoryMB:  float64(initialMemStats.Alloc) / 1024 / 1024,
		PeakMemoryMB:     float64(finalMemStats.Sys) / 1024 / 1024,
		FinalMemoryMB:    float64(finalMemStats.Alloc) / 1024 / 1024,
		MemoryEfficiency: float64(totalLines) / (float64(finalMemStats.Alloc) / 1024 / 1024),

		TotalLines:       totalLines,
		ProcessedLines:   result.ProcessedLines,
		TransformedLines: result.TransformedLines,
		ErrorsDetected:   result.ErrorsDetected,

		CPUUsage:       0.0,
		GoroutineCount: runtime.NumGoroutine(),
		GCCount:        finalMemStats.NumGC - initialMemStats.NumGC,
		GCPauseTime:    pts.safeUint64ToTimeDuration(finalMemStats.PauseTotalNs - initialMemStats.PauseTotalNs),
	}
}

func (pts *PerformanceTestSuite) calculateAverageMetrics(measurements []*PerformanceMetrics) *PerformanceMetrics {
	if len(measurements) == 0 {
		return &PerformanceMetrics{}
	}

	avg := &PerformanceMetrics{}
	for _, m := range measurements {
		avg.TotalTime += m.TotalTime
		avg.ProcessingRate += m.ProcessingRate
		avg.PeakMemoryMB += m.PeakMemoryMB
		avg.FinalMemoryMB += m.FinalMemoryMB
		avg.MemoryEfficiency += m.MemoryEfficiency
		avg.TotalLines = m.TotalLines
		avg.ProcessedLines = m.ProcessedLines
		avg.TransformedLines = m.TransformedLines
		avg.ErrorsDetected = m.ErrorsDetected
		avg.GoroutineCount += m.GoroutineCount
		avg.GCCount += m.GCCount
		avg.GCPauseTime += m.GCPauseTime
	}

	count := float64(len(measurements))
	avg.TotalTime = time.Duration(float64(avg.TotalTime) / count)
	avg.ProcessingRate /= count
	avg.PeakMemoryMB /= count
	avg.FinalMemoryMB /= count
	avg.MemoryEfficiency /= count
	avg.GoroutineCount = int(float64(avg.GoroutineCount) / count)
	avg.GCCount = uint32(float64(avg.GCCount) / count)
	avg.GCPauseTime = time.Duration(float64(avg.GCPauseTime) / count)

	avg.AverageTimePerLine = avg.TotalTime / time.Duration(avg.TotalLines)

	return avg
}

func (pts *PerformanceTestSuite) validatePerformanceRequirements(testName string, metrics *PerformanceMetrics) {
	pts.t.Helper()

	if metrics.ProcessingRate < pts.tolerance.MinProcessingRate {
		pts.t.Errorf("%s: 処理レートが要件を下回りました: %.2f < %.2f lines/sec",
			testName, metrics.ProcessingRate, pts.tolerance.MinProcessingRate)
	}

	memoryPerLine := metrics.PeakMemoryMB / float64(metrics.TotalLines)
	if memoryPerLine > pts.tolerance.MaxMemoryPerLine {
		pts.t.Errorf("%s: 行あたりメモリ使用量が要件を上回りました: %.6f > %.6f MB/line",
			testName, memoryPerLine, pts.tolerance.MaxMemoryPerLine)
	}

	if metrics.GCCount > 10 {
		pts.t.Logf("%s: GC回数が多い可能性があります: %d回", testName, metrics.GCCount)
	}
}

func (pts *PerformanceTestSuite) countLines(filename string) int {
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	return lines
}

func (pts *PerformanceTestSuite) GenerateTestFile(lines int) string {
	pts.t.Helper()

	testFile := fmt.Sprintf("%s/test-file-%d.sh", pts.tempDir, lines)

	content := strings.Builder{}
	content.WriteString("#!/bin/bash\n")
	content.WriteString("# Generated performance test file\n\n")

	commands := []string{
		"usacloud server list",
		"usacloud disk list --output-type csv",
		"usacloud server read 123456789",
		"usacloud database list",
		"usacloud invalid-command",
		"usacloud server invalid-subcommand",
		"usacloud iso-image list",
		"echo 'non-usacloud command'",
	}

	for i := 0; i < lines; i++ {
		command := commands[i%len(commands)]
		content.WriteString(fmt.Sprintf("%s  # Line %d\n", command, i+1))
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		pts.t.Fatalf("テストファイル生成エラー: %v", err)
	}

	return testFile
}
