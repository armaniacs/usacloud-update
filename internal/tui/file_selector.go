package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/scanner"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Preview display constants for PBI-034
const (
	MinPreviewLines = 10  // 最小表示行数
	MaxPreviewLines = 100 // 最大表示行数
	HeaderLines     = 10  // ヘッダー情報の行数
	MarginLines     = 2   // 余白の行数
)

// FileSelector represents a file selection TUI
type FileSelector struct {
	app           *tview.Application
	config        *config.SandboxConfig
	scanner       *scanner.Scanner
	scanResult    *scanner.BasicScanResult
	selectedFiles []string

	// UI components
	fileList      *tview.List
	previewPane   *tview.TextView
	statusBar     *tview.TextView
	helpText      *tview.TextView
	previewNotice *tview.TextView
	mainGrid      *tview.Grid

	// State
	helpVisible bool

	// Callbacks
	onFilesSelected func([]string)
	onCancel        func()
}

// NewFileSelector creates a new file selector
func NewFileSelector(cfg *config.SandboxConfig) *FileSelector {
	fs := &FileSelector{
		app:           tview.NewApplication(),
		config:        cfg,
		scanner:       scanner.NewScanner(),
		selectedFiles: make([]string, 0),
		helpVisible:   true, // Default to visible
	}

	fs.setupUI()
	return fs
}

// SetOnFilesSelected sets the callback for when files are selected
func (fs *FileSelector) SetOnFilesSelected(callback func([]string)) {
	fs.onFilesSelected = callback
}

// SetOnCancel sets the callback for when the user cancels
func (fs *FileSelector) SetOnCancel(callback func()) {
	fs.onCancel = callback
}

// Run starts the file selector and scans the specified directory
func (fs *FileSelector) Run(directory string) error {
	// Scan directory for files
	result, err := fs.scanner.Scan(directory)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	fs.scanResult = result
	fs.populateFileList()
	fs.updateStatusBar()

	return fs.app.Run()
}

// Stop stops the file selector
func (fs *FileSelector) Stop() {
	fs.app.Stop()
}

// setupUI initializes the UI components
func (fs *FileSelector) setupUI() {
	fs.setupFileList()
	fs.setupPreviewPane()
	fs.setupStatusBar()
	fs.setupHelpText()
	fs.setupPreviewNotice()
	fs.setupLayout()
	fs.setupKeyBindings()
}

// setupFileList initializes the file list widget
func (fs *FileSelector) setupFileList() {
	fs.fileList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedFunc(fs.onFileToggle).
		SetChangedFunc(fs.onSelectionChanged)

	fs.fileList.SetTitle("📁 Script Files").SetBorder(true)
	fs.fileList.SetTitleAlign(tview.AlignLeft)
}

// setupPreviewPane initializes the preview pane
func (fs *FileSelector) setupPreviewPane() {
	fs.previewPane = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)

	fs.previewPane.SetTitle("👁️ Preview").SetBorder(true)
	fs.previewPane.SetTitleAlign(tview.AlignLeft)
}

// setupStatusBar initializes the status bar
func (fs *FileSelector) setupStatusBar() {
	fs.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
}

// setupHelpText initializes the help text
func (fs *FileSelector) setupHelpText() {
	helpContent := `[yellow]Key Bindings:[white]
[green]Space[white] - Select/deselect file    [green]Enter[white] - Confirm selection    [green]a[white] - Select all
[green]n[white] - Select none                [green]u[white] - Toggle usacloud files   [green]q[white] - Cancel
[green]↑↓[white] - Navigate                 [green]Tab[white] - Switch panes         [green]?[white] - Toggle help`

	fs.helpText = tview.NewTextView().
		SetText(helpContent).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	fs.helpText.SetTitle("❓ Help").SetBorder(true)
}

// setupPreviewNotice initializes the preview notice text
func (fs *FileSelector) setupPreviewNotice() {
	fs.previewNotice = tview.NewTextView().
		SetText("[black:yellow:b] TUIはPreviewとして提供中 [::-]").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
}

// setupLayout creates the main layout
func (fs *FileSelector) setupLayout() {
	// Create main grid layout
	fs.mainGrid = tview.NewGrid()

	// Initial layout with help visible
	fs.updateLayout()

	fs.app.SetRoot(fs.mainGrid, true)
}

// updateLayout updates the grid layout based on help visibility
func (fs *FileSelector) updateLayout() {
	fs.mainGrid.Clear()

	if fs.helpVisible {
		// Layout with help: Main content, status bar, help, preview notice
		fs.mainGrid.SetRows(0, 1, 4, 1).
			SetColumns(0, 0).
			SetBorders(false)

		fs.mainGrid.AddItem(fs.fileList, 0, 0, 1, 1, 0, 0, true).
			AddItem(fs.previewPane, 0, 1, 1, 1, 0, 0, false).
			AddItem(fs.statusBar, 1, 0, 1, 2, 0, 0, false).
			AddItem(fs.helpText, 2, 0, 1, 2, 0, 0, false).
			AddItem(fs.previewNotice, 3, 0, 1, 2, 0, 0, false)
	} else {
		// Layout without help: Main content, status bar, preview notice
		fs.mainGrid.SetRows(0, 1, 1).
			SetColumns(0, 0).
			SetBorders(false)

		fs.mainGrid.AddItem(fs.fileList, 0, 0, 1, 1, 0, 0, true).
			AddItem(fs.previewPane, 0, 1, 1, 1, 0, 0, false).
			AddItem(fs.statusBar, 1, 0, 1, 2, 0, 0, false).
			AddItem(fs.previewNotice, 2, 0, 1, 2, 0, 0, false)
	}
}

// setupKeyBindings configures global key bindings
func (fs *FileSelector) setupKeyBindings() {
	fs.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			if fs.onCancel != nil {
				fs.onCancel()
			}
			fs.app.Stop()
			return nil
		case 'a':
			fs.selectAll()
			return nil
		case 'n':
			fs.selectNone()
			return nil
		case 'u':
			fs.toggleUsacloudFiles()
			return nil
		case ' ':
			// Let the list handle space for selection
			return event
		case '?':
			fs.toggleHelp()
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			if fs.onCancel != nil {
				fs.onCancel()
			}
			fs.app.Stop()
			return nil
		case tcell.KeyEnter:
			fs.confirmSelection()
			return nil
		}

		return event
	})
}

// populateFileList populates the file list with scanned files
func (fs *FileSelector) populateFileList() {
	fs.fileList.Clear()

	if fs.scanResult == nil || len(fs.scanResult.Files) == 0 {
		fs.fileList.AddItem("[red]No script files found[white]",
			"    No .sh or .bash files found in the current directory", 0, nil)
		return
	}

	for _, file := range fs.scanResult.Files {
		fs.addFileToList(file)
	}
}

// addFileToList adds a single file to the list
func (fs *FileSelector) addFileToList(file *scanner.FileInfo) {
	// Check if file is selected
	isSelected := fs.isFileSelected(file.Path)
	prefix := "  "
	if isSelected {
		prefix = "✓ "
	}

	// Create main text
	relPath := file.GetRelativePath(fs.scanResult.Directory)
	mainText := fmt.Sprintf("%s[white]%s", prefix, relPath)

	// Create secondary text with file info
	var badges []string

	if file.IsExecutable {
		badges = append(badges, "[green]executable[white]")
	}

	// Check if file has usacloud commands
	if hasUsacloud, err := file.HasUsacloudCommands(); err == nil && hasUsacloud {
		badges = append(badges, "[yellow]usacloud[white]")
	}

	secondaryText := fmt.Sprintf("    %s  %s  %s",
		file.FormatSize(),
		file.FormatModTime(),
		strings.Join(badges, " "))

	fs.fileList.AddItem(mainText, secondaryText, 0, nil)
}

// onFileToggle handles file selection toggle
func (fs *FileSelector) onFileToggle(index int, mainText, secondaryText string, shortcut rune) {
	if fs.scanResult == nil || index >= len(fs.scanResult.Files) {
		return
	}

	file := fs.scanResult.Files[index]

	// Toggle selection
	if fs.isFileSelected(file.Path) {
		fs.removeSelectedFile(file.Path)
	} else {
		fs.selectedFiles = append(fs.selectedFiles, file.Path)
	}

	// Refresh the list to update selection indicators
	fs.populateFileList()
	fs.fileList.SetCurrentItem(index)
	fs.updateStatusBar()
}

// onSelectionChanged handles list selection changes for preview
func (fs *FileSelector) onSelectionChanged(index int, mainText, secondaryText string, shortcut rune) {
	if fs.scanResult == nil || index >= len(fs.scanResult.Files) {
		fs.previewPane.Clear()
		return
	}

	file := fs.scanResult.Files[index]
	fs.updatePreview(file)
}

// updatePreview updates the preview pane with file content
func (fs *FileSelector) updatePreview(file *scanner.FileInfo) {
	var content strings.Builder

	// File information
	content.WriteString(fmt.Sprintf("[yellow]File:[white] %s\n", file.Name))
	content.WriteString(fmt.Sprintf("[yellow]Path:[white] %s\n", file.GetRelativePath(fs.scanResult.Directory)))
	content.WriteString(fmt.Sprintf("[yellow]Size:[white] %s\n", file.FormatSize()))
	content.WriteString(fmt.Sprintf("[yellow]Modified:[white] %s\n", file.FormatModTime()))

	if file.IsExecutable {
		content.WriteString("[yellow]Permissions:[white] [green]executable[white]\n")
	}

	// Check usacloud commands
	if hasUsacloud, err := file.HasUsacloudCommands(); err == nil {
		if hasUsacloud {
			content.WriteString("[yellow]Contains:[white] [yellow]usacloud commands[white]\n")
		} else {
			content.WriteString("[yellow]Contains:[white] [gray]no usacloud commands[white]\n")
		}
	}

	content.WriteString("\n[yellow]Preview:[white]\n")
	content.WriteString("[gray]" + strings.Repeat("─", 50) + "[white]\n")

	// File preview with dynamic line calculation
	dynamicLines := fs.calculateDynamicPreviewLines()
	preview, err := file.Preview(dynamicLines)
	if err != nil {
		content.WriteString(fmt.Sprintf("[red]Error reading file: %v[white]\n", err))
	} else {
		for i, line := range preview {
			if i >= dynamicLines { // Limit preview lines dynamically
				break
			}
			content.WriteString(fmt.Sprintf("%s\n", line))
		}

		// Show truncated message only if file has more content than displayed
		if len(preview) == dynamicLines {
			// Check if file actually has more lines
			fileContent, readErr := os.ReadFile(file.Path)
			if readErr == nil {
				totalLines := len(strings.Split(string(fileContent), "\n"))
				if totalLines > dynamicLines {
					content.WriteString("[gray]... (truncated)[white]\n")
				}
			}
		}
	}

	fs.previewPane.SetText(content.String())
}

// calculateDynamicPreviewLines calculates optimal preview lines based on available height
func (fs *FileSelector) calculateDynamicPreviewLines() int {
	// Get the actual height of the preview pane
	_, _, _, height := fs.previewPane.GetRect()

	// Calculate available lines for content
	availableLines := height - HeaderLines - MarginLines

	// Apply min/max constraints
	if availableLines < MinPreviewLines {
		availableLines = MinPreviewLines
	}
	if availableLines > MaxPreviewLines {
		availableLines = MaxPreviewLines
	}

	return availableLines
}

// updateStatusBar updates the status bar text
func (fs *FileSelector) updateStatusBar() {
	if fs.scanResult == nil {
		fs.statusBar.SetText("[gray]Scanning...[white]")
		return
	}

	totalFiles := len(fs.scanResult.Files)
	selectedCount := len(fs.selectedFiles)

	var usacloudCount int
	for _, file := range fs.scanResult.Files {
		if hasUsacloud, err := file.HasUsacloudCommands(); err == nil && hasUsacloud {
			usacloudCount++
		}
	}

	status := fmt.Sprintf(
		"[yellow]Directory:[white] %s  [blue]Files:[white] %d  [green]Selected:[white] %d  [yellow]usacloud:[white] %d",
		fs.scanResult.Directory, totalFiles, selectedCount, usacloudCount)

	if len(fs.scanResult.Errors) > 0 {
		status += fmt.Sprintf("  [red]Errors:[white] %d", len(fs.scanResult.Errors))
	}

	fs.statusBar.SetText(status)
}

// isFileSelected checks if a file is selected
func (fs *FileSelector) isFileSelected(filePath string) bool {
	for _, selected := range fs.selectedFiles {
		if selected == filePath {
			return true
		}
	}
	return false
}

// removeSelectedFile removes a file from selection
func (fs *FileSelector) removeSelectedFile(filePath string) {
	for i, selected := range fs.selectedFiles {
		if selected == filePath {
			fs.selectedFiles = append(fs.selectedFiles[:i], fs.selectedFiles[i+1:]...)
			break
		}
	}
}

// selectAll selects all files
func (fs *FileSelector) selectAll() {
	if fs.scanResult == nil {
		return
	}

	fs.selectedFiles = make([]string, 0, len(fs.scanResult.Files))
	for _, file := range fs.scanResult.Files {
		fs.selectedFiles = append(fs.selectedFiles, file.Path)
	}

	fs.populateFileList()
	fs.updateStatusBar()
}

// selectNone deselects all files
func (fs *FileSelector) selectNone() {
	fs.selectedFiles = make([]string, 0)
	fs.populateFileList()
	fs.updateStatusBar()
}

// toggleUsacloudFiles selects/deselects files containing usacloud commands
func (fs *FileSelector) toggleUsacloudFiles() {
	if fs.scanResult == nil {
		return
	}

	// Check if any usacloud files are currently selected
	var usacloudFiles []string
	for _, file := range fs.scanResult.Files {
		if hasUsacloud, err := file.HasUsacloudCommands(); err == nil && hasUsacloud {
			usacloudFiles = append(usacloudFiles, file.Path)
		}
	}

	// Check if all usacloud files are already selected
	allUsacloudSelected := true
	for _, usacloudFile := range usacloudFiles {
		if !fs.isFileSelected(usacloudFile) {
			allUsacloudSelected = false
			break
		}
	}

	if allUsacloudSelected {
		// Deselect all usacloud files
		for _, usacloudFile := range usacloudFiles {
			fs.removeSelectedFile(usacloudFile)
		}
	} else {
		// Select all usacloud files
		for _, usacloudFile := range usacloudFiles {
			if !fs.isFileSelected(usacloudFile) {
				fs.selectedFiles = append(fs.selectedFiles, usacloudFile)
			}
		}
	}

	fs.populateFileList()
	fs.updateStatusBar()
}

// confirmSelection confirms the current selection
func (fs *FileSelector) confirmSelection() {
	if len(fs.selectedFiles) == 0 {
		return // No files selected, do nothing
	}

	if fs.onFilesSelected != nil {
		fs.onFilesSelected(fs.selectedFiles)
	}

	fs.app.Stop()
}

// GetSelectedFiles returns the currently selected files
func (fs *FileSelector) GetSelectedFiles() []string {
	result := make([]string, len(fs.selectedFiles))
	copy(result, fs.selectedFiles)
	return result
}

// toggleHelp toggles the visibility of the help text
func (fs *FileSelector) toggleHelp() {
	fs.helpVisible = !fs.helpVisible

	// Update help text title to reflect current state
	if fs.helpVisible {
		fs.helpText.SetTitle("❓ Help")
	} else {
		fs.helpText.SetTitle("❓ Help (Hidden)")
	}

	// Re-layout the UI
	fs.updateLayout()

	// Force redraw
	fs.app.Draw()
}
