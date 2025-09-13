package security

import (
	"io"
	"regexp"
)

// SensitiveDataFilter handles filtering of sensitive information from logs and output
type SensitiveDataFilter struct {
	patterns   []SensitivePattern
	maskChar   rune
	maskLength int
}

// SensitivePattern defines a pattern for sensitive data detection
type SensitivePattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
	Description string
}

// NewSensitiveDataFilter creates a new sensitive data filter with default patterns
func NewSensitiveDataFilter() *SensitiveDataFilter {
	return &SensitiveDataFilter{
		patterns: []SensitivePattern{
			{
				Name:        "sakura-access-token",
				Pattern:     regexp.MustCompile(`(?i)(sakuracloud_access_token\s*[:=]\s*["']?)[a-zA-Z0-9]{20,}(["']?)`),
				Replacement: "${1}[FILTERED]${2}",
				Description: "Sakura Cloud アクセストークン",
			},
			{
				Name:        "sakura-secret",
				Pattern:     regexp.MustCompile(`(?i)(sakuracloud_access_token_secret\s*[:=]\s*["']?)[a-zA-Z0-9+/=]{20,}(["']?)`),
				Replacement: "${1}[FILTERED]${2}",
				Description: "Sakura Cloud アクセストークンシークレット",
			},
			{
				Name:        "password-field",
				Pattern:     regexp.MustCompile(`(?i)(password\s*[:=]\s*["']?)[^\s"']{8,}(["']?)`),
				Replacement: "${1}[FILTERED]${2}",
				Description: "パスワードフィールド",
			},
			{
				Name:        "api-key-generic",
				Pattern:     regexp.MustCompile(`(?i)(api[_-]?key\s*[:=]\s*["']?)[a-zA-Z0-9]{16,}(["']?)`),
				Replacement: "${1}[FILTERED]${2}",
				Description: "汎用APIキー",
			},
			{
				Name:        "bearer-token",
				Pattern:     regexp.MustCompile(`(?i)(bearer\s+)[a-zA-Z0-9+/=]{20,}`),
				Replacement: "${1}[FILTERED]",
				Description: "Bearerトークン",
			},
			{
				Name:        "authorization-header",
				Pattern:     regexp.MustCompile(`(?i)(authorization:\s*)[a-zA-Z0-9+/=]{20,}`),
				Replacement: "${1}[FILTERED]",
				Description: "Authorizationヘッダー",
			},
		},
		maskChar:   '*',
		maskLength: 8,
	}
}

// NewSensitiveDataFilterWithCustomPatterns creates a filter with custom patterns
func NewSensitiveDataFilterWithCustomPatterns(patterns []SensitivePattern) *SensitiveDataFilter {
	return &SensitiveDataFilter{
		patterns:   patterns,
		maskChar:   '*',
		maskLength: 8,
	}
}

// FilterString filters sensitive information from a string
func (sdf *SensitiveDataFilter) FilterString(input string) string {
	filtered := input

	for _, pattern := range sdf.patterns {
		filtered = pattern.Pattern.ReplaceAllString(filtered, pattern.Replacement)
	}

	return filtered
}

// FilterLogEntry filters sensitive information from a log entry
func (sdf *SensitiveDataFilter) FilterLogEntry(entry string) string {
	return sdf.FilterString(entry)
}

// AddPattern adds a new sensitive pattern to the filter
func (sdf *SensitiveDataFilter) AddPattern(pattern SensitivePattern) {
	sdf.patterns = append(sdf.patterns, pattern)
}

// RemovePattern removes a pattern by name
func (sdf *SensitiveDataFilter) RemovePattern(name string) {
	for i, pattern := range sdf.patterns {
		if pattern.Name == name {
			sdf.patterns = append(sdf.patterns[:i], sdf.patterns[i+1:]...)
			break
		}
	}
}

// GetPatterns returns all registered patterns
func (sdf *SensitiveDataFilter) GetPatterns() []SensitivePattern {
	return sdf.patterns
}

// SetMaskChar sets the character used for masking
func (sdf *SensitiveDataFilter) SetMaskChar(char rune) {
	sdf.maskChar = char
}

// SetMaskLength sets the length of the mask
func (sdf *SensitiveDataFilter) SetMaskLength(length int) {
	sdf.maskLength = length
}

// SecureLogWriter wraps an io.Writer to filter sensitive data from logs
type SecureLogWriter struct {
	writer io.Writer
	filter *SensitiveDataFilter
}

// NewSecureLogWriter creates a new secure log writer
func NewSecureLogWriter(writer io.Writer, filter *SensitiveDataFilter) *SecureLogWriter {
	if filter == nil {
		filter = NewSensitiveDataFilter()
	}

	return &SecureLogWriter{
		writer: writer,
		filter: filter,
	}
}

// Write implements io.Writer interface with sensitive data filtering
func (slw *SecureLogWriter) Write(p []byte) (n int, err error) {
	filtered := slw.filter.FilterString(string(p))
	written, err := slw.writer.Write([]byte(filtered))

	// Return original length to maintain compatibility
	if err == nil {
		return len(p), nil
	}
	return written, err
}

// FilterBytes filters sensitive data from byte slice
func (sdf *SensitiveDataFilter) FilterBytes(data []byte) []byte {
	filtered := sdf.FilterString(string(data))
	return []byte(filtered)
}

// ValidatePattern validates a sensitive pattern
func ValidatePattern(pattern SensitivePattern) error {
	if pattern.Name == "" {
		return ErrInvalidPatternName
	}

	if pattern.Pattern == nil {
		return ErrInvalidPatternRegex
	}

	// Test the pattern to ensure it compiles and works
	testString := "test input"
	_ = pattern.Pattern.ReplaceAllString(testString, pattern.Replacement)

	return nil
}
