package validation

import (
	"sort"
	"testing"
)

func TestMiscCommandsCount(t *testing.T) {
	expectedCount := 3
	actualCount := GetMiscCommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d miscellaneous commands, but got %d", expectedCount, actualCount)
	}
}

func TestGetMiscCommandSubcommands(t *testing.T) {
	tests := []struct {
		command          string
		expectExists     bool
		expectedMinCount int
	}{
		{"config", true, 6},         // should have 6 config management subcommands
		{"rest", true, 4},           // should have 4 HTTP method subcommands
		{"webaccelerator", true, 6}, // should have 6 subcommands including purge
		{"nonexistent", false, 0},
		{"", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			subcommands, exists := GetMiscCommandSubcommands(tt.command)

			if exists != tt.expectExists {
				t.Errorf("GetMiscCommandSubcommands(%s): expected exists=%t, got %t",
					tt.command, tt.expectExists, exists)
			}

			if tt.expectExists && len(subcommands) < tt.expectedMinCount {
				t.Errorf("GetMiscCommandSubcommands(%s): expected at least %d subcommands, got %d",
					tt.command, tt.expectedMinCount, len(subcommands))
			}
		})
	}
}

func TestIsValidMiscCommand(t *testing.T) {
	tests := []struct {
		command string
		valid   bool
	}{
		{"config", true},
		{"rest", true},
		{"webaccelerator", true},
		{"server", false}, // IaaS command, not misc
		{"nonexistent", false},
		{"", false},
		{"CONFIG", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := IsValidMiscCommand(tt.command)
			if result != tt.valid {
				t.Errorf("IsValidMiscCommand(%s): expected %t, got %t",
					tt.command, tt.valid, result)
			}
		})
	}
}

func TestIsValidMiscSubcommand(t *testing.T) {
	tests := []struct {
		command    string
		subcommand string
		valid      bool
	}{
		{"config", "list", true},
		{"config", "use", true},
		{"config", "edit", true},
		{"config", "boot", false}, // config doesn't have boot command
		{"rest", "get", true},
		{"rest", "post", true},
		{"rest", "put", true},
		{"rest", "delete", true},
		{"rest", "list", false}, // rest doesn't have list command
		{"webaccelerator", "purge", true},
		{"webaccelerator", "list", true},
		{"webaccelerator", "boot", false}, // webaccelerator doesn't have boot
		{"nonexistent", "list", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.command+"_"+tt.subcommand, func(t *testing.T) {
			result := IsValidMiscSubcommand(tt.command, tt.subcommand)
			if result != tt.valid {
				t.Errorf("IsValidMiscSubcommand(%s, %s): expected %t, got %t",
					tt.command, tt.subcommand, tt.valid, result)
			}
		})
	}
}

func TestGetAllMiscCommands(t *testing.T) {
	commands := GetAllMiscCommands()

	// Check that we have the expected count
	expectedCount := 3
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// Check that all expected commands are present
	expectedCommands := []string{"config", "rest", "webaccelerator"}
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

func TestSpecificCommandTypes(t *testing.T) {
	tests := []struct {
		command  string
		isConfig bool
		isRest   bool
		isWebAcc bool
	}{
		{"config", true, false, false},
		{"rest", false, true, false},
		{"webaccelerator", false, false, true},
		{"server", false, false, false}, // not a misc command
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			configResult := IsConfigCommand(tt.command)
			restResult := IsRestCommand(tt.command)
			webAccResult := IsWebAcceleratorCommand(tt.command)

			if configResult != tt.isConfig {
				t.Errorf("IsConfigCommand(%s): expected %t, got %t",
					tt.command, tt.isConfig, configResult)
			}

			if restResult != tt.isRest {
				t.Errorf("IsRestCommand(%s): expected %t, got %t",
					tt.command, tt.isRest, restResult)
			}

			if webAccResult != tt.isWebAcc {
				t.Errorf("IsWebAcceleratorCommand(%s): expected %t, got %t",
					tt.command, tt.isWebAcc, webAccResult)
			}
		})
	}
}

func TestHTTPMethodSubcommands(t *testing.T) {
	httpMethods := GetHTTPMethodSubcommands()
	expectedMethods := []string{"get", "post", "put", "delete"}

	if len(httpMethods) != len(expectedMethods) {
		t.Errorf("Expected %d HTTP methods, got %d", len(expectedMethods), len(httpMethods))
	}

	methodMap := make(map[string]bool)
	for _, method := range httpMethods {
		methodMap[method] = true
	}

	for _, expected := range expectedMethods {
		if !methodMap[expected] {
			t.Errorf("Expected HTTP method '%s' not found", expected)
		}
	}

	// Test IsHTTPMethodSubcommand function
	for _, method := range expectedMethods {
		if !IsHTTPMethodSubcommand(method) {
			t.Errorf("IsHTTPMethodSubcommand(%s): expected true, got false", method)
		}
	}

	// Test with non-HTTP methods
	nonMethods := []string{"list", "read", "create", "purge", "boot"}
	for _, nonMethod := range nonMethods {
		if IsHTTPMethodSubcommand(nonMethod) {
			t.Errorf("IsHTTPMethodSubcommand(%s): expected false, got true", nonMethod)
		}
	}
}

func TestConfigCommandSubcommands(t *testing.T) {
	expectedConfigSubcommands := []string{"list", "show", "use", "create", "edit", "delete"}

	subcommands, exists := GetMiscCommandSubcommands("config")
	if !exists {
		t.Fatal("Config command not found")
	}

	if len(subcommands) != len(expectedConfigSubcommands) {
		t.Errorf("Expected %d config subcommands, got %d",
			len(expectedConfigSubcommands), len(subcommands))
	}

	subcommandMap := make(map[string]bool)
	for _, sc := range subcommands {
		subcommandMap[sc] = true
	}

	for _, expected := range expectedConfigSubcommands {
		if !subcommandMap[expected] {
			t.Errorf("Expected config subcommand '%s' not found", expected)
		}
	}
}

func TestWebAcceleratorSpecialSubcommand(t *testing.T) {
	// Test that webaccelerator has the special 'purge' subcommand
	subcommands, exists := GetMiscCommandSubcommands("webaccelerator")
	if !exists {
		t.Fatal("webaccelerator command not found")
	}

	hasPurge := false
	for _, sc := range subcommands {
		if sc == "purge" {
			hasPurge = true
			break
		}
	}

	if !hasPurge {
		t.Error("webaccelerator command missing special 'purge' subcommand")
	}

	// Verify purge is valid for webaccelerator
	if !IsValidMiscSubcommand("webaccelerator", "purge") {
		t.Error("purge should be valid subcommand for webaccelerator")
	}

	// Verify purge is not valid for other commands
	if IsValidMiscSubcommand("config", "purge") {
		t.Error("purge should not be valid subcommand for config")
	}

	if IsValidMiscSubcommand("rest", "purge") {
		t.Error("purge should not be valid subcommand for rest")
	}
}
