package filter

import (
	"testing"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

func TestNewStatusFilter(t *testing.T) {
	filter := NewStatusFilter()

	if filter == nil {
		t.Fatal("NewStatusFilter should not return nil")
	}

	if filter.Name() != "ステータス" {
		t.Errorf("Expected name 'ステータス', got %s", filter.Name())
	}

	if filter.Description() != "実行ステータスでフィルタリング" {
		t.Errorf("Expected description '実行ステータスでフィルタリング', got %s", filter.Description())
	}

	if filter.IsActive() {
		t.Error("New status filter should not be active by default")
	}
}

func TestStatusFilter_GetSetConfig(t *testing.T) {
	filter := NewStatusFilter()

	config := FilterConfig{
		"statuses": []string{"success", "failed"},
	}

	err := filter.SetConfig(config)
	if err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	retrievedConfig := filter.GetConfig()
	statuses, ok := retrievedConfig["statuses"]
	if !ok {
		t.Error("Config should contain 'statuses' key")
	}

	statusList, ok := statuses.([]string)
	if !ok {
		t.Error("Statuses should be a []string")
	}

	if len(statusList) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statusList))
	}
}

func TestStatusFilter_Apply_ExecutionResults(t *testing.T) {
	filter := NewStatusFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"statuses": []string{"success", "failed"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: false,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud switch list",
			Success: false,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud router list",
			Success: true,
		},
	}

	result := filter.Apply(items)

	if len(result) != 4 {
		t.Errorf("Expected 4 items after filtering (all items), got %d", len(result))
	}

	for _, item := range result {
		execResult, ok := item.(*sandbox.ExecutionResult)
		if !ok {
			t.Fatal("Result item is not an ExecutionResult")
		}

		// Check that the results match our filter criteria (success or failed)
		// Since ExecutionResult only has Success bool, we check both true and false are included
		if execResult.Command == "" {
			t.Errorf("Empty command in result")
		}
	}
}

func TestStatusFilter_Apply_CommandPreview(t *testing.T) {
	filter := NewStatusFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"statuses": []string{"pending"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&preview.CommandPreview{
			Original:    "usacloud server list",
			Transformed: "usacloud server list --output-type=json",
			Description: "Convert to JSON output",
			Category:    "output-format",
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: true,
		},
	}

	result := filter.Apply(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 item after filtering, got %d", len(result))
	}

	preview, ok := result[0].(*preview.CommandPreview)
	if !ok {
		t.Fatal("Result item should be a CommandPreview")
	}

	if preview.Original != "usacloud server list" {
		t.Errorf("Expected command 'usacloud server list', got %s", preview.Original)
	}
}

func TestStatusFilter_Apply_Mixed(t *testing.T) {
	filter := NewStatusFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"statuses": []string{"success", "pending"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: false,
		},
		&preview.CommandPreview{
			Original:    "usacloud switch list",
			Transformed: "usacloud switch list --output-type=json",
			Description: "Convert to JSON output",
			Category:    "output-format",
		},
	}

	result := filter.Apply(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 items after filtering, got %d", len(result))
	}

	// First item should be success ExecutionResult
	execResult, ok := result[0].(*sandbox.ExecutionResult)
	if !ok {
		t.Fatal("First result item should be ExecutionResult")
	}
	if !execResult.Success {
		t.Error("First result should be successful")
	}

	// Second item should be pending CommandPreview
	preview, ok := result[1].(*preview.CommandPreview)
	if !ok {
		t.Fatal("Second result item should be CommandPreview")
	}
	if preview.Original != "usacloud switch list" {
		t.Errorf("Expected command 'usacloud switch list', got %s", preview.Original)
	}
}

func TestStatusFilter_Apply_NoFilter(t *testing.T) {
	filter := NewStatusFilter()
	// Filter is not active

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: false,
		},
	}

	result := filter.Apply(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 items when filter is inactive, got %d", len(result))
	}
}

func TestStatusFilter_Apply_EmptyStatuses(t *testing.T) {
	filter := NewStatusFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"statuses": []string{},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: false,
		},
	}

	result := filter.Apply(items)

	if len(result) != 0 {
		t.Errorf("Expected 0 items when no statuses are allowed, got %d", len(result))
	}
}

func TestStatusFilter_Apply_SkippedResults(t *testing.T) {
	filter := NewStatusFilter()
	filter.SetActive(true)

	config := FilterConfig{
		"statuses": []string{"skipped"},
	}
	filter.SetConfig(config)

	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "usacloud server list",
			Success: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud disk list",
			Success: false,
			Skipped: true,
		},
		&sandbox.ExecutionResult{
			Command: "usacloud switch list",
			Success: false,
		},
	}

	result := filter.Apply(items)

	if len(result) != 1 {
		t.Errorf("Expected 1 skipped item, got %d", len(result))
	}

	execResult, ok := result[0].(*sandbox.ExecutionResult)
	if !ok {
		t.Fatal("Result item should be ExecutionResult")
	}

	if !execResult.Skipped {
		t.Error("Result should be skipped")
	}
}

func TestStatusFilter_ToggleStatus(t *testing.T) {
	filter := NewStatusFilter()

	// Initially all statuses should be allowed
	if !filter.config.AllowedStatuses[StatusSuccess] {
		t.Error("StatusSuccess should be allowed initially")
	}

	// Toggle success status
	filter.ToggleStatus(StatusSuccess)

	if filter.config.AllowedStatuses[StatusSuccess] {
		t.Error("StatusSuccess should be disabled after toggle")
	}

	// Toggle again
	filter.ToggleStatus(StatusSuccess)

	if !filter.config.AllowedStatuses[StatusSuccess] {
		t.Error("StatusSuccess should be enabled after second toggle")
	}
}

func TestStatusFilter_GetStatusFromItem(t *testing.T) {
	filter := NewStatusFilter()

	tests := []struct {
		name     string
		item     interface{}
		expected string
	}{
		{
			name: "Success ExecutionResult",
			item: &sandbox.ExecutionResult{
				Command: "test",
				Success: true,
			},
			expected: "success",
		},
		{
			name: "Failed ExecutionResult",
			item: &sandbox.ExecutionResult{
				Command: "test",
				Success: false,
			},
			expected: "failed",
		},
		{
			name: "Skipped ExecutionResult",
			item: &sandbox.ExecutionResult{
				Command: "test",
				Success: false,
				Skipped: true,
			},
			expected: "skipped",
		},
		{
			name: "CommandPreview",
			item: &preview.CommandPreview{
				Original: "test command",
			},
			expected: "pending",
		},
		{
			name:     "String",
			item:     "test string",
			expected: "pending",
		},
		{
			name:     "Unknown type",
			item:     123,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.getStatusFromItem(tt.item)
			if result != tt.expected {
				t.Errorf("Expected status %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStatusFilter_SetConfig_InvalidStatus(t *testing.T) {
	filter := NewStatusFilter()

	config := FilterConfig{
		"statuses": []string{"invalid_status"},
	}

	err := filter.SetConfig(config)
	if err != nil {
		t.Fatalf("SetConfig should not fail with invalid status: %v", err)
	}

	// Invalid statuses should be ignored
	filter.SetActive(true)
	items := []interface{}{
		&sandbox.ExecutionResult{
			Command: "test",
			Success: true,
		},
	}

	result := filter.Apply(items)
	if len(result) != 0 {
		t.Errorf("Expected 0 items with invalid status filter, got %d", len(result))
	}
}
