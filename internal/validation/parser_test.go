package validation

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("NewParser() returned nil")
	}
}

func TestIsUsacloudCommand(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input    string
		expected bool
	}{
		{"usacloud server list", true},
		{"usacloud", true},
		{"usacloud --help", true},
		{"  usacloud server list  ", true},
		{"not-usacloud command", false},
		{"aws s3 ls", false},
		{"", false},
		{"   ", false},
		{"usacloud-update", false}, // doesn't start with "usacloud "
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parser.IsUsacloudCommand(tt.input)
			if result != tt.expected {
				t.Errorf("IsUsacloudCommand(%q): expected %t, got %t",
					tt.input, tt.expected, result)
			}
		})
	}
}

func TestParseBasicCommands(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input           string
		expectedMain    string
		expectedSub     string
		expectedArgs    []string
		expectedOptions map[string]string
		expectedFlags   []string
	}{
		{
			input:           "usacloud server list",
			expectedMain:    "server",
			expectedSub:     "list",
			expectedArgs:    []string{},
			expectedOptions: map[string]string{},
			expectedFlags:   []string{},
		},
		{
			input:           "usacloud disk create",
			expectedMain:    "disk",
			expectedSub:     "create",
			expectedArgs:    []string{},
			expectedOptions: map[string]string{},
			expectedFlags:   []string{},
		},
		{
			input:           "usacloud config show",
			expectedMain:    "config",
			expectedSub:     "show",
			expectedArgs:    []string{},
			expectedOptions: map[string]string{},
			expectedFlags:   []string{},
		},
		{
			input:           "usacloud version",
			expectedMain:    "version",
			expectedSub:     "",
			expectedArgs:    []string{},
			expectedOptions: map[string]string{},
			expectedFlags:   []string{},
		},
		{
			input:           "usacloud",
			expectedMain:    "",
			expectedSub:     "",
			expectedArgs:    []string{},
			expectedOptions: map[string]string{},
			expectedFlags:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}

			if result.MainCommand != tt.expectedMain {
				t.Errorf("MainCommand: expected %q, got %q", tt.expectedMain, result.MainCommand)
			}

			if result.SubCommand != tt.expectedSub {
				t.Errorf("SubCommand: expected %q, got %q", tt.expectedSub, result.SubCommand)
			}

			if len(result.Arguments) != len(tt.expectedArgs) {
				t.Errorf("Arguments length: expected %d, got %d", len(tt.expectedArgs), len(result.Arguments))
			} else {
				for i, arg := range tt.expectedArgs {
					if result.Arguments[i] != arg {
						t.Errorf("Arguments[%d]: expected %q, got %q", i, arg, result.Arguments[i])
					}
				}
			}

			if len(result.Options) != len(tt.expectedOptions) {
				t.Errorf("Options length: expected %d, got %d", len(tt.expectedOptions), len(result.Options))
			} else {
				for key, value := range tt.expectedOptions {
					if result.Options[key] != value {
						t.Errorf("Options[%s]: expected %q, got %q", key, value, result.Options[key])
					}
				}
			}

			if len(result.Flags) != len(tt.expectedFlags) {
				t.Errorf("Flags length: expected %d, got %d", len(tt.expectedFlags), len(result.Flags))
			} else {
				for i, flag := range tt.expectedFlags {
					if result.Flags[i] != flag {
						t.Errorf("Flags[%d]: expected %q, got %q", i, flag, result.Flags[i])
					}
				}
			}
		})
	}
}

func TestParseComplexCommands(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input           string
		expectedMain    string
		expectedSub     string
		expectedArgs    []string
		expectedOptions map[string]string
		expectedFlags   []string
	}{
		{
			input:        "usacloud server create --name test --cpu 2 --memory 4",
			expectedMain: "server",
			expectedSub:  "create",
			expectedArgs: []string{},
			expectedOptions: map[string]string{
				"name":   "test",
				"cpu":    "2",
				"memory": "4",
			},
			expectedFlags: []string{},
		},
		{
			input:        "usacloud disk connect --server-id 123456789 disk-id",
			expectedMain: "disk",
			expectedSub:  "connect",
			expectedArgs: []string{"disk-id"},
			expectedOptions: map[string]string{
				"server-id": "123456789",
			},
			expectedFlags: []string{},
		},
		{
			input:        "usacloud server ssh --user root --key ~/.ssh/id_rsa --force server-name",
			expectedMain: "server",
			expectedSub:  "ssh",
			expectedArgs: []string{"server-name"},
			expectedOptions: map[string]string{
				"user": "root",
				"key":  "~/.ssh/id_rsa",
			},
			expectedFlags: []string{"force"},
		},
		{
			input:        "usacloud server list --zone=tk1a --output-type=json",
			expectedMain: "server",
			expectedSub:  "list",
			expectedArgs: []string{},
			expectedOptions: map[string]string{
				"zone":        "tk1a",
				"output-type": "json",
			},
			expectedFlags: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}

			if result.MainCommand != tt.expectedMain {
				t.Errorf("MainCommand: expected %q, got %q", tt.expectedMain, result.MainCommand)
			}

			if result.SubCommand != tt.expectedSub {
				t.Errorf("SubCommand: expected %q, got %q", tt.expectedSub, result.SubCommand)
			}

			if len(result.Arguments) != len(tt.expectedArgs) {
				t.Errorf("Arguments: expected %v, got %v", tt.expectedArgs, result.Arguments)
			}

			if len(result.Options) != len(tt.expectedOptions) {
				t.Errorf("Options: expected %v, got %v", tt.expectedOptions, result.Options)
			}

			for key, expectedValue := range tt.expectedOptions {
				if actualValue, exists := result.Options[key]; !exists || actualValue != expectedValue {
					t.Errorf("Option %s: expected %q, got %q (exists: %t)", key, expectedValue, actualValue, exists)
				}
			}

			if len(result.Flags) != len(tt.expectedFlags) {
				t.Errorf("Flags: expected %v, got %v", tt.expectedFlags, result.Flags)
			}
		})
	}
}

func TestParseWithQuotes(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input           string
		expectedMain    string
		expectedSub     string
		expectedArgs    []string
		expectedOptions map[string]string
	}{
		{
			input:        `usacloud server create --name "test server" --description 'A test server'`,
			expectedMain: "server",
			expectedSub:  "create",
			expectedArgs: []string{},
			expectedOptions: map[string]string{
				"name":        "test server",
				"description": "A test server",
			},
		},
		{
			input:        `usacloud server ssh --command "ls -la" server-name`,
			expectedMain: "server",
			expectedSub:  "ssh",
			expectedArgs: []string{"server-name"},
			expectedOptions: map[string]string{
				"command": "ls -la",
			},
		},
		{
			input:        `usacloud server create --name="quoted name" regular-arg`,
			expectedMain: "server",
			expectedSub:  "create",
			expectedArgs: []string{"regular-arg"},
			expectedOptions: map[string]string{
				"name": "quoted name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}

			if result.MainCommand != tt.expectedMain {
				t.Errorf("MainCommand: expected %q, got %q", tt.expectedMain, result.MainCommand)
			}

			if result.SubCommand != tt.expectedSub {
				t.Errorf("SubCommand: expected %q, got %q", tt.expectedSub, result.SubCommand)
			}

			if len(result.Arguments) != len(tt.expectedArgs) {
				t.Errorf("Arguments: expected %v, got %v", tt.expectedArgs, result.Arguments)
			}

			for key, expectedValue := range tt.expectedOptions {
				if actualValue, exists := result.Options[key]; !exists || actualValue != expectedValue {
					t.Errorf("Option %s: expected %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input       string
		expectedErr error
	}{
		{"", ErrEmptyCommand},
		{"   ", ErrEmptyCommand},
		{"aws s3 ls", ErrNotUsacloudCommand},
		{"not-a-command", ErrNotUsacloudCommand},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err == nil {
				t.Errorf("Parse(%q) should have returned error %v, but got result: %v", tt.input, tt.expectedErr, result)
			}
			if err != tt.expectedErr {
				t.Errorf("Parse(%q) error: expected %v, got %v", tt.input, tt.expectedErr, err)
			}
		})
	}
}

func TestParseErrorUnclosedQuote(t *testing.T) {
	parser := NewParser()

	tests := []string{
		`usacloud server create --name "unclosed quote`,
		`usacloud server create --name 'unclosed quote`,
		`usacloud server create --name "mixed quotes'`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result, err := parser.Parse(input)
			if err == nil {
				t.Errorf("Parse(%q) should have returned error for unclosed quote, but got result: %v", input, result)
			}

			if parseErr, ok := err.(*ParseError); ok {
				if !strings.Contains(parseErr.Message, "unclosed quote") {
					t.Errorf("Parse(%q) error should mention unclosed quote, got: %v", input, parseErr.Message)
				}
			} else {
				t.Errorf("Parse(%q) should return ParseError for unclosed quote, got: %T", input, err)
			}
		})
	}
}

func TestCommandLineGetCommandType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input        string
		expectedType string
	}{
		{"usacloud server list", "iaas"},
		{"usacloud config show", "misc"},
		{"usacloud version", "root"},
		{"usacloud iso-image list", "deprecated"},
		{"usacloud invalid-command", "unknown"},
		{"usacloud", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}

			cmdType := result.GetCommandType()
			if cmdType != tt.expectedType {
				t.Errorf("GetCommandType() for %q: expected %s, got %s", tt.input, tt.expectedType, cmdType)
			}
		})
	}
}

func TestCommandLineIsValid(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input    string
		expected bool
	}{
		{"usacloud server list", true},
		{"usacloud disk create", true},
		{"usacloud config show", true},
		{"usacloud version", true},
		{"usacloud completion bash", true},
		{"usacloud iso-image list", false},  // deprecated
		{"usacloud invalid-command", false}, // unknown
		{"usacloud server", false},          // missing subcommand for IaaS
		{"usacloud config", false},          // missing subcommand for misc
		{"usacloud", false},                 // no command
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}

			isValid := result.IsValid()
			if isValid != tt.expected {
				t.Errorf("IsValid() for %q: expected %t, got %t", tt.input, tt.expected, isValid)
			}
		})
	}
}

func TestCommandLineHelperMethods(t *testing.T) {
	parser := NewParser()

	input := "usacloud server create --name test --cpu 2 --force --dry-run server-id"
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}

	// Test HasOption
	if !result.HasOption("name") {
		t.Error("HasOption('name') should return true")
	}
	if result.HasOption("nonexistent") {
		t.Error("HasOption('nonexistent') should return false")
	}

	// Test GetOption
	if result.GetOption("name") != "test" {
		t.Errorf("GetOption('name'): expected 'test', got %q", result.GetOption("name"))
	}
	if result.GetOption("nonexistent") != "" {
		t.Errorf("GetOption('nonexistent'): expected '', got %q", result.GetOption("nonexistent"))
	}

	// Test HasFlag
	if !result.HasFlag("force") {
		t.Error("HasFlag('force') should return true")
	}
	if result.HasFlag("nonexistent") {
		t.Error("HasFlag('nonexistent') should return false")
	}

	// Test GetArgument
	if result.GetArgument(0) != "server-id" {
		t.Errorf("GetArgument(0): expected 'server-id', got %q", result.GetArgument(0))
	}
	if result.GetArgument(1) != "" {
		t.Errorf("GetArgument(1): expected '', got %q", result.GetArgument(1))
	}
	if result.GetArgument(-1) != "" {
		t.Errorf("GetArgument(-1): expected '', got %q", result.GetArgument(-1))
	}
}

func TestCommandLineString(t *testing.T) {
	parser := NewParser()

	input := "usacloud server create --name test --force server-id"
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}

	str := result.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}

	// Should contain the main components
	expectedParts := []string{"MainCommand: server", "SubCommand: create", "Arguments:", "Options:", "Flags:"}
	for _, part := range expectedParts {
		if !strings.Contains(str, part) {
			t.Errorf("String() should contain %q, got: %s", part, str)
		}
	}
}

func TestTokenizeEdgeCases(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		input    string
		expected []string
	}{
		{"usacloud  server   list", []string{"usacloud", "server", "list"}},
		{"usacloud\tserver\tlist", []string{"usacloud", "server", "list"}},
		{"usacloud server list   ", []string{"usacloud", "server", "list"}},
		{"   usacloud server list", []string{"usacloud", "server", "list"}},
		{"usacloud server create --name=\"\"", []string{"usacloud", "server", "create", "--name="}},
		{`usacloud server create --description "line1\nline2"`, []string{"usacloud", "server", "create", "--description", "line1\nline2"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens, err := parser.tokenize(tt.input)
			if err != nil {
				t.Fatalf("tokenize(%q) returned error: %v", tt.input, err)
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("tokenize(%q): expected %v, got %v", tt.input, tt.expected, tokens)
			} else {
				for i, expected := range tt.expected {
					if tokens[i] != expected {
						t.Errorf("tokenize(%q)[%d]: expected %q, got %q", tt.input, i, expected, tokens[i])
					}
				}
			}
		})
	}
}

func TestParsePerformance(t *testing.T) {
	parser := NewParser()

	// Test with a reasonably complex command
	input := "usacloud server create --name test --cpu 2 --memory 4 --zone tk1a --disk-plan ssd --os-type centos --password mypassword --ssh-key ~/.ssh/id_rsa --network-interface eth0 --tag env=test --tag project=demo --force --dry-run server-arg1 server-arg2"

	// Run parsing multiple times to check for performance issues
	for i := 0; i < 100; i++ {
		result, err := parser.Parse(input)
		if err != nil {
			t.Fatalf("Parse iteration %d failed: %v", i, err)
		}

		if result.MainCommand != "server" || result.SubCommand != "create" {
			t.Errorf("Parse iteration %d: unexpected result: %v", i, result)
		}
	}
}
