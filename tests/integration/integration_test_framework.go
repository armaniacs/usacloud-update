package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type IntegrationTestSuite struct {
	t           *testing.T
	tempDir     string
	binaryPath  string
	configPath  string
	testDataDir string

	timeout     time.Duration
	parallelism int
	verbose     bool
}

func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		t:           t,
		tempDir:     t.TempDir(),
		testDataDir: "testdata",
		timeout:     30 * time.Second,
		parallelism: 4,
		verbose:     testing.Verbose(),
	}

	suite.setupTestEnvironment()
	return suite
}

func (its *IntegrationTestSuite) setupTestEnvironment() {
	its.t.Helper()

	its.buildTestBinary()
	its.createTestConfig()
}

func (its *IntegrationTestSuite) buildTestBinary() {
	its.t.Helper()

	its.binaryPath = filepath.Join(its.tempDir, "usacloud-update-test")
	if runtime.GOOS == "windows" {
		its.binaryPath += ".exe"
	}

	// プロジェクトルートを堅牢に検索
	rootDir := its.findProjectRoot()
	if rootDir == "" {
		its.t.Fatalf("プロジェクトルート（go.mod）が見つかりません")
	}

	cmdArgs := []string{"build", "-o", its.binaryPath, "./cmd/usacloud-update"}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = rootDir
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		its.t.Fatalf("テストバイナリビルドエラー in %s: %v\n出力: %s", rootDir, err, output)
	}

	if _, err := os.Stat(its.binaryPath); os.IsNotExist(err) {
		its.t.Fatalf("ビルドされたバイナリが見つかりません: %s", its.binaryPath)
	}
}

// findProjectRoot はgo.modファイルを探してプロジェクトルートを特定
func (its *IntegrationTestSuite) findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// go.modファイルを探してプロジェクトルートを特定
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// ルートディレクトリに到達した場合は空文字を返す
			return ""
		}
		dir = parent
	}
}

func (its *IntegrationTestSuite) createTestConfig() {
	its.t.Helper()

	its.configPath = filepath.Join(its.tempDir, "test-config.conf")
	configContent := `[general]
version = "1.9.0"
color_output = true
language = "ja"
verbose = false
profile = "default"

[validation]
enable_validation = true
strict_mode = false
max_suggestions = 5
typo_detection_enabled = true

[output]
format = "auto"
show_progress = true
report_level = "summary"
`

	if err := os.WriteFile(its.configPath, []byte(configContent), 0644); err != nil {
		its.t.Fatalf("テスト設定ファイル作成エラー: %v", err)
	}
}

type TestScenario struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Input       ScenarioInput          `yaml:"input"`
	Expected    ScenarioExpected       `yaml:"expected"`
	Config      map[string]interface{} `yaml:"config"`
	Environment map[string]string      `yaml:"environment"`
	Timeout     string                 `yaml:"timeout"`
	Tags        []string               `yaml:"tags"`
}

type ScenarioInput struct {
	Type        string            `yaml:"type"`
	Content     string            `yaml:"content"`
	FilePath    string            `yaml:"file_path"`
	Arguments   []string          `yaml:"arguments"`
	Environment map[string]string `yaml:"environment"`
}

type ScenarioExpected struct {
	ExitCode          int             `yaml:"exit_code"`
	OutputContains    []string        `yaml:"output_contains"`
	OutputNotContains []string        `yaml:"output_not_contains"`
	ErrorContains     []string        `yaml:"error_contains"`
	FilesCreated      []string        `yaml:"files_created"`
	FilesModified     []string        `yaml:"files_modified"`
	Metrics           ExpectedMetrics `yaml:"metrics"`
}

type ExpectedMetrics struct {
	ProcessedLines   int `yaml:"processed_lines"`
	TransformedLines int `yaml:"transformed_lines"`
	ErrorsFound      int `yaml:"errors_found"`
	SuggestionsShown int `yaml:"suggestions_shown"`
	ExecutionTimeMs  int `yaml:"execution_time_ms"`
}

type ScenarioFile struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Scenarios   []TestScenario `yaml:"scenarios"`
}

func (its *IntegrationTestSuite) LoadScenarioFile(filename string) (*ScenarioFile, error) {
	path := filepath.Join(its.testDataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("シナリオファイル読み込みエラー %s: %w", path, err)
	}

	var scenarioFile ScenarioFile
	if err := yaml.Unmarshal(data, &scenarioFile); err != nil {
		return nil, fmt.Errorf("YAMLパースエラー %s: %w", path, err)
	}

	return &scenarioFile, nil
}

func (its *IntegrationTestSuite) RunScenarioFromFile(filename string) {
	its.t.Helper()

	scenarioFile, err := its.LoadScenarioFile(filename)
	if err != nil {
		its.t.Fatalf("シナリオファイルロードエラー: %v", err)
	}

	for _, scenario := range scenarioFile.Scenarios {
		its.t.Run(scenario.Name, func(t *testing.T) {
			suite := &IntegrationTestSuite{
				t:           t,
				tempDir:     its.tempDir,
				binaryPath:  its.binaryPath,
				configPath:  its.configPath,
				testDataDir: its.testDataDir,
				timeout:     its.timeout,
				parallelism: its.parallelism,
				verbose:     its.verbose,
			}
			suite.RunScenario(scenario)
		})
	}
}

type ExecutionResult struct {
	ExitCode      int
	Stdout        string
	Stderr        string
	ExecutionTime time.Duration
	Error         error
}

func (its *IntegrationTestSuite) RunScenario(scenario TestScenario) {
	its.t.Helper()

	timeout := its.timeout
	if scenario.Timeout != "" {
		if parsed, err := time.ParseDuration(scenario.Timeout); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := its.executeScenario(ctx, scenario)

	its.validateScenarioResult(scenario, result)
}

func (its *IntegrationTestSuite) executeScenario(ctx context.Context, scenario TestScenario) *ExecutionResult {
	start := time.Now()

	var cmd *exec.Cmd
	switch scenario.Input.Type {
	case "command":
		cmd = its.buildCommand(scenario)
	case "file":
		cmd = its.buildFileCommand(scenario)
	case "stdin":
		cmd = its.buildStdinCommand(scenario)
	default:
		return &ExecutionResult{
			Error: fmt.Errorf("未対応の入力タイプ: %s", scenario.Input.Type),
		}
	}

	// Context handling for older Go versions
	if ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()
	}

	for key, value := range scenario.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	for key, value := range scenario.Input.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	stdout, stderr, err := its.runCommand(cmd)
	executionTime := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &ExecutionResult{
		ExitCode:      exitCode,
		Stdout:        stdout,
		Stderr:        stderr,
		ExecutionTime: executionTime,
		Error:         err,
	}
}

func (its *IntegrationTestSuite) buildCommand(scenario TestScenario) *exec.Cmd {
	args := scenario.Input.Arguments

	if scenario.Input.Content != "" {
		inputFile := filepath.Join(its.tempDir, "input.txt")
		if err := os.WriteFile(inputFile, []byte(scenario.Input.Content), 0644); err != nil {
			its.t.Fatalf("入力ファイル作成エラー: %v", err)
		}
		args = append(args, "--in", inputFile)
	}

	return exec.Command(its.binaryPath, args...)
}

func (its *IntegrationTestSuite) buildFileCommand(scenario TestScenario) *exec.Cmd {
	args := scenario.Input.Arguments

	if scenario.Input.FilePath != "" {
		// FilePath が絶対パスの場合はそのまま使用、相対パスの場合はtestDataDirと結合
		var fullPath string
		if filepath.IsAbs(scenario.Input.FilePath) {
			fullPath = scenario.Input.FilePath
		} else {
			fullPath = filepath.Join(its.testDataDir, scenario.Input.FilePath)
		}
		args = append(args, "--in", fullPath)
	}

	return exec.Command(its.binaryPath, args...)
}

func (its *IntegrationTestSuite) buildStdinCommand(scenario TestScenario) *exec.Cmd {
	args := scenario.Input.Arguments

	cmd := exec.Command(its.binaryPath, args...)
	if scenario.Input.Content != "" {
		cmd.Stdin = strings.NewReader(scenario.Input.Content)
	}

	return cmd
}

func (its *IntegrationTestSuite) runCommand(cmd *exec.Cmd) (string, string, error) {
	if its.verbose {
		its.t.Logf("実行コマンド: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

func (its *IntegrationTestSuite) validateScenarioResult(scenario TestScenario, result *ExecutionResult) {
	its.t.Helper()

	if result.Error != nil && scenario.Expected.ExitCode == 0 {
		its.t.Errorf("実行エラー: %v\n標準出力: %s\n標準エラー: %s", result.Error, result.Stdout, result.Stderr)
		return
	}

	if result.ExitCode != scenario.Expected.ExitCode {
		its.t.Errorf("終了コード不一致: 期待=%d, 実際=%d", scenario.Expected.ExitCode, result.ExitCode)
	}

	for _, expected := range scenario.Expected.OutputContains {
		if !strings.Contains(result.Stdout, expected) {
			its.t.Errorf("標準出力に期待する文字列が含まれていません: '%s'\n実際の出力: %s", expected, result.Stdout)
		}
	}

	for _, notExpected := range scenario.Expected.OutputNotContains {
		if strings.Contains(result.Stdout, notExpected) {
			its.t.Errorf("標準出力に含まれてはいけない文字列が含まれています: '%s'", notExpected)
		}
	}

	for _, expected := range scenario.Expected.ErrorContains {
		if !strings.Contains(result.Stderr, expected) {
			its.t.Errorf("標準エラーに期待する文字列が含まれていません: '%s'\n実際のエラー: %s", expected, result.Stderr)
		}
	}

	if scenario.Expected.Metrics.ExecutionTimeMs > 0 {
		maxDuration := time.Duration(scenario.Expected.Metrics.ExecutionTimeMs) * time.Millisecond
		if result.ExecutionTime > maxDuration {
			its.t.Errorf("実行時間が制限を超過: %v > %v", result.ExecutionTime, maxDuration)
		}
	}

	for _, filename := range scenario.Expected.FilesCreated {
		fullPath := filepath.Join(its.tempDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			its.t.Errorf("期待されるファイルが作成されていません: %s", filename)
		}
	}
}

func (its *IntegrationTestSuite) Cleanup() {
}

type PerformanceResult struct {
	ExecutionTime time.Duration
	MaxMemoryMB   int
	CPUUsage      float64
}

func (its *IntegrationTestSuite) MeasurePerformance(fn func()) *PerformanceResult {
	start := time.Now()
	fn()
	executionTime := time.Since(start)

	return &PerformanceResult{
		ExecutionTime: executionTime,
		MaxMemoryMB:   0,
		CPUUsage:      0.0,
	}
}

func (its *IntegrationTestSuite) GenerateTestFile(lines int) string {
	its.t.Helper()

	testFile := filepath.Join(its.tempDir, fmt.Sprintf("test-file-%d.sh", lines))

	content := strings.Builder{}
	content.WriteString("#!/bin/bash\n")
	content.WriteString("# Generated test file\n\n")

	commands := []string{
		"usacloud server list",
		"usacloud disk list --output-type csv",
		"usacloud server read 123456789",
		"usacloud database list",
		"usacloud invalid-command",
		"usacloud server invalid-subcommand",
		"usacloud iso-image list",
		"echo 'non-usacloud command'",
	}

	for i := 0; i < lines; i++ {
		command := commands[i%len(commands)]
		content.WriteString(fmt.Sprintf("%s  # Line %d\n", command, i+1))
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		its.t.Fatalf("テストファイル生成エラー: %v", err)
	}

	return testFile
}

func (its *IntegrationTestSuite) ExecuteCommand(args ...string) *ExecutionResult {
	fullArgs := args

	cmd := exec.Command(its.binaryPath, fullArgs...)
	stdout, stderr, err := its.runCommand(cmd)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return &ExecutionResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Error:    err,
	}
}

func (its *IntegrationTestSuite) ExecuteCommandWithFile(inputFile string) *ExecutionResult {
	return its.ExecuteCommand("--in", inputFile, "--out", "-")
}
