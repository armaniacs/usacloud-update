package filter

import (
	"testing"
	"time"
)

func TestExecutionStatus_String(t *testing.T) {
	tests := []struct {
		status   ExecutionStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusSuccess, "success"},
		{StatusFailed, "failed"},
		{StatusSkipped, "skipped"},
	}

	for _, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			if string(test.status) != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, string(test.status))
			}
		})
	}
}

func TestFilterConfig_Creation(t *testing.T) {
	config := FilterConfig{
		"filter_id": "test-filter",
		"active":    true,
		"value":     "test-value",
	}

	if config["filter_id"] != "test-filter" {
		t.Errorf("Expected filter_id 'test-filter', got '%s'", config["filter_id"])
	}

	if active, ok := config["active"].(bool); !ok || !active {
		t.Errorf("Expected active to be true, got %v", config["active"])
	}

	if config["value"] != "test-value" {
		t.Errorf("Expected value 'test-value', got '%v'", config["value"])
	}
}

func TestFilterSet_Creation(t *testing.T) {
	now := time.Now()
	filters := []FilterExport{
		{Name: "filter1", Active: true, Config: FilterConfig{"filter_id": "filter1", "value": "value1"}},
		{Name: "filter2", Active: false, Config: FilterConfig{"filter_id": "filter2", "value": "value2"}},
	}

	filterSet := FilterSet{
		ID:      "test-set",
		Name:    "Test Set",
		Filters: filters,
		Created: now,
	}

	if filterSet.ID != "test-set" {
		t.Errorf("Expected ID 'test-set', got '%s'", filterSet.ID)
	}

	if filterSet.Name != "Test Set" {
		t.Errorf("Expected Name 'Test Set', got '%s'", filterSet.Name)
	}

	if len(filterSet.Filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filterSet.Filters))
	}

	if !filterSet.Created.Equal(now) {
		t.Errorf("Expected Created time %v, got %v", now, filterSet.Created)
	}
}

func TestTextSearchConfig_Defaults(t *testing.T) {
	config := TextSearchConfig{
		Fields: []string{"command", "description"},
	}

	if config.CaseSensitive {
		t.Error("Expected CaseSensitive to be false by default")
	}

	if config.UseRegex {
		t.Error("Expected UseRegex to be false by default")
	}

	if len(config.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(config.Fields))
	}

	if config.Fields[0] != "command" {
		t.Errorf("Expected first field 'command', got '%s'", config.Fields[0])
	}
}

func TestCategoryFilterConfig_Creation(t *testing.T) {
	selectedCategories := map[string]bool{
		"infrastructure": true,
		"storage":        false,
	}

	availableCategories := []string{
		"infrastructure", "storage", "network",
	}

	config := CategoryFilterConfig{
		SelectedCategories:  selectedCategories,
		AvailableCategories: availableCategories,
	}

	if len(config.SelectedCategories) != 2 {
		t.Errorf("Expected 2 selected categories, got %d", len(config.SelectedCategories))
	}

	if !config.SelectedCategories["infrastructure"] {
		t.Error("Expected infrastructure to be selected")
	}

	if config.SelectedCategories["storage"] {
		t.Error("Expected storage to not be selected")
	}

	if len(config.AvailableCategories) != 3 {
		t.Errorf("Expected 3 available categories, got %d", len(config.AvailableCategories))
	}
}

func TestStatusFilterConfig_Creation(t *testing.T) {
	allowedStatuses := map[ExecutionStatus]bool{
		StatusSuccess: true,
		StatusFailed:  true,
		StatusPending: false,
	}

	config := StatusFilterConfig{
		AllowedStatuses: allowedStatuses,
	}

	if len(config.AllowedStatuses) != 3 {
		t.Errorf("Expected 3 allowed statuses, got %d", len(config.AllowedStatuses))
	}

	if !config.AllowedStatuses[StatusSuccess] {
		t.Error("Expected StatusSuccess to be allowed")
	}

	if !config.AllowedStatuses[StatusFailed] {
		t.Error("Expected StatusFailed to be allowed")
	}

	if config.AllowedStatuses[StatusPending] {
		t.Error("Expected StatusPending to not be allowed")
	}
}

func TestFilterStats_Creation(t *testing.T) {
	now := time.Now()
	duration := 100 * time.Millisecond

	stats := FilterStats{
		TotalItems:     100,
		FilteredItems:  50,
		FilterDuration: duration,
		LastUpdated:    now,
	}

	if stats.TotalItems != 100 {
		t.Errorf("Expected TotalItems 100, got %d", stats.TotalItems)
	}

	if stats.FilteredItems != 50 {
		t.Errorf("Expected FilteredItems 50, got %d", stats.FilteredItems)
	}

	if stats.FilterDuration != duration {
		t.Errorf("Expected FilterDuration %v, got %v", duration, stats.FilterDuration)
	}

	if !stats.LastUpdated.Equal(now) {
		t.Errorf("Expected LastUpdated %v, got %v", now, stats.LastUpdated)
	}
}

func TestFilterEvent_Creation(t *testing.T) {
	now := time.Now()

	event := FilterEvent{
		Type:      EventFilterActivated,
		FilterID:  "test-filter",
		Timestamp: now,
		Data:      "test-data",
	}

	if event.Type != EventFilterActivated {
		t.Errorf("Expected Type %s, got %s", EventFilterActivated, event.Type)
	}

	if event.FilterID != "test-filter" {
		t.Errorf("Expected FilterID 'test-filter', got '%s'", event.FilterID)
	}

	if !event.Timestamp.Equal(now) {
		t.Errorf("Expected Timestamp %v, got %v", now, event.Timestamp)
	}

	if event.Data != "test-data" {
		t.Errorf("Expected Data 'test-data', got '%v'", event.Data)
	}
}

func TestFilterEventType_Values(t *testing.T) {
	events := []FilterEventType{
		EventFilterActivated,
		EventFilterDeactivated,
		EventFilterConfigured,
		EventFilterApplied,
		EventPresetSaved,
		EventPresetLoaded,
	}

	expectedValues := []string{
		"filter_activated",
		"filter_deactivated",
		"filter_configured",
		"filter_applied",
		"preset_saved",
		"preset_loaded",
	}

	if len(events) != len(expectedValues) {
		t.Errorf("Expected %d event types, got %d", len(expectedValues), len(events))
	}

	for i, event := range events {
		if string(event) != expectedValues[i] {
			t.Errorf("Expected event type %s, got %s", expectedValues[i], string(event))
		}
	}
}

// Mock implementations for testing
type MockFilterableItem struct {
	category       string
	status         ExecutionStatus
	searchableText []string
}

func (m *MockFilterableItem) GetCategory() string {
	return m.category
}

func (m *MockFilterableItem) GetStatus() ExecutionStatus {
	return m.status
}

func (m *MockFilterableItem) GetSearchableText() []string {
	return m.searchableText
}

func TestMockFilterableItem(t *testing.T) {
	item := &MockFilterableItem{
		category:       "infrastructure",
		status:         StatusSuccess,
		searchableText: []string{"test command", "test description"},
	}

	if item.GetCategory() != "infrastructure" {
		t.Errorf("Expected category 'infrastructure', got '%s'", item.GetCategory())
	}

	if item.GetStatus() != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, item.GetStatus())
	}

	searchableText := item.GetSearchableText()
	if len(searchableText) != 2 {
		t.Errorf("Expected 2 searchable texts, got %d", len(searchableText))
	}

	if searchableText[0] != "test command" {
		t.Errorf("Expected first text 'test command', got '%s'", searchableText[0])
	}
}

func TestConfigValidation(t *testing.T) {
	// Test valid config
	config := FilterConfig{
		"filter_id": "valid-filter",
		"active":    true,
		"value":     "valid-value",
	}

	if filterID, ok := config["filter_id"].(string); !ok || filterID == "" {
		t.Errorf("filter_id should not be empty, got %v", config["filter_id"])
	}

	// Test empty config
	emptyConfig := FilterConfig{}
	if filterID, exists := emptyConfig["filter_id"]; exists && filterID != "" {
		t.Errorf("Empty config should not have filter_id, got %v", filterID)
	}

	if active, exists := emptyConfig["active"]; exists && active == true {
		t.Errorf("Empty config should not have active=true, got %v", active)
	}

	if value, exists := emptyConfig["value"]; exists && value != nil {
		t.Errorf("Empty config should not have non-nil value, got %v", value)
	}
}
