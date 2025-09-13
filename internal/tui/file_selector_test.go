package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/usacloud-update/internal/config"
	"github.com/armaniacs/usacloud-update/internal/scanner"
)

func TestNewFileSelector(t *testing.T) {
	cfg := &config.SandboxConfig{
		Enabled:     true,
		Debug:       false,
		DryRun:      false,
		Interactive: true,
		Timeout:     30 * time.Second,
	}

	fs := NewFileSelector(cfg)
	if fs == nil {
		t.Error("NewFileSelector() returned nil")
	}

	if fs.config != cfg {
		t.Error("NewFileSelector() did not set config correctly")
	}

	if fs.app == nil {
		t.Error("NewFileSelector() did not initialize app")
	}

	if fs.scanner == nil {
		t.Error("NewFileSelector() did not initialize scanner")
	}

	if fs.selectedFiles == nil {
		t.Error("NewFileSelector() did not initialize selectedFiles")
	}

	if len(fs.selectedFiles) != 0 {
		t.Error("NewFileSelector() should initialize empty selectedFiles")
	}
}

func TestFileSelector_SetCallbacks(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Test OnFilesSelected callback
	var selectedFiles []string
	fs.SetOnFilesSelected(func(files []string) {
		selectedFiles = files
	})

	// Simulate callback execution
	testFiles := []string{"test1.sh", "test2.sh"}
	if fs.onFilesSelected != nil {
		fs.onFilesSelected(testFiles)
	}

	if len(selectedFiles) != 2 {
		t.Errorf("Expected 2 selected files, got %d", len(selectedFiles))
	}

	// Test OnCancel callback
	var cancelCalled bool
	fs.SetOnCancel(func() {
		cancelCalled = true
	})

	// Simulate callback execution
	if fs.onCancel != nil {
		fs.onCancel()
	}

	if !cancelCalled {
		t.Error("OnCancel callback was not called")
	}
}

func TestFileSelector_FileSelection(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Create mock file info
	file1 := &scanner.FileInfo{
		Path:         "/test/file1.sh",
		Name:         "file1.sh",
		Size:         1024,
		ModTime:      time.Now(),
		IsExecutable: true,
	}

	file2 := &scanner.FileInfo{
		Path:         "/test/file2.sh",
		Name:         "file2.sh",
		Size:         2048,
		ModTime:      time.Now(),
		IsExecutable: false,
	}

	// Create mock scan result
	fs.scanResult = &scanner.BasicScanResult{
		Directory: "/test",
		Files:     []*scanner.FileInfo{file1, file2},
		Errors:    []string{},
	}

	t.Run("IsFileSelected", func(t *testing.T) {
		// Initially no files selected
		if fs.isFileSelected(file1.Path) {
			t.Error("File1 should not be selected initially")
		}

		// Add file to selection
		fs.selectedFiles = append(fs.selectedFiles, file1.Path)

		if !fs.isFileSelected(file1.Path) {
			t.Error("File1 should be selected after adding")
		}

		if fs.isFileSelected(file2.Path) {
			t.Error("File2 should not be selected")
		}
	})

	t.Run("RemoveSelectedFile", func(t *testing.T) {
		// Setup: file1 and file2 selected
		fs.selectedFiles = []string{file1.Path, file2.Path}

		// Remove file1
		fs.removeSelectedFile(file1.Path)

		if fs.isFileSelected(file1.Path) {
			t.Error("File1 should not be selected after removal")
		}

		if !fs.isFileSelected(file2.Path) {
			t.Error("File2 should still be selected")
		}

		if len(fs.selectedFiles) != 1 {
			t.Errorf("Expected 1 selected file, got %d", len(fs.selectedFiles))
		}
	})

	t.Run("SelectAll", func(t *testing.T) {
		// Clear selection
		fs.selectedFiles = []string{}

		// Select all files
		fs.selectAll()

		if len(fs.selectedFiles) != 2 {
			t.Errorf("Expected 2 selected files, got %d", len(fs.selectedFiles))
		}

		if !fs.isFileSelected(file1.Path) {
			t.Error("File1 should be selected")
		}

		if !fs.isFileSelected(file2.Path) {
			t.Error("File2 should be selected")
		}
	})

	t.Run("SelectNone", func(t *testing.T) {
		// Setup: all files selected
		fs.selectedFiles = []string{file1.Path, file2.Path}

		// Clear selection
		fs.selectNone()

		if len(fs.selectedFiles) != 0 {
			t.Errorf("Expected 0 selected files, got %d", len(fs.selectedFiles))
		}

		if fs.isFileSelected(file1.Path) {
			t.Error("File1 should not be selected")
		}

		if fs.isFileSelected(file2.Path) {
			t.Error("File2 should not be selected")
		}
	})
}

func TestFileSelector_GetSelectedFiles(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Test with empty selection
	selected := fs.GetSelectedFiles()
	if len(selected) != 0 {
		t.Errorf("Expected 0 selected files, got %d", len(selected))
	}

	// Test with files selected
	testFiles := []string{"test1.sh", "test2.sh", "test3.sh"}
	fs.selectedFiles = testFiles

	selected = fs.GetSelectedFiles()
	if len(selected) != 3 {
		t.Errorf("Expected 3 selected files, got %d", len(selected))
	}

	// Ensure it's a copy, not the same slice
	selected[0] = "modified.sh"
	if fs.selectedFiles[0] == "modified.sh" {
		t.Error("GetSelectedFiles() should return a copy, not the original slice")
	}
}

func TestFileSelector_OnFileToggle(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Create test files
	file1 := &scanner.FileInfo{
		Path:         "/test/file1.sh",
		Name:         "file1.sh",
		Size:         1024,
		ModTime:      time.Now(),
		IsExecutable: true,
	}

	file2 := &scanner.FileInfo{
		Path:         "/test/file2.sh",
		Name:         "file2.sh",
		Size:         2048,
		ModTime:      time.Now(),
		IsExecutable: false,
	}

	fs.scanResult = &scanner.BasicScanResult{
		Directory: "/test",
		Files:     []*scanner.FileInfo{file1, file2},
		Errors:    []string{},
	}

	// Initialize UI components
	fs.setupUI()

	t.Run("ToggleSelection", func(t *testing.T) {
		// Initially no selection
		if len(fs.selectedFiles) != 0 {
			t.Error("Should start with no selected files")
		}

		// Toggle file1 selection (select)
		fs.onFileToggle(0, "", "", 0)

		if len(fs.selectedFiles) != 1 {
			t.Errorf("Expected 1 selected file, got %d", len(fs.selectedFiles))
		}

		if !fs.isFileSelected(file1.Path) {
			t.Error("File1 should be selected")
		}

		// Toggle file1 selection again (deselect)
		fs.onFileToggle(0, "", "", 0)

		if len(fs.selectedFiles) != 0 {
			t.Errorf("Expected 0 selected files, got %d", len(fs.selectedFiles))
		}

		if fs.isFileSelected(file1.Path) {
			t.Error("File1 should not be selected")
		}
	})

	t.Run("InvalidIndex", func(t *testing.T) {
		initialCount := len(fs.selectedFiles)

		// Try to toggle with invalid index
		fs.onFileToggle(999, "", "", 0)

		// Should not change selection
		if len(fs.selectedFiles) != initialCount {
			t.Error("Selection should not change for invalid index")
		}
	})
}

func TestFileSelector_WithRealFiles(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "file-selector-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"script1.sh":     "#!/bin/bash\necho 'Hello World'\nusacloud server list\n",
		"script2.bash":   "#!/bin/bash\nls -la\necho 'No usacloud here'\n",
		"script3.sh":     "#!/bin/bash\n# usacloud-update: converted\nusacloud server show 123\n",
		"not-script.txt": "This is not a script file\n",
		"empty.sh":       "",
	}

	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Make .sh files executable
		if strings.HasSuffix(filename, ".sh") {
			err = os.Chmod(filePath, 0755)
			if err != nil {
				t.Fatalf("Failed to make file executable: %v", err)
			}
		}
	}

	cfg := &config.SandboxConfig{
		Interactive: true,
	}
	fs := NewFileSelector(cfg)

	// Test scanning
	t.Run("Scanning", func(t *testing.T) {
		// Note: We can't call fs.Run() in tests because it starts the TUI
		// Instead, we'll test the scanner component directly
		result, err := fs.scanner.Scan(tempDir)
		if err != nil {
			t.Fatalf("Scanner.Scan() failed: %v", err)
		}

		fs.scanResult = result
		fs.populateFileList()

		// Should find 4 script files (.sh and .bash, including empty.sh)
		expectedFiles := 4
		if len(result.Files) != expectedFiles {
			t.Errorf("Expected %d files, got %d", expectedFiles, len(result.Files))
		}

		// Verify files are sorted and contain expected files
		foundFiles := make(map[string]bool)
		for _, file := range result.Files {
			foundFiles[file.Name] = true
		}

		expectedNames := []string{"script1.sh", "script2.bash", "script3.sh", "empty.sh"}
		for _, name := range expectedNames {
			if !foundFiles[name] {
				t.Errorf("Expected file %s not found", name)
			}
		}

		// Should not find .txt file
		if foundFiles["not-script.txt"] {
			t.Error("Should not include .txt files")
		}
	})
}

func TestFileSelector_UIComponents(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Test UI initialization
	if fs.fileList == nil {
		t.Error("fileList should be initialized")
	}

	if fs.previewPane == nil {
		t.Error("previewPane should be initialized")
	}

	if fs.statusBar == nil {
		t.Error("statusBar should be initialized")
	}

	if fs.helpText == nil {
		t.Error("helpText should be initialized")
	}

	if fs.app == nil {
		t.Error("app should be initialized")
	}

	// Test that components have proper configuration
	// Note: This is mainly a smoke test since we can't easily test TUI components
}

func TestFileSelector_ToggleUsacloudFiles(t *testing.T) {
	cfg := &config.SandboxConfig{}
	fs := NewFileSelector(cfg)

	// Create mock files - some with usacloud, some without
	// Note: In real implementation, HasUsacloudCommands() reads file content
	// For testing, we would need to mock this or create actual files
	file1 := &scanner.FileInfo{
		Path: "/test/with-usacloud.sh",
		Name: "with-usacloud.sh",
	}

	file2 := &scanner.FileInfo{
		Path: "/test/without-usacloud.sh",
		Name: "without-usacloud.sh",
	}

	fs.scanResult = &scanner.BasicScanResult{
		Directory: "/test",
		Files:     []*scanner.FileInfo{file1, file2},
		Errors:    []string{},
	}

	// Initialize UI
	fs.setupUI()

	t.Run("EmptyScanResult", func(t *testing.T) {
		fs.scanResult = nil
		initialCount := len(fs.selectedFiles)

		fs.toggleUsacloudFiles()

		if len(fs.selectedFiles) != initialCount {
			t.Error("toggleUsacloudFiles should not change selection when scanResult is nil")
		}
	})

	// Note: Full testing of toggleUsacloudFiles would require mocking
	// the HasUsacloudCommands() method or creating real files with content
}
