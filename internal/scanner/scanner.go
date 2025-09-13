package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo represents information about a script file
type FileInfo struct {
	Path         string    // Full path to the file
	Name         string    // File name
	Size         int64     // File size in bytes
	ModTime      time.Time // Last modification time
	IsExecutable bool      // Whether file has execute permission
}

// BasicScanResult represents the result of a directory scan
type BasicScanResult struct {
	Directory string      // Scanned directory path
	Files     []*FileInfo // Found script files
	Errors    []string    // Any errors encountered during scan
}

// Scanner handles directory scanning for script files
type Scanner struct {
	extensions  []string // File extensions to scan for
	excludeDirs []string // Directory names to exclude
	maxDepth    int      // Maximum recursion depth (0 = current dir only)
}

// NewScanner creates a new directory scanner
func NewScanner() *Scanner {
	return &Scanner{
		extensions: []string{".sh", ".bash"},
		excludeDirs: []string{
			"node_modules", ".git", ".svn", ".hg",
			"vendor", "bin", "build", "dist",
			".vscode", ".idea", "__pycache__",
		},
		maxDepth: 2, // Scan current directory and 1 level deep
	}
}

// WithExtensions sets the file extensions to scan for
func (s *Scanner) WithExtensions(extensions []string) *Scanner {
	s.extensions = make([]string, len(extensions))
	copy(s.extensions, extensions)
	return s
}

// WithMaxDepth sets the maximum recursion depth
func (s *Scanner) WithMaxDepth(depth int) *Scanner {
	s.maxDepth = depth
	return s
}

// Scan scans the specified directory for script files
func (s *Scanner) Scan(directory string) (*BasicScanResult, error) {
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	result := &BasicScanResult{
		Directory: absDir,
		Files:     make([]*FileInfo, 0),
		Errors:    make([]string, 0),
	}

	err = s.scanRecursive(absDir, 0, result)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// Sort files by name for consistent ordering
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Name < result.Files[j].Name
	})

	return result, nil
}

// scanRecursive recursively scans directories up to maxDepth
func (s *Scanner) scanRecursive(directory string, depth int, result *BasicScanResult) error {
	if depth > s.maxDepth {
		return nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", directory, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(directory, entry.Name())

		if entry.IsDir() {
			// Skip excluded directories
			if s.isExcludedDir(entry.Name()) {
				continue
			}

			// Recurse into subdirectory
			if depth < s.maxDepth {
				if err := s.scanRecursive(fullPath, depth+1, result); err != nil {
					result.Errors = append(result.Errors, err.Error())
				}
			}
			continue
		}

		// Check if file has target extension
		if !s.hasTargetExtension(entry.Name()) {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("failed to get info for %s: %v", fullPath, err))
			continue
		}

		// Skip hidden files (starting with .)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip backup files
		if s.isBackupFile(entry.Name()) {
			continue
		}

		fileInfo := &FileInfo{
			Path:         fullPath,
			Name:         entry.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsExecutable: s.isExecutable(info),
		}

		result.Files = append(result.Files, fileInfo)
	}

	return nil
}

// hasTargetExtension checks if the file has a target extension
func (s *Scanner) hasTargetExtension(filename string) bool {
	ext := filepath.Ext(filename)
	for _, targetExt := range s.extensions {
		if strings.EqualFold(ext, targetExt) {
			return true
		}
	}
	return false
}

// isExcludedDir checks if a directory should be excluded
func (s *Scanner) isExcludedDir(dirname string) bool {
	for _, excludeDir := range s.excludeDirs {
		if dirname == excludeDir {
			return true
		}
	}
	return false
}

// isBackupFile checks if a file appears to be a backup file
func (s *Scanner) isBackupFile(filename string) bool {
	backupSuffixes := []string{
		".bak", ".backup", ".old", ".orig", ".save",
		"~", ".tmp", ".temp",
	}

	lowerName := strings.ToLower(filename)
	for _, suffix := range backupSuffixes {
		if strings.HasSuffix(lowerName, suffix) {
			return true
		}
	}

	return false
}

// isExecutable checks if a file has execute permission
func (s *Scanner) isExecutable(info os.FileInfo) bool {
	mode := info.Mode()
	return mode&0111 != 0 // Check if any execute bit is set
}

// GetRelativePath returns the relative path from the base directory
func (f *FileInfo) GetRelativePath(baseDir string) string {
	relPath, err := filepath.Rel(baseDir, f.Path)
	if err != nil {
		return f.Path // Fall back to absolute path
	}
	return relPath
}

// FormatSize returns a human-readable file size
func (f *FileInfo) FormatSize() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	size := float64(f.Size)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", size/KB)
	default:
		return fmt.Sprintf("%d B", f.Size)
	}
}

// FormatModTime returns a human-readable modification time
func (f *FileInfo) FormatModTime() string {
	now := time.Now()
	diff := now.Sub(f.ModTime)

	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	default:
		return f.ModTime.Format("2006-01-02 15:04")
	}
}

// HasUsacloudCommands checks if the file likely contains usacloud commands
func (f *FileInfo) HasUsacloudCommands() (bool, error) {
	content, err := os.ReadFile(f.Path)
	if err != nil {
		return false, err
	}

	// Simple check for usacloud command presence
	contentStr := string(content)
	return strings.Contains(contentStr, "usacloud "), nil
}

// Preview returns the first few lines of the file for preview
func (f *FileInfo) Preview(lines int) ([]string, error) {
	content, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	allLines := strings.Split(contentStr, "\n")

	if len(allLines) > lines {
		return allLines[:lines], nil
	}

	return allLines, nil
}
