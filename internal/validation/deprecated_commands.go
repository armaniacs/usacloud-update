// Package validation provides command validation functionality for usacloud-update
package validation

import "fmt"

// DeprecatedCommands maps deprecated command names to their new equivalents
// Empty string indicates the command has been discontinued with no direct replacement
var DeprecatedCommands = map[string]string{
	// Resource name changes (renamed commands)
	"iso-image":      "cdrom",     // CD-ROM resource name unification
	"startup-script": "note",      // Note resource name unification
	"ipv4":           "ipaddress", // IP address resource name unification

	// Product alias consolidation (renamed commands)
	"product-disk":     "disk-plan",     // Disk plan resource name unification
	"product-internet": "internet-plan", // Internet plan resource name unification
	"product-server":   "server-plan",   // Server plan resource name unification

	// Completely discontinued commands
	"summary":        "", // Discontinued - use bill/self/list commands or rest instead
	"object-storage": "", // Discontinued - service ended, use S3-compatible tools
	"ojs":            "", // Discontinued - alias for object-storage
}

// DeprecatedCommandMessages provides detailed explanations for each deprecated command
var DeprecatedCommandMessages = map[string]string{
	// Resource name changes
	"iso-image":      "iso-imageコマンドはv1で廃止されました。cdromコマンドを使用してください",
	"startup-script": "startup-scriptコマンドはv1で廃止されました。noteコマンドを使用してください",
	"ipv4":           "ipv4コマンドはv1で廃止されました。ipaddressコマンドを使用してください",

	// Product alias consolidation
	"product-disk":     "product-diskコマンドはv1で廃止されました。disk-planコマンドを使用してください",
	"product-internet": "product-internetコマンドはv1で廃止されました。internet-planコマンドを使用してください",
	"product-server":   "product-serverコマンドはv1で廃止されました。server-planコマンドを使用してください",

	// Completely discontinued commands
	"summary":        "summaryコマンドはv1で廃止されました。bill、self、各listコマンドまたはrestコマンドを使用してください",
	"object-storage": "object-storageコマンドはv1で廃止されました。S3互換ツールやTerraformの使用を検討してください",
	"ojs":            "ojsコマンドはv1で廃止されました。S3互換ツールやTerraformの使用を検討してください",
}

// DeprecatedCommandTypes categorizes the type of deprecation for each command
var DeprecatedCommandTypes = map[string]string{
	// Resource name changes
	"iso-image":      "renamed",
	"startup-script": "renamed",
	"ipv4":           "renamed",

	// Product alias consolidation
	"product-disk":     "renamed",
	"product-internet": "renamed",
	"product-server":   "renamed",

	// Completely discontinued commands
	"summary":        "discontinued",
	"object-storage": "discontinued",
	"ojs":            "discontinued",
}

// GetDeprecatedCommands returns the complete map of deprecated commands
func GetDeprecatedCommands() map[string]string {
	return DeprecatedCommands
}

// GetDeprecatedCommandCount returns the total number of deprecated commands
func GetDeprecatedCommandCount() int {
	return len(DeprecatedCommands)
}

// IsDeprecatedCommand checks if the given command is deprecated
func IsDeprecatedCommand(command string) bool {
	_, exists := DeprecatedCommands[command]
	return exists
}

// GetReplacementCommand returns the replacement command for a deprecated command
// Returns empty string and false if the command is not deprecated
// Returns empty string and true if the command is deprecated but discontinued
func GetReplacementCommand(command string) (string, bool) {
	replacement, exists := DeprecatedCommands[command]
	return replacement, exists
}

// GetDeprecatedCommandMessage returns the user-friendly message for a deprecated command
func GetDeprecatedCommandMessage(command string) (string, bool) {
	message, exists := DeprecatedCommandMessages[command]
	return message, exists
}

// GetDeprecatedCommandType returns the deprecation type for a command
func GetDeprecatedCommandType(command string) (string, bool) {
	commandType, exists := DeprecatedCommandTypes[command]
	return commandType, exists
}

// GetRenamedCommands returns all commands that have been renamed (not discontinued)
func GetRenamedCommands() map[string]string {
	renamed := make(map[string]string)
	for old, new := range DeprecatedCommands {
		if new != "" {
			renamed[old] = new
		}
	}
	return renamed
}

// GetDiscontinuedCommands returns all commands that have been discontinued
func GetDiscontinuedCommands() []string {
	var discontinued []string
	for old, new := range DeprecatedCommands {
		if new == "" {
			discontinued = append(discontinued, old)
		}
	}
	return discontinued
}

// IsRenamedCommand checks if a command has been renamed (has a replacement)
func IsRenamedCommand(command string) bool {
	replacement, exists := DeprecatedCommands[command]
	return exists && replacement != ""
}

// IsDiscontinuedCommand checks if a command has been discontinued (no replacement)
func IsDiscontinuedCommand(command string) bool {
	replacement, exists := DeprecatedCommands[command]
	return exists && replacement == ""
}

// GetDeprecatedCommandsByType returns commands filtered by deprecation type
func GetDeprecatedCommandsByType(commandType string) []string {
	var commands []string
	for command, depType := range DeprecatedCommandTypes {
		if depType == commandType {
			commands = append(commands, command)
		}
	}
	return commands
}

// GetAllDeprecatedCommandTypes returns all available deprecation types
func GetAllDeprecatedCommandTypes() []string {
	typeMap := make(map[string]bool)
	for _, commandType := range DeprecatedCommandTypes {
		typeMap[commandType] = true
	}

	var types []string
	for commandType := range typeMap {
		types = append(types, commandType)
	}
	return types
}

// ValidateDeprecatedCommandConsistency checks internal consistency of the deprecated command maps
func ValidateDeprecatedCommandConsistency() error {
	// Check that all commands in DeprecatedCommands have corresponding entries in other maps
	for command := range DeprecatedCommands {
		if _, exists := DeprecatedCommandMessages[command]; !exists {
			return fmt.Errorf("command %s missing from DeprecatedCommandMessages", command)
		}
		if _, exists := DeprecatedCommandTypes[command]; !exists {
			return fmt.Errorf("command %s missing from DeprecatedCommandTypes", command)
		}
	}

	// Check that all entries in other maps have corresponding commands
	for command := range DeprecatedCommandMessages {
		if _, exists := DeprecatedCommands[command]; !exists {
			return fmt.Errorf("command %s in DeprecatedCommandMessages but not in DeprecatedCommands", command)
		}
	}

	for command := range DeprecatedCommandTypes {
		if _, exists := DeprecatedCommands[command]; !exists {
			return fmt.Errorf("command %s in DeprecatedCommandTypes but not in DeprecatedCommands", command)
		}
	}

	return nil
}
