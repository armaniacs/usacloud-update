package filter

import "time"

// FilterSystem manages a collection of filters and their application
type FilterSystem struct {
	filters   []Filter
	activeSet *FilterSet
	presets   map[string]*FilterSet
	callback  func([]interface{})
}

// Filter defines the interface for all filter types
type Filter interface {
	Name() string
	Description() string
	Apply(items []interface{}) []interface{}
	IsActive() bool
	SetActive(bool)
	GetConfig() FilterConfig
	SetConfig(FilterConfig) error
}

// FilterSet represents a collection of filter configurations
type FilterSet struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Filters []FilterExport `json:"filters"`
	Created time.Time      `json:"created"`
}

// FilterConfig holds configuration data for a filter
type FilterConfig map[string]interface{}

// ExecutionStatus represents the status of command execution
type ExecutionStatus string

const (
	StatusPending ExecutionStatus = "pending"
	StatusRunning ExecutionStatus = "running"
	StatusSuccess ExecutionStatus = "success"
	StatusFailed  ExecutionStatus = "failed"
	StatusSkipped ExecutionStatus = "skipped"
)

// FilterableItem defines what items can be filtered
type FilterableItem interface {
	GetCategory() string
	GetStatus() ExecutionStatus
	GetSearchableText() []string
}

// TextSearchConfig holds configuration for text search filter
type TextSearchConfig struct {
	SearchTerm    string   `json:"search_term"`
	CaseSensitive bool     `json:"case_sensitive"`
	UseRegex      bool     `json:"use_regex"`
	Fields        []string `json:"fields"`
}

// CategoryFilterConfig holds configuration for category filter
type CategoryFilterConfig struct {
	SelectedCategories  map[string]bool `json:"selected_categories"`
	AvailableCategories []string        `json:"available_categories"`
}

// StatusFilterConfig holds configuration for status filter
type StatusFilterConfig struct {
	AllowedStatuses map[ExecutionStatus]bool `json:"allowed_statuses"`
}

// FilterStats holds statistics about filter performance
type FilterStats struct {
	TotalItems     int           `json:"total_items"`
	FilteredItems  int           `json:"filtered_items"`
	FilterDuration time.Duration `json:"filter_duration"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// FilterStatistics holds detailed filter statistics
type FilterStatistics struct {
	TotalItems     int            `json:"total_items"`
	FilteredItems  int            `json:"filtered_items"`
	CategoryCounts map[string]int `json:"category_counts"`
	StatusCounts   map[string]int `json:"status_counts"`
	LastUpdated    time.Time      `json:"last_updated"`
}

// FilterExport represents exported filter configuration
type FilterExport struct {
	Name   string       `json:"name"`
	Active bool         `json:"active"`
	Config FilterConfig `json:"config"`
}

// FilterEvent represents events in the filter system
type FilterEvent struct {
	Type      FilterEventType `json:"type"`
	FilterID  string          `json:"filter_id"`
	Timestamp time.Time       `json:"timestamp"`
	Data      interface{}     `json:"data"`
}

// FilterEventType defines types of filter events
type FilterEventType string

const (
	EventFilterActivated   FilterEventType = "filter_activated"
	EventFilterDeactivated FilterEventType = "filter_deactivated"
	EventFilterConfigured  FilterEventType = "filter_configured"
	EventFilterApplied     FilterEventType = "filter_applied"
	EventPresetSaved       FilterEventType = "preset_saved"
	EventPresetLoaded      FilterEventType = "preset_loaded"
)
