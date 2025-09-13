package preview

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Widget represents the TUI preview widget
type Widget struct {
	*tview.Flex

	// UI components
	originalView    *tview.TextView
	transformedView *tview.TextView
	changesView     *tview.TextView
	impactView      *tview.TextView
	descriptionView *tview.TextView
	warningsView    *tview.TextView

	// State
	currentPreview *CommandPreview
	app            *tview.Application
	visible        bool
}

// NewWidget creates a new preview widget
func NewWidget() *Widget {
	pw := &Widget{
		Flex:    tview.NewFlex(),
		visible: true,
	}

	pw.setupViews()
	pw.layoutViews()

	return pw
}

// setupViews initializes all the sub-views
func (pw *Widget) setupViews() {
	// Original command view
	pw.originalView = tview.NewTextView()
	pw.originalView.SetDynamicColors(true).
		SetTitle("🔍 変換前").
		SetBorder(true).
		SetBorderColor(tcell.ColorGray)
	pw.originalView.SetWrap(true)

	// Transformed command view
	pw.transformedView = tview.NewTextView()
	pw.transformedView.SetDynamicColors(true).
		SetTitle("✨ 変換後").
		SetBorder(true).
		SetBorderColor(tcell.ColorGreen)
	pw.transformedView.SetWrap(true)

	// Changes detail view
	pw.changesView = tview.NewTextView()
	pw.changesView.SetDynamicColors(true).
		SetTitle("📋 変更詳細").
		SetBorder(true)
	pw.changesView.SetWrap(true)

	// Impact analysis view
	pw.impactView = tview.NewTextView()
	pw.impactView.SetDynamicColors(true).
		SetTitle("⚠️ 影響分析").
		SetBorder(true)
	pw.impactView.SetWrap(true)

	// Description view
	pw.descriptionView = tview.NewTextView()
	pw.descriptionView.SetDynamicColors(true).
		SetTitle("📖 コマンド説明").
		SetBorder(true)
	pw.descriptionView.SetWrap(true)

	// Warnings view
	pw.warningsView = tview.NewTextView()
	pw.warningsView.SetDynamicColors(true).
		SetTitle("⚠️ 注意事項").
		SetBorder(true)
	pw.warningsView.SetWrap(true)
}

// layoutViews arranges the views in the layout
func (pw *Widget) layoutViews() {
	// Top row: Original | Transformed
	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(pw.originalView, 0, 1, false).
		AddItem(pw.transformedView, 0, 1, false)

	// Middle row: Changes | Impact
	middleFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(pw.changesView, 0, 1, false).
		AddItem(pw.impactView, 0, 1, false)

	// Bottom row: Description | Warnings
	bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(pw.descriptionView, 0, 2, false).
		AddItem(pw.warningsView, 0, 1, false)

	// Main layout: Top | Middle | Bottom
	pw.Flex.SetDirection(tview.FlexRow).
		AddItem(topFlex, 0, 2, false).
		AddItem(middleFlex, 0, 2, false).
		AddItem(bottomFlex, 0, 1, false)
}

// UpdatePreview updates the widget with a new preview
func (pw *Widget) UpdatePreview(preview *CommandPreview) {
	pw.currentPreview = preview

	if preview == nil {
		pw.clearAllViews()
		return
	}

	pw.updateOriginalView()
	pw.updateTransformedView()
	pw.updateChangesView()
	pw.updateImpactView()
	pw.updateDescriptionView()
	pw.updateWarningsView()
}

// clearAllViews clears all views
func (pw *Widget) clearAllViews() {
	pw.originalView.Clear()
	pw.transformedView.Clear()
	pw.changesView.Clear()
	pw.impactView.Clear()
	pw.descriptionView.Clear()
	pw.warningsView.Clear()
}

// updateOriginalView updates the original command view
func (pw *Widget) updateOriginalView() {
	pw.originalView.Clear()

	if pw.currentPreview.Original == "" {
		_, _ = fmt.Fprintf(pw.originalView, "[gray]空行またはコメント[white]")
		return
	}

	// Highlight the original command
	highlighted := pw.highlightSyntax(pw.currentPreview.Original, false)
	_, _ = fmt.Fprintf(pw.originalView, "%s", highlighted)
}

// updateTransformedView updates the transformed command view
func (pw *Widget) updateTransformedView() {
	pw.transformedView.Clear()

	if pw.currentPreview.Transformed == "" {
		_, _ = fmt.Fprintf(pw.transformedView, "[gray]空行またはコメント[white]")
		return
	}

	// Highlight changes in the transformed command
	highlighted := pw.highlightChanges(pw.currentPreview.Transformed, pw.currentPreview.Changes)
	_, _ = fmt.Fprintf(pw.transformedView, "%s", highlighted)
}

// updateChangesView updates the changes detail view
func (pw *Widget) updateChangesView() {
	pw.changesView.Clear()

	if len(pw.currentPreview.Changes) == 0 {
		_, _ = fmt.Fprintf(pw.changesView, "[gray]変更はありません[white]")
		return
	}

	for i, change := range pw.currentPreview.Changes {
		color := pw.getChangeColor(change.Type)

		_, _ = fmt.Fprintf(pw.changesView, "[%s]%d. %s[white]\n", color, i+1, change.Reason)

		if change.RuleName != "" {
			_, _ = fmt.Fprintf(pw.changesView, "   [blue]ルール:[white] %s\n", change.RuleName)
		}

		if change.Original != "" {
			_, _ = fmt.Fprintf(pw.changesView, "   [red]削除:[white] %s\n", change.Original)
		}

		if change.Replacement != "" {
			_, _ = fmt.Fprintf(pw.changesView, "   [green]追加:[white] %s\n", change.Replacement)
		}

		_, _ = fmt.Fprintf(pw.changesView, "\n")
	}
}

// updateImpactView updates the impact analysis view
func (pw *Widget) updateImpactView() {
	pw.impactView.Clear()

	impact := pw.currentPreview.Impact
	if impact == nil {
		_, _ = fmt.Fprintf(pw.impactView, "[gray]影響分析情報がありません[white]")
		return
	}

	// Risk level with color
	riskColor := pw.getRiskColor(impact.Risk)
	_, _ = fmt.Fprintf(pw.impactView, "[%s]リスクレベル:[white] [%s]%s[white]\n\n",
		riskColor, riskColor, strings.ToUpper(string(impact.Risk)))

	// Description
	_, _ = fmt.Fprintf(pw.impactView, "[yellow]説明:[white]\n%s\n\n", impact.Description)

	// Complexity
	complexityColor := pw.getComplexityColor(impact.Complexity)
	_, _ = fmt.Fprintf(pw.impactView, "[%s]複雑度:[white] %d/10\n\n", complexityColor, impact.Complexity)

	// Resources
	if len(impact.Resources) > 0 {
		_, _ = fmt.Fprintf(pw.impactView, "[cyan]影響するリソース:[white]\n")
		for _, resource := range impact.Resources {
			_, _ = fmt.Fprintf(pw.impactView, "  • %s\n", resource)
		}
		_, _ = fmt.Fprintf(pw.impactView, "\n")
	}

	// Dependencies
	if len(impact.Dependencies) > 0 {
		_, _ = fmt.Fprintf(pw.impactView, "[magenta]依存関係:[white]\n")
		for _, dep := range impact.Dependencies {
			_, _ = fmt.Fprintf(pw.impactView, "  • %s\n", dep)
		}
	}
}

// updateDescriptionView updates the command description view
func (pw *Widget) updateDescriptionView() {
	pw.descriptionView.Clear()

	if pw.currentPreview.Description == "" {
		_, _ = fmt.Fprintf(pw.descriptionView, "[gray]説明がありません[white]")
		return
	}

	_, _ = fmt.Fprintf(pw.descriptionView, "[white]%s[white]", pw.currentPreview.Description)

	// Add metadata
	if pw.currentPreview.Metadata != nil {
		_, _ = fmt.Fprintf(pw.descriptionView, "\n\n[gray]--- メタデータ ---[white]\n")
		_, _ = fmt.Fprintf(pw.descriptionView, "[gray]行番号:[white] %d\n", pw.currentPreview.Metadata.LineNumber)
		_, _ = fmt.Fprintf(pw.descriptionView, "[gray]カテゴリ:[white] %s\n", pw.currentPreview.Category)
		_, _ = fmt.Fprintf(pw.descriptionView, "[gray]処理時間:[white] %v\n", pw.currentPreview.Metadata.ProcessingTime)
	}
}

// updateWarningsView updates the warnings view
func (pw *Widget) updateWarningsView() {
	pw.warningsView.Clear()

	if len(pw.currentPreview.Warnings) == 0 {
		_, _ = fmt.Fprintf(pw.warningsView, "[green]注意事項はありません[white]")
		return
	}

	for i, warning := range pw.currentPreview.Warnings {
		_, _ = fmt.Fprintf(pw.warningsView, "[yellow]%d.[white] %s\n\n", i+1, warning)
	}
}

// highlightChanges highlights changes in the transformed command
func (pw *Widget) highlightChanges(text string, changes []ChangeHighlight) string {
	highlighted := text

	// Sort changes by position to avoid conflicts
	sortedChanges := make([]ChangeHighlight, len(changes))
	copy(sortedChanges, changes)

	for _, change := range sortedChanges {
		if change.Replacement != "" {
			color := pw.getChangeColor(change.Type)
			highlighted = strings.ReplaceAll(highlighted, change.Replacement,
				fmt.Sprintf("[%s]%s[white]", color, change.Replacement))
		}
	}

	return pw.highlightSyntax(highlighted, true)
}

// highlightSyntax adds basic syntax highlighting
func (pw *Widget) highlightSyntax(text string, isTransformed bool) string {
	if strings.TrimSpace(text) == "" {
		return "[gray](空行)[white]"
	}

	if strings.HasPrefix(strings.TrimSpace(text), "#") {
		return fmt.Sprintf("[gray]%s[white]", text)
	}

	// Basic usacloud command highlighting
	if strings.HasPrefix(strings.TrimSpace(text), "usacloud") {
		parts := strings.Fields(text)
		if len(parts) >= 2 {
			highlighted := fmt.Sprintf("[cyan]%s[white] [yellow]%s[white]", parts[0], parts[1])
			if len(parts) > 2 {
				highlighted += " " + strings.Join(parts[2:], " ")
			}
			return highlighted
		}
	}

	return text
}

// getChangeColor returns the color for a change type
func (pw *Widget) getChangeColor(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeOption:
		return "green"
	case ChangeTypeArgument:
		return "blue"
	case ChangeTypeCommand:
		return "yellow"
	case ChangeTypeFormat:
		return "cyan"
	case ChangeTypeRemoval:
		return "red"
	case ChangeTypeAddition:
		return "green"
	default:
		return "white"
	}
}

// getRiskColor returns the color for a risk level
func (pw *Widget) getRiskColor(risk RiskLevel) string {
	switch risk {
	case RiskLow:
		return "green"
	case RiskMedium:
		return "yellow"
	case RiskHigh:
		return "red"
	case RiskCritical:
		return "magenta"
	default:
		return "white"
	}
}

// getComplexityColor returns the color for complexity level
func (pw *Widget) getComplexityColor(complexity int) string {
	switch {
	case complexity <= 3:
		return "green"
	case complexity <= 6:
		return "yellow"
	case complexity <= 8:
		return "red"
	default:
		return "magenta"
	}
}

// SetVisible sets the visibility of the preview widget
func (pw *Widget) SetVisible(visible bool) {
	pw.visible = visible
}

// IsVisible returns whether the preview widget is visible
func (pw *Widget) IsVisible() bool {
	return pw.visible
}

// GetCurrentPreview returns the current preview
func (pw *Widget) GetCurrentPreview() *CommandPreview {
	return pw.currentPreview
}

// Focus sets focus to a specific sub-view
func (pw *Widget) Focus(viewName string) {
	switch viewName {
	case "original":
		pw.app.SetFocus(pw.originalView)
	case "transformed":
		pw.app.SetFocus(pw.transformedView)
	case "changes":
		pw.app.SetFocus(pw.changesView)
	case "impact":
		pw.app.SetFocus(pw.impactView)
	case "description":
		pw.app.SetFocus(pw.descriptionView)
	case "warnings":
		pw.app.SetFocus(pw.warningsView)
	}
}

// SetApplication sets the tview application reference
func (pw *Widget) SetApplication(app *tview.Application) {
	pw.app = app
}

// GetFocusable returns all focusable views
func (pw *Widget) GetFocusable() []tview.Primitive {
	return []tview.Primitive{
		pw.originalView,
		pw.transformedView,
		pw.changesView,
		pw.impactView,
		pw.descriptionView,
		pw.warningsView,
	}
}
