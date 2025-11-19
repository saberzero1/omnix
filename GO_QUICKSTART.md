# Quick Start Guide for Go Development

This guide helps developers get started with the Go version of Omnix.

## Prerequisites

- Go 1.22 or later
- golangci-lint (for linting)

## Setup

```bash
# Clone the repository
git clone https://github.com/saberzero1/omnix
cd omnix

# Install dependencies
go mod download

# Install golangci-lint (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Common Commands

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...

# Run tests for specific package
go test ./pkg/common/...
```

### Building

```bash
# Build the binary
go build -o bin/om ./cmd/om

# Build for production (stripped, static)
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o bin/om ./cmd/om

# Run without building
go run ./cmd/om [args]
```

### Linting

```bash
# Run linter
golangci-lint run

# Run linter on specific files
golangci-lint run ./pkg/common/...

# Auto-fix issues where possible
golangci-lint run --fix
```

### Formatting

```bash
# Format all code
go fmt ./...

# Check formatting without modifying
gofmt -l .

# Format and fix imports
goimports -w .
```

## Project Structure

```
omnix/
├── cmd/om/              # Main binary
│   └── main.go
├── pkg/                 # Public packages
│   └── common/          # Common utilities (Phase 1 ✅)
│       ├── logging.go   # Logging setup
│       ├── check.go     # System checks
│       ├── fs.go        # Filesystem utilities
│       ├── config.go    # Configuration parsing
│       └── markdown.go  # Markdown rendering
├── internal/            # Private packages
│   └── testutil/        # Test utilities
├── go.mod              # Module definition
└── go.sum              # Dependency checksums
```

## Development Workflow

### 1. Make Changes
Edit Go files in `pkg/` or `cmd/`

### 2. Run Tests
```bash
go test ./...
```

### 3. Format Code
```bash
go fmt ./...
```

### 4. Run Linter
```bash
golangci-lint run
```

### 5. Build
```bash
go build -o bin/om ./cmd/om
```

### 6. Test Binary
```bash
./bin/om
```

## Writing Tests

### Table-Driven Tests (Recommended)

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Using Temp Directories

```go
func TestFileOperation(t *testing.T) {
    // Create temp directory (automatically cleaned up)
    dir := t.TempDir()
    
    // Create test file
    testFile := filepath.Join(dir, "test.txt")
    err := os.WriteFile(testFile, []byte("content"), 0644)
    if err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    // Run your test
    // ...
}
```

## Debugging

### Print Debugging
```go
import "fmt"

fmt.Printf("Debug: value = %+v\n", myStruct)
```

### Using Logger
```go
import "go.uber.org/zap"

logger := common.Logger()
logger.Debug("debugging message", zap.String("key", "value"))
```

### Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a test
dlv test ./pkg/common -- -test.run TestMyFunction

# Debug the binary
dlv debug ./cmd/om -- [args]
```

## Common Issues

### Import Cycle
If you get an import cycle error, you may need to:
- Move shared types to a separate package
- Use interfaces to break the cycle
- Reorganize package dependencies

### Race Conditions
Always run tests with race detector:
```bash
go test -race ./...
```

### Dependency Issues
```bash
# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy

# Verify dependencies
go mod verify
```

## Best Practices

1. **Error Handling**: Always handle errors explicitly
   ```go
   result, err := DoSomething()
   if err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }
   ```

2. **Naming**: Use descriptive names
   - Exported: `PascalCase`
   - Unexported: `camelCase`
   - Acronyms: `HTTPServer` not `HttpServer`

3. **Comments**: Document all exported functions
   ```go
   // DoSomething does something useful with the input.
   // It returns an error if the input is invalid.
   func DoSomething(input string) error {
       // ...
   }
   ```

4. **Testing**: Aim for 80%+ coverage
   - Test happy paths
   - Test error cases
   - Test edge cases

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [DESIGN_DOCUMENT.md](./DESIGN_DOCUMENT.md) - Full migration plan
- [GO_MIGRATION.md](./GO_MIGRATION.md) - Migration guide and patterns
- [PHASE1_SUMMARY.md](./PHASE1_SUMMARY.md) - Phase 1 completion summary

## Getting Help

- Check existing documentation in `doc/`
- Review code examples in tests
- Ask in GitHub Discussions
- Check the design document for architecture decisions

---

Happy coding! 🚀
