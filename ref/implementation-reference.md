# Implementation Reference

## Overview

The usacloud-update project is a command-line tool that automatically transforms usacloud v0.x scripts to v1.1 compatibility. The implementation follows a clean, layered architecture with clear separation of concerns.

## Architecture Layers

### 1. CLI Layer (`cmd/usacloud-update/main.go`)

**Purpose**: User-facing command-line interface that handles I/O operations and user feedback.

**Key Features**:
- **Flexible I/O**: Supports stdin/stdout or file-based input/output via `-in` and `-out` flags
- **Statistics Reporting**: Colored change summary to stderr (controllable via `-stats` flag)
- **Large File Support**: 1MB buffer for processing large shell scripts
- **User Feedback**: Detailed line-by-line transformation reports with rule names

**Implementation Highlights**:
- Uses `bufio.Scanner` with extended buffer capacity
- Integrates `fatih/color` for terminal output formatting
- Processes files line-by-line for memory efficiency
- Automatically adds transformation header to output files

### 2. Engine Layer (`internal/transform/engine.go`)

**Purpose**: Central orchestration of the transformation process.

**Core Types**:
```go
type Change struct {
    RuleName string
    Before   string
    After    string
}

type Result struct {
    Line     string
    Changed  bool
    Changes  []Change
}

type Engine struct {
    rules []Rule
}
```

**Key Functions**:
- `NewDefaultEngine()`: Creates engine with standard rule set
- `Apply(line string) Result`: Transforms a single line using all applicable rules

**Smart Processing**:
- Skips empty lines and comments automatically
- Applies rules sequentially
- Tracks all changes for detailed reporting
- Provides whitespace normalization utilities

### 3. Rule System (`internal/transform/rules.go`)

**Purpose**: Extensible rule infrastructure using regex-based pattern matching.

**Core Implementation**:
```go
type simpleRule struct {
    name   string
    re     *regexp.Regexp
    repl   func([]string) string
    reason string
    url    string
}
```

**Rule Features**:
- **Regex Matching**: Compiled patterns for efficient matching
- **Dynamic Replacements**: Custom replacement functions with capture group access
- **Automatic Documentation**: Adds explanatory comments with reasons and URLs
- **Duplicate Prevention**: Avoids adding multiple usacloud-update comments
- **Fragment Tracking**: Captures before/after snippets for reporting

**Helper Function**:
```go
func mk(name, pattern string, repl func([]string) string, reason, url string) Rule
```

### 4. Business Logic (`internal/transform/ruledefs.go`)

**Purpose**: Contains all specific transformation rules for usacloud migration.

**Generated Header**:
```bash
# Updated for usacloud v1.1 by usacloud-update — DO NOT EDIT ABOVE THIS LINE
```

**Transformation Categories** (9 total):

1. **Output Format Migration**
   - Pattern: `--output-type=(csv|tsv)` → `--output-type=json`
   - Reason: CSV/TSV deprecated in v1.0

2. **Selector Deprecation** 
   - Pattern: `--selector value` → `value`
   - Reason: --selector flag removed, use arguments instead

3. **Resource Name Changes**
   - `iso-image` → `cdrom`
   - `startup-script` → `note` 
   - `ipv4` → `ipaddress`

4. **Product Alias Cleanup**
   - `product-disk` → `disk-plan`
   - `product-internet` → `internet-plan`
   - `product-server` → `server-plan`

5. **Command Deprecation**
   - `summary` → commented out (removed in v1)
   - `object-storage`/`ojs` → commented out (no longer supported)

6. **Parameter Normalization**
   - `--zone = all` → `--zone=all`

**Rule Documentation**: Each rule includes:
- Descriptive reason for the change
- Link to relevant official documentation
- Context about version compatibility

### 5. Test Suite (`internal/transform/engine_test.go`)

**Purpose**: Golden file testing to ensure transformation accuracy.

**Testing Strategy**:
- **End-to-End Validation**: Tests complete file transformation pipeline
- **Golden Files**: Compares output against pre-approved expected results
- **Update Mode**: `--update` flag for regenerating expectations
- **Real Data**: Uses actual mixed v0/v1 script as test input

**Key Test**: `TestGolden_SampleMixed`
- Input: `testdata/sample_v0_v1_mixed.sh`
- Expected: `testdata/expected_v1_1.sh`
- Process: Transforms entire file and compares byte-for-byte

## Code Organization

```
usacloud-update/
├── cmd/usacloud-update/     # CLI entry point
│   └── main.go             # Command-line interface
├── internal/transform/      # Core transformation logic
│   ├── engine.go           # Rule orchestration
│   ├── rules.go            # Rule infrastructure  
│   ├── ruledefs.go         # Specific transformation rules
│   └── engine_test.go      # Golden file tests
├── testdata/               # Test fixtures
│   ├── sample_v0_v1_mixed.sh    # Input test data
│   └── expected_v1_1.sh         # Expected output
└── ref/                    # Documentation
```

## Extension Points

### Adding New Rules

1. Add rule definition in `ruledefs.go`:
```go
rules = append(rules, mk(
    "rule-name",
    `regex-pattern`,
    func(m []string) string { return replacement },
    "reason for change",
    "https://documentation-url",
))
```

2. Update golden test file if needed:
```bash
go test -run Golden -update ./...
```

### Custom Rule Types

Implement the `Rule` interface:
```go
type Rule interface {
    Name() string
    Apply(line string) (after string, changed bool, before string, afterFragment string)
}
```

## Performance Considerations

- **Memory Efficient**: Line-by-line processing avoids loading entire files
- **Compiled Regexes**: Patterns compiled once during engine initialization  
- **Sequential Processing**: Rules applied in order for predictable results
- **Large Buffer**: 1MB buffer handles large shell scripts efficiently

## Error Handling

- **File Operations**: Comprehensive error checking for I/O operations
- **Regex Compilation**: Validation during rule creation
- **Test Failures**: Clear reporting when transformations don't match expectations
- **User Feedback**: Detailed error messages with context

## Documentation Integration

Each transformation includes:
- Clear explanation of why the change is needed
- Link to relevant official documentation
- Version compatibility information
- Example usage patterns

This ensures users understand not just what changed, but why, enabling them to make informed decisions about their migration strategy.