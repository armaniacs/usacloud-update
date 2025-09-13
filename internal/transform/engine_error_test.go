package transform

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestEngine_ErrorHandling(t *testing.T) {
	engine := NewDefaultEngine()

	// nil input handling
	result := engine.Apply("")
	_ = result // Should not crash

	// extremely long line handling
	longLine := strings.Repeat("usacloud server list ", 1000)
	result = engine.Apply(longLine)
	_ = result // Should handle gracefully

	// malformed command handling
	malformedLines := []string{
		"usacloud --invalid-syntax",
		"usacloud server list --output-type=",
		"usacloud server list --zone=",
		"usacloud server list --output-type=invalid",
		"usacloud server list --zone = invalid_zone",
		"usacloud invalid-command",
		"usacloud server invalid-subcommand",
	}

	for _, line := range malformedLines {
		result = engine.Apply(line)
		// Should not crash, even with malformed input
		_ = result
	}
}

func TestEngine_ConcurrentAccessErrors(t *testing.T) {
	engine := NewDefaultEngine()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 100 concurrent transformations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			line := fmt.Sprintf("usacloud server list --output-type=csv # test %d", id)
			result := engine.Apply(line)

			// Should not crash during concurrent access
			_ = result
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestEngine_EmptyAndWhitespaceLines(t *testing.T) {
	engine := NewDefaultEngine()

	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"single space", " "},
		{"multiple spaces", "   "},
		{"tab character", "\t"},
		{"mixed whitespace", " \t \n"},
		{"newline only", "\n"},
		{"carriage return", "\r"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)
			// Whitespace should be preserved
			if result.Line != tc.input {
				t.Errorf("Expected whitespace to be preserved: %q, got: %q", tc.input, result.Line)
			}
		})
	}
}

func TestEngine_CommentHandling(t *testing.T) {
	engine := NewDefaultEngine()

	commentTests := []struct {
		name     string
		input    string
		expected bool // should be changed
	}{
		{"comment only", "# This is a comment", false},
		{"shebang", "#!/bin/bash", false},
		{"inline comment after command", "usacloud server list # comment", false},         // Comments are not transformed
		{"command with hash in quotes", "usacloud server list --name='test#name'", false}, // Hash in quotes is not transformed
		{"multiple hashes", "usacloud server list # comment # more", false},               // Comment lines are not transformed
	}

	for _, tc := range commentTests {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if tc.expected && !result.Changed {
				t.Errorf("Expected line to be changed: %s", tc.input)
			}
			if !tc.expected && result.Changed {
				t.Errorf("Expected line to remain unchanged: %s", tc.input)
			}
		})
	}
}

func TestEngine_SpecialCharacters(t *testing.T) {
	engine := NewDefaultEngine()

	specialCharTests := []string{
		"usacloud server list --name='test\"file'",
		"usacloud server list --name=\"test'file\"",
		"usacloud server list --name=test\\file",
		"usacloud server list --name=test\nfile",
		"usacloud server list --name=test\tfile",
		"usacloud server list --name=テスト",
		"usacloud server list --name=тест",
		"usacloud server list --name=测试",
		"usacloud server list --name=🚀",
	}

	for i, input := range specialCharTests {
		t.Run(fmt.Sprintf("special_char_%d", i), func(t *testing.T) {
			_ = engine.Apply(input)
		})
	}
}

func TestEngine_LargeInputs(t *testing.T) {
	engine := NewDefaultEngine()

	largeSizes := []int{
		1000,   // 1KB
		10000,  // 10KB
		100000, // 100KB
	}

	for _, size := range largeSizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			// Create a large command with many flags
			baseCmd := "usacloud server list"
			flags := strings.Repeat(" --very-long-flag-name-that-takes-up-space", size/50)
			input := baseCmd + flags

			_ = engine.Apply(input)
		})
	}
}

func TestEngine_RecoveryFromPanic(t *testing.T) {
	// This test ensures the engine recovers gracefully from potential panics
	engine := NewDefaultEngine()

	// These inputs might cause edge cases in regex processing
	edgeCases := []string{
		strings.Repeat("(", 1000), // Unbalanced parentheses
		strings.Repeat("[", 1000), // Unbalanced brackets
		strings.Repeat("*", 1000), // Many wildcards
		strings.Repeat("\\", 500), // Many escapes
	}

	for i, input := range edgeCases {
		t.Run(fmt.Sprintf("edge_case_%d", i), func(t *testing.T) {
			// Wrap in defer to catch any panics
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Engine panicked on input: %v", r)
				}
			}()

			result := engine.Apply(input)
			// Should not panic, result can be anything
			_ = result
		})
	}
}
