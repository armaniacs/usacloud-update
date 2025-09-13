// Package validation provides command validation functionality for usacloud-update
package validation

// RootCommands contains the complete dictionary of 3 root-level usacloud commands
// and their supported subcommands. These commands are tool management functions
// executed directly at the root level.
var RootCommands = map[string][]string{
	// Shell completion script generation command
	"completion": {
		"bash",       // Generate bash completion script
		"zsh",        // Generate zsh completion script
		"fish",       // Generate fish completion script
		"powershell", // Generate PowerShell completion script
	},

	// Version display command (standalone - no subcommands)
	"version": {},

	// Self-update command (standalone - no subcommands)
	"update-self": {},
}

// GetRootCommandSubcommands returns the list of valid subcommands for the given root command
func GetRootCommandSubcommands(command string) ([]string, bool) {
	subcommands, exists := RootCommands[command]
	return subcommands, exists
}

// IsValidRootCommand checks if the given command is a valid root command
func IsValidRootCommand(command string) bool {
	_, exists := RootCommands[command]
	return exists
}

// IsValidRootSubcommand checks if the given subcommand is valid for the given root command
func IsValidRootSubcommand(command, subcommand string) bool {
	subcommands, exists := RootCommands[command]
	if !exists {
		return false
	}

	// For standalone commands (no subcommands), empty subcommand is valid
	if len(subcommands) == 0 && subcommand == "" {
		return true
	}

	// For commands with subcommands, check if subcommand matches
	for _, sc := range subcommands {
		if sc == subcommand {
			return true
		}
	}
	return false
}

// GetAllRootCommands returns a slice of all available root commands
func GetAllRootCommands() []string {
	commands := make([]string, 0, len(RootCommands))
	for command := range RootCommands {
		commands = append(commands, command)
	}
	return commands
}

// GetRootCommandCount returns the total number of defined root commands
func GetRootCommandCount() int {
	return len(RootCommands)
}

// IsStandaloneCommand checks if the command is a standalone command (no subcommands)
func IsStandaloneCommand(command string) bool {
	subcommands, exists := RootCommands[command]
	return exists && len(subcommands) == 0
}

// IsCompletionCommand checks if the command is the completion command
func IsCompletionCommand(command string) bool {
	return command == "completion"
}

// IsVersionCommand checks if the command is the version command
func IsVersionCommand(command string) bool {
	return command == "version"
}

// IsUpdateSelfCommand checks if the command is the update-self command
func IsUpdateSelfCommand(command string) bool {
	return command == "update-self"
}

// GetCompletionShells returns the list of supported shell types for completion
func GetCompletionShells() []string {
	return []string{"bash", "zsh", "fish", "powershell"}
}

// IsValidCompletionShell checks if the shell type is supported for completion
func IsValidCompletionShell(shell string) bool {
	shells := GetCompletionShells()
	for _, s := range shells {
		if s == shell {
			return true
		}
	}
	return false
}

// ValidateRootCommandUsage validates if a root command is used correctly
// Returns an error message if the usage is invalid, empty string if valid
func ValidateRootCommandUsage(command, subcommand string) string {
	if !IsValidRootCommand(command) {
		return "invalid root command"
	}

	// For standalone commands, subcommand should be empty
	if IsStandaloneCommand(command) && subcommand != "" {
		return "standalone command does not accept subcommands"
	}

	// For completion command, subcommand is required and must be valid
	if IsCompletionCommand(command) {
		if subcommand == "" {
			return "completion command requires a shell type subcommand"
		}
		if !IsValidRootSubcommand(command, subcommand) {
			return "invalid shell type for completion command"
		}
	}

	return "" // Valid usage
}
