package security

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
	ClientIP  string                 `json:"client_ip,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	mu         sync.RWMutex
	logger     *log.Logger
	logFile    *os.File
	filter     *SensitiveDataFilter
	config     *AuditConfig
	buffer     []AuditEvent
	bufferSize int
}

// AuditConfig provides configuration for audit logging
type AuditConfig struct {
	LogFile       string
	BufferSize    int
	FlushInterval time.Duration
	MaxFileSize   int64
	RotateFiles   bool
	MaxFiles      int
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		LogFile:       "audit.log",
		BufferSize:    100,
		FlushInterval: 5 * time.Second,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		RotateFiles:   true,
		MaxFiles:      10,
	}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config *AuditConfig) (*AuditLogger, error) {
	if config == nil {
		config = DefaultAuditConfig()
	}

	// Create log directory if it doesn't exist
	logDir := filepath.Dir(config.LogFile)
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	filter := NewSensitiveDataFilter()
	secureWriter := NewSecureLogWriter(logFile, filter)
	logger := log.New(secureWriter, "", 0)

	auditLogger := &AuditLogger{
		logger:     logger,
		logFile:    logFile,
		filter:     filter,
		config:     config,
		buffer:     make([]AuditEvent, 0, config.BufferSize),
		bufferSize: config.BufferSize,
	}

	// Start flush routine if buffering is enabled
	if config.BufferSize > 0 && config.FlushInterval > 0 {
		go auditLogger.flushRoutine()
	}

	return auditLogger, nil
}

// LogCredentialAccess logs credential access events
func (al *AuditLogger) LogCredentialAccess(credentialKey, action string) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "credential_access",
		Action:    action,
		Resource:  credentialKey,
		Status:    "success",
		Details: map[string]interface{}{
			"credential_type": "sakura_cloud_api",
		},
	}

	al.logEvent(event)
}

// LogCredentialRotation logs credential rotation events
func (al *AuditLogger) LogCredentialRotation(credentialKey string, oldVersion, newVersion int) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "credential_rotation",
		Action:    "rotate",
		Resource:  credentialKey,
		Status:    "success",
		Details: map[string]interface{}{
			"old_version": oldVersion,
			"new_version": newVersion,
		},
	}

	al.logEvent(event)
}

// LogSecurityViolation logs security violation events
func (al *AuditLogger) LogSecurityViolation(violation string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "security_violation",
		Action:    "alert",
		Resource:  "system",
		Status:    "violation",
		Details:   details,
	}

	al.logEvent(event)
}

// LogAuthenticationEvent logs authentication events
func (al *AuditLogger) LogAuthenticationEvent(userID, action, status string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "authentication",
		UserID:    userID,
		Action:    action,
		Resource:  "auth_system",
		Status:    status,
		Details:   details,
	}

	al.logEvent(event)
}

// LogConfigurationChange logs configuration change events
func (al *AuditLogger) LogConfigurationChange(resource, action string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: "configuration_change",
		Action:    action,
		Resource:  resource,
		Status:    "success",
		Details:   details,
	}

	al.logEvent(event)
}

// LogSystemEvent logs general system events
func (al *AuditLogger) LogSystemEvent(eventType, action, resource, status string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		Action:    action,
		Resource:  resource,
		Status:    status,
		Details:   details,
	}

	al.logEvent(event)
}

// logEvent logs a single audit event
func (al *AuditLogger) logEvent(event AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.bufferSize > 0 {
		al.buffer = append(al.buffer, event)
		if len(al.buffer) >= al.bufferSize {
			al.flushBuffer()
		}
	} else {
		al.writeEvent(event)
	}
}

// writeEvent writes an event to the log
func (al *AuditLogger) writeEvent(event AuditEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		al.logger.Printf("Failed to marshal audit event: %v", err)
		return
	}

	al.logger.Println(string(eventJSON))
}

// flushBuffer flushes the event buffer to the log
func (al *AuditLogger) flushBuffer() {
	for _, event := range al.buffer {
		al.writeEvent(event)
	}
	al.buffer = al.buffer[:0]
}

// flushRoutine periodically flushes the buffer
func (al *AuditLogger) flushRoutine() {
	ticker := time.NewTicker(al.config.FlushInterval)
	defer ticker.Stop()

	for range ticker.C {
		al.mu.Lock()
		if len(al.buffer) > 0 {
			al.flushBuffer()
		}
		al.mu.Unlock()
	}
}

// Flush manually flushes the buffer
func (al *AuditLogger) Flush() {
	al.mu.Lock()
	defer al.mu.Unlock()

	if len(al.buffer) > 0 {
		al.flushBuffer()
	}
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	al.Flush()

	if al.logFile != nil {
		return al.logFile.Close()
	}

	return nil
}

// RotateLog rotates the log file if rotation is enabled
func (al *AuditLogger) RotateLog() error {
	if !al.config.RotateFiles {
		return nil
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// Check file size
	info, err := al.logFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get log file info: %w", err)
	}

	if info.Size() < al.config.MaxFileSize {
		return nil
	}

	// Close current file
	if err := al.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %w", err)
	}

	// Rotate files
	baseName := al.config.LogFile
	for i := al.config.MaxFiles - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", baseName, i)
		newName := fmt.Sprintf("%s.%d", baseName, i+1)

		if _, err := os.Stat(oldName); err == nil {
			if renameErr := os.Rename(oldName, newName); renameErr != nil {
				// Continue rotation even if one file fails
			}
		}
	}

	// Move current log to .1
	rotatedName := fmt.Sprintf("%s.1", baseName)
	if err := os.Rename(baseName, rotatedName); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Create new log file
	newFile, err := os.OpenFile(baseName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	al.logFile = newFile
	secureWriter := NewSecureLogWriter(newFile, al.filter)
	al.logger = log.New(secureWriter, "", 0)

	return nil
}

// GetLogFile returns the current log file path
func (al *AuditLogger) GetLogFile() string {
	return al.config.LogFile
}

// SetUserContext sets user context for subsequent log entries
func (al *AuditLogger) SetUserContext(userID, sessionID, clientIP string) *ContextualAuditLogger {
	return &ContextualAuditLogger{
		logger:    al,
		userID:    userID,
		sessionID: sessionID,
		clientIP:  clientIP,
	}
}

// ContextualAuditLogger provides audit logging with pre-set context
type ContextualAuditLogger struct {
	logger    *AuditLogger
	userID    string
	sessionID string
	clientIP  string
}

// LogEvent logs an event with context
func (cal *ContextualAuditLogger) LogEvent(eventType, action, resource, status string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		UserID:    cal.userID,
		Action:    action,
		Resource:  resource,
		Status:    status,
		Details:   details,
		ClientIP:  cal.clientIP,
		SessionID: cal.sessionID,
	}

	cal.logger.logEvent(event)
}

// AuditEventFilter provides filtering for audit events
type AuditEventFilter struct {
	EventTypes []string
	Actions    []string
	Resources  []string
	TimeRange  TimeRange
}

// TimeRange represents a time range for filtering
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Matches checks if an event matches the filter criteria
func (filter *AuditEventFilter) Matches(event AuditEvent) bool {
	if len(filter.EventTypes) > 0 {
		found := false
		for _, eventType := range filter.EventTypes {
			if event.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.Actions) > 0 {
		found := false
		for _, action := range filter.Actions {
			if event.Action == action {
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
			if event.Resource == resource {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if !filter.TimeRange.Start.IsZero() && event.Timestamp.Before(filter.TimeRange.Start) {
		return false
	}

	if !filter.TimeRange.End.IsZero() && event.Timestamp.After(filter.TimeRange.End) {
		return false
	}

	return true
}
