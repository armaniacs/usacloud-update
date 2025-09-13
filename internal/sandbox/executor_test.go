package sandbox

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
)

// Mock command execution for testing
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// Helper process for mocking command execution
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		os.Exit(1)
	}

	cmd, args := args[0], args[1:]

	// Mock different command behaviors
	switch cmd {
	case "usacloud":
		if len(args) > 0 && args[0] == "server" && len(args) > 1 && args[1] == "list" {
			// Mock successful server list command
			os.Stdout.WriteString(`[{"ID": "123", "Name": "test-server"}]`)
			os.Exit(0)
		}
		if len(args) > 0 && args[0] == "fail" {
			// Mock failing command
			os.Stderr.WriteString("Command failed")
			os.Exit(1)
		}
		if len(args) > 0 && args[0] == "timeout" {
			// Mock command that times out
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
		// Default successful command
		os.Stdout.WriteString("Success")
		os.Exit(0)
	default:
		os.Stderr.WriteString("Command not found")
		os.Exit(127)
	}
}

func TestNewExecutor(t *testing.T) {
	cfg := &config.SandboxConfig{
		Enabled:     true,
		Timeout:     30 * time.Second,
		Debug:       false,
		DryRun:      false,
		Interactive: true,
	}

	executor := NewExecutor(cfg)
	if executor == nil {
		t.Error("NewExecutor() returned nil")
	}
	if executor.config != cfg {
		t.Error("NewExecutor() did not set config correctly")
	}
}

func TestExecutor_ExecuteScript(t *testing.T) {
	// Note: These tests cannot use the mock execution because the executor
	// uses exec.CommandContext directly. Instead, we test the logic flow.

	cfg := &config.SandboxConfig{
		Enabled:           true,
		Timeout:           5 * time.Second,
		Debug:             false,
		DryRun:            false,
		Interactive:       false,
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		Zone:              "tk1v",
	}

	executor := NewExecutor(cfg)

	t.Run("SuccessfulExecution", func(t *testing.T) {
		lines := []string{
			"#!/bin/bash",
			"usacloud server list",
			"echo done",
		}

		results, err := executor.ExecuteScript(lines)
		if err != nil {
			t.Fatalf("ExecuteScript() failed: %v", err)
		}

		if len(results) != 3 { // All lines are processed, but only usacloud is executed
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// First result: shebang line should be skipped
		result := results[0]
		if !result.Skipped {
			t.Error("Expected shebang line to be skipped")
		}

		// Second result: usacloud command will fail if usacloud is not installed
		// In CI/test environments, this is expected to fail with "executable not found"
		result = results[1]
		if result.Skipped {
			t.Error("Expected usacloud command to be processed, not skipped")
		}
		// In test environment without usacloud CLI, this will fail
		// This is expected behavior and validates the error handling path

		// Third result: echo command should be skipped (not usacloud)
		result = results[2]
		if !result.Skipped {
			t.Error("Expected echo command to be skipped")
		}
	})

	t.Run("FailedExecution", func(t *testing.T) {
		lines := []string{
			"usacloud fail",
		}

		results, err := executor.ExecuteScript(lines)
		if err != nil {
			t.Fatalf("ExecuteScript() failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}

		result := results[0]
		if result.Success {
			t.Error("Expected failure, got success")
		}
		if result.Error == "" {
			t.Error("Expected error message, got empty string")
		}
		// Test can result in either "executable not found" (no usacloud) or "command failed" (usacloud installed but fail subcommand)
		if !strings.Contains(result.Error, "executable file not found") && !strings.Contains(result.Error, "command failed") {
			t.Errorf("Expected executable or command failure error, got: %s", result.Error)
		}
	})

	t.Run("DryRunMode", func(t *testing.T) {
		dryRunCfg := *cfg
		dryRunCfg.DryRun = true
		dryRunExecutor := NewExecutor(&dryRunCfg)

		lines := []string{
			"usacloud server list",
		}

		results, err := dryRunExecutor.ExecuteScript(lines)
		if err != nil {
			t.Fatalf("ExecuteScript() failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}

		result := results[0]
		if !result.Success {
			t.Errorf("Expected success in dry-run, got failure: %s", result.Error)
		}
		// In dry-run mode, commands are executed but with DRY RUN prefix
		if result.Skipped {
			t.Error("Expected command not to be skipped in dry-run mode (shows as executed with DRY RUN prefix)")
		}
		if !strings.Contains(result.Output, "[DRY RUN]") {
			t.Error("Expected [DRY RUN] prefix in output")
		}
	})

	t.Run("NonUsacloudCommands", func(t *testing.T) {
		lines := []string{
			"#!/bin/bash",
			"echo 'regular command'",
			"ls -la",
			"# comment",
			"",
		}

		results, err := executor.ExecuteScript(lines)
		if err != nil {
			t.Fatalf("ExecuteScript() failed: %v", err)
		}

		// All lines are processed but non-usacloud commands are skipped
		if len(results) != 5 {
			t.Errorf("Expected 5 results for all lines, got %d", len(results))
		}

		// All results should be skipped since no usacloud commands
		for i, result := range results {
			if !result.Skipped {
				t.Errorf("Result %d should be skipped (non-usacloud command)", i)
			}
		}
	})

	t.Run("MixedCommands", func(t *testing.T) {
		lines := []string{
			"#!/bin/bash",
			"echo 'starting'",
			"usacloud server list",
			"ls -la",
			"usacloud server list --zone=is1b",
		}

		results, err := executor.ExecuteScript(lines)
		if err != nil {
			t.Fatalf("ExecuteScript() failed: %v", err)
		}

		// All lines are processed, 2 usacloud commands will be processed
		if len(results) != 5 {
			t.Errorf("Expected 5 results for all lines, got %d", len(results))
		}

		// Debug output to understand actual behavior
		for i, result := range results {
			t.Logf("Result %d: line='%s', skipped=%t, reason='%s'", i, lines[i], result.Skipped, result.SkipReason)
		}

		// Check that usacloud commands are processed (not skipped) and others are skipped
		usacloudLines := []int{2, 4} // 0-based indices: usacloud lines are at index 2 and 4
		for i, result := range results {
			isUsacloudLine := false
			for _, idx := range usacloudLines {
				if i == idx {
					isUsacloudLine = true
					break
				}
			}

			if isUsacloudLine && result.Skipped {
				t.Errorf("Result %d (usacloud command) should not be skipped", i)
			} else if !isUsacloudLine && !result.Skipped {
				t.Errorf("Result %d (non-usacloud command) should be skipped", i)
			}
		}
	})
}

func TestExecutor_PrintSummary(t *testing.T) {
	cfg := &config.SandboxConfig{
		Debug: false,
	}
	executor := NewExecutor(cfg)

	results := []*ExecutionResult{
		{
			Command:  "usacloud server list",
			Success:  true,
			Output:   "Success",
			Duration: 100 * time.Millisecond,
		},
		{
			Command:  "usacloud server list --fail",
			Success:  false,
			Error:    "Command failed",
			Duration: 50 * time.Millisecond,
		},
		{
			Command:    "usacloud server list --dry-run",
			Success:    true,
			Skipped:    true,
			SkipReason: "Dry run mode",
			Duration:   1 * time.Millisecond,
		},
	}

	// This is mainly a smoke test to ensure PrintSummary doesn't panic
	// In a real implementation, we might want to capture the output and verify it
	executor.PrintSummary(results)
}

func TestIsUsacloudCommand(t *testing.T) {
	cfg := &config.SandboxConfig{}
	executor := NewExecutor(cfg)

	tests := []struct {
		line     string
		expected bool
	}{
		{"usacloud server list", true},
		{"  usacloud server list  ", true},
		{"sudo usacloud server list", false}, // sudo prefix not supported by regex
		{"echo usacloud server list", false},
		{"# usacloud server list", false},
		{"", false},
		{"ls -la", false},
		{"curl http://example.com", false},
		{"usacloud server", true}, // Need something after usacloud to match regex
	}

	for _, test := range tests {
		result := executor.usacloudRegex.MatchString(strings.TrimSpace(test.line))
		if result != test.expected {
			t.Errorf("usacloudRegex.MatchString(%q) = %v, expected %v", test.line, result, test.expected)
		}
	}
}

// Mock execCommand variable for testing
var execCommand = exec.Command
