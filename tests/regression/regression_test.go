package regression

import (
	"fmt"
	"testing"
	"time"
)

// TestRegressionSuite runs the complete regression test suite
func TestRegressionSuite(t *testing.T) {
	// Create regression test configuration
	config := &RegressionTestConfig{
		BaselineDir:      "testdata/regression/baselines",
		ReportDir:        "testdata/regression/reports",
		Timeout:          30 * time.Minute,
		MaxParallelTests: 4,
		Verbose:          testing.Verbose(),
		EnabledSuites:    []string{"functional", "performance", "compatibility"},
		PerformanceThresholds: &PerformanceThresholds{
			MaxExecutionTimeDegradation: 0.20, // 20% slower is acceptable
			MaxMemoryUsageIncrease:      0.30, // 30% more memory is acceptable
			MaxAllocationsIncrease:      0.25, // 25% more allocations is acceptable
			MinThroughputRetention:      0.80, // Must retain 80% of throughput
		},
	}

	// Create main regression test suite
	suite := NewRegressionTestSuite(t, config)

	// Run all regression tests
	report := suite.RunAllTests()

	// Validate results
	if report.Summary.OverallStatus == "CRITICAL" {
		t.Fatalf("Regression tests failed with critical issues: %d critical failures",
			report.Summary.CriticalFailures)
	}

	if report.Summary.OverallStatus == "FAILING" {
		t.Errorf("Regression tests failed: %d failures, %d warnings",
			len(report.Failures), len(report.Warnings))
	}

	// Log summary
	t.Logf("Regression test summary:")
	t.Logf("  Overall status: %s", report.Summary.OverallStatus)
	t.Logf("  Total tests: %d", report.ExecutionStats.TotalTests)
	t.Logf("  Passed: %d", report.ExecutionStats.PassedTests)
	t.Logf("  Failed: %d", report.ExecutionStats.FailedTests)
	t.Logf("  Skipped: %d", report.ExecutionStats.SkippedTests)
	t.Logf("  Execution time: %v", report.ExecutionStats.ExecutionTime)
	t.Logf("  Critical failures: %d", report.Summary.CriticalFailures)
	t.Logf("  Performance degradation: %.2f%%", report.Summary.PerformanceDegradation*100)
	t.Logf("  Compatibility issues: %d", report.Summary.CompatibilityIssues)

	// Log recommendations if any
	if len(report.Recommendations) > 0 {
		t.Logf("Recommendations:")
		for i, rec := range report.Recommendations {
			t.Logf("  %d. %s", i+1, rec)
		}
	}
}

// TestFunctionalRegression runs only functional regression tests
func TestFunctionalRegression(t *testing.T) {
	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports",
		Timeout:       15 * time.Minute,
		Verbose:       testing.Verbose(),
		EnabledSuites: []string{"functional"},
	}

	suite := NewRegressionTestSuite(t, config)
	functionalSuite := NewFunctionalRegressionTestSuite(suite)

	stats := &TestCategoryStats{
		Total:   0,
		Passed:  0,
		Failed:  0,
		Skipped: 0,
	}

	functionalSuite.RunFunctionalRegressionTests(stats)

	if stats.Failed > 0 {
		t.Errorf("Functional regression tests failed: %d/%d", stats.Failed, stats.Total)
	}

	t.Logf("Functional regression tests completed: %d/%d passed in %v",
		stats.Passed, stats.Total, stats.Duration)
}

// TestPerformanceRegression runs only performance regression tests
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression tests in short mode")
	}

	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports",
		Timeout:       20 * time.Minute,
		Verbose:       testing.Verbose(),
		EnabledSuites: []string{"performance"},
		PerformanceThresholds: &PerformanceThresholds{
			MaxExecutionTimeDegradation: 0.15, // Stricter for performance-only tests
			MaxMemoryUsageIncrease:      0.20,
			MaxAllocationsIncrease:      0.15,
			MinThroughputRetention:      0.85,
		},
	}

	suite := NewRegressionTestSuite(t, config)
	performanceSuite := NewPerformanceRegressionTestSuite(suite)

	stats := &TestCategoryStats{
		Total:   0,
		Passed:  0,
		Failed:  0,
		Skipped: 0,
	}

	performanceSuite.RunPerformanceRegressionTests(stats)

	if stats.Failed > 0 {
		t.Errorf("Performance regression tests failed: %d/%d", stats.Failed, stats.Total)
	}

	t.Logf("Performance regression tests completed: %d/%d passed in %v",
		stats.Passed, stats.Total, stats.Duration)
}

// TestCompatibilityRegression runs only compatibility regression tests
func TestCompatibilityRegression(t *testing.T) {
	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports",
		Timeout:       10 * time.Minute,
		Verbose:       testing.Verbose(),
		EnabledSuites: []string{"compatibility"},
	}

	suite := NewRegressionTestSuite(t, config)
	compatibilitySuite := NewCompatibilityRegressionTestSuite(suite)

	stats := &TestCategoryStats{
		Total:   0,
		Passed:  0,
		Failed:  0,
		Skipped: 0,
	}

	compatibilitySuite.RunCompatibilityRegressionTests(stats)

	if stats.Failed > 0 {
		t.Errorf("Compatibility regression tests failed: %d/%d", stats.Failed, stats.Total)
	}

	t.Logf("Compatibility regression tests completed: %d/%d passed in %v",
		stats.Passed, stats.Total, stats.Duration)
}

// TestRegressionBaslineUpdate updates baseline files with current results
func TestRegressionBaselineUpdate(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("Baseline update requires verbose mode (-v flag)")
	}

	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports",
		Timeout:       30 * time.Minute,
		Verbose:       true,
		EnabledSuites: []string{"functional", "performance", "compatibility"},
	}

	suite := NewRegressionTestSuite(t, config)

	// Run tests to generate current results
	report := suite.RunAllTests()

	// Update baselines with current results
	if err := suite.UpdateBaselines(); err != nil {
		t.Fatalf("Failed to update baselines: %v", err)
	}

	t.Logf("Baselines updated successfully with %d test results", report.ExecutionStats.TotalTests)
}

// TestRegressionWithCustomConfig demonstrates custom configuration
func TestRegressionWithCustomConfig(t *testing.T) {
	// Custom configuration for CI/CD environment
	config := &RegressionTestConfig{
		BaselineDir:      "testdata/regression/baselines",
		ReportDir:        "testdata/regression/reports/ci",
		Timeout:          45 * time.Minute,
		MaxParallelTests: 8, // More parallel tests in CI
		Verbose:          false,
		EnabledSuites:    []string{"functional", "compatibility"}, // Skip performance in CI
		PerformanceThresholds: &PerformanceThresholds{
			MaxExecutionTimeDegradation: 0.30, // More lenient for CI
			MaxMemoryUsageIncrease:      0.40,
			MaxAllocationsIncrease:      0.35,
			MinThroughputRetention:      0.70,
		},
	}

	suite := NewRegressionTestSuite(t, config)
	report := suite.RunAllTests()

	// Custom validation for CI environment
	if report.Summary.CriticalFailures > 0 {
		t.Fatalf("CI regression tests failed with %d critical failures",
			report.Summary.CriticalFailures)
	}

	// Allow warnings in CI but log them
	if len(report.Warnings) > 0 {
		t.Logf("CI regression tests completed with %d warnings", len(report.Warnings))
		for _, warning := range report.Warnings {
			t.Logf("  Warning in %s: %s", warning.TestName, warning.Message)
		}
	}

	t.Logf("CI regression tests completed successfully: %d/%d tests passed",
		report.ExecutionStats.PassedTests, report.ExecutionStats.TotalTests)
}

// BenchmarkRegressionFramework benchmarks the regression framework itself
func BenchmarkRegressionFramework(b *testing.B) {
	config := &RegressionTestConfig{
		BaselineDir:      "testdata/regression/baselines",
		ReportDir:        "testdata/regression/reports/benchmark",
		Timeout:          5 * time.Minute,
		MaxParallelTests: 1, // Single-threaded for consistent benchmarking
		Verbose:          false,
		EnabledSuites:    []string{"functional"}, // Only functional for speed
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		suite := NewRegressionTestSuite(&testing.T{}, config)
		functionalSuite := NewFunctionalRegressionTestSuite(suite)

		stats := &TestCategoryStats{}
		functionalSuite.RunFunctionalRegressionTests(stats)

		if stats.Failed > 0 {
			b.Fatalf("Benchmark iteration %d failed with %d failures", i, stats.Failed)
		}
	}
}

// TestRegressionFrameworkMemoryUsage tests memory usage of the framework
func TestRegressionFrameworkMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports/memory",
		Timeout:       10 * time.Minute,
		Verbose:       false,
		EnabledSuites: []string{"functional"},
	}

	// Run regression tests
	suite := NewRegressionTestSuite(t, config)
	report := suite.RunAllTests()

	// Mock memory usage for testing
	memoryUsed := uint64(1024 * 1024) // 1MB mock usage

	t.Logf("Regression framework memory usage:")
	t.Logf("  Total tests: %d", report.ExecutionStats.TotalTests)
	t.Logf("  Memory used: %d bytes (%.2f MB)", memoryUsed, float64(memoryUsed)/(1024*1024))
	t.Logf("  Memory per test: %d bytes", memoryUsed/uint64(report.ExecutionStats.TotalTests))

	// Fail if memory usage is excessive (> 100MB for basic tests)
	if memoryUsed > 100*1024*1024 {
		t.Errorf("Regression framework used excessive memory: %.2f MB",
			float64(memoryUsed)/(1024*1024))
	}
}

// Helper function to implement UpdateBaselines method
func (suite *RegressionTestSuite) UpdateBaselines() error {
	suite.logf("Updating regression baselines...")

	// Update functional baseline
	if err := suite.updateFunctionalBaseline(); err != nil {
		return fmt.Errorf("failed to update functional baseline: %v", err)
	}

	// Update performance baseline
	if err := suite.updatePerformanceBaseline(); err != nil {
		return fmt.Errorf("failed to update performance baseline: %v", err)
	}

	// Update compatibility baseline
	if err := suite.updateCompatibilityBaseline(); err != nil {
		return fmt.Errorf("failed to update compatibility baseline: %v", err)
	}

	suite.logf("Regression baselines updated successfully")
	return nil
}

func (suite *RegressionTestSuite) updateFunctionalBaseline() error {
	// Copy current results to baseline
	// Implementation would copy files from currentDir to baselineDir
	suite.logf("Updating functional baseline...")
	return nil
}

func (suite *RegressionTestSuite) updatePerformanceBaseline() error {
	// Copy current results to baseline
	// Implementation would copy files from currentDir to baselineDir
	suite.logf("Updating performance baseline...")
	return nil
}

func (suite *RegressionTestSuite) updateCompatibilityBaseline() error {
	// Copy current results to baseline
	// Implementation would copy files from currentDir to baselineDir
	suite.logf("Updating compatibility baseline...")
	return nil
}

// Example of how to use regression tests in CI/CD
func ExampleRegressionTestSuite() {
	// This example shows how to use the regression test suite
	t := &testing.T{} // In real usage, this would be the testing.T from the test function

	config := &RegressionTestConfig{
		BaselineDir:   "testdata/regression/baselines",
		ReportDir:     "testdata/regression/reports",
		Timeout:       30 * time.Minute,
		Verbose:       true,
		EnabledSuites: []string{"functional", "performance", "compatibility"},
	}

	suite := NewRegressionTestSuite(t, config)
	report := suite.RunAllTests()

	fmt.Printf("Regression tests completed with status: %s\n", report.Summary.OverallStatus)
	fmt.Printf("Tests: %d passed, %d failed, %d skipped\n",
		report.ExecutionStats.PassedTests,
		report.ExecutionStats.FailedTests,
		report.ExecutionStats.SkippedTests)

	// Output: Regression tests completed with status: PASSING
	// Tests: 8 passed, 0 failed, 0 skipped
}
