package preview

import (
	"strings"
	"testing"
	"time"

	"github.com/rivo/tview"
)

func TestNewWidget(t *testing.T) {
	widget := NewWidget()

	if widget == nil {
		t.Error("Widget should not be nil")
	}

	if widget.Flex == nil {
		t.Error("Flex container should not be nil")
	}

	if !widget.visible {
		t.Error("Widget should be visible by default")
	}

	// Check that all views are initialized
	if widget.originalView == nil {
		t.Error("Original view should be initialized")
	}

	if widget.transformedView == nil {
		t.Error("Transformed view should be initialized")
	}

	if widget.changesView == nil {
		t.Error("Changes view should be initialized")
	}

	if widget.impactView == nil {
		t.Error("Impact view should be initialized")
	}

	if widget.descriptionView == nil {
		t.Error("Description view should be initialized")
	}

	if widget.warningsView == nil {
		t.Error("Warnings view should be initialized")
	}
}

func TestWidget_SetVisible(t *testing.T) {
	widget := NewWidget()

	// Test setting visible to false
	widget.SetVisible(false)
	if widget.IsVisible() {
		t.Error("Widget should not be visible after SetVisible(false)")
	}

	// Test setting visible to true
	widget.SetVisible(true)
	if !widget.IsVisible() {
		t.Error("Widget should be visible after SetVisible(true)")
	}
}

func TestWidget_UpdatePreview_Nil(t *testing.T) {
	widget := NewWidget()

	// Test with nil preview
	widget.UpdatePreview(nil)

	if widget.GetCurrentPreview() != nil {
		t.Error("Current preview should be nil after updating with nil")
	}

	// Views should be cleared but not nil
	if widget.originalView == nil {
		t.Error("Original view should not be nil")
	}
}

func TestWidget_UpdatePreview_Valid(t *testing.T) {
	widget := NewWidget()

	preview := &CommandPreview{
		Original:    "usacloud server list --output-type=csv",
		Transformed: "usacloud server list --output-type=json",
		Changes: []ChangeHighlight{
			{
				Type:        ChangeTypeOption,
				Original:    "csv",
				Replacement: "json",
				Reason:      "出力フォーマット変更",
				RuleName:    "output-format",
			},
		},
		Description: "サーバ一覧取得コマンド",
		Impact: &ImpactAnalysis{
			Risk:        RiskLow,
			Description: "情報取得のため影響なし",
			Complexity:  2,
			Resources:   []string{"server"},
		},
		Warnings: []string{"出力形式が変更されます"},
		Category: "server",
		Metadata: &PreviewMetadata{
			LineNumber:     1,
			GeneratedAt:    time.Now(),
			ProcessingTime: 50 * time.Millisecond,
			Version:        "1.9.0",
		},
	}

	widget.UpdatePreview(preview)

	if widget.GetCurrentPreview() != preview {
		t.Error("Current preview should match the updated preview")
	}

	// Check that views were updated (we can't easily test the content without a full TUI setup)
	if widget.originalView == nil {
		t.Error("Original view should not be nil after update")
	}
}

func TestWidget_GetChangeColor(t *testing.T) {
	widget := NewWidget()

	tests := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeTypeOption, "green"},
		{ChangeTypeArgument, "blue"},
		{ChangeTypeCommand, "yellow"},
		{ChangeTypeFormat, "cyan"},
		{ChangeTypeRemoval, "red"},
		{ChangeTypeAddition, "green"},
		{ChangeType("unknown"), "white"},
	}

	for _, test := range tests {
		t.Run(string(test.changeType), func(t *testing.T) {
			color := widget.getChangeColor(test.changeType)
			if color != test.expected {
				t.Errorf("Expected color %s, got %s", test.expected, color)
			}
		})
	}
}

func TestWidget_GetRiskColor(t *testing.T) {
	widget := NewWidget()

	tests := []struct {
		risk     RiskLevel
		expected string
	}{
		{RiskLow, "green"},
		{RiskMedium, "yellow"},
		{RiskHigh, "red"},
		{RiskCritical, "magenta"},
		{RiskLevel("unknown"), "white"},
	}

	for _, test := range tests {
		t.Run(string(test.risk), func(t *testing.T) {
			color := widget.getRiskColor(test.risk)
			if color != test.expected {
				t.Errorf("Expected color %s, got %s", test.expected, color)
			}
		})
	}
}

func TestWidget_GetComplexityColor(t *testing.T) {
	widget := NewWidget()

	tests := []struct {
		complexity int
		expected   string
	}{
		{1, "green"},
		{3, "green"},
		{4, "yellow"},
		{6, "yellow"},
		{7, "red"},
		{8, "red"},
		{9, "magenta"},
		{10, "magenta"},
	}

	for _, test := range tests {
		t.Run(string(rune(test.complexity)), func(t *testing.T) {
			color := widget.getComplexityColor(test.complexity)
			if color != test.expected {
				t.Errorf("Expected color %s, got %s for complexity %d", test.expected, color, test.complexity)
			}
		})
	}
}

func TestWidget_HighlightSyntax(t *testing.T) {
	widget := NewWidget()

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "empty line",
			text:     "",
			expected: "[gray](空行)[white]",
		},
		{
			name:     "comment line",
			text:     "# This is a comment",
			expected: "[gray]# This is a comment[white]",
		},
		{
			name:     "usacloud command",
			text:     "usacloud server list",
			expected: "[cyan]usacloud[white] [yellow]server[white] list",
		},
		{
			name:     "non-usacloud command",
			text:     "echo hello",
			expected: "echo hello",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := widget.highlightSyntax(test.text, false)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestWidget_HighlightChanges(t *testing.T) {
	widget := NewWidget()

	text := "usacloud server list --output-type=json"
	changes := []ChangeHighlight{
		{
			Type:        ChangeTypeOption,
			Replacement: "json",
		},
	}

	result := widget.highlightChanges(text, changes)

	// Should contain the highlighted replacement
	if !strings.Contains(result, "[green]json[white]") {
		t.Errorf("Expected highlighted 'json', got '%s'", result)
	}
}

func TestWidget_SetApplication(t *testing.T) {
	widget := NewWidget()
	app := tview.NewApplication()

	widget.SetApplication(app)

	if widget.app != app {
		t.Error("Application should be set correctly")
	}
}

func TestWidget_GetFocusable(t *testing.T) {
	widget := NewWidget()

	focusable := widget.GetFocusable()

	expectedCount := 6 // All 6 sub-views should be focusable
	if len(focusable) != expectedCount {
		t.Errorf("Expected %d focusable items, got %d", expectedCount, len(focusable))
	}

	// Check that all views are included
	viewFound := make(map[string]bool)
	for _, item := range focusable {
		switch item {
		case widget.originalView:
			viewFound["original"] = true
		case widget.transformedView:
			viewFound["transformed"] = true
		case widget.changesView:
			viewFound["changes"] = true
		case widget.impactView:
			viewFound["impact"] = true
		case widget.descriptionView:
			viewFound["description"] = true
		case widget.warningsView:
			viewFound["warnings"] = true
		}
	}

	expectedViews := []string{"original", "transformed", "changes", "impact", "description", "warnings"}
	for _, view := range expectedViews {
		if !viewFound[view] {
			t.Errorf("View '%s' not found in focusable items", view)
		}
	}
}

func TestWidget_Focus(t *testing.T) {
	widget := NewWidget()
	app := tview.NewApplication()
	widget.SetApplication(app)

	// Test focusing different views
	views := []string{"original", "transformed", "changes", "impact", "description", "warnings"}

	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			// This won't change actual focus without a running app, but should not panic
			widget.Focus(view)
		})
	}

	// Test focusing invalid view (should not panic)
	widget.Focus("invalid")
}

func TestWidget_EmptyPreviewHandling(t *testing.T) {
	widget := NewWidget()

	// Test with preview containing empty/nil fields
	preview := &CommandPreview{
		Original:    "",
		Transformed: "",
		Changes:     []ChangeHighlight{},
		Description: "",
		Impact:      nil,
		Warnings:    []string{},
		Category:    "",
		Metadata:    nil,
	}

	// Should not panic
	widget.UpdatePreview(preview)

	if widget.GetCurrentPreview() != preview {
		t.Error("Should handle empty preview correctly")
	}
}

func TestWidget_LargeDataHandling(t *testing.T) {
	widget := NewWidget()

	// Test with large amounts of data
	largeText := strings.Repeat("usacloud server list ", 1000)
	manyChanges := make([]ChangeHighlight, 100)
	for i := range manyChanges {
		manyChanges[i] = ChangeHighlight{
			Type:        ChangeTypeOption,
			Reason:      "Test change " + string(rune(i)),
			Replacement: "test" + string(rune(i)),
		}
	}

	manyWarnings := make([]string, 50)
	for i := range manyWarnings {
		manyWarnings[i] = "Warning " + string(rune(i))
	}

	preview := &CommandPreview{
		Original:    largeText,
		Transformed: largeText,
		Changes:     manyChanges,
		Description: strings.Repeat("Description ", 200),
		Impact: &ImpactAnalysis{
			Risk:        RiskMedium,
			Description: "Large impact description",
			Resources:   []string{"server", "disk", "network"},
			Complexity:  8,
		},
		Warnings: manyWarnings,
		Category: "server",
	}

	// Should handle large data without panicking
	widget.UpdatePreview(preview)

	if widget.GetCurrentPreview() != preview {
		t.Error("Should handle large preview data correctly")
	}
}

func TestWidget_ConcurrentAccess(t *testing.T) {
	widget := NewWidget()

	// Test concurrent updates (basic safety check)
	done := make(chan bool, 2)

	preview1 := &CommandPreview{
		Original:    "usacloud server list",
		Transformed: "usacloud server list --output-type=json",
		Category:    "server",
	}

	preview2 := &CommandPreview{
		Original:    "usacloud database list",
		Transformed: "usacloud database list --output-type=json",
		Category:    "database",
	}

	go func() {
		for i := 0; i < 10; i++ {
			widget.UpdatePreview(preview1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			widget.UpdatePreview(preview2)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have one of the previews
	currentPreview := widget.GetCurrentPreview()
	if currentPreview != preview1 && currentPreview != preview2 {
		t.Error("Should have one of the test previews after concurrent updates")
	}
}
