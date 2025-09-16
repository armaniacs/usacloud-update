# Developer Onboarding Guide

A practical guide for developers to quickly understand and contribute to the usacloud-update project.

## 🚀 Quick Start (5 Minutes)

### Prerequisites
- Go 1.24.1+
- Basic understanding of CLI tools
- Familiarity with Sakura Cloud usacloud command

### First Steps
```bash
# 1. Clone and setup
git clone <repository>
cd usacloud-update
make build

# 2. Run basic transformation
echo 'usacloud server ls --output-type csv' | ./usacloud-update
# Output: usacloud server ls --output-type json  # Converted!

# 3. Run tests to verify everything works
make test
```

## 📚 Understanding the Project (15 Minutes)

### What This Tool Does
usacloud-update transforms bash scripts containing mixed v0, v1.0, and v1.1 usacloud commands to work consistently with v1.1.

**Example Transformation**:
```bash
# Input (old v0/v1.0 commands)
usacloud server ls --output-type csv
usacloud iso-image ls
usacloud server create --selector "name=web01" 

# Output (v1.1 compatible)
usacloud server ls --output-type json  # csv→json 
usacloud cdrom ls                      # iso-image→cdrom
usacloud server create web01          # selector→argument
```

### Project Architecture
```
usacloud-update/
├── cmd/usacloud-update/main.go          # CLI entry point
├── internal/
│   ├── transform/                       # Core transformation engine
│   │   ├── engine.go                    # Rule application engine
│   │   ├── rules.go                     # Rule interface & base implementation
│   │   └── ruledefs.go                  # Actual transformation rules (9 categories)
│   ├── validation/                      # Command validation system
│   ├── sandbox/                         # Real usacloud command execution
│   ├── config/                          # Configuration management
│   └── tui/                            # Terminal user interface
├── testdata/                            # Sample scripts & expected output
└── ref/                                # Documentation
```

### Key Concepts

1. **Rules**: Individual transformation patterns (csv→json, iso-image→cdrom)
2. **Engine**: Applies rules line-by-line to input
3. **Validator**: Checks command correctness and suggests fixes
4. **Sandbox**: Executes commands in real Sakura Cloud environment for testing
5. **TUI**: Interactive terminal interface for command selection

## 🔧 Common Development Tasks

### Adding a New Transformation Rule

**File**: `internal/transform/ruledefs.go`

```go
// Add to DefaultRules() function
mk(
    "new-rule-name",
    `old-pattern-regex`,
    `new-replacement`,
    "URL_TO_DOCUMENTATION"
),
```

**Example**:
```go
mk(
    "network-interface-rename",
    `usacloud nic\b`,
    `usacloud interface`,
    "https://docs.sakura.ad.jp/xxx"
),
```

**Test your rule**:
```bash
# Update golden files with your new transformation
make golden

# Verify tests pass
make test
```

### Extending Command Validation

**File**: `internal/validation/main_command_validator.go`

Add new commands to the validator:
```go
// In initializeCommands()
commands["new-command"] = true
```

### Adding TUI Features

**File**: `internal/tui/app.go`

Key methods to understand:
- `setupKeyBindings()`: Add new keyboard shortcuts
- `onCommandSelected()`: Handle command execution
- `updateDetailView()`: Customize information display

### Configuration Changes

**File**: `internal/config/integrated_config.go`

Add new configuration options:
```go
type GeneralConfig struct {
    NewOption bool `ini:"new_option"`
    // ... existing fields
}
```

## 🧪 Testing Your Changes

### Test Hierarchy (run in order)
```bash
# 1. Unit tests (fastest)
go test ./internal/transform/
go test ./internal/validation/

# 2. Golden file tests (transformation accuracy)
make test

# 3. Integration tests (BDD scenarios)
make bdd

# 4. Manual verification
echo "usacloud server ls --output-type csv" | ./usacloud-update
```

### Understanding Golden File Testing
Golden file testing compares actual output with expected output:
- **Input**: `testdata/sample_v0_v1_mixed.sh`
- **Expected output**: `testdata/expected_v1_1.sh`
- **Update expected**: `make golden`

When you add a new rule, the expected output changes, so you must run `make golden` to update the expectation.

## 🐛 Debugging Common Issues

### "make test" Fails After Adding Rule
**Problem**: New transformation changed expected output
**Solution**: 
```bash
make golden  # Update expected output
make test    # Should pass now
```

### Command Not Recognized
**Problem**: Added command but validation fails
**Solution**: Add command to validator in `internal/validation/main_command_validator.go`

### TUI Crashes
**Problem**: UI component error
**Solution**: Check terminal size requirements and error handling in `internal/tui/app.go`

### Configuration Not Loading
**Problem**: New config option not appearing
**Solution**: 
1. Add to struct in `integrated_config.go`
2. Add default value in `createDefaultConfig()`
3. Test with `usacloud-update config`

## 📖 Code Reading Strategy

### For New Contributors (Read in this order)

1. **Start Here**: `cmd/usacloud-update/main.go`
   - Understand CLI structure and entry points
   - See how different modes (transform/sandbox/tui) are selected

2. **Core Logic**: `internal/transform/engine.go`
   - Understand how transformation works
   - See the Rule interface and application logic

3. **Actual Rules**: `internal/transform/ruledefs.go`
   - See all 9 categories of transformations
   - Understand rule patterns and comments

4. **Validation**: `internal/validation/main_command_validator.go`
   - Understand command validation logic
   - See error handling and suggestions

5. **Advanced Features**: Pick one based on interest
   - Sandbox: `internal/sandbox/executor.go`
   - TUI: `internal/tui/app.go`
   - Config: `internal/config/integrated_config.go`

### For Bug Fixes
1. Reproduce the issue with a minimal test case
2. Identify which component handles that functionality
3. Add debug logging to understand the flow
4. Write a test case for your fix
5. Ensure all existing tests still pass

### For New Features
1. Check if similar functionality exists
2. Identify the appropriate component or create new one
3. Design the interface first (structs, functions)
4. Implement with comprehensive error handling
5. Add tests (unit + golden file if it affects output)
6. Update documentation

## 🎯 Contribution Guidelines

### Code Standards
- **Go formatting**: `make fmt` before committing
- **Testing**: All new code must have tests
- **Documentation**: Update relevant docs in `/ref`
- **Comments**: Add explanatory comments for transformation rules

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and test
make build test

# Commit with descriptive message
git commit -m "feat: add transformation for network-interface commands"

# Update golden files if needed
make golden
git add testdata/expected_v1_1.sh
git commit -m "test: update golden files for network-interface transformation"
```

### Pull Request Process
1. Ensure all tests pass locally
2. Update documentation if needed
3. Add example usage in PR description
4. Include before/after transformation examples

## 🔍 Useful Development Commands

```bash
# Development workflow
make build          # Build binary
make test           # Run all tests
make golden         # Update golden files
make bdd            # Run behavior tests
make clean          # Clean build artifacts

# Debugging
make run            # Run with sample data
make verify-sample  # Test against known good output

# Code quality
make fmt            # Format code
make vet            # Static analysis
make tidy           # Clean up modules
```

## 📞 Getting Help

### When You're Stuck
1. **Check existing tests**: They show expected behavior
2. **Read similar code**: Find similar functionality elsewhere
3. **Use debug output**: Add logging to understand flow
4. **Test incrementally**: Make small changes and test frequently

### Resources
- **Architecture docs**: `/ref/component-architecture.md`
- **API reference**: `/ref/api-reference.md`
- **Detailed implementation**: `/ref/detailed-implementation-reference.md`
- **Current state**: `/ref/current-implementation-state.md`

### Code Examples
Most components have comprehensive test files that serve as usage examples:
- `internal/transform/engine_test.go`: How to use the transform engine
- `internal/validation/main_command_validator_test.go`: Validation examples
- `internal/config/integrated_config_test.go`: Configuration usage

## 🎉 Success Metrics

You'll know you're ready to contribute when you can:
- [ ] Build and run the project successfully
- [ ] Add a simple transformation rule and test it
- [ ] Understand the flow from CLI input to output
- [ ] Run and interpret the test suite
- [ ] Make a small change and verify it doesn't break existing functionality

Welcome to the usacloud-update project! 🚀