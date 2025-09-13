package validation

import (
	"strings"
	"testing"
)

func TestNewDeprecatedCommandDetector(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	if detector == nil {
		t.Error("Expected detector to be created, got nil")
	}

	// Check that deprecated commands were initialized
	expectedCount := 9 // 6 renamed + 3 discontinued
	actualCount := detector.GetDeprecatedCommandCount()
	if actualCount != expectedCount {
		t.Errorf("Expected %d deprecated commands, got %d", expectedCount, actualCount)
	}
}

func TestDetectRenamedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		deprecated  string
		replacement string
	}{
		{"iso-image", "cdrom"},
		{"startup-script", "note"},
		{"ipv4", "ipaddress"},
		{"product-disk", "disk-plan"},
		{"product-internet", "internet-plan"},
		{"product-server", "server-plan"},
	}

	for _, tt := range tests {
		t.Run(tt.deprecated, func(t *testing.T) {
			info := detector.Detect(tt.deprecated)

			if info == nil {
				t.Errorf("Expected to detect deprecated command '%s'", tt.deprecated)
				return
			}

			if info.Command != tt.deprecated {
				t.Errorf("Expected command '%s', got '%s'", tt.deprecated, info.Command)
			}

			if info.ReplacementCommand != tt.replacement {
				t.Errorf("Expected replacement '%s', got '%s'", tt.replacement, info.ReplacementCommand)
			}

			if info.DeprecationType != "renamed" {
				t.Errorf("Expected deprecation type 'renamed', got '%s'", info.DeprecationType)
			}

			if info.Message == "" {
				t.Error("Expected non-empty message")
			}

			if info.DocumentationURL == "" {
				t.Error("Expected non-empty documentation URL")
			}
		})
	}
}

func TestDetectDiscontinuedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []string{"summary", "object-storage", "ojs"}

	for _, deprecated := range tests {
		t.Run(deprecated, func(t *testing.T) {
			info := detector.Detect(deprecated)

			if info == nil {
				t.Errorf("Expected to detect deprecated command '%s'", deprecated)
				return
			}

			if info.Command != deprecated {
				t.Errorf("Expected command '%s', got '%s'", deprecated, info.Command)
			}

			if info.ReplacementCommand != "" {
				t.Errorf("Expected empty replacement for discontinued command, got '%s'", info.ReplacementCommand)
			}

			if info.DeprecationType != "discontinued" {
				t.Errorf("Expected deprecation type 'discontinued', got '%s'", info.DeprecationType)
			}

			if info.Message == "" {
				t.Error("Expected non-empty message")
			}

			if len(info.AlternativeActions) == 0 {
				t.Error("Expected alternative actions for discontinued command")
			}

			if info.DocumentationURL == "" {
				t.Error("Expected non-empty documentation URL")
			}
		})
	}
}

func TestDetectNonDeprecatedCommand(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	nonDeprecatedCommands := []string{
		"server", "disk", "config", "version", "invalid", "",
	}

	for _, cmd := range nonDeprecatedCommands {
		t.Run(cmd, func(t *testing.T) {
			info := detector.Detect(cmd)
			if info != nil {
				t.Errorf("Expected command '%s' to not be deprecated", cmd)
			}
		})
	}
}

func TestIsDeprecated(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		command    string
		deprecated bool
	}{
		// Renamed commands
		{"iso-image", true},
		{"startup-script", true},
		{"ipv4", true},
		{"product-disk", true},
		{"product-internet", true},
		{"product-server", true},

		// Discontinued commands
		{"summary", true},
		{"object-storage", true},
		{"ojs", true},

		// Non-deprecated commands
		{"server", false},
		{"disk", false},
		{"config", false},
		{"cdrom", false},   // replacement command
		{"note", false},    // replacement command
		{"invalid", false}, // non-existent command
		{"", false},        // empty command
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := detector.IsDeprecated(tt.command)
			if result != tt.deprecated {
				t.Errorf("Expected IsDeprecated('%s') = %t, got %t", tt.command, tt.deprecated, result)
			}
		})
	}
}

func TestDetectorGetReplacementCommand(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		deprecated  string
		replacement string
	}{
		// Renamed commands
		{"iso-image", "cdrom"},
		{"startup-script", "note"},
		{"ipv4", "ipaddress"},
		{"product-disk", "disk-plan"},
		{"product-internet", "internet-plan"},
		{"product-server", "server-plan"},

		// Discontinued commands (no replacement)
		{"summary", ""},
		{"object-storage", ""},
		{"ojs", ""},

		// Non-deprecated commands
		{"server", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.deprecated, func(t *testing.T) {
			result := detector.GetReplacementCommand(tt.deprecated)
			if result != tt.replacement {
				t.Errorf("Expected replacement '%s' for '%s', got '%s'",
					tt.replacement, tt.deprecated, result)
			}
		})
	}
}

func TestGetDeprecationType(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		command string
		depType string
	}{
		// Renamed commands
		{"iso-image", "renamed"},
		{"startup-script", "renamed"},
		{"ipv4", "renamed"},

		// Discontinued commands
		{"summary", "discontinued"},
		{"object-storage", "discontinued"},
		{"ojs", "discontinued"},

		// Non-deprecated commands
		{"server", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := detector.GetDeprecationType(tt.command)
			if result != tt.depType {
				t.Errorf("Expected deprecation type '%s' for '%s', got '%s'",
					tt.depType, tt.command, result)
			}
		})
	}
}

func TestGetDeprecationMessage(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	deprecatedCommands := []string{
		"iso-image", "startup-script", "ipv4", "product-disk",
		"summary", "object-storage", "ojs",
	}

	for _, cmd := range deprecatedCommands {
		t.Run(cmd, func(t *testing.T) {
			message := detector.GetDeprecationMessage(cmd)
			if message == "" {
				t.Errorf("Expected non-empty message for deprecated command '%s'", cmd)
			}
			if !strings.Contains(message, "廃止") {
				t.Errorf("Expected message for '%s' to contain '廃止', got: %s", cmd, message)
			}
		})
	}

	// Test non-deprecated commands
	nonDeprecated := []string{"server", "invalid", ""}
	for _, cmd := range nonDeprecated {
		t.Run("non_deprecated_"+cmd, func(t *testing.T) {
			message := detector.GetDeprecationMessage(cmd)
			if message != "" {
				t.Errorf("Expected empty message for non-deprecated command '%s', got: %s", cmd, message)
			}
		})
	}
}

func TestGetAlternativeActions(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	// Test discontinued commands (should have alternatives)
	discontinuedCommands := []string{"summary", "object-storage", "ojs"}
	for _, cmd := range discontinuedCommands {
		t.Run(cmd, func(t *testing.T) {
			actions := detector.GetAlternativeActions(cmd)
			if len(actions) == 0 {
				t.Errorf("Expected alternative actions for discontinued command '%s'", cmd)
			}
		})
	}

	// Test renamed commands (should not have alternatives)
	renamedCommands := []string{"iso-image", "startup-script", "ipv4"}
	for _, cmd := range renamedCommands {
		t.Run("renamed_"+cmd, func(t *testing.T) {
			actions := detector.GetAlternativeActions(cmd)
			if len(actions) != 0 {
				t.Errorf("Expected no alternative actions for renamed command '%s', got %v", cmd, actions)
			}
		})
	}

	// Test non-deprecated commands
	nonDeprecated := []string{"server", "invalid"}
	for _, cmd := range nonDeprecated {
		t.Run("non_deprecated_"+cmd, func(t *testing.T) {
			actions := detector.GetAlternativeActions(cmd)
			if len(actions) != 0 {
				t.Errorf("Expected no alternative actions for non-deprecated command '%s', got %v", cmd, actions)
			}
		})
	}
}

func TestGenerateMigrationMessage(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		command     string
		expectEmpty bool
		mustContain []string
	}{
		{
			command:     "iso-image",
			expectEmpty: false,
			mustContain: []string{"iso-image", "cdrom", "名称変更", "詳細"},
		},
		{
			command:     "summary",
			expectEmpty: false,
			mustContain: []string{"summary", "廃止", "代替手段", "詳細"},
		},
		{
			command:     "server",
			expectEmpty: true,
			mustContain: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			message := detector.GenerateMigrationMessage(tt.command)

			if tt.expectEmpty && message != "" {
				t.Errorf("Expected empty message for '%s', got: %s", tt.command, message)
				return
			}

			if !tt.expectEmpty && message == "" {
				t.Errorf("Expected non-empty message for '%s'", tt.command)
				return
			}

			for _, must := range tt.mustContain {
				if !strings.Contains(message, must) {
					t.Errorf("Expected message for '%s' to contain '%s', got: %s", tt.command, must, message)
				}
			}
		})
	}
}

func TestGetAllDeprecatedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()
	all := detector.GetAllDeprecatedCommands()

	expectedCount := 9
	if len(all) != expectedCount {
		t.Errorf("Expected %d deprecated commands, got %d", expectedCount, len(all))
	}

	// Check that all expected commands are present
	expectedCommands := []string{
		"iso-image", "startup-script", "ipv4", "product-disk", "product-internet", "product-server",
		"summary", "object-storage", "ojs",
	}

	for _, expected := range expectedCommands {
		if info, exists := all[expected]; !exists {
			t.Errorf("Expected deprecated command '%s' not found", expected)
		} else if info.Command != expected {
			t.Errorf("Expected command field to be '%s', got '%s'", expected, info.Command)
		}
	}

	// Test that modifying the returned map doesn't affect the original
	all["test"] = &DeprecationInfo{Command: "test"}
	allAgain := detector.GetAllDeprecatedCommands()
	if _, exists := allAgain["test"]; exists {
		t.Error("Modifying returned map should not affect original data")
	}
}

func TestDetectorGetRenamedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()
	renamed := detector.GetRenamedCommands()

	expectedRenamed := map[string]string{
		"iso-image":        "cdrom",
		"startup-script":   "note",
		"ipv4":             "ipaddress",
		"product-disk":     "disk-plan",
		"product-internet": "internet-plan",
		"product-server":   "server-plan",
	}

	if len(renamed) != len(expectedRenamed) {
		t.Errorf("Expected %d renamed commands, got %d", len(expectedRenamed), len(renamed))
	}

	for old, expectedNew := range expectedRenamed {
		if actualNew, exists := renamed[old]; !exists {
			t.Errorf("Expected renamed command '%s' not found", old)
		} else if actualNew != expectedNew {
			t.Errorf("Expected '%s' -> '%s', got '%s' -> '%s'", old, expectedNew, old, actualNew)
		}
	}

	// Ensure discontinued commands are not included
	discontinuedCommands := []string{"summary", "object-storage", "ojs"}
	for _, cmd := range discontinuedCommands {
		if _, exists := renamed[cmd]; exists {
			t.Errorf("Discontinued command '%s' should not be in renamed commands", cmd)
		}
	}
}

func TestDetectorGetDiscontinuedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()
	discontinued := detector.GetDiscontinuedCommands()

	expectedDiscontinued := []string{"summary", "object-storage", "ojs"}

	if len(discontinued) != len(expectedDiscontinued) {
		t.Errorf("Expected %d discontinued commands, got %d", len(expectedDiscontinued), len(discontinued))
	}

	discontinuedMap := make(map[string]bool)
	for _, cmd := range discontinued {
		discontinuedMap[cmd] = true
	}

	for _, expected := range expectedDiscontinued {
		if !discontinuedMap[expected] {
			t.Errorf("Expected discontinued command '%s' not found", expected)
		}
	}

	// Ensure renamed commands are not included
	renamedCommands := []string{"iso-image", "startup-script", "ipv4"}
	for _, cmd := range renamedCommands {
		if discontinuedMap[cmd] {
			t.Errorf("Renamed command '%s' should not be in discontinued commands", cmd)
		}
	}
}

func TestCommandCounts(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	totalCount := detector.GetDeprecatedCommandCount()
	renamedCount := detector.GetRenamedCommandCount()
	discontinuedCount := detector.GetDiscontinuedCommandCount()

	expectedTotal := 9
	expectedRenamed := 6
	expectedDiscontinued := 3

	if totalCount != expectedTotal {
		t.Errorf("Expected total count %d, got %d", expectedTotal, totalCount)
	}

	if renamedCount != expectedRenamed {
		t.Errorf("Expected renamed count %d, got %d", expectedRenamed, renamedCount)
	}

	if discontinuedCount != expectedDiscontinued {
		t.Errorf("Expected discontinued count %d, got %d", expectedDiscontinued, discontinuedCount)
	}

	if renamedCount+discontinuedCount != totalCount {
		t.Errorf("Renamed count (%d) + discontinued count (%d) should equal total count (%d)",
			renamedCount, discontinuedCount, totalCount)
	}
}

func TestDetectorCaseSensitivity(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	tests := []struct {
		command string
		should  bool
	}{
		{"iso-image", true},
		{"ISO-IMAGE", true},
		{"Iso-Image", true},
		{"summary", true},
		{"SUMMARY", true},
		{"Summary", true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := detector.IsDeprecated(tt.command)
			if result != tt.should {
				t.Errorf("Expected IsDeprecated('%s') = %t, got %t", tt.command, tt.should, result)
			}
		})
	}
}

func TestValidateConsistencyWithDeprecatedCommands(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	err := detector.ValidateConsistencyWithDeprecatedCommands()
	if err != nil {
		t.Errorf("Expected consistency validation to pass, got error: %v", err)
	}
}

func TestHandleMethods(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	// Test handleRenamedCommand
	renamedInfo := &DeprecationInfo{
		Command:            "iso-image",
		ReplacementCommand: "cdrom",
		DeprecationType:    "renamed",
		DocumentationURL:   "https://example.com",
	}
	renamedMessage := detector.handleRenamedCommand(renamedInfo)
	if !strings.Contains(renamedMessage, "iso-image") || !strings.Contains(renamedMessage, "cdrom") {
		t.Errorf("Renamed command message should contain both old and new command names")
	}

	// Test handleDiscontinuedCommand
	discontinuedInfo := &DeprecationInfo{
		Command:         "summary",
		DeprecationType: "discontinued",
		Message:         "Summary is discontinued",
		AlternativeActions: []string{
			"Use bill command",
			"Use self command",
		},
		DocumentationURL: "https://example.com",
	}
	discontinuedMessage := detector.handleDiscontinuedCommand(discontinuedInfo)
	if !strings.Contains(discontinuedMessage, "代替手段") ||
		!strings.Contains(discontinuedMessage, "bill command") ||
		!strings.Contains(discontinuedMessage, "self command") {
		t.Errorf("Discontinued command message should contain alternatives")
	}
}

func TestSpecificAlternativeActions(t *testing.T) {
	detector := NewDeprecatedCommandDetector()

	// Test summary alternatives
	summaryActions := detector.GetAlternativeActions("summary")
	expectedSummaryKeywords := []string{"bill", "self", "list", "rest"}
	for _, keyword := range expectedSummaryKeywords {
		found := false
		for _, action := range summaryActions {
			if strings.Contains(action, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected summary alternatives to mention '%s'", keyword)
		}
	}

	// Test object-storage alternatives
	objActions := detector.GetAlternativeActions("object-storage")
	expectedObjKeywords := []string{"S3", "Terraform", "クラウドストレージ"}
	for _, keyword := range expectedObjKeywords {
		found := false
		for _, action := range objActions {
			if strings.Contains(action, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected object-storage alternatives to mention '%s'", keyword)
		}
	}
}
