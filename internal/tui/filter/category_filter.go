package filter

import (
	"fmt"
	"strings"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

// CategoryFilter implements category-based filtering
type CategoryFilter struct {
	*BaseFilter
	config CategoryFilterConfig
}

// NewCategoryFilter creates a new category filter
func NewCategoryFilter() *CategoryFilter {
	return &CategoryFilter{
		BaseFilter: NewBaseFilter("カテゴリ", "コマンドのカテゴリでフィルタリング"),
		config: CategoryFilterConfig{
			SelectedCategories: make(map[string]bool),
			AvailableCategories: []string{
				"infrastructure",
				"storage",
				"network",
				"managed-service",
				"security",
				"monitoring",
				"other",
			},
		},
	}
}

// Apply filters items based on category criteria
func (f *CategoryFilter) Apply(items []interface{}) []interface{} {
	if !f.IsActive() || len(f.config.SelectedCategories) == 0 {
		return items
	}

	var filtered []interface{}

	for _, item := range items {
		category := f.getCategoryFromItem(item)
		if f.config.SelectedCategories[category] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// getCategoryFromItem extracts category from different item types
func (f *CategoryFilter) getCategoryFromItem(item interface{}) string {
	switch v := item.(type) {
	case *preview.CommandPreview:
		return v.Category

	case *sandbox.ExecutionResult:
		return f.categorizeCommand(v.Command)

	case FilterableItem:
		return v.GetCategory()

	default:
		if str, ok := item.(string); ok {
			return f.categorizeCommand(str)
		}
		return "other"
	}
}

// categorizeCommand categorizes a command string
func (f *CategoryFilter) categorizeCommand(command string) string {
	if command == "" {
		return "other"
	}

	command = strings.ToLower(strings.TrimSpace(command))
	parts := strings.Fields(command)

	if len(parts) < 2 {
		return "other"
	}

	// Skip "usacloud" prefix
	if parts[0] == "usacloud" && len(parts) > 2 {
		parts = parts[1:]
	}

	subcommand := parts[0]

	// Infrastructure resources
	infrastructureCommands := []string{
		"server", "interface", "bridge", "packet-filter", "ipaddress", "internet",
	}
	for _, cmd := range infrastructureCommands {
		if subcommand == cmd {
			return "infrastructure"
		}
	}

	// Storage resources
	storageCommands := []string{
		"disk", "archive", "cdrom", "note", "iso-image", "startup-script",
		"volume", "backup",
	}
	for _, cmd := range storageCommands {
		if subcommand == cmd {
			return "storage"
		}
	}

	// Network resources
	networkCommands := []string{
		"switch", "router", "dns", "gslb", "proxylb", "loadbalancer", "subnet",
		"vpc", "nat", "firewall",
	}
	for _, cmd := range networkCommands {
		if subcommand == cmd {
			return "network"
		}
	}

	// Managed services
	managedServiceCommands := []string{
		"database", "nfs", "mariadb", "postgres", "mysql",
		"redis", "memcached",
	}
	for _, cmd := range managedServiceCommands {
		if subcommand == cmd {
			return "managed-service"
		}
	}

	// Security resources
	securityCommands := []string{
		"certificateauthority", "certificate", "cert", "ssl", "ssh-key", "private-host",
		"auth", "token", "license",
	}
	for _, cmd := range securityCommands {
		if subcommand == cmd {
			return "security"
		}
	}

	// Monitoring resources
	monitoringCommands := []string{
		"bill", "monitor", "activity", "metric", "log", "alert",
		"dashboard", "graph",
	}
	for _, cmd := range monitoringCommands {
		if subcommand == cmd {
			return "monitoring"
		}
	}

	return "other"
}

// ToggleCategory toggles selection of a category
func (f *CategoryFilter) ToggleCategory(category string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.config.SelectedCategories[category] {
		delete(f.config.SelectedCategories, category)
	} else {
		f.config.SelectedCategories[category] = true
	}

	f.SetActive(len(f.config.SelectedCategories) > 0)
}

// SelectCategory selects a category
func (f *CategoryFilter) SelectCategory(category string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.config.SelectedCategories[category] = true
	f.SetActive(true)
}

// DeselectCategory deselects a category
func (f *CategoryFilter) DeselectCategory(category string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.config.SelectedCategories, category)
	f.SetActive(len(f.config.SelectedCategories) > 0)
}

// IsCategorySelected returns whether a category is selected
func (f *CategoryFilter) IsCategorySelected(category string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.config.SelectedCategories[category]
}

// GetSelectedCategories returns all selected categories
func (f *CategoryFilter) GetSelectedCategories() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var selected []string
	for category := range f.config.SelectedCategories {
		selected = append(selected, category)
	}
	return selected
}

// GetAvailableCategories returns all available categories
func (f *CategoryFilter) GetAvailableCategories() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return append([]string(nil), f.config.AvailableCategories...)
}

// SelectAllCategories selects all available categories
func (f *CategoryFilter) SelectAllCategories() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for _, category := range f.config.AvailableCategories {
		f.config.SelectedCategories[category] = true
	}
	f.SetActive(true)
}

// ClearSelection clears all category selections
func (f *CategoryFilter) ClearSelection() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.config.SelectedCategories = make(map[string]bool)
	f.SetActive(false)
}

// GetConfig returns the filter configuration
func (f *CategoryFilter) GetConfig() FilterConfig {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	var categories []string
	for category, selected := range f.config.SelectedCategories {
		if selected {
			categories = append(categories, category)
		}
	}

	return FilterConfig{
		"categories": categories,
	}
}

// SetConfig sets the filter configuration
func (f *CategoryFilter) SetConfig(config FilterConfig) error {
	if err := f.BaseFilter.SetConfig(config); err != nil {
		return err
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if categoriesValue, exists := config["categories"]; exists {
		if categories, ok := categoriesValue.([]interface{}); ok {
			// Reset all categories to false first
			for category := range f.config.SelectedCategories {
				f.config.SelectedCategories[category] = false
			}
			// Set selected categories to true
			for _, categoryInterface := range categories {
				if categoryStr, ok := categoryInterface.(string); ok {
					f.config.SelectedCategories[categoryStr] = true
				}
			}
		} else if categories, ok := categoriesValue.([]string); ok {
			// Reset all categories to false first
			for category := range f.config.SelectedCategories {
				f.config.SelectedCategories[category] = false
			}
			// Set selected categories to true
			for _, categoryStr := range categories {
				f.config.SelectedCategories[categoryStr] = true
			}
		} else {
			return fmt.Errorf("categories must be []string or []interface{}, got %T", categoriesValue)
		}
	}

	return nil
}

// SetAvailableCategories sets the list of available categories
func (f *CategoryFilter) SetAvailableCategories(categories []string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.config.AvailableCategories = append([]string(nil), categories...)
}

// GetCategoryCount returns the number of items in each category
func (f *CategoryFilter) GetCategoryCount(items []interface{}) map[string]int {
	counts := make(map[string]int)

	for _, item := range items {
		category := f.getCategoryFromItem(item)
		counts[category]++
	}

	return counts
}
