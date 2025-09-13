package validation

import (
	"strings"
	"testing"
)

func TestNewMainCommandValidator(t *testing.T) {
	validator := NewMainCommandValidator()

	if validator == nil {
		t.Error("Expected validator to be created, got nil")
	}

	// Check that commands were initialized
	expectedTotal := 50 // 44 IaaS + 3 misc + 3 root
	actualTotal := validator.GetCommandCount()
	if actualTotal < expectedTotal {
		t.Errorf("Expected at least %d commands, got %d", expectedTotal, actualTotal)
	}
}

func TestValidateValidCommands(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command      string
		expectedType string
	}{
		// IaaS commands
		{"server", "iaas"},
		{"disk", "iaas"},
		{"database", "iaas"},
		{"loadbalancer", "iaas"},
		{"dns", "iaas"},
		{"gslb", "iaas"},
		{"proxylb", "iaas"},
		{"autobackup", "iaas"},
		{"archive", "iaas"},
		{"cdrom", "iaas"},
		{"bridge", "iaas"},
		{"packetfilter", "iaas"},
		{"internet", "iaas"},
		{"ipaddress", "iaas"},
		{"zone", "iaas"},
		{"region", "iaas"},
		{"bill", "iaas"},
		{"self", "iaas"},
		{"disk-plan", "iaas"},
		{"internet-plan", "iaas"},
		{"server-plan", "iaas"},

		// Misc commands
		{"config", "misc"},
		{"rest", "misc"},
		{"webaccelerator", "misc"},

		// Root commands
		{"completion", "root"},
		{"version", "root"},
		{"update-self", "root"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validator.Validate(tt.command)

			if !result.IsValid {
				t.Errorf("Expected command '%s' to be valid, got invalid: %s", tt.command, result.Message)
			}

			if result.CommandType != tt.expectedType {
				t.Errorf("Expected command '%s' to be type '%s', got '%s'",
					tt.command, tt.expectedType, result.CommandType)
			}

			if result.Command != tt.command {
				t.Errorf("Expected result command to be '%s', got '%s'", tt.command, result.Command)
			}
		})
	}
}

func TestValidateInvalidCommands(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command       string
		expectedError string
	}{
		{"", "empty_command"},
		{"invalid-command", "unknown_command"},
		{"nonexistent", "unknown_command"},
		{"xyz", "unknown_command"},
		{"servr", "unknown_command"}, // typo
		{"dsk", "unknown_command"},   // typo
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validator.Validate(tt.command)

			if result.IsValid {
				t.Errorf("Expected command '%s' to be invalid, got valid", tt.command)
			}

			if result.ErrorType != tt.expectedError {
				t.Errorf("Expected error type '%s', got '%s'", tt.expectedError, result.ErrorType)
			}

			if result.Message == "" {
				t.Error("Expected error message to be non-empty")
			}
		})
	}
}

func TestValidateDeprecatedCommands(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command         string
		expectedError   string
		hasReplacement  bool
		expectedReplace string
	}{
		{"iso-image", "deprecated_command", true, "cdrom"},
		{"startup-script", "deprecated_command", true, "note"},
		{"ipv4", "deprecated_command", true, "ipaddress"},
		{"product-disk", "deprecated_command", true, "disk-plan"},
		{"product-internet", "deprecated_command", true, "internet-plan"},
		{"product-server", "deprecated_command", true, "server-plan"},
		{"summary", "discontinued_command", false, ""},
		{"object-storage", "discontinued_command", false, ""},
		{"ojs", "discontinued_command", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validator.Validate(tt.command)

			if result.IsValid {
				t.Errorf("Expected deprecated command '%s' to be invalid, got valid", tt.command)
			}

			if result.ErrorType != tt.expectedError {
				t.Errorf("Expected error type '%s', got '%s'", tt.expectedError, result.ErrorType)
			}

			if tt.hasReplacement {
				if len(result.Suggestions) == 0 {
					t.Errorf("Expected suggestions for deprecated command '%s'", tt.command)
				} else if result.Suggestions[0] != tt.expectedReplace {
					t.Errorf("Expected replacement '%s', got '%s'",
						tt.expectedReplace, result.Suggestions[0])
				}
			}
		})
	}
}

func TestValidateCaseSensitivity(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command string
		valid   bool
		hasCase bool
	}{
		{"SERVER", true, true},
		{"Server", true, true},
		{"server", true, false},
		{"DISK", true, true},
		{"CONFIG", true, true},
		{"VERSION", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validator.Validate(tt.command)

			if result.IsValid != tt.valid {
				t.Errorf("Expected command '%s' valid=%t, got %t", tt.command, tt.valid, result.IsValid)
			}

			if tt.hasCase && result.Message == "" {
				t.Errorf("Expected case sensitivity message for '%s'", tt.command)
			}

			if !tt.hasCase && strings.Contains(result.Message, "小文字") {
				t.Errorf("Unexpected case sensitivity message for '%s': %s", tt.command, result.Message)
			}
		})
	}
}

func TestValidateCommandLine(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		name        string
		cmdLine     *CommandLine
		expectValid bool
		expectError string
	}{
		{
			name: "Valid IaaS command with subcommand",
			cmdLine: &CommandLine{
				MainCommand: "server",
				SubCommand:  "list",
			},
			expectValid: true,
		},
		{
			name: "Valid standalone command without subcommand",
			cmdLine: &CommandLine{
				MainCommand: "version",
				SubCommand:  "",
			},
			expectValid: true,
		},
		{
			name: "Invalid standalone command with subcommand",
			cmdLine: &CommandLine{
				MainCommand: "version",
				SubCommand:  "list",
			},
			expectValid: false,
			expectError: "unexpected_subcommand",
		},
		{
			name: "Invalid standalone command with subcommand (update-self)",
			cmdLine: &CommandLine{
				MainCommand: "update-self",
				SubCommand:  "show",
			},
			expectValid: false,
			expectError: "unexpected_subcommand",
		},
		{
			name: "Empty main command",
			cmdLine: &CommandLine{
				MainCommand: "",
				SubCommand:  "list",
			},
			expectValid: false,
			expectError: "empty_command",
		},
		{
			name: "Unknown command",
			cmdLine: &CommandLine{
				MainCommand: "unknown",
				SubCommand:  "list",
			},
			expectValid: false,
			expectError: "unknown_command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateCommandLine(tt.cmdLine)

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected valid=%t, got %t: %s", tt.expectValid, result.IsValid, result.Message)
			}

			if !tt.expectValid && result.ErrorType != tt.expectError {
				t.Errorf("Expected error type '%s', got '%s'", tt.expectError, result.ErrorType)
			}
		})
	}
}

func TestIsValidCommand(t *testing.T) {
	validator := NewMainCommandValidator()

	validCommands := []string{"server", "disk", "config", "version", "SERVER", "Disk"}
	invalidCommands := []string{"", "invalid", "nonexistent"}

	for _, cmd := range validCommands {
		if !validator.IsValidCommand(cmd) {
			t.Errorf("Expected command '%s' to be valid", cmd)
		}
	}

	for _, cmd := range invalidCommands {
		if validator.IsValidCommand(cmd) {
			t.Errorf("Expected command '%s' to be invalid", cmd)
		}
	}
}

func TestGetCommandType(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command      string
		expectedType string
	}{
		{"server", "iaas"},
		{"config", "misc"},
		{"version", "root"},
		{"SERVER", "iaas"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		result := validator.GetCommandType(tt.command)
		if result != tt.expectedType {
			t.Errorf("Expected type '%s' for command '%s', got '%s'",
				tt.expectedType, tt.command, result)
		}
	}
}

func TestGetAllCommands(t *testing.T) {
	validator := NewMainCommandValidator()
	commands := validator.GetAllCommands()

	// Check that all types are present
	expectedTypes := []string{"iaas", "misc", "root"}
	for _, cmdType := range expectedTypes {
		if _, exists := commands[cmdType]; !exists {
			t.Errorf("Expected command type '%s' to be present", cmdType)
		}

		if len(commands[cmdType]) == 0 {
			t.Errorf("Expected command type '%s' to have commands", cmdType)
		}
	}

	// Check specific commands
	iaasCommands := commands["iaas"]
	found := false
	for _, cmd := range iaasCommands {
		if cmd == "server" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'server' to be in IaaS commands")
	}
}

func TestGetCommandCount(t *testing.T) {
	validator := NewMainCommandValidator()
	count := validator.GetCommandCount()

	// Should have at least 50 commands (47 IaaS + 3 misc + 3 root)
	if count < 50 {
		t.Errorf("Expected at least 50 commands, got %d", count)
	}

	// Should not be too many (sanity check)
	if count > 100 {
		t.Errorf("Expected less than 100 commands, got %d", count)
	}
}

func TestGetSimilarCommands(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		input          string
		maxSuggestions int
		expectSome     bool
	}{
		{"serv", 3, true}, // Should match "server"
		{"dis", 2, true},  // Should match "disk"
		{"xyz", 2, false}, // Should not match anything
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			suggestions := validator.getSimilarCommands(tt.input, tt.maxSuggestions)

			if tt.expectSome && len(suggestions) == 0 {
				t.Errorf("Expected suggestions for input '%s', got none", tt.input)
			}

			if len(suggestions) > tt.maxSuggestions {
				t.Errorf("Expected at most %d suggestions, got %d", tt.maxSuggestions, len(suggestions))
			}
		})
	}
}

func TestIsStandaloneCommand(t *testing.T) {
	validator := NewMainCommandValidator()

	tests := []struct {
		command    string
		standalone bool
	}{
		{"version", true},
		{"update-self", true},
		{"VERSION", true},
		{"server", false},
		{"config", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := validator.IsStandaloneCommand(tt.command)
		if result != tt.standalone {
			t.Errorf("Expected IsStandaloneCommand('%s') = %t, got %t",
				tt.command, tt.standalone, result)
		}
	}
}

func TestCommandInitialization(t *testing.T) {
	validator := NewMainCommandValidator()

	// Test that all expected command categories are properly initialized
	allCommands := validator.GetAllCommands()

	// Check IaaS commands include key examples
	iaasCommands := allCommands["iaas"]
	expectedIaaS := []string{"server", "disk", "database", "cdrom", "disk-plan", "internet-plan", "server-plan"}
	for _, expected := range expectedIaaS {
		found := false
		for _, cmd := range iaasCommands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected IaaS command '%s' not found", expected)
		}
	}

	// Check misc commands
	miscCommands := allCommands["misc"]
	expectedMisc := []string{"config", "rest", "webaccelerator"}
	for _, expected := range expectedMisc {
		found := false
		for _, cmd := range miscCommands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected misc command '%s' not found", expected)
		}
	}

	// Check root commands
	rootCommands := allCommands["root"]
	expectedRoot := []string{"completion", "version", "update-self"}
	for _, expected := range expectedRoot {
		found := false
		for _, cmd := range rootCommands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected root command '%s' not found", expected)
		}
	}
}

func TestValidationResultFields(t *testing.T) {
	validator := NewMainCommandValidator()

	// Test valid command
	result := validator.Validate("server")
	if result.Command != "server" {
		t.Errorf("Expected command field to be 'server', got '%s'", result.Command)
	}
	if result.CommandType != "iaas" {
		t.Errorf("Expected command type to be 'iaas', got '%s'", result.CommandType)
	}
	if result.ErrorType != "" {
		t.Errorf("Expected empty error type for valid command, got '%s'", result.ErrorType)
	}

	// Test invalid command
	result = validator.Validate("invalid")
	if result.Command != "invalid" {
		t.Errorf("Expected command field to be 'invalid', got '%s'", result.Command)
	}
	if result.ErrorType != "unknown_command" {
		t.Errorf("Expected error type 'unknown_command', got '%s'", result.ErrorType)
	}
	if result.Message == "" {
		t.Error("Expected non-empty message for invalid command")
	}
}
