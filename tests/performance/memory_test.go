package performance

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/transform"
)

func TestMemoryUsage_LargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("メモリテストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	testCases := []struct {
		name        string
		lines       int
		maxMemoryMB float64
		maxGCCount  uint32
	}{
		{
			name:        "Medium file memory usage",
			lines:       1000,
			maxMemoryMB: 50,
			maxGCCount:  5,
		},
		{
			name:        "Large file memory usage",
			lines:       10000,
			maxMemoryMB: 200,
			maxGCCount:  20,
		},
		{
			name:        "Huge file memory usage",
			lines:       50000,
			maxMemoryMB: 500,
			maxGCCount:  50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile := suite.GenerateTestFile(tc.lines)
			defer os.Remove(inputFile)

			metrics := suite.MeasurePerformance(tc.name, inputFile, func() *ProcessingResult {
				engine := transform.NewDefaultEngine()
				return processFile(engine, inputFile)
			})

			if metrics.PeakMemoryMB > tc.maxMemoryMB {
				t.Errorf("メモリ使用量が制限超過: %.2fMB > %.2fMB",
					metrics.PeakMemoryMB, tc.maxMemoryMB)
			}

			if metrics.GCCount > tc.maxGCCount {
				t.Logf("GC回数が多い: %d > %d", metrics.GCCount, tc.maxGCCount)
			}

			if metrics.MemoryEfficiency < 50 {
				t.Logf("メモリ効率が低い: %.2f lines/MB", metrics.MemoryEfficiency)
			}

			t.Logf("%s メモリ使用状況:", tc.name)
			t.Logf("  初期メモリ: %.2f MB", metrics.InitialMemoryMB)
			t.Logf("  ピークメモリ: %.2f MB", metrics.PeakMemoryMB)
			t.Logf("  最終メモリ: %.2f MB", metrics.FinalMemoryMB)
			t.Logf("  メモリ効率: %.2f lines/MB", metrics.MemoryEfficiency)
			t.Logf("  GC回数: %d", metrics.GCCount)
			t.Logf("  GC時間: %v", metrics.GCPauseTime)
		})
	}
}

func TestMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("メモリリークテストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	var memorySnapshots []float64
	const iterations = 10

	for i := 0; i < iterations; i++ {
		inputFile := suite.GenerateTestFile(1000)

		engine := transform.NewDefaultEngine()
		processFile(engine, inputFile)

		var memStats runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memStats)
		memorySnapshots = append(memorySnapshots, float64(memStats.Alloc)/1024/1024)

		os.Remove(inputFile)

		time.Sleep(10 * time.Millisecond)
	}

	firstSnapshot := memorySnapshots[0]
	lastSnapshot := memorySnapshots[len(memorySnapshots)-1]

	memoryGrowth := (lastSnapshot - firstSnapshot) / firstSnapshot
	if memoryGrowth > 0.2 {
		t.Logf("メモリリークの可能性: %.2f%% 増加", memoryGrowth*100)
		for i, snapshot := range memorySnapshots {
			t.Logf("  実行 %d: %.2f MB", i+1, snapshot)
		}
	} else {
		t.Logf("メモリリークなし: %.2f%% 変化", memoryGrowth*100)
	}
}

func TestMemoryProfile_Processing(t *testing.T) {
	if testing.Short() {
		t.Skip("メモリプロファイルテストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	inputFile := suite.GenerateTestFile(5000)
	defer os.Remove(inputFile)

	var beforeStats, afterStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&beforeStats)

	engine := transform.NewDefaultEngine()
	result := processFile(engine, inputFile)

	runtime.ReadMemStats(&afterStats)

	allocatedMB := float64(afterStats.Alloc-beforeStats.Alloc) / 1024 / 1024
	totalAllocMB := float64(afterStats.TotalAlloc-beforeStats.TotalAlloc) / 1024 / 1024

	t.Logf("メモリプロファイル:")
	t.Logf("  処理行数: %d", result.ProcessedLines)
	t.Logf("  変換行数: %d", result.TransformedLines)
	t.Logf("  現在のメモリ増加: %.2f MB", allocatedMB)
	t.Logf("  総メモリ使用量: %.2f MB", totalAllocMB)
	t.Logf("  行あたりメモリ: %.6f MB/line", allocatedMB/float64(result.ProcessedLines))
	t.Logf("  Heap objects: %d", afterStats.HeapObjects)
	t.Logf("  GC回数: %d", afterStats.NumGC-beforeStats.NumGC)

	memoryPerLine := allocatedMB / float64(result.ProcessedLines)
	if memoryPerLine > 0.01 {
		t.Errorf("行あたりメモリ使用量が高い: %.6f MB/line", memoryPerLine)
	}
}

func TestMemoryEfficiency_EngineReuse(t *testing.T) {
	suite := NewPerformanceTestSuite(t)

	testFiles := make([]string, 5)
	for i := range testFiles {
		testFiles[i] = suite.GenerateTestFile(500)
		defer os.Remove(testFiles[i])
	}

	var beforeStats, afterStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&beforeStats)

	engine := transform.NewDefaultEngine()

	var totalProcessed int
	for _, file := range testFiles {
		result := processFile(engine, file)
		totalProcessed += result.ProcessedLines
	}

	runtime.ReadMemStats(&afterStats)

	allocatedMB := float64(afterStats.Alloc-beforeStats.Alloc) / 1024 / 1024
	memoryPerLine := allocatedMB / float64(totalProcessed)

	t.Logf("エンジン再利用効率:")
	t.Logf("  処理ファイル数: %d", len(testFiles))
	t.Logf("  総処理行数: %d", totalProcessed)
	t.Logf("  総メモリ使用: %.2f MB", allocatedMB)
	t.Logf("  行あたりメモリ: %.6f MB/line", memoryPerLine)

	if memoryPerLine > 0.005 {
		t.Logf("エンジン再利用時のメモリ効率が改善の余地あり")
	}
}
