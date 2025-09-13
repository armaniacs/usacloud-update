# Testing Framework Reference

This document provides comprehensive documentation for the usacloud-update testing frameworks, including E2E testing, integration testing, BDD testing, performance testing, and testing utilities.

## Table of Contents

1. [Testing Strategy Overview](#testing-strategy-overview)
2. [E2E Testing Framework](#e2e-testing-framework)
3. [Integration Testing Framework](#integration-testing-framework)
4. [BDD Testing Framework](#bdd-testing-framework)
5. [Performance Testing Framework](#performance-testing-framework)
6. [Golden File Testing](#golden-file-testing)
7. [Testing Utilities and Helpers](#testing-utilities-and-helpers)
8. [Test Data Management](#test-data-management)
9. [CI/CD Integration](#cicd-integration)
10. [Best Practices](#best-practices)

---

## Testing Strategy Overview

The usacloud-update project employs a **multi-layered testing strategy** with **56.1% overall test coverage**:

### Testing Pyramid

```
     /\     
    /  \    BDD & E2E Tests (User Scenarios)
   /____\   
  /      \  Integration Tests (Component Interaction)
 /________\ 
/__________\ Unit Tests (Individual Components)
```

### Testing Layers

1. **Unit Tests**: Individual component testing (5,175+ lines of test code)
2. **Integration Tests**: Cross-component interaction validation
3. **E2E Tests**: End-to-end user workflow verification
4. **BDD Tests**: Behavior-driven development scenarios
5. **Performance Tests**: Benchmark and load testing
6. **Golden File Tests**: Output comparison and regression detection

### Test Coverage by Package

| Package | Coverage | Test Files | Focus Area |
|---------|----------|------------|------------|
| `internal/transform` | 100% | 7 files | Core transformation logic |
| `internal/validation` | 89.2% | 14 files | Command validation |
| `internal/config` | 85.7% | 6 files | Configuration management |
| `internal/sandbox` | 78.5% | 6 files | Sandbox execution |
| `internal/tui` | 62.7% | 8 files | Terminal UI |
| **Overall** | **56.1%** | **41 files** | **System-wide** |

---

## E2E Testing Framework

**Location**: `tests/e2e/e2e_test_framework.go` (608 lines)

### Framework Architecture

```go
type E2ETestSuite struct {
    t          *testing.T
    testDir    string
    binaryPath string
    tempFiles  []string
    cleanup    []func()
}

type E2ETestOptions struct {
    // Command execution
    Arguments           []string
    StdinInput         string
    Environment        map[string]string
    
    // Expected results
    ExpectedExitCode   int
    ExpectedStdout     []string
    ExpectedStderr     []string
    ExpectedFiles      []FileExpectation
    
    // Performance validation
    ValidatePerformance bool
    MaxExecutionTime    string
    
    // Advanced options
    WorkingDirectory   string
    Timeout            time.Duration
}
```

### Core Capabilities

#### 1. Process Execution Testing

```go
func (suite *E2ETestSuite) RunE2ETest(name string, options *E2ETestOptions) *E2ETestResult {
    // Binary execution with environment isolation
    // Timeout control and graceful termination
    // Comprehensive output capture
    // Performance metrics collection
}
```

**Features**:
- Binary execution with timeout control
- Environment variable injection
- Working directory isolation
- Standard I/O capture and validation
- Exit code verification

#### 2. File System Validation

```go
type FileExpectation struct {
    Path            string
    ShouldExist     bool
    ContentContains []string
    ContentExcludes []string
    Permissions     os.FileMode
    MinSize         int64
    MaxSize         int64
}
```

**Capabilities**:
- Output file existence verification
- Content validation with pattern matching
- File permission checking
- Size constraints validation
- Temporary file cleanup

#### 3. Environment Management

```go
func (suite *E2ETestSuite) CreateTempFile(name, content string) string {
    // Creates isolated temporary files
}

func (suite *E2ETestSuite) CreateTempFileBytes(name string, data []byte) string {
    // Handles binary test data
}

func (suite *E2ETestSuite) CreateTempDir(name string) string {
    // Creates isolated test directories
}
```

### Usage Examples

#### Basic E2E Test

```go
func TestBasicTransformation(t *testing.T) {
    suite := e2e.NewE2ETestSuite(t)
    defer suite.Cleanup()
    
    inputScript := `#!/bin/bash
usacloud server list --output-type=csv
usacloud disk create --selector name=test`
    
    inputFile := suite.CreateTempFile("input.sh", inputScript)
    outputFile := suite.GetTestDir() + "/output.sh"
    
    options := &e2e.E2ETestOptions{
        Arguments: []string{
            "--in", inputFile,
            "--out", outputFile,
        },
        ExpectedExitCode: 0,
        ExpectedStdout: []string{"✅ 変換完了"},
        ExpectedFiles: []e2e.FileExpectation{
            {
                Path:        "output.sh",
                ShouldExist: true,
                ContentContains: []string{
                    "--output-type=json",
                    "test # usacloud-update:",
                },
            },
        },
    }
    
    result := suite.RunE2ETest("BasicTransformation", options)
    
    // Additional custom validations
    if result.ExecutionTime > time.Second*5 {
        t.Errorf("Transformation took too long: %v", result.ExecutionTime)
    }
}
```

#### Sandbox E2E Test

```go
func TestSandboxExecution(t *testing.T) {
    suite := e2e.NewE2ETestSuite(t)
    defer suite.Cleanup()
    
    options := &e2e.E2ETestOptions{
        Arguments: []string{
            "--sandbox",
            "--dry-run",
            "--batch",
            "--in", inputFile,
        },
        Environment: map[string]string{
            "SAKURACLOUD_ACCESS_TOKEN":        "test-token",
            "SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
        },
        ExpectedExitCode: 0,
        ValidatePerformance: true,
        MaxExecutionTime: "30s",
    }
    
    result := suite.RunE2ETest("SandboxExecution", options)
    
    // Validate sandbox-specific behavior
    validateSandboxSafety(t, result)
}
```

---

## Integration Testing Framework

**Location**: `tests/integration/integration_test_framework.go` (476 lines)

### Framework Structure

```go
type IntegrationTestFramework struct {
    scenarios []TestScenario
    config    *TestConfig
    reporter  *TestReporter
    golden    *GoldenFileManager
}

type TestScenario struct {
    Name        string
    Description string
    Input       TestInput
    Expected    TestExpected
    Validation  ValidationConfig
    Tags        []string
}
```

### Scenario-Based Testing

#### Test Scenario Configuration

```go
type TestInput struct {
    ScriptContent    string
    CommandLineArgs  []string
    ConfigSettings   map[string]interface{}
    EnvironmentVars  map[string]string
    InputFiles       []string
}

type TestExpected struct {
    ExitCode         int
    OutputPatterns   []string
    ErrorPatterns    []string
    FileChanges      []FileChange
    MetricThresholds map[string]float64
}
```

#### User Workflow Integration

```go
// Beginner user workflow
func TestBeginnerWorkflow(t *testing.T) {
    framework := integration.NewTestFramework()
    
    scenario := &integration.TestScenario{
        Name: "BeginnerUserTypos",
        Input: integration.TestInput{
            ScriptContent: "usacloud serv list", // Intentional typo
            CommandLineArgs: []string{"--validate-only"},
        },
        Expected: integration.TestExpected{
            ExitCode: 1,
            ErrorPatterns: []string{
                "もしかして以下のコマンドですか？",
                "server",
                "詳細情報:",
            },
        },
    }
    
    result := framework.RunScenario(scenario)
    framework.ValidateResult(t, result)
}
```

### Cross-Component Integration

#### Config + Transform Integration

```go
func TestConfigTransformIntegration(t *testing.T) {
    // Test configuration loading affecting transformation
    configContent := `[transform]
strict_mode = true
suggestion_level = 5`
    
    configFile := createTempConfig(configContent)
    defer os.Remove(configFile)
    
    result := runIntegrationTest(&IntegrationTestOptions{
        ConfigFile:    configFile,
        InputScript:   "usacloud serv list",
        ExpectedBehavior: StrictValidationMode,
    })
    
    validateStrictMode(t, result)
}
```

#### TUI + Sandbox Integration

```go
func TestTUISandboxIntegration(t *testing.T) {
    // Test TUI interface with sandbox execution
    framework := integration.NewTestFramework()
    
    scenario := framework.CreateTUIScenario(&TUIScenarioConfig{
        Commands:       []string{"usacloud server list"},
        InteractionSequence: []TUIInteraction{
            {Key: tcell.KeyEnter, ExpectedState: "CommandSelected"},
            {Key: tcell.KeyRune, Rune: 'e', ExpectedState: "Executing"},
        },
        SandboxEnabled: true,
    })
    
    result := framework.RunTUIScenario(scenario)
    validateTUIBehavior(t, result)
}
```

---

## BDD Testing Framework

**Location**: `internal/bdd/steps.go` (589 lines)

### Behavior-Driven Development

#### Step Definitions

```go
type SandboxSteps struct {
    executor       *sandbox.Executor
    config         *config.SandboxConfig
    lastResult     *sandbox.ExecutionResult
    tuiApp         *tui.App
    testScript     string
    commandResults []sandbox.ExecutionResult
}

// Context setup
func (s *SandboxSteps) iHaveAScriptWithUsacloudCommands(scriptContent string) error {
    s.testScript = scriptContent
    return nil
}

// Action execution
func (s *SandboxSteps) iRunUsacloudUpdateInSandboxMode() error {
    config, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    s.executor = sandbox.NewExecutor(config)
    return nil
}

// State validation
func (s *SandboxSteps) tuiInterfaceIsDisplayed() error {
    if s.tuiApp == nil {
        return fmt.Errorf("TUI interface was not initialized")
    }
    
    // Validate TUI state
    if !s.tuiApp.IsVisible() {
        return fmt.Errorf("TUI interface is not visible")
    }
    
    return nil
}
```

#### Feature Scenarios

**File**: `features/sandbox.feature`

```gherkin
Feature: Sandbox Execution
  As a user
  I want to execute usacloud commands safely in a sandbox environment
  So that I can test my scripts without affecting production resources

  @sandbox @interactive
  Scenario: Interactive TUI execution
    Given I have a script with usacloud commands:
      """
      usacloud server list --output-type=csv
      usacloud disk list --zone=is1a
      """
    When I run usacloud-update in sandbox mode
    Then TUI interface is displayed
    And list of converted commands is displayed
    And following options are provided for each command:
      | Option | Description |
      | Execute | Run the command |
      | Preview | Show converted command |
      | Skip | Skip this command |

  @sandbox @batch
  Scenario: Batch execution with validation
    Given I have a script with validation errors
    When I run usacloud-update in sandbox batch mode
    Then execution results are displayed in TUI
    And environment variable setup is guided
    And env sample file is referenced
```

#### BDD Implementation Patterns

```go
func (s *SandboxSteps) followingOptionsAreProvidedForEachCommand(table *godog.Table) error {
    expectedOptions := parseOptionsTable(table)
    
    for _, command := range s.commandResults {
        availableOptions := command.AvailableActions
        
        for _, expected := range expectedOptions {
            if !contains(availableOptions, expected.Option) {
                return fmt.Errorf("option %q not available for command %q", 
                    expected.Option, command.Command)
            }
        }
    }
    
    return nil
}
```

### BDD Test Execution

```bash
# Run all BDD scenarios
make bdd

# Run specific tags
godog -t @sandbox features/

# Generate BDD reports
godog -f junit features/ > bdd-results.xml
```

---

## Performance Testing Framework

**Location**: `tests/performance/performance_test_framework.go`

### Benchmark Testing

#### Performance Benchmarks

```go
func BenchmarkTransformationEngine(b *testing.B) {
    engine := transform.NewEngine()
    testLine := "usacloud server list --output-type=csv --selector name=test"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result := engine.Process(testLine)
        if !result.Changed {
            b.Errorf("Expected transformation but got unchanged result")
        }
    }
}

func BenchmarkLargeScriptProcessing(b *testing.B) {
    framework := performance.NewFramework()
    largeScript := generateLargeScript(10000) // 10K lines
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result := framework.ProcessScript(largeScript)
        framework.ValidateResult(result)
    }
}
```

#### Memory Testing

```go
func TestMemoryUsage(t *testing.T) {
    framework := performance.NewFramework()
    
    testCases := []struct {
        name      string
        lineCount int
        maxMemory int64 // bytes
    }{
        {"SmallScript", 100, 10 * 1024 * 1024},    // 10MB
        {"MediumScript", 1000, 50 * 1024 * 1024},  // 50MB
        {"LargeScript", 10000, 100 * 1024 * 1024}, // 100MB
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            script := generateScript(tc.lineCount)
            
            memBefore := getCurrentMemoryUsage()
            result := framework.ProcessScript(script)
            memAfter := getCurrentMemoryUsage()
            
            memUsed := memAfter - memBefore
            if memUsed > tc.maxMemory {
                t.Errorf("Memory usage exceeded limit: used=%d, limit=%d", 
                    memUsed, tc.maxMemory)
            }
            
            framework.ValidateResult(result)
        })
    }
}
```

#### Stress Testing

```go
func TestConcurrentExecution(t *testing.T) {
    framework := performance.NewFramework()
    numWorkers := 10
    scriptsPerWorker := 100
    
    var wg sync.WaitGroup
    errChan := make(chan error, numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            for j := 0; j < scriptsPerWorker; j++ {
                script := generateRandomScript()
                result := framework.ProcessScript(script)
                
                if err := framework.ValidateResult(result); err != nil {
                    errChan <- fmt.Errorf("worker %d, script %d: %w", 
                        workerID, j, err)
                    return
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errChan)
    
    for err := range errChan {
        t.Error(err)
    }
}
```

---

## Golden File Testing

**Location**: `internal/testing/golden_test_framework.go`

### Golden File Management

```go
type GoldenFileManager struct {
    testdataDir string
    updateMode  bool
}

func (g *GoldenFileManager) CompareOutput(t *testing.T, name string, got string) {
    goldenFile := filepath.Join(g.testdataDir, name+".golden")
    
    if g.updateMode {
        g.updateGoldenFile(goldenFile, got)
        return
    }
    
    expected, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("Failed to read golden file %s: %v", goldenFile, err)
    }
    
    if string(expected) != got {
        t.Errorf("Golden file mismatch for %s\nExpected:\n%s\nGot:\n%s\nDiff:\n%s",
            name, expected, got, getDiff(string(expected), got))
    }
}
```

### Golden File Tests

```go
func TestGoldenFileTransformation(t *testing.T) {
    golden := NewGoldenFileManager("testdata/golden")
    
    testCases := []struct {
        name  string
        input string
    }{
        {"BasicTransform", "testdata/inputs/basic.sh"},
        {"ComplexTransform", "testdata/inputs/complex.sh"},
        {"ErrorScenarios", "testdata/inputs/errors.sh"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            input, err := os.ReadFile(tc.input)
            if err != nil {
                t.Fatalf("Failed to read input: %v", err)
            }
            
            result := processInput(string(input))
            golden.CompareOutput(t, tc.name, result)
        })
    }
}
```

### Golden File Updates

```bash
# Update all golden files
make golden

# Update specific test golden files
go test -run TestGoldenFileTransformation -update ./...

# Environment variable control
UPDATE_GOLDEN=true go test ./...
```

---

## Testing Utilities and Helpers

### Test Data Generation

```go
// testdata/inputs/ management
func generateTestScript(config *ScriptGenerationConfig) string {
    var script strings.Builder
    script.WriteString("#!/bin/bash\n")
    script.WriteString("# Generated test script\n\n")
    
    for _, command := range config.Commands {
        if config.IncludeErrors && rand.Float64() < config.ErrorRate {
            script.WriteString(introduceError(command))
        } else {
            script.WriteString(command)
        }
        script.WriteString("\n")
    }
    
    return script.String()
}

func generateLargeScript(lines int) string {
    commands := []string{
        "usacloud server list --output-type csv",
        "usacloud disk list --selector name=test", 
        "usacloud iso-image list",
        "usacloud startup-script list",
    }
    
    var script strings.Builder
    script.WriteString("#!/bin/bash\n")
    
    for i := 0; i < lines; i++ {
        cmd := commands[i%len(commands)]
        script.WriteString(fmt.Sprintf("# Line %d\n%s\n", i+1, cmd))
    }
    
    return script.String()
}
```

### Assertion Helpers

```go
func assertContains(t *testing.T, text, substring string) {
    t.Helper()
    if !strings.Contains(text, substring) {
        t.Errorf("Expected %q to contain %q", text, substring)
    }
}

func assertExitCode(t *testing.T, result *E2ETestResult, expected int) {
    t.Helper()
    if result.ExitCode != expected {
        t.Errorf("Expected exit code %d, got %d\nStderr: %s", 
            expected, result.ExitCode, result.Stderr)
    }
}

func assertFileExists(t *testing.T, path string) {
    t.Helper()
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("Expected file %s to exist", path)
    }
}
```

### Mock and Stub Utilities

```go
type MockExecutor struct {
    executeFunc func(string) (*sandbox.ExecutionResult, error)
    callLog     []string
}

func (m *MockExecutor) Execute(command string) (*sandbox.ExecutionResult, error) {
    m.callLog = append(m.callLog, command)
    if m.executeFunc != nil {
        return m.executeFunc(command)
    }
    
    return &sandbox.ExecutionResult{
        Command:   command,
        ExitCode:  0,
        Stdout:    "mock output",
        Success:   true,
        Duration:  time.Millisecond * 100,
        Timestamp: time.Now(),
    }, nil
}

func (m *MockExecutor) GetCallLog() []string {
    return m.callLog
}
```

---

## Test Data Management

### Test Data Organization

```
testdata/
├── configs/              # Configuration files for testing
│   ├── beginner.conf
│   ├── default.conf
│   └── strict.conf
├── inputs/               # Input test scripts
│   ├── sample_v0_v1_mixed.sh
│   ├── special_characters.sh
│   ├── unicode_content.sh
│   ├── large_5000_lines.sh
│   └── error_scenarios.sh
└── golden/               # Expected output files
    ├── integration/
    │   ├── BasicTransform.golden
    │   ├── ComplexTransform.golden
    │   └── ErrorScenarios.golden
    └── integration_en/
        └── MultiLanguage_en.golden
```

### Test Data Factories

```go
type TestDataFactory struct {
    baseDir string
}

func (f *TestDataFactory) CreateBeginnerScript() string {
    return `#!/bin/bash
# Beginner script with common mistakes
usacloud serv list              # typo
usacloud server lst             # typo
usacloud disk create size 100   # missing --`
}

func (f *TestDataFactory) CreateComplexScript() string {
    return f.loadFromFile("complex_scenario.sh")
}

func (f *TestDataFactory) CreateMultiLanguageScript() string {
    return f.loadFromFile("unicode_content.sh")
}
```

### Configuration Test Data

```go
func createTestConfig(overrides map[string]interface{}) string {
    baseConfig := map[string]interface{}{
        "general": map[string]interface{}{
            "color_output": true,
            "language":     "ja",
        },
        "transform": map[string]interface{}{
            "strict_mode":      false,
            "suggestion_level": 3,
        },
        "validation": map[string]interface{}{
            "validate_only":     false,
            "skip_deprecated":   false,
        },
    }
    
    // Apply overrides
    for key, value := range overrides {
        baseConfig[key] = value
    }
    
    return marshalConfig(baseConfig)
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/golden-tests.yml
name: Golden File Tests
on: [push, pull_request]

jobs:
  golden-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run golden file tests
        run: make test
      
      - name: Check for golden file changes
        run: |
          if ! git diff --exit-code testdata/golden/; then
            echo "Golden files need updating"
            echo "Run 'make golden' to update"
            exit 1
          fi
```

### Test Reports Generation

```bash
# Generate test coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Generate benchmark reports
go test -bench=. -benchmem ./... > benchmark.txt

# Generate BDD reports
godog -f junit features/ > bdd-results.xml
```

### Parallel Test Execution

```bash
# Run tests in parallel
go test -parallel 4 ./...

# Run specific test patterns
go test -run TestE2E ./...
go test -run TestIntegration ./...
go test -run TestBDD ./...
```

---

## Best Practices

### Test Organization

1. **File Naming Convention**
   - Unit tests: `*_test.go`
   - Integration tests: `integration_*_test.go`
   - E2E tests: `e2e_*_test.go`
   - BDD tests: `*_steps.go` + `*.feature`

2. **Test Structure**
   ```go
   func TestFeatureName(t *testing.T) {
       // Setup
       setup := createTestSetup()
       defer setup.Cleanup()
       
       // Execute
       result := executeTestCase(setup)
       
       // Verify
       assertExpectedBehavior(t, result)
   }
   ```

3. **Test Data Isolation**
   - Each test creates its own temporary files
   - No shared state between tests
   - Cleanup after each test

### Error Testing

```go
func TestErrorHandling(t *testing.T) {
    testCases := []struct {
        name        string
        input       string
        expectedErr string
    }{
        {
            name:        "InvalidCommand",
            input:       "usacloud invalid-cmd",
            expectedErr: "invalid command",
        },
        {
            name:        "MalformedSyntax",
            input:       "usacloud server list --",
            expectedErr: "malformed syntax",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := processInput(tc.input)
            
            if !strings.Contains(result.Error, tc.expectedErr) {
                t.Errorf("Expected error %q, got %q", 
                    tc.expectedErr, result.Error)
            }
        })
    }
}
```

### Performance Testing Guidelines

1. **Benchmark Stability**
   ```go
   func BenchmarkStableFunction(b *testing.B) {
       b.ResetTimer()
       b.ReportAllocs()
       
       for i := 0; i < b.N; i++ {
           result := functionUnderTest()
           // Prevent compiler optimization
           _ = result
       }
   }
   ```

2. **Memory Leak Detection**
   ```go
   func TestNoMemoryLeaks(t *testing.T) {
       runtime.GC()
       m1 := runtime.MemStats{}
       runtime.ReadMemStats(&m1)
       
       // Execute test logic multiple times
       for i := 0; i < 1000; i++ {
           executeFunction()
       }
       
       runtime.GC()
       m2 := runtime.MemStats{}
       runtime.ReadMemStats(&m2)
       
       if m2.Alloc > m1.Alloc*2 {
           t.Errorf("Potential memory leak detected")
       }
   }
   ```

### Debugging Test Failures

1. **Verbose Output**
   ```bash
   go test -v ./...
   go test -v -run TestSpecificTest ./...
   ```

2. **Test Debugging**
   ```go
   func TestWithDebugging(t *testing.T) {
       if testing.Verbose() {
           t.Logf("Debug: input=%q", input)
           t.Logf("Debug: result=%+v", result)
       }
       
       // Test logic
   }
   ```

3. **Golden File Debugging**
   ```bash
   # Compare golden files manually
   diff testdata/golden/TestName.golden actual_output.txt
   
   # Update golden files selectively
   UPDATE_GOLDEN=true go test -run TestSpecificGolden ./...
   ```

---

This testing framework reference provides comprehensive guidance for understanding, using, and extending the usacloud-update testing infrastructure. The multi-layered approach ensures robust validation of all system components while maintaining efficiency and reliability.