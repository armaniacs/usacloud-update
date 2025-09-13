package validation

import (
	"sort"
	"testing"
)

func TestServerSubcommandsCount(t *testing.T) {
	expectedCount := 15
	actualCount := GetServerSubcommandCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d server subcommands, but got %d", expectedCount, actualCount)
	}
}

func TestGetServerSubcommands(t *testing.T) {
	subcommands := GetServerSubcommands()
	expectedCount := 15

	if len(subcommands) != expectedCount {
		t.Errorf("Expected %d subcommands, got %d", expectedCount, len(subcommands))
	}

	// Check that all expected subcommands are present
	expectedSubcommands := []string{
		"list", "read", "create", "update", "delete",
		"boot", "shutdown", "reset",
		"send-nmi",
		"monitor-cpu",
		"ssh", "vnc", "rdp",
		"wait-until-ready", "wait-until-shutdown",
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

func TestIsValidServerSubcommand(t *testing.T) {
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

		// Power control operations
		{"boot", true},
		{"shutdown", true},
		{"reset", true},

		// Management operations
		{"send-nmi", true},

		// Monitoring operations
		{"monitor-cpu", true},

		// Connection operations
		{"ssh", true},
		{"vnc", true},
		{"rdp", true},

		// Wait operations
		{"wait-until-ready", true},
		{"wait-until-shutdown", true},

		// Invalid subcommands
		{"invalid", false},
		{"connect", false},    // This is for disk, not server
		{"disconnect", false}, // This is for disk, not server
		{"", false},
		{"LIST", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			result := IsValidServerSubcommand(tt.subcommand)
			if result != tt.valid {
				t.Errorf("IsValidServerSubcommand(%s): expected %t, got %t",
					tt.subcommand, tt.valid, result)
			}
		})
	}
}

func TestGetServerSubcommandDescription(t *testing.T) {
	tests := []struct {
		subcommand   string
		expectExists bool
		contains     string // Check if description contains this text
	}{
		{"list", true, "一覧"},
		{"create", true, "作成"},
		{"boot", true, "起動"},
		{"shutdown", true, "シャットダウン"},
		{"ssh", true, "SSH"},
		{"monitor-cpu", true, "CPU"},
		{"send-nmi", true, "NMI"},
		{"wait-until-ready", true, "待機"},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			description, exists := GetServerSubcommandDescription(tt.subcommand)

			if exists != tt.expectExists {
				t.Errorf("GetServerSubcommandDescription(%s): expected exists=%t, got %t",
					tt.subcommand, tt.expectExists, exists)
			}

			if tt.expectExists && tt.contains != "" {
				if len(description) == 0 {
					t.Errorf("GetServerSubcommandDescription(%s): expected non-empty description",
						tt.subcommand)
				}
			}
		})
	}
}

func TestBasicCRUDSubcommands(t *testing.T) {
	crudCommands := GetBasicCRUDSubcommands()
	expectedCRUD := []string{"list", "read", "create", "update", "delete"}

	if len(crudCommands) != len(expectedCRUD) {
		t.Errorf("Expected %d CRUD commands, got %d", len(expectedCRUD), len(crudCommands))
	}

	crudMap := make(map[string]bool)
	for _, cmd := range crudCommands {
		crudMap[cmd] = true
	}

	for _, expected := range expectedCRUD {
		if !crudMap[expected] {
			t.Errorf("Expected CRUD command '%s' not found", expected)
		}
	}

	// Test IsBasicCRUDSubcommand function
	for _, cmd := range expectedCRUD {
		if !IsBasicCRUDSubcommand(cmd) {
			t.Errorf("IsBasicCRUDSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-CRUD commands
	nonCRUDCommands := []string{"boot", "ssh", "monitor-cpu", "wait-until-ready"}
	for _, cmd := range nonCRUDCommands {
		if IsBasicCRUDSubcommand(cmd) {
			t.Errorf("IsBasicCRUDSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestPowerControlSubcommands(t *testing.T) {
	powerCommands := GetPowerControlSubcommands()
	expectedPower := []string{"boot", "shutdown", "reset"}

	if len(powerCommands) != len(expectedPower) {
		t.Errorf("Expected %d power control commands, got %d", len(expectedPower), len(powerCommands))
	}

	powerMap := make(map[string]bool)
	for _, cmd := range powerCommands {
		powerMap[cmd] = true
	}

	for _, expected := range expectedPower {
		if !powerMap[expected] {
			t.Errorf("Expected power control command '%s' not found", expected)
		}
	}

	// Test IsPowerControlSubcommand function
	for _, cmd := range expectedPower {
		if !IsPowerControlSubcommand(cmd) {
			t.Errorf("IsPowerControlSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-power commands
	nonPowerCommands := []string{"list", "ssh", "monitor-cpu", "wait-until-ready"}
	for _, cmd := range nonPowerCommands {
		if IsPowerControlSubcommand(cmd) {
			t.Errorf("IsPowerControlSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestConnectionSubcommands(t *testing.T) {
	connectionCommands := GetConnectionSubcommands()
	expectedConnection := []string{"ssh", "vnc", "rdp"}

	if len(connectionCommands) != len(expectedConnection) {
		t.Errorf("Expected %d connection commands, got %d", len(expectedConnection), len(connectionCommands))
	}

	connectionMap := make(map[string]bool)
	for _, cmd := range connectionCommands {
		connectionMap[cmd] = true
	}

	for _, expected := range expectedConnection {
		if !connectionMap[expected] {
			t.Errorf("Expected connection command '%s' not found", expected)
		}
	}

	// Test IsConnectionSubcommand function
	for _, cmd := range expectedConnection {
		if !IsConnectionSubcommand(cmd) {
			t.Errorf("IsConnectionSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-connection commands
	nonConnectionCommands := []string{"list", "boot", "monitor-cpu", "wait-until-ready"}
	for _, cmd := range nonConnectionCommands {
		if IsConnectionSubcommand(cmd) {
			t.Errorf("IsConnectionSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestWaitSubcommands(t *testing.T) {
	waitCommands := GetWaitSubcommands()
	expectedWait := []string{"wait-until-ready", "wait-until-shutdown"}

	if len(waitCommands) != len(expectedWait) {
		t.Errorf("Expected %d wait commands, got %d", len(expectedWait), len(waitCommands))
	}

	waitMap := make(map[string]bool)
	for _, cmd := range waitCommands {
		waitMap[cmd] = true
	}

	for _, expected := range expectedWait {
		if !waitMap[expected] {
			t.Errorf("Expected wait command '%s' not found", expected)
		}
	}

	// Test IsWaitSubcommand function
	for _, cmd := range expectedWait {
		if !IsWaitSubcommand(cmd) {
			t.Errorf("IsWaitSubcommand(%s): expected true, got false", cmd)
		}
	}

	// Test with non-wait commands
	nonWaitCommands := []string{"list", "boot", "ssh", "monitor-cpu"}
	for _, cmd := range nonWaitCommands {
		if IsWaitSubcommand(cmd) {
			t.Errorf("IsWaitSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestManagementSubcommands(t *testing.T) {
	// Test IsManagementSubcommand function
	if !IsManagementSubcommand("send-nmi") {
		t.Error("IsManagementSubcommand(send-nmi): expected true, got false")
	}

	// Test with non-management commands
	nonManagementCommands := []string{"list", "boot", "ssh", "monitor-cpu", "wait-until-ready"}
	for _, cmd := range nonManagementCommands {
		if IsManagementSubcommand(cmd) {
			t.Errorf("IsManagementSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestMonitoringSubcommands(t *testing.T) {
	// Test IsMonitoringSubcommand function
	if !IsMonitoringSubcommand("monitor-cpu") {
		t.Error("IsMonitoringSubcommand(monitor-cpu): expected true, got false")
	}

	// Test with non-monitoring commands
	nonMonitoringCommands := []string{"list", "boot", "ssh", "send-nmi", "wait-until-ready"}
	for _, cmd := range nonMonitoringCommands {
		if IsMonitoringSubcommand(cmd) {
			t.Errorf("IsMonitoringSubcommand(%s): expected false, got true", cmd)
		}
	}
}

func TestSubcommandCategorization(t *testing.T) {
	tests := []struct {
		subcommand   string
		isCRUD       bool
		isPower      bool
		isConnection bool
		isWait       bool
		isManagement bool
		isMonitoring bool
	}{
		{"list", true, false, false, false, false, false},
		{"create", true, false, false, false, false, false},
		{"boot", false, true, false, false, false, false},
		{"shutdown", false, true, false, false, false, false},
		{"ssh", false, false, true, false, false, false},
		{"rdp", false, false, true, false, false, false},
		{"wait-until-ready", false, false, false, true, false, false},
		{"wait-until-shutdown", false, false, false, true, false, false},
		{"send-nmi", false, false, false, false, true, false},
		{"monitor-cpu", false, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			if IsBasicCRUDSubcommand(tt.subcommand) != tt.isCRUD {
				t.Errorf("IsBasicCRUDSubcommand(%s): expected %t", tt.subcommand, tt.isCRUD)
			}
			if IsPowerControlSubcommand(tt.subcommand) != tt.isPower {
				t.Errorf("IsPowerControlSubcommand(%s): expected %t", tt.subcommand, tt.isPower)
			}
			if IsConnectionSubcommand(tt.subcommand) != tt.isConnection {
				t.Errorf("IsConnectionSubcommand(%s): expected %t", tt.subcommand, tt.isConnection)
			}
			if IsWaitSubcommand(tt.subcommand) != tt.isWait {
				t.Errorf("IsWaitSubcommand(%s): expected %t", tt.subcommand, tt.isWait)
			}
			if IsManagementSubcommand(tt.subcommand) != tt.isManagement {
				t.Errorf("IsManagementSubcommand(%s): expected %t", tt.subcommand, tt.isManagement)
			}
			if IsMonitoringSubcommand(tt.subcommand) != tt.isMonitoring {
				t.Errorf("IsMonitoringSubcommand(%s): expected %t", tt.subcommand, tt.isMonitoring)
			}
		})
	}
}

func TestAllSubcommandsHaveDescriptions(t *testing.T) {
	subcommands := GetServerSubcommands()

	for _, sc := range subcommands {
		description, exists := GetServerSubcommandDescription(sc)
		if !exists {
			t.Errorf("Subcommand '%s' missing description", sc)
		}
		if len(description) == 0 {
			t.Errorf("Subcommand '%s' has empty description", sc)
		}
	}

	// Check that descriptions map doesn't have extra entries
	descriptionCount := len(ServerSubcommandDescriptions)
	subcommandCount := len(subcommands)

	if descriptionCount != subcommandCount {
		t.Errorf("Description count (%d) doesn't match subcommand count (%d)",
			descriptionCount, subcommandCount)
	}
}
