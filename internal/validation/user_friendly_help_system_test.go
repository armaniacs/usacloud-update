package validation

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewUserFriendlyHelpSystem(t *testing.T) {
	cmdValidator := NewMainCommandValidator()
	subValidator := NewSubcommandValidator(cmdValidator)
	formatter := NewDefaultComprehensiveErrorFormatter()

	helpSystem := NewUserFriendlyHelpSystem(cmdValidator, subValidator, formatter, true)
	if helpSystem == nil {
		t.Error("Expected help system to be created, got nil")
	}
	if !helpSystem.interactiveModeEnabled {
		t.Error("Expected interactive mode to be enabled")
	}
}

func TestNewDefaultUserFriendlyHelpSystem(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()
	if helpSystem == nil {
		t.Error("Expected help system to be created, got nil")
	}
	if !helpSystem.interactiveModeEnabled {
		t.Error("Expected interactive mode to be enabled by default")
	}
}

func TestCreateDefaultContext(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()
	context := helpSystem.createDefaultContext()

	if context == nil {
		t.Error("Expected context to be created, got nil")
	}
	if context.UserSkillLevel != SkillBeginner {
		t.Error("Expected default skill level to be beginner")
	}
	if context.PreferredFormat != FormatBasic {
		t.Error("Expected default format to be basic")
	}
}

func TestShowBasicHelp(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()

	tests := []struct {
		name       string
		skillLevel SkillLevel
	}{
		{"Beginner help", SkillBeginner},
		{"Intermediate help", SkillIntermediate},
		{"Advanced help", SkillAdvanced},
		{"Expert help", SkillExpert},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &HelpContext{
				UserSkillLevel:  tt.skillLevel,
				PreferredFormat: FormatBasic,
				LastAccessed:    time.Now(),
			}

			err := helpSystem.showBasicHelp(context)
			if err != nil {
				t.Errorf("showBasicHelp() error = %v", err)
			}
		})
	}
}

func TestShowHelp(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()

	tests := []struct {
		name    string
		context *HelpContext
	}{
		{
			name:    "Nil context",
			context: nil,
		},
		{
			name: "Basic format",
			context: &HelpContext{
				PreferredFormat: FormatBasic,
				UserSkillLevel:  SkillBeginner,
			},
		},
		{
			name: "Detailed format",
			context: &HelpContext{
				PreferredFormat: FormatDetailed,
				UserSkillLevel:  SkillIntermediate,
			},
		},
		{
			name: "Example format",
			context: &HelpContext{
				PreferredFormat: FormatExample,
				UserSkillLevel:  SkillAdvanced,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helpSystem.ShowHelp(tt.context)
			if err != nil {
				t.Errorf("ShowHelp() error = %v", err)
			}
		})
	}
}

func TestNewHelpDatabase(t *testing.T) {
	db := NewHelpDatabase()
	if db == nil {
		t.Error("Expected help database to be created, got nil")
	}

	if len(db.commonMistakes) == 0 {
		t.Error("Expected common mistakes to be initialized")
	}

	if len(db.tutorialSteps) == 0 {
		t.Error("Expected tutorial steps to be initialized")
	}

	if len(db.conceptMap) == 0 {
		t.Error("Expected concept map to be initialized")
	}

	if len(db.migrationGuides) == 0 {
		t.Error("Expected migration guides to be initialized")
	}
}

func TestGetCommonMistakes(t *testing.T) {
	mistakes := getCommonMistakes()
	if len(mistakes) == 0 {
		t.Error("Expected common mistakes to be returned")
	}

	// Check for expected mistakes
	expectedPatterns := []string{
		"usacloud server show",
		"usacloud server list --selector",
		"usacloud iso-image list",
	}

	for _, expectedPattern := range expectedPatterns {
		found := false
		for _, mistake := range mistakes {
			if mistake.Pattern == expectedPattern {
				found = true
				// Verify mistake has required fields
				if mistake.Description == "" {
					t.Errorf("Mistake pattern '%s' missing description", expectedPattern)
				}
				if len(mistake.CorrectExamples) == 0 {
					t.Errorf("Mistake pattern '%s' missing correct examples", expectedPattern)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected mistake pattern '%s' not found", expectedPattern)
		}
	}
}

func TestGetTutorialSteps(t *testing.T) {
	steps := getTutorialSteps()
	if len(steps) == 0 {
		t.Error("Expected tutorial steps to be returned")
	}

	// Check for expected steps
	expectedStepIDs := []string{"step1", "step2", "step3"}
	for _, expectedID := range expectedStepIDs {
		found := false
		for _, step := range steps {
			if step.StepID == expectedID {
				found = true
				// Verify step has required fields
				if step.Title == "" {
					t.Errorf("Tutorial step '%s' missing title", expectedID)
				}
				if step.Description == "" {
					t.Errorf("Tutorial step '%s' missing description", expectedID)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected tutorial step '%s' not found", expectedID)
		}
	}
}

func TestInteractiveCommandBuilderGenerateFinalCommand(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()
	builder := &InteractiveCommandBuilder{
		helpSystem: helpSystem,
		options:    make(map[string]string),
	}

	tests := []struct {
		name     string
		mainCmd  string
		subCmd   string
		options  map[string]string
		expected string
	}{
		{
			name:     "Basic command",
			mainCmd:  "server",
			subCmd:   "list",
			options:  map[string]string{},
			expected: "usacloud server list",
		},
		{
			name:    "Command with options",
			mainCmd: "server",
			subCmd:  "create",
			options: map[string]string{
				"name": "test-server",
				"zone": "tk1v",
			},
			expected: "usacloud server create --name=test-server --zone=tk1v",
		},
		{
			name:     "Command without subcommand",
			mainCmd:  "config",
			subCmd:   "",
			options:  map[string]string{},
			expected: "usacloud config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.generateFinalCommand(tt.mainCmd, tt.subCmd, tt.options)

			// Check that all expected parts are present
			if !strings.Contains(result, "usacloud") {
				t.Error("Expected command to contain 'usacloud'")
			}
			if !strings.Contains(result, tt.mainCmd) {
				t.Errorf("Expected command to contain main command '%s'", tt.mainCmd)
			}
			if tt.subCmd != "" && !strings.Contains(result, tt.subCmd) {
				t.Errorf("Expected command to contain subcommand '%s'", tt.subCmd)
			}

			for key, value := range tt.options {
				expectedOption := fmt.Sprintf("--%s=%s", key, value)
				if !strings.Contains(result, expectedOption) {
					t.Errorf("Expected command to contain option '%s'", expectedOption)
				}
			}
		})
	}
}

func TestLoadOrCreateUserProfile(t *testing.T) {
	profile := loadOrCreateUserProfile()
	if profile == nil {
		t.Error("Expected user profile to be created, got nil")
	}

	if profile.UserID == "" {
		t.Error("Expected user ID to be set")
	}

	if profile.SkillLevel < SkillBeginner || profile.SkillLevel > SkillExpert {
		t.Error("Expected valid skill level")
	}

	if profile.PreferredFormat < FormatBasic || profile.PreferredFormat > FormatExample {
		t.Error("Expected valid help format")
	}

	// Check that slices are initialized (not nil)
	if profile.CompletedTasks == nil {
		t.Error("Expected completed tasks to be initialized")
	}

	if profile.LearningGoals == nil {
		t.Error("Expected learning goals to be initialized")
	}
}

func TestHelpContextValidation(t *testing.T) {
	// Test ErrorHistory structure
	history := ErrorHistory{
		Timestamp:   time.Now(),
		Command:     "usacloud server show",
		ErrorType:   "deprecated_command",
		WasResolved: true,
		Resolution:  "Used 'read' instead of 'show'",
	}

	if history.Command == "" {
		t.Error("Expected command to be set")
	}

	if history.ErrorType == "" {
		t.Error("Expected error type to be set")
	}

	// Test HelpContext structure
	context := HelpContext{
		RequestedCommand: "server",
		PreviousErrors:   []ErrorHistory{history},
		UserSkillLevel:   SkillBeginner,
		PreferredFormat:  FormatBasic,
		LastAccessed:     time.Now(),
	}

	if context.RequestedCommand == "" {
		t.Error("Expected requested command to be set")
	}

	if len(context.PreviousErrors) == 0 {
		t.Error("Expected previous errors to be set")
	}
}

func TestCommonMistakeValidation(t *testing.T) {
	mistakes := getCommonMistakes()

	for _, mistake := range mistakes {
		// Validate required fields
		if mistake.Pattern == "" {
			t.Error("Common mistake should have pattern")
		}

		if mistake.Description == "" {
			t.Error("Common mistake should have description")
		}

		if len(mistake.CorrectExamples) == 0 {
			t.Error("Common mistake should have correct examples")
		}

		if mistake.Frequency < 0 || mistake.Frequency > 100 {
			t.Errorf("Common mistake frequency should be 0-100, got %d", mistake.Frequency)
		}

		// Validate that correct examples are not empty
		for _, example := range mistake.CorrectExamples {
			if example == "" {
				t.Error("Correct example should not be empty")
			}
		}
	}
}

func TestTutorialStepValidation(t *testing.T) {
	steps := getTutorialSteps()

	for _, step := range steps {
		// Validate required fields
		if step.StepID == "" {
			t.Error("Tutorial step should have ID")
		}

		if step.Title == "" {
			t.Error("Tutorial step should have title")
		}

		if step.Description == "" {
			t.Error("Tutorial step should have description")
		}

		// Commands can be empty, but if present should be valid
		for _, command := range step.Commands {
			if command == "" {
				t.Error("Tutorial command should not be empty")
			}
			if !strings.HasPrefix(command, "usacloud ") {
				t.Errorf("Tutorial command should start with 'usacloud ', got '%s'", command)
			}
		}
	}
}

func TestGetSkillLevelString(t *testing.T) {
	tests := []struct {
		skill    SkillLevel
		expected string
	}{
		{SkillBeginner, "Beginner"},
		{SkillIntermediate, "Intermediate"},
		{SkillAdvanced, "Advanced"},
		{SkillExpert, "Expert"},
		{SkillLevel(999), "Unknown"}, // Invalid skill level
	}

	for _, tt := range tests {
		result := GetSkillLevelString(tt.skill)
		if result != tt.expected {
			t.Errorf("GetSkillLevelString(%v) = %s, expected %s", tt.skill, result, tt.expected)
		}
	}
}

func TestGetHelpFormatString(t *testing.T) {
	tests := []struct {
		format   HelpFormat
		expected string
	}{
		{FormatBasic, "Basic"},
		{FormatDetailed, "Detailed"},
		{FormatInteractive, "Interactive"},
		{FormatExample, "Example"},
		{HelpFormat(999), "Unknown"}, // Invalid format
	}

	for _, tt := range tests {
		result := GetHelpFormatString(tt.format)
		if result != tt.expected {
			t.Errorf("GetHelpFormatString(%v) = %s, expected %s", tt.format, result, tt.expected)
		}
	}
}

func TestHelpDatabaseInitialization(t *testing.T) {
	db := NewHelpDatabase()

	// Test concept map initialization
	if db.conceptMap["crud"] == nil {
		t.Error("Expected 'crud' concept to be initialized")
	}

	if db.conceptMap["selector"] == nil {
		t.Error("Expected 'selector' concept to be initialized")
	}

	// Validate concept structure
	crudConcept := db.conceptMap["crud"]
	if crudConcept.ConceptID != "crud" {
		t.Error("Expected concept ID to match key")
	}

	if crudConcept.Title == "" {
		t.Error("Expected concept to have title")
	}

	if crudConcept.Description == "" {
		t.Error("Expected concept to have description")
	}

	// Test migration guide initialization
	if db.migrationGuides["v0_to_v1"] == nil {
		t.Error("Expected 'v0_to_v1' migration guide to be initialized")
	}

	// Validate migration guide structure
	migrationGuide := db.migrationGuides["v0_to_v1"]
	if migrationGuide.FromVersion == "" || migrationGuide.ToVersion == "" {
		t.Error("Expected migration guide to have version information")
	}

	if len(migrationGuide.Changes) == 0 {
		t.Error("Expected migration guide to have changes")
	}

	if len(migrationGuide.Examples) == 0 {
		t.Error("Expected migration guide to have examples")
	}
}

func TestUserProfileFields(t *testing.T) {
	profile := loadOrCreateUserProfile()

	// Test that all required fields are initialized
	requiredStringFields := map[string]string{
		"UserID": profile.UserID,
	}

	for fieldName, fieldValue := range requiredStringFields {
		if fieldValue == "" {
			t.Errorf("Expected %s to be initialized", fieldName)
		}
	}

	// Test that numeric fields are within valid ranges
	if profile.TotalCommands < 0 {
		t.Error("Expected TotalCommands to be non-negative")
	}

	if profile.ErrorCount < 0 {
		t.Error("Expected ErrorCount to be non-negative")
	}

	if profile.SuccessRate < 0.0 || profile.SuccessRate > 1.0 {
		t.Errorf("Expected SuccessRate to be 0.0-1.0, got %f", profile.SuccessRate)
	}

	// Test that time fields are reasonable
	if profile.LastActivity.IsZero() {
		t.Error("Expected LastActivity to be set")
	}
}

func TestInteractiveBuilderFields(t *testing.T) {
	helpSystem := NewDefaultUserFriendlyHelpSystem()
	builder := &InteractiveCommandBuilder{
		helpSystem:  helpSystem,
		currentStep: StepMainCommand,
		command:     []string{"usacloud", "server"},
		options:     make(map[string]string),
	}

	if builder.helpSystem == nil {
		t.Error("Expected help system to be set")
	}

	if builder.currentStep != StepMainCommand {
		t.Error("Expected current step to be set")
	}

	if builder.options == nil {
		t.Error("Expected options map to be initialized")
	}

	if len(builder.command) == 0 {
		t.Error("Expected command parts to be set")
	}
}

func TestLearningTrackerStructures(t *testing.T) {
	// Test CompletedTask structure
	task := CompletedTask{
		TaskID:     "task1",
		Command:    "usacloud server list",
		Timestamp:  time.Now(),
		Difficulty: 5,
		Success:    true,
	}

	if task.TaskID == "" {
		t.Error("Expected task ID to be set")
	}

	if task.Difficulty < 1 || task.Difficulty > 10 {
		t.Errorf("Expected difficulty to be 1-10, got %d", task.Difficulty)
	}

	// Test LearningGoal structure
	goal := LearningGoal{
		GoalID:      "goal1",
		Title:       "Master server operations",
		Description: "Learn all server-related commands",
		Steps:       []string{"Learn list", "Learn create", "Learn delete"},
		Progress:    0.5,
		Deadline:    nil,
	}

	if goal.GoalID == "" {
		t.Error("Expected goal ID to be set")
	}

	if goal.Progress < 0.0 || goal.Progress > 1.0 {
		t.Errorf("Expected progress to be 0.0-1.0, got %f", goal.Progress)
	}

	if len(goal.Steps) == 0 {
		t.Error("Expected goal to have steps")
	}
}
