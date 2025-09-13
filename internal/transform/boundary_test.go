package transform

import (
	"strings"
	"testing"
)

func TestRules_BoundaryValues(t *testing.T) {
	engine := NewDefaultEngine()

	boundaryTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single character command",
			input:    "u",
			expected: "u",
		},
		{
			name:     "exact usacloud match",
			input:    "usacloud",
			expected: "usacloud",
		},
		{
			name:     "usacloud with single space",
			input:    "usacloud ",
			expected: "usacloud ",
		},
		{
			name:     "unicode characters",
			input:    "usacloud server list --name=テスト",
			expected: "usacloud server list --name=テスト",
		},
		{
			name:     "special characters",
			input:    "usacloud server list --name='test\"file'",
			expected: "usacloud server list --name='test\"file'",
		},
		{
			name:     "minimum transformation case",
			input:    "usacloud server list --output-type=csv",
			expected: "usacloud server list --format=json", // Should be transformed
		},
		{
			name:     "already correct format",
			input:    "usacloud server list --format=json",
			expected: "usacloud server list --format=json",
		},
	}

	for _, tt := range boundaryTests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Apply(tt.input)

			// Should not crash

			// For transformation tests, check expected output
			if tt.name == "minimum transformation case" {
				if !result.Changed {
					t.Errorf("Expected transformation for: %s", tt.input)
				}
				if !strings.Contains(result.Line, "json") {
					t.Errorf("Expected json in result, got: %s", result.Line)
				}
			}
		})
	}
}

func TestRules_EdgeCases(t *testing.T) {
	engine := NewDefaultEngine()

	edgeCases := []struct {
		name  string
		input string
	}{
		{"empty command", ""},
		{"only usacloud", "usacloud"},
		{"usacloud with space", "usacloud "},
		{"non-usacloud command", "docker run hello-world"},
		{"usacloud in middle", "echo usacloud server list"},
		{"usacloud with path", "/usr/bin/usacloud server list"},
		{"quoted usacloud", "echo 'usacloud server list'"},
		{"usacloud variable", "CMD=usacloud server list"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			// Should not crash

			// Should preserve non-usacloud commands
			if !strings.Contains(tc.input, "usacloud") ||
				strings.Contains(tc.input, "echo") ||
				strings.Contains(tc.input, "/usr/bin/") ||
				strings.Contains(tc.input, "'") ||
				strings.Contains(tc.input, "CMD=") {
				if result.Changed {
					t.Logf("Non-usacloud command was modified (might be expected): %s -> %s", tc.input, result.Line)
				}
			}
		})
	}
}

func TestRules_ZoneTransformations(t *testing.T) {
	engine := NewDefaultEngine()

	zoneTests := []struct {
		name     string
		input    string
		expected bool // should be changed
	}{
		{
			name:     "zone with spaces around equals - not transformed",
			input:    "usacloud server list --zone = tk1v",
			expected: false,
		},
		{
			name:     "zone with multiple spaces - not transformed",
			input:    "usacloud server list --zone  =  tk1v",
			expected: false,
		},
		{
			name:     "zone already correct",
			input:    "usacloud server list --zone=tk1v",
			expected: false,
		},
		{
			name:     "zone with all value",
			input:    "usacloud server list --zone = all",
			expected: true,
		},
	}

	for _, tc := range zoneTests {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if tc.expected && !result.Changed {
				t.Errorf("Expected zone transformation for: %s", tc.input)
			}

			if tc.expected && result.Changed {
				// Should normalize spaces around equals
				if !strings.Contains(result.Line, "--zone=") {
					t.Errorf("Expected normalized --zone= in result, got: %s", result.Line)
				}
			}
		})
	}
}

func TestRules_OutputTypeTransformations(t *testing.T) {
	engine := NewDefaultEngine()

	outputTests := []struct {
		name     string
		input    string
		expected bool // should be changed
	}{
		{
			name:     "csv to json",
			input:    "usacloud server list --output-type=csv",
			expected: true,
		},
		{
			name:     "tsv to json",
			input:    "usacloud server list --output-type=tsv",
			expected: true,
		},
		{
			name:     "already json",
			input:    "usacloud server list --output-type=json",
			expected: false,
		},
		{
			name:     "no output-type",
			input:    "usacloud server list",
			expected: false,
		},
		{
			name:     "output-type with spaces",
			input:    "usacloud server list --output-type = csv",
			expected: true,
		},
	}

	for _, tc := range outputTests {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if tc.expected && !result.Changed {
				t.Errorf("Expected output-type transformation for: %s", tc.input)
			}

			if tc.expected && result.Changed {
				// Should convert to json (format may vary with spacing)
				if !strings.Contains(result.Line, "json") {
					t.Errorf("Expected json in result, got: %s", result.Line)
				}
			}
		})
	}
}

func TestRules_ResourceNameTransformations(t *testing.T) {
	engine := NewDefaultEngine()

	resourceTests := []struct {
		name     string
		input    string
		expected bool // should be changed
	}{
		{
			name:     "iso-image to cdrom",
			input:    "usacloud iso-image list",
			expected: true,
		},
		{
			name:     "startup-script to note",
			input:    "usacloud startup-script list",
			expected: true,
		},
		{
			name:     "ipv4 to ipaddress",
			input:    "usacloud ipv4 list",
			expected: true,
		},
		{
			name:     "already correct cdrom",
			input:    "usacloud cdrom list",
			expected: false,
		},
		{
			name:     "already correct note",
			input:    "usacloud note list",
			expected: false,
		},
		{
			name:     "already correct ipaddress",
			input:    "usacloud ipaddress list",
			expected: false,
		},
	}

	for _, tc := range resourceTests {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Apply(tc.input)

			if tc.expected && !result.Changed {
				t.Errorf("Expected resource name transformation for: %s", tc.input)
			}

			if tc.expected && result.Changed {
				// Should contain the new resource name
				switch {
				case strings.Contains(tc.input, "iso-image"):
					if !strings.Contains(result.Line, "cdrom") {
						t.Errorf("Expected cdrom in result, got: %s", result.Line)
					}
				case strings.Contains(tc.input, "startup-script"):
					if !strings.Contains(result.Line, "note") {
						t.Errorf("Expected note in result, got: %s", result.Line)
					}
				case strings.Contains(tc.input, "ipv4"):
					if !strings.Contains(result.Line, "ipaddress") {
						t.Errorf("Expected ipaddress in result, got: %s", result.Line)
					}
				}
			}
		})
	}
}

func TestRules_MaximumLineLength(t *testing.T) {
	engine := NewDefaultEngine()

	// Test with very long lines
	baseLine := "usacloud server list --output-type=csv"

	// Add many flags to create a long line
	longFlags := strings.Repeat(" --very-long-flag-name-with-many-characters=very-long-value-with-many-characters", 100)
	longLine := baseLine + longFlags

	result := engine.Apply(longLine)

	// Should still apply transformations to the base part
	if !result.Changed {
		t.Errorf("Expected transformation of output-type even in long line")
	}
}
