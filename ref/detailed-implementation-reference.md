# Detailed Implementation Reference

This document provides comprehensive implementation details for the usacloud-update project, covering all key components, architectures, and extension points.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Main Application Entry Point](#main-application-entry-point)
3. [Transform Engine](#transform-engine)
4. [Validation System](#validation-system)
5. [Configuration Management](#configuration-management)
6. [Sandbox Execution](#sandbox-execution)
7. [Terminal User Interface](#terminal-user-interface)
8. [Testing Frameworks](#testing-frameworks)
9. [Extension Points](#extension-points)
10. [Implementation Patterns](#implementation-patterns)

---

## Architecture Overview

The usacloud-update project follows a **modular, multi-modal architecture** with the following key characteristics:

### Core Design Principles
- **Single Responsibility**: Each component has a well-defined purpose
- **Dependency Injection**: Components receive dependencies through interfaces
- **Configuration-Driven**: Behavior controlled through structured configuration
- **Safety-First**: Multiple validation layers for secure operation
- **User-Centric**: Rich feedback and interactive capabilities

### Component Interaction Flow
```
CLI Entry Point (main.go)
    ↓
Config Management (internal/config)
    ↓
Mode Selection (Transform | Validation | Sandbox)
    ↓
Transform Engine (internal/transform) ← → Validation Engine (internal/validation)
    ↓
Output Processing (TUI | File | STDOUT)
```

---

## Main Application Entry Point

**File**: `cmd/usacloud-update/main.go` (1,313 lines)

### Core Architecture

#### Key Types and Structures

```go
// ProcessResult integrates transformation and validation results
type ProcessResult struct {
    LineNumber       int
    OriginalLine     string
    TransformResult  *transform.Result
    ValidationResult *ValidationResult
}

// Config represents unified application configuration
type Config struct {
    // Basic I/O settings
    InputPath  string
    OutputPath string
    ShowStats  bool
    
    // Validation settings
    ValidateOnly     bool
    StrictValidation bool
    InteractiveMode  bool
    
    // User Experience
    HelpMode         string
    SuggestionLevel  int
    ColorEnabled     bool
    LanguageCode     string
    
    // Sandbox settings
    SandboxMode      bool
    BatchMode        bool
    DryRunMode       bool
    InteractiveTUI   bool
    ConfigFile       string
}
```

#### Operational Modes

1. **Transform Mode** (Default)
   - Line-by-line script transformation
   - Statistical output generation
   - Error recovery and reporting

2. **Validation-Only Mode** (`--validate-only`)
   - Comprehensive syntax checking
   - Command structure validation
   - Error reporting without transformation

3. **Sandbox Mode** (`--sandbox`)
   - Real command execution in tk1v zone
   - Interactive TUI interface
   - Batch and dry-run capabilities

4. **Interactive Mode** (`--interactive-mode`)
   - Step-by-step validation and fixing
   - User-guided error resolution

#### Key Functions

- **`parseFlags()`**: Command-line argument processing with flag validation
- **`processLine()`**: Core line processing logic integrating transform and validation
- **`runTransformMode()`**: Traditional transformation pipeline
- **`runValidationMode()`**: Validation-only processing
- **`runSandboxMode()`**: Sandbox execution with TUI
- **`runInteractiveMode()`**: Interactive validation and correction

---

## Transform Engine

**Files**: 
- `internal/transform/engine.go` - Core transformation logic
- `internal/transform/rules.go` - Rule interface and implementations  
- `internal/transform/ruledefs.go` - Rule definitions and catalog

### Core Architecture

#### Engine Structure
```go
type Engine struct {
    rules []Rule
}

type Result struct {
    Line         string  // Transformed line
    Changed      bool    // Whether transformation occurred
    BeforePhrase string  // Original phrase that was changed
    AfterPhrase  string  // New phrase after transformation
    RuleName     string  // Name of rule that was applied
    Comment      string  // Explanatory comment
    URL          string  // Documentation URL
}

type Rule interface {
    Name() string
    Apply(line string) (newLine string, changed bool, beforeFrag string, afterFrag string)
}
```

#### Rule Categories (9 types)

1. **Output Format Migration**
   - Converts `--output-type=csv/tsv` → `--output-type=json`
   - Adds guidance for jq/query usage

2. **Selector Elimination**
   - Transforms `--selector name=value` → `value` (as argument)
   - Handles ID, name, and tag selectors

3. **Resource Renaming**
   - `iso-image` → `cdrom`
   - `startup-script` → `note`
   - `ipv4` → `ipaddress`

4. **Product Alias Consolidation**
   - `product-disk` → `disk-plan`
   - `product-internet` → `internet-plan`
   - `product-server` → `server-plan`

5. **Command Deprecation**
   - Comments out `usacloud summary`
   - Provides alternative suggestions

6. **Service Discontinuation**
   - Handles `object-storage` removal
   - Suggests migration paths

7. **Parameter Normalization**
   - `--zone = all` → `--zone=all`
   - Standardizes spacing

8. **Context-Aware Processing**
   - Only processes usacloud commands
   - Preserves non-usacloud lines

9. **Documentation Integration**
   - Automatic comment generation
   - URL references for further reading

#### Implementation Pattern

```go
func (e *Engine) Process(line string) *Result {
    // Skip empty lines and comments
    if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
        return &Result{Line: line, Changed: false}
    }
    
    // Apply rules sequentially
    for _, rule := range e.rules {
        if newLine, changed, before, after := rule.Apply(line); changed {
            return &Result{
                Line:         newLine,
                Changed:      true,
                BeforePhrase: before,
                AfterPhrase:  after,
                RuleName:     rule.Name(),
                Comment:      generateComment(rule),
                URL:          getDocumentationURL(rule),
            }
        }
    }
    
    return &Result{Line: line, Changed: false}
}
```

---

## Validation System

**Files**:
- `internal/validation/comprehensive_error_formatter.go` (636 lines) - Advanced error formatting
- `internal/validation/parser.go` (393 lines) - Command parsing
- `internal/validation/main_command_validator.go` - Command validation

### Comprehensive Error Formatting

#### Architecture Features
- **Multi-language Support**: Japanese/English error messages
- **User Intent Inference**: Context-aware error interpretation
- **Visual Enhancement**: Colors, icons, and structured output
- **Detailed Guidance**: Specific fix suggestions with examples

#### Key Components

```go
type ComprehensiveErrorFormatter struct {
    templates map[string]MessageTemplate
    language  string
}

type MessageTemplate struct {
    SectionHeader    string
    ErrorHeader      string
    WarningHeader    string
    SuggestionsHeader string
    SeeAlso          string
    ValidationStatus  string
    SummaryTemplate   string
}
```

#### Error Classification System

1. **Critical Errors** (🔴)
   - Invalid main commands
   - Syntax errors
   - Missing required parameters

2. **Warnings** (🟡)
   - Deprecated commands
   - Performance concerns
   - Style recommendations

3. **Info** (ℹ️)
   - Migration suggestions
   - Best practices
   - Documentation references

#### User Intent Inference Engine

```go
func (f *ComprehensiveErrorFormatter) inferUserIntent(command string, suggestions []SimilarityResult) UserIntent {
    // Analyze command patterns
    // Consider similarity scores
    // Examine context clues
    // Return inferred intent with confidence
}
```

### Command Parser

#### Parsing Strategy
- **Flexible Tokenization**: Handles various shell syntax
- **Context Preservation**: Maintains original formatting
- **Error Recovery**: Graceful handling of malformed input

```go
type ParsedCommand struct {
    Original    string
    Command     string
    SubCommand  string
    Arguments   []string
    Flags       map[string]string
    IsUsacloud  bool
    LineNumber  int
}
```

---

## Configuration Management

**Files**:
- `internal/config/env.go` (259 lines) - Environment configuration
- `internal/config/file.go` (292 lines) - File-based configuration

### Configuration Architecture

#### Multi-Source Priority System
1. **Command Line Flags** (Highest Priority)
2. **Custom Config File** (`--config` parameter)
3. **Default Config File** (`~/.config/usacloud-update/usacloud-update.conf`)
4. **Environment Variables** (Legacy support)
5. **Built-in Defaults** (Lowest Priority)

#### Configuration Structure
```go
type SandboxConfig struct {
    // Sakura Cloud API Credentials
    AccessToken       string
    AccessTokenSecret string
    
    // Execution Settings
    DefaultZone       string
    TimeoutDuration   int
    EnableSafeMode    bool
    
    // User Interface
    ColorOutput       bool
    Language          string
    LogLevel          string
    
    // Advanced Features
    ProfileName       string
    CustomEndpoint    string
}
```

#### Cross-Platform Support

**Directory Resolution Logic**:
- **Linux/macOS**: `~/.config/usacloud-update/`
- **Windows**: `%APPDATA%\usacloud-update\`
- **Custom**: `$USACLOUD_UPDATE_CONFIG_DIR`

#### Migration and Validation

```go
func LoadConfig(customPath ...string) (*SandboxConfig, error) {
    // 1. Determine config source priority
    // 2. Load from appropriate source
    // 3. Validate configuration values
    // 4. Offer migration assistance if needed
    // 5. Return validated configuration
}
```

---

## Sandbox Execution

**File**: `internal/sandbox/executor.go` (338 lines)

### Security and Safety Architecture

#### Multi-Layer Safety System

1. **Zone Enforcement**
   - All commands executed with `--zone=tk1v`
   - Automatic zone parameter injection
   - Zone validation before execution

2. **Command Filtering**
   - Dangerous operations blocked (`delete`, `shutdown`, `reset`)
   - Read-only operations preferred
   - Command whitelist validation

3. **Execution Limits**
   - 30-second timeout per command
   - Resource usage monitoring
   - Graceful interruption handling

#### Execution Result Tracking

```go
type ExecutionResult struct {
    Command      string
    ExitCode     int
    Stdout       string
    Stderr       string
    Duration     time.Duration
    Timestamp    time.Time
    Success      bool
    ErrorType    ErrorType
}

type SafetyValidator struct {
    dangerousCommands []string
    requiredZone      string
    timeoutDuration   time.Duration
}
```

#### Real-Time Monitoring

```go
func (e *Executor) ExecuteWithMonitoring(cmd string) (*ExecutionResult, error) {
    // 1. Pre-execution validation
    // 2. Command safety checking
    // 3. Zone parameter injection
    // 4. Timeout-controlled execution
    // 5. Result capture and analysis
    // 6. Post-execution cleanup
}
```

---

## Terminal User Interface

**File**: `internal/tui/app.go` (524 lines)

### TUI Architecture

#### Component-Based Design

```go
type App struct {
    app            *tview.Application
    commandList    *tview.List
    detailsPanel   *tview.TextView
    resultsPanel   *tview.TextView
    helpPanel      *tview.TextView
    
    // State management
    commands       []ConvertedCommand
    selectedIndex  int
    helpVisible    bool
    
    // Interactive features
    executionMode  ExecutionMode
    filterSystem   *filter.System
}
```

#### Dynamic Layout Management

**Grid-Based Layout System**:
- **Commands Panel** (Left): Command selection and status
- **Details Panel** (Right-Top): Command details and transformation info  
- **Results Panel** (Right-Bottom): Execution results and logs
- **Help Panel** (Toggle): Context-sensitive help system

#### Interactive Features

1. **Command Selection**
   - Arrow key navigation
   - Space bar for selection toggle
   - Batch selection capabilities

2. **Execution Control**
   - Individual command execution (Enter)
   - Batch execution (e key)  
   - Dry run preview (d key)

3. **Help System** (✨ v1.9.0 Feature)
   - `?` key toggles help visibility
   - Dynamic layout adjustment
   - Context-sensitive guidance

#### Event Handling System

```go
func (a *App) setupKeyBindings() {
    a.commandList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyEnter:
            return a.executeCurrentCommand()
        case tcell.KeyRune:
            switch event.Rune() {
            case 'a': return a.selectAll()
            case 'n': return a.selectNone() 
            case 'e': return a.executeBatch()
            case '?': return a.toggleHelp()
            case 'q': return a.quit()
            }
        }
        return event
    })
}
```

---

## Testing Frameworks

### E2E Testing Framework

**File**: `tests/e2e/e2e_test_framework.go` (608 lines)

#### Testing Architecture

```go
type E2ETestSuite struct {
    t         *testing.T
    testDir   string
    binaryPath string
    tempFiles []string
}

type E2ETestOptions struct {
    Arguments           []string
    StdinInput         string
    Environment        map[string]string
    ExpectedExitCode   int
    ExpectedStdout     []string
    ExpectedStderr     []string
    ExpectedFiles      []FileExpectation
    ValidatePerformance bool
    MaxExecutionTime    string
}

type E2ETestResult struct {
    ExitCode         int
    Stdout           string
    Stderr           string
    ExecutionTime    time.Duration
    FilesCreated     []string
    EnvironmentDump  map[string]string
}
```

#### Test Capabilities

1. **Process Execution Testing**
   - Binary execution with timeout
   - Environment variable injection
   - Standard I/O capture

2. **File System Validation**
   - Output file verification
   - Content validation
   - Permission checking

3. **Performance Monitoring**
   - Execution time measurement
   - Memory usage tracking
   - Resource consumption analysis

### Integration Testing Framework

**File**: `tests/integration/integration_test_framework.go` (476 lines)

#### Integration Test Strategy

```go
type IntegrationTestFramework struct {
    scenarios []TestScenario
    config    *TestConfig
    reporter  *TestReporter
}

type TestScenario struct {
    Name         string
    Input        TestInput
    Expected     TestExpected
    Validation   ValidationConfig
}
```

#### Scenario-Based Testing

1. **User Workflow Scenarios**
   - Beginner user journey
   - Expert batch processing
   - Error recovery workflows

2. **Cross-Component Integration**
   - Transform + Validation integration
   - Config + Sandbox integration
   - TUI + Execution integration

### BDD Testing Framework

**File**: `internal/bdd/steps.go` (589 lines)

#### Behavior-Driven Development

```go
func (s *SandboxSteps) iRunUsacloudUpdateInSandboxMode() error {
    // Execute sandbox mode with validation
}

func (s *SandboxSteps) tuiInterfaceIsDisplayed() error {
    // Validate TUI interface rendering
}

func (s *SandboxSteps) listOfConvertedCommandsIsDisplayed() error {
    // Verify command conversion and display
}
```

#### BDD Capabilities

1. **Natural Language Scenarios**
   - Gherkin-format feature files
   - Business-readable test cases
   - Stakeholder-friendly documentation

2. **Automated Validation**
   - UI state verification
   - Command execution validation
   - Error scenario testing

---

## Extension Points

### 1. Transform Rule Extension

```go
// Implement the Rule interface for custom transformations
type CustomRule struct {
    name    string
    pattern *regexp.Regexp
    replacement string
}

func (r *CustomRule) Name() string { return r.name }

func (r *CustomRule) Apply(line string) (string, bool, string, string) {
    // Custom transformation logic
}

// Register with engine
engine := transform.NewEngine()
engine.AddRule(&CustomRule{...})
```

### 2. Validation Extension

```go
// Custom validator implementation
type CustomValidator struct {
    validationRules []ValidationRule
}

func (v *CustomValidator) Validate(cmd *ParsedCommand) []ValidationIssue {
    // Custom validation logic
}
```

### 3. TUI Component Extension

```go
// Custom TUI panel
type CustomPanel struct {
    *tview.TextView
    updateChannel chan string
}

func (p *CustomPanel) Update(data interface{}) {
    // Custom UI update logic
}
```

### 4. Configuration Extension

```go
// Custom configuration provider
type CustomConfigProvider struct {
    source string
}

func (p *CustomConfigProvider) Load() (*Config, error) {
    // Custom configuration loading
}
```

---

## Implementation Patterns

### 1. Error Handling Pattern

```go
// Structured error handling with context
type DetailedError struct {
    Code    ErrorCode
    Message string
    Context map[string]interface{}
    Cause   error
}

func (e *DetailedError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

### 2. Configuration Pattern

```go
// Priority-based configuration resolution
func ResolveConfig() *Config {
    config := &Config{}
    
    // Apply defaults
    applyDefaults(config)
    
    // Override with environment
    applyEnvironment(config)
    
    // Override with files
    applyConfigFile(config)
    
    // Override with flags
    applyCommandLine(config)
    
    return config
}
```

### 3. Plugin Pattern

```go
// Plugin interface for extensibility
type Plugin interface {
    Name() string
    Initialize(config *Config) error
    Process(input *ProcessInput) (*ProcessOutput, error)
    Cleanup() error
}

// Plugin registry
type PluginManager struct {
    plugins map[string]Plugin
}
```

### 4. Observer Pattern

```go
// Event notification system
type EventBus struct {
    subscribers map[EventType][]EventHandler
}

func (bus *EventBus) Notify(event *Event) {
    for _, handler := range bus.subscribers[event.Type] {
        go handler.Handle(event)
    }
}
```

---

## Performance and Scalability

### 1. Memory Management
- **Line-by-Line Processing**: Constant memory usage regardless of file size
- **Buffer Optimization**: 1MB buffer for handling long lines
- **Resource Cleanup**: Explicit cleanup of temporary resources

### 2. Concurrency
- **Safe Parallel Processing**: Channel-based communication
- **Background Tasks**: Non-blocking UI updates
- **Resource Synchronization**: Mutex protection for shared state

### 3. Caching Strategy
- **Rule Compilation**: Pre-compiled regular expressions
- **Command Validation**: Cached validation results
- **Configuration**: In-memory configuration caching

---

This implementation reference provides comprehensive coverage of the usacloud-update project's architecture, components, and extension mechanisms. Each section includes practical examples and implementation guidance for developers working with or extending the system.