package preview

import (
	"testing"
	"time"
)

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeTypeOption, "option"},
		{ChangeTypeArgument, "argument"},
		{ChangeTypeCommand, "command"},
		{ChangeTypeFormat, "format"},
		{ChangeTypeRemoval, "removal"},
		{ChangeTypeAddition, "addition"},
	}

	for _, test := range tests {
		t.Run(string(test.changeType), func(t *testing.T) {
			if string(test.changeType) != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, string(test.changeType))
			}
		})
	}
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		riskLevel RiskLevel
		expected  string
	}{
		{RiskLow, "low"},
		{RiskMedium, "medium"},
		{RiskHigh, "high"},
		{RiskCritical, "critical"},
	}

	for _, test := range tests {
		t.Run(string(test.riskLevel), func(t *testing.T) {
			if string(test.riskLevel) != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, string(test.riskLevel))
			}
		})
	}
}

func TestCommandPreview_Creation(t *testing.T) {
	preview := &CommandPreview{
		Original:    "usacloud server list --output-type=csv",
		Transformed: "usacloud server list --output-type=json",
		Changes: []ChangeHighlight{
			{
				Type:        ChangeTypeOption,
				Original:    "csv",
				Replacement: "json",
				Reason:      "出力フォーマットをJSONに変更",
			},
		},
		Description: "サーバ一覧取得コマンド",
		Impact: &ImpactAnalysis{
			Risk:        RiskLow,
			Description: "情報取得のみのため影響なし",
			Complexity:  2,
		},
		Warnings: []string{"出力形式が変更されます"},
		Category: "server",
		Metadata: &PreviewMetadata{
			LineNumber:  1,
			GeneratedAt: time.Now(),
			Version:     "1.9.0",
		},
	}

	if preview.Original == "" {
		t.Error("Original should not be empty")
	}
	if preview.Transformed == "" {
		t.Error("Transformed should not be empty")
	}
	if len(preview.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(preview.Changes))
	}
	if preview.Impact == nil {
		t.Error("Impact should not be nil")
	}
	if preview.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestChangeHighlight_Fields(t *testing.T) {
	change := ChangeHighlight{
		Type: ChangeTypeOption,
		Position: Range{
			Start: 10,
			End:   20,
		},
		Original:    "--output-type=csv",
		Replacement: "--output-type=json",
		Reason:      "出力フォーマット変更",
		RuleName:    "output-format-rule",
	}

	if change.Type != ChangeTypeOption {
		t.Errorf("Expected ChangeTypeOption, got %v", change.Type)
	}
	if change.Position.Start != 10 {
		t.Errorf("Expected start position 10, got %d", change.Position.Start)
	}
	if change.Position.End != 20 {
		t.Errorf("Expected end position 20, got %d", change.Position.End)
	}
	if change.Original != "--output-type=csv" {
		t.Errorf("Expected '--output-type=csv', got '%s'", change.Original)
	}
	if change.Replacement != "--output-type=json" {
		t.Errorf("Expected '--output-type=json', got '%s'", change.Replacement)
	}
}

func TestImpactAnalysis_RiskLevels(t *testing.T) {
	tests := []struct {
		risk     RiskLevel
		expected bool
	}{
		{RiskLow, true},
		{RiskMedium, true},
		{RiskHigh, true},
		{RiskCritical, true},
		{RiskLevel("invalid"), true}, // Should still be valid as it's just a string
	}

	for _, test := range tests {
		t.Run(string(test.risk), func(t *testing.T) {
			impact := &ImpactAnalysis{
				Risk:        test.risk,
				Description: "Test description",
				Complexity:  5,
			}

			if impact.Risk != test.risk {
				t.Errorf("Expected risk %v, got %v", test.risk, impact.Risk)
			}
		})
	}
}

func TestPreviewMetadata_Timing(t *testing.T) {
	start := time.Now()

	metadata := &PreviewMetadata{
		LineNumber:     42,
		GeneratedAt:    time.Now(),
		ProcessingTime: 100 * time.Millisecond,
		Version:        "1.9.0",
	}

	end := time.Now()

	if metadata.GeneratedAt.Before(start) || metadata.GeneratedAt.After(end) {
		t.Error("GeneratedAt should be between start and end times")
	}

	if metadata.ProcessingTime != 100*time.Millisecond {
		t.Errorf("Expected processing time 100ms, got %v", metadata.ProcessingTime)
	}

	if metadata.LineNumber != 42 {
		t.Errorf("Expected line number 42, got %d", metadata.LineNumber)
	}

	if metadata.Version != "1.9.0" {
		t.Errorf("Expected version '1.9.0', got '%s'", metadata.Version)
	}
}

func TestPreviewFilter_DefaultValues(t *testing.T) {
	filter := &PreviewFilter{
		ShowOnlyChanged: false,
		Categories:      []CommandCategory{},
		RiskLevels:      []RiskLevel{},
		SearchQuery:     "",
	}

	if filter.ShowOnlyChanged {
		t.Error("ShowOnlyChanged should default to false")
	}

	if len(filter.Categories) != 0 {
		t.Errorf("Categories should be empty by default, got %d", len(filter.Categories))
	}

	if len(filter.RiskLevels) != 0 {
		t.Errorf("RiskLevels should be empty by default, got %d", len(filter.RiskLevels))
	}

	if filter.SearchQuery != "" {
		t.Errorf("SearchQuery should be empty by default, got '%s'", filter.SearchQuery)
	}
}

func TestPreviewOptions_DefaultValues(t *testing.T) {
	options := &PreviewOptions{
		IncludeDescription: true,
		IncludeImpact:      true,
		IncludeWarnings:    true,
		MaxDescLength:      500,
		Timeout:            5 * time.Second,
	}

	if !options.IncludeDescription {
		t.Error("IncludeDescription should default to true")
	}

	if !options.IncludeImpact {
		t.Error("IncludeImpact should default to true")
	}

	if !options.IncludeWarnings {
		t.Error("IncludeWarnings should default to true")
	}

	if options.MaxDescLength != 500 {
		t.Errorf("Expected MaxDescLength 500, got %d", options.MaxDescLength)
	}

	if options.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", options.Timeout)
	}
}

func TestCommandCategory_Values(t *testing.T) {
	categories := []CommandCategory{
		CategoryServer,
		CategoryDatabase,
		CategoryNetwork,
		CategoryStorage,
		CategorySecurity,
		CategoryMonitoring,
		CategoryOther,
	}

	expectedValues := []string{
		"server",
		"database",
		"network",
		"storage",
		"security",
		"monitoring",
		"other",
	}

	if len(categories) != len(expectedValues) {
		t.Errorf("Expected %d categories, got %d", len(expectedValues), len(categories))
	}

	for i, category := range categories {
		if string(category) != expectedValues[i] {
			t.Errorf("Expected category %s, got %s", expectedValues[i], string(category))
		}
	}
}

func TestRange_Validation(t *testing.T) {
	tests := []struct {
		name  string
		start int
		end   int
		valid bool
	}{
		{"valid range", 0, 10, true},
		{"zero range", 5, 5, true},
		{"invalid range", 10, 5, false},
		{"negative start", -1, 5, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := Range{
				Start: test.start,
				End:   test.end,
			}

			isValid := r.Start >= 0 && r.End >= r.Start
			if isValid != test.valid {
				t.Errorf("Expected validity %v, got %v for range %d-%d",
					test.valid, isValid, test.start, test.end)
			}
		})
	}
}
