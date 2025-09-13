package validation

import (
	"testing"
)

func TestNewSubcommandValidator(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	if validator == nil {
		t.Error("Expected validator to be created, got nil")
	}

	if validator.mainValidator == nil {
		t.Error("Expected main validator to be set")
	}

	// Check that some subcommands were initialized
	serverSubs := validator.GetAvailableSubcommands("server")
	if len(serverSubs) == 0 {
		t.Error("Expected server subcommands to be initialized")
	}
}

func TestValidateValidSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		mainCommand string
		subCommand  string
	}{
		// IaaS commands
		{"server", "list"},
		{"server", "create"},
		{"server", "boot"},
		{"server", "shutdown"},
		{"disk", "list"},
		{"disk", "create"},
		{"disk", "connect"},
		{"disk", "disconnect"},

		// Misc commands
		{"config", "list"},
		{"config", "show"},
		{"config", "use"},
		{"rest", "get"},
		{"rest", "post"},
		{"rest", "put"},
		{"rest", "delete"},
		{"webaccelerator", "purge"},

		// Root commands
		{"completion", "bash"},
		{"completion", "zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.mainCommand+"_"+tt.subCommand, func(t *testing.T) {
			result := validator.Validate(tt.mainCommand, tt.subCommand)

			if !result.IsValid {
				t.Errorf("Expected subcommand '%s %s' to be valid, got invalid: %s",
					tt.mainCommand, tt.subCommand, result.Message)
			}

			if result.MainCommand != tt.mainCommand {
				t.Errorf("Expected main command to be '%s', got '%s'",
					tt.mainCommand, result.MainCommand)
			}

			if result.SubCommand != tt.subCommand {
				t.Errorf("Expected sub command to be '%s', got '%s'",
					tt.subCommand, result.SubCommand)
			}
		})
	}
}

func TestValidateInvalidSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		mainCommand   string
		subCommand    string
		expectedError string
	}{
		{"server", "invalid", ErrorTypeInvalidSubcommand},
		{"disk", "nonexistent", ErrorTypeInvalidSubcommand},
		{"config", "invalid", ErrorTypeInvalidSubcommand},
		{"rest", "invalid", ErrorTypeInvalidSubcommand},
		{"webaccelerator", "invalid", ErrorTypeInvalidSubcommand},
		{"completion", "invalid", ErrorTypeInvalidSubcommand},
	}

	for _, tt := range tests {
		t.Run(tt.mainCommand+"_"+tt.subCommand, func(t *testing.T) {
			result := validator.Validate(tt.mainCommand, tt.subCommand)

			if result.IsValid {
				t.Errorf("Expected subcommand '%s %s' to be invalid, got valid",
					tt.mainCommand, tt.subCommand)
			}

			if result.ErrorType != tt.expectedError {
				t.Errorf("Expected error type '%s', got '%s'",
					tt.expectedError, result.ErrorType)
			}

			if result.Message == "" {
				t.Error("Expected error message to be non-empty")
			}

			if len(result.Available) == 0 {
				t.Error("Expected available subcommands to be provided")
			}
		})
	}
}

func TestValidateStandaloneCommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		name        string
		mainCommand string
		subCommand  string
		expectValid bool
		expectError string
	}{
		{
			name:        "version without subcommand",
			mainCommand: "version",
			subCommand:  "",
			expectValid: true,
		},
		{
			name:        "update-self without subcommand",
			mainCommand: "update-self",
			subCommand:  "",
			expectValid: true,
		},
		{
			name:        "version with invalid subcommand",
			mainCommand: "version",
			subCommand:  "list",
			expectValid: false,
			expectError: ErrorTypeUnexpectedSubcommand,
		},
		{
			name:        "update-self with invalid subcommand",
			mainCommand: "update-self",
			subCommand:  "run",
			expectValid: false,
			expectError: ErrorTypeUnexpectedSubcommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(tt.mainCommand, tt.subCommand)

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected valid=%t, got %t: %s", tt.expectValid, result.IsValid, result.Message)
			}

			if !tt.expectValid && result.ErrorType != tt.expectError {
				t.Errorf("Expected error type '%s', got '%s'", tt.expectError, result.ErrorType)
			}
		})
	}
}

func TestValidateMissingSubcommand(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []string{"server", "disk", "config", "rest", "webaccelerator", "completion"}

	for _, mainCommand := range tests {
		t.Run(mainCommand, func(t *testing.T) {
			result := validator.Validate(mainCommand, "")

			if result.IsValid {
				t.Errorf("Expected command '%s' without subcommand to be invalid", mainCommand)
			}

			if result.ErrorType != ErrorTypeMissingSubcommand {
				t.Errorf("Expected error type '%s', got '%s'",
					ErrorTypeMissingSubcommand, result.ErrorType)
			}

			if len(result.Suggestions) == 0 {
				t.Error("Expected suggestions to be provided for missing subcommand")
			}

			if len(result.Available) == 0 {
				t.Error("Expected available subcommands to be provided")
			}
		})
	}
}

func TestValidateInvalidMainCommand(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	result := validator.Validate("invalid", "list")

	if result.IsValid {
		t.Error("Expected validation to fail for invalid main command")
	}

	if result.ErrorType != "invalid_main_command" {
		t.Errorf("Expected error type 'invalid_main_command', got '%s'", result.ErrorType)
	}
}

func TestSubcommandValidateCommandLine(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		name        string
		cmdLine     *CommandLine
		expectValid bool
		expectError string
	}{
		{
			name: "Valid server list",
			cmdLine: &CommandLine{
				MainCommand: "server",
				SubCommand:  "list",
			},
			expectValid: true,
		},
		{
			name: "Valid disk connect",
			cmdLine: &CommandLine{
				MainCommand: "disk",
				SubCommand:  "connect",
			},
			expectValid: true,
		},
		{
			name: "Invalid server subcommand",
			cmdLine: &CommandLine{
				MainCommand: "server",
				SubCommand:  "invalid",
			},
			expectValid: false,
			expectError: ErrorTypeInvalidSubcommand,
		},
		{
			name: "Version with subcommand",
			cmdLine: &CommandLine{
				MainCommand: "version",
				SubCommand:  "list",
			},
			expectValid: false,
			expectError: ErrorTypeUnexpectedSubcommand,
		},
		{
			name: "Missing subcommand",
			cmdLine: &CommandLine{
				MainCommand: "server",
				SubCommand:  "",
			},
			expectValid: false,
			expectError: ErrorTypeMissingSubcommand,
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

func TestIsValidSubcommand(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	validCases := [][]string{
		{"server", "list"},
		{"disk", "connect"},
		{"config", "show"},
		{"rest", "get"},
	}

	invalidCases := [][]string{
		{"server", "invalid"},
		{"disk", "invalid"},
		{"config", "invalid"},
		{"invalid", "list"},
	}

	for _, test := range validCases {
		if !validator.IsValidSubcommand(test[0], test[1]) {
			t.Errorf("Expected '%s %s' to be valid", test[0], test[1])
		}
	}

	for _, test := range invalidCases {
		if validator.IsValidSubcommand(test[0], test[1]) {
			t.Errorf("Expected '%s %s' to be invalid", test[0], test[1])
		}
	}
}

func TestGetAvailableSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		command     string
		expectEmpty bool
		expectSome  []string // subcommands that should be present
	}{
		{"server", false, []string{"list", "create", "boot", "shutdown"}},
		{"disk", false, []string{"list", "create", "connect", "disconnect"}},
		{"config", false, []string{"list", "show", "use"}},
		{"rest", false, []string{"get", "post", "put", "delete"}},
		{"webaccelerator", false, []string{"purge"}},
		{"completion", false, []string{"bash", "zsh"}},
		{"version", true, nil},     // standalone command
		{"update-self", true, nil}, // standalone command
		{"nonexistent", true, nil}, // invalid command
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			subcommands := validator.GetAvailableSubcommands(tt.command)

			if tt.expectEmpty && len(subcommands) != 0 {
				t.Errorf("Expected no subcommands for '%s', got %v", tt.command, subcommands)
			}

			if !tt.expectEmpty && len(subcommands) == 0 {
				t.Errorf("Expected subcommands for '%s', got none", tt.command)
			}

			// Check for specific expected subcommands
			if tt.expectSome != nil {
				for _, expected := range tt.expectSome {
					found := false
					for _, sub := range subcommands {
						if sub == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected subcommand '%s' for command '%s', not found in %v",
							expected, tt.command, subcommands)
					}
				}
			}
		})
	}
}

func TestSubcommandIsStandaloneCommand(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		command    string
		standalone bool
	}{
		{"version", true},
		{"update-self", true},
		{"VERSION", true}, // case insensitive
		{"server", false},
		{"disk", false},
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

func TestGetCommandSubcommandCount(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	count := validator.GetCommandSubcommandCount()
	if count == 0 {
		t.Error("Expected non-zero subcommand count")
	}

	// Should be reasonable number (not too high, not too low)
	if count < 50 || count > 500 {
		t.Errorf("Expected subcommand count between 50-500, got %d", count)
	}
}

func TestGetCommandsWithSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	commands := validator.GetCommandsWithSubcommands()
	if len(commands) == 0 {
		t.Error("Expected commands with subcommands")
	}

	// Check that some expected commands are present
	expectedCommands := []string{"server", "disk", "config", "rest"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' to have subcommands", expected)
		}
	}
}

func TestGetSimilarSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		mainCommand string
		subCommand  string
		expectSome  bool
	}{
		{"server", "lis", true},  // Should match "list"
		{"server", "crea", true}, // Should match "create"
		{"disk", "conn", true},   // Should match "connect"
		{"config", "xyz", false}, // Should not match anything well
	}

	for _, tt := range tests {
		t.Run(tt.mainCommand+"_"+tt.subCommand, func(t *testing.T) {
			result := validator.Validate(tt.mainCommand, tt.subCommand)

			if result.IsValid {
				t.Errorf("Expected invalid subcommand for test: %s %s", tt.mainCommand, tt.subCommand)
				return
			}

			if tt.expectSome && len(result.Suggestions) == 0 {
				t.Errorf("Expected suggestions for '%s %s', got none", tt.mainCommand, tt.subCommand)
			}

			if len(result.Suggestions) > 3 {
				t.Errorf("Expected at most 3 suggestions, got %d", len(result.Suggestions))
			}
		})
	}
}

func TestGetAllSubcommandsByCommand(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	allSubcommands := validator.GetAllSubcommandsByCommand()

	if len(allSubcommands) == 0 {
		t.Error("Expected non-empty subcommands map")
	}

	// Check that server has expected subcommands
	serverSubs, exists := allSubcommands["server"]
	if !exists {
		t.Error("Expected server to have subcommands")
	}

	expectedServerSubs := []string{"list", "create", "boot", "shutdown"}
	for _, expected := range expectedServerSubs {
		found := false
		for _, sub := range serverSubs {
			if sub == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected server subcommand '%s', not found in %v", expected, serverSubs)
		}
	}
}

func TestHasSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		command    string
		hasSubcmds bool
	}{
		{"server", true},
		{"disk", true},
		{"config", true},
		{"rest", true},
		{"version", false},     // standalone
		{"update-self", false}, // standalone
		{"invalid", false},     // doesn't exist
	}

	for _, tt := range tests {
		result := validator.HasSubcommands(tt.command)
		if result != tt.hasSubcmds {
			t.Errorf("Expected HasSubcommands('%s') = %t, got %t",
				tt.command, tt.hasSubcmds, result)
		}
	}
}

func TestCaseSensitivity(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	tests := []struct {
		mainCommand string
		subCommand  string
		expectValid bool
	}{
		{"SERVER", "LIST", true},  // Both uppercase
		{"Server", "List", true},  // Mixed case
		{"server", "list", true},  // Both lowercase
		{"DISK", "CONNECT", true}, // Both uppercase
	}

	for _, tt := range tests {
		t.Run(tt.mainCommand+"_"+tt.subCommand, func(t *testing.T) {
			result := validator.Validate(tt.mainCommand, tt.subCommand)

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected case insensitive validation to work for '%s %s', got valid=%t",
					tt.mainCommand, tt.subCommand, result.IsValid)
			}
		})
	}
}

func TestSpecificSubcommands(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	// Test disk-specific subcommands
	diskSpecific := []string{"connect", "disconnect"}
	for _, sub := range diskSpecific {
		if !validator.IsValidSubcommand("disk", sub) {
			t.Errorf("Expected disk subcommand '%s' to be valid", sub)
		}
	}

	// Test webaccelerator-specific subcommand
	if !validator.IsValidSubcommand("webaccelerator", "purge") {
		t.Error("Expected webaccelerator subcommand 'purge' to be valid")
	}

	// Test rest HTTP methods
	httpMethods := []string{"get", "post", "put", "delete"}
	for _, method := range httpMethods {
		if !validator.IsValidSubcommand("rest", method) {
			t.Errorf("Expected rest subcommand '%s' to be valid", method)
		}
	}

	// Test completion shells
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		if !validator.IsValidSubcommand("completion", shell) {
			t.Errorf("Expected completion subcommand '%s' to be valid", shell)
		}
	}
}

func TestSubcommandValidationResultFields(t *testing.T) {
	mainValidator := NewMainCommandValidator()
	validator := NewSubcommandValidator(mainValidator)

	// Test valid subcommand
	result := validator.Validate("server", "list")
	if result.MainCommand != "server" {
		t.Errorf("Expected MainCommand to be 'server', got '%s'", result.MainCommand)
	}
	if result.SubCommand != "list" {
		t.Errorf("Expected SubCommand to be 'list', got '%s'", result.SubCommand)
	}
	if result.ErrorType != "" {
		t.Errorf("Expected empty ErrorType for valid subcommand, got '%s'", result.ErrorType)
	}

	// Test invalid subcommand
	result = validator.Validate("server", "invalid")
	if result.MainCommand != "server" {
		t.Errorf("Expected MainCommand to be 'server', got '%s'", result.MainCommand)
	}
	if result.SubCommand != "invalid" {
		t.Errorf("Expected SubCommand to be 'invalid', got '%s'", result.SubCommand)
	}
	if result.ErrorType != ErrorTypeInvalidSubcommand {
		t.Errorf("Expected ErrorType '%s', got '%s'", ErrorTypeInvalidSubcommand, result.ErrorType)
	}
	if result.Message == "" {
		t.Error("Expected non-empty message for invalid subcommand")
	}
	if len(result.Available) == 0 {
		t.Error("Expected non-empty Available list")
	}
}
