package transform

import (
	"strings"
	"testing"
)

// Phase 2 Coverage Improvement Tests - integrated_engine.go

func TestNewDefaultIntegrationConfig(t *testing.T) {
	config := NewDefaultIntegrationConfig()
	if config == nil {
		t.Error("Expected config to be created, got nil")
	}

	// Test default values
	if !config.EnablePreValidation {
		t.Error("Expected EnablePreValidation to be true by default")
	}
	if !config.EnablePostValidation {
		t.Error("Expected EnablePostValidation to be true by default")
	}
	if !config.EnableRuleConflictCheck {
		t.Error("Expected EnableRuleConflictCheck to be true by default")
	}
	if config.StrictMode {
		t.Error("Expected StrictMode to be false by default")
	}
	if config.ValidationPriority != PriorityBalanced {
		t.Errorf("Expected ValidationPriority to be PriorityBalanced, got %v", config.ValidationPriority)
	}
	if config.PerformanceMode {
		t.Error("Expected PerformanceMode to be false by default")
	}
	if config.ParallelMode {
		t.Error("Expected ParallelMode to be false by default")
	}
	if config.BatchSize != 100 {
		t.Errorf("Expected BatchSize to be 100, got %d", config.BatchSize)
	}
	if !config.CacheEnabled {
		t.Error("Expected CacheEnabled to be true by default")
	}
}

func TestNewIntegratedEngine(t *testing.T) {
	config := NewDefaultIntegrationConfig()
	engine := NewIntegratedEngine(config)

	if engine == nil {
		t.Error("Expected engine to be created, got nil")
	}
	if engine.engine == nil {
		t.Error("Expected engine.engine to be initialized")
	}
	if engine.rules == nil {
		t.Error("Expected engine.rules to be initialized")
	}
	if engine.mainValidator == nil {
		t.Error("Expected engine.mainValidator to be initialized")
	}
	if engine.deprecatedDetector == nil {
		t.Error("Expected engine.deprecatedDetector to be initialized")
	}
	if engine.similarSuggester == nil {
		t.Error("Expected engine.similarSuggester to be initialized")
	}
	if engine.errorFormatter == nil {
		t.Error("Expected engine.errorFormatter to be initialized")
	}
	if engine.parser == nil {
		t.Error("Expected engine.parser to be initialized")
	}
	if engine.config != config {
		t.Error("Expected engine.config to be set to the provided config")
	}
	if engine.stats == nil {
		t.Error("Expected engine.stats to be initialized")
	}
	if engine.cache == nil {
		t.Error("Expected engine.cache to be initialized")
	}
}

func TestNewIntegratedEngine_NilConfig(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	if engine == nil {
		t.Error("Expected engine to be created even with nil config")
	}
	if engine.config == nil {
		t.Error("Expected engine.config to be created with default config when nil provided")
	}

	// Should use default config values
	if !engine.config.EnablePreValidation {
		t.Error("Expected default EnablePreValidation to be true")
	}
}

func TestNewIntegratedStats(t *testing.T) {
	stats := NewIntegratedStats()

	if stats == nil {
		t.Error("Expected stats to be created, got nil")
	}
	if stats.TotalLines != 0 {
		t.Errorf("Expected TotalLines to be 0, got %d", stats.TotalLines)
	}
	if stats.ProcessedLines != 0 {
		t.Errorf("Expected ProcessedLines to be 0, got %d", stats.ProcessedLines)
	}
	if stats.TransformedLines != 0 {
		t.Errorf("Expected TransformedLines to be 0, got %d", stats.TransformedLines)
	}
	if stats.PreValidationIssues != 0 {
		t.Errorf("Expected PreValidationIssues to be 0, got %d", stats.PreValidationIssues)
	}
	if stats.AverageConfidence != 0.0 {
		t.Errorf("Expected AverageConfidence to be 0.0, got %f", stats.AverageConfidence)
	}
}

func TestIntegratedEngine_Process_BasicLine(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	testLine := "usacloud server list"
	result := engine.Process(testLine, 1)

	if result == nil {
		t.Error("Expected result to be non-nil")
	}
	if result.LineNumber != 1 {
		t.Errorf("Expected LineNumber to be 1, got %d", result.LineNumber)
	}
	if result.OriginalLine != testLine {
		t.Errorf("Expected OriginalLine to be '%s', got '%s'", testLine, result.OriginalLine)
	}
}

func TestIntegratedEngine_Process_WithTransformation(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	testLine := "usacloud server list --output-type=csv"
	result := engine.Process(testLine, 1)

	if result == nil {
		t.Error("Expected result to be non-nil")
	}
	if result.TransformedLine == "" {
		t.Error("Expected TransformedLine to be non-empty")
	}

	// Should transform csv to json and add comment
	if !strings.Contains(result.TransformedLine, "--output-type=json") {
		t.Errorf("Expected transformed line to contain '--output-type=json', got '%s'", result.TransformedLine)
	}
	if !strings.Contains(result.TransformedLine, "usacloud-update:") {
		t.Errorf("Expected transformed line to contain 'usacloud-update:' comment, got '%s'", result.TransformedLine)
	}
}

func TestIntegratedEngine_Process_EmptyLine(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	result := engine.Process("", 1)

	if result == nil {
		t.Error("Expected result to be non-nil for empty line")
	}
	if result.OriginalLine != "" {
		t.Errorf("Expected OriginalLine to be empty, got '%s'", result.OriginalLine)
	}
}

func TestIntegratedEngine_GetStats(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	// Process some lines
	engine.Process("usacloud server list --output-type=csv", 1)
	engine.Process("usacloud disk list", 2)

	stats := engine.GetStats()
	if stats == nil {
		t.Error("Expected stats to be non-nil")
	}
	if stats.TotalLines == 0 {
		t.Error("Expected stats to show processed lines")
	}
}

func TestIntegratedEngine_ResetStats(t *testing.T) {
	engine := NewIntegratedEngine(nil)

	// Process some lines
	engine.Process("usacloud server list", 1)

	// Verify stats have data
	if engine.stats.TotalLines == 0 {
		t.Error("Expected stats to have data before reset")
	}

	// Reset stats
	engine.ResetStats()

	if engine.stats.TotalLines != 0 {
		t.Errorf("Expected TotalLines to be 0 after ResetStats, got %d", engine.stats.TotalLines)
	}
}

func TestIntegratedEngine_ClearCache(t *testing.T) {
	config := NewDefaultIntegrationConfig()
	config.CacheEnabled = true
	engine := NewIntegratedEngine(config)

	// Process some lines to populate cache
	engine.Process("usacloud server list", 1)
	engine.Process("usacloud disk list", 2)

	// Clear cache
	engine.ClearCache()

	if len(engine.cache) != 0 {
		t.Errorf("Expected cache to be empty after ClearCache, got %d items", len(engine.cache))
	}
}
