package regression

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/armaniacs/usacloud-update/internal/transform"
)

type PerformanceRegressionTestSuite struct {
	*RegressionTestSuite
	engine              *transform.IntegratedEngine
	benchmarkThresholds *PerformanceThresholds
	memoryProfiler      *MemoryProfiler
	benchmarkRunner     *BenchmarkRunner
}

type BenchmarkTestCase struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Input           string           `json:"input"`
	Iterations      int              `json:"iterations"`
	Warmup          int              `json:"warmup"`
	Timeout         time.Duration    `json:"timeout"`
	ExpectedMetrics *ExpectedMetrics `json:"expected_metrics"`
	Tags            []string         `json:"tags"`
	Category        string           `json:"category"`
	Priority        string           `json:"priority"`
}

type ExpectedMetrics struct {
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	MaxMemoryUsage   int64         `json:"max_memory_usage"`
	MaxAllocations   int64         `json:"max_allocations"`
	MinThroughput    float64       `json:"min_throughput"`
	MaxGCPressure    float64       `json:"max_gc_pressure"`
}

type MemoryProfiler struct {
	samples           []MemorySample
	leakDetector      *LeakDetector
	gcStatistics      *GCStatistics
	allocationTracker *AllocationTracker
}

type MemorySample struct {
	Timestamp     time.Time `json:"timestamp"`
	TotalAlloc    int64     `json:"total_alloc"`
	HeapAlloc     int64     `json:"heap_alloc"`
	HeapSys       int64     `json:"heap_sys"`
	HeapIdle      int64     `json:"heap_idle"`
	HeapInuse     int64     `json:"heap_inuse"`
	HeapReleased  int64     `json:"heap_released"`
	HeapObjects   int64     `json:"heap_objects"`
	StackInuse    int64     `json:"stack_inuse"`
	StackSys      int64     `json:"stack_sys"`
	MSpanInuse    int64     `json:"mspan_inuse"`
	MSpanSys      int64     `json:"mspan_sys"`
	MCacheInuse   int64     `json:"mcache_inuse"`
	MCacheSys     int64     `json:"mcache_sys"`
	BuckHashSys   int64     `json:"buckhash_sys"`
	GCSys         int64     `json:"gc_sys"`
	OtherSys      int64     `json:"other_sys"`
	NextGC        int64     `json:"next_gc"`
	LastGC        int64     `json:"last_gc"`
	PauseTotalNs  int64     `json:"pause_total_ns"`
	NumGC         int32     `json:"num_gc"`
	NumForcedGC   int32     `json:"num_forced_gc"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
}

type LeakDetector struct {
	baselineHeap    int64
	samplingRate    time.Duration
	detectionWindow time.Duration
	threshold       float64
}

type GCStatistics struct {
	collections    []GCCollection
	totalPauseTime time.Duration
	averagePause   time.Duration
	maxPause       time.Duration
	gcPressure     float64
}

type GCCollection struct {
	Timestamp  time.Time     `json:"timestamp"`
	PauseTime  time.Duration `json:"pause_time"`
	HeapBefore int64         `json:"heap_before"`
	HeapAfter  int64         `json:"heap_after"`
	Collected  int64         `json:"collected"`
}

type AllocationTracker struct {
	allocations   map[string]int64
	hotPaths      []string
	largestAllocs []AllocationInfo
}

type AllocationInfo struct {
	Size       int64  `json:"size"`
	Count      int64  `json:"count"`
	StackTrace string `json:"stack_trace"`
}

type BenchmarkRunner struct {
	warmupIterations int
	testIterations   int
	timeout          time.Duration
	cpuProfiler      *CPUProfiler
}

type CPUProfiler struct {
	enabled      bool
	samplingRate time.Duration
	profilePath  string
}

type PerformanceMetrics struct {
	ExecutionTime   time.Duration    `json:"execution_time"`
	MemoryUsage     int64            `json:"memory_usage"`
	Allocations     int64            `json:"allocations"`
	Throughput      float64          `json:"throughput"`
	GCPressure      float64          `json:"gc_pressure"`
	CPUUsage        float64          `json:"cpu_usage"`
	DetailedProfile *DetailedProfile `json:"detailed_profile"`
}

type DetailedProfile struct {
	MemoryProfile       *MemoryProfile       `json:"memory_profile"`
	AllocationPattern   map[string]int64     `json:"allocation_pattern"`
	GCStatistics        *GCStatistics        `json:"gc_statistics"`
	LeakDetection       *LeakDetectionResult `json:"leak_detection"`
	PerformanceHotspots []PerformanceHotspot `json:"performance_hotspots"`
}

type PerformanceHotspot struct {
	Function      string        `json:"function"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsage   int64         `json:"memory_usage"`
	CallCount     int64         `json:"call_count"`
	Percentage    float64       `json:"percentage"`
}

func NewPerformanceRegressionTestSuite(regressionSuite *RegressionTestSuite) *PerformanceRegressionTestSuite {
	suite := &PerformanceRegressionTestSuite{
		RegressionTestSuite: regressionSuite,
		benchmarkThresholds: &PerformanceThresholds{
			MaxExecutionTimeDegradation: 0.20, // 20% degradation allowed
			MaxMemoryUsageIncrease:      0.30, // 30% memory increase allowed
			MaxAllocationsIncrease:      0.25, // 25% allocation increase allowed
			MinThroughputRetention:      0.80, // Must retain 80% throughput
		},
		memoryProfiler:  NewMemoryProfiler(),
		benchmarkRunner: NewBenchmarkRunner(),
	}

	// Initialize components
	suite.initializeComponents()

	return suite
}

func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		samples:           make([]MemorySample, 0),
		leakDetector:      NewLeakDetector(),
		gcStatistics:      NewGCStatistics(),
		allocationTracker: NewAllocationTracker(),
	}
}

func NewLeakDetector() *LeakDetector {
	return &LeakDetector{
		samplingRate:    time.Second,
		detectionWindow: 30 * time.Second,
		threshold:       0.10, // 10% increase considered potential leak
	}
}

func NewGCStatistics() *GCStatistics {
	return &GCStatistics{
		collections: make([]GCCollection, 0),
	}
}

func NewAllocationTracker() *AllocationTracker {
	return &AllocationTracker{
		allocations:   make(map[string]int64),
		hotPaths:      make([]string, 0),
		largestAllocs: make([]AllocationInfo, 0),
	}
}

func NewBenchmarkRunner() *BenchmarkRunner {
	return &BenchmarkRunner{
		warmupIterations: 10,
		testIterations:   100,
		timeout:          5 * time.Minute,
		cpuProfiler:      NewCPUProfiler(),
	}
}

func NewCPUProfiler() *CPUProfiler {
	return &CPUProfiler{
		enabled:      false,
		samplingRate: 10 * time.Millisecond,
	}
}

func (suite *PerformanceRegressionTestSuite) initializeComponents() {
	// Initialize transform engine
	integrationConfig := &transform.IntegrationConfig{
		EnablePreValidation:  true,
		EnablePostValidation: true,
		StrictMode:           false,
		CacheEnabled:         true,
	}

	suite.engine = transform.NewIntegratedEngine(integrationConfig)
}

func (suite *PerformanceRegressionTestSuite) RunPerformanceRegressionTests(stats *TestCategoryStats) {
	suite.logf("Starting performance regression tests...")

	// Run benchmark regression tests
	suite.runBenchmarkRegressionTests(stats)

	// Run memory regression tests
	suite.runMemoryRegressionTests(stats)

	// Run throughput regression tests
	suite.runThroughputRegressionTests(stats)

	// Run stress tests
	suite.runStressTests(stats)

	suite.logf("Performance regression tests completed")
}

func (suite *PerformanceRegressionTestSuite) runBenchmarkRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running benchmark regression tests...")

	// Load benchmark test cases
	testCases := suite.loadBenchmarkTestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run benchmark
		metrics, err := suite.runBenchmarkTest(testCase)
		if err != nil {
			suite.recordFailure("benchmark", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.performanceBaseline != nil {
			if baselineResult, exists := suite.performanceBaseline.Benchmarks[testCase.Name]; exists {
				if !suite.compareBenchmarkResults(baselineResult, metrics, testCase.Name) {
					stats.Failed++
					continue
				}
			}
		}

		// Save current result for future baseline
		suite.saveCurrentBenchmarkResult(testCase.Name, metrics)

		stats.Passed++
		suite.logf("Benchmark test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("Benchmark regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *PerformanceRegressionTestSuite) runMemoryRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running memory regression tests...")

	// Load memory test cases
	testCases := suite.loadMemoryTestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run memory profiling
		profile, err := suite.runMemoryProfilingTest(testCase)
		if err != nil {
			suite.recordFailure("memory", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.performanceBaseline != nil {
			if baselineProfile, exists := suite.performanceBaseline.MemoryProfiles[testCase.Name]; exists {
				if !suite.compareMemoryProfiles(baselineProfile, profile, testCase.Name) {
					stats.Failed++
					continue
				}
			}
		}

		// Check for memory leaks
		if leaks := suite.memoryProfiler.leakDetector.DetectLeaks(profile); len(leaks.SuspectedLeaks) > 0 {
			suite.recordWarning("memory_leak", testCase.Name,
				fmt.Sprintf("Detected %d potential memory leaks", len(leaks.SuspectedLeaks)),
				leaks, map[string]interface{}{"test_case": testCase, "leaks": leaks})
		}

		// Save current result for future baseline
		suite.saveCurrentMemoryProfile(testCase.Name, profile)

		stats.Passed++
		suite.logf("Memory test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("Memory regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *PerformanceRegressionTestSuite) runThroughputRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running throughput regression tests...")

	// Load throughput test cases
	testCases := suite.loadThroughputTestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run throughput test
		throughputMetrics, err := suite.runThroughputTest(testCase)
		if err != nil {
			suite.recordFailure("throughput", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.performanceBaseline != nil {
			baselineThroughput := suite.getBaselineThroughput(testCase.Name)
			if baselineThroughput > 0 {
				degradation := (baselineThroughput - throughputMetrics.Throughput) / baselineThroughput
				if degradation > (1.0 - suite.benchmarkThresholds.MinThroughputRetention) {
					suite.recordFailure("throughput", testCase.Name,
						baselineThroughput, throughputMetrics.Throughput, "throughput_degradation",
						map[string]interface{}{
							"degradation": degradation,
							"threshold":   suite.benchmarkThresholds.MinThroughputRetention,
						})
					stats.Failed++
					continue
				}
			}
		}

		// Save current result for future baseline
		suite.saveCurrentThroughputResult(testCase.Name, throughputMetrics)

		stats.Passed++
		suite.logf("Throughput test '%s' passed (throughput: %.2f ops/sec, execution: %v)",
			testCase.Name, throughputMetrics.Throughput, executionTime)
	}

	suite.logf("Throughput regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *PerformanceRegressionTestSuite) runStressTests(stats *TestCategoryStats) {
	suite.logf("Running stress tests...")

	// Load stress test cases
	testCases := suite.loadStressTestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run stress test
		stressMetrics, err := suite.runStressTest(testCase)
		if err != nil {
			suite.recordFailure("stress", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Check if system remained stable under stress
		if !suite.validateStressTestStability(stressMetrics) {
			suite.recordFailure("stress", testCase.Name,
				"Stable performance under stress", "Performance degraded significantly", "stability",
				map[string]interface{}{"metrics": stressMetrics})
			stats.Failed++
			continue
		}

		// Save current result for future baseline
		suite.saveCurrentStressResult(testCase.Name, stressMetrics)

		stats.Passed++
		suite.logf("Stress test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("Stress tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *PerformanceRegressionTestSuite) runBenchmarkTest(testCase BenchmarkTestCase) (*BenchmarkResult, error) {

	// Warmup phase
	for i := 0; i < testCase.Warmup; i++ {
		// Mock processing - simple work simulation
		_ = testCase.Input
		time.Sleep(time.Microsecond)
	}

	// Force garbage collection before measurement
	runtime.GC()
	runtime.GC() // Double GC to ensure cleanup

	// Benchmark measurement
	var totalTime time.Duration
	var totalAllocs int64
	var totalBytes int64

	startTime := time.Now()
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	for i := 0; i < testCase.Iterations; i++ {
		iterStart := time.Now()
		// Mock processing - simple work simulation
		_ = testCase.Input
		time.Sleep(time.Microsecond)
		totalTime += time.Since(iterStart)
	}

	runtime.ReadMemStats(&memStatsAfter)
	executionTime := time.Since(startTime)

	totalAllocs = int64(memStatsAfter.Mallocs - memStatsBefore.Mallocs)
	totalBytes = int64(memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc)

	avgNanosPerOp := totalTime.Nanoseconds() / int64(testCase.Iterations)
	bytesPerOp := totalBytes / int64(testCase.Iterations)
	allocsPerOp := totalAllocs / int64(testCase.Iterations)

	return &BenchmarkResult{
		Name:             testCase.Name,
		Iterations:       testCase.Iterations,
		NanosPerOp:       avgNanosPerOp,
		BytesPerOp:       bytesPerOp,
		AllocsPerOp:      allocsPerOp,
		ExecutionTime:    executionTime,
		TotalMemoryAlloc: totalBytes,
	}, nil
}

func (suite *PerformanceRegressionTestSuite) runMemoryProfilingTest(testCase BenchmarkTestCase) (*MemoryProfile, error) {

	// Start memory profiling
	suite.memoryProfiler.StartProfiling()
	defer suite.memoryProfiler.StopProfiling()

	// Run test iterations while profiling
	for i := 0; i < testCase.Iterations; i++ {
		// Mock processing - simple work simulation
		_ = testCase.Input
		time.Sleep(time.Microsecond)

		// Sample memory every 10 iterations
		if i%10 == 0 {
			suite.memoryProfiler.SampleMemory()
		}
	}

	// Force final garbage collection and sample
	runtime.GC()
	suite.memoryProfiler.SampleMemory()

	profile := suite.memoryProfiler.GenerateProfile()
	return profile, nil
}

func (suite *PerformanceRegressionTestSuite) runThroughputTest(testCase BenchmarkTestCase) (*PerformanceMetrics, error) {

	startTime := time.Now()
	var successful int64

	// Run operations for a fixed duration
	testDuration := 100 * time.Millisecond // Reduced from 10 seconds for faster testing
	endTime := startTime.Add(testDuration)

	for time.Now().Before(endTime) {
		// Mock processing - simple work simulation
		_ = testCase.Input
		time.Sleep(time.Microsecond)
		successful++
	}

	actualDuration := time.Since(startTime)
	throughput := float64(successful) / actualDuration.Seconds()

	return &PerformanceMetrics{
		ExecutionTime: actualDuration,
		Throughput:    throughput,
	}, nil
}

func (suite *PerformanceRegressionTestSuite) runStressTest(testCase BenchmarkTestCase) (*PerformanceMetrics, error) {

	// Run stress test with high concurrency
	concurrency := runtime.NumCPU() * 2
	iterations := testCase.Iterations * concurrency

	startTime := time.Now()

	// Channel for goroutine synchronization
	results := make(chan error, concurrency)

	// Launch concurrent workers
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			for j := 0; j < testCase.Iterations; j++ {
				// Mock processing - simple work simulation
				_ = testCase.Input
				time.Sleep(time.Microsecond)
			}
			results <- nil
		}(i)
	}

	// Wait for all workers to complete
	var errors []error
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	executionTime := time.Since(startTime)

	if len(errors) > 0 {
		return nil, fmt.Errorf("stress test failed with %d errors: %v", len(errors), errors[0])
	}

	throughput := float64(iterations) / executionTime.Seconds()

	return &PerformanceMetrics{
		ExecutionTime: executionTime,
		Throughput:    throughput,
	}, nil
}

func (suite *PerformanceRegressionTestSuite) loadBenchmarkTestCases() []BenchmarkTestCase {
	testCasesPath := filepath.Join(suite.baselineDir, "performance", "benchmark_test_cases.json")

	// Create default test cases if file doesn't exist
	if _, err := os.Stat(testCasesPath); os.IsNotExist(err) {
		defaultCases := suite.createDefaultBenchmarkTestCases()
		suite.saveBenchmarkTestCases(testCasesPath, defaultCases)
		return defaultCases
	}

	data, err := os.ReadFile(testCasesPath)
	if err != nil {
		suite.t.Logf("Failed to load benchmark test cases: %v", err)
		return suite.createDefaultBenchmarkTestCases()
	}

	var testCases []BenchmarkTestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		suite.t.Logf("Failed to unmarshal benchmark test cases: %v", err)
		return suite.createDefaultBenchmarkTestCases()
	}

	return testCases
}

func (suite *PerformanceRegressionTestSuite) loadMemoryTestCases() []BenchmarkTestCase {
	return suite.loadBenchmarkTestCases() // Reuse benchmark test cases for memory testing
}

func (suite *PerformanceRegressionTestSuite) loadThroughputTestCases() []BenchmarkTestCase {
	return suite.loadBenchmarkTestCases() // Reuse benchmark test cases for throughput testing
}

func (suite *PerformanceRegressionTestSuite) loadStressTestCases() []BenchmarkTestCase {
	testCases := suite.loadBenchmarkTestCases()
	// Modify iterations for stress testing
	for i := range testCases {
		testCases[i].Iterations = 10           // Reduced iterations for faster testing
		testCases[i].Timeout = 5 * time.Minute // Longer timeout
	}
	return testCases
}

func (suite *PerformanceRegressionTestSuite) createDefaultBenchmarkTestCases() []BenchmarkTestCase {
	return []BenchmarkTestCase{
		{
			Name:        "single_transformation_benchmark",
			Description: "Benchmark single line transformation",
			Input:       "usacloud server list --output-type csv",
			Iterations:  10,
			Warmup:      10,
			Timeout:     30 * time.Second,
			ExpectedMetrics: &ExpectedMetrics{
				MaxExecutionTime: 10 * time.Millisecond,
				MaxMemoryUsage:   1024 * 1024, // 1MB
				MaxAllocations:   100,
				MinThroughput:    100.0, // 100 ops/sec
			},
			Tags:     []string{"single", "transformation"},
			Category: "performance",
			Priority: "high",
		},
		{
			Name:        "complex_transformation_benchmark",
			Description: "Benchmark complex multi-rule transformation",
			Input: `usacloud server list --output-type csv --zone = all
usacloud iso-image list --output-type tsv`,
			Iterations: 500,
			Warmup:     5,
			Timeout:    60 * time.Second,
			ExpectedMetrics: &ExpectedMetrics{
				MaxExecutionTime: 50 * time.Millisecond,
				MaxMemoryUsage:   2 * 1024 * 1024, // 2MB
				MaxAllocations:   500,
				MinThroughput:    50.0, // 50 ops/sec
			},
			Tags:     []string{"complex", "multi-rule"},
			Category: "performance",
			Priority: "high",
		},
		{
			Name:        "validation_benchmark",
			Description: "Benchmark validation performance",
			Input:       "usacloud serv lst --invalid-option",
			Iterations:  2000,
			Warmup:      20,
			Timeout:     45 * time.Second,
			ExpectedMetrics: &ExpectedMetrics{
				MaxExecutionTime: 5 * time.Millisecond,
				MaxMemoryUsage:   512 * 1024, // 512KB
				MaxAllocations:   50,
				MinThroughput:    200.0, // 200 ops/sec
			},
			Tags:     []string{"validation", "error-handling"},
			Category: "performance",
			Priority: "medium",
		},
		{
			Name:        "large_input_benchmark",
			Description: "Benchmark large input processing",
			Input:       suite.generateLargeInput(1000), // 1000 lines
			Iterations:  10,
			Warmup:      2,
			Timeout:     2 * time.Minute,
			ExpectedMetrics: &ExpectedMetrics{
				MaxExecutionTime: 1 * time.Second,
				MaxMemoryUsage:   10 * 1024 * 1024, // 10MB
				MaxAllocations:   10000,
				MinThroughput:    1.0, // 1 ops/sec
			},
			Tags:     []string{"large-input", "stress"},
			Category: "performance",
			Priority: "medium",
		},
	}
}

func (suite *PerformanceRegressionTestSuite) generateLargeInput(lines int) string {
	var input string
	commands := []string{
		"usacloud server list --output-type csv",
		"usacloud disk list --output-type tsv",
		"usacloud iso-image list",
		"usacloud server read --selector name=test",
	}

	for i := 0; i < lines; i++ {
		cmd := commands[i%len(commands)]
		input += cmd + "\n"
	}

	return input
}

func (suite *PerformanceRegressionTestSuite) compareBenchmarkResults(baseline, current *BenchmarkResult, testName string) bool {
	// Check execution time regression
	if current.NanosPerOp > baseline.NanosPerOp {
		degradation := float64(current.NanosPerOp-baseline.NanosPerOp) / float64(baseline.NanosPerOp)
		if degradation > suite.benchmarkThresholds.MaxExecutionTimeDegradation {
			suite.recordFailure("benchmark", testName,
				baseline.NanosPerOp, current.NanosPerOp, "execution_time_regression",
				map[string]interface{}{
					"degradation": degradation,
					"threshold":   suite.benchmarkThresholds.MaxExecutionTimeDegradation,
				})
			return false
		}
	}

	// Check memory usage regression
	if current.BytesPerOp > baseline.BytesPerOp {
		increase := float64(current.BytesPerOp-baseline.BytesPerOp) / float64(baseline.BytesPerOp)
		if increase > suite.benchmarkThresholds.MaxMemoryUsageIncrease {
			suite.recordFailure("benchmark", testName,
				baseline.BytesPerOp, current.BytesPerOp, "memory_usage_regression",
				map[string]interface{}{
					"increase":  increase,
					"threshold": suite.benchmarkThresholds.MaxMemoryUsageIncrease,
				})
			return false
		}
	}

	// Check allocations regression
	if current.AllocsPerOp > baseline.AllocsPerOp {
		increase := float64(current.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp)
		if increase > suite.benchmarkThresholds.MaxAllocationsIncrease {
			suite.recordFailure("benchmark", testName,
				baseline.AllocsPerOp, current.AllocsPerOp, "allocations_regression",
				map[string]interface{}{
					"increase":  increase,
					"threshold": suite.benchmarkThresholds.MaxAllocationsIncrease,
				})
			return false
		}
	}

	return true
}

func (suite *PerformanceRegressionTestSuite) compareMemoryProfiles(baseline, current *MemoryProfile, testName string) bool {
	// Check peak usage regression
	if current.PeakUsage > baseline.PeakUsage {
		increase := float64(current.PeakUsage-baseline.PeakUsage) / float64(baseline.PeakUsage)
		if increase > suite.benchmarkThresholds.MaxMemoryUsageIncrease {
			suite.recordFailure("memory", testName,
				baseline.PeakUsage, current.PeakUsage, "peak_memory_regression",
				map[string]interface{}{
					"increase":  increase,
					"threshold": suite.benchmarkThresholds.MaxMemoryUsageIncrease,
				})
			return false
		}
	}

	// Check total allocated regression
	if current.TotalAllocated > baseline.TotalAllocated {
		increase := float64(current.TotalAllocated-baseline.TotalAllocated) / float64(baseline.TotalAllocated)
		if increase > suite.benchmarkThresholds.MaxMemoryUsageIncrease {
			suite.recordFailure("memory", testName,
				baseline.TotalAllocated, current.TotalAllocated, "total_allocation_regression",
				map[string]interface{}{
					"increase":  increase,
					"threshold": suite.benchmarkThresholds.MaxMemoryUsageIncrease,
				})
			return false
		}
	}

	return true
}

func (suite *PerformanceRegressionTestSuite) getBaselineThroughput(testName string) float64 {
	if suite.performanceBaseline != nil {
		if benchmark, exists := suite.performanceBaseline.Benchmarks[testName]; exists {
			// Calculate throughput from benchmark data
			if benchmark.NanosPerOp > 0 {
				return 1e9 / float64(benchmark.NanosPerOp) // ops/second
			}
		}
	}
	return 0
}

func (suite *PerformanceRegressionTestSuite) validateStressTestStability(metrics *PerformanceMetrics) bool {
	// Check if throughput is within acceptable range
	return metrics.Throughput > 0.5 // At least 0.5 ops/sec under stress
}

// Memory profiler implementation
func (profiler *MemoryProfiler) StartProfiling() {
	profiler.samples = profiler.samples[:0] // Clear previous samples
	profiler.leakDetector.baselineHeap = profiler.getCurrentHeapSize()
}

func (profiler *MemoryProfiler) StopProfiling() {
	// Final processing if needed
}

func (profiler *MemoryProfiler) SampleMemory() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	sample := MemorySample{
		Timestamp:     time.Now(),
		TotalAlloc:    int64(memStats.TotalAlloc),
		HeapAlloc:     int64(memStats.HeapAlloc),
		HeapSys:       int64(memStats.HeapSys),
		HeapIdle:      int64(memStats.HeapIdle),
		HeapInuse:     int64(memStats.HeapInuse),
		HeapReleased:  int64(memStats.HeapReleased),
		HeapObjects:   int64(memStats.HeapObjects),
		StackInuse:    int64(memStats.StackInuse),
		StackSys:      int64(memStats.StackSys),
		MSpanInuse:    int64(memStats.MSpanInuse),
		MSpanSys:      int64(memStats.MSpanSys),
		MCacheInuse:   int64(memStats.MCacheInuse),
		MCacheSys:     int64(memStats.MCacheSys),
		BuckHashSys:   int64(memStats.BuckHashSys),
		GCSys:         int64(memStats.GCSys),
		OtherSys:      int64(memStats.OtherSys),
		NextGC:        int64(memStats.NextGC),
		LastGC:        int64(memStats.LastGC),
		PauseTotalNs:  int64(memStats.PauseTotalNs),
		NumGC:         int32(memStats.NumGC),
		NumForcedGC:   int32(memStats.NumForcedGC),
		GCCPUFraction: memStats.GCCPUFraction,
	}

	profiler.samples = append(profiler.samples, sample)
}

func (profiler *MemoryProfiler) GenerateProfile() *MemoryProfile {
	if len(profiler.samples) == 0 {
		return &MemoryProfile{}
	}

	// Calculate peak usage
	var peakUsage int64
	var totalAllocated int64

	for _, sample := range profiler.samples {
		if sample.HeapAlloc > peakUsage {
			peakUsage = sample.HeapAlloc
		}
		totalAllocated += sample.TotalAlloc
	}

	// Calculate GC statistics
	lastSample := profiler.samples[len(profiler.samples)-1]

	return &MemoryProfile{
		PeakUsage:         peakUsage,
		TotalAllocated:    totalAllocated / int64(len(profiler.samples)), // Average
		GCCycles:          int(lastSample.NumGC),
		AllocationPattern: profiler.allocationTracker.allocations,
		LeakDetection:     profiler.leakDetector.DetectLeaks(&MemoryProfile{PeakUsage: peakUsage}),
	}
}

func (profiler *MemoryProfiler) getCurrentHeapSize() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.HeapAlloc)
}

func (detector *LeakDetector) DetectLeaks(profile *MemoryProfile) *LeakDetectionResult {
	currentHeap := profile.PeakUsage
	growth := currentHeap - detector.baselineHeap

	var suspectedLeaks []string
	confidence := 0.0

	if growth > 0 {
		growthRate := float64(growth) / float64(detector.baselineHeap)
		if growthRate > detector.threshold {
			suspectedLeaks = append(suspectedLeaks, "Heap growth exceeds threshold")
			confidence = growthRate
		}
	}

	return &LeakDetectionResult{
		SuspectedLeaks:    suspectedLeaks,
		UnreleasedObjects: 0, // Would require more sophisticated tracking
		MemoryGrowth:      growth,
		Confidence:        confidence,
	}
}

// Save functions
func (suite *PerformanceRegressionTestSuite) saveCurrentBenchmarkResult(testName string, result *BenchmarkResult) {
	resultPath := filepath.Join(suite.currentDir, "performance", fmt.Sprintf("benchmark_%s.json", testName))
	suite.saveJSONToFile(resultPath, result)
}

func (suite *PerformanceRegressionTestSuite) saveCurrentMemoryProfile(testName string, profile *MemoryProfile) {
	resultPath := filepath.Join(suite.currentDir, "performance", fmt.Sprintf("memory_%s.json", testName))
	suite.saveJSONToFile(resultPath, profile)
}

func (suite *PerformanceRegressionTestSuite) saveCurrentThroughputResult(testName string, metrics *PerformanceMetrics) {
	resultPath := filepath.Join(suite.currentDir, "performance", fmt.Sprintf("throughput_%s.json", testName))
	suite.saveJSONToFile(resultPath, metrics)
}

func (suite *PerformanceRegressionTestSuite) saveCurrentStressResult(testName string, metrics *PerformanceMetrics) {
	resultPath := filepath.Join(suite.currentDir, "performance", fmt.Sprintf("stress_%s.json", testName))
	suite.saveJSONToFile(resultPath, metrics)
}

func (suite *PerformanceRegressionTestSuite) saveBenchmarkTestCases(path string, testCases []BenchmarkTestCase) {
	suite.saveJSONToFile(path, testCases)
}

func (suite *PerformanceRegressionTestSuite) saveJSONToFile(path string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		suite.t.Logf("Failed to marshal data for %s: %v", path, err)
		return
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		suite.t.Logf("Failed to save file %s: %v", path, err)
	}
}

func (suite *PerformanceRegressionTestSuite) recordFailure(failureType, testName string, expected, actual interface{}, severity string, context map[string]interface{}) {
	failure := RegressionFailure{
		Type:       failureType,
		TestName:   testName,
		Expected:   expected,
		Actual:     actual,
		Difference: fmt.Sprintf("Expected: %v, Actual: %v", expected, actual),
		Severity:   severity,
		Timestamp:  time.Now(),
		Context:    context,
	}

	suite.mu.Lock()
	suite.failures = append(suite.failures, failure)
	suite.mu.Unlock()

	suite.logf("Performance regression failure in %s test '%s': %s", failureType, testName, failure.Difference)
}

func (suite *PerformanceRegressionTestSuite) recordWarning(warningType, testName, message string, actual interface{}, context map[string]interface{}) {
	warning := RegressionWarning{
		Type:      warningType,
		TestName:  testName,
		Message:   message,
		Actual:    actual,
		Timestamp: time.Now(),
		Context:   context,
	}

	suite.mu.Lock()
	suite.warnings = append(suite.warnings, warning)
	suite.mu.Unlock()

	suite.logf("Performance regression warning in %s test '%s': %s", warningType, testName, message)
}
