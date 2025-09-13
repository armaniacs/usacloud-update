package filter

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FilterPanel represents the TUI component for filter management
type FilterPanel struct {
	*tview.Flex

	// UI components
	searchInput    *tview.InputField
	categoryList   *tview.List
	statusForm     *tview.Form
	presetDropdown *tview.DropDown
	clearButton    *tview.Button
	summaryText    *tview.TextView

	// State
	filterSystem *FilterSystem
	onUpdate     func()
	visible      bool
}

// NewFilterPanel creates a new filter panel
func NewFilterPanel(fs *FilterSystem) *FilterPanel {
	fp := &FilterPanel{
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		filterSystem: fs,
		visible:      true,
	}

	fp.setupComponents()
	fp.layoutComponents()
	fp.setupKeyBindings()

	return fp
}

// setupComponents initializes all UI components
func (fp *FilterPanel) setupComponents() {
	// Search input field
	fp.searchInput = tview.NewInputField()
	fp.searchInput.SetLabel("🔍 検索: ").
		SetPlaceholder("コマンド、説明、出力で検索...").
		SetChangedFunc(fp.onSearchChanged)
	fp.searchInput.SetBorder(true).SetTitle("テキスト検索")

	// Category list
	fp.categoryList = tview.NewList()
	fp.categoryList.SetTitle("📂 カテゴリ").SetBorder(true)
	fp.setupCategoryList()

	// Status form with checkboxes
	fp.statusForm = tview.NewForm()
	fp.statusForm.SetTitle("📊 ステータス").SetBorder(true)
	fp.setupStatusForm()

	// Preset dropdown
	fp.presetDropdown = tview.NewDropDown()
	fp.presetDropdown.SetLabel("💾 プリセット: ").
		SetOptions([]string{"<なし>"}, nil)
	fp.presetDropdown.SetBorder(true).SetTitle("プリセット管理")

	// Clear button
	fp.clearButton = tview.NewButton("🗑️  すべてクリア")
	fp.clearButton.SetSelectedFunc(fp.onClearAll)

	// Summary text
	fp.summaryText = tview.NewTextView()
	fp.summaryText.SetDynamicColors(true).
		SetTitle("📈 フィルタ状況").
		SetBorder(true)
	fp.updateSummary()
}

// setupCategoryList configures the category list
func (fp *FilterPanel) setupCategoryList() {
	categoryFilter := fp.filterSystem.GetFilter("category").(*CategoryFilter)
	categories := categoryFilter.GetAvailableCategories()

	for i, category := range categories {
		fp.categoryList.AddItem(
			fp.getCategoryDisplayName(category),
			fp.getCategoryDescription(category),
			rune('1'+i),
			fp.onCategoryToggle(category),
		)
	}
}

// setupStatusForm configures the status form
func (fp *FilterPanel) setupStatusForm() {
	statusFilter := fp.filterSystem.GetFilter("status").(*StatusFilter)
	statuses := statusFilter.GetAllStatuses()

	for _, status := range statuses {
		displayName := statusFilter.GetStatusDisplayName(status)
		icon := statusFilter.GetStatusIcon(status)
		label := fmt.Sprintf("%s %s", icon, displayName)

		fp.statusForm.AddCheckbox(
			label,
			statusFilter.IsStatusAllowed(status),
			fp.onStatusToggle(status),
		)
	}
}

// layoutComponents arranges the UI components
func (fp *FilterPanel) layoutComponents() {
	// Top section: Search and clear
	topFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(fp.searchInput, 0, 3, true).
		AddItem(fp.clearButton, 15, 0, false)

	// Middle section: Category and Status
	middleFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(fp.categoryList, 0, 1, false).
		AddItem(fp.statusForm, 0, 1, false)

	// Bottom section: Preset and Summary
	bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(fp.presetDropdown, 0, 1, false).
		AddItem(fp.summaryText, 0, 1, false)

	// Main layout
	fp.Flex.AddItem(topFlex, 3, 0, true).
		AddItem(middleFlex, 0, 2, false).
		AddItem(bottomFlex, 6, 0, false)
}

// setupKeyBindings configures key bindings
func (fp *FilterPanel) setupKeyBindings() {
	fp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlR:
			fp.resetAllFilters()
			return nil
		case tcell.KeyCtrlS:
			fp.showSavePresetDialog()
			return nil
		case tcell.KeyF5:
			fp.refreshData()
			return nil
		}

		switch event.Rune() {
		case '/':
			// Focus search input (handled by app)
			return nil
		case 'c':
			// Focus category list (handled by app)
			return nil
		case 's':
			// Focus status form (handled by app)
			return nil
		case 'p':
			// Focus preset dropdown (handled by app)
			return nil
		}

		return event
	})
}

// Event handlers
func (fp *FilterPanel) onSearchChanged(text string) {
	textFilter := fp.filterSystem.GetFilter("text-search").(*TextSearchFilter)
	textFilter.SetSearchTerm(text)
	fp.updateSummary()
	fp.triggerUpdate()
}

func (fp *FilterPanel) onCategoryToggle(category string) func() {
	return func() {
		categoryFilter := fp.filterSystem.GetFilter("category").(*CategoryFilter)
		categoryFilter.ToggleCategory(category)
		fp.updateCategoryList()
		fp.updateSummary()
		fp.triggerUpdate()
	}
}

func (fp *FilterPanel) onStatusToggle(status ExecutionStatus) func(bool) {
	return func(checked bool) {
		statusFilter := fp.filterSystem.GetFilter("status").(*StatusFilter)
		statusFilter.SetStatusAllowed(status, checked)
		fp.updateSummary()
		fp.triggerUpdate()
	}
}

func (fp *FilterPanel) onClearAll() {
	fp.filterSystem.ClearAllFilters()
	fp.searchInput.SetText("")
	fp.updateCategoryList()
	fp.updateStatusForm()
	fp.updateSummary()
	fp.triggerUpdate()
}

// Update methods
func (fp *FilterPanel) updateCategoryList() {
	categoryFilter := fp.filterSystem.GetFilter("category").(*CategoryFilter)

	for i := 0; i < fp.categoryList.GetItemCount(); i++ {
		mainText, _ := fp.categoryList.GetItemText(i)
		category := fp.getCategoryFromDisplayName(mainText)

		if categoryFilter.IsCategorySelected(category) {
			fp.categoryList.SetItemText(i, "✓ "+mainText, "")
		} else {
			// Remove checkmark if present
			if len(mainText) > 2 && mainText[:2] == "✓ " {
				fp.categoryList.SetItemText(i, mainText[2:], "")
			}
		}
	}
}

func (fp *FilterPanel) updateStatusForm() {
	statusFilter := fp.filterSystem.GetFilter("status").(*StatusFilter)

	// Clear and rebuild form
	fp.statusForm.Clear(true)
	statuses := statusFilter.GetAllStatuses()

	for _, status := range statuses {
		displayName := statusFilter.GetStatusDisplayName(status)
		icon := statusFilter.GetStatusIcon(status)
		label := fmt.Sprintf("%s %s", icon, displayName)

		fp.statusForm.AddCheckbox(
			label,
			statusFilter.IsStatusAllowed(status),
			fp.onStatusToggle(status),
		)
	}
}

func (fp *FilterPanel) updateSummary() {
	activeFilters := fp.filterSystem.GetActiveFilters()

	var summary string
	summary += fmt.Sprintf("[yellow]アクティブフィルタ:[white] %d\n\n", len(activeFilters))

	for _, filter := range activeFilters {
		summary += fmt.Sprintf("• [green]%s[white]\n", filter.Name())

		switch f := filter.(type) {
		case *TextSearchFilter:
			term := f.GetSearchTerm()
			if len(term) > 20 {
				term = term[:20] + "..."
			}
			summary += fmt.Sprintf("  検索: [cyan]%s[white]\n", term)

		case *CategoryFilter:
			categories := f.GetSelectedCategories()
			if len(categories) > 0 {
				summary += fmt.Sprintf("  カテゴリ: [cyan]%d個[white]\n", len(categories))
			}

		case *StatusFilter:
			statuses := f.GetAllowedStatuses()
			summary += fmt.Sprintf("  ステータス: [cyan]%d個[white]\n", len(statuses))
		}
		summary += "\n"
	}

	if len(activeFilters) == 0 {
		summary += "[gray]フィルタは適用されていません[white]"
	}

	fp.summaryText.SetText(summary)
}

// Helper methods
func (fp *FilterPanel) getCategoryDisplayName(category string) string {
	displayNames := map[string]string{
		"infrastructure":  "🏗️  インフラ",
		"storage":         "💾 ストレージ",
		"network":         "🌐 ネットワーク",
		"managed-service": "⚙️  マネージドサービス",
		"security":        "🔒 セキュリティ",
		"monitoring":      "📊 監視",
		"other":           "📦 その他",
	}

	if display, exists := displayNames[category]; exists {
		return display
	}
	return category
}

func (fp *FilterPanel) getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"infrastructure":  "サーバ、ディスク、ネットワーク機器",
		"storage":         "アーカイブ、ISO、ノート",
		"network":         "DNS、ロードバランサ、VPC",
		"managed-service": "データベース、NFS",
		"security":        "証明書、SSH鍵、認証",
		"monitoring":      "モニタ、アラート、メトリクス",
		"other":           "その他のコマンド",
	}

	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return ""
}

func (fp *FilterPanel) getCategoryFromDisplayName(displayName string) string {
	// Remove checkmark and icon
	name := displayName
	if len(name) > 2 && name[:2] == "✓ " {
		name = name[2:]
	}

	reverseMap := map[string]string{
		"🏗️  インフラ":      "infrastructure",
		"💾 ストレージ":       "storage",
		"🌐 ネットワーク":      "network",
		"⚙️  マネージドサービス": "managed-service",
		"🔒 セキュリティ":      "security",
		"📊 監視":          "monitoring",
		"📦 その他":         "other",
	}

	if category, exists := reverseMap[name]; exists {
		return category
	}
	return name
}

func (fp *FilterPanel) triggerUpdate() {
	if fp.onUpdate != nil {
		fp.onUpdate()
	}
}

func (fp *FilterPanel) resetAllFilters() {
	fp.onClearAll()
}

func (fp *FilterPanel) showSavePresetDialog() {
	// TODO: Implement preset save dialog
}

func (fp *FilterPanel) refreshData() {
	fp.updateCategoryList()
	fp.updateStatusForm()
	fp.updateSummary()
	fp.triggerUpdate()
}

// Public interface methods
func (fp *FilterPanel) SetOnUpdate(callback func()) {
	fp.onUpdate = callback
}

func (fp *FilterPanel) SetVisible(visible bool) {
	fp.visible = visible
}

func (fp *FilterPanel) IsVisible() bool {
	return fp.visible
}

func (fp *FilterPanel) GetFocusable() []tview.Primitive {
	return []tview.Primitive{
		fp.searchInput,
		fp.categoryList,
		fp.statusForm,
		fp.presetDropdown,
		fp.clearButton,
	}
}

func (fp *FilterPanel) FocusSearch() {
	// Focus is handled by the calling application
}

func (fp *FilterPanel) GetSearchTerm() string {
	return fp.searchInput.GetText()
}

func (fp *FilterPanel) SetSearchTerm(term string) {
	fp.searchInput.SetText(term)
	fp.onSearchChanged(term)
}

func (fp *FilterPanel) GetActiveFilterCount() int {
	return len(fp.filterSystem.GetActiveFilters())
}
