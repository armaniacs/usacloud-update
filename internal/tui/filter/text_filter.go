package filter

import (
	"regexp"
	"strings"

	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

// TextFilter implements text-based filtering
type TextFilter struct {
	*BaseFilter
	query         string
	caseSensitive bool
	useRegex      bool
	searchFields  []string
}

// NewTextFilter creates a new text filter
func NewTextFilter() *TextFilter {
	return &TextFilter{
		BaseFilter:   NewBaseFilter("テキスト検索", "テキスト検索でフィルタリング"),
		searchFields: []string{"command", "description", "output"},
	}
}

// TextSearchFilter is an alias for TextFilter for backward compatibility
type TextSearchFilter = TextFilter

// NewTextSearchFilter creates a new text search filter
func NewTextSearchFilter() *TextSearchFilter {
	return &TextSearchFilter{
		BaseFilter:   NewBaseFilter("text-search", "Filter items by text search"),
		searchFields: []string{"command", "description", "output"},
	}
}

// Apply filters items based on text search criteria
func (f *TextFilter) Apply(items []interface{}) []interface{} {
	if !f.IsActive() || f.query == "" {
		return items
	}

	var filtered []interface{}

	for _, item := range items {
		if f.matchesItem(item) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// matchesItem checks if an item matches the search criteria
func (f *TextFilter) matchesItem(item interface{}) bool {
	searchTexts := f.getSearchableTexts(item)

	for _, text := range searchTexts {
		if f.searchInText(text) {
			return true
		}
	}

	return false
}

// getSearchableTexts extracts searchable text from different item types
func (f *TextFilter) getSearchableTexts(item interface{}) []string {
	var texts []string

	switch v := item.(type) {
	case *preview.CommandPreview:
		texts = append(texts, v.Original, v.Transformed, v.Description)
		for _, warning := range v.Warnings {
			texts = append(texts, warning)
		}
		if v.Impact != nil {
			texts = append(texts, v.Impact.Description)
		}

	case *sandbox.ExecutionResult:
		texts = append(texts, v.Command, v.Output, v.Error, v.SkipReason)

	case FilterableItem:
		texts = v.GetSearchableText()

	default:
		// Try to extract text from string representation
		if str, ok := item.(string); ok {
			texts = append(texts, str)
		}
	}

	return texts
}

// searchInText performs the actual text search
func (f *TextFilter) searchInText(text string) bool {
	if text == "" {
		return false
	}

	searchTerm := f.query

	if !f.caseSensitive {
		text = strings.ToLower(text)
		searchTerm = strings.ToLower(searchTerm)
	}

	if f.useRegex {
		matched, err := regexp.MatchString(searchTerm, text)
		if err != nil {
			// Fallback to simple string search if regex is invalid
			return strings.Contains(text, searchTerm)
		}
		return matched
	}

	return strings.Contains(text, searchTerm)
}

// SetSearchTerm sets the search term
func (f *TextFilter) SetSearchTerm(term string) {
	var shouldBeActive bool

	// Update query under lock
	func() {
		f.BaseFilter.mutex.Lock()
		defer f.BaseFilter.mutex.Unlock()
		f.query = term
		shouldBeActive = term != ""
	}()

	// Set active status outside the lock to avoid deadlock
	f.SetActive(shouldBeActive)
}

// GetSearchTerm returns the current search term
func (f *TextFilter) GetSearchTerm() string {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()
	return f.query
}

// SetCaseSensitive sets case sensitivity
func (f *TextFilter) SetCaseSensitive(sensitive bool) {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()
	f.caseSensitive = sensitive
}

// IsCaseSensitive returns whether search is case sensitive
func (f *TextFilter) IsCaseSensitive() bool {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()
	return f.caseSensitive
}

// SetUseRegex sets whether to use regex
func (f *TextFilter) SetUseRegex(useRegex bool) {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()
	f.useRegex = useRegex
}

// IsUsingRegex returns whether regex is enabled
func (f *TextFilter) IsUsingRegex() bool {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()
	return f.useRegex
}

// GetConfig returns the filter configuration
func (f *TextFilter) GetConfig() FilterConfig {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()

	return FilterConfig{
		"query":          f.query,
		"case_sensitive": f.caseSensitive,
		"use_regex":      f.useRegex,
		"regex_mode":     f.useRegex,
		"search_fields":  f.searchFields,
	}
}

// SetConfig sets the filter configuration
func (f *TextFilter) SetConfig(config FilterConfig) error {
	if err := f.BaseFilter.SetConfig(config); err != nil {
		return err
	}

	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	if query, ok := config["query"].(string); ok {
		f.query = query
	}
	if caseSensitive, ok := config["case_sensitive"].(bool); ok {
		f.caseSensitive = caseSensitive
	}
	if useRegex, ok := config["use_regex"].(bool); ok {
		f.useRegex = useRegex
	}
	if regexMode, ok := config["regex_mode"].(bool); ok {
		f.useRegex = regexMode
	}
	if fields, ok := config["search_fields"].([]interface{}); ok {
		f.searchFields = make([]string, len(fields))
		for i, field := range fields {
			if str, ok := field.(string); ok {
				f.searchFields[i] = str
			}
		}
	} else if fields, ok := config["search_fields"].([]string); ok {
		f.searchFields = fields
	} else if fields, ok := config["fields"].([]interface{}); ok {
		f.searchFields = make([]string, len(fields))
		for i, field := range fields {
			if str, ok := field.(string); ok {
				f.searchFields[i] = str
			}
		}
	} else if fields, ok := config["fields"].([]string); ok {
		f.searchFields = fields
	}

	return nil
}

// ClearSearch clears the search term and deactivates the filter
func (f *TextFilter) ClearSearch() {
	// Update query under lock
	func() {
		f.BaseFilter.mutex.Lock()
		defer f.BaseFilter.mutex.Unlock()
		f.query = ""
	}()

	// Set active status outside the lock to avoid deadlock
	f.SetActive(false)
}

// GetFields returns the search fields
func (f *TextFilter) GetFields() []string {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()
	return append([]string(nil), f.searchFields...)
}

// SetFields sets the search fields
func (f *TextFilter) SetFields(fields []string) {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()
	f.searchFields = append([]string(nil), fields...)
}
