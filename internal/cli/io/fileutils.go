package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// BufferSize defines the buffer size for reading files (1MB)
	BufferSize = 1024 * 1024
	// BinaryDetectionSize defines how many bytes to check for binary content
	BinaryDetectionSize = 512
)

// FileReader provides unified file reading capabilities
type FileReader struct {
	enableBinaryDetection bool
}

// NewFileReader creates a new file reader with binary detection enabled by default
func NewFileReader() *FileReader {
	return &FileReader{
		enableBinaryDetection: true,
	}
}

// SetBinaryDetection enables or disables binary file detection
func (fr *FileReader) SetBinaryDetection(enabled bool) {
	fr.enableBinaryDetection = enabled
}

// ReadInputFile reads from the specified path or stdin if path is "-"
// Returns an io.Reader for the content and any error encountered
func (fr *FileReader) ReadInputFile(path string) (io.Reader, error) {
	if path == "-" {
		return os.Stdin, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Check for binary content if enabled
	if fr.enableBinaryDetection {
		if err := fr.DetectBinaryContent(f); err != nil {
			f.Close()
			return nil, err
		}

		// Reset file position to beginning after binary check
		if _, err := f.Seek(0, 0); err != nil {
			f.Close()
			return nil, fmt.Errorf("ファイル位置のリセットに失敗: %w", err)
		}
	}

	return f, nil
}

// ReadInputLines reads lines from the specified path or stdin if path is "-"
// Returns a slice of lines and any error encountered
func (fr *FileReader) ReadInputLines(path string) ([]string, error) {
	reader, err := fr.ReadInputFile(path)
	if err != nil {
		return nil, err
	}

	// Close file if it's not stdin
	if f, ok := reader.(*os.File); ok && f != os.Stdin {
		defer f.Close()
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, BufferSize), BufferSize)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Note: Empty files are allowed - the original readFileLines function allowed them

	return lines, nil
}

// DetectBinaryContent checks if the reader contains binary content by looking for null bytes
func (fr *FileReader) DetectBinaryContent(reader io.Reader) error {
	// Create a buffer to read the first bytes
	firstBytes := make([]byte, BinaryDetectionSize)
	n, err := reader.Read(firstBytes)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ファイル読み込み中にエラーが発生: %w", err)
	}

	if n > 0 {
		// Check if content contains null bytes (binary indicator)
		for i := 0; i < n; i++ {
			if firstBytes[i] == 0 {
				return &BinaryFileError{Message: "バイナリファイルは処理できません"}
			}
		}
	}

	return nil
}

// WriteOutputFile writes content to the specified path or stdout if path is "-"
func WriteOutputFile(path string, content string) error {
	var w io.Writer = os.Stdout
	if path != "-" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	_, err := io.WriteString(w, content)
	return err
}

// BinaryFileError represents an error when a binary file is detected
type BinaryFileError struct {
	Message string
}

func (e *BinaryFileError) Error() string {
	return e.Message
}

// IsBinaryFileError checks if the error is a binary file error
func IsBinaryFileError(err error) bool {
	_, ok := err.(*BinaryFileError)
	return ok
}

// ReadFileLines is a standalone function for reading file lines
// Useful for compatibility with existing code
func ReadFileLines(path string) ([]string, error) {
	reader := NewFileReader()
	return reader.ReadInputLines(path)
}

// DetectBinaryFile is a standalone function for binary detection
// Useful for quick binary file checks
func DetectBinaryFile(path string) error {
	if path == "-" {
		return nil // Don't check stdin for binary content
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := NewFileReader()
	return reader.DetectBinaryContent(f)
}

// ValidateFilePath checks if a file path is valid for reading
func ValidateFilePath(path string) error {
	if path == "-" {
		return nil // stdin is always valid
	}

	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("ファイルパスが空です")
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ファイルが見つかりません: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("ファイルへのアクセス権限がありません: %s", path)
		}
		return fmt.Errorf("ファイルアクセスエラー: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return fmt.Errorf("指定されたパスはディレクトリです: %s", path)
	}

	return nil
}

// ValidateOutputPath checks if an output path is valid for writing
func ValidateOutputPath(path string) error {
	if path == "-" {
		return nil // stdout is always valid
	}

	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("出力パスが空です")
	}

	// Check if the directory exists
	dir := strings.TrimSuffix(path, "/"+strings.Split(path, "/")[len(strings.Split(path, "/"))-1])
	if dir != path { // Only check if there's a directory part
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("出力ディレクトリが見つかりません: %s", dir)
			}
		}
	}

	return nil
}
