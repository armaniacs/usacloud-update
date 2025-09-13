package transform

import (
	"strings"
	"testing"
)

// TestEngine_EdgeCases tests edge cases and error conditions for the Engine
func TestEngine_EdgeCases(t *testing.T) {
	engine := NewDefaultEngine()

	testCases := []struct {
		name         string
		input        string
		shouldChange bool
		description  string
	}{
		{
			name:         "EmptyString",
			input:        "",
			shouldChange: false,
			description:  "Empty input should remain unchanged",
		},
		{
			name:         "WhitespaceOnly",
			input:        "   \t\n  ",
			shouldChange: false,
			description:  "Whitespace-only input should remain unchanged",
		},
		{
			name:         "CommentOnly",
			input:        "# This is a comment",
			shouldChange: false,
			description:  "Comment lines should not be processed",
		},
		{
			name:         "VeryLongLine",
			input:        strings.Repeat("usacloud server list --output-type=csv ", 100),
			shouldChange: true,
			description:  "Very long lines should still be processed",
		},
		{
			name:         "UnicodeCharacters",
			input:        "usacloud server list --output-type=csv ◯△□",
			shouldChange: true,
			description:  "Unicode characters should be preserved",
		},
		{
			name:         "SpecialCharacters",
			input:        "usacloud server list --output-type=csv !@#$%^&*()",
			shouldChange: true,
			description:  "Special characters should be preserved",
		},
		{
			name:         "MixedCase",
			input:        "UsaCloud SERVER list --OUTPUT-TYPE=CSV",
			shouldChange: true,
			description:  "Case-insensitive matching should work",
		},
		{
			name:         "MultipleSpaces",
			input:        "usacloud    server     list    --output-type=csv",
			shouldChange: true,
			description:  "Multiple spaces should not break parsing",
		},
		{
			name:         "TabCharacters",
			input:        "usacloud\tserver\tlist\t--output-type=csv",
			shouldChange: true,
			description:  "Tab characters should be handled",
		},
		{
			name:         "LeadingWhitespace",
			input:        "    usacloud server list --output-type=csv",
			shouldChange: true,
			description:  "Leading whitespace should be preserved",
		},
		{
			name:         "TrailingWhitespace",
			input:        "usacloud server list --output-type=csv    ",
			shouldChange: true,
			description:  "Trailing whitespace should be preserved",
		},
		{
			name:         "NewlineCharacters",
			input:        "usacloud server list --output-type=csv\n",
			shouldChange: true,
			description:  "Newline characters should be handled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if result.Changed != tc.shouldChange {
				t.Errorf("Expected changed=%v, got %v for input: %q", tc.shouldChange, result.Changed, tc.input)
			}

			if result.Line == "" && tc.input != "" {
				t.Error("Output line should not be empty for non-empty input")
			}

			if tc.shouldChange {
				if len(result.Changes) == 0 {
					t.Error("Changes slice should not be empty when changed=true")
				}

				// Verify the rule name is set
				for _, change := range result.Changes {
					if change.RuleName == "" {
						t.Error("Rule name should not be empty for changes")
					}
					if change.Before == "" || change.After == "" {
						t.Error("Before and After should not be empty for changes")
					}
				}
			} else {
				if len(result.Changes) > 0 {
					t.Error("Changes slice should be empty when changed=false")
				}
			}
		})
	}
}

// TestEngine_ErrorConditions tests error conditions and malformed input
func TestEngine_ErrorConditions(t *testing.T) {
	engine := NewDefaultEngine()

	testCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "MalformedCommand",
			input:       "usacloud --output-type=csv server list", // parameters before subcommand
			description: "Malformed commands should not crash",
		},
		{
			name:        "OnlyDashes",
			input:       "----------",
			description: "Input with only dashes should not crash",
		},
		{
			name:        "OnlyEquals",
			input:       "==========",
			description: "Input with only equals should not crash",
		},
		{
			name:        "BinaryData",
			input:       string([]byte{0, 1, 2, 3, 255, 254, 253}),
			description: "Binary data should not crash the engine",
		},
		{
			name:        "VeryLongWord",
			input:       "usacloud " + strings.Repeat("a", 10000),
			description: "Very long words should not crash",
		},
		{
			name:        "ManyRepeatedPatterns",
			input:       strings.Repeat("--output-type=csv ", 1000),
			description: "Many repeated patterns should not crash",
		},
		{
			name:        "NestedQuotes",
			input:       `usacloud server list --output-type="csv \"nested\" quotes"`,
			description: "Nested quotes should be handled",
		},
		{
			name:        "EscapedCharacters",
			input:       "usacloud server list --output-type=csv\\n\\t\\r",
			description: "Escaped characters should be handled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The main test is that this doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Engine panicked on input %q: %v", tc.input, r)
				}
			}()

			result := engine.Apply(tc.input)

			// Basic sanity checks
			if result.Line == "" && tc.input != "" {
				t.Errorf("Result line should not be empty for non-empty input: %q", tc.input)
			}

			// Verify result structure is valid
			if result.Changed && len(result.Changes) == 0 {
				t.Error("If changed=true, changes slice should not be empty")
			}

			if !result.Changed && len(result.Changes) > 0 {
				t.Error("If changed=false, changes slice should be empty")
			}
		})
	}
}

// TestEngine_ConcurrentAccess tests concurrent access to the engine
func TestEngine_ConcurrentAccess(t *testing.T) {
	engine := NewDefaultEngine()

	// Test concurrent access to the same engine
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for j := 0; j < numOperations; j++ {
				input := "usacloud server list --output-type=csv"
				result := engine.Apply(input)

				if !result.Changed {
					t.Errorf("Goroutine %d operation %d: expected change", id, j)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestEngine_MultipleRuleApplication tests complex scenarios with multiple rule applications
func TestEngine_MultipleRuleApplication(t *testing.T) {
	engine := NewDefaultEngine()

	testCases := []struct {
		name          string
		input         string
		expectedRules int
		description   string
	}{
		{
			name:          "MultipleTransformations",
			input:         "usacloud startup-script list --output-type=csv --selector name=test",
			expectedRules: 3, // startup-script->note, csv->json, selector removal
			description:   "Input that matches multiple rules should apply all",
		},
		{
			name:          "OverlappingRules",
			input:         "usacloud product-disk list --zone = all",
			expectedRules: 2, // product-disk->disk-plan, zone normalization
			description:   "Rules with overlapping patterns should work",
		},
		{
			name:          "AllRuleTypes",
			input:         "usacloud ipv4 list --output-type=tsv --zone = all",
			expectedRules: 3, // ipv4->ipaddress, tsv->json, zone normalize
			description:   "Input matching various rule categories",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if !result.Changed {
				t.Error("Expected input to be changed")
			}

			if len(result.Changes) != tc.expectedRules {
				t.Errorf("Expected %d rule applications, got %d", tc.expectedRules, len(result.Changes))
			}

			// Verify each change has valid metadata
			for i, change := range result.Changes {
				if change.RuleName == "" {
					t.Errorf("Change %d: rule name should not be empty", i)
				}
				if change.Before == "" {
					t.Errorf("Change %d: before fragment should not be empty", i)
				}
				if change.After == "" {
					t.Errorf("Change %d: after fragment should not be empty", i)
				}
			}
		})
	}
}

// TestEngine_RuleOrderDependency tests that rule application order produces consistent results
func TestEngine_RuleOrderDependency(t *testing.T) {
	// Create multiple engines and apply them multiple times to ensure consistency
	engines := []*Engine{
		NewDefaultEngine(),
		NewDefaultEngine(),
		NewDefaultEngine(),
	}

	testInputs := []string{
		"usacloud server list --output-type=csv",
		"usacloud startup-script list --zone = all",
		"usacloud product-disk read --selector id=123",
		"usacloud summary --verbose",
	}

	for _, input := range testInputs {
		var firstResult *Result

		for i, engine := range engines {
			result := engine.Apply(input)

			if i == 0 {
				firstResult = &result
			} else {
				// All engines should produce identical results
				if result.Changed != firstResult.Changed {
					t.Errorf("Engine %d produced different Changed result for %q", i, input)
				}

				if result.Line != firstResult.Line {
					t.Errorf("Engine %d produced different Line result for %q", i, input)
				}

				if len(result.Changes) != len(firstResult.Changes) {
					t.Errorf("Engine %d produced different number of changes for %q", i, input)
				}
			}
		}
	}
}
