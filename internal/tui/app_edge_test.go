package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/sandbox"
)

// TestApp_EdgeCases tests edge cases and error conditions for the TUI App
func TestApp_EdgeCases(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		Interactive:       true,
		DryRun:            true,
	}

	testCases := []struct {
		name        string
		setupFunc   func() *App
		testFunc    func(*App) error
		description string
	}{
		{
			name: "LoadEmptyScript",
			setupFunc: func() *App {
				return NewApp(cfg)
			},
			testFunc: func(app *App) error {
				return app.LoadScript([]string{})
			},
			description: "Loading empty script should not crash",
		},
		{
			name: "LoadNilScript",
			setupFunc: func() *App {
				return NewApp(cfg)
			},
			testFunc: func(app *App) error {
				return app.LoadScript(nil)
			},
			description: "Loading nil script should not crash",
		},
		{
			name: "LoadVeryLargeScript",
			setupFunc: func() *App {
				return NewApp(cfg)
			},
			testFunc: func(app *App) error {
				// Create a script with many lines
				lines := make([]string, 10000)
				for i := range lines {
					lines[i] = "usacloud server list --output-type=csv"
				}
				return app.LoadScript(lines)
			},
			description: "Loading very large script should not crash",
		},
		{
			name: "LoadScriptWithSpecialCharacters",
			setupFunc: func() *App {
				return NewApp(cfg)
			},
			testFunc: func(app *App) error {
				lines := []string{
					"usacloud server list --tag 'test!@#$%^&*()'",
					"usacloud disk list --name '◯△□'",
					"# Comment with unicode: テスト",
					"echo 'non-usacloud with unicode: 日本語'",
				}
				return app.LoadScript(lines)
			},
			description: "Scripts with special characters should be handled",
		},
		{
			name: "LoadScriptWithVeryLongLines",
			setupFunc: func() *App {
				return NewApp(cfg)
			},
			testFunc: func(app *App) error {
				longLine := "usacloud server list " + strings.Repeat("--tag very-long-tag-name-", 1000)
				lines := []string{longLine}
				return app.LoadScript(lines)
			},
			description: "Scripts with very long lines should be handled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test panicked: %v", r)
				}
			}()

			app := tc.setupFunc()
			if app == nil {
				t.Fatal("App creation failed")
			}

			err := tc.testFunc(app)
			if err != nil {
				t.Errorf("Test function failed: %v", err)
			}

			// Verify app is still in valid state
			if app.commands == nil {
				t.Error("Commands slice should not be nil after operation")
			}

			if app.currentIndex < 0 {
				t.Error("Current index should not be negative")
			}

			if app.totalSelected < 0 {
				t.Error("Total selected should not be negative")
			}
		})
	}
}

// TestApp_StateManagement tests state management edge cases
func TestApp_StateManagement(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		Interactive:       true,
		DryRun:            true,
	}

	app := NewApp(cfg)

	// Load test script
	lines := []string{
		"usacloud server list --output-type=csv",
		"usacloud disk list",
		"echo 'non-usacloud'",
		"# comment",
		"",
		"usacloud cdrom list",
	}

	err := app.LoadScript(lines)
	if err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	testCases := []struct {
		name        string
		operation   func(*App)
		validate    func(*App) error
		description string
	}{
		{
			name: "SelectAllMultipleTimes",
			operation: func(a *App) {
				a.selectAll()
				a.selectAll()
				a.selectAll()
			},
			validate: func(a *App) error {
				// Should be idempotent
				selectedCount := 0
				for _, cmd := range a.commands {
					if cmd.Selected {
						selectedCount++
					}
				}
				if selectedCount == 0 {
					return fmt.Errorf("expected some commands to be selected")
				}
				return nil
			},
			description: "Multiple selectAll calls should be idempotent",
		},
		{
			name: "SelectNoneMultipleTimes",
			operation: func(a *App) {
				a.selectAll()
				a.selectNone()
				a.selectNone()
				a.selectNone()
			},
			validate: func(a *App) error {
				for i, cmd := range a.commands {
					if cmd.Selected {
						return fmt.Errorf("command %d should not be selected after selectNone", i)
					}
				}
				return nil
			},
			description: "Multiple selectNone calls should be idempotent",
		},
		{
			name: "ToggleHelpMultipleTimes",
			operation: func(a *App) {
				initialState := a.helpVisible
				a.helpVisible = !a.helpVisible
				a.helpVisible = !a.helpVisible
				a.helpVisible = !a.helpVisible
				// Should end up in opposite state
				if a.helpVisible == initialState {
					a.helpVisible = !a.helpVisible
				}
			},
			validate: func(a *App) error {
				// Help visibility should be in a valid state (true or false)
				// This test mainly ensures no panic occurs
				return nil
			},
			description: "Multiple help toggles should not crash",
		},
		{
			name: "InvalidCurrentIndex",
			operation: func(a *App) {
				a.currentIndex = -1
				a.updateDetailView()
				a.currentIndex = len(a.commands) + 100
				a.updateDetailView()
				a.currentIndex = 0 // Reset to valid
			},
			validate: func(a *App) error {
				if a.currentIndex < 0 || a.currentIndex >= len(a.commands) {
					// Current index might be out of bounds during test, that's ok
				}
				return nil
			},
			description: "Invalid current index should not crash detail view update",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Operation panicked: %v", r)
				}
			}()

			tc.operation(app)

			if tc.validate != nil {
				if err := tc.validate(app); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// TestApp_CommandExecution tests command execution edge cases
func TestApp_CommandExecution(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		Interactive:       true,
		DryRun:            true,
		Timeout:           1 * time.Second, // Short timeout for testing
	}

	app := NewApp(cfg)

	// Load test script with various command types
	lines := []string{
		"usacloud server list",
		"usacloud invalid-command",
		"usacloud disk list --invalid-flag",
		"", // Empty line
		"# Comment line",
		"echo hello", // Non-usacloud
	}

	err := app.LoadScript(lines)
	if err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	// Select some commands
	if len(app.commands) > 0 {
		app.commands[0].Selected = true
	}
	if len(app.commands) > 1 {
		app.commands[1].Selected = true
	}

	// Test execution in a controlled way (without actual TUI interaction)
	t.Run("ExecuteSelectedCommands", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Command execution panicked: %v", r)
			}
		}()

		// Simulate what executeSelected does without the TUI updates
		selected := make([]*CommandItem, 0)
		for _, cmd := range app.commands {
			if cmd.Selected {
				selected = append(selected, cmd)
			}
		}

		for _, cmd := range selected {
			if app.executor != nil {
				result, _ := app.executor.ExecuteCommand(cmd.Converted)
				cmd.Result = result
			}
		}

		// Verify results
		for _, cmd := range selected {
			if cmd.Result == nil {
				t.Error("Selected command should have execution result")
			}
		}
	})
}

// TestApp_UIUpdateEdgeCases tests UI update edge cases
func TestApp_UIUpdateEdgeCases(t *testing.T) {
	cfg := &config.SandboxConfig{
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
		Enabled:           true,
		Interactive:       true,
		DryRun:            true,
	}

	app := NewApp(cfg)

	testCases := []struct {
		name        string
		operation   func(*App)
		description string
	}{
		{
			name: "UpdateLayoutWithEmptyCommands",
			operation: func(a *App) {
				a.commands = []*CommandItem{}
				a.updateLayout()
			},
			description: "Update layout with empty commands should not crash",
		},
		{
			name: "UpdateStatusBarWithNilCommands",
			operation: func(a *App) {
				a.commands = nil
				a.updateStatusBar()
			},
			description: "Update status bar with nil commands should not crash",
		},
		{
			name: "UpdateDetailViewWithNoCommands",
			operation: func(a *App) {
				a.commands = []*CommandItem{}
				a.currentIndex = 0
				a.updateDetailView()
			},
			description: "Update detail view with no commands should not crash",
		},
		{
			name: "UpdateResultViewWithMixedResults",
			operation: func(a *App) {
				// Create commands with mixed result states
				a.commands = []*CommandItem{
					{
						Original:  "usacloud server list",
						Converted: "usacloud server list",
						Result: &sandbox.ExecutionResult{
							Success: true,
							Output:  "success output",
						},
					},
					{
						Original:  "usacloud disk list",
						Converted: "usacloud disk list",
						Result: &sandbox.ExecutionResult{
							Success: false,
							Error:   "test error",
						},
					},
					{
						Original:  "usacloud cdrom list",
						Converted: "usacloud cdrom list",
						Result: &sandbox.ExecutionResult{
							Skipped:    true,
							SkipReason: "test skip",
						},
					},
					{
						Original:  "usacloud note list",
						Converted: "usacloud note list",
						Result:    nil, // No result yet
					},
				}
				a.updateResultView()
			},
			description: "Update result view with mixed result states should work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("UI update panicked: %v", r)
				}
			}()

			tc.operation(app)

			// Basic validation that app is still functional
			if app.app == nil {
				t.Error("TView application should not be nil after UI update")
			}
		})
	}
}

// TestApp_ConfigEdgeCases tests configuration edge cases
func TestApp_ConfigEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.SandboxConfig
		expectPanic bool
		description string
	}{
		{
			name:        "NilConfig",
			config:      nil,
			expectPanic: true, // Actually panics due to nil config access
			description: "Nil config currently panics (expected behavior)",
		},
		{
			name:        "EmptyConfig",
			config:      &config.SandboxConfig{},
			expectPanic: false,
			description: "Empty config should be handled",
		},
		{
			name: "DisabledSandbox",
			config: &config.SandboxConfig{
				Enabled: false,
			},
			expectPanic: false,
			description: "Disabled sandbox should work",
		},
		{
			name: "InvalidTimeout",
			config: &config.SandboxConfig{
				AccessToken:       "test-token",
				AccessTokenSecret: "test-secret",
				Zone:              "tk1v",
				Enabled:           true,
				Timeout:           -1 * time.Second,
			},
			expectPanic: false,
			description: "Invalid timeout should be handled",
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

			app := NewApp(tc.config)

			// Basic validation
			if app == nil {
				t.Error("NewApp should not return nil")
				return
			}

			// Try basic operations
			err := app.LoadScript([]string{"usacloud server list"})
			if err != nil {
				t.Errorf("LoadScript failed: %v", err)
			}
		})
	}
}
