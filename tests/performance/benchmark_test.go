package performance

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/armaniacs/usacloud-update/internal/transform"
)

func BenchmarkMainCommandValidation(b *testing.B) {
	commands := []string{
		"server", "disk", "database", "invalid-command",
		"serv", "dsk", "databse",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		command := commands[i%len(commands)]
		_ = len(command) > 0
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	testPairs := []struct {
		s1, s2 string
	}{
		{"server", "serv"},
		{"database", "databse"},
		{"webaccelerator", "webacelerator"},
	}

	for _, pair := range testPairs {
		b.Run(fmt.Sprintf("%s_%s", pair.s1, pair.s2), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = levenshteinDistance(pair.s1, pair.s2)
			}
		})
	}
}

func levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min3(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len1][len2]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func BenchmarkTransformEngine(b *testing.B) {
	engine := transform.NewDefaultEngine()

	testLines := []string{
		"usacloud server list --output-type csv",
		"usacloud iso-image list",
		"usacloud disk create --size 100",
		"echo 'non-usacloud command'",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		line := testLines[i%len(testLines)]
		engine.Apply(line)
	}
}

func BenchmarkFileProcessing(b *testing.B) {
	fileSizes := []struct {
		name  string
		lines int
	}{
		{"Small", 100},
		{"Medium", 1000},
		{"Large", 10000},
	}

	for _, size := range fileSizes {
		b.Run(size.name, func(b *testing.B) {
			suite := NewPerformanceTestSuite(&testing.T{})
			inputFile := suite.GenerateTestFile(size.lines)
			defer os.Remove(inputFile)

			engine := transform.NewDefaultEngine()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				processFile(engine, inputFile)
			}
		})
	}
}

func processFile(engine *transform.Engine, inputFile string) *ProcessingResult {
	file, err := os.Open(inputFile)
	if err != nil {
		return &ProcessingResult{}
	}
	defer file.Close()

	result := &ProcessingResult{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		result.ProcessedLines++

		engineResult := engine.Apply(line)
		if engineResult.Changed {
			result.TransformedLines++
		}
	}

	return result
}

func TestPerformance_SmallFile(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	inputFile := suite.GenerateTestFile(100)
	defer os.Remove(inputFile)

	metrics := suite.MeasurePerformance("Small file", inputFile, func() *ProcessingResult {
		engine := transform.NewDefaultEngine()
		return processFile(engine, inputFile)
	})

	t.Logf("Small file performance:")
	t.Logf("  Processing rate: %.2f lines/sec", metrics.ProcessingRate)
	t.Logf("  Memory efficiency: %.2f lines/MB", metrics.MemoryEfficiency)
	t.Logf("  Total time: %v", metrics.TotalTime)
	t.Logf("  Peak memory: %.2f MB", metrics.PeakMemoryMB)
}

func TestPerformance_MediumFile(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	inputFile := suite.GenerateTestFile(1000)
	defer os.Remove(inputFile)

	metrics := suite.MeasurePerformance("Medium file", inputFile, func() *ProcessingResult {
		engine := transform.NewDefaultEngine()
		return processFile(engine, inputFile)
	})

	t.Logf("Medium file performance:")
	t.Logf("  Processing rate: %.2f lines/sec", metrics.ProcessingRate)
	t.Logf("  Memory efficiency: %.2f lines/MB", metrics.MemoryEfficiency)
	t.Logf("  Total time: %v", metrics.TotalTime)
	t.Logf("  Peak memory: %.2f MB", metrics.PeakMemoryMB)
}

func TestPerformance_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	inputFile := suite.GenerateTestFile(10000)
	defer os.Remove(inputFile)

	metrics := suite.MeasurePerformance("Large file", inputFile, func() *ProcessingResult {
		engine := transform.NewDefaultEngine()
		return processFile(engine, inputFile)
	})

	t.Logf("Large file performance:")
	t.Logf("  Processing rate: %.2f lines/sec", metrics.ProcessingRate)
	t.Logf("  Memory efficiency: %.2f lines/MB", metrics.MemoryEfficiency)
	t.Logf("  Total time: %v", metrics.TotalTime)
	t.Logf("  Peak memory: %.2f MB", metrics.PeakMemoryMB)
}

func TestPerformance_ComplexScript(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストは短時間モードではスキップ")
	}

	suite := NewPerformanceTestSuite(t)

	complexScript := filepath.Join(suite.tempDir, "complex.sh")
	content := `#!/bin/bash
# Complex mixed script with various usacloud commands
usacloud server list --output-type csv
usacloud disk list --output-type tsv
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 list
usacloud product-disk list
usacloud summary
usacloud object-storage list
usacloud server list --zone = all
echo "Non-usacloud command"
python3 -c "print('hello')"
curl -X GET https://api.example.com
usacloud serv list  # typo
usacloud disk invalid-action  # invalid subcommand
# usacloud commented command
usacloud database create --name test --type postgresql
`

	if err := os.WriteFile(complexScript, []byte(content), 0644); err != nil {
		t.Fatalf("複雑スクリプト作成エラー: %v", err)
	}
	defer os.Remove(complexScript)

	metrics := suite.MeasurePerformance("Complex script", complexScript, func() *ProcessingResult {
		engine := transform.NewDefaultEngine()
		return processFile(engine, complexScript)
	})

	t.Logf("Complex script performance:")
	t.Logf("  Processing rate: %.2f lines/sec", metrics.ProcessingRate)
	t.Logf("  Memory efficiency: %.2f lines/MB", metrics.MemoryEfficiency)
	t.Logf("  Total time: %v", metrics.TotalTime)
	t.Logf("  Peak memory: %.2f MB", metrics.PeakMemoryMB)
	t.Logf("  Lines processed: %d", metrics.ProcessedLines)
	t.Logf("  Lines transformed: %d", metrics.TransformedLines)
}
