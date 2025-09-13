// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"fmt"
	"strings"
)

// Error types for subcommand validation
const (
	ErrorTypeUnexpectedSubcommand = "unexpected_subcommand" // Standalone command with subcommand
	ErrorTypeInvalidSubcommand    = "invalid_subcommand"    // Non-existent subcommand
	ErrorTypeMissingSubcommand    = "missing_subcommand"    // Subcommand required but not specified
)

// SubcommandValidationResult represents subcommand validation result
type SubcommandValidationResult struct {
	IsValid     bool     // Whether the subcommand is valid
	MainCommand string   // Main command
	SubCommand  string   // Subcommand being validated
	ErrorType   string   // Error type
	Message     string   // Detailed message
	Suggestions []string // Suggested subcommands
	Available   []string // Available subcommands list
}

// SubcommandValidator represents a subcommand validator
type SubcommandValidator struct {
	commandSubcommands map[string][]string   // command -> subcommands mapping
	standaloneCommands map[string]bool       // commands that don't take subcommands
	mainValidator      *MainCommandValidator // main command validator
}

// NewSubcommandValidator creates a new subcommand validator
func NewSubcommandValidator(mainValidator *MainCommandValidator) *SubcommandValidator {
	validator := &SubcommandValidator{
		commandSubcommands: make(map[string][]string),
		standaloneCommands: map[string]bool{
			"version":     true,
			"update-self": true,
		},
		mainValidator: mainValidator,
	}

	// Initialize subcommand dictionaries
	validator.initializeSubcommands()

	return validator
}

// initializeSubcommands initializes the subcommand dictionaries
func (v *SubcommandValidator) initializeSubcommands() {
	// IaaS commands subcommands - use existing dictionary functions
	iaasCommands := v.mainValidator.GetAllCommands()["iaas"]
	for _, cmd := range iaasCommands {
		switch cmd {
		case "server":
			v.commandSubcommands[cmd] = GetServerSubcommands()
		case "disk":
			v.commandSubcommands[cmd] = GetDiskSubcommands()
		default:
			// Common IaaS subcommands for other commands
			v.commandSubcommands[cmd] = []string{"list", "read", "create", "update", "delete"}
		}
	}

	// Misc commands subcommands
	v.commandSubcommands["config"] = []string{"list", "show", "use", "create", "edit", "delete"}
	v.commandSubcommands["rest"] = []string{"get", "post", "put", "delete"}
	v.commandSubcommands["webaccelerator"] = []string{"list", "read", "create", "update", "delete", "purge"}

	// Root commands subcommands
	v.commandSubcommands["completion"] = []string{"bash", "zsh", "fish", "powershell"}
	// version, update-self are handled in standaloneCommands
}

// Validate validates a subcommand for a given main command
func (v *SubcommandValidator) Validate(mainCommand, subCommand string) *SubcommandValidationResult {
	result := &SubcommandValidationResult{
		MainCommand: mainCommand,
		SubCommand:  subCommand,
	}

	// First validate that the main command exists (case insensitive)
	if !v.mainValidator.IsValidCommand(mainCommand) {
		result.IsValid = false
		result.ErrorType = "invalid_main_command"
		result.Message = fmt.Sprintf("メインコマンド '%s' が無効です", mainCommand)
		return result
	}

	normalizedMain := strings.ToLower(mainCommand)
	normalizedSub := strings.ToLower(subCommand)

	// Check if it's a standalone command
	if v.standaloneCommands[normalizedMain] {
		if subCommand != "" {
			result.IsValid = false
			result.ErrorType = ErrorTypeUnexpectedSubcommand
			result.Message = fmt.Sprintf("コマンド '%s' はサブコマンドを受け付けません", mainCommand)
			return result
		}
		// Standalone command without subcommand is valid
		result.IsValid = true
		return result
	}

	// Get available subcommands for this main command
	availableSubcommands, exists := v.commandSubcommands[normalizedMain]
	if !exists {
		result.IsValid = false
		result.ErrorType = ErrorTypeInvalidSubcommand
		result.Message = fmt.Sprintf("コマンド '%s' のサブコマンド辞書が見つかりません", mainCommand)
		return result
	}

	result.Available = availableSubcommands

	// If no subcommand provided but required
	if subCommand == "" {
		result.IsValid = false
		result.ErrorType = ErrorTypeMissingSubcommand
		result.Message = fmt.Sprintf("コマンド '%s' にはサブコマンドが必要です", mainCommand)
		result.Suggestions = availableSubcommands[:min(5, len(availableSubcommands))] // Show first 5 as suggestions
		return result
	}

	// Check if subcommand exists
	for _, validSub := range availableSubcommands {
		if normalizedSub == validSub {
			result.IsValid = true
			return result
		}
	}

	// Subcommand doesn't exist
	result.IsValid = false
	result.ErrorType = ErrorTypeInvalidSubcommand
	result.Message = fmt.Sprintf("サブコマンド '%s' はコマンド '%s' では利用できません", subCommand, mainCommand)
	result.Suggestions = v.getSimilarSubcommands(normalizedMain, normalizedSub)

	return result
}

// ValidateCommandLine validates a parsed command line for subcommand
func (v *SubcommandValidator) ValidateCommandLine(cmdLine *CommandLine) *SubcommandValidationResult {
	return v.Validate(cmdLine.MainCommand, cmdLine.SubCommand)
}

// IsValidSubcommand checks if a subcommand is valid for a given main command
func (v *SubcommandValidator) IsValidSubcommand(mainCommand, subCommand string) bool {
	result := v.Validate(mainCommand, subCommand)
	return result.IsValid
}

// GetAvailableSubcommands returns available subcommands for a main command
func (v *SubcommandValidator) GetAvailableSubcommands(mainCommand string) []string {
	normalized := strings.ToLower(mainCommand)

	// Check if it's a standalone command
	if v.standaloneCommands[normalized] {
		return []string{} // No subcommands for standalone commands
	}

	subcommands, exists := v.commandSubcommands[normalized]
	if !exists {
		return []string{}
	}

	// Return a copy to prevent modification
	result := make([]string, len(subcommands))
	copy(result, subcommands)
	return result
}

// IsStandaloneCommand checks if a command is a standalone command
func (v *SubcommandValidator) IsStandaloneCommand(mainCommand string) bool {
	normalized := strings.ToLower(mainCommand)
	return v.standaloneCommands[normalized]
}

// GetCommandSubcommandCount returns the total number of subcommands across all commands
func (v *SubcommandValidator) GetCommandSubcommandCount() int {
	total := 0
	for _, subcommands := range v.commandSubcommands {
		total += len(subcommands)
	}
	return total
}

// GetCommandsWithSubcommands returns all commands that have subcommands
func (v *SubcommandValidator) GetCommandsWithSubcommands() []string {
	var commands []string
	for cmd := range v.commandSubcommands {
		commands = append(commands, cmd)
	}
	return commands
}

// getSimilarSubcommands finds similar subcommands using simple string matching
// This is a basic implementation - more sophisticated algorithm in PBI-012
func (v *SubcommandValidator) getSimilarSubcommands(mainCommand, subCommand string) []string {
	availableSubcommands, exists := v.commandSubcommands[mainCommand]
	if !exists {
		return []string{}
	}

	var suggestions []string
	maxSuggestions := 3

	// Simple similarity check - subcommands that contain the input or vice versa
	for _, validSub := range availableSubcommands {
		if len(suggestions) >= maxSuggestions {
			break
		}

		// Check if subcommand contains input or input contains subcommand
		if strings.Contains(validSub, subCommand) || strings.Contains(subCommand, validSub) {
			suggestions = append(suggestions, validSub)
		}
	}

	// If no substring matches, look for subcommands with similar prefixes
	if len(suggestions) == 0 && len(subCommand) > 1 {
		prefix := subCommand[:min(2, len(subCommand))]
		for _, validSub := range availableSubcommands {
			if len(suggestions) >= maxSuggestions {
				break
			}
			if strings.HasPrefix(validSub, prefix) {
				suggestions = append(suggestions, validSub)
			}
		}
	}

	// If still no matches, return first few available subcommands as fallback
	if len(suggestions) == 0 {
		maxFallback := min(3, len(availableSubcommands))
		suggestions = make([]string, maxFallback)
		copy(suggestions, availableSubcommands[:maxFallback])
	}

	return suggestions
}

// GetAllSubcommandsByCommand returns a map of command to subcommands
func (v *SubcommandValidator) GetAllSubcommandsByCommand() map[string][]string {
	result := make(map[string][]string)
	for cmd, subcommands := range v.commandSubcommands {
		result[cmd] = make([]string, len(subcommands))
		copy(result[cmd], subcommands)
	}
	return result
}

// HasSubcommands checks if a main command has any subcommands
func (v *SubcommandValidator) HasSubcommands(mainCommand string) bool {
	normalized := strings.ToLower(mainCommand)
	if v.standaloneCommands[normalized] {
		return false
	}
	subcommands, exists := v.commandSubcommands[normalized]
	return exists && len(subcommands) > 0
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
