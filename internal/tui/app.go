package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/sandbox"
	"github.com/armaniacs/usacloud-update/internal/transform"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CommandItem represents a converted command with its metadata
type CommandItem struct {
	Original   string
	Converted  string
	LineNumber int
	Changed    bool
	RuleName   string
	Selected   bool
	Result     *sandbox.ExecutionResult
}

// App represents the TUI application
type App struct {
	app      *tview.Application
	pages    *tview.Pages
	config   *config.SandboxConfig
	executor *sandbox.Executor
	commands []*CommandItem

	// UI components
	commandList *tview.List
	detailView  *tview.TextView
	resultView  *tview.TextView
	statusBar   *tview.TextView
	progressBar *tview.TextView
	helpText    *tview.TextView
	mainGrid    *tview.Grid

	// State
	currentIndex  int
	executedCount int
	totalSelected int
	helpVisible   bool
}

// NewApp creates a new TUI application
func NewApp(cfg *config.SandboxConfig) *App {
	app := &App{
		app:         tview.NewApplication(),
		config:      cfg,
		executor:    sandbox.NewExecutor(cfg),
		commands:    make([]*CommandItem, 0),
		helpVisible: true, // Default to visible
	}

	app.setupUI()
	return app
}

// LoadScript loads and converts a script for interactive execution
func (a *App) LoadScript(lines []string) error {
	engine := transform.NewDefaultEngine()

	for i, line := range lines {
		result := engine.Apply(line)

		item := &CommandItem{
			Original:   line,
			Converted:  result.Line,
			LineNumber: i + 1,
			Changed:    result.Changed,
			Selected:   false,
		}

		if result.Changed && len(result.Changes) > 0 {
			item.RuleName = result.Changes[0].RuleName
		}

		a.commands = append(a.commands, item)
	}

	a.refreshCommandList()
	return nil
}

// Run starts the TUI application
func (a *App) Run() error {
	return a.app.Run()
}

// Stop stops the TUI application
func (a *App) Stop() {
	a.app.Stop()
}

// setupUI initializes the UI components
func (a *App) setupUI() {
	a.setupCommandList()
	a.setupDetailView()
	a.setupResultView()
	a.setupStatusBar()
	a.setupProgressBar()
	a.setupHelpText()
	a.setupLayout()
	a.setupKeyBindings()
}

// setupCommandList initializes the command list widget
func (a *App) setupCommandList() {
	a.commandList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedFunc(a.onCommandSelected).
		SetChangedFunc(a.onSelectionChanged)

	a.commandList.SetTitle("📋 Converted Commands").SetBorder(true)
	a.commandList.SetTitleAlign(tview.AlignLeft)
}

// setupDetailView initializes the detail view widget
func (a *App) setupDetailView() {
	a.detailView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true)

	a.detailView.SetTitle("🔍 Command Details").SetBorder(true)
	a.detailView.SetTitleAlign(tview.AlignLeft)
}

// setupResultView initializes the result view widget
func (a *App) setupResultView() {
	a.resultView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(0)

	a.resultView.SetTitle("📊 Execution Results").SetBorder(true)
	a.resultView.SetTitleAlign(tview.AlignLeft)
}

// setupStatusBar initializes the status bar
func (a *App) setupStatusBar() {
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.updateStatusBar()
}

// setupProgressBar initializes the progress bar
func (a *App) setupProgressBar() {
	a.progressBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
}

// setupHelpText initializes the help text
func (a *App) setupHelpText() {
	helpContent := `[yellow]Key Bindings:[white]
[green]Enter[white] - Execute selected command    [green]Space[white] - Toggle selection    [green]a[white] - Select all
[green]n[white] - Select none                    [green]e[white] - Execute selected      [green]q[white] - Quit
[green]↑↓[white] - Navigate                    [green]Tab[white] - Switch panels      [green]?[white] - Toggle help`

	a.helpText = tview.NewTextView().
		SetText(helpContent).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	a.helpText.SetTitle("❓ Help").SetBorder(true)
}

// setupLayout creates the main layout
func (a *App) setupLayout() {
	// Create main grid layout
	a.mainGrid = tview.NewGrid()

	// Initial layout with help visible
	a.updateLayout()

	a.app.SetRoot(a.mainGrid, true)
}

// updateLayout updates the grid layout based on help visibility
func (a *App) updateLayout() {
	a.mainGrid.Clear()

	// Left column: command list
	leftPanel := tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(a.commandList, 0, 0, 1, 1, 0, 0, true)

	// Right column: detail and result views
	rightPanel := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(0).
		AddItem(a.detailView, 0, 0, 1, 1, 0, 0, false).
		AddItem(a.resultView, 1, 0, 1, 1, 0, 0, false)

	if a.helpVisible {
		// Layout with help: Main content, status bar, progress, help
		a.mainGrid.SetRows(0, 1, 3, 8).
			SetColumns(0, 0).
			SetBorders(false)

		a.mainGrid.AddItem(leftPanel, 0, 0, 1, 1, 0, 0, true).
			AddItem(rightPanel, 0, 1, 1, 1, 0, 0, false).
			AddItem(a.statusBar, 1, 0, 1, 2, 0, 0, false).
			AddItem(a.progressBar, 2, 0, 1, 2, 0, 0, false).
			AddItem(a.helpText, 3, 0, 1, 2, 0, 0, false)
	} else {
		// Layout without help: Main content, status bar, progress
		a.mainGrid.SetRows(0, 1, 3).
			SetColumns(0, 0).
			SetBorders(false)

		a.mainGrid.AddItem(leftPanel, 0, 0, 1, 1, 0, 0, true).
			AddItem(rightPanel, 0, 1, 1, 1, 0, 0, false).
			AddItem(a.statusBar, 1, 0, 1, 2, 0, 0, false).
			AddItem(a.progressBar, 2, 0, 1, 2, 0, 0, false)
	}
}

// setupKeyBindings configures global key bindings
func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case 'a':
			a.selectAll()
			return nil
		case 'n':
			a.selectNone()
			return nil
		case 'e':
			go a.executeSelected()
			return nil
		case '?':
			a.toggleHelp()
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}

		return event
	})
}

// refreshCommandList updates the command list display
func (a *App) refreshCommandList() {
	a.commandList.Clear()

	for _, cmd := range a.commands {
		prefix := "  "
		if cmd.Selected {
			prefix = "✓ "
		}

		color := "[white]"
		if cmd.Changed {
			color = "[yellow]"
		}
		if cmd.Result != nil {
			if cmd.Result.Success {
				color = "[green]"
			} else if !cmd.Result.Skipped {
				color = "[red]"
			}
		}

		mainText := fmt.Sprintf("%s%sL%d: %s", prefix, color, cmd.LineNumber, truncateString(cmd.Converted, 60))

		var secondaryText string
		if cmd.Changed {
			secondaryText = fmt.Sprintf("    [blue]%s[white]", cmd.RuleName)
		} else if strings.TrimSpace(cmd.Converted) == "" || strings.HasPrefix(strings.TrimSpace(cmd.Converted), "#") {
			secondaryText = "    [gray]Comment or empty line[white]"
		} else {
			secondaryText = "    [gray]No changes needed[white]"
		}

		a.commandList.AddItem(mainText, secondaryText, 0, nil)
	}

	a.updateStatusBar()
}

// onSelectionChanged handles list selection changes
func (a *App) onSelectionChanged(index int, mainText, secondaryText string, shortcut rune) {
	a.currentIndex = index
	a.updateDetailView()
}

// onCommandSelected handles command selection (Enter key)
func (a *App) onCommandSelected(index int, mainText, secondaryText string, shortcut rune) {
	if index >= 0 && index < len(a.commands) {
		cmd := a.commands[index]
		cmd.Selected = !cmd.Selected
		a.refreshCommandList()
		a.commandList.SetCurrentItem(index)
	}
}

// updateDetailView updates the detail view with current command info
func (a *App) updateDetailView() {
	if a.currentIndex < 0 || a.currentIndex >= len(a.commands) {
		a.detailView.Clear()
		return
	}

	cmd := a.commands[a.currentIndex]

	var content strings.Builder
	content.WriteString(fmt.Sprintf("[yellow]Line %d:[white]\n\n", cmd.LineNumber))

	if cmd.Original != cmd.Converted {
		content.WriteString("[red]Original:[white]\n")
		content.WriteString(fmt.Sprintf("  %s\n\n", cmd.Original))
		content.WriteString("[green]Converted:[white]\n")
		content.WriteString(fmt.Sprintf("  %s\n\n", cmd.Converted))
		content.WriteString(fmt.Sprintf("[blue]Rule:[white] %s\n\n", cmd.RuleName))
	} else {
		content.WriteString("[white]Command:[white]\n")
		content.WriteString(fmt.Sprintf("  %s\n\n", cmd.Converted))
	}

	if cmd.Result != nil {
		content.WriteString("[yellow]Execution Result:[white]\n")
		if cmd.Result.Skipped {
			content.WriteString(fmt.Sprintf("[yellow]Skipped:[white] %s\n", cmd.Result.SkipReason))
		} else if cmd.Result.Success {
			content.WriteString("[green]Status:[white] Success\n")
			content.WriteString(fmt.Sprintf("[cyan]Duration:[white] %v\n", cmd.Result.Duration))
		} else {
			content.WriteString("[red]Status:[white] Failed\n")
			content.WriteString(fmt.Sprintf("[red]Error:[white] %s\n", cmd.Result.Error))
		}
	}

	a.detailView.SetText(content.String())
}

// updateStatusBar updates the status bar text
func (a *App) updateStatusBar() {
	selected := 0
	executed := 0
	successful := 0

	for _, cmd := range a.commands {
		if cmd.Selected {
			selected++
		}
		if cmd.Result != nil && !cmd.Result.Skipped {
			executed++
			if cmd.Result.Success {
				successful++
			}
		}
	}

	status := fmt.Sprintf("[yellow]Commands:[white] %d  [green]Selected:[white] %d  [blue]Executed:[white] %d  [green]Successful:[white] %d",
		len(a.commands), selected, executed, successful)

	if a.config.DryRun {
		status += "  [red]DRY RUN MODE[white]"
	}

	a.statusBar.SetText(status)
}

// selectAll selects all usacloud commands
func (a *App) selectAll() {
	for _, cmd := range a.commands {
		trimmed := strings.TrimSpace(cmd.Converted)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && strings.HasPrefix(trimmed, "usacloud ") {
			cmd.Selected = true
		}
	}
	a.refreshCommandList()
}

// selectNone deselects all commands
func (a *App) selectNone() {
	for _, cmd := range a.commands {
		cmd.Selected = false
	}
	a.refreshCommandList()
}

// executeSelected executes all selected commands
func (a *App) executeSelected() {
	selected := make([]*CommandItem, 0)
	for _, cmd := range a.commands {
		if cmd.Selected {
			selected = append(selected, cmd)
		}
	}

	if len(selected) == 0 {
		return
	}

	a.updateProgressBar(0, len(selected), "Preparing execution...")

	for i, cmd := range selected {
		a.updateProgressBar(i+1, len(selected), fmt.Sprintf("Executing: %s", truncateString(cmd.Converted, 40)))

		result, _ := a.executor.ExecuteCommand(cmd.Converted)
		cmd.Result = result

		a.app.QueueUpdateDraw(func() {
			a.refreshCommandList()
			a.updateDetailView()
			a.updateResultView()
		})

		// Small delay for visual feedback
		time.Sleep(100 * time.Millisecond)
	}

	a.updateProgressBar(len(selected), len(selected), "Execution completed")
}

// updateProgressBar updates the progress bar
func (a *App) updateProgressBar(current, total int, message string) {
	if total == 0 {
		a.progressBar.SetText("")
		return
	}

	percentage := float64(current) / float64(total) * 100
	barWidth := 40
	filledWidth := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	text := fmt.Sprintf("[green]Progress:[white] [%s] %.1f%% (%d/%d) - %s",
		bar, percentage, current, total, message)

	a.app.QueueUpdateDraw(func() {
		a.progressBar.SetText(text)
	})
}

// updateResultView updates the result summary view
func (a *App) updateResultView() {
	var content strings.Builder

	executed := 0
	successful := 0
	failed := 0
	skipped := 0

	for _, cmd := range a.commands {
		if cmd.Result != nil {
			if cmd.Result.Skipped {
				skipped++
			} else {
				executed++
				if cmd.Result.Success {
					successful++
				} else {
					failed++
				}
			}
		}
	}

	content.WriteString("[yellow]Execution Summary:[white]\n\n")
	content.WriteString(fmt.Sprintf("[blue]Total Executed:[white] %d\n", executed))
	content.WriteString(fmt.Sprintf("[green]Successful:[white] %d\n", successful))
	content.WriteString(fmt.Sprintf("[red]Failed:[white] %d\n", failed))
	content.WriteString(fmt.Sprintf("[yellow]Skipped:[white] %d\n\n", skipped))

	if failed > 0 {
		content.WriteString("[red]Failed Commands:[white]\n")
		for _, cmd := range a.commands {
			if cmd.Result != nil && !cmd.Result.Success && !cmd.Result.Skipped {
				content.WriteString(fmt.Sprintf("  L%d: %s\n", cmd.LineNumber, truncateString(cmd.Converted, 50)))
				content.WriteString(fmt.Sprintf("    [red]Error:[white] %s\n\n", cmd.Result.Error))
			}
		}
	}

	a.resultView.SetText(content.String())
}

// toggleHelp toggles the visibility of the help text
func (a *App) toggleHelp() {
	a.helpVisible = !a.helpVisible

	// Update help text title to reflect current state
	if a.helpVisible {
		a.helpText.SetTitle("❓ Help")
	} else {
		a.helpText.SetTitle("❓ Help (Hidden)")
	}

	// Re-layout the UI
	a.updateLayout()

	// Force redraw
	a.app.Draw()
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
