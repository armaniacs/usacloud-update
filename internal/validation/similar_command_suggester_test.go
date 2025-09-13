package validation

import (
	"strings"
	"testing"
)

func TestNewSimilarCommandSuggester(t *testing.T) {
	suggester := NewSimilarCommandSuggester(3, 5)
	if suggester == nil {
		t.Error("Expected suggester to be created, got nil")
	}
	if suggester.maxDistance != 3 {
		t.Errorf("Expected maxDistance to be 3, got %d", suggester.maxDistance)
	}
	if suggester.maxSuggestions != 5 {
		t.Errorf("Expected maxSuggestions to be 5, got %d", suggester.maxSuggestions)
	}
}

func TestNewDefaultSimilarCommandSuggester(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()
	if suggester == nil {
		t.Error("Expected suggester to be created, got nil")
	}
	if suggester.maxDistance != DefaultMaxDistance {
		t.Errorf("Expected maxDistance to be %d, got %d", DefaultMaxDistance, suggester.maxDistance)
	}
	if suggester.maxSuggestions != DefaultMaxSuggestions {
		t.Errorf("Expected maxSuggestions to be %d, got %d", DefaultMaxSuggestions, suggester.maxSuggestions)
	}
}

func TestLevenshteinDistance(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "def", 3},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"server", "sever", 1},
		{"list", "lst", 1},
		{"create", "creat", 1},
		{"Server", "server", 0},     // Case insensitive
		{"DATABASE", "database", 0}, // Case insensitive
		{"disk", "disc", 1},
		{"switch", "swich", 1},
	}

	for _, tt := range tests {
		result := suggester.LevenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestGetAdaptiveMaxDistance(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		input    string
		expected int
	}{
		{"a", 1},
		{"ab", 1},
		{"abc", 1},
		{"abcd", 2},
		{"abcde", 2},
		{"abcdef", 2},
		{"abcdefg", 3},
		{"very-long-command", 3},
	}

	for _, tt := range tests {
		result := suggester.getAdaptiveMaxDistance(tt.input)
		if result != tt.expected {
			t.Errorf("getAdaptiveMaxDistance(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateScore(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		input     string
		candidate string
		distance  int
		minScore  float64
	}{
		{"server", "server", 0, 1.0}, // Exact match
		{"server", "sever", 1, 0.8},  // Close match
		{"list", "lst", 1, 0.5},      // Acceptable match
		{"create", "creat", 1, 0.8},  // Close match
		{"abc", "xyz", 3, 0.0},       // Poor match
	}

	for _, tt := range tests {
		result := suggester.calculateScore(tt.input, tt.candidate, tt.distance)
		if result < tt.minScore {
			t.Errorf("calculateScore(%q, %q, %d) = %.3f, expected >= %.3f",
				tt.input, tt.candidate, tt.distance, result, tt.minScore)
		}
	}
}

func TestGetTypoScore(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		input     string
		candidate string
		expected  float64
	}{
		{"sever", "server", 0.2},  // Common typo
		{"disc", "disk", 0.2},     // Common typo
		{"srv", "server", 0.2},    // Common abbreviation
		{"lst", "list", 0.2},      // Common typo
		{"creat", "create", 0.2},  // Common typo
		{"xyz", "server", 0.0},    // No pattern match
		{"server", "server", 0.0}, // Exact match (no bonus)
	}

	for _, tt := range tests {
		result := suggester.getTypoScore(tt.input, tt.candidate)
		if result != tt.expected {
			t.Errorf("getTypoScore(%q, %q) = %.3f, expected %.3f",
				tt.input, tt.candidate, result, tt.expected)
		}
	}
}

func TestFilterByPrefix(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()
	candidates := []string{"server", "disk", "swytch", "database", "snapshot", "service"}

	tests := []struct {
		input    string
		expected []string
	}{
		{"a", candidates},                     // Too short, return all
		{"se", []string{"server", "service"}}, // Prefix match
		{"di", []string{"disk"}},              // Prefix match
		{"xyz", candidates},                   // No match, return all
		{"sw", []string{"swytch"}},            // Single match
	}

	for _, tt := range tests {
		result := suggester.filterByPrefix(tt.input, candidates)
		if len(result) != len(tt.expected) {
			t.Errorf("filterByPrefix(%q) returned %d candidates, expected %d",
				tt.input, len(result), len(tt.expected))
			continue
		}

		// Check if all expected candidates are present
		for _, expected := range tt.expected {
			found := false
			for _, actual := range result {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("filterByPrefix(%q) missing expected candidate %q", tt.input, expected)
			}
		}
	}
}

func TestSuggestMainCommands(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		input             string
		expectSuggestions bool
		shouldContain     []string
		shouldNotContain  []string
	}{
		{
			input:             "sever",
			expectSuggestions: true,
			shouldContain:     []string{"server"},
		},
		{
			input:             "lst",
			expectSuggestions: false, // lst is not close enough to any main command
		},
		{
			input:             "servr",
			expectSuggestions: true,
			shouldContain:     []string{"server"},
		},
		{
			input:             "disc",
			expectSuggestions: true,
			shouldContain:     []string{"disk"},
		},
		{
			input:             "databse",
			expectSuggestions: true,
			shouldContain:     []string{"database"},
		},
		{
			input:             "",
			expectSuggestions: false, // Empty input
		},
		{
			input:             "xyz",
			expectSuggestions: false, // No similar commands
		},
	}

	for _, tt := range tests {
		results := suggester.SuggestMainCommands(tt.input)

		if tt.expectSuggestions && len(results) == 0 {
			t.Errorf("SuggestMainCommands(%q) expected suggestions but got none", tt.input)
			continue
		}

		if !tt.expectSuggestions && len(results) > 0 {
			t.Errorf("SuggestMainCommands(%q) expected no suggestions but got %d", tt.input, len(results))
			continue
		}

		// Check for expected suggestions
		for _, expected := range tt.shouldContain {
			found := false
			for _, result := range results {
				if result.Command == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SuggestMainCommands(%q) should contain %q", tt.input, expected)
			}
		}

		// Check that unwanted suggestions are not present
		for _, unwanted := range tt.shouldNotContain {
			for _, result := range results {
				if result.Command == unwanted {
					t.Errorf("SuggestMainCommands(%q) should not contain %q", tt.input, unwanted)
				}
			}
		}

		// Check score ordering (should be descending)
		for i := 1; i < len(results); i++ {
			if results[i-1].Score < results[i].Score {
				t.Errorf("SuggestMainCommands(%q) results not properly sorted by score", tt.input)
				break
			}
		}

		// Check suggestion limit
		if len(results) > suggester.maxSuggestions {
			t.Errorf("SuggestMainCommands(%q) returned %d suggestions, max is %d",
				tt.input, len(results), suggester.maxSuggestions)
		}
	}
}

func TestSuggestSubcommands(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	tests := []struct {
		mainCommand       string
		input             string
		expectSuggestions bool
		shouldContain     []string
	}{
		{
			mainCommand:       "server",
			input:             "lst",
			expectSuggestions: true,
			shouldContain:     []string{"list"},
		},
		{
			mainCommand:       "server",
			input:             "creat",
			expectSuggestions: true,
			shouldContain:     []string{"create"},
		},
		{
			mainCommand:       "server",
			input:             "delet",
			expectSuggestions: true,
			shouldContain:     []string{"delete"},
		},
		{
			mainCommand:       "nonexistent",
			input:             "list",
			expectSuggestions: false, // Unknown main command
		},
		{
			mainCommand:       "server",
			input:             "",
			expectSuggestions: false, // Empty input
		},
		{
			mainCommand:       "",
			input:             "list",
			expectSuggestions: false, // Empty main command
		},
		{
			mainCommand:       "disk",
			input:             "connct",
			expectSuggestions: true,
			shouldContain:     []string{"connect"},
		},
	}

	for _, tt := range tests {
		results := suggester.SuggestSubcommands(tt.mainCommand, tt.input)

		if tt.expectSuggestions && len(results) == 0 {
			t.Errorf("SuggestSubcommands(%q, %q) expected suggestions but got none",
				tt.mainCommand, tt.input)
			continue
		}

		if !tt.expectSuggestions && len(results) > 0 {
			t.Errorf("SuggestSubcommands(%q, %q) expected no suggestions but got %d",
				tt.mainCommand, tt.input, len(results))
			continue
		}

		// Check for expected suggestions
		for _, expected := range tt.shouldContain {
			found := false
			for _, result := range results {
				if result.Command == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SuggestSubcommands(%q, %q) should contain %q",
					tt.mainCommand, tt.input, expected)
			}
		}

		// Check score ordering (should be descending)
		for i := 1; i < len(results); i++ {
			if results[i-1].Score < results[i].Score {
				t.Errorf("SuggestSubcommands(%q, %q) results not properly sorted by score",
					tt.mainCommand, tt.input)
				break
			}
		}

		// Check suggestion limit
		if len(results) > suggester.maxSuggestions {
			t.Errorf("SuggestSubcommands(%q, %q) returned %d suggestions, max is %d",
				tt.mainCommand, tt.input, len(results), suggester.maxSuggestions)
		}
	}
}

func TestSuggesterGetAllCommands(t *testing.T) {
	commands := getAllCommands()

	if len(commands) == 0 {
		t.Error("getAllCommands() should return non-empty list")
	}

	// Check for some expected commands
	expectedCommands := []string{"server", "disk", "database", "swytch"}
	for _, expected := range expectedCommands {
		found := false
		for _, command := range commands {
			if command == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getAllCommands() should contain %q", expected)
		}
	}
}

func TestGetAllCommandSubcommands(t *testing.T) {
	subcommands := getAllCommandSubcommands()

	if len(subcommands) == 0 {
		t.Error("getAllCommandSubcommands() should return non-empty map")
	}

	// Check for some expected command-subcommand mappings
	expectedMappings := map[string][]string{
		"server": {"list", "create", "delete"},
		"disk":   {"list", "create", "delete"},
	}

	for command, expectedSubs := range expectedMappings {
		actualSubs, exists := subcommands[command]
		if !exists {
			t.Errorf("getAllCommandSubcommands() should contain command %q", command)
			continue
		}

		for _, expectedSub := range expectedSubs {
			found := false
			for _, actualSub := range actualSubs {
				if actualSub == expectedSub {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("getAllCommandSubcommands()[%q] should contain %q", command, expectedSub)
			}
		}
	}
}

func TestSimilarityResultFields(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()
	results := suggester.SuggestMainCommands("sever")

	if len(results) == 0 {
		t.Error("Expected at least one result for 'sever'")
		return
	}

	result := results[0]
	if result.Command == "" {
		t.Error("SimilarityResult.Command should not be empty")
	}
	if result.Distance < 0 {
		t.Error("SimilarityResult.Distance should not be negative")
	}
	if result.Score < 0.0 || result.Score > 1.0 {
		t.Errorf("SimilarityResult.Score should be between 0.0 and 1.0, got %.3f", result.Score)
	}
}

func TestCommonTypoPatterns(t *testing.T) {
	// Test that CommonTypoPatterns is properly defined
	if len(CommonTypoPatterns) == 0 {
		t.Error("CommonTypoPatterns should not be empty")
	}

	// Check some expected patterns
	expectedPatterns := map[string][]string{
		"server": {"sever", "serv"},
		"disk":   {"disc"},
		"list":   {"lst"},
		"create": {"creat"},
	}

	for command, expectedTypos := range expectedPatterns {
		actualTypos, exists := CommonTypoPatterns[command]
		if !exists {
			t.Errorf("CommonTypoPatterns should contain patterns for %q", command)
			continue
		}

		for _, expectedTypo := range expectedTypos {
			found := false
			for _, actualTypo := range actualTypos {
				if actualTypo == expectedTypo {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("CommonTypoPatterns[%q] should contain %q", command, expectedTypo)
			}
		}
	}
}

func TestPerformance(t *testing.T) {
	// Basic performance test - should complete within reasonable time
	suggester := NewDefaultSimilarCommandSuggester()

	// Test with various inputs
	testInputs := []string{
		"sever", "lst", "creat", "delet", "databse", "swich", "archiv", "snap",
		"xyz", "abcdef", "very-long-command-that-does-not-exist",
	}

	for _, input := range testInputs {
		// Test main command suggestions
		results := suggester.SuggestMainCommands(input)

		// Just verify it completes and returns reasonable results
		if len(results) > 10 {
			t.Errorf("SuggestMainCommands(%q) returned too many results: %d", input, len(results))
		}

		// Test subcommand suggestions for some commands
		if len(results) > 0 {
			subResults := suggester.SuggestSubcommands("server", input)
			if len(subResults) > 10 {
				t.Errorf("SuggestSubcommands('server', %q) returned too many results: %d", input, len(subResults))
			}
		}
	}
}

func TestEdgeCases(t *testing.T) {
	suggester := NewDefaultSimilarCommandSuggester()

	// Test very short inputs
	_ = suggester.SuggestMainCommands("a")
	// Should not crash and may return some results

	// Test very long inputs
	longInput := strings.Repeat("abcdefghijk", 10)
	_ = suggester.SuggestMainCommands(longInput)
	// Should not crash and should return empty or very few results

	// Test special characters
	_ = suggester.SuggestMainCommands("server-test")
	// Should handle hyphens gracefully

	// Test numbers
	_ = suggester.SuggestMainCommands("server123")
	// Should handle numbers gracefully

	// All tests should complete without panicking
	t.Log("Edge case tests completed successfully")
}

func TestSuggesterMinMaxHelpers(t *testing.T) {
	// Test suggesterMin function
	if suggesterMin(5, 3) != 3 {
		t.Error("suggesterMin(5, 3) should return 3")
	}
	if suggesterMin(3, 5) != 3 {
		t.Error("suggesterMin(3, 5) should return 3")
	}
	if suggesterMin(5, 5) != 5 {
		t.Error("suggesterMin(5, 5) should return 5")
	}

	// Test suggesterMax function
	if suggesterMax(5, 3) != 5 {
		t.Error("suggesterMax(5, 3) should return 5")
	}
	if suggesterMax(3, 5) != 5 {
		t.Error("suggesterMax(3, 5) should return 5")
	}
	if suggesterMax(5, 5) != 5 {
		t.Error("suggesterMax(5, 5) should return 5")
	}
}
