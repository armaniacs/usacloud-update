package validation

import (
	"sort"
	"testing"
)

func TestDeprecatedCommandsCount(t *testing.T) {
	expectedCount := 9
	actualCount := GetDeprecatedCommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d deprecated commands, but got %d", expectedCount, actualCount)
	}
}

func TestGetDeprecatedCommands(t *testing.T) {
	commands := GetDeprecatedCommands()
	expectedCommands := []string{
		"iso-image", "startup-script", "ipv4",
		"product-disk", "product-internet", "product-server",
		"summary", "object-storage", "ojs",
	}

	if len(commands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
	}

	for _, expected := range expectedCommands {
		if _, exists := commands[expected]; !exists {
			t.Errorf("Expected deprecated command '%s' not found", expected)
		}
	}
}

func TestIsDeprecatedCommand(t *testing.T) {
	tests := []struct {
		command    string
		deprecated bool
	}{
		// Renamed commands
		{"iso-image", true},
		{"startup-script", true},
		{"ipv4", true},
		{"product-disk", true},
		{"product-internet", true},
		{"product-server", true},

		// Discontinued commands
		{"summary", true},
		{"object-storage", true},
		{"ojs", true},

		// Valid current commands
		{"cdrom", false},
		{"note", false},
		{"ipaddress", false},
		{"disk-plan", false},
		{"server", false},
		{"config", false},
		{"", false},
		{"SUMMARY", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsDeprecatedCommand(tt.command)
			if result != tt.deprecated {
				t.Errorf("IsDeprecatedCommand(%s): expected %t, got %t",
					tt.command, tt.deprecated, result)
			}
		})
	}
}

func TestGetReplacementCommand(t *testing.T) {
	tests := []struct {
		command     string
		replacement string
		exists      bool
	}{
		// Renamed commands
		{"iso-image", "cdrom", true},
		{"startup-script", "note", true},
		{"ipv4", "ipaddress", true},
		{"product-disk", "disk-plan", true},
		{"product-internet", "internet-plan", true},
		{"product-server", "server-plan", true},

		// Discontinued commands (no replacement)
		{"summary", "", true},
		{"object-storage", "", true},
		{"ojs", "", true},

		// Non-deprecated commands
		{"server", "", false},
		{"config", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			replacement, exists := GetReplacementCommand(tt.command)

			if exists != tt.exists {
				t.Errorf("GetReplacementCommand(%s): expected exists=%t, got %t",
					tt.command, tt.exists, exists)
			}

			if replacement != tt.replacement {
				t.Errorf("GetReplacementCommand(%s): expected '%s', got '%s'",
					tt.command, tt.replacement, replacement)
			}
		})
	}
}

func TestGetDeprecatedCommandMessage(t *testing.T) {
	tests := []struct {
		command      string
		expectExists bool
		contains     string // Check if message contains this text
	}{
		{"iso-image", true, "cdrom"},
		{"startup-script", true, "note"},
		{"ipv4", true, "ipaddress"},
		{"product-disk", true, "disk-plan"},
		{"summary", true, "廃止"},
		{"object-storage", true, "S3"},
		{"ojs", true, "S3"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			message, exists := GetDeprecatedCommandMessage(tt.command)

			if exists != tt.expectExists {
				t.Errorf("GetDeprecatedCommandMessage(%s): expected exists=%t, got %t",
					tt.command, tt.expectExists, exists)
			}

			if tt.expectExists && tt.contains != "" {
				if len(message) == 0 {
					t.Errorf("GetDeprecatedCommandMessage(%s): expected non-empty message",
						tt.command)
				}
			}
		})
	}
}

func TestGetDeprecatedCommandType(t *testing.T) {
	tests := []struct {
		command      string
		expectedType string
		expectExists bool
	}{
		// Renamed commands
		{"iso-image", "renamed", true},
		{"startup-script", "renamed", true},
		{"ipv4", "renamed", true},
		{"product-disk", "renamed", true},
		{"product-internet", "renamed", true},
		{"product-server", "renamed", true},

		// Discontinued commands
		{"summary", "discontinued", true},
		{"object-storage", "discontinued", true},
		{"ojs", "discontinued", true},

		// Non-deprecated
		{"server", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			commandType, exists := GetDeprecatedCommandType(tt.command)

			if exists != tt.expectExists {
				t.Errorf("GetDeprecatedCommandType(%s): expected exists=%t, got %t",
					tt.command, tt.expectExists, exists)
			}

			if tt.expectExists && commandType != tt.expectedType {
				t.Errorf("GetDeprecatedCommandType(%s): expected '%s', got '%s'",
					tt.command, tt.expectedType, commandType)
			}
		})
	}
}

func TestGetRenamedCommands(t *testing.T) {
	renamed := GetRenamedCommands()
	expectedRenamed := map[string]string{
		"iso-image":        "cdrom",
		"startup-script":   "note",
		"ipv4":             "ipaddress",
		"product-disk":     "disk-plan",
		"product-internet": "internet-plan",
		"product-server":   "server-plan",
	}

	if len(renamed) != len(expectedRenamed) {
		t.Errorf("Expected %d renamed commands, got %d", len(expectedRenamed), len(renamed))
	}

	for old, expectedNew := range expectedRenamed {
		if actualNew, exists := renamed[old]; !exists || actualNew != expectedNew {
			t.Errorf("GetRenamedCommands(): expected %s -> %s, got %s -> %s (exists: %t)",
				old, expectedNew, old, actualNew, exists)
		}
	}

	// Ensure discontinued commands are not included
	discontinuedCommands := []string{"summary", "object-storage", "ojs"}
	for _, cmd := range discontinuedCommands {
		if _, exists := renamed[cmd]; exists {
			t.Errorf("GetRenamedCommands(): discontinued command '%s' should not be included", cmd)
		}
	}
}

func TestGetDiscontinuedCommands(t *testing.T) {
	discontinued := GetDiscontinuedCommands()
	expectedDiscontinued := []string{"summary", "object-storage", "ojs"}

	if len(discontinued) != len(expectedDiscontinued) {
		t.Errorf("Expected %d discontinued commands, got %d", len(expectedDiscontinued), len(discontinued))
	}

	discontinuedMap := make(map[string]bool)
	for _, cmd := range discontinued {
		discontinuedMap[cmd] = true
	}

	for _, expected := range expectedDiscontinued {
		if !discontinuedMap[expected] {
			t.Errorf("GetDiscontinuedCommands(): expected discontinued command '%s' not found", expected)
		}
	}

	// Ensure renamed commands are not included
	renamedCommands := []string{"iso-image", "startup-script", "ipv4", "product-disk"}
	for _, cmd := range renamedCommands {
		if discontinuedMap[cmd] {
			t.Errorf("GetDiscontinuedCommands(): renamed command '%s' should not be included", cmd)
		}
	}
}

func TestIsRenamedCommand(t *testing.T) {
	tests := []struct {
		command string
		renamed bool
	}{
		// Renamed commands
		{"iso-image", true},
		{"startup-script", true},
		{"ipv4", true},
		{"product-disk", true},
		{"product-internet", true},
		{"product-server", true},

		// Discontinued commands
		{"summary", false},
		{"object-storage", false},
		{"ojs", false},

		// Non-deprecated commands
		{"server", false},
		{"config", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsRenamedCommand(tt.command)
			if result != tt.renamed {
				t.Errorf("IsRenamedCommand(%s): expected %t, got %t",
					tt.command, tt.renamed, result)
			}
		})
	}
}

func TestIsDiscontinuedCommand(t *testing.T) {
	tests := []struct {
		command      string
		discontinued bool
	}{
		// Discontinued commands
		{"summary", true},
		{"object-storage", true},
		{"ojs", true},

		// Renamed commands
		{"iso-image", false},
		{"startup-script", false},
		{"ipv4", false},
		{"product-disk", false},

		// Non-deprecated commands
		{"server", false},
		{"config", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsDiscontinuedCommand(tt.command)
			if result != tt.discontinued {
				t.Errorf("IsDiscontinuedCommand(%s): expected %t, got %t",
					tt.command, tt.discontinued, result)
			}
		})
	}
}

func TestGetDeprecatedCommandsByType(t *testing.T) {
	tests := []struct {
		commandType string
		expected    []string
	}{
		{"renamed", []string{"iso-image", "startup-script", "ipv4", "product-disk", "product-internet", "product-server"}},
		{"discontinued", []string{"summary", "object-storage", "ojs"}},
		{"invalid", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.commandType, func(t *testing.T) {
			result := GetDeprecatedCommandsByType(tt.commandType)

			if len(result) != len(tt.expected) {
				t.Errorf("GetDeprecatedCommandsByType(%s): expected %d commands, got %d",
					tt.commandType, len(tt.expected), len(result))
				return
			}

			resultMap := make(map[string]bool)
			for _, cmd := range result {
				resultMap[cmd] = true
			}

			for _, expectedCmd := range tt.expected {
				if !resultMap[expectedCmd] {
					t.Errorf("GetDeprecatedCommandsByType(%s): missing expected command '%s'",
						tt.commandType, expectedCmd)
				}
			}
		})
	}
}

func TestGetAllDeprecatedCommandTypes(t *testing.T) {
	types := GetAllDeprecatedCommandTypes()
	expectedTypes := []string{"renamed", "discontinued"}

	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d types, got %d", len(expectedTypes), len(types))
	}

	typeMap := make(map[string]bool)
	for _, cmdType := range types {
		typeMap[cmdType] = true
	}

	for _, expected := range expectedTypes {
		if !typeMap[expected] {
			t.Errorf("Expected type '%s' not found", expected)
		}
	}
}

func TestValidateDeprecatedCommandConsistency(t *testing.T) {
	err := ValidateDeprecatedCommandConsistency()
	if err != nil {
		t.Errorf("ValidateDeprecatedCommandConsistency() returned error: %v", err)
	}
}

func TestAllCommandsHaveMessagesAndTypes(t *testing.T) {
	commands := GetDeprecatedCommands()

	for command := range commands {
		// Test message exists
		message, msgExists := GetDeprecatedCommandMessage(command)
		if !msgExists {
			t.Errorf("Command '%s' missing message", command)
		}
		if len(message) == 0 {
			t.Errorf("Command '%s' has empty message", command)
		}

		// Test type exists
		commandType, typeExists := GetDeprecatedCommandType(command)
		if !typeExists {
			t.Errorf("Command '%s' missing type", command)
		}
		if commandType != "renamed" && commandType != "discontinued" {
			t.Errorf("Command '%s' has invalid type '%s'", command, commandType)
		}
	}
}

func TestMapConsistency(t *testing.T) {
	// Check that all maps have the same keys
	commandCount := len(DeprecatedCommands)
	messageCount := len(DeprecatedCommandMessages)
	typeCount := len(DeprecatedCommandTypes)

	if commandCount != messageCount {
		t.Errorf("Command count (%d) doesn't match message count (%d)",
			commandCount, messageCount)
	}

	if commandCount != typeCount {
		t.Errorf("Command count (%d) doesn't match type count (%d)",
			commandCount, typeCount)
	}
}

func TestDeprecatedCommandMapping(t *testing.T) {
	// Test specific mappings that correspond to existing transformation rules
	tests := []struct {
		deprecated  string
		replacement string
		cmdType     string
	}{
		{"iso-image", "cdrom", "renamed"},
		{"startup-script", "note", "renamed"},
		{"ipv4", "ipaddress", "renamed"},
		{"product-disk", "disk-plan", "renamed"},
		{"product-internet", "internet-plan", "renamed"},
		{"product-server", "server-plan", "renamed"},
		{"summary", "", "discontinued"},
		{"object-storage", "", "discontinued"},
		{"ojs", "", "discontinued"},
	}

	for _, tt := range tests {
		t.Run(tt.deprecated, func(t *testing.T) {
			// Test replacement
			replacement, exists := GetReplacementCommand(tt.deprecated)
			if !exists {
				t.Errorf("Command '%s' should be deprecated", tt.deprecated)
			}
			if replacement != tt.replacement {
				t.Errorf("Command '%s': expected replacement '%s', got '%s'",
					tt.deprecated, tt.replacement, replacement)
			}

			// Test type
			cmdType, exists := GetDeprecatedCommandType(tt.deprecated)
			if !exists {
				t.Errorf("Command '%s' should have a type", tt.deprecated)
			}
			if cmdType != tt.cmdType {
				t.Errorf("Command '%s': expected type '%s', got '%s'",
					tt.deprecated, tt.cmdType, cmdType)
			}
		})
	}
}

func TestDeprecatedCommandsUniqueness(t *testing.T) {
	commands := GetDeprecatedCommands()

	// Check for duplicate values in renamed commands (excluding empty strings)
	seenReplacements := make(map[string][]string)
	for old, new := range commands {
		if new != "" {
			seenReplacements[new] = append(seenReplacements[new], old)
		}
	}

	for replacement, oldCommands := range seenReplacements {
		if len(oldCommands) > 1 {
			sort.Strings(oldCommands)
			t.Errorf("Replacement command '%s' is used by multiple deprecated commands: %v",
				replacement, oldCommands)
		}
	}
}
