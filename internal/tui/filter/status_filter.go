package filter

import (
	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/tui/preview"
)

// StatusFilter implements status-based filtering
type StatusFilter struct {
	*BaseFilter
	config StatusFilterConfig
}

// NewStatusFilter creates a new status filter
func NewStatusFilter() *StatusFilter {
	return &StatusFilter{
		BaseFilter: NewBaseFilter("ステータス", "実行ステータスでフィルタリング"),
		config: StatusFilterConfig{
			AllowedStatuses: map[ExecutionStatus]bool{
				StatusPending: true,
				StatusRunning: true,
				StatusSuccess: true,
				StatusFailed:  true,
				StatusSkipped: true,
			},
		},
	}
}

// Apply filters items based on status criteria
func (f *StatusFilter) Apply(items []interface{}) []interface{} {
	if !f.IsActive() {
		return items
	}

	var filtered []interface{}

	for _, item := range items {
		statusStr := f.getStatusFromItem(item)
		status := ExecutionStatus(statusStr)
		if f.config.AllowedStatuses[status] {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// getStatusFromItem extracts status from different item types
func (f *StatusFilter) getStatusFromItem(item interface{}) string {
	switch v := item.(type) {
	case *sandbox.ExecutionResult:
		if v.Success {
			return "success"
		} else if v.Skipped {
			return "skipped"
		} else {
			return "failed"
		}
	case *preview.CommandPreview:
		return "pending"
	case FilterableItem:
		return string(v.GetStatus())
	case string:
		return "pending"
	default:
		// Default to empty for unsupported types
		return ""
	}
}

// ToggleStatus toggles the allowed state of a status
func (f *StatusFilter) ToggleStatus(status ExecutionStatus) {
	var hasAllowed bool

	// Update status configuration under lock
	func() {
		f.BaseFilter.mutex.Lock()
		defer f.BaseFilter.mutex.Unlock()

		if f.config.AllowedStatuses[status] {
			f.config.AllowedStatuses[status] = false
		} else {
			f.config.AllowedStatuses[status] = true
		}

		// Check if any status is still allowed
		hasAllowed = false
		for _, allowed := range f.config.AllowedStatuses {
			if allowed {
				hasAllowed = true
				break
			}
		}
	}()

	// Set active status outside the lock to avoid deadlock
	f.SetActive(hasAllowed)
}

// SetStatusAllowed sets whether a status is allowed
func (f *StatusFilter) SetStatusAllowed(status ExecutionStatus, allowed bool) {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	f.config.AllowedStatuses[status] = allowed

	// Check if any status is still allowed
	hasAllowed := false
	for _, isAllowed := range f.config.AllowedStatuses {
		if isAllowed {
			hasAllowed = true
			break
		}
	}

	f.SetActive(hasAllowed)
}

// IsStatusAllowed returns whether a status is allowed
func (f *StatusFilter) IsStatusAllowed(status ExecutionStatus) bool {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()
	return f.config.AllowedStatuses[status]
}

// GetAllowedStatuses returns all allowed statuses
func (f *StatusFilter) GetAllowedStatuses() []ExecutionStatus {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()

	var allowed []ExecutionStatus
	for status, isAllowed := range f.config.AllowedStatuses {
		if isAllowed {
			allowed = append(allowed, status)
		}
	}
	return allowed
}

// GetAllStatuses returns all possible statuses
func (f *StatusFilter) GetAllStatuses() []ExecutionStatus {
	return []ExecutionStatus{
		StatusPending,
		StatusRunning,
		StatusSuccess,
		StatusFailed,
		StatusSkipped,
	}
}

// AllowAllStatuses allows all statuses
func (f *StatusFilter) AllowAllStatuses() {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	for status := range f.config.AllowedStatuses {
		f.config.AllowedStatuses[status] = true
	}
	f.SetActive(true)
}

// AllowOnlyStatus allows only the specified status
func (f *StatusFilter) AllowOnlyStatus(status ExecutionStatus) {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	for s := range f.config.AllowedStatuses {
		f.config.AllowedStatuses[s] = (s == status)
	}
	f.SetActive(true)
}

// AllowSuccessful allows only successful statuses
func (f *StatusFilter) AllowSuccessful() {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	f.config.AllowedStatuses[StatusPending] = false
	f.config.AllowedStatuses[StatusRunning] = false
	f.config.AllowedStatuses[StatusSuccess] = true
	f.config.AllowedStatuses[StatusFailed] = false
	f.config.AllowedStatuses[StatusSkipped] = false

	f.SetActive(true)
}

// AllowFailed allows only failed statuses
func (f *StatusFilter) AllowFailed() {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	f.config.AllowedStatuses[StatusPending] = false
	f.config.AllowedStatuses[StatusRunning] = false
	f.config.AllowedStatuses[StatusSuccess] = false
	f.config.AllowedStatuses[StatusFailed] = true
	f.config.AllowedStatuses[StatusSkipped] = false

	f.SetActive(true)
}

// DisallowAllStatuses disallows all statuses (effectively disabling the filter)
func (f *StatusFilter) DisallowAllStatuses() {
	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	for status := range f.config.AllowedStatuses {
		f.config.AllowedStatuses[status] = false
	}
	f.SetActive(false)
}

// GetConfig returns the filter configuration
func (f *StatusFilter) GetConfig() FilterConfig {
	f.BaseFilter.mutex.RLock()
	defer f.BaseFilter.mutex.RUnlock()

	var statuses []string
	for status, allowed := range f.config.AllowedStatuses {
		if allowed {
			statuses = append(statuses, string(status))
		}
	}

	return FilterConfig{
		"statuses": statuses,
	}
}

// SetConfig sets the filter configuration
func (f *StatusFilter) SetConfig(config FilterConfig) error {
	if err := f.BaseFilter.SetConfig(config); err != nil {
		return err
	}

	f.BaseFilter.mutex.Lock()
	defer f.BaseFilter.mutex.Unlock()

	if statuses, ok := config["statuses"].([]interface{}); ok {
		// Reset all statuses to false first
		for status := range f.config.AllowedStatuses {
			f.config.AllowedStatuses[status] = false
		}
		// Set allowed statuses to true
		for _, statusInterface := range statuses {
			if statusStr, ok := statusInterface.(string); ok {
				f.config.AllowedStatuses[ExecutionStatus(statusStr)] = true
			}
		}
	} else if statuses, ok := config["statuses"].([]string); ok {
		// Reset all statuses to false first
		for status := range f.config.AllowedStatuses {
			f.config.AllowedStatuses[status] = false
		}
		// Set allowed statuses to true
		for _, statusStr := range statuses {
			f.config.AllowedStatuses[ExecutionStatus(statusStr)] = true
		}
	}

	return nil
}

// GetStatusCount returns the number of items with each status
func (f *StatusFilter) GetStatusCount(items []interface{}) map[ExecutionStatus]int {
	counts := make(map[ExecutionStatus]int)

	for _, item := range items {
		statusStr := f.getStatusFromItem(item)
		status := ExecutionStatus(statusStr)
		counts[status]++
	}

	return counts
}

// GetStatusDisplayName returns a human-readable name for a status
func (f *StatusFilter) GetStatusDisplayName(status ExecutionStatus) string {
	switch status {
	case StatusPending:
		return "未実行"
	case StatusRunning:
		return "実行中"
	case StatusSuccess:
		return "成功"
	case StatusFailed:
		return "失敗"
	case StatusSkipped:
		return "スキップ"
	default:
		return string(status)
	}
}

// GetStatusIcon returns an icon for a status
func (f *StatusFilter) GetStatusIcon(status ExecutionStatus) string {
	switch status {
	case StatusPending:
		return "⏳"
	case StatusRunning:
		return "🔄"
	case StatusSuccess:
		return "✅"
	case StatusFailed:
		return "❌"
	case StatusSkipped:
		return "⏭️"
	default:
		return "❓"
	}
}

// Global helper functions

// GetStatusIcon returns an icon for a status string
func GetStatusIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "running":
		return "🔄"
	case "success":
		return "✅"
	case "failed":
		return "❌"
	case "skipped":
		return "⏭️"
	default:
		return "❓"
	}
}

// GetStatusDisplayName returns a human-readable name for a status string
func GetStatusDisplayName(status string) string {
	switch status {
	case "pending":
		return "待機中"
	case "running":
		return "実行中"
	case "success":
		return "成功"
	case "failed":
		return "失敗"
	case "skipped":
		return "スキップ"
	default:
		return "不明"
	}
}
