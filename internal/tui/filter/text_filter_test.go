package filter

import (
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

func TestNewTextSearchFilter(t *testing.T) {
	filter := NewTextSearchFilter()

	if filter == nil {
		t.Error("Filter should not be nil")
	}

	if filter.Name() != "text-search" {
		t.Errorf("Expected name 'text-search', got '%s'", filter.Name())
	}

	if filter.IsActive() {
		t.Error("Filter should be inactive by default")
	}

	if filter.GetSearchTerm() != "" {
		t.Errorf("Expected empty search term, got '%s'", filter.GetSearchTerm())
	}

	fields := filter.GetFields()
	expectedFields := []string{"command", "description", "output"}
	if len(fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestTextSearchFilter_SetSearchTerm(t *testing.T) {
	filter := NewTextSearchFilter()

	// Test setting search term
	filter.SetSearchTerm("test")

	if filter.GetSearchTerm() != "test" {
		t.Errorf("Expected search term 'test', got '%s'", filter.GetSearchTerm())
	}

	if !filter.IsActive() {
		t.Error("Filter should be active when search term is set")
	}

	// Test clearing search term
	filter.SetSearchTerm("")

	if filter.IsActive() {
		t.Error("Filter should be inactive when search term is empty")
	}
}

func TestTextSearchFilter_CaseSensitivity(t *testing.T) {
	filter := NewTextSearchFilter()

	// Default should be case insensitive
	if filter.IsCaseSensitive() {
		t.Error("Filter should be case insensitive by default")
	}

	filter.SetCaseSensitive(true)
	if !filter.IsCaseSensitive() {
		t.Error("Filter should be case sensitive after setting")
	}

	filter.SetCaseSensitive(false)
	if filter.IsCaseSensitive() {
		t.Error("Filter should be case insensitive after setting")
	}
}

func TestTextSearchFilter_Regex(t *testing.T) {
	filter := NewTextSearchFilter()

	// Default should be no regex
	if filter.IsUsingRegex() {
		t.Error("Filter should not use regex by default")
	}

	filter.SetUseRegex(true)
	if !filter.IsUsingRegex() {
		t.Error("Filter should use regex after setting")
	}

	filter.SetUseRegex(false)
	if filter.IsUsingRegex() {
		t.Error("Filter should not use regex after setting")
	}
}

func TestTextSearchFilter_Apply_StringItems(t *testing.T) {
	filter := NewTextSearchFilter()

	items := []interface{}{
		"usacloud server list",
		"usacloud database create",
		"usacloud disk attach",
		"echo hello world",
	}

	// Test with inactive filter
	result := filter.Apply(items)
	if len(result) != len(items) {
		t.Errorf("Expected %d items with inactive filter, got %d", len(items), len(result))
	}

	// Test with active filter
	filter.SetSearchTerm("usacloud")
	result = filter.Apply(items)
	if len(result) != 3 {
		t.Errorf("Expected 3 items matching 'usacloud', got %d", len(result))
	}

	// Test case sensitivity
	filter.SetSearchTerm("USACLOUD")
	filter.SetCaseSensitive(false)
	result = filter.Apply(items)
	if len(result) != 3 {
		t.Errorf("Expected 3 items with case insensitive search, got %d", len(result))
	}

	filter.SetCaseSensitive(true)
	result = filter.Apply(items)
	if len(result) != 0 {
		t.Errorf("Expected 0 items with case sensitive search for 'USACLOUD', got %d", len(result))
	}
}

func TestTextSearchFilter_Apply_PreviewItems(t *testing.T) {
	filter := NewTextSearchFilter()

	items := []interface{}{
		&preview.CommandPreview{
			Original:    "usacloud server list --output-type=csv",
			Transformed: "usacloud server list --output-type=json",
			Description: "List all servers",
			Warnings:    []string{"Output format changed"},
		},
		&preview.CommandPreview{
			Original:    "usacloud database create",
			Transformed: "usacloud database create",
			Description: "Create a new database",
		},
	}

	// Test searching in original command
	filter.SetSearchTerm("server")
	result := filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'server', got %d", len(result))
	}

	// Test searching in description
	filter.SetSearchTerm("database")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'database', got %d", len(result))
	}

	// Test searching in warnings
	filter.SetSearchTerm("format")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'format' in warnings, got %d", len(result))
	}
}

func TestTextSearchFilter_Apply_ExecutionResultItems(t *testing.T) {
	filter := NewTextSearchFilter()

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command:    "usacloud server list",
			Output:     "Server1: running\nServer2: stopped",
			Error:      "",
			Success:    true,
			SkipReason: "",
		},
		&sandbox.ExecutionResult{
			Command:    "usacloud database create",
			Output:     "",
			Error:      "Connection failed",
			Success:    false,
			SkipReason: "",
		},
	}

	// Test searching in command
	filter.SetSearchTerm("server")
	result := filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'server', got %d", len(result))
	}

	// Test searching in output
	filter.SetSearchTerm("running")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'running', got %d", len(result))
	}

	// Test searching in error
	filter.SetSearchTerm("Connection")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'Connection', got %d", len(result))
	}
}

func TestTextSearchFilter_RegexAdvanced(t *testing.T) {
	filter := NewTextSearchFilter()
	filter.SetUseRegex(true)

	items := []interface{}{
		"usacloud server list",
		"usacloud server create",
		"usacloud database list",
		"echo test123",
	}

	// Test simple regex
	filter.SetSearchTerm("server.*list")
	result := filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching regex 'server.*list', got %d", len(result))
	}

	// Test character class regex
	filter.SetSearchTerm("test[0-9]+")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching regex 'test[0-9]+', got %d", len(result))
	}

	// Test invalid regex (should fallback to string search)
	filter.SetSearchTerm("[invalid")
	result = filter.Apply(items)
	if len(result) != 0 {
		t.Errorf("Expected 0 items with invalid regex, got %d", len(result))
	}
}

func TestTextSearchFilter_GetConfig(t *testing.T) {
	filter := NewTextSearchFilter()
	filter.SetSearchTerm("test")
	filter.SetCaseSensitive(true)
	filter.SetUseRegex(true)

	config := filter.GetConfig()

	// FilterConfig no longer has FilterID field - check by other means
	if filter.Name() != "text-search" {
		t.Errorf("Expected filter name 'text-search', got '%s'", filter.Name())
	}

	if !filter.IsActive() {
		t.Error("Expected filter to be active")
	}

	// Config is now map[string]interface{} - check individual fields
	if query, ok := config["query"].(string); !ok || query != "test" {
		t.Errorf("Expected query 'test', got %v", config["query"])
		return
	}

	if caseSensitive, ok := config["case_sensitive"].(bool); !ok || !caseSensitive {
		t.Errorf("Expected case_sensitive true, got %v", config["case_sensitive"])
	}

	if useRegex, ok := config["regex_mode"].(bool); !ok || !useRegex {
		t.Errorf("Expected regex_mode true, got %v", config["regex_mode"])
	}
}

func TestTextSearchFilter_SetConfig(t *testing.T) {
	filter := NewTextSearchFilter()

	config := FilterConfig{
		"query":          "test config",
		"case_sensitive": true,
		"regex_mode":     true,
		"fields":         []string{"command", "output"},
	}

	err := filter.SetConfig(config)
	if err != nil {
		t.Errorf("SetConfig failed: %v", err)
	}

	if filter.GetSearchTerm() != "test config" {
		t.Errorf("Expected search term 'test config', got '%s'", filter.GetSearchTerm())
	}

	if !filter.IsCaseSensitive() {
		t.Error("Expected case sensitive to be true")
	}

	if !filter.IsUsingRegex() {
		t.Error("Expected use regex to be true")
	}

	fields := filter.GetFields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}
}

func TestTextSearchFilter_SetConfig_MapInterface(t *testing.T) {
	filter := NewTextSearchFilter()

	// New FilterConfig format
	config := FilterConfig{
		"query":          "map config",
		"case_sensitive": true,
		"regex_mode":     false,
		"fields":         []interface{}{"command", "description"},
	}

	err := filter.SetConfig(config)
	if err != nil {
		t.Errorf("SetConfig with map failed: %v", err)
	}

	if filter.GetSearchTerm() != "map config" {
		t.Errorf("Expected search term 'map config', got '%s'", filter.GetSearchTerm())
	}

	if !filter.IsCaseSensitive() {
		t.Error("Expected case sensitive to be true")
	}

	if filter.IsUsingRegex() {
		t.Error("Expected use regex to be false")
	}
}

func TestTextSearchFilter_ClearSearch(t *testing.T) {
	filter := NewTextSearchFilter()
	filter.SetSearchTerm("test")
	filter.SetCaseSensitive(true)

	if !filter.IsActive() {
		t.Error("Filter should be active before clear")
	}

	filter.ClearSearch()

	if filter.IsActive() {
		t.Error("Filter should be inactive after clear")
	}

	if filter.GetSearchTerm() != "" {
		t.Errorf("Expected empty search term after clear, got '%s'", filter.GetSearchTerm())
	}

	// Other settings should remain
	if !filter.IsCaseSensitive() {
		t.Error("Case sensitivity should remain after clear")
	}
}

func TestTextSearchFilter_SetFields(t *testing.T) {
	filter := NewTextSearchFilter()

	newFields := []string{"command", "error"}
	filter.SetFields(newFields)

	fields := filter.GetFields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	if fields[0] != "command" {
		t.Errorf("Expected first field 'command', got '%s'", fields[0])
	}

	if fields[1] != "error" {
		t.Errorf("Expected second field 'error', got '%s'", fields[1])
	}

	// Test that original slice is not modified
	newFields[0] = "modified"
	fields = filter.GetFields()
	if fields[0] == "modified" {
		t.Error("Fields should be independent of original slice")
	}
}

func TestTextSearchFilter_EdgeCases(t *testing.T) {
	filter := NewTextSearchFilter()

	// Test with empty items
	result := filter.Apply([]interface{}{})
	if len(result) != 0 {
		t.Errorf("Expected 0 items with empty input, got %d", len(result))
	}

	// Test with nil items
	result = filter.Apply(nil)
	if len(result) != 0 {
		t.Errorf("Expected 0 items with nil input, got %d", len(result))
	}

	// Test with mixed types
	items := []interface{}{
		"string item",
		123,
		nil,
		&preview.CommandPreview{Original: "test command"},
	}

	filter.SetSearchTerm("test")
	result = filter.Apply(items)
	if len(result) != 1 {
		t.Errorf("Expected 1 item matching 'test', got %d", len(result))
	}

	// Test with very long search term
	longTerm := strings.Repeat("a", 1000)
	filter.SetSearchTerm(longTerm)
	result = filter.Apply(items)
	if len(result) != 0 {
		t.Errorf("Expected 0 items with long search term, got %d", len(result))
	}
}

func TestTextSearchFilter_Performance(t *testing.T) {
	filter := NewTextSearchFilter()
	filter.SetSearchTerm("test")

	// Create a large number of items
	items := make([]interface{}, 10000)
	for i := 0; i < 10000; i++ {
		if i%2 == 0 {
			items[i] = "test item " + string(rune(i))
		} else {
			items[i] = "other item " + string(rune(i))
		}
	}

	// Apply filter and measure (basic performance test)
	result := filter.Apply(items)

	// Should filter out roughly half the items
	if len(result) < 4000 || len(result) > 6000 {
		t.Errorf("Expected around 5000 items, got %d", len(result))
	}
}
