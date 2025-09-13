package filter

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

func TestFilterSystem_NewFilterSystem(t *testing.T) {
	system := NewFilterSystem()

	if system == nil {
		t.Fatal("NewFilterSystem should not return nil")
	}

	filters := system.GetFilters()
	if len(filters) == 0 {
		t.Error("FilterSystem should have default filters")
	}

	expectedFilters := []string{"テキスト検索", "カテゴリ", "ステータス"}
	foundFilters := make(map[string]bool)

	for _, filter := range filters {
		foundFilters[filter.Name()] = true
	}

	for _, expected := range expectedFilters {
		if !foundFilters[expected] {
			t.Errorf("Expected filter '%s' not found", expected)
		}
	}
}

func TestFilterSystem_AddRemoveFilter(t *testing.T) {
	system := NewFilterSystem()
	initialCount := len(system.GetFilters())

	customFilter := NewTextFilter()
	system.AddFilter(customFilter)

	filters := system.GetFilters()
	if len(filters) != initialCount+1 {
		t.Errorf("Expected %d filters after adding one, got %d", initialCount+1, len(filters))
	}

	system.RemoveFilter("テキスト検索")
	filters = system.GetFilters()

	if len(filters) != initialCount {
		t.Errorf("Expected %d filters after removing one, got %d", initialCount, len(filters))
	}
}

func TestFilterSystem_ApplyFilters(t *testing.T) {
	system := NewFilterSystem()

	items := []interface{}{
		&preview.CommandPreview{
			Original:    "usacloud server list",
			Transformed: "usacloud server list --output-type=json",
			Category:    "infrastructure",
		},
		&preview.CommandPreview{
			Original:    "usacloud disk list",
			Transformed: "usacloud disk list --output-type=json",
			Category:    "storage",
		},
		&sandbox.ExecutionResult{
			Command: "usacloud switch list",
			Success: true,
		},
	}

	result := system.ApplyFilters(items)

	if len(result) != len(items) {
		t.Errorf("Expected %d items when no filters are active, got %d", len(items), len(result))
	}

	textFilter := system.GetFilter("テキスト検索").(*TextFilter)
	textFilter.SetActive(true)
	textFilter.SetConfig(FilterConfig{
		"query": "server",
	})

	result = system.ApplyFilters(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 item after text filtering, got %d", len(result))
	}

	preview, ok := result[0].(*preview.CommandPreview)
	if !ok {
		t.Fatal("Result item is not a CommandPreview")
	}

	if !strings.Contains(preview.Original, "server") {
		t.Errorf("Result should contain 'server', got '%s'", preview.Original)
	}
}

func TestFilterSystem_MultipleActiveFilters(t *testing.T) {
	system := NewFilterSystem()

	items := []interface{}{
		&preview.CommandPreview{
			Original:    "usacloud server list",
			Transformed: "usacloud server list --output-type=json",
			Category:    "infrastructure",
		},
		&preview.CommandPreview{
			Original:    "usacloud disk list",
			Transformed: "usacloud disk list --output-type=json",
			Category:    "storage",
		},
		&preview.CommandPreview{
			Original:    "usacloud server create",
			Transformed: "usacloud server create --output-type=json",
			Category:    "infrastructure",
		},
	}

	textFilter := system.GetFilter("テキスト検索").(*TextFilter)
	textFilter.SetActive(true)
	textFilter.SetConfig(FilterConfig{
		"query": "server",
	})

	categoryFilter := system.GetFilter("カテゴリ").(*CategoryFilter)
	categoryFilter.SetActive(true)
	categoryFilter.SetConfig(FilterConfig{
		"categories": []string{"infrastructure"},
	})

	result := system.ApplyFilters(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 items after multiple filtering, got %d", len(result))
	}

	for _, item := range result {
		preview := item.(*preview.CommandPreview)
		if !strings.Contains(preview.Original, "server") {
			t.Errorf("Result should contain 'server', got '%s'", preview.Original)
		}
		if preview.Category != "infrastructure" {
			t.Errorf("Result should have infrastructure category, got '%s'", preview.Category)
		}
	}
}

func TestFilterSystem_GetActiveFilters(t *testing.T) {
	system := NewFilterSystem()

	activeFilters := system.GetActiveFilters()
	if len(activeFilters) != 0 {
		t.Errorf("Expected 0 active filters initially, got %d", len(activeFilters))
	}

	textFilter := system.GetFilter("テキスト検索")
	textFilter.SetActive(true)

	activeFilters = system.GetActiveFilters()
	if len(activeFilters) != 1 {
		t.Errorf("Expected 1 active filter, got %d", len(activeFilters))
	}

	if activeFilters[0].Name() != "テキスト検索" {
		t.Errorf("Expected active filter 'テキスト検索', got '%s'", activeFilters[0].Name())
	}
}

func TestFilterSystem_ClearAllFilters(t *testing.T) {
	system := NewFilterSystem()

	textFilter := system.GetFilter("テキスト検索")
	textFilter.SetActive(true)

	categoryFilter := system.GetFilter("カテゴリ")
	categoryFilter.SetActive(true)

	activeFilters := system.GetActiveFilters()
	if len(activeFilters) != 2 {
		t.Errorf("Expected 2 active filters before clearing, got %d", len(activeFilters))
	}

	system.ClearAllFilters()

	activeFilters = system.GetActiveFilters()
	if len(activeFilters) != 0 {
		t.Errorf("Expected 0 active filters after clearing, got %d", len(activeFilters))
	}
}

func TestFilterSystem_GetStatistics(t *testing.T) {
	system := NewFilterSystem()

	items := []interface{}{
		&preview.CommandPreview{Category: "infrastructure"},
		&preview.CommandPreview{Category: "storage"},
		&sandbox.ExecutionResult{Command: "usacloud server list", Success: true},
		&sandbox.ExecutionResult{Command: "usacloud disk list", Success: false},
	}

	stats := system.GetStatistics(items)

	if stats.TotalItems != 4 {
		t.Errorf("Expected 4 total items, got %d", stats.TotalItems)
	}

	if stats.FilteredItems != 4 {
		t.Errorf("Expected 4 filtered items when no filters active, got %d", stats.FilteredItems)
	}

	if len(stats.CategoryCounts) == 0 {
		t.Error("Expected category counts to be populated")
	}

	if len(stats.StatusCounts) == 0 {
		t.Error("Expected status counts to be populated")
	}

	textFilter := system.GetFilter("テキスト検索").(*TextFilter)
	textFilter.SetActive(true)
	textFilter.SetConfig(FilterConfig{
		"query": "nonexistent",
	})

	filteredItems := system.ApplyFilters(items)
	stats = system.GetStatistics(filteredItems)

	if stats.FilteredItems != 0 {
		t.Errorf("Expected 0 filtered items after filtering, got %d", stats.FilteredItems)
	}
}

func TestFilterSystem_ExportImportConfig(t *testing.T) {
	system := NewFilterSystem()

	textFilter := system.GetFilter("テキスト検索").(*TextFilter)
	textFilter.SetActive(true)
	textFilter.SetConfig(FilterConfig{
		"query":          "server",
		"case_sensitive": true,
		"use_regex":      false,
	})

	categoryFilter := system.GetFilter("カテゴリ").(*CategoryFilter)
	categoryFilter.SetActive(true)
	categoryFilter.SetConfig(FilterConfig{
		"categories": []string{"infrastructure", "storage"},
	})

	exported := system.ExportConfig()

	if len(exported) != 2 {
		t.Errorf("Expected 2 exported configs, got %d", len(exported))
	}

	system2 := NewFilterSystem()
	err := system2.ImportConfig(exported)
	if err != nil {
		t.Fatalf("Failed to import config: %v", err)
	}

	activeFilters := system2.GetActiveFilters()
	if len(activeFilters) != 2 {
		t.Errorf("Expected 2 active filters after import, got %d", len(activeFilters))
	}

	importedTextFilter := system2.GetFilter("テキスト検索").(*TextFilter)
	config := importedTextFilter.GetConfig()
	if config["query"] != "server" {
		t.Errorf("Expected query 'server', got '%v'", config["query"])
	}
}

func TestFilterSystem_ConcurrentAccess(t *testing.T) {
	system := NewFilterSystem()

	items := []interface{}{
		&preview.CommandPreview{Original: "usacloud server list"},
		&preview.CommandPreview{Original: "usacloud disk list"},
		&preview.CommandPreview{Original: "usacloud switch list"},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if id%2 == 0 {
				result := system.ApplyFilters(items)
				if len(result) != len(items) {
					t.Errorf("Goroutine %d: unexpected result length", id)
				}
			} else {
				stats := system.GetStatistics(items)
				if stats.TotalItems != len(items) {
					t.Errorf("Goroutine %d: unexpected total items", id)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestFilterSystem_Performance(t *testing.T) {
	system := NewFilterSystem()

	var items []interface{}
	for i := 0; i < 10000; i++ {
		category := "infrastructure"
		if i%3 == 0 {
			category = "storage"
		} else if i%3 == 1 {
			category = "network"
		}

		items = append(items, &preview.CommandPreview{
			Original:    "usacloud server list",
			Transformed: "usacloud server list --output-type=json",
			Category:    category,
		})
	}

	start := time.Now()
	result := system.ApplyFilters(items)
	duration := time.Since(start)

	if len(result) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(result))
	}

	if duration > time.Second {
		t.Errorf("Filter application took too long: %v", duration)
	}

	categoryFilter := system.GetFilter("カテゴリ").(*CategoryFilter)
	categoryFilter.SetActive(true)
	categoryFilter.SetConfig(FilterConfig{
		"categories": []string{"infrastructure"},
	})

	start = time.Now()
	result = system.ApplyFilters(items)
	duration = time.Since(start)

	expectedCount := len(items) / 3
	if abs(len(result)-expectedCount) > 100 {
		t.Errorf("Expected approximately %d items after filtering, got %d", expectedCount, len(result))
	}

	if duration > time.Second {
		t.Errorf("Filtered application took too long: %v", duration)
	}
}

func TestFilterSystem_SaveLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "filter-config.json")

	system := NewFilterSystem()

	textFilter := system.GetFilter("テキスト検索").(*TextFilter)
	textFilter.SetActive(true)
	textFilter.SetConfig(FilterConfig{
		"query": "server",
	})

	err := system.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should exist after saving")
	}

	system2 := NewFilterSystem()
	err = system2.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	activeFilters := system2.GetActiveFilters()
	if len(activeFilters) != 1 {
		t.Errorf("Expected 1 active filter after loading, got %d", len(activeFilters))
	}

	loadedTextFilter := system2.GetFilter("テキスト検索").(*TextFilter)
	config := loadedTextFilter.GetConfig()
	if config["query"] != "server" {
		t.Errorf("Expected query 'server', got '%v'", config["query"])
	}
}

func TestFilterSystem_ErrorHandling(t *testing.T) {
	system := NewFilterSystem()

	invalidConfig := []FilterExport{
		{
			Name:   "nonexistent",
			Active: true,
			Config: FilterConfig{},
		},
	}

	err := system.ImportConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error when importing config for nonexistent filter")
	}

	err = system.LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error when loading from nonexistent path")
	}

	err = system.SaveConfig("/invalid/path/config.json")
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestFilterSystem_GetFilter_NotFound(t *testing.T) {
	system := NewFilterSystem()

	filter := system.GetFilter("nonexistent")
	if filter != nil {
		t.Error("Expected nil for nonexistent filter")
	}
}

func TestFilterSystem_EmptyItems(t *testing.T) {
	system := NewFilterSystem()

	result := system.ApplyFilters([]interface{}{})
	if len(result) != 0 {
		t.Errorf("Expected 0 items for empty input, got %d", len(result))
	}

	stats := system.GetStatistics([]interface{}{})
	if stats.TotalItems != 0 {
		t.Errorf("Expected 0 total items for empty input, got %d", stats.TotalItems)
	}
}

func TestFilterSystem_NilItems(t *testing.T) {
	system := NewFilterSystem()

	items := []interface{}{nil, nil, nil}
	result := system.ApplyFilters(items)

	if len(result) != 0 {
		t.Errorf("Expected 0 items after filtering nil items, got %d", len(result))
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
