# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## ドキュメントのルール

- CLAUDE.md 以外のファイルは日本語で書く
- @CHANGELOG.md に更新についてサマリを記録する

## Overview

usacloud-update is a comprehensive Go CLI tool that transforms bash scripts containing mixed v0, v1.0, and v1.1 usacloud commands to work consistently with v1.1. Beyond transformation, v1.9.0 introduces sandbox execution, interactive TUI, and comprehensive testing framework. The tool applies transformation rules line-by-line and can execute commands in Sakura Cloud's sandbox environment for real testing.

## Key Commands

### Build and Run
```bash
make build        # Build the binary (includes tidy and fmt)
make run          # Build and run with sample input
make test         # Run unit tests (golden file comparison)
make golden       # Update golden files with current output
make verify-sample # Run sample and diff against expected output
```

### Development
```bash
make tidy         # go mod tidy
make fmt          # go fmt ./...
make vet          # go vet ./...
make clean        # Remove bin/ directory
```

### Testing
- **Comprehensive Test Coverage**: 56.1% coverage with 8 new test files (5,175+ lines)
- **Golden file testing**: `make test` compares output against `testdata/expected_v1_1.sh`
- **BDD testing**: `make bdd` runs behavior-driven tests for sandbox functionality
- **Edge case testing**: Concurrent access, error conditions, boundary values
- To update expected output after rule changes: `make golden`
- Input samples: `testdata/sample_v0_v1_mixed.sh`, `testdata/mixed_with_non_usacloud.sh`

## Architecture

### Core Components

**Main Application (`cmd/usacloud-update/main.go`)**
- **Dual-mode CLI**: Traditional conversion vs sandbox execution
- **Sandbox flags**: `--sandbox`, `--interactive`, `--dry-run`, `--batch`
- **Flexible input**: stdin, files, directory scanning with TUI file selector
- **Multi-file processing**: Batch processing of multiple script files
- Line-by-line processing with colored output to stderr for changes

**Transform Engine (`internal/transform/engine.go`)**
- `Engine` struct applies rules sequentially to each line
- `Result` contains transformed line and change metadata  
- Skips empty lines and comments (lines starting with `#`)
- **100% test coverage** achieved

**Rule System (`internal/transform/rules.go` + `ruledefs.go`)**
- `Rule` interface: `Name()` and `Apply(line) (newLine, changed, beforeFrag, afterFrag)`
- `simpleRule` implementation uses regex pattern matching
- Rules automatically append explanatory comments with URLs
- `DefaultRules()` returns all transformation rules
- **100% test coverage** with comprehensive rule validation

**Sandbox Execution (`internal/sandbox/executor.go`) - v1.9.0**
- Real usacloud command execution in Sakura Cloud tk1v zone
- Command validation, timeout protection, rate limiting
- Comprehensive result tracking and error handling
- **78.5% test coverage** with edge case testing

**Terminal UI (`internal/tui/`) - v1.9.0**
- Interactive command selection and execution interface
- File selector with directory scanning capabilities
- Help panel toggle with `?` key, dynamic layout management
- **62.7% test coverage** with comprehensive TUI testing

**Configuration Management (`internal/config/`) - v1.9.0**
- File-based configuration (`~/.config/usacloud-update/usacloud-update.conf`)
- Interactive configuration setup for first-time users
- Environment variable support for legacy compatibility
- **57.2% test coverage** with environment handling tests

### Rule Categories

1. **Output format**: CSV/TSV → JSON
2. **Selector migration**: `--selector` → command arguments  
3. **Resource renaming**: `iso-image` → `cdrom`, `startup-script` → `note`, `ipv4` → `ipaddress`
4. **Product aliases**: `product-*` → `*-plan`
5. **Removed commands**: `summary`, `object-storage` (commented out with manual action required)
6. **Zone normalization**: `--zone = all` → `--zone=all`

### Testing Strategy

**Multi-layered Testing Approach**:
- **Golden file testing**: Transform sample input and compare against expected output
- **Comprehensive unit testing**: 8 dedicated test files with 5,175+ lines of test code
- **Edge case testing**: Concurrent access, error conditions, boundary value testing
- **BDD testing**: Behavior-driven development with Godog framework
- **Integration testing**: Cross-component interaction validation

**Test Management**:
- Run `go test -run Golden -update ./...` to update golden files
- `make golden` wraps this for convenience
- `make bdd` runs BDD tests for sandbox functionality
- **Overall test coverage: 56.1%** across all packages

## Development Notes

- All transformation rules add explanatory comments with relevant documentation URLs
- Rules are applied sequentially - order matters for overlapping patterns
- Manual intervention required for deprecated commands (marked with comments)
- Large buffer size (1MB) used for handling long bash script lines

## Detailed Reference Documentation

For comprehensive implementation details, see the `/ref` directory:

### 🏗️ Core Architecture Documentation
- **[Component Architecture](/ref/component-architecture.md)** - Complete v1.9.0 component structure, data flow, and inter-component relationships
- **[Architecture Overview v2](/ref/architecture-v2.md)** - System architecture with sandbox execution, TUI, and multi-mode CLI
- **[Code Organization](/ref/code-organization.md)** - File structure, naming conventions, coding standards, and project layout
- **[Core Algorithms](/ref/core-algorithms.md)** - Detailed algorithms for Transform Engine, Validation System, and TUI components

### 📚 Implementation References
- **[Current Implementation State](/ref/current-implementation-state.md)** - ✨**NEW**: Concise overview of what's actually implemented and working in v1.9.3
- **[Developer Onboarding Guide](/ref/developer-onboarding-guide.md)** - ✨**NEW**: Quick start guide for new developers to understand and contribute to the project
- **[Implementation Overview](/ref/implementation-overview.md)** - Comprehensive implementation details, component architecture, and system design patterns
- **[Implementation Reference](/ref/implementation-reference.md)** - Technical implementation details, code organization, and extension points
- **[Detailed Implementation Reference](/ref/detailed-implementation-reference.md)** - In-depth implementation guide with 1,313+ lines covering all core components, extension points, and implementation patterns
- **[API Reference](/ref/api-reference.md)** - Full API documentation for all types, interfaces, and functions

### 🔧 Core Systems Documentation
- **[Transformation Rules](/ref/transformation-rules.md)** - Complete catalog of all 9 rule categories with examples and implementation details

### 🧪 Development and Testing
- **[Testing Guide](/ref/testing-guide.md)** - Multi-layered testing methodology with 56.1% coverage details
- **[Testing Framework Reference](/ref/testing-framework-reference.md)** - ✨**NEW**: Comprehensive testing framework documentation covering E2E, Integration, BDD, Performance, and Golden File testing
- **[Test Coverage Report](/ref/test-coverage-report.md)** - Comprehensive coverage analysis and quality metrics
- **[Test Data Reference](/ref/test-data-reference.md)** - Test data structure, transformation examples, and validation processes
- **[Development Workflow](/ref/development-workflow.md)** - Complete development processes, debugging strategies, and release preparation

### 🚀 Build and Deployment
- **[Build & Deployment](/ref/build-deployment.md)** - Build system, deployment options, CI/CD integration, and troubleshooting
- **[Project Dependencies](/ref/project-dependencies.md)** - Dependency analysis, version management, and platform support

### ⚡ Advanced System Components
- **[Performance Optimization Engine](/ref/performance-optimization-engine.md)** - Comprehensive resource management, intelligent caching, CPU scheduling, and adaptive optimization
- **[Monitoring System Architecture](/ref/monitoring-system-architecture.md)** - Enterprise-grade monitoring with metric collection, alerting, and web-based dashboards

### 📖 Legacy Architecture (for reference)
- **[Architecture Overview (v1)](/ref/architecture.md)** - Original architecture documentation for historical reference

## Quick Start for Developers

### Understanding the Codebase
1. **Start Here**: [Current Implementation State](/ref/current-implementation-state.md) - ✨**NEW**: What's actually implemented and working right now
2. **New Developer Guide**: [Developer Onboarding Guide](/ref/developer-onboarding-guide.md) - ✨**NEW**: 5-minute quick start for new contributors
3. **Architecture Overview**: [Component Architecture](/ref/component-architecture.md) - Get an overview of all system components
4. **Implementation Deep Dive**: [Detailed Implementation Reference](/ref/detailed-implementation-reference.md) - Comprehensive guide to all core components and implementation patterns
5. **Code Structure**: [Code Organization](/ref/code-organization.md) - Understand file layout and naming conventions
6. **Core Logic**: [Core Algorithms](/ref/core-algorithms.md) - Deep dive into Transform Engine and Validation System algorithms
7. **API Integration**: [API Reference](/ref/api-reference.md) - Complete API documentation for system integration

### Development Setup
```bash
# Clone and setup
git clone <repository>
cd usacloud-update

# Install dependencies
go mod tidy

# Build and test
make build
make test
```

### Platform-specific Build Guides

For detailed platform-specific build instructions, troubleshooting, and optimization:

- **[Windows Build Guide](README-Build-Windows.md)** - PowerShell/Git Bash development environment
- **[Linux Build Guide](README-Build-Linux.md)** - Ubuntu/CentOS/Debian/Fedora support
- **[macOS Build Guide](README-Build-macOS.md)** - Intel/Apple Silicon Mac support

Each guide includes comprehensive setup instructions, common issues resolution, performance optimization tips, and cross-compilation procedures.

### Common Development Tasks
- **Quick Start**: Use [Developer Onboarding Guide](/ref/developer-onboarding-guide.md) for step-by-step instructions
- **Current Capabilities**: Check [Current Implementation State](/ref/current-implementation-state.md) to understand what's working
- **Add new transformation rule**: See [Transformation Rules](/ref/transformation-rules.md) and [Developer Onboarding Guide](/ref/developer-onboarding-guide.md#adding-a-new-transformation-rule)
- **Extend validation**: Check [Core Algorithms](/ref/core-algorithms.md) validation section and [Detailed Implementation Reference](/ref/detailed-implementation-reference.md#validation-system)
- **Add TUI features**: Reference [Component Architecture](/ref/component-architecture.md) TUI system and [Detailed Implementation Reference](/ref/detailed-implementation-reference.md#terminal-user-interface)
- **Testing**: Follow [Testing Guide](/ref/testing-guide.md) methodology and [Testing Framework Reference](/ref/testing-framework-reference.md) for comprehensive testing strategies
- **API Integration**: Use [API Reference](/ref/api-reference.md) for system integration and extension

### Project Status (v1.9.3)
- ✅ **Stable**: Core transformation engine, validation system, basic TUI
- ✅ **Feature Complete**: Sandbox execution, BDD testing, configuration management, environment variable detection
- ✅ **Ready for v2.0**: All core features implemented with 100% make test success rate
- 🔄 **In Progress**: Documentation updates, test coverage improvement
- ⏸️ **Postponed**: Advanced features (see [NOT-TODO.md](NOT-TODO.md) and [TODO-target-2_0.md](TODO-target-2_0.md))

**For detailed current status, see**: [Current Implementation State](/ref/current-implementation-state.md)