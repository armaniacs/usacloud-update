package filter

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

// NewFilterSystem creates a new filter system with default filters
func NewFilterSystem() *FilterSystem {
	fs := &FilterSystem{
		filters: make([]Filter, 0),
		presets: make(map[string]*FilterSet),
	}

	// Add default filters
	fs.AddFilter(NewTextFilter())
	fs.AddFilter(NewCategoryFilter())
	fs.AddFilter(NewStatusFilter())

	return fs
}

// AddFilter adds a filter to the system
func (fs *FilterSystem) AddFilter(filter Filter) {
	fs.filters = append(fs.filters, filter)
}

// GetFilter retrieves a filter by name
func (fs *FilterSystem) GetFilter(name string) Filter {
	for _, filter := range fs.filters {
		if filter.Name() == name {
			return filter
		}
	}
	return nil
}

// Apply applies all active filters to the given items
func (fs *FilterSystem) Apply(items []interface{}) []interface{} {
	// First filter out nil items
	var nonNilItems []interface{}
	for _, item := range items {
		if item != nil {
			nonNilItems = append(nonNilItems, item)
		}
	}

	result := nonNilItems

	for _, filter := range fs.filters {
		if filter.IsActive() {
			result = filter.Apply(result)
		}
	}

	if fs.callback != nil {
		fs.callback(result)
	}

	return result
}

// ApplyFilters is an alias for Apply for compatibility
func (fs *FilterSystem) ApplyFilters(items []interface{}) []interface{} {
	return fs.Apply(items)
}

// GetFilters returns all filters in the system
func (fs *FilterSystem) GetFilters() []Filter {
	return fs.filters
}

// RemoveFilter removes a filter by name
func (fs *FilterSystem) RemoveFilter(name string) {
	for i, filter := range fs.filters {
		if filter.Name() == name {
			fs.filters = append(fs.filters[:i], fs.filters[i+1:]...)
			break
		}
	}
}

// SetCallback sets the callback function for filter updates
func (fs *FilterSystem) SetCallback(callback func([]interface{})) {
	fs.callback = callback
}

// GetActiveFilters returns all currently active filters
func (fs *FilterSystem) GetActiveFilters() []Filter {
	var active []Filter
	for _, filter := range fs.filters {
		if filter.IsActive() {
			active = append(active, filter)
		}
	}
	return active
}

// ClearAllFilters deactivates all filters
func (fs *FilterSystem) ClearAllFilters() {
	for _, filter := range fs.filters {
		filter.SetActive(false)
	}
}

// ExportConfig exports current filter configuration
func (fs *FilterSystem) ExportConfig() []FilterExport {
	exports := make([]FilterExport, 0)
	for _, filter := range fs.filters {
		if filter.IsActive() {
			exports = append(exports, FilterExport{
				Name:   filter.Name(),
				Active: filter.IsActive(),
				Config: filter.GetConfig(),
			})
		}
	}
	return exports
}

// ImportConfig imports filter configuration
func (fs *FilterSystem) ImportConfig(exports []FilterExport) error {
	for _, export := range exports {
		filter := fs.GetFilter(export.Name)
		if filter == nil {
			return fmt.Errorf("filter '%s' not found", export.Name)
		}

		filter.SetActive(export.Active)
		if err := filter.SetConfig(export.Config); err != nil {
			return fmt.Errorf("failed to set config for filter %s: %w", export.Name, err)
		}
	}
	return nil
}

// SaveConfig saves the current filter configuration to a file
func (fs *FilterSystem) SaveConfig(path string) error {
	exports := fs.ExportConfig()
	data, err := json.MarshalIndent(exports, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// LoadConfig loads filter configuration from a file
func (fs *FilterSystem) LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var exports []FilterExport
	if err := json.Unmarshal(data, &exports); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return fs.ImportConfig(exports)
}

// GetStats returns statistics about filter performance
func (fs *FilterSystem) GetStats(items []interface{}) *FilterStats {
	start := time.Now()
	filtered := fs.Apply(items)
	duration := time.Since(start)

	return &FilterStats{
		TotalItems:     len(items),
		FilteredItems:  len(filtered),
		FilterDuration: duration,
		LastUpdated:    time.Now(),
	}
}

// GetStatistics returns detailed statistics including category and status counts
func (fs *FilterSystem) GetStatistics(items []interface{}) *FilterStatistics {
	categoryCounts := make(map[string]int)
	statusCounts := make(map[string]int)

	for _, item := range items {
		// Count categories
		var category string
		switch v := item.(type) {
		case *preview.CommandPreview:
			category = v.Category
		case *sandbox.ExecutionResult:
			// Extract category from command if possible
			category = "command"
		default:
			category = "other"
		}
		if category == "" {
			category = "other"
		}
		categoryCounts[category]++

		// Count statuses
		var status string
		switch v := item.(type) {
		case *sandbox.ExecutionResult:
			if v.Success {
				status = "success"
			} else if v.Skipped {
				status = "skipped"
			} else {
				status = "failed"
			}
		default:
			status = "pending"
		}
		statusCounts[status]++
	}

	return &FilterStatistics{
		TotalItems:     len(items),
		FilteredItems:  len(items), // This will be updated by calling code if filters are applied
		CategoryCounts: categoryCounts,
		StatusCounts:   statusCounts,
		LastUpdated:    time.Now(),
	}
}

// BaseFilter provides common functionality for all filters
type BaseFilter struct {
	name        string
	description string
	active      bool
	mutex       sync.RWMutex
}

// NewBaseFilter creates a new base filter
func NewBaseFilter(name, description string) *BaseFilter {
	return &BaseFilter{
		name:        name,
		description: description,
		active:      false,
	}
}

// Name returns the filter name
func (bf *BaseFilter) Name() string {
	return bf.name
}

// Description returns the filter description
func (bf *BaseFilter) Description() string {
	return bf.description
}

// IsActive returns whether the filter is active
func (bf *BaseFilter) IsActive() bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()
	return bf.active
}

// SetActive sets the filter active state
func (bf *BaseFilter) SetActive(active bool) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	bf.active = active
}

// GetConfig returns the basic filter configuration
func (bf *BaseFilter) GetConfig() FilterConfig {
	return FilterConfig{
		"name":   bf.name,
		"active": bf.IsActive(),
	}
}

// SetConfig sets the basic filter configuration
func (bf *BaseFilter) SetConfig(config FilterConfig) error {
	if name, ok := config["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != bf.name {
			return fmt.Errorf("filter name mismatch: expected %s, got %s", bf.name, nameStr)
		}
	}

	if active, ok := config["active"]; ok {
		if activeBool, ok := active.(bool); ok {
			bf.SetActive(activeBool)
		}
	}
	return nil
}

// Apply is a default implementation that returns items unchanged
func (bf *BaseFilter) Apply(items []interface{}) []interface{} {
	return items
}
