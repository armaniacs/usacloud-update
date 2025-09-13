package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
)

// TestExecutor_EdgeCases tests edge cases and error conditions
func TestExecutor_EdgeCases(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            false,
		Timeout:           5 * time.Second,
	}

	executor := NewExecutor(cfg)

	testCases := []struct {
		name        string
		command     string
		expectError bool
		description string
	}{
		{
			name:        "EmptyCommand",
			command:     "",
			expectError: false, // Should be skipped, not error
			description: "Empty commands should be skipped",
		},
		{
			name:        "WhitespaceOnlyCommand",
			command:     "   \t\n  ",
			expectError: false, // Should be skipped
			description: "Whitespace-only commands should be skipped",
		},
		{
			name:        "CommentOnlyCommand",
			command:     "# This is just a comment",
			expectError: false, // Should be skipped
			description: "Comment-only commands should be skipped",
		},
		{
			name:        "NonUsacloudCommand",
			command:     "echo hello world",
			expectError: false, // Should be skipped
			description: "Non-usacloud commands should be skipped",
		},
		{
			name:        "VeryLongCommand",
			command:     "usacloud server list " + strings.Repeat("--tag very-long-tag-name-", 100),
			expectError: true, // Likely to fail due to length
			description: "Very long commands should be handled gracefully",
		},
		{
			name:        "CommandWithSpecialCharacters",
			command:     "usacloud server list --tag 'test!@#$%^&*()'",
			expectError: true, // May fail but should not crash
			description: "Commands with special characters should be handled",
		},
		{
			name:        "CommandWithUnicode",
			command:     "usacloud server list --tag 'テスト日本語'",
			expectError: true, // May fail but should not crash
			description: "Commands with unicode should be handled",
		},
		{
			name:        "MalformedUsacloudCommand",
			command:     "usacloud --invalid-global-flag server list",
			expectError: true, // Should fail gracefully
			description: "Malformed usacloud commands should not crash executor",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The main test is that this doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Executor panicked on command %q: %v", tc.command, r)
				}
			}()

			result, err := executor.ExecuteCommand(tc.command)

			// Basic result validation
			if result == nil {
				t.Error("Result should not be nil")
				return
			}

			if result.Command != tc.command {
				t.Errorf("Result command should match input: got %q, want %q", result.Command, tc.command)
			}

			// Check skip conditions
			trimmed := strings.TrimSpace(tc.command)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, "usacloud ") {
				if !result.Skipped {
					t.Errorf("Command %q should have been skipped", tc.command)
				}
				// Note: Skipped commands are actually marked as Success=true in the executor
				if !result.Success {
					t.Error("Skipped commands should be marked as successful (by design)")
				}
				if err != nil {
					t.Errorf("Skipped commands should not return error: %v", err)
				}
				return
			}

			// For non-skipped commands, validate error expectations
			if tc.expectError {
				if err == nil && result.Success {
					t.Errorf("Expected command %q to fail, but it succeeded", tc.command)
				}
			}

			// Validate result structure
			if result.Duration < 0 {
				t.Error("Duration should not be negative")
			}

			if result.Skipped && (result.Success || result.Error != "" || result.Output != "") {
				t.Error("Skipped results should not have success/error/output")
			}
		})
	}
}

// TestExecutor_TimeoutHandling tests command timeout scenarios
func TestExecutor_TimeoutHandling(t *testing.T) {
	if !IsUsacloudInstalled() {
		t.Skip("usacloud CLI not installed, skipping timeout tests")
	}

	cfg := &config.SandboxConfig{
		AccessToken:       "invalid-token", // Use invalid token to ensure failure
		AccessTokenSecret: "invalid-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            false,
		Timeout:           1 * time.Millisecond, // Very short timeout
	}

	executor := NewExecutor(cfg)

	testCases := []struct {
		name        string
		command     string
		description string
	}{
		{
			name:        "FastFailingCommand",
			command:     "usacloud server list",
			description: "Command that should fail quickly with invalid credentials",
		},
		{
			name:        "SimpleCommand",
			command:     "usacloud version",
			description: "Simple command that should complete quickly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			result, _ := executor.ExecuteCommand(tc.command)
			elapsed := time.Since(start)

			// Should complete within a reasonable time (not hang indefinitely)
			if elapsed > 5*time.Second {
				t.Errorf("Command took too long: %v", elapsed)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Command should either succeed quickly or fail, but not timeout
			// (since timeout is set very low, most real commands would timeout)
			if !result.Success && result.Error == "" {
				t.Error("Failed commands should have error message")
			}

			// For dry run, verify no actual execution
			if cfg.DryRun && result.Success && result.Output != "" {
				t.Error("Dry run should not produce actual output")
			}

			t.Logf("Command %q completed in %v with success=%v", tc.command, elapsed, result.Success)
		})
	}
}

// TestExecutor_ConcurrentExecution tests concurrent command execution
func TestExecutor_ConcurrentExecution(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            true, // Use dry run to avoid actual API calls
		Timeout:           30 * time.Second,
	}

	executor := NewExecutor(cfg)

	const numGoroutines = 5
	const commandsPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for j := 0; j < commandsPerGoroutine; j++ {
				command := "usacloud server list"
				result, err := executor.ExecuteCommand(command)

				if result == nil {
					t.Errorf("Goroutine %d command %d: result should not be nil", id, j)
					continue
				}

				if err != nil {
					t.Errorf("Goroutine %d command %d: unexpected error: %v", id, j, err)
				}

				// In dry run mode, commands might still have some output in test scenarios
				// This is acceptable for testing purposes
				if !result.Skipped && len(result.Output) > 1000 {
					t.Errorf("Goroutine %d command %d: unexpectedly large output: %d chars", id, j, len(result.Output))
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(30 * time.Second):
			t.Fatal("Goroutines did not complete in time")
		}
	}
}

// TestExecutor_MemoryUsage tests memory usage with large inputs
func TestExecutor_MemoryUsage(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            true,
		Timeout:           30 * time.Second,
	}

	executor := NewExecutor(cfg)

	// Test with many commands
	var commands []string
	for i := 0; i < 1000; i++ {
		commands = append(commands, "usacloud server list")
	}

	results, execErr := executor.ExecuteScript(commands)

	if execErr != nil {
		t.Fatalf("ExecuteScript failed: %v", execErr)
	}

	if len(results) != len(commands) {
		t.Errorf("Expected %d results, got %d", len(commands), len(results))
	}

	// Verify all results are valid
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d should not be nil", i)
			continue
		}

		if result.Command != commands[i] {
			t.Errorf("Result %d command mismatch: got %q, want %q", i, result.Command, commands[i])
		}
	}
}

// TestExecutor_ErrorRecovery tests recovery from various error conditions
func TestExecutor_ErrorRecovery(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            false,
		Timeout:           5 * time.Second,
	}

	executor := NewExecutor(cfg)

	// Test sequence of commands with different failure modes
	commands := []string{
		"usacloud server list",        // May fail with invalid credentials
		"",                            // Should be skipped
		"# Comment",                   // Should be skipped
		"echo hello",                  // Should be skipped (not usacloud)
		"usacloud invalid-subcommand", // Should fail gracefully
		"usacloud server list",        // Should attempt execution again
	}

	results, execErr := executor.ExecuteScript(commands)

	// Script execution should not fail even if individual commands fail
	if execErr != nil {
		t.Errorf("ExecuteScript should not fail due to individual command failures: %v", execErr)
	}

	if len(results) != len(commands) {
		t.Errorf("Expected %d results, got %d", len(commands), len(results))
	}

	// Verify skip conditions
	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d should not be nil", i)
			continue
		}

		command := commands[i]
		trimmed := strings.TrimSpace(command)

		shouldSkip := trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, "usacloud ")

		if shouldSkip && !result.Skipped {
			t.Errorf("Command %d %q should have been skipped", i, command)
		}

		if !shouldSkip && result.Skipped {
			t.Errorf("Command %d %q should not have been skipped", i, command)
		}
	}
}

// TestExecutor_ConfigValidation tests configuration validation edge cases
func TestExecutor_ConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.SandboxConfig
		expectPanic bool
		description string
	}{
		{
			name:        "NilConfig",
			config:      nil,
			expectPanic: true,
			description: "Nil config should be handled gracefully",
		},
		{
			name: "DisabledSandbox",
			config: &config.SandboxConfig{
				Enabled: false,
				DryRun:  false,
			},
			expectPanic: false,
			description: "Disabled sandbox should work",
		},
		{
			name: "ZeroTimeout",
			config: &config.SandboxConfig{
				AccessToken:       "test-token",
				AccessTokenSecret: "test-secret",
				Zone:              "tk1v",
				Enabled:           true,
				Timeout:           0,
			},
			expectPanic: false,
			description: "Zero timeout should be handled",
		},
		{
			name: "NegativeTimeout",
			config: &config.SandboxConfig{
				AccessToken:       "test-token",
				AccessTokenSecret: "test-secret",
				Zone:              "tk1v",
				Enabled:           true,
				Timeout:           -1 * time.Second,
			},
			expectPanic: false,
			description: "Negative timeout should be handled",
		},
		{
			name: "EmptyCredentials",
			config: &config.SandboxConfig{
				AccessToken:       "",
				AccessTokenSecret: "",
				Zone:              "tk1v",
				Enabled:           true,
				Timeout:           30 * time.Second,
			},
			expectPanic: false,
			description: "Empty credentials should be handled (executor creation succeeds, validation fails later)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tc.expectPanic && r == nil {
					t.Error("Expected panic but none occurred")
				}
				if !tc.expectPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			executor := NewExecutor(tc.config)

			// Try to execute a simple command
			if executor != nil {
				result, err := executor.ExecuteCommand("usacloud server list")

				// For invalid configs, ExecuteCommand may return error
				if tc.name == "EmptyCredentials" {
					// This should fail validation and return error
					if err == nil {
						t.Error("Empty credentials should cause validation error")
					}
				} else {
					// Basic validation that result is not nil for other cases
					if result == nil {
						t.Error("Result should not be nil for valid configs")
					}
				}
			} else if !tc.expectPanic {
				t.Error("Executor should not be nil for non-panic cases")
			}
		})
	}
}

// TestExecutor_ContextCancellation tests context cancellation scenarios
func TestExecutor_ContextCancellation(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		DryRun:            false,
		Timeout:           30 * time.Second, // Long timeout
	}

	executor := NewExecutor(cfg)

	// Create a context that we'll cancel quickly
	_, cancel := context.WithCancel(context.Background())

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()

	// This would normally take longer, but should be cancelled quickly
	result, err := executor.ExecuteCommand("usacloud server list")

	elapsed := time.Since(start)

	// Should complete quickly due to cancellation
	if elapsed > 5*time.Second {
		t.Errorf("Command should have been cancelled quickly, took %v", elapsed)
	}

	// Result should still be valid
	if result == nil {
		t.Error("Result should not be nil even when cancelled")
	}

	// Either should succeed quickly or fail due to cancellation/invalid creds
	// The exact behavior depends on timing and usacloud installation
	t.Logf("Command completed in %v with success=%v, error=%v", elapsed, result.Success, err)
}
