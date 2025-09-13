package security

import (
	"fmt"
	"sync"
	"time"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Resource    string                 `json:"resource"`
	CreatedAt   time.Time              `json:"created_at"`
	Remediation string                 `json:"remediation"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// CredentialMonitor monitors credential security and generates alerts
type CredentialMonitor struct {
	mu              sync.RWMutex
	storage         *SecureStorage
	auditLogger     *AuditLogger
	alertThresholds map[string]time.Duration
	config          *MonitorConfig
	alerts          []SecurityAlert
}

// MonitorConfig provides configuration for credential monitoring
type MonitorConfig struct {
	DefaultAgeThreshold  time.Duration
	UnusedThreshold      time.Duration
	MaxAlerts            int
	CheckInterval        time.Duration
	AlertRetentionPeriod time.Duration
	EnableNotifications  bool
	NotificationChannels []string
}

// DefaultMonitorConfig returns default monitor configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		DefaultAgeThreshold:  90 * 24 * time.Hour, // 90 days
		UnusedThreshold:      30 * 24 * time.Hour, // 30 days
		MaxAlerts:            1000,
		CheckInterval:        24 * time.Hour,      // daily
		AlertRetentionPeriod: 30 * 24 * time.Hour, // 30 days
		EnableNotifications:  true,
		NotificationChannels: []string{"log"},
	}
}

// NewCredentialMonitor creates a new credential monitor
func NewCredentialMonitor(storage *SecureStorage, auditLogger *AuditLogger, config *MonitorConfig) *CredentialMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	monitor := &CredentialMonitor{
		storage:         storage,
		auditLogger:     auditLogger,
		alertThresholds: make(map[string]time.Duration),
		config:          config,
		alerts:          make([]SecurityAlert, 0),
	}

	// Start monitoring routine if check interval is set
	if config.CheckInterval > 0 {
		go monitor.monitoringRoutine()
	}

	return monitor
}

// CheckCredentialAge checks for aging credentials and generates alerts
func (cm *CredentialMonitor) CheckCredentialAge() []SecurityAlert {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var alerts []SecurityAlert

	credentials, err := cm.storage.ListCredentials()
	if err != nil {
		cm.logError("failed to list credentials for age check", err)
		return alerts
	}

	for _, cred := range credentials {
		age := time.Since(cred.CreatedAt)
		threshold := cm.getThresholdForCredential(cred.Key)

		if age > threshold {
			alert := SecurityAlert{
				Type:        "credential_age",
				Severity:    cm.getSeverityForAge(age, threshold),
				Message:     fmt.Sprintf("認証情報 '%s' が%d日前に作成されました。更新を検討してください。", cred.Key, int(age.Hours()/24)),
				Resource:    cred.Key,
				CreatedAt:   time.Now(),
				Remediation: "新しい認証情報を生成し、設定を更新してください。",
				Details: map[string]interface{}{
					"age_days":       int(age.Hours() / 24),
					"threshold_days": int(threshold.Hours() / 24),
					"created_at":     cred.CreatedAt,
				},
			}
			alerts = append(alerts, alert)

			cm.auditLogger.LogSecurityViolation("credential_age_warning", map[string]interface{}{
				"credential_key": cred.Key,
				"age_days":       int(age.Hours() / 24),
				"threshold_days": int(threshold.Hours() / 24),
			})
		}
	}

	cm.addAlerts(alerts)
	return alerts
}

// CheckUnusedCredentials checks for unused credentials and generates alerts
func (cm *CredentialMonitor) CheckUnusedCredentials() []SecurityAlert {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var alerts []SecurityAlert

	credentials, err := cm.storage.ListCredentials()
	if err != nil {
		cm.logError("failed to list credentials for unused check", err)
		return alerts
	}

	for _, cred := range credentials {
		unused := time.Since(cred.LastUsedAt)

		if unused > cm.config.UnusedThreshold {
			alert := SecurityAlert{
				Type:        "unused_credential",
				Severity:    "info",
				Message:     fmt.Sprintf("認証情報 '%s' が%d日間使用されていません。", cred.Key, int(unused.Hours()/24)),
				Resource:    cred.Key,
				CreatedAt:   time.Now(),
				Remediation: "不要な場合は削除を検討してください。",
				Details: map[string]interface{}{
					"unused_days":  int(unused.Hours() / 24),
					"last_used_at": cred.LastUsedAt,
				},
			}
			alerts = append(alerts, alert)

			cm.auditLogger.LogSecurityViolation("unused_credential_warning", map[string]interface{}{
				"credential_key": cred.Key,
				"unused_days":    int(unused.Hours() / 24),
			})
		}
	}

	cm.addAlerts(alerts)
	return alerts
}

// CheckExpiredCredentials checks for credentials that should be rotated
func (cm *CredentialMonitor) CheckExpiredCredentials() []SecurityAlert {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var alerts []SecurityAlert

	credentials, err := cm.storage.ListCredentials()
	if err != nil {
		cm.logError("failed to list credentials for expiry check", err)
		return alerts
	}

	for _, cred := range credentials {
		age := time.Since(cred.CreatedAt)
		threshold := cm.getThresholdForCredential(cred.Key)

		// Check if credential is approaching expiry (90% of threshold)
		warningThreshold := time.Duration(float64(threshold) * 0.9)

		if age > warningThreshold && age <= threshold {
			alert := SecurityAlert{
				Type:        "credential_expiry_warning",
				Severity:    "warning",
				Message:     fmt.Sprintf("認証情報 '%s' の有効期限が近づいています（作成から%d日）。", cred.Key, int(age.Hours()/24)),
				Resource:    cred.Key,
				CreatedAt:   time.Now(),
				Remediation: "早めに認証情報の更新を行ってください。",
				Details: map[string]interface{}{
					"age_days":          int(age.Hours() / 24),
					"threshold_days":    int(threshold.Hours() / 24),
					"days_until_expiry": int((threshold - age).Hours() / 24),
				},
			}
			alerts = append(alerts, alert)
		}
	}

	cm.addAlerts(alerts)
	return alerts
}

// CheckCredentialVersions checks for version-related issues
func (cm *CredentialMonitor) CheckCredentialVersions() []SecurityAlert {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var alerts []SecurityAlert

	credentials, err := cm.storage.ListCredentials()
	if err != nil {
		cm.logError("failed to list credentials for version check", err)
		return alerts
	}

	for _, cred := range credentials {
		// Check for old versions (version 1 credentials older than 30 days)
		if cred.Version == 1 && time.Since(cred.CreatedAt) > 30*24*time.Hour {
			alert := SecurityAlert{
				Type:        "credential_version_old",
				Severity:    "info",
				Message:     fmt.Sprintf("認証情報 '%s' はバージョン%dで、30日以上更新されていません。", cred.Key, cred.Version),
				Resource:    cred.Key,
				CreatedAt:   time.Now(),
				Remediation: "セキュリティ向上のため、認証情報の更新を検討してください。",
				Details: map[string]interface{}{
					"version":    cred.Version,
					"created_at": cred.CreatedAt,
				},
			}
			alerts = append(alerts, alert)
		}
	}

	cm.addAlerts(alerts)
	return alerts
}

// RunAllChecks runs all credential checks and returns combined alerts
func (cm *CredentialMonitor) RunAllChecks() []SecurityAlert {
	var allAlerts []SecurityAlert

	allAlerts = append(allAlerts, cm.CheckCredentialAge()...)
	allAlerts = append(allAlerts, cm.CheckUnusedCredentials()...)
	allAlerts = append(allAlerts, cm.CheckExpiredCredentials()...)
	allAlerts = append(allAlerts, cm.CheckCredentialVersions()...)

	return allAlerts
}

// GetAlerts returns current alerts with optional filtering
func (cm *CredentialMonitor) GetAlerts(filter *AlertFilter) []SecurityAlert {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if filter == nil {
		return cm.alerts
	}

	var filtered []SecurityAlert
	for _, alert := range cm.alerts {
		if filter.Matches(alert) {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// ClearAlerts clears alerts based on filter criteria
func (cm *CredentialMonitor) ClearAlerts(filter *AlertFilter) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if filter == nil {
		count := len(cm.alerts)
		cm.alerts = cm.alerts[:0]
		return count
	}

	var remaining []SecurityAlert
	cleared := 0

	for _, alert := range cm.alerts {
		if filter.Matches(alert) {
			cleared++
		} else {
			remaining = append(remaining, alert)
		}
	}

	cm.alerts = remaining
	return cleared
}

// SetThreshold sets age threshold for a specific credential
func (cm *CredentialMonitor) SetThreshold(credentialKey string, threshold time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.alertThresholds[credentialKey] = threshold
}

// GetThreshold gets age threshold for a specific credential
func (cm *CredentialMonitor) GetThreshold(credentialKey string) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.getThresholdForCredential(credentialKey)
}

// getThresholdForCredential gets threshold for credential (internal, no lock)
func (cm *CredentialMonitor) getThresholdForCredential(credentialKey string) time.Duration {
	if threshold, exists := cm.alertThresholds[credentialKey]; exists {
		return threshold
	}
	return cm.config.DefaultAgeThreshold
}

// getSeverityForAge determines severity based on age and threshold
func (cm *CredentialMonitor) getSeverityForAge(age, threshold time.Duration) string {
	ratio := float64(age) / float64(threshold)

	switch {
	case ratio >= 2.0:
		return "critical"
	case ratio >= 1.5:
		return "high"
	case ratio >= 1.2:
		return "warning"
	default:
		return "info"
	}
}

// addAlerts adds alerts to the monitor (internal, must hold lock)
func (cm *CredentialMonitor) addAlerts(alerts []SecurityAlert) {
	cm.alerts = append(cm.alerts, alerts...)

	// Trim alerts if we exceed max
	if len(cm.alerts) > cm.config.MaxAlerts {
		// Keep the most recent alerts
		start := len(cm.alerts) - cm.config.MaxAlerts
		cm.alerts = cm.alerts[start:]
	}

	// Clean up old alerts
	cm.cleanupOldAlerts()
}

// cleanupOldAlerts removes alerts older than retention period
func (cm *CredentialMonitor) cleanupOldAlerts() {
	cutoff := time.Now().Add(-cm.config.AlertRetentionPeriod)
	var remaining []SecurityAlert

	for _, alert := range cm.alerts {
		if alert.CreatedAt.After(cutoff) {
			remaining = append(remaining, alert)
		}
	}

	cm.alerts = remaining
}

// logError logs an error to the audit logger
func (cm *CredentialMonitor) logError(message string, err error) {
	if cm.auditLogger != nil {
		cm.auditLogger.LogSystemEvent("monitor_error", "error", "credential_monitor", "error", map[string]interface{}{
			"message": message,
			"error":   err.Error(),
		})
	}
}

// monitoringRoutine runs periodic checks
func (cm *CredentialMonitor) monitoringRoutine() {
	ticker := time.NewTicker(cm.config.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		cm.RunAllChecks()
	}
}

// AlertFilter provides filtering for security alerts
type AlertFilter struct {
	Types      []string
	Severities []string
	Resources  []string
	TimeRange  *TimeRange
	MaxAge     time.Duration
}

// Matches checks if an alert matches the filter criteria
func (filter *AlertFilter) Matches(alert SecurityAlert) bool {
	if len(filter.Types) > 0 {
		found := false
		for _, alertType := range filter.Types {
			if alert.Type == alertType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.Severities) > 0 {
		found := false
		for _, severity := range filter.Severities {
			if alert.Severity == severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.Resources) > 0 {
		found := false
		for _, resource := range filter.Resources {
			if alert.Resource == resource {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if filter.TimeRange != nil {
		if !filter.TimeRange.Start.IsZero() && alert.CreatedAt.Before(filter.TimeRange.Start) {
			return false
		}

		if !filter.TimeRange.End.IsZero() && alert.CreatedAt.After(filter.TimeRange.End) {
			return false
		}
	}

	if filter.MaxAge > 0 && time.Since(alert.CreatedAt) > filter.MaxAge {
		return false
	}

	return true
}

// GetStatistics returns monitoring statistics
func (cm *CredentialMonitor) GetStatistics() *MonitorStatistics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := &MonitorStatistics{
		TotalAlerts:      len(cm.alerts),
		AlertsByType:     make(map[string]int),
		AlertsBySeverity: make(map[string]int),
	}

	for _, alert := range cm.alerts {
		stats.AlertsByType[alert.Type]++
		stats.AlertsBySeverity[alert.Severity]++
	}

	return stats
}

// MonitorStatistics provides monitoring statistics
type MonitorStatistics struct {
	TotalAlerts      int            `json:"total_alerts"`
	AlertsByType     map[string]int `json:"alerts_by_type"`
	AlertsBySeverity map[string]int `json:"alerts_by_severity"`
}
