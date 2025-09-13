package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScriptDetector_ScanFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"usacloud_script.sh": `#!/bin/bash
usacloud server list
usacloud disk create --name test-disk
# This is a usacloud script`,
		"no_usacloud.sh": `#!/bin/bash
echo "Hello world"
ls -la`,
		"mixed_script.sh": `#!/bin/bash
usacloud server list
docker ps
usacloud disk list`,
		"binary_file.bin": "\x00\x01\x02\x03",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	config := &DetectionConfig{
		MinConfidence:   0.3,
		MaxFileSize:     1024 * 1024,
		ScanBinaryFiles: false,
	}

	detector := NewScriptDetector(config)

	tests := []struct {
		filename      string
		expectScript  bool
		minCommands   int
		minConfidence float64
	}{
		{"usacloud_script.sh", true, 2, 0.5},
		{"no_usacloud.sh", false, 0, 0.0},
		{"mixed_script.sh", true, 2, 0.5},
		{"binary_file.bin", false, 0, 0.0},
	}

	for _, test := range tests {
		filePath := filepath.Join(tempDir, test.filename)
		result, err := detector.ScanFile(filePath)

		if err != nil {
			t.Errorf("ScanFile(%s) failed: %v", test.filename, err)
			continue
		}

		if result.IsScript != test.expectScript {
			t.Errorf("ScanFile(%s): expected IsScript=%v, got %v",
				test.filename, test.expectScript, result.IsScript)
		}

		if result.CommandCount < test.minCommands {
			t.Errorf("ScanFile(%s): expected at least %d commands, got %d",
				test.filename, test.minCommands, result.CommandCount)
		}

		if result.Confidence < test.minConfidence {
			t.Errorf("ScanFile(%s): expected confidence >= %.2f, got %.2f",
				test.filename, test.minConfidence, result.Confidence)
		}
	}
}

func TestDetectionPatterns(t *testing.T) {
	detector := NewScriptDetector(nil)

	testCases := []struct {
		content      string
		expectedType string
		description  string
	}{
		{
			content:      "usacloud server list",
			expectedType: "usacloud-command",
			description:  "Direct usacloud command",
		},
		{
			content:      "USACLOUD=usacloud",
			expectedType: "usacloud-variable",
			description:  "Usacloud variable assignment",
		},
		{
			content:      "echo 'test' | usacloud server create",
			expectedType: "usacloud-in-pipe",
			description:  "Usacloud in pipe",
		},
		{
			content:      "usacloud server create --name test",
			expectedType: "infrastructure-commands",
			description:  "Infrastructure command",
		},
		{
			content:      "This script uses Sakura Cloud services",
			expectedType: "sakura-cloud-reference",
			description:  "Sakura Cloud reference",
		},
	}

	for _, tc := range testCases {
		score, commands := detector.detectByPatterns(tc.content)

		if score == 0 {
			t.Errorf("Pattern detection failed for: %s", tc.description)
			continue
		}

		if len(commands) == 0 {
			t.Errorf("No commands detected for: %s", tc.description)
			continue
		}

		found := false
		for _, cmd := range commands {
			if cmd.CommandType == tc.expectedType {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected command type '%s' not found for: %s",
				tc.expectedType, tc.description)
		}
	}
}

func TestContentAnalyzer(t *testing.T) {
	tempDir := t.TempDir()

	testFiles := map[string]string{
		"script.sh": `#!/bin/bash
# This is a comment
usacloud server list
echo "test"
`,
		"text.txt":  "This is a plain text file\nwith multiple lines",
		"empty.txt": "",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	config := &DetectionConfig{
		MaxFileSize: 1024,
	}
	analyzer := NewContentAnalyzer(config)

	tests := []struct {
		filename    string
		expectText  bool
		expectLang  string
		expectLines int
	}{
		{"script.sh", true, "bash", 5},
		{"text.txt", true, "unknown", 2},
		{"empty.txt", true, "unknown", 1},
	}

	for _, test := range tests {
		filePath := filepath.Join(tempDir, test.filename)
		analysis, err := analyzer.AnalyzeFile(filePath)

		if err != nil {
			t.Errorf("AnalyzeFile(%s) failed: %v", test.filename, err)
			continue
		}

		if analysis.IsText != test.expectText {
			t.Errorf("AnalyzeFile(%s): expected IsText=%v, got %v",
				test.filename, test.expectText, analysis.IsText)
		}

		if analysis.Language != test.expectLang {
			t.Errorf("AnalyzeFile(%s): expected Language=%s, got %s",
				test.filename, test.expectLang, analysis.Language)
		}

		if analysis.LineCount != test.expectLines {
			t.Errorf("AnalyzeFile(%s): expected LineCount=%d, got %d",
				test.filename, test.expectLines, analysis.LineCount)
		}
	}
}

func TestImportanceScorer(t *testing.T) {
	weights := ScoreWeights{
		CommandCount:     1.0,
		FileSize:         0.1,
		InfraCommands:    1.5,
		DeprecatedUsage:  2.0,
		ComplexityFactor: 0.2,
	}

	scorer := NewImportanceScorer(weights)

	result := &DetectionResult{
		CommandCount: 3,
		Commands: []DetectedCommand{
			{CommandType: "infrastructure-commands", Deprecated: false},
			{CommandType: "usacloud-command", Deprecated: true},
			{CommandType: "usacloud-command", Deprecated: false},
		},
	}

	analysis := &FileAnalysis{
		FileSize: 2048,
		TextMetrics: &TextMetrics{
			Complexity: 5,
		},
	}

	score := scorer.CalculateScore(result, analysis)

	// Expected: 3*1.0 (commands) + log(2048)*0.1 (file size) + 1*1.5 (infra) + 1*2.0 (deprecated) + 5*0.2 (complexity)
	expectedMin := 3.0 + 1.5 + 2.0 + 1.0 // Minimum expected

	if score < expectedMin {
		t.Errorf("Expected score >= %.2f, got %.2f", expectedMin, score)
	}

	priority := scorer.AssignPriority(score)
	if priority == PriorityLow {
		t.Errorf("Expected higher priority for score %.2f, got %s", score, priority.String())
	}
}

func TestDeprecatedCommandDetection(t *testing.T) {
	detector := NewScriptDetector(nil)

	deprecatedCommands := []string{
		"usacloud summary",
		"usacloud object-storage list",
		"usacloud iso-image create",
		"usacloud startup-script show",
		"usacloud ipv4 list",
		"usacloud product-server list",
	}

	for _, cmd := range deprecatedCommands {
		if !detector.isDeprecatedCommand(cmd) {
			t.Errorf("Command should be detected as deprecated: %s", cmd)
		}
	}

	modernCommands := []string{
		"usacloud server list",
		"usacloud disk create",
		"usacloud switch info",
	}

	for _, cmd := range modernCommands {
		if detector.isDeprecatedCommand(cmd) {
			t.Errorf("Command should NOT be detected as deprecated: %s", cmd)
		}
	}
}

func TestBinaryFileDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Create binary file
	binaryFile := filepath.Join(tempDir, "binary.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0x89, 0x50, 0x4E, 0x47}
	err := os.WriteFile(binaryFile, binaryContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// Create text file
	textFile := filepath.Join(tempDir, "text.txt")
	textContent := "This is a text file with usacloud commands\nusacloud server list"
	err = os.WriteFile(textFile, []byte(textContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	config := &DetectionConfig{
		ScanBinaryFiles: false,
		MinConfidence:   0.3,
	}
	detector := NewScriptDetector(config)

	// Test binary file - should not be detected as script
	result, err := detector.ScanFile(binaryFile)
	if err != nil {
		t.Fatalf("ScanFile failed for binary file: %v", err)
	}

	if result.IsScript {
		t.Errorf("Binary file should not be detected as script")
	}

	// Test text file - should be detected as script
	result, err = detector.ScanFile(textFile)
	if err != nil {
		t.Fatalf("ScanFile failed for text file: %v", err)
	}

	if !result.IsScript {
		t.Errorf("Text file with usacloud commands should be detected as script")
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
	}

	for _, test := range tests {
		actual := test.priority.String()
		if actual != test.expected {
			t.Errorf("Priority.String(): expected %s, got %s", test.expected, actual)
		}
	}
}

func TestLanguageDetection(t *testing.T) {
	analyzer := NewContentAnalyzer(&DetectionConfig{})

	tests := []struct {
		content  string
		expected string
	}{
		{"#!/bin/bash\necho hello", "bash"},
		{"#!/bin/sh\necho hello", "sh"},
		{"#!/usr/bin/python\nprint('hello')", "python"},
		{"function test() {\n  echo hello\n}", "bash"},
		{"def test():\n  print('hello')", "python"},
		{"echo hello world", "unknown"},
	}

	for _, test := range tests {
		actual := analyzer.detectScriptLanguage(test.content)
		if actual != test.expected {
			t.Errorf("detectScriptLanguage(%q): expected %s, got %s",
				test.content, test.expected, actual)
		}
	}
}

func TestTextMetricsCalculation(t *testing.T) {
	analyzer := NewContentAnalyzer(&DetectionConfig{})

	content := `#!/bin/bash
# This is a comment
echo "Hello world"

if [ -f file.txt ]; then
    # Another comment
    usacloud server list
fi
`

	metrics := analyzer.calculateTextMetrics(content)

	expectedCommentLines := 3 // # This is a comment, # Another comment, plus more
	expectedBlankLines := 2   // 2つの空行
	expectedCodeLines := 4    // #!/bin/bash, echo, if, fi (usacloud is counted separately)
	expectedComplexity := 1   // one if statement

	if metrics.CommentLines != expectedCommentLines {
		t.Errorf("Expected %d comment lines, got %d", expectedCommentLines, metrics.CommentLines)
	}

	if metrics.BlankLines != expectedBlankLines {
		t.Errorf("Expected %d blank lines, got %d", expectedBlankLines, metrics.BlankLines)
	}

	if metrics.CodeLines != expectedCodeLines {
		t.Errorf("Expected %d code lines, got %d", expectedCodeLines, metrics.CodeLines)
	}

	if metrics.Complexity != expectedComplexity {
		t.Errorf("Expected complexity %d, got %d", expectedComplexity, metrics.Complexity)
	}

	expectedRatio := float64(expectedCommentLines) / float64(expectedCommentLines+expectedBlankLines+expectedCodeLines)
	if metrics.CommentRatio != expectedRatio {
		t.Errorf("Expected comment ratio %.3f, got %.3f", expectedRatio, metrics.CommentRatio)
	}
}

func TestFileTypeDetection(t *testing.T) {
	analyzer := NewContentAnalyzer(&DetectionConfig{})

	tests := []struct {
		filename string
		content  string
		expected string
	}{
		{"script.sh", "#!/bin/bash", "shell"},
		{"script.bash", "echo hello", "shell"},
		{"script.py", "print('hello')", "python"},
		{"data.json", "{\"key\": \"value\"}", "json"},
		{"config.yaml", "key: value", "yaml"},
		{"readme.txt", "This is a readme", "text"},
		{"unknown_file", "#!/bin/bash", "script"},
		{"no_extension", "some content", "unknown"},
	}

	for _, test := range tests {
		actual := analyzer.detectFileType(test.filename, test.content)
		if actual != test.expected {
			t.Errorf("detectFileType(%s, %q): expected %s, got %s",
				test.filename, test.content, test.expected, actual)
		}
	}
}

func TestConfidenceCalculation(t *testing.T) {
	detector := NewScriptDetector(nil)

	tests := []struct {
		patternScore  float64
		analysis      *FileAnalysis
		minConfidence float64
		description   string
	}{
		{
			patternScore: 3.0,
			analysis: &FileAnalysis{
				Language: "bash",
				FilePath: "script.sh",
			},
			minConfidence: 0.8,
			description:   "High pattern score with bash file",
		},
		{
			patternScore: 1.0,
			analysis: &FileAnalysis{
				Language: "unknown",
				FilePath: "unknown_file",
			},
			minConfidence: 0.2,
			description:   "Low pattern score with unknown file",
		},
	}

	for _, test := range tests {
		confidence := detector.calculateConfidence(test.patternScore, test.analysis)

		if confidence < test.minConfidence {
			t.Errorf("%s: expected confidence >= %.2f, got %.2f",
				test.description, test.minConfidence, confidence)
		}

		if confidence > 1.0 {
			t.Errorf("%s: confidence should not exceed 1.0, got %.2f",
				test.description, confidence)
		}
	}
}
