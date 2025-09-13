package security

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// SecureInput provides secure input methods for sensitive data
type SecureInput struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// NewSecureInput creates a new secure input instance
func NewSecureInput() *SecureInput {
	return &SecureInput{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// NewSecureInputWithStreams creates a secure input with custom streams
func NewSecureInputWithStreams(stdin io.Reader, stdout, stderr io.Writer) *SecureInput {
	return &SecureInput{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

// ReadPassword reads a password with no echo
func (si *SecureInput) ReadPassword(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(si.stdout, prompt)
	}

	// Check if stdin is a terminal
	if file, ok := si.stdin.(*os.File); ok {
		fd := int(file.Fd())
		if term.IsTerminal(fd) {
			password, err := term.ReadPassword(fd)
			if err != nil {
				return "", fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Fprintln(si.stdout) // Add newline after password input
			return string(password), nil
		}
	}

	// Fallback for non-terminal input
	return si.readLineFromReader()
}

// ReadSensitiveValue reads a sensitive value with validation
func (si *SecureInput) ReadSensitiveValue(prompt, fieldName string) (string, error) {
	fullPrompt := fmt.Sprintf("%s (入力は非表示になります): ", prompt)

	value, err := si.ReadPassword(fullPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", fieldName, err)
	}

	// Basic validation
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", fieldName)
	}

	return value, nil
}

// ReadConfirmation reads a yes/no confirmation
func (si *SecureInput) ReadConfirmation(message string) (bool, error) {
	fmt.Fprintf(si.stdout, "%s (y/N): ", message)

	response, err := si.readLineFromReader()
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// ConfirmOverwrite asks for confirmation before overwriting
func (si *SecureInput) ConfirmOverwrite(resource string) (bool, error) {
	message := fmt.Sprintf("'%s' already exists. Overwrite?", resource)
	return si.ReadConfirmation(message)
}

// ReadAPIKey reads an API key with validation
func (si *SecureInput) ReadAPIKey(prompt string) (string, error) {
	apiKey, err := si.ReadSensitiveValue(prompt, "API key")
	if err != nil {
		return "", err
	}

	// Basic API key validation (length check)
	if len(apiKey) < 16 {
		return "", fmt.Errorf("API key appears to be too short (minimum 16 characters)")
	}

	return apiKey, nil
}

// ReadSecret reads a secret token with validation
func (si *SecureInput) ReadSecret(prompt string) (string, error) {
	secret, err := si.ReadSensitiveValue(prompt, "secret")
	if err != nil {
		return "", err
	}

	// Basic secret validation (length check)
	if len(secret) < 16 {
		return "", fmt.Errorf("secret appears to be too short (minimum 16 characters)")
	}

	return secret, nil
}

// ReadText reads regular text input (non-sensitive)
func (si *SecureInput) ReadText(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(si.stdout, prompt)
	}

	text, err := si.readLineFromReader()
	if err != nil {
		return "", fmt.Errorf("failed to read text: %w", err)
	}

	return strings.TrimSpace(text), nil
}

// ReadChoice reads a choice from multiple options
func (si *SecureInput) ReadChoice(prompt string, choices []string) (string, error) {
	fmt.Fprintf(si.stdout, "%s\n", prompt)
	for i, choice := range choices {
		fmt.Fprintf(si.stdout, "%d) %s\n", i+1, choice)
	}
	fmt.Fprint(si.stdout, "Enter choice (1-", len(choices), "): ")

	input, err := si.readLineFromReader()
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}

	input = strings.TrimSpace(input)

	// Try to parse as number
	for i, choice := range choices {
		if input == fmt.Sprintf("%d", i+1) {
			return choice, nil
		}
	}

	// Try to match by name (case insensitive)
	inputLower := strings.ToLower(input)
	for _, choice := range choices {
		if strings.ToLower(choice) == inputLower {
			return choice, nil
		}
	}

	return "", fmt.Errorf("invalid choice: %s", input)
}

// readLineFromReader reads a line from the configured reader
func (si *SecureInput) readLineFromReader() (string, error) {
	reader := bufio.NewReader(si.stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Remove trailing newline
	return strings.TrimSuffix(line, "\n"), nil
}

// IsTerminal checks if the input is from a terminal
func (si *SecureInput) IsTerminal() bool {
	if file, ok := si.stdin.(*os.File); ok {
		return term.IsTerminal(int(file.Fd()))
	}
	return false
}

// SecureInputConfig provides configuration for secure input
type SecureInputConfig struct {
	MinPasswordLength int
	MaxInputLength    int
	AllowEmptyInput   bool
	ConfirmSensitive  bool
}

// DefaultSecureInputConfig returns default secure input configuration
func DefaultSecureInputConfig() *SecureInputConfig {
	return &SecureInputConfig{
		MinPasswordLength: 8,
		MaxInputLength:    1024,
		AllowEmptyInput:   false,
		ConfirmSensitive:  true,
	}
}

// ValidateInput validates input according to configuration
func (config *SecureInputConfig) ValidateInput(input, fieldName string) error {
	if !config.AllowEmptyInput && strings.TrimSpace(input) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(input) > config.MaxInputLength {
		return fmt.Errorf("%s too long (maximum %d characters)", fieldName, config.MaxInputLength)
	}

	return nil
}

// ValidatePassword validates password according to configuration
func (config *SecureInputConfig) ValidatePassword(password string) error {
	if err := config.ValidateInput(password, "password"); err != nil {
		return err
	}

	if len(password) < config.MinPasswordLength {
		return fmt.Errorf("password too short (minimum %d characters)", config.MinPasswordLength)
	}

	return nil
}

// GetTerminalSize returns the terminal size if available
func GetTerminalSize() (width, height int, err error) {
	width, height, err = term.GetSize(int(syscall.Stdin))
	return
}
