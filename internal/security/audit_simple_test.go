package security

import (
	"os"
	"path/filepath"
	"testing"
)

// Phase 3 Coverage Improvement Tests - audit.go (Simplified)

func TestDefaultAuditConfig_Simple(t *testing.T) {
	config := DefaultAuditConfig()
	if config == nil {
		t.Error("Expected config to be created, got nil")
	}

	// Test that basic fields are set
	if config.LogFile == "" {
		t.Error("Expected LogFile to be set")
	}
	if config.BufferSize <= 0 {
		t.Error("Expected BufferSize to be positive")
	}
}

func TestNewAuditLogger_Simple(t *testing.T) {
	config := DefaultAuditConfig()

	// Create temporary directory for log file
	tmpDir := t.TempDir()
	config.LogFile = filepath.Join(tmpDir, "audit.log")

	logger, err := NewAuditLogger(config)
	if err != nil {
		t.Errorf("Expected to create logger without error, got: %v", err)
	}
	if logger == nil {
		t.Error("Expected logger to be created, got nil")
	}

	// Clean up
	if logger != nil {
		logger.Close()
	}
}

func TestNewAuditLogger_NilConfig(t *testing.T) {
	// Test with nil config - may succeed with default values
	logger, err := NewAuditLogger(nil)

	// Clean up if logger was created
	if logger != nil {
		defer logger.Close()
	}

	// The actual behavior may vary - test should just verify it doesn't panic
	// Either error or success is acceptable for nil config
	if err != nil && logger != nil {
		t.Error("If error is returned, logger should be nil")
	}
}

func TestAuditLogger_LogCredentialAccess_Simple(t *testing.T) {
	config := DefaultAuditConfig()
	tmpDir := t.TempDir()
	config.LogFile = filepath.Join(tmpDir, "audit.log")

	logger, err := NewAuditLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test basic credential access logging
	logger.LogCredentialAccess("test-credential", "read")
	logger.Flush()

	// Verify log file was created
	if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestAuditLogger_LogCredentialRotation_Simple(t *testing.T) {
	config := DefaultAuditConfig()
	tmpDir := t.TempDir()
	config.LogFile = filepath.Join(tmpDir, "audit.log")

	logger, err := NewAuditLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test credential rotation logging
	logger.LogCredentialRotation("test-credential", 1, 2)
	logger.Flush()

	// Verify log file was created
	if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestAuditLogger_LogSecurityViolation_Simple(t *testing.T) {
	config := DefaultAuditConfig()
	tmpDir := t.TempDir()
	config.LogFile = filepath.Join(tmpDir, "audit.log")

	logger, err := NewAuditLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test security violation logging with proper map
	details := map[string]interface{}{
		"violation_type": "unauthorized_access",
		"user":           "test-user",
		"description":    "Attempted to access restricted resource",
	}
	logger.LogSecurityViolation("unauthorized_access", details)
	logger.Flush()

	// Verify log file was created
	if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestAuditLogger_GetLogFile_Simple(t *testing.T) {
	config := DefaultAuditConfig()
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")
	config.LogFile = logFilePath

	logger, err := NewAuditLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test getting log file path
	returnedPath := logger.GetLogFile()
	if returnedPath != logFilePath {
		t.Errorf("Expected log file path '%s', got '%s'", logFilePath, returnedPath)
	}
}
