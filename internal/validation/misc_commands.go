// Package validation provides command validation functionality for usacloud-update
package validation

// MiscCommands contains the complete dictionary of 3 miscellaneous usacloud commands
// and their supported subcommands. These commands serve special purposes different from
// IaaS resource management.
var MiscCommands = map[string][]string{
	// Configuration management command
	"config": {
		"list",   // List all configurations
		"show",   // Show configuration details
		"use",    // Switch to a configuration
		"create", // Create new configuration
		"edit",   // Edit existing configuration
		"delete", // Delete configuration
	},

	// REST API direct access command
	"rest": {
		"get",    // HTTP GET request
		"post",   // HTTP POST request
		"put",    // HTTP PUT request
		"delete", // HTTP DELETE request
	},

	// Web Accelerator (CDN) management command
	"webaccelerator": {
		"list",   // List web accelerators
		"read",   // Read web accelerator details
		"create", // Create new web accelerator
		"update", // Update web accelerator
		"delete", // Delete web accelerator
		"purge",  // Purge CDN cache (special subcommand)
	},
}

// GetMiscCommandSubcommands returns the list of valid subcommands for the given miscellaneous command
func GetMiscCommandSubcommands(command string) ([]string, bool) {
	subcommands, exists := MiscCommands[command]
	return subcommands, exists
}

// IsValidMiscCommand checks if the given command is a valid miscellaneous command
func IsValidMiscCommand(command string) bool {
	_, exists := MiscCommands[command]
	return exists
}

// IsValidMiscSubcommand checks if the given subcommand is valid for the given miscellaneous command
func IsValidMiscSubcommand(command, subcommand string) bool {
	subcommands, exists := MiscCommands[command]
	if !exists {
		return false
	}

	for _, sc := range subcommands {
		if sc == subcommand {
			return true
		}
	}
	return false
}

// GetAllMiscCommands returns a slice of all available miscellaneous commands
func GetAllMiscCommands() []string {
	commands := make([]string, 0, len(MiscCommands))
	for command := range MiscCommands {
		commands = append(commands, command)
	}
	return commands
}

// GetMiscCommandCount returns the total number of defined miscellaneous commands
func GetMiscCommandCount() int {
	return len(MiscCommands)
}

// IsConfigCommand checks if the command is the config management command
func IsConfigCommand(command string) bool {
	return command == "config"
}

// IsRestCommand checks if the command is the REST API command
func IsRestCommand(command string) bool {
	return command == "rest"
}

// IsWebAcceleratorCommand checks if the command is the web accelerator command
func IsWebAcceleratorCommand(command string) bool {
	return command == "webaccelerator"
}

// GetHTTPMethodSubcommands returns HTTP method-based subcommands for REST API
func GetHTTPMethodSubcommands() []string {
	return []string{"get", "post", "put", "delete"}
}

// IsHTTPMethodSubcommand checks if the subcommand is an HTTP method
func IsHTTPMethodSubcommand(subcommand string) bool {
	methods := GetHTTPMethodSubcommands()
	for _, method := range methods {
		if method == subcommand {
			return true
		}
	}
	return false
}
