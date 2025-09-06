# Build and Deployment Reference

## Build System

### Makefile Configuration

The project uses a comprehensive Makefile for build automation and development tasks.

**Key Variables**:
```makefile
GO       ?= go
BINARY   := usacloud-update
BIN_DIR  := bin
CMD_PKG  := ./cmd/$(BINARY)
PKGS     := ./...
```

**Test Data Configuration**:
```makefile
IN_SAMPLE  := testdata/sample_v0_v1_mixed.sh
OUT_SAMPLE := /tmp/out.sh
GOLDEN     := testdata/expected_v1_1.sh
```

### Available Make Targets

#### Core Build Targets

##### `make build`
- **Purpose**: Compile the binary with code quality checks
- **Dependencies**: Runs `tidy` and `fmt` first
- **Output**: Creates `bin/usacloud-update` executable
- **Command**: `$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)`

##### `make all` (default)
- **Purpose**: Alias for `make build`
- **Use Case**: Default target for quick builds

#### Development and Testing

##### `make run`
- **Purpose**: Build and execute with sample data
- **Dependencies**: Requires successful build
- **Action**: Processes test input and writes to `/tmp/out.sh`
- **Command**: `$(BIN_DIR)/$(BINARY) --in $(IN_SAMPLE) --out $(OUT_SAMPLE)`

##### `make test`
- **Purpose**: Run unit tests and golden file tests
- **Coverage**: Tests all packages in the project
- **Command**: `$(GO) test $(PKGS)`

##### `make golden`
- **Purpose**: Update golden test files with current output
- **Use Case**: After implementing new transformation rules
- **Command**: `$(GO) test -run Golden -update $(PKGS)`
- **⚠️ Warning**: Only run after verifying new output is correct

##### `make verify-sample`
- **Purpose**: Manual verification of transformations
- **Process**: 
  1. Runs transformation on sample input
  2. Compares output with golden file
  3. Shows differences (if any)
- **Command**: `diff -u $(GOLDEN) $(OUT_SAMPLE) || true`

#### Code Quality

##### `make tidy`
- **Purpose**: Clean up Go module dependencies
- **Command**: `$(GO) mod tidy`
- **Effect**: Removes unused dependencies, adds missing ones

##### `make fmt`
- **Purpose**: Format Go source code
- **Command**: `$(GO) fmt $(PKGS)`
- **Standard**: Applies standard Go formatting rules

##### `make vet`
- **Purpose**: Static analysis of Go code
- **Command**: `$(GO) vet $(PKGS)`
- **Checks**: Suspicious constructs, potential bugs

#### Cleanup

##### `make clean`
- **Purpose**: Remove build artifacts
- **Action**: Deletes `bin/` directory
- **Command**: `rm -rf $(BIN_DIR)`

## Build Process

### Standard Development Workflow

1. **Code Changes**: Make modifications to source code
2. **Format and Tidy**: `make fmt tidy`
3. **Build**: `make build`
4. **Test**: `make test`
5. **Manual Verification**: `make verify-sample`

### Quality Assurance Pipeline

```bash
# Full quality check
make tidy fmt vet test
make build
make verify-sample
```

### Golden File Update Process

When adding or modifying transformation rules:

1. **Implement Rule**: Add rule in `internal/transform/ruledefs.go`
2. **Run Tests**: `make test` (will likely fail)
3. **Update Golden**: `make golden`
4. **Verify Output**: Review `testdata/expected_v1_1.sh` changes
5. **Retest**: `make test` (should now pass)
6. **Manual Check**: `make verify-sample`

## Deployment Options

### Go Install (Recommended)

**Direct Installation**:
```bash
go install github.com/armaniacs/usacloud-update/cmd/usacloud-update@latest
```

**Benefits**:
- Automatically handles dependencies
- Installs to `$GOPATH/bin` or `$GOBIN`
- Easy version management
- No manual build required

**Requirements**:
- Go 1.24.1 or later
- Internet connection for dependency download
- GOPATH/GOBIN in system PATH

### From Source

**Clone and Build**:
```bash
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update
make build
```

**Manual Installation**:
```bash
# Copy binary to system path
sudo cp bin/usacloud-update /usr/local/bin/
```

**Benefits**:
- Full control over build process
- Can modify source before building
- Useful for development and debugging

### Binary Distribution

**Creating Distributable Binary**:
```bash
# Build for current platform
make build

# Binary location
ls -la bin/usacloud-update
```

**Cross-Platform Builds**:
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/usacloud-update-linux-amd64 ./cmd/usacloud-update

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/usacloud-update-windows-amd64.exe ./cmd/usacloud-update

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o bin/usacloud-update-darwin-arm64 ./cmd/usacloud-update
```

**Distribution Benefits**:
- No Go installation required on target systems
- Single executable file
- No runtime dependencies
- Platform-specific optimization

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.24.1'
    
    - name: Run tests
      run: |
        make tidy fmt vet
        make test
        make build
        make verify-sample
    
    - name: Build for multiple platforms
      run: |
        GOOS=linux GOARCH=amd64 make build
        GOOS=windows GOARCH=amd64 make build
        GOOS=darwin GOARCH=amd64 make build
```

### Docker Build

**Dockerfile Example**:
```dockerfile
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/usacloud-update .
ENTRYPOINT ["./usacloud-update"]
```

## Performance Considerations

### Build Optimization

**Compilation Flags**:
```bash
# Optimized build
go build -ldflags="-s -w" -o bin/usacloud-update ./cmd/usacloud-update

# With version information
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/usacloud-update ./cmd/usacloud-update
```

**Size Optimization**:
- `-s`: Strip symbol table and debug information
- `-w`: Strip DWARF debug information
- Results in smaller binary size

### Runtime Performance

**Memory Efficiency**:
- Line-by-line processing avoids loading large files into memory
- 1MB buffer size optimized for typical shell script sizes
- Compiled regex patterns cached during initialization

**Processing Speed**:
- Sequential rule application with early termination
- Efficient string operations using Go's optimized string package
- Minimal allocations during transformation process

## Troubleshooting

### Common Build Issues

#### Go Version Compatibility
```bash
# Check Go version
go version

# Minimum required: go1.24.1
```

#### Module Issues
```bash
# Clean and rebuild modules
go clean -modcache
go mod download
make tidy
```

#### Missing Dependencies
```bash
# Verify all dependencies
go mod verify

# Install missing tools
go install golang.org/x/tools/cmd/goimports@latest
```

### Test Failures

#### Golden File Mismatches
```bash
# View differences
make verify-sample

# Update if changes are intentional
make golden

# Rerun tests
make test
```

#### Environment Issues
```bash
# Check test environment
echo $TMPDIR
ls -la testdata/

# Verify file permissions
chmod +x bin/usacloud-update
```

## Version Management

### Release Process

1. **Update Version**: Tag with semantic version
2. **Build and Test**: Full quality pipeline
3. **Cross-Platform Builds**: Generate binaries for all platforms
4. **Documentation**: Update changelog and documentation
5. **Release**: Create GitHub release with binaries

### Semantic Versioning

- **Major**: Breaking changes to CLI interface or output format
- **Minor**: New transformation rules or features
- **Patch**: Bug fixes and documentation improvements

Example: `v1.2.3`
- Major: 1 (stable CLI interface)
- Minor: 2 (new transformation rules added)
- Patch: 3 (bug fixes)