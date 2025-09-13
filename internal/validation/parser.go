// Package validation provides command validation functionality for usacloud-update
package validation

import (
	"errors"
	"fmt"
	"strings"
)

// CommandLine represents parsed command line information
type CommandLine struct {
	Raw         string            // Original command line
	MainCommand string            // Main command (server, disk, config, etc.)
	SubCommand  string            // Subcommand (list, create, show, etc.)
	Arguments   []string          // Positional arguments
	Options     map[string]string // Options (--name=value)
	Flags       []string          // Flags (--force, --dry-run, etc.)
}

// ParseError represents a parsing error
type ParseError struct {
	Message  string
	Position int
	Input    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at position %d: %s in '%s'", e.Position, e.Message, e.Input)
}

// Parser represents a command line parser
type Parser struct {
	// Configuration and cache can be added here later
}

// Common parsing errors
var (
	ErrEmptyCommand       = errors.New("empty command")
	ErrNotUsacloudCommand = errors.New("not a usacloud command")
	ErrInvalidSyntax      = errors.New("invalid command syntax")
	ErrMissingArgument    = errors.New("missing required argument")
)

// NewParser creates a new command line parser
func NewParser() *Parser {
	return &Parser{}
}

// IsUsacloudCommand checks if the command line starts with usacloud
func (p *Parser) IsUsacloudCommand(commandLine string) bool {
	trimmed := strings.TrimSpace(commandLine)
	return strings.HasPrefix(trimmed, "usacloud ") || trimmed == "usacloud"
}

// Parse parses a command line string into CommandLine struct
func (p *Parser) Parse(commandLine string) (*CommandLine, error) {
	if commandLine == "" {
		return nil, ErrEmptyCommand
	}

	trimmed := strings.TrimSpace(commandLine)
	if trimmed == "" {
		return nil, ErrEmptyCommand
	}

	if !p.IsUsacloudCommand(trimmed) {
		return nil, ErrNotUsacloudCommand
	}

	// Initialize result
	result := &CommandLine{
		Raw:       commandLine,
		Arguments: []string{},
		Options:   make(map[string]string),
		Flags:     []string{},
	}

	// Tokenize the command line
	tokens, err := p.tokenize(trimmed)
	if err != nil {
		return nil, err
	}

	// Skip "usacloud" token
	if len(tokens) == 0 || tokens[0] != "usacloud" {
		return nil, &ParseError{
			Message:  "command must start with usacloud",
			Position: 0,
			Input:    trimmed,
		}
	}
	tokens = tokens[1:]

	if len(tokens) == 0 {
		// Just "usacloud" with no arguments
		return result, nil
	}

	// Parse main command
	result.MainCommand = tokens[0]
	tokens = tokens[1:]

	// Parse subcommand if it doesn't start with --
	if len(tokens) > 0 && !strings.HasPrefix(tokens[0], "--") {
		result.SubCommand = tokens[0]
		tokens = tokens[1:]
	}

	// Parse remaining tokens (options, flags, arguments)
	err = p.parseOptionsAndArguments(tokens, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// tokenize splits command line into tokens, respecting quotes
func (p *Parser) tokenize(commandLine string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar byte

	for i := 0; i < len(commandLine); i++ {
		char := commandLine[i]

		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteByte(char)
			}
		case char == ' ' || char == '\t':
			if inQuotes {
				current.WriteByte(char)
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			}
		case char == '\\' && i+1 < len(commandLine):
			// Handle escape sequences
			next := commandLine[i+1]
			switch next {
			case 'n':
				current.WriteByte('\n')
			case 't':
				current.WriteByte('\t')
			case 'r':
				current.WriteByte('\r')
			case '\\':
				current.WriteByte('\\')
			case '"', '\'':
				current.WriteByte(next)
			default:
				current.WriteByte(char)
				current.WriteByte(next)
			}
			i++ // Skip next character
		default:
			current.WriteByte(char)
		}
	}

	if inQuotes {
		return nil, &ParseError{
			Message:  "unclosed quote",
			Position: len(commandLine) - 1,
			Input:    commandLine,
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// parseOptionsAndArguments parses options, flags, and positional arguments
func (p *Parser) parseOptionsAndArguments(tokens []string, result *CommandLine) error {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if strings.HasPrefix(token, "--") {
			// This is an option or flag
			err := p.parseOption(token, tokens, &i, result)
			if err != nil {
				return err
			}
		} else {
			// This is a positional argument
			result.Arguments = append(result.Arguments, token)
		}
	}

	return nil
}

// parseOption parses a single option or flag
func (p *Parser) parseOption(token string, tokens []string, index *int, result *CommandLine) error {
	optionName := token[2:] // Remove "--" prefix

	if optionName == "" {
		return &ParseError{
			Message:  "empty option name",
			Position: 0,
			Input:    token,
		}
	}

	// Check if it's in --key=value format
	if equalIndex := strings.Index(optionName, "="); equalIndex != -1 {
		key := optionName[:equalIndex]
		value := optionName[equalIndex+1:]

		if key == "" {
			return &ParseError{
				Message:  "empty option key",
				Position: 0,
				Input:    token,
			}
		}

		result.Options[key] = value
		return nil
	}

	// List of known flags that don't take values
	knownFlags := map[string]bool{
		"force":    true,
		"dry-run":  true,
		"verbose":  true,
		"quiet":    true,
		"help":     true,
		"version":  true,
		"no-color": true,
		"debug":    true,
	}

	// If it's a known flag, treat it as a flag
	if knownFlags[optionName] {
		result.Flags = append(result.Flags, optionName)
		return nil
	}

	// Check if next token is a value (doesn't start with --)
	if *index+1 < len(tokens) && !strings.HasPrefix(tokens[*index+1], "--") {
		// --key value format
		value := tokens[*index+1]
		result.Options[optionName] = value
		*index++ // Skip the value token
	} else {
		// This is a flag (boolean option)
		result.Flags = append(result.Flags, optionName)
	}

	return nil
}

// GetCommandType returns the type of command (iaas, misc, root, deprecated, unknown)
func (c *CommandLine) GetCommandType() string {
	if c.MainCommand == "" {
		return "unknown"
	}

	// Check if it's a deprecated command
	if IsDeprecatedCommand(c.MainCommand) {
		return "deprecated"
	}

	// Check if it's a root command
	if IsValidRootCommand(c.MainCommand) {
		return "root"
	}

	// Check if it's a misc command
	if IsValidMiscCommand(c.MainCommand) {
		return "misc"
	}

	// Check if it's an IaaS command
	if IsValidIaaSCommand(c.MainCommand) {
		return "iaas"
	}

	return "unknown"
}

// IsValid checks if the parsed command is valid according to the dictionaries
func (c *CommandLine) IsValid() bool {
	commandType := c.GetCommandType()

	switch commandType {
	case "deprecated":
		return false // Deprecated commands are not valid
	case "unknown":
		return false // Unknown commands are not valid
	case "root":
		return c.isValidRootCommand()
	case "misc":
		return c.isValidMiscCommand()
	case "iaas":
		return c.isValidIaaSCommand()
	default:
		return false
	}
}

// isValidRootCommand validates root commands
func (c *CommandLine) isValidRootCommand() bool {
	// Validate using root command validation
	errorMsg := ValidateRootCommandUsage(c.MainCommand, c.SubCommand)
	return errorMsg == ""
}

// isValidMiscCommand validates miscellaneous commands
func (c *CommandLine) isValidMiscCommand() bool {
	if c.SubCommand == "" {
		return false // Misc commands require subcommands
	}
	return IsValidMiscSubcommand(c.MainCommand, c.SubCommand)
}

// isValidIaaSCommand validates IaaS commands
func (c *CommandLine) isValidIaaSCommand() bool {
	if c.SubCommand == "" {
		return false // IaaS commands require subcommands
	}
	return IsValidIaaSSubcommand(c.MainCommand, c.SubCommand)
}

// String returns a string representation of the CommandLine
func (c *CommandLine) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("MainCommand: %s", c.MainCommand))

	if c.SubCommand != "" {
		parts = append(parts, fmt.Sprintf("SubCommand: %s", c.SubCommand))
	}

	if len(c.Arguments) > 0 {
		parts = append(parts, fmt.Sprintf("Arguments: %v", c.Arguments))
	}

	if len(c.Options) > 0 {
		parts = append(parts, fmt.Sprintf("Options: %v", c.Options))
	}

	if len(c.Flags) > 0 {
		parts = append(parts, fmt.Sprintf("Flags: %v", c.Flags))
	}

	return fmt.Sprintf("CommandLine{%s}", strings.Join(parts, ", "))
}

// HasOption checks if a specific option exists
func (c *CommandLine) HasOption(key string) bool {
	_, exists := c.Options[key]
	return exists
}

// GetOption returns the value of an option, or empty string if not found
func (c *CommandLine) GetOption(key string) string {
	return c.Options[key]
}

// HasFlag checks if a specific flag exists
func (c *CommandLine) HasFlag(flag string) bool {
	for _, f := range c.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// GetArgument returns the argument at the specified index, or empty string if not found
func (c *CommandLine) GetArgument(index int) string {
	if index >= 0 && index < len(c.Arguments) {
		return c.Arguments[index]
	}
	return ""
}
