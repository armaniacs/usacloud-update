package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewScanner(t *testing.T) {
	scanner := NewScanner()

	if len(scanner.extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(scanner.extensions))
	}

	expectedExts := []string{".sh", ".bash"}
	for i, ext := range expectedExts {
		if scanner.extensions[i] != ext {
			t.Errorf("Expected extension %s, got %s", ext, scanner.extensions[i])
		}
	}

	if scanner.maxDepth != 2 {
		t.Errorf("Expected maxDepth 2, got %d", scanner.maxDepth)
	}
}

func TestWithExtensions(t *testing.T) {
	scanner := NewScanner()
	customExts := []string{".py", ".js"}

	scanner = scanner.WithExtensions(customExts)

	if len(scanner.extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(scanner.extensions))
	}

	for i, ext := range customExts {
		if scanner.extensions[i] != ext {
			t.Errorf("Expected extension %s, got %s", ext, scanner.extensions[i])
		}
	}
}

func TestWithMaxDepth(t *testing.T) {
	scanner := NewScanner()
	scanner = scanner.WithMaxDepth(5)

	if scanner.maxDepth != 5 {
		t.Errorf("Expected maxDepth 5, got %d", scanner.maxDepth)
	}
}

func TestHasTargetExtension(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.sh", true},
		{"test.bash", true},
		{"test.SH", true},   // Case insensitive
		{"test.BASH", true}, // Case insensitive
		{"test.py", false},
		{"test.txt", false},
		{"test", false},
	}

	for _, test := range tests {
		result := scanner.hasTargetExtension(test.filename)
		if result != test.expected {
			t.Errorf("hasTargetExtension(%s) = %v, expected %v",
				test.filename, result, test.expected)
		}
	}
}

func TestIsExcludedDir(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		dirname  string
		expected bool
	}{
		{"node_modules", true},
		{".git", true},
		{"vendor", true},
		{"src", false},
		{"scripts", false},
		{"test", false},
	}

	for _, test := range tests {
		result := scanner.isExcludedDir(test.dirname)
		if result != test.expected {
			t.Errorf("isExcludedDir(%s) = %v, expected %v",
				test.dirname, result, test.expected)
		}
	}
}

func TestIsBackupFile(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.sh", false},
		{"test.sh.bak", true},
		{"test.sh.backup", true},
		{"test.sh~", true},
		{"test.old", true},
		{"test.tmp", true},
		{"Test.BAK", true}, // Case insensitive
	}

	for _, test := range tests {
		result := scanner.isBackupFile(test.filename)
		if result != test.expected {
			t.Errorf("isBackupFile(%s) = %v, expected %v",
				test.filename, result, test.expected)
		}
	}
}

func TestScan(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "scanner-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []struct {
		path    string
		content string
	}{
		{"script1.sh", "#!/bin/bash\nusacloud server list\n"},
		{"script2.bash", "#!/bin/bash\necho hello\n"},
		{"script3.py", "print('hello')"},  // Should be excluded
		{"script4.sh.bak", "backup file"}, // Should be excluded
		{".hidden.sh", "hidden file"},     // Should be excluded
		{"subdir/nested.sh", "nested script"},
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file.path)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}

		// Create file
		if err := os.WriteFile(fullPath, []byte(file.content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	scanner := NewScanner()
	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find script1.sh, script2.bash, and subdir/nested.sh
	expectedCount := 3
	if len(result.Files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(result.Files))
		for _, file := range result.Files {
			t.Logf("Found file: %s", file.Name)
		}
	}

	// Check that files are sorted by name
	for i := 1; i < len(result.Files); i++ {
		if result.Files[i-1].Name > result.Files[i].Name {
			t.Errorf("Files not sorted: %s > %s",
				result.Files[i-1].Name, result.Files[i].Name)
		}
	}
}

func TestFileInfoMethods(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := "#!/bin/bash\nusacloud server list\necho done\n"
	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Get file info
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	fileInfo := &FileInfo{
		Path:    tempFile.Name(),
		Name:    filepath.Base(tempFile.Name()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	// Test GetRelativePath
	tempDir := filepath.Dir(tempFile.Name())
	relPath := fileInfo.GetRelativePath(tempDir)
	if relPath != fileInfo.Name {
		t.Errorf("GetRelativePath() = %s, expected %s", relPath, fileInfo.Name)
	}

	// Test FormatSize
	sizeStr := fileInfo.FormatSize()
	if !strings.Contains(sizeStr, "B") {
		t.Errorf("FormatSize() = %s, expected to contain 'B'", sizeStr)
	}

	// Test FormatModTime
	modTimeStr := fileInfo.FormatModTime()
	if modTimeStr == "" {
		t.Errorf("FormatModTime() returned empty string")
	}

	// Test HasUsacloudCommands
	hasUsacloud, err := fileInfo.HasUsacloudCommands()
	if err != nil {
		t.Errorf("HasUsacloudCommands() failed: %v", err)
	}
	if !hasUsacloud {
		t.Errorf("HasUsacloudCommands() = false, expected true")
	}

	// Test Preview
	preview, err := fileInfo.Preview(2)
	if err != nil {
		t.Errorf("Preview() failed: %v", err)
	}
	if len(preview) != 2 {
		t.Errorf("Preview(2) returned %d lines, expected 2", len(preview))
	}
	if preview[0] != "#!/bin/bash" {
		t.Errorf("Preview first line = %s, expected '#!/bin/bash'", preview[0])
	}
}

func TestScanWithMaxDepth(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "scanner-depth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested structure: tempDir/level1/level2/level3/script.sh
	nestedPath := filepath.Join(tempDir, "level1", "level2", "level3")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create files at different levels
	files := []string{
		filepath.Join(tempDir, "root.sh"),
		filepath.Join(tempDir, "level1", "level1.sh"),
		filepath.Join(tempDir, "level1", "level2", "level2.sh"),
		filepath.Join(tempDir, "level1", "level2", "level3", "level3.sh"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("#!/bin/bash\n"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Test with maxDepth = 0 (only root directory)
	scanner := NewScanner().WithMaxDepth(0)
	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan with depth 0 failed: %v", err)
	}
	if len(result.Files) != 1 {
		t.Errorf("With depth 0, expected 1 file, got %d", len(result.Files))
	}

	// Test with maxDepth = 2 (root + 2 levels)
	scanner = NewScanner().WithMaxDepth(2)
	result, err = scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan with depth 2 failed: %v", err)
	}
	if len(result.Files) != 3 {
		t.Errorf("With depth 2, expected 3 files, got %d", len(result.Files))
	}
}
