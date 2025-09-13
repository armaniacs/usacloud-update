package filter

import (
	"strings"
	"testing"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

func TestCategoryFilter_Name(t *testing.T) {
	filter := NewCategoryFilter()
	if filter.Name() != "カテゴリ" {
		t.Errorf("Expected name 'カテゴリ', got '%s'", filter.Name())
	}
}

func TestCategoryFilter_Description(t *testing.T) {
	filter := NewCategoryFilter()
	desc := filter.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if desc != "コマンドのカテゴリでフィルタリング" {
		t.Errorf("Unexpected description: %s", desc)
	}
}

func TestCategoryFilter_SetGetConfig(t *testing.T) {
	filter := NewCategoryFilter()

	config := FilterConfig{
		"categories": []string{"infrastructure", "storage"},
	}

	err := filter.SetConfig(config)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	retrievedConfig := filter.GetConfig()
	categories, ok := retrievedConfig["categories"].([]string)
	if !ok {
		t.Fatal("Categories not found in config or wrong type")
	}

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	expectedCategories := map[string]bool{"infrastructure": true, "storage": true}
	for _, cat := range categories {
		if !expectedCategories[cat] {
			t.Errorf("Unexpected category: %s", cat)
		}
	}
}

func TestCategoryFilter_Apply_CommandPreview(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"infrastructure"},
	}
	filter.SetConfig(config)

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
			Original:    "usacloud switch list",
			Transformed: "usacloud switch list --output-type=json",
			Category:    "network",
		},
	}

	result := filter.Apply(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 item after filtering, got %d", len(result))
	}

	preview, ok := result[0].(*preview.CommandPreview)
	if !ok {
		t.Fatal("Result item is not a CommandPreview")
	}

	if preview.Category != "infrastructure" {
		t.Errorf("Expected infrastructure category, got %s", preview.Category)
	}
}

func TestCategoryFilter_Apply_ExecutionResult(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"storage", "network"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud switch list",
			Success: false,
		},
	}

	result := filter.Apply(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 items after filtering, got %d", len(result))
	}

	for _, item := range result {
		execResult, ok := item.(*sandbox.ExecutionResult)
		if !ok {
			t.Fatal("Result item is not an ExecutionResult")
		}

		// Category is now determined from command, so we just verify the command type
		if !strings.Contains(execResult.Command, "disk") && !strings.Contains(execResult.Command, "switch") {
			t.Errorf("Unexpected command in result: %s", execResult.Command)
		}
	}
}

func TestCategoryFilter_Apply_StringItems(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"infrastructure"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		"usacloud server list",
		"usacloud disk create",
		"usacloud switch list",
	}

	result := filter.Apply(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 item after filtering, got %d", len(result))
	}

	cmd, ok := result[0].(string)
	if !ok {
		t.Fatal("Result item is not a string")
	}

	if cmd != "usacloud server list" {
		t.Errorf("Expected 'usacloud server list', got '%s'", cmd)
	}
}

func TestCategoryFilter_Apply_InactiveFilter(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(false)

	config := FilterConfig{
		"categories": []string{"infrastructure"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&preview.CommandPreview{Category: "infrastructure"},
		&preview.CommandPreview{Category: "storage"},
		&preview.CommandPreview{Category: "network"},
	}

	result := filter.Apply(items)

	if len(result) != len(items) {
		t.Errorf("Expected all items when filter is inactive, got %d out of %d", len(result), len(items))
	}
}

func TestCategoryFilter_Apply_EmptyCategories(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&preview.CommandPreview{Category: "infrastructure"},
		&preview.CommandPreview{Category: "storage"},
	}

	result := filter.Apply(items)

	if len(result) != len(items) {
		t.Errorf("Expected all items when no categories selected, got %d out of %d", len(result), len(items))
	}
}

func TestCategoryFilter_Apply_UnsupportedType(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"infrastructure"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		42,
		struct{ Name string }{Name: "test"},
		&preview.CommandPreview{Category: "infrastructure"},
	}

	result := filter.Apply(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 item after filtering, got %d", len(result))
	}

	preview, ok := result[0].(*preview.CommandPreview)
	if !ok {
		t.Fatal("Result item is not a CommandPreview")
	}

	if preview.Category != "infrastructure" {
		t.Errorf("Expected infrastructure category, got %s", preview.Category)
	}
}

func TestCategoryFilter_CategorizeCommand(t *testing.T) {
	testCases := []struct {
		command  string
		expected string
	}{
		{"usacloud server list", "infrastructure"},
		{"usacloud server create", "infrastructure"},
		{"usacloud database list", "managed-service"},
		{"usacloud disk list", "storage"},
		{"usacloud archive list", "storage"},
		{"usacloud switch list", "network"},
		{"usacloud router list", "network"},
		{"usacloud loadbalancer list", "network"},
		{"usacloud certificateauthority list", "security"},
		{"usacloud ssh-key list", "security"},
		{"usacloud activity list", "monitoring"},
		{"usacloud bill list", "monitoring"},
		{"usacloud unknown-command list", "other"},
		{"not-usacloud command", "other"},
	}

	filter := &CategoryFilter{}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result := filter.categorizeCommand(tc.command)
			if result != tc.expected {
				t.Errorf("For command '%s', expected category '%s', got '%s'", tc.command, tc.expected, result)
			}
		})
	}
}

func TestCategoryFilter_IsActive(t *testing.T) {
	filter := NewCategoryFilter()

	if filter.IsActive() {
		t.Error("Filter should be inactive by default")
	}

	filter.SetActive(true)
	if !filter.IsActive() {
		t.Error("Filter should be active after SetActive(true)")
	}

	filter.SetActive(false)
	if filter.IsActive() {
		t.Error("Filter should be inactive after SetActive(false)")
	}
}

func TestCategoryFilter_SetConfig_InvalidConfig(t *testing.T) {
	filter := NewCategoryFilter()

	testCases := []struct {
		name   string
		config FilterConfig
	}{
		{
			name: "wrong type for categories",
			config: FilterConfig{
				"categories": "not an array",
			},
		},
		{
			name: "array of wrong type",
			config: FilterConfig{
				"categories": []int{1, 2, 3},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := filter.SetConfig(tc.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestCategoryFilter_Performance(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"infrastructure"},
	}
	filter.SetConfig(config)

	var items []interface{}
	for i := 0; i < 1000; i++ {
		category := "infrastructure"
		if i%2 == 0 {
			category = "storage"
		}
		items = append(items, &preview.CommandPreview{
			Original:    "usacloud server list",
			Transformed: "usacloud server list --output-type=json",
			Category:    category,
		})
	}

	result := filter.Apply(items)

	expectedCount := 500
	if len(result) != expectedCount {
		t.Errorf("Expected %d items after filtering, got %d", expectedCount, len(result))
	}
}

func TestCategoryFilter_MultipleCategories(t *testing.T) {
	filter := NewCategoryFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"categories": []string{"infrastructure", "storage", "network"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&preview.CommandPreview{Category: "infrastructure"},
		&preview.CommandPreview{Category: "storage"},
		&preview.CommandPreview{Category: "network"},
		&preview.CommandPreview{Category: "managed-service"},
		&preview.CommandPreview{Category: "security"},
	}

	result := filter.Apply(items)

	if len(result) != 3 {
		t.Errorf("Expected 3 items after filtering, got %d", len(result))
	}

	expectedCategories := map[string]bool{
		"infrastructure": true,
		"storage":        true,
		"network":        true,
	}

	for _, item := range result {
		preview := item.(*preview.CommandPreview)
		if !expectedCategories[preview.Category] {
			t.Errorf("Unexpected category in result: %s", preview.Category)
		}
	}
}
