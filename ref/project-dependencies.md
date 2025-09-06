# Project Dependencies

## Go Module Configuration

**Module**: `github.com/armaniacs/usacloud-update`  
**Go Version**: 1.24.1

## Direct Dependencies

### Core Dependencies

#### github.com/fatih/color v1.18.0
- **Purpose**: Terminal color output for transformation statistics
- **Usage**: Provides colored text output in CLI for better user experience
- **Key Features**:
  - Cross-platform color support
  - ANSI color codes for Unix-like systems
  - Windows console color support
  - Color detection and fallback handling

**Used in**: `cmd/usacloud-update/main.go`
```go
color.YellowString("#L%-5d %s => %s [%s]\n", lineNum, before, after, ruleName)
```

## Indirect Dependencies

### github.com/mattn/go-colorable v0.1.13
- **Purpose**: Cross-platform colored terminal support
- **Relationship**: Transitive dependency of fatih/color
- **Function**: Provides colorable writer interface for Windows compatibility

### github.com/mattn/go-isatty v0.0.20  
- **Purpose**: TTY detection for color output decisions
- **Relationship**: Transitive dependency of fatih/color
- **Function**: Determines if output is connected to a terminal

### golang.org/x/sys v0.25.0
- **Purpose**: System-level functionality
- **Relationship**: Transitive dependency for low-level system calls
- **Function**: Provides platform-specific system interfaces

## Standard Library Usage

### Primary Standard Library Packages

- **bufio**: Buffered I/O for efficient file processing
- **flag**: Command-line flag parsing
- **fmt**: Formatted I/O operations
- **io**: Basic I/O primitives
- **os**: Operating system interface
- **strings**: String manipulation utilities
- **regexp**: Regular expression matching
- **testing**: Testing framework for golden file tests

## Dependency Analysis

### Minimal Dependency Footprint
The project maintains a very small dependency footprint with only one direct dependency:
- **Single External Dependency**: Only fatih/color for terminal output enhancement
- **Standard Library Focus**: Heavy reliance on Go standard library for core functionality
- **No Heavy Frameworks**: Avoids web frameworks, ORMs, or other heavy dependencies

### Security Considerations
- **Trusted Sources**: All dependencies from well-known, maintained repositories
- **Minimal Attack Surface**: Small dependency tree reduces potential security vulnerabilities
- **No Network Dependencies**: No dependencies that make network calls or handle external resources

### Version Management
- **Semantic Versioning**: All dependencies follow semantic versioning
- **Stable Versions**: Uses stable releases rather than development branches
- **Compatibility**: Dependencies chosen for Go 1.24.1 compatibility

## Build Requirements

### Go Version
- **Minimum**: Go 1.24.1
- **Recommendation**: Use Go 1.24.1 or later for full compatibility
- **Features Used**: Modern Go features requiring recent version

### Build Tools
- **go mod**: Dependency management via Go modules
- **go build**: Standard Go build toolchain
- **go test**: Built-in testing framework
- **make**: Build automation via Makefile

### Development Dependencies
While not in go.mod, the project expects:
- **make**: For build automation and development tasks
- **git**: For version control and development workflow
- **diff**: For test result comparison (typically available on Unix systems)

## Installation Methods

### Via Go Install
```bash
go install github.com/armaniacs/usacloud-update/cmd/usacloud-update@latest
```

### From Source
```bash
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update
make build
```

### Binary Distribution
The built binary has no runtime dependencies and can be distributed as a single executable file.

## Platform Support

### Supported Platforms
- **Linux**: Full support with colored output
- **macOS**: Full support with colored output  
- **Windows**: Full support with colored output via go-colorable
- **Other Unix**: Should work on any platform supported by Go

### Color Output Support
- **Unix-like Systems**: Native ANSI color support
- **Windows**: Color support via mattn/go-colorable
- **TTY Detection**: Automatic color disabling for non-terminal output

## License Compatibility

All dependencies use permissive licenses compatible with the project's MIT license:
- **fatih/color**: MIT License
- **mattn/go-colorable**: MIT License  
- **mattn/go-isatty**: MIT License
- **golang.org/x/sys**: BSD-3-Clause License

## Maintenance Considerations

### Dependency Updates
- **Regular Updates**: Monitor for security updates in dependencies
- **Testing**: Run full test suite after dependency updates
- **Compatibility**: Verify Go version compatibility with updates

### Minimal Maintenance Burden
- **Small Dependency Tree**: Few dependencies to monitor and update
- **Stable Dependencies**: Uses well-maintained, stable libraries
- **Standard Library Focus**: Reduces external maintenance dependencies