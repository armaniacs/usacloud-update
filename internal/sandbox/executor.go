package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/fatih/color"
)

// ExecutionResult represents the result of executing a command
type ExecutionResult struct {
	Command    string        `json:"command"`
	Success    bool          `json:"success"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Skipped    bool          `json:"skipped"`
	SkipReason string        `json:"skip_reason,omitempty"`
}

// Executor handles sandbox execution of usacloud commands
type Executor struct {
	config        *config.SandboxConfig
	usacloudRegex *regexp.Regexp
}

// NewExecutor creates a new sandbox executor
func NewExecutor(cfg *config.SandboxConfig) *Executor {
	// Regex to identify usacloud commands
	usacloudRegex := regexp.MustCompile(`^\s*usacloud\s+`)

	return &Executor{
		config:        cfg,
		usacloudRegex: usacloudRegex,
	}
}

// ExecuteScript executes all usacloud commands in the provided script lines
func (e *Executor) ExecuteScript(lines []string) ([]*ExecutionResult, error) {
	if err := e.config.Validate(); err != nil {
		return nil, fmt.Errorf("sandbox configuration validation failed: %w", err)
	}

	var results []*ExecutionResult

	for i, line := range lines {
		lineNum := i + 1

		if e.config.Debug {
			fmt.Fprintf(os.Stderr, color.CyanString("[DEBUG] Processing line %d: %s\n"), lineNum, line)
		}

		result := e.executeLine(line, lineNum)
		results = append(results, result)

		// Add small delay between commands to avoid rate limiting
		if !result.Skipped && !e.config.DryRun {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return results, nil
}

// ExecuteCommand executes a single usacloud command
func (e *Executor) ExecuteCommand(command string) (*ExecutionResult, error) {
	if err := e.config.Validate(); err != nil {
		return nil, fmt.Errorf("sandbox configuration validation failed: %w", err)
	}

	return e.executeLine(command, 1), nil
}

// executeLine processes and executes a single line
func (e *Executor) executeLine(line string, lineNum int) *ExecutionResult {
	start := time.Now()
	result := &ExecutionResult{
		Command: line,
		Success: false,
	}

	// Skip empty lines and comments
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		result.Skipped = true
		result.SkipReason = "Empty line or comment"
		result.Success = true
		result.Duration = time.Since(start)
		return result
	}

	// Skip non-usacloud commands
	if !e.usacloudRegex.MatchString(trimmed) {
		result.Skipped = true
		result.SkipReason = "Not a usacloud command"
		result.Success = true
		result.Duration = time.Since(start)
		return result
	}

	// Skip commented usacloud commands (those marked for manual intervention)
	if strings.Contains(trimmed, "# usacloud-update:") && strings.HasPrefix(trimmed, "# ") {
		result.Skipped = true
		result.SkipReason = "Commented usacloud command (manual intervention required)"
		result.Success = true
		result.Duration = time.Since(start)
		return result
	}

	// Extract the usacloud command
	command := e.extractUsacloudCommand(trimmed)
	if command == "" {
		result.Skipped = true
		result.SkipReason = "Failed to extract usacloud command"
		result.Duration = time.Since(start)
		return result
	}

	// Validate command for sandbox safety
	if err := e.validateCommand(command); err != nil {
		result.Error = fmt.Sprintf("Command validation failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Execute in dry-run mode
	if e.config.DryRun {
		result.Output = fmt.Sprintf("[DRY RUN] Would execute: %s", command)
		result.Success = true
		result.Duration = time.Since(start)
		return result
	}

	// Execute the command
	ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
	defer cancel()

	if e.config.Debug {
		fmt.Fprintf(os.Stderr, color.BlueString("[EXEC] %s\n"), command)
	}

	output, err := e.executeUsacloudCommand(ctx, command)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		result.Output = output
		return result
	}

	result.Success = true
	result.Output = output
	return result
}

// extractUsacloudCommand extracts the usacloud command from a line
func (e *Executor) extractUsacloudCommand(line string) string {
	// Remove leading/trailing whitespace
	line = strings.TrimSpace(line)

	// Handle shell constructs (&&, ||, ;, |)
	// For now, extract only the usacloud portion
	parts := regexp.MustCompile(`\s*(?:&&|\|\||;|\|)\s*`).Split(line, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if e.usacloudRegex.MatchString(part) {
			return part
		}
	}

	return ""
}

// validateCommand validates that a command is safe for sandbox execution
func (e *Executor) validateCommand(command string) error {
	// Ensure command starts with usacloud
	if !strings.HasPrefix(command, "usacloud ") {
		return fmt.Errorf("command must start with 'usacloud'")
	}

	// Ensure zone is set to tk1v (sandbox zone)
	if strings.Contains(command, "--zone") && !strings.Contains(command, "--zone=tk1v") && !strings.Contains(command, "--zone tk1v") {
		return fmt.Errorf("sandbox commands must use --zone=tk1v")
	}

	// Check for potentially dangerous operations
	dangerousOps := []string{
		"delete",
		"shutdown",
		"reset",
		"power-off",
		"boot",
		"reboot",
	}

	parts := strings.Fields(command)
	if len(parts) >= 2 {
		for _, op := range dangerousOps {
			if parts[1] == op {
				return fmt.Errorf("operation '%s' not allowed in sandbox mode for safety", op)
			}
		}
	}

	return nil
}

// executeUsacloudCommand executes a usacloud command with proper environment
func (e *Executor) executeUsacloudCommand(ctx context.Context, command string) (string, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Ensure zone is set to sandbox zone
	args = e.ensureSandboxZone(args)

	// Create command
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = e.config.GetUsacloudEnv()

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Check for context timeout
	if ctx.Err() == context.DeadlineExceeded {
		return outputStr, fmt.Errorf("command timed out after %v", e.config.Timeout)
	}

	if err != nil {
		// Try to provide helpful error messages for common issues
		if strings.Contains(outputStr, "authentication") || strings.Contains(outputStr, "unauthorized") {
			return outputStr, fmt.Errorf("authentication failed - check SAKURACLOUD_ACCESS_TOKEN and SAKURACLOUD_ACCESS_TOKEN_SECRET")
		}

		if strings.Contains(outputStr, "not found") && strings.Contains(command, "usacloud") {
			return outputStr, fmt.Errorf("usacloud command not found - please install usacloud CLI")
		}

		return outputStr, fmt.Errorf("command failed: %w", err)
	}

	// Add sandbox warning to output
	if outputStr != "" {
		outputStr += "\n" + color.YellowString("⚠️  Executed in Sakura Cloud Sandbox (tk1v) - resources may not function normally")
	}

	return outputStr, nil
}

// ensureSandboxZone ensures that the command uses the sandbox zone
func (e *Executor) ensureSandboxZone(args []string) []string {
	// If zone is already specified, ensure it's tk1v
	for i, arg := range args {
		if arg == "--zone" && i+1 < len(args) {
			args[i+1] = "tk1v"
			return args
		}
		if strings.HasPrefix(arg, "--zone=") {
			args[i] = "--zone=tk1v"
			return args
		}
	}

	// If no zone specified, add --zone=tk1v
	result := make([]string, 0, len(args)+2)
	if len(args) > 0 {
		result = append(result, args[0]) // usacloud
		result = append(result, "--zone=tk1v")
		if len(args) > 1 {
			result = append(result, args[1:]...)
		}
	}

	return result
}

// PrintSummary prints a summary of execution results
func (e *Executor) PrintSummary(results []*ExecutionResult) {
	total := len(results)
	executed := 0
	successful := 0
	skipped := 0
	failed := 0

	for _, result := range results {
		if result.Skipped {
			skipped++
		} else {
			executed++
			if result.Success {
				successful++
			} else {
				failed++
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\n%s\n", color.HiWhiteString("🏖️  Sandbox Execution Summary"))
	fmt.Fprintf(os.Stderr, "Total lines:     %d\n", total)
	fmt.Fprintf(os.Stderr, "Executed:        %s\n", color.BlueString("%d", executed))
	fmt.Fprintf(os.Stderr, "Successful:      %s\n", color.GreenString("%d", successful))
	fmt.Fprintf(os.Stderr, "Failed:          %s\n", color.RedString("%d", failed))
	fmt.Fprintf(os.Stderr, "Skipped:         %s\n", color.YellowString("%d", skipped))

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "\n%s\n", color.HiRedString("❌ Failed Commands:"))
		for i, result := range results {
			if !result.Success && !result.Skipped {
				fmt.Fprintf(os.Stderr, "  Line %d: %s\n", i+1, result.Command)
				fmt.Fprintf(os.Stderr, "  Error: %s\n\n", color.RedString(result.Error))
			}
		}
	}

	if e.config.Debug {
		fmt.Fprintf(os.Stderr, "\n%s\n", color.HiCyanString("🔍 Debug Information:"))
		fmt.Fprintf(os.Stderr, "Zone:           %s\n", e.config.Zone)
		fmt.Fprintf(os.Stderr, "API Endpoint:   %s\n", e.config.APIEndpoint)
		fmt.Fprintf(os.Stderr, "Dry Run:        %t\n", e.config.DryRun)
		fmt.Fprintf(os.Stderr, "Timeout:        %s\n", e.config.Timeout)
	}
}

// IsUsacloudInstalled checks if usacloud CLI is installed
func IsUsacloudInstalled() bool {
	_, err := exec.LookPath("usacloud")
	return err == nil
}
