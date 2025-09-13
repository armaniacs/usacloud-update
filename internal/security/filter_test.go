package security

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestSensitiveDataFilter_FilterString(t *testing.T) {
	filter := NewSensitiveDataFilter()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Sakura Cloud access token",
			input:    `sakuracloud_access_token="abcdefghij1234567890"`,
			expected: `sakuracloud_access_token="[FILTERED]"`,
		},
		{
			name:     "Sakura Cloud access token with spaces",
			input:    `sakuracloud_access_token = "abcdefghij1234567890"`,
			expected: `sakuracloud_access_token = "[FILTERED]"`,
		},
		{
			name:     "Sakura Cloud secret",
			input:    `sakuracloud_access_token_secret="abcdefghij1234567890+/="`,
			expected: `sakuracloud_access_token_secret="[FILTERED]"`,
		},
		{
			name:     "Password field",
			input:    `password="mysecretpassword123"`,
			expected: `password="[FILTERED]"`,
		},
		{
			name:     "API key",
			input:    `api_key="1234567890abcdef"`,
			expected: `api_key="[FILTERED]"`,
		},
		{
			name:     "Bearer token",
			input:    `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9`,
			expected: `Authorization: Bearer [FILTERED]`,
		},
		{
			name:     "Authorization header",
			input:    `authorization: YWxhZGRpbjpvcGVuc2VzYW1l1234567890abc`,
			expected: `authorization: [FILTERED]`,
		},
		{
			name:     "Multiple patterns in one string",
			input:    `password="secret123" api_key="abcdef1234567890"`,
			expected: `password="[FILTERED]" api_key="[FILTERED]"`,
		},
		{
			name:     "No sensitive data",
			input:    `echo "Hello, World!"`,
			expected: `echo "Hello, World!"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter.FilterString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

func TestSensitiveDataFilter_FilterLogEntry(t *testing.T) {
	filter := NewSensitiveDataFilter()

	logEntry := `INFO: Connecting with sakuracloud_access_token="secrettoken12345678901" and password="mypassword12345"`
	expected := `INFO: Connecting with sakuracloud_access_token="[FILTERED]" and password="[FILTERED]"`

	result := filter.FilterLogEntry(logEntry)
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestSensitiveDataFilter_CustomPatterns(t *testing.T) {
	customPatterns := []SensitivePattern{
		{
			Name:        "custom-token",
			Pattern:     regexp.MustCompile(`(custom_token=)[a-zA-Z0-9]+`),
			Replacement: "${1}[CUSTOM_FILTERED]",
			Description: "Custom token pattern",
		},
	}

	filter := NewSensitiveDataFilterWithCustomPatterns(customPatterns)

	input := `custom_token=abc123xyz`
	expected := `custom_token=[CUSTOM_FILTERED]`

	result := filter.FilterString(input)
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}
}

func TestSensitiveDataFilter_AddRemovePattern(t *testing.T) {
	filter := NewSensitiveDataFilter()

	// Add custom pattern
	customPattern := SensitivePattern{
		Name:        "test-pattern",
		Pattern:     regexp.MustCompile(`(test_secret=)[a-zA-Z0-9]+`),
		Replacement: "${1}[TEST_FILTERED]",
		Description: "Test pattern",
	}

	filter.AddPattern(customPattern)

	input := `test_secret=mysecret123`
	expected := `test_secret=[TEST_FILTERED]`

	result := filter.FilterString(input)
	if result != expected {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}

	// Remove pattern
	filter.RemovePattern("test-pattern")

	// Should not filter anymore
	result = filter.FilterString(input)
	if result != input {
		t.Errorf("Pattern should be removed, but still filtering: %s", result)
	}
}

func TestSensitiveDataFilter_GetPatterns(t *testing.T) {
	filter := NewSensitiveDataFilter()

	patterns := filter.GetPatterns()

	// Check that default patterns are present
	expectedPatterns := []string{
		"sakura-access-token",
		"sakura-secret",
		"password-field",
		"api-key-generic",
		"bearer-token",
		"authorization-header",
	}

	patternNames := make(map[string]bool)
	for _, pattern := range patterns {
		patternNames[pattern.Name] = true
	}

	for _, expected := range expectedPatterns {
		if !patternNames[expected] {
			t.Errorf("Expected pattern %s not found", expected)
		}
	}
}

func TestSensitiveDataFilter_SetMaskOptions(t *testing.T) {
	filter := NewSensitiveDataFilter()

	// Test setting mask character
	filter.SetMaskChar('#')
	if filter.maskChar != '#' {
		t.Errorf("Expected mask char '#', got %c", filter.maskChar)
	}

	// Test setting mask length
	filter.SetMaskLength(12)
	if filter.maskLength != 12 {
		t.Errorf("Expected mask length 12, got %d", filter.maskLength)
	}
}

func TestSensitiveDataFilter_FilterBytes(t *testing.T) {
	filter := NewSensitiveDataFilter()

	input := []byte(`password="secret123"`)
	expected := []byte(`password="[FILTERED]"`)

	result := filter.FilterBytes(input)
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected: %s, Got: %s", expected, result)
	}
}

func TestSecureLogWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	filter := NewSensitiveDataFilter()
	writer := NewSecureLogWriter(&buf, filter)

	logData := []byte(`INFO: User login with password="secret123"`)
	expectedData := `INFO: User login with password="[FILTERED]"`

	n, err := writer.Write(logData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Should return original length for compatibility
	if n != len(logData) {
		t.Errorf("Expected write length %d, got %d", len(logData), n)
	}

	result := buf.String()
	if result != expectedData {
		t.Errorf("Expected: %s, Got: %s", expectedData, result)
	}
}

func TestSecureLogWriter_WithNilFilter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSecureLogWriter(&buf, nil)

	logData := []byte(`password="secret123"`)

	_, err := writer.Write(logData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Should still filter with default filter
	result := buf.String()
	if strings.Contains(result, "secret123") {
		t.Error("Expected filtering even with nil filter")
	}
}

func TestValidatePattern(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     SensitivePattern
		expectError bool
	}{
		{
			name: "Valid pattern",
			pattern: SensitivePattern{
				Name:        "test-pattern",
				Pattern:     regexp.MustCompile(`test=\w+`),
				Replacement: "test=[FILTERED]",
			},
			expectError: false,
		},
		{
			name: "Empty name",
			pattern: SensitivePattern{
				Name:        "",
				Pattern:     regexp.MustCompile(`test=\w+`),
				Replacement: "test=[FILTERED]",
			},
			expectError: true,
		},
		{
			name: "Nil pattern",
			pattern: SensitivePattern{
				Name:        "test-pattern",
				Pattern:     nil,
				Replacement: "test=[FILTERED]",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePattern(tc.pattern)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSensitiveDataFilter_CaseInsensitive(t *testing.T) {
	filter := NewSensitiveDataFilter()

	testCases := []struct {
		name  string
		input string
	}{
		{"lowercase", `password="secret123"`},
		{"uppercase", `PASSWORD="secret123"`},
		{"mixed case", `Password="secret123"`},
		{"api key lowercase", `api_key="1234567890abcdef"`},
		{"api key uppercase", `API_KEY="1234567890abcdef"`},
		{"sakura token mixed", `SakuraCloud_Access_Token="token12345678901234567890"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter.FilterString(tc.input)
			if strings.Contains(result, "secret123") || strings.Contains(result, "1234567890abcdef") || strings.Contains(result, "token12345678901234567890") {
				t.Errorf("Sensitive data not filtered in case insensitive test: %s", result)
			}
		})
	}
}

func TestSensitiveDataFilter_ComplexLogLine(t *testing.T) {
	filter := NewSensitiveDataFilter()

	input := `[2023-01-01 12:00:00] INFO: Authenticating user with credentials: sakuracloud_access_token="abc123def45678901234567890" password="mypassword12345" api_key="xyz789uvw01234567890"`

	result := filter.FilterString(input)

	// Check that all sensitive values are filtered
	sensitiveValues := []string{"abc123def45678901234567890", "mypassword12345", "xyz789uvw01234567890"}
	for _, value := range sensitiveValues {
		if strings.Contains(result, value) {
			t.Errorf("Sensitive value %s not filtered in complex log line", value)
		}
	}

	// Check that [FILTERED] appears
	if !strings.Contains(result, "[FILTERED]") {
		t.Error("Expected [FILTERED] to appear in filtered result")
	}
}

func TestSensitiveDataFilter_EdgeCases(t *testing.T) {
	filter := NewSensitiveDataFilter()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   \n\t  ",
			expected: "   \n\t  ",
		},
		{
			name:     "No sensitive data",
			input:    "This is a normal log message without secrets",
			expected: "This is a normal log message without secrets",
		},
		{
			name:     "Partial match (too short)",
			input:    `password="short"`,
			expected: `password="short"`, // Should not match if too short
		},
		{
			name:     "Multiple spaces in assignment",
			input:    `password   =   "secret123456"`,
			expected: `password   =   "[FILTERED]"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter.FilterString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}

func BenchmarkSensitiveDataFilter_FilterString(b *testing.B) {
	filter := NewSensitiveDataFilter()
	testString := `INFO: Authentication with sakuracloud_access_token="abc123def456ghi789" password="secretpassword123" api_key="xyz789uvw012abc345"`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter.FilterString(testString)
	}
}

func BenchmarkSecureLogWriter_Write(b *testing.B) {
	filter := NewSensitiveDataFilter()
	writer := NewSecureLogWriter(&bytes.Buffer{}, filter)
	logData := []byte(`INFO: Authentication with password="secret123" api_key="1234567890abcdef"`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writer.Write(logData)
	}
}

func TestSensitiveDataFilter_ThreadSafety(t *testing.T) {
	filter := NewSensitiveDataFilter()

	// Test concurrent filtering
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			input := string(rune(id)) + `password="secret123"`
			result := filter.FilterString(input)
			if strings.Contains(result, "secret123") {
				t.Errorf("Goroutine %d: sensitive data not filtered", id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
