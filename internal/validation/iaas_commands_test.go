package validation

import (
	"sort"
	"testing"
)

func TestIaaSCommandsCount(t *testing.T) {
	expectedCount := 49
	actualCount := GetTotalCommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d IaaS commands, but got %d", expectedCount, actualCount)
	}
}

func TestGetIaaSCommandSubcommands(t *testing.T) {
	tests := []struct {
		command          string
		expectExists     bool
		expectedMinCount int
	}{
		{"server", true, 5},  // should have at least basic CRUD + boot/shutdown
		{"disk", true, 5},    // should have at least basic CRUD + connect/disconnect
		{"archive", true, 5}, // should have at least basic CRUD + download/extract
		{"nonexistent", false, 0},
		{"", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			subcommands, exists := GetIaaSCommandSubcommands(tt.command)

			if exists != tt.expectExists {
				t.Errorf("GetIaaSCommandSubcommands(%s): expected exists=%t, got %t",
					tt.command, tt.expectExists, exists)
			}

			if tt.expectExists && len(subcommands) < tt.expectedMinCount {
				t.Errorf("GetIaaSCommandSubcommands(%s): expected at least %d subcommands, got %d",
					tt.command, tt.expectedMinCount, len(subcommands))
			}
		})
	}
}

func TestIsValidIaaSCommand(t *testing.T) {
	tests := []struct {
		command string
		valid   bool
	}{
		{"server", true},
		{"disk", true},
		{"archive", true},
		{"database", true},
		{"dns", true},
		{"nonexistent", false},
		{"", false},
		{"SERVER", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsValidIaaSCommand(tt.command)
			if result != tt.valid {
				t.Errorf("IsValidIaaSCommand(%s): expected %t, got %t",
					tt.command, tt.valid, result)
			}
		})
	}
}

func TestIsValidIaaSSubcommand(t *testing.T) {
	tests := []struct {
		command    string
		subcommand string
		valid      bool
	}{
		{"server", "list", true},
		{"server", "boot", true},
		{"server", "shutdown", true},
		{"server", "nonexistent", false},
		{"disk", "connect", true},
		{"disk", "disconnect", true},
		{"disk", "boot", false}, // disk doesn't have boot command
		{"nonexistent", "list", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command+"_"+tt.subcommand, func(t *testing.T) {
			result := IsValidIaaSSubcommand(tt.command, tt.subcommand)
			if result != tt.valid {
				t.Errorf("IsValidIaaSSubcommand(%s, %s): expected %t, got %t",
					tt.command, tt.subcommand, tt.valid, result)
			}
		})
	}
}

func TestGetAllIaaSCommands(t *testing.T) {
	commands := GetAllIaaSCommands()

	// Check that we have the expected count
	expectedCount := 49
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// Check that some key commands are present
	expectedCommands := []string{"server", "disk", "database", "dns", "archive"}
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

func TestAllCommandsHaveBasicSubcommands(t *testing.T) {
	// Most commands should have basic CRUD operations
	basicCommands := []string{"list", "read"}

	for command := range IaaSCommands {
		subcommands := IaaSCommands[command]

		// Skip read-only commands that might not have all basic operations
		if command == "authstatus" || command == "bill" || command == "coupon" ||
			command == "diskplan" || command == "internetplan" || command == "license" ||
			command == "licenseinfo" || command == "privatehostplan" || command == "region" ||
			command == "serverplan" || command == "serviceclass" || command == "zone" ||
			command == "category" || command == "ipaddress" || command == "self" {
			continue
		}

		subcommandMap := make(map[string]bool)
		for _, sc := range subcommands {
			subcommandMap[sc] = true
		}

		for _, basicCmd := range basicCommands {
			if !subcommandMap[basicCmd] {
				t.Errorf("Command '%s' missing basic subcommand '%s'", command, basicCmd)
			}
		}
	}
}

func TestSpecificCommandSubcommands(t *testing.T) {
	tests := []struct {
		command             string
		expectedSubcommands []string
	}{
		{
			"server",
			[]string{"list", "read", "create", "update", "delete", "boot", "shutdown", "reset"},
		},
		{
			"vpcrouter",
			[]string{"list", "read", "create", "update", "delete", "boot", "shutdown"},
		},
		{
			"database",
			[]string{"list", "read", "create", "update", "delete", "boot", "shutdown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			subcommands, exists := GetIaaSCommandSubcommands(tt.command)

			if !exists {
				t.Fatalf("Command '%s' not found", tt.command)
			}

			subcommandMap := make(map[string]bool)
			for _, sc := range subcommands {
				subcommandMap[sc] = true
			}

			for _, expected := range tt.expectedSubcommands {
				if !subcommandMap[expected] {
					t.Errorf("Command '%s' missing expected subcommand '%s'",
						tt.command, expected)
				}
			}
		})
	}
}

func TestCommandCategorization(t *testing.T) {
	// Test that commands are properly categorized by checking specific command groups
	storageCommands := []string{"archive", "cdrom", "disk"}
	networkCommands := []string{"swytch", "bridge", "internet", "subnet", "dns"}
	serverCommands := []string{"server", "serverplan"}

	allCommands := GetAllIaaSCommands()
	commandMap := make(map[string]bool)
	for _, cmd := range allCommands {
		commandMap[cmd] = true
	}

	// Check storage commands
	for _, cmd := range storageCommands {
		if !commandMap[cmd] {
			t.Errorf("Storage command '%s' not found in IaaS commands", cmd)
		}
	}

	// Check network commands
	for _, cmd := range networkCommands {
		if !commandMap[cmd] {
			t.Errorf("Network command '%s' not found in IaaS commands", cmd)
		}
	}

	// Check server commands
	for _, cmd := range serverCommands {
		if !commandMap[cmd] {
			t.Errorf("Server command '%s' not found in IaaS commands", cmd)
		}
	}
}
