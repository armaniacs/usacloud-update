// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"fmt"
	"strings"
)

// ValidationResult represents main command validation result
type ValidationResult struct {
	IsValid     bool     // Whether the command is valid
	Command     string   // Command being validated
	CommandType string   // Command type (iaas/misc/root)
	ErrorType   string   // Error type (if invalid)
	Message     string   // Detailed message
	Suggestions []string // Suggested commands (similar commands, etc.)
}

// MainCommandValidator represents a main command validator
type MainCommandValidator struct {
	iaasCommands map[string]bool
	miscCommands map[string]bool
	rootCommands map[string]bool
	allCommands  map[string]string // command -> type mapping
}

// Standalone commands that don't take subcommands
var standaloneCommands = map[string]bool{
	"version":     true,
	"update-self": true,
}

// NewMainCommandValidator creates a new main command validator
func NewMainCommandValidator() *MainCommandValidator {
	validator := &MainCommandValidator{
		iaasCommands: make(map[string]bool),
		miscCommands: make(map[string]bool),
		rootCommands: make(map[string]bool),
		allCommands:  make(map[string]string),
	}

	// Initialize command dictionaries
	validator.initializeCommands()

	return validator
}

// initializeCommands initializes the command dictionaries
func (v *MainCommandValidator) initializeCommands() {
	// IaaS commands (44 commands)
	iaasCommandList := []string{
		"server", "disk", "database", "loadbalancer", "dns", "gslb", "proxylb",
		"autobackup", "archive", "cdrom", "bridge", "packetfilter", "internet",
		"ipaddress", "ipv6addr", "ipv6net", "subnet", "swytch", "localrouter",
		"vpcrouter", "mobilegateway", "sim", "nfs", "license", "licenseinfo",
		"sshkey", "note", "icon", "privatehost", "privatehostplan", "zone",
		"region", "bill", "coupon", "authstatus", "self", "serviceclass",
		"enhanceddb", "containerregistry", "certificateauthority", "esme",
		"simplemonitor", "autoscale", "category", "disk-plan", "internet-plan", "server-plan",
	}

	// Misc commands (3 commands)
	miscCommandList := []string{
		"config", "rest", "webaccelerator",
	}

	// Root commands (3 commands)
	rootCommandList := []string{
		"completion", "version", "update-self",
	}

	// Populate IaaS commands
	for _, cmd := range iaasCommandList {
		v.iaasCommands[cmd] = true
		v.allCommands[cmd] = "iaas"
	}

	// Populate misc commands
	for _, cmd := range miscCommandList {
		v.miscCommands[cmd] = true
		v.allCommands[cmd] = "misc"
	}

	// Populate root commands
	for _, cmd := range rootCommandList {
		v.rootCommands[cmd] = true
		v.allCommands[cmd] = "root"
	}
}

// Validate validates a main command
func (v *MainCommandValidator) Validate(command string) *ValidationResult {
	if command == "" {
		return &ValidationResult{
			IsValid:   false,
			Command:   command,
			ErrorType: "empty_command",
			Message:   "メインコマンドが指定されていません",
		}
	}

	// Normalize to lowercase for checking
	normalized := strings.ToLower(command)
	originalCommand := command

	// Check if command exists with normalized form
	commandType, exists := v.allCommands[normalized]
	if !exists {
		// Check for deprecated commands
		if IsDeprecatedCommand(normalized) {
			replacement, _ := GetReplacementCommand(normalized)
			if replacement != "" {
				return &ValidationResult{
					IsValid:     false,
					Command:     originalCommand,
					CommandType: "deprecated",
					ErrorType:   "deprecated_command",
					Message:     fmt.Sprintf("コマンド '%s' は廃止されました。'%s' を使用してください", normalized, replacement),
					Suggestions: []string{replacement},
				}
			} else {
				message, _ := GetDeprecatedCommandMessage(normalized)
				return &ValidationResult{
					IsValid:     false,
					Command:     originalCommand,
					CommandType: "deprecated",
					ErrorType:   "discontinued_command",
					Message:     message,
				}
			}
		}

		// Command doesn't exist
		suggestions := v.getSimilarCommands(normalized, 2)
		return &ValidationResult{
			IsValid:     false,
			Command:     originalCommand,
			ErrorType:   "unknown_command",
			Message:     fmt.Sprintf("コマンド '%s' は存在しません", originalCommand),
			Suggestions: suggestions,
		}
	}

	// If case differs, provide a case sensitivity message but still valid
	if originalCommand != normalized {
		return &ValidationResult{
			IsValid:     true,
			Command:     originalCommand,
			CommandType: commandType,
			Message:     fmt.Sprintf("コマンド '%s' は有効ですが、小文字 '%s' を推奨します", originalCommand, normalized),
			Suggestions: []string{normalized},
		}
	}

	return &ValidationResult{
		IsValid:     true,
		Command:     originalCommand,
		CommandType: commandType,
		Message:     "",
	}
}

// ValidateCommandLine validates a parsed command line for main command
func (v *MainCommandValidator) ValidateCommandLine(cmdLine *CommandLine) *ValidationResult {
	if cmdLine.MainCommand == "" {
		return &ValidationResult{
			IsValid:   false,
			Command:   "",
			ErrorType: "empty_command",
			Message:   "メインコマンドが指定されていません",
		}
	}

	// First validate the main command exists
	result := v.Validate(cmdLine.MainCommand)
	if !result.IsValid {
		return result
	}

	// Check standalone commands
	if standaloneCommands[strings.ToLower(cmdLine.MainCommand)] && cmdLine.SubCommand != "" {
		return &ValidationResult{
			IsValid:     false,
			Command:     cmdLine.MainCommand,
			CommandType: result.CommandType,
			ErrorType:   "unexpected_subcommand",
			Message:     fmt.Sprintf("コマンド '%s' はサブコマンドを受け付けません", cmdLine.MainCommand),
		}
	}

	return result
}

// IsValidCommand checks if a command is valid
func (v *MainCommandValidator) IsValidCommand(command string) bool {
	// Check case-insensitively
	normalized := strings.ToLower(command)
	return v.allCommands[normalized] != ""
}

// GetCommandType returns the command type
func (v *MainCommandValidator) GetCommandType(command string) string {
	normalized := strings.ToLower(command)
	return v.allCommands[normalized]
}

// GetAllCommands returns all valid commands by type
func (v *MainCommandValidator) GetAllCommands() map[string][]string {
	result := map[string][]string{
		"iaas": make([]string, 0, len(v.iaasCommands)),
		"misc": make([]string, 0, len(v.miscCommands)),
		"root": make([]string, 0, len(v.rootCommands)),
	}

	for cmd := range v.iaasCommands {
		result["iaas"] = append(result["iaas"], cmd)
	}
	for cmd := range v.miscCommands {
		result["misc"] = append(result["misc"], cmd)
	}
	for cmd := range v.rootCommands {
		result["root"] = append(result["root"], cmd)
	}

	return result
}

// GetCommandCount returns the total number of valid commands
func (v *MainCommandValidator) GetCommandCount() int {
	return len(v.allCommands)
}

// getSimilarCommands finds similar commands using simple string matching
// This is a basic implementation - more sophisticated algorithm in PBI-012
func (v *MainCommandValidator) getSimilarCommands(command string, maxSuggestions int) []string {
	var suggestions []string

	// Simple similarity check - commands that contain the input or vice versa
	for cmd := range v.allCommands {
		if len(suggestions) >= maxSuggestions {
			break
		}

		// Check if command contains input or input contains command
		if strings.Contains(cmd, command) || strings.Contains(command, cmd) {
			suggestions = append(suggestions, cmd)
		}
	}

	// If no substring matches, look for commands with similar prefixes
	if len(suggestions) == 0 && len(command) > 2 {
		prefix := command[:2]
		for cmd := range v.allCommands {
			if len(suggestions) >= maxSuggestions {
				break
			}
			if strings.HasPrefix(cmd, prefix) {
				suggestions = append(suggestions, cmd)
			}
		}
	}

	return suggestions
}

// IsStandaloneCommand checks if a command is a standalone command
func (v *MainCommandValidator) IsStandaloneCommand(command string) bool {
	normalized := strings.ToLower(command)
	return standaloneCommands[normalized]
}
