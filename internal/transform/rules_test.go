package transform

import (
	"regexp"
	"strings"
	"testing"
)

func TestSimpleRuleCreation(t *testing.T) {
	rule := mk(
		"test-rule",
		`test-(\w+)`,
		func(m []string) string { return "replaced-" + m[1] },
		"test reason",
		"https://example.com",
	)

	if rule.Name() != "test-rule" {
		t.Errorf("Expected rule name 'test-rule', got '%s'", rule.Name())
	}
}

func TestSimpleRuleApplyMatch(t *testing.T) {
	rule := mk(
		"test-rule",
		`test-(\w+)`,
		func(m []string) string { return "replaced-" + m[1] },
		"test reason",
		"https://example.com",
	)

	line, changed, before, after := rule.Apply("This is test-hello world")

	if !changed {
		t.Error("Rule should have matched and changed the line")
	}

	expectedLine := "This is replaced-hello world # usacloud-update: test reason (https://example.com)"
	if line != expectedLine {
		t.Errorf("Expected line '%s', got '%s'", expectedLine, line)
	}

	if before != "test-hello" {
		t.Errorf("Expected before fragment 'test-hello', got '%s'", before)
	}

	if after != "replaced-hello" {
		t.Errorf("Expected after fragment 'replaced-hello', got '%s'", after)
	}
}

func TestSimpleRuleApplyNoMatch(t *testing.T) {
	rule := mk(
		"test-rule",
		`test-(\w+)`,
		func(m []string) string { return "replaced-" + m[1] },
		"test reason",
		"https://example.com",
	)

	originalLine := "This has no match"
	line, changed, before, after := rule.Apply(originalLine)

	if changed {
		t.Error("Rule should not have matched")
	}

	if line != originalLine {
		t.Errorf("Line should be unchanged, got '%s'", line)
	}

	if before != "" || after != "" {
		t.Error("Before and after fragments should be empty when no match")
	}
}

func TestSimpleRuleCommentNotDuplicated(t *testing.T) {
	rule := mk(
		"test-rule",
		`test-(\w+)`,
		func(m []string) string { return "replaced-" + m[1] },
		"test reason",
		"https://example.com",
	)

	// Line already has usacloud-update comment
	inputLine := "This is test-hello world # usacloud-update: existing comment"
	line, changed, _, _ := rule.Apply(inputLine)

	if !changed {
		t.Error("Rule should have matched")
	}

	// Should not add duplicate comment
	if strings.Count(line, "# usacloud-update:") != 1 {
		t.Errorf("Should have exactly one usacloud-update comment, got: '%s'", line)
	}
}

func TestSimpleRuleRegexPatterns(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		input       string
		shouldMatch bool
	}{
		{
			name:        "Case insensitive match",
			pattern:     `(?i)usacloud`,
			input:       "USACLOUD server list",
			shouldMatch: true,
		},
		{
			name:        "Word boundary match",
			pattern:     `\busacloud\b`,
			input:       "usacloud server list",
			shouldMatch: true,
		},
		{
			name:        "Word boundary no match",
			pattern:     `\busacloud\b`,
			input:       "notusacloud server list", // No dash, truly no word boundary
			shouldMatch: false,
		},
		{
			name:        "Complex pattern",
			pattern:     `usacloud\s+\w+\s+list\s+--output-type=csv`,
			input:       "usacloud server list --output-type=csv",
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.pattern)
			matches := re.MatchString(tc.input)

			if matches != tc.shouldMatch {
				t.Errorf("Pattern '%s' with input '%s' - expected match: %v, got: %v",
					tc.pattern, tc.input, tc.shouldMatch, matches)
			}
		})
	}
}

func TestSimpleRuleReplacementFunction(t *testing.T) {
	testCases := []struct {
		name     string
		pattern  string
		input    string
		replFunc func([]string) string
		expected string
	}{
		{
			name:     "Simple replacement",
			pattern:  `old-(\w+)`,
			input:    "old-value",
			replFunc: func(m []string) string { return "new-" + m[1] },
			expected: "new-value",
		},
		{
			name:     "Multiple captures",
			pattern:  `(\w+)-(\w+)-(\w+)`,
			input:    "one-two-three",
			replFunc: func(m []string) string { return m[3] + "-" + m[2] + "-" + m[1] },
			expected: "three-two-one",
		},
		{
			name:     "Case transformation",
			pattern:  `test-(\w+)`,
			input:    "test-hello",
			replFunc: func(m []string) string { return "TEST-" + strings.ToUpper(m[1]) },
			expected: "TEST-HELLO",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := mk(
				"test-rule",
				tc.pattern,
				tc.replFunc,
				"test reason",
				"https://example.com",
			)

			line, changed, _, after := rule.Apply(tc.input)

			if !changed {
				t.Errorf("Rule should have matched input '%s'", tc.input)
				return
			}

			if !strings.Contains(line, tc.expected) {
				t.Errorf("Expected line to contain '%s', got '%s'", tc.expected, line)
			}

			if after != tc.expected {
				t.Errorf("Expected after fragment '%s', got '%s'", tc.expected, after)
			}
		})
	}
}

func TestSimpleRuleNameAndMetadata(t *testing.T) {
	rule := mk(
		"output-type-csv-tsv",
		`--output-type=csv`,
		func(m []string) string { return "--output-type=json" },
		"CSV is deprecated in v1.0",
		"https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	)

	if rule.Name() != "output-type-csv-tsv" {
		t.Errorf("Expected rule name 'output-type-csv-tsv', got '%s'", rule.Name())
	}

	line, changed, _, _ := rule.Apply("usacloud server list --output-type=csv")

	if !changed {
		t.Error("Rule should have matched")
	}

	if !strings.Contains(line, "CSV is deprecated in v1.0") {
		t.Error("Comment should contain the reason")
	}

	if !strings.Contains(line, "https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/") {
		t.Error("Comment should contain the URL")
	}
}

func TestSimpleRuleEdgeCases(t *testing.T) {
	rule := mk(
		"test-rule",
		`test-(\w+)`,
		func(m []string) string { return "replaced-" + m[1] },
		"test reason",
		"https://example.com",
	)

	testCases := []struct {
		name        string
		input       string
		shouldMatch bool
	}{
		{"Empty string", "", false},
		{"Only whitespace", "   ", false},
		{"Match at start", "test-hello world", true},
		{"Match at end", "world test-hello", true},
		{"Match in middle", "start test-hello end", true},
		{"Multiple matches", "test-one and test-two", true}, // Should match first occurrence
		{"No word after dash", "test-", false},
		{"Numbers in word", "test-hello123", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, changed, _, _ := rule.Apply(tc.input)

			if changed != tc.shouldMatch {
				t.Errorf("Input '%s' - expected match: %v, got: %v", tc.input, tc.shouldMatch, changed)
			}
		})
	}
}
