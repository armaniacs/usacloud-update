package validation

import (
	"sort"
	"testing"
)

func TestRootCommandsCount(t *testing.T) {
	expectedCount := 3
	actualCount := GetRootCommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d root commands, but got %d", expectedCount, actualCount)
	}
}

func TestGetRootCommandSubcommands(t *testing.T) {
	tests := []struct {
		command          string
		expectExists     bool
		expectedMinCount int
	}{
		{"completion", true, 4},  // should have 4 shell completion subcommands
		{"version", true, 0},     // standalone command - no subcommands
		{"update-self", true, 0}, // standalone command - no subcommands
		{"nonexistent", false, 0},
		{"", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			subcommands, exists := GetRootCommandSubcommands(tt.command)

			if exists != tt.expectExists {
				t.Errorf("GetRootCommandSubcommands(%s): expected exists=%t, got %t",
					tt.command, tt.expectExists, exists)
			}

			if tt.expectExists && len(subcommands) != tt.expectedMinCount {
				t.Errorf("GetRootCommandSubcommands(%s): expected %d subcommands, got %d",
					tt.command, tt.expectedMinCount, len(subcommands))
			}
		})
	}
}

func TestIsValidRootCommand(t *testing.T) {
	tests := []struct {
		command string
		valid   bool
	}{
		{"completion", true},
		{"version", true},
		{"update-self", true},
		{"server", false}, // IaaS command, not root
		{"config", false}, // Misc command, not root
		{"nonexistent", false},
		{"", false},
		{"COMPLETION", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsValidRootCommand(tt.command)
			if result != tt.valid {
				t.Errorf("IsValidRootCommand(%s): expected %t, got %t",
					tt.command, tt.valid, result)
			}
		})
	}
}

func TestIsValidRootSubcommand(t *testing.T) {
	tests := []struct {
		command    string
		subcommand string
		valid      bool
	}{
		{"completion", "bash", true},
		{"completion", "zsh", true},
		{"completion", "fish", true},
		{"completion", "powershell", true},
		{"completion", "invalid", false},
		{"completion", "", false},            // completion requires subcommand
		{"version", "", true},                // standalone command
		{"version", "subcommand", false},     // standalone doesn't accept subcommands
		{"update-self", "", true},            // standalone command
		{"update-self", "subcommand", false}, // standalone doesn't accept subcommands
		{"nonexistent", "anything", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command+"_"+tt.subcommand, func(t *testing.T) {
			result := IsValidRootSubcommand(tt.command, tt.subcommand)
			if result != tt.valid {
				t.Errorf("IsValidRootSubcommand(%s, %s): expected %t, got %t",
					tt.command, tt.subcommand, tt.valid, result)
			}
		})
	}
}

func TestGetAllRootCommands(t *testing.T) {
	commands := GetAllRootCommands()

	// Check that we have the expected count
	expectedCount := 3
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// Check that all expected commands are present
	expectedCommands := []string{"completion", "version", "update-self"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command '%s' not found in commands list", expected)
		}
	}

	// Check for duplicates
	sort.Strings(commands)
	for i := 1; i < len(commands); i++ {
		if commands[i] == commands[i-1] {
			t.Errorf("Duplicate command found: %s", commands[i])
		}
	}
}

func TestStandaloneCommands(t *testing.T) {
	tests := []struct {
		command    string
		standalone bool
	}{
		{"version", true},
		{"update-self", true},
		{"completion", false}, // has subcommands
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsStandaloneCommand(tt.command)
			if result != tt.standalone {
				t.Errorf("IsStandaloneCommand(%s): expected %t, got %t",
					tt.command, tt.standalone, result)
			}
		})
	}
}

func TestRootSpecificCommandTypes(t *testing.T) {
	tests := []struct {
		command      string
		isCompletion bool
		isVersion    bool
		isUpdateSelf bool
	}{
		{"completion", true, false, false},
		{"version", false, true, false},
		{"update-self", false, false, true},
		{"server", false, false, false}, // not a root command
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			completionResult := IsCompletionCommand(tt.command)
			versionResult := IsVersionCommand(tt.command)
			updateSelfResult := IsUpdateSelfCommand(tt.command)

			if completionResult != tt.isCompletion {
				t.Errorf("IsCompletionCommand(%s): expected %t, got %t",
					tt.command, tt.isCompletion, completionResult)
			}

			if versionResult != tt.isVersion {
				t.Errorf("IsVersionCommand(%s): expected %t, got %t",
					tt.command, tt.isVersion, versionResult)
			}

			if updateSelfResult != tt.isUpdateSelf {
				t.Errorf("IsUpdateSelfCommand(%s): expected %t, got %t",
					tt.command, tt.isUpdateSelf, updateSelfResult)
			}
		})
	}
}

func TestCompletionShells(t *testing.T) {
	shells := GetCompletionShells()
	expectedShells := []string{"bash", "zsh", "fish", "powershell"}

	if len(shells) != len(expectedShells) {
		t.Errorf("Expected %d completion shells, got %d", len(expectedShells), len(shells))
	}

	shellMap := make(map[string]bool)
	for _, shell := range shells {
		shellMap[shell] = true
	}

	for _, expected := range expectedShells {
		if !shellMap[expected] {
			t.Errorf("Expected shell '%s' not found", expected)
		}
	}

	// Test IsValidCompletionShell function
	for _, shell := range expectedShells {
		if !IsValidCompletionShell(shell) {
			t.Errorf("IsValidCompletionShell(%s): expected true, got false", shell)
		}
	}

	// Test with invalid shells
	invalidShells := []string{"cmd", "sh", "csh", "tcsh", "unknown"}
	for _, invalidShell := range invalidShells {
		if IsValidCompletionShell(invalidShell) {
			t.Errorf("IsValidCompletionShell(%s): expected false, got true", invalidShell)
		}
	}
}

func TestValidateRootCommandUsage(t *testing.T) {
	tests := []struct {
		command    string
		subcommand string
		expectErr  string
	}{
		{"completion", "bash", ""}, // valid
		{"completion", "zsh", ""},  // valid
		{"completion", "", "completion command requires a shell type subcommand"}, // completion needs subcommand
		{"completion", "invalid", "invalid shell type for completion command"},    // invalid shell
		{"version", "", ""}, // valid standalone
		{"version", "subcommand", "standalone command does not accept subcommands"}, // standalone with subcommand
		{"update-self", "", ""}, // valid standalone
		{"update-self", "subcommand", "standalone command does not accept subcommands"}, // standalone with subcommand
		{"invalid", "", "invalid root command"},                                         // invalid command
	}

	for _, tt := range tests {
		t.Run(tt.command+"_"+tt.subcommand, func(t *testing.T) {
			result := ValidateRootCommandUsage(tt.command, tt.subcommand)
			if result != tt.expectErr {
				t.Errorf("ValidateRootCommandUsage(%s, %s): expected '%s', got '%s'",
					tt.command, tt.subcommand, tt.expectErr, result)
			}
		})
	}
}

func TestCompletionCommandSubcommands(t *testing.T) {
	expectedCompletionSubcommands := []string{"bash", "zsh", "fish", "powershell"}

	subcommands, exists := GetRootCommandSubcommands("completion")
	if !exists {
		t.Fatal("Completion command not found")
	}

	if len(subcommands) != len(expectedCompletionSubcommands) {
		t.Errorf("Expected %d completion subcommands, got %d",
			len(expectedCompletionSubcommands), len(subcommands))
	}

	subcommandMap := make(map[string]bool)
	for _, sc := range subcommands {
		subcommandMap[sc] = true
	}

	for _, expected := range expectedCompletionSubcommands {
		if !subcommandMap[expected] {
			t.Errorf("Expected completion subcommand '%s' not found", expected)
		}
	}
}

func TestStandaloneCommandsNoSubcommands(t *testing.T) {
	standaloneCommands := []string{"version", "update-self"}

	for _, cmd := range standaloneCommands {
		t.Run(cmd, func(t *testing.T) {
			subcommands, exists := GetRootCommandSubcommands(cmd)
			if !exists {
				t.Errorf("Command '%s' should exist", cmd)
			}

			if len(subcommands) != 0 {
				t.Errorf("Standalone command '%s' should have no subcommands, got %d",
					cmd, len(subcommands))
			}
		})
	}
}
