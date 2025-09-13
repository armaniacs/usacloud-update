package validation

import (
	"sort"
	"testing"
)

func TestDiskSubcommandsCount(t *testing.T) {
	expectedCount := 9
	actualCount := GetDiskSubcommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d disk subcommands, but got %d", expectedCount, actualCount)
	}
}

func TestGetDiskSubcommands(t *testing.T) {
	subcommands := GetDiskSubcommands()
	expectedCount := 9

	if len(subcommands) != expectedCount {
		t.Errorf("Expected %d subcommands, got %d", expectedCount, len(subcommands))
	}

	// Check that all expected subcommands are present
	expectedSubcommands := []string{
		"list", "read", "create", "update", "delete",
		"connect", "disconnect",
		"clone", "resize",
	}

	subcommandMap := make(map[string]bool)
	for _, sc := range subcommands {
		subcommandMap[sc] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandMap[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}

	// Check for duplicates
	sort.Strings(subcommands)
	for i := 1; i < len(subcommands); i++ {
		if subcommands[i] == subcommands[i-1] {
			t.Errorf("Duplicate subcommand found: %s", subcommands[i])
		}
	}
}

func TestIsValidDiskSubcommand(t *testing.T) {
	tests := []struct {
		subcommand string
		valid      bool
	}{
		// Basic CRUD operations
		{"list", true},
		{"read", true},
		{"create", true},
		{"update", true},
		{"delete", true},

		// Attachment operations
		{"connect", true},
		{"disconnect", true},

		// Management operations
		{"clone", true},
		{"resize", true},

		// Invalid subcommands
		{"invalid", false},
		{"boot", false},     // This is for server, not disk
		{"shutdown", false}, // This is for server, not disk
		{"ssh", false},      // This is for server, not disk
		{"", false},
		{"LIST", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			result := IsValidDiskSubcommand(tt.subcommand)
			if result != tt.valid {
				t.Errorf("IsValidDiskSubcommand(%s): expected %t, got %t",
					tt.subcommand, tt.valid, result)
			}
		})
	}
}

func TestGetDiskSubcommandDescription(t *testing.T) {
	tests := []struct {
		subcommand   string
		expectExists bool
		contains     string // Check if description contains this text
	}{
		{"list", true, "一覧"},
		{"create", true, "作成"},
		{"connect", true, "接続"},
		{"disconnect", true, "切断"},
		{"clone", true, "クローン"},
		{"resize", true, "サイズ"},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			description, exists := GetDiskSubcommandDescription(tt.subcommand)

			if exists != tt.expectExists {
				t.Errorf("GetDiskSubcommandDescription(%s): expected exists=%t, got %t",
					tt.subcommand, tt.expectExists, exists)
			}

			if tt.expectExists && tt.contains != "" {
				if len(description) == 0 {
					t.Errorf("GetDiskSubcommandDescription(%s): expected non-empty description",
						tt.subcommand)
				}
			}
		})
	}
}

func TestGetDiskSubcommandCategory(t *testing.T) {
	tests := []struct {
		subcommand       string
		expectedCategory string
		expectExists     bool
	}{
		{"list", "basic", true},
		{"create", "basic", true},
		{"connect", "attachment", true},
		{"disconnect", "attachment", true},
		{"clone", "management", true},
		{"resize", "management", true},
		{"invalid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			category, exists := GetDiskSubcommandCategory(tt.subcommand)

			if exists != tt.expectExists {
				t.Errorf("GetDiskSubcommandCategory(%s): expected exists=%t, got %t",
					tt.subcommand, tt.expectExists, exists)
			}

			if tt.expectExists && category != tt.expectedCategory {
				t.Errorf("GetDiskSubcommandCategory(%s): expected '%s', got '%s'",
					tt.subcommand, tt.expectedCategory, category)
			}
		})
	}
}

func TestBasicDiskSubcommands(t *testing.T) {
	basicCommands := GetBasicDiskSubcommands()
	expectedBasic := []string{"list", "read", "create", "update", "delete"}

	if len(basicCommands) != len(expectedBasic) {
		t.Errorf("Expected %d basic commands, got %d", len(expectedBasic), len(basicCommands))
	}

	basicMap := make(map[string]bool)
	for _, cmd := range basicCommands {
		basicMap[cmd] = true
	}

	for _, expected := range expectedBasic {
		if !basicMap[expected] {
			t.Errorf("Expected basic command '%s' not found", expected)
		}
	}

	// Test IsBasicDiskSubcommand function
	for _, cmd := range expectedBasic {
		if !IsBasicDiskSubcommand(cmd) {
			t.Errorf("IsBasicDiskSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-basic commands
	nonBasicCommands := []string{"connect", "clone", "resize"}
	for _, cmd := range nonBasicCommands {
		if IsBasicDiskSubcommand(cmd) {
			t.Errorf("IsBasicDiskSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestAttachmentDiskSubcommands(t *testing.T) {
	attachmentCommands := GetAttachmentDiskSubcommands()
	expectedAttachment := []string{"connect", "disconnect"}

	if len(attachmentCommands) != len(expectedAttachment) {
		t.Errorf("Expected %d attachment commands, got %d", len(expectedAttachment), len(attachmentCommands))
	}

	attachmentMap := make(map[string]bool)
	for _, cmd := range attachmentCommands {
		attachmentMap[cmd] = true
	}

	for _, expected := range expectedAttachment {
		if !attachmentMap[expected] {
			t.Errorf("Expected attachment command '%s' not found", expected)
		}
	}

	// Test IsAttachmentDiskSubcommand function
	for _, cmd := range expectedAttachment {
		if !IsAttachmentDiskSubcommand(cmd) {
			t.Errorf("IsAttachmentDiskSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-attachment commands
	nonAttachmentCommands := []string{"list", "clone", "resize"}
	for _, cmd := range nonAttachmentCommands {
		if IsAttachmentDiskSubcommand(cmd) {
			t.Errorf("IsAttachmentDiskSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestManagementDiskSubcommands(t *testing.T) {
	managementCommands := GetManagementDiskSubcommands()
	expectedManagement := []string{"clone", "resize"}

	if len(managementCommands) != len(expectedManagement) {
		t.Errorf("Expected %d management commands, got %d", len(expectedManagement), len(managementCommands))
	}

	managementMap := make(map[string]bool)
	for _, cmd := range managementCommands {
		managementMap[cmd] = true
	}

	for _, expected := range expectedManagement {
		if !managementMap[expected] {
			t.Errorf("Expected management command '%s' not found", expected)
		}
	}

	// Test IsManagementDiskSubcommand function
	for _, cmd := range expectedManagement {
		if !IsManagementDiskSubcommand(cmd) {
			t.Errorf("IsManagementDiskSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-management commands
	nonManagementCommands := []string{"list", "connect", "disconnect"}
	for _, cmd := range nonManagementCommands {
		if IsManagementDiskSubcommand(cmd) {
			t.Errorf("IsManagementDiskSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestDiskSubcommandCategorization(t *testing.T) {
	tests := []struct {
		subcommand       string
		isBasic          bool
		isAttachment     bool
		isManagement     bool
		expectedCategory string
	}{
		{"list", true, false, false, "basic"},
		{"create", true, false, false, "basic"},
		{"connect", false, true, false, "attachment"},
		{"disconnect", false, true, false, "attachment"},
		{"clone", false, false, true, "management"},
		{"resize", false, false, true, "management"},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			if IsBasicDiskSubcommand(tt.subcommand) != tt.isBasic {
				t.Errorf("IsBasicDiskSubcommand(%s): expected %t", tt.subcommand, tt.isBasic)
			}
			if IsAttachmentDiskSubcommand(tt.subcommand) != tt.isAttachment {
				t.Errorf("IsAttachmentDiskSubcommand(%s): expected %t", tt.subcommand, tt.isAttachment)
			}
			if IsManagementDiskSubcommand(tt.subcommand) != tt.isManagement {
				t.Errorf("IsManagementDiskSubcommand(%s): expected %t", tt.subcommand, tt.isManagement)
			}

			// Test category consistency
			category, exists := GetDiskSubcommandCategory(tt.subcommand)
			if !exists {
				t.Errorf("GetDiskSubcommandCategory(%s): expected category to exist", tt.subcommand)
			}
			if category != tt.expectedCategory {
				t.Errorf("GetDiskSubcommandCategory(%s): expected %s, got %s",
					tt.subcommand, tt.expectedCategory, category)
			}
		})
	}
}

func TestGetDiskSubcommandsByCategory(t *testing.T) {
	tests := []struct {
		category string
		expected []string
	}{
		{"basic", []string{"list", "read", "create", "update", "delete"}},
		{"attachment", []string{"connect", "disconnect"}},
		{"management", []string{"clone", "resize"}},
		{"invalid", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			result := GetDiskSubcommandsByCategory(tt.category)

			if len(result) != len(tt.expected) {
				t.Errorf("GetDiskSubcommandsByCategory(%s): expected %d commands, got %d",
					tt.category, len(tt.expected), len(result))
				return
			}

			resultMap := make(map[string]bool)
			for _, cmd := range result {
				resultMap[cmd] = true
			}

			for _, expectedCmd := range tt.expected {
				if !resultMap[expectedCmd] {
					t.Errorf("GetDiskSubcommandsByCategory(%s): missing expected command '%s'",
						tt.category, expectedCmd)
				}
			}
		})
	}
}

func TestGetAllDiskCategories(t *testing.T) {
	categories := GetAllDiskCategories()
	expectedCategories := []string{"basic", "attachment", "management"}

	if len(categories) != len(expectedCategories) {
		t.Errorf("Expected %d categories, got %d", len(expectedCategories), len(categories))
	}

	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}

	for _, expected := range expectedCategories {
		if !categoryMap[expected] {
			t.Errorf("Expected category '%s' not found", expected)
		}
	}
}

func TestAllSubcommandsHaveDescriptionsAndCategories(t *testing.T) {
	subcommands := GetDiskSubcommands()

	for _, sc := range subcommands {
		// Test description exists
		description, descExists := GetDiskSubcommandDescription(sc)
		if !descExists {
			t.Errorf("Subcommand '%s' missing description", sc)
		}
		if len(description) == 0 {
			t.Errorf("Subcommand '%s' has empty description", sc)
		}

		// Test category exists
		category, catExists := GetDiskSubcommandCategory(sc)
		if !catExists {
			t.Errorf("Subcommand '%s' missing category", sc)
		}
		if len(category) == 0 {
			t.Errorf("Subcommand '%s' has empty category", sc)
		}
	}

	// Check that descriptions and categories maps don't have extra entries
	descriptionCount := len(DiskSubcommandDescriptions)
	categoryCount := len(DiskSubcommandCategories)
	subcommandCount := len(subcommands)

	if descriptionCount != subcommandCount {
		t.Errorf("Description count (%d) doesn't match subcommand count (%d)",
			descriptionCount, subcommandCount)
	}

	if categoryCount != subcommandCount {
		t.Errorf("Category count (%d) doesn't match subcommand count (%d)",
			categoryCount, subcommandCount)
	}
}

func TestDiskSpecificOperations(t *testing.T) {
	// Test that disk-specific operations are properly categorized
	tests := []struct {
		subcommand  string
		category    string
		description string
	}{
		{"connect", "attachment", "接続"},
		{"disconnect", "attachment", "切断"},
		{"clone", "management", "クローン"},
		{"resize", "management", "サイズ"},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			// Verify it's a valid disk subcommand
			if !IsValidDiskSubcommand(tt.subcommand) {
				t.Errorf("IsValidDiskSubcommand(%s): expected true, got false", tt.subcommand)
			}

			// Verify category
			category, exists := GetDiskSubcommandCategory(tt.subcommand)
			if !exists || category != tt.category {
				t.Errorf("GetDiskSubcommandCategory(%s): expected '%s', got '%s'",
					tt.subcommand, tt.category, category)
			}

			// Verify description exists
			_, exists = GetDiskSubcommandDescription(tt.subcommand)
			if !exists {
				t.Errorf("GetDiskSubcommandDescription(%s): expected description to exist", tt.subcommand)
			}
		})
	}
}
