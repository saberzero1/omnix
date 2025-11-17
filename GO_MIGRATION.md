# Go Migration Guide

This document tracks the progress of migrating Omnix from Rust to Go according to the DESIGN_DOCUMENT.md.

## Current Status: Phase 1 - Foundation ✅

### Completed Tasks

#### 1. Go Module Structure
- ✅ Initialized Go module (`github.com/juspay/omnix`)
- ✅ Created directory structure:
  - `cmd/om/` - Main binary entry point
  - `pkg/common/` - Shared utilities (replaces omnix-common)
  - `internal/testutil/` - Internal test utilities

#### 2. Package Migration: `pkg/common`

Successfully migrated all modules from `omnix-common` Rust crate:

| Rust Module | Go Module | Status | Test Coverage |
|-------------|-----------|--------|---------------|
| `logging.rs` | `logging.go` | ✅ Complete | ✅ Covered |
| `check.rs` | `check.go` | ✅ Complete | ✅ Covered |
| `fs.rs` | `fs.go` | ✅ Complete | ✅ Covered |
| `config.rs` | `config.go` | ✅ Complete | ✅ Covered |
| `markdown.rs` | `markdown.go` | ✅ Complete | ✅ Covered |

**Overall Test Coverage: 80.4%** ✅ (Meets 80% target)

#### 3. Dependencies

| Purpose | Rust Crate | Go Package |
|---------|-----------|------------|
| Structured Logging | `tracing` | `go.uber.org/zap` |
| YAML Parsing | `serde_yaml` | `gopkg.in/yaml.v3` |
| Markdown Rendering | `pulldown-cmark-mdcat` | `github.com/charmbracelet/glamour` |

#### 4. Development Tools

- ✅ golangci-lint configuration (`.golangci.yml`)
- ✅ Updated `.gitignore` for Go artifacts
- ✅ Added Go targets to `justfile`

#### 5. Quality Assurance

- ✅ All tests passing
- ✅ Linting passing (golangci-lint)
- ✅ Code formatted (gofmt)
- ✅ Binary builds successfully

## Quick Start

### Run Tests
```bash
just go-test
```

### Run Tests with Coverage
```bash
just go-test-coverage
```

### Run Linter
```bash
just go-lint
```

### Build Binary
```bash
just go-build
```

### Run Complete CI
```bash
just go-ci
```

### Run the Binary
```bash
./bin/om
# or
just go-run
```

## Key Design Decisions

### 1. Logging: tracing → zap
- Used `go.uber.org/zap` for high-performance structured logging
- Implemented similar verbosity levels (Error, Warn, Info, Debug, Trace)
- Supports both bare and formatted output modes
- Respects `OMNIX_LOG` environment variable

### 2. Markdown: pulldown-cmark → glamour
- Used `charmbracelet/glamour` for terminal markdown rendering
- Provides auto-styling based on terminal capabilities
- Similar API to Rust version

### 3. Config: serde → standard library + yaml.v3
- Used `encoding/json` and `gopkg.in/yaml.v3`
- Implemented flexible config tree structure
- Maintains compatibility with existing om.yaml format

### 4. Filesystem: async → standard library
- Converted async operations to synchronous (Go doesn't need async for file I/O)
- Preserved symlink handling
- Maintained permission management

## Next Steps: Phase 2 - Core Nix Integration

The next phase will focus on migrating the `nix_rs` crate to `pkg/nix`:

- [ ] `command.go` - Nix command execution
- [ ] `flake.go` - Flake operations
- [ ] `store.go` - Store path operations
- [ ] `config.go` - Nix configuration parsing
- [ ] Integration tests with actual Nix commands

## Migration Patterns

### Error Handling
**Rust:**
```rust
use anyhow::{Context, Result};

fn do_something() -> Result<String> {
    fs::read_to_string(path)
        .context("Failed to read file")?;
}
```

**Go:**
```go
import "fmt"

func doSomething() (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }
    return string(content), nil
}
```

### Async/Await → Standard Functions
**Rust:**
```rust
async fn copy_dir(src: &Path, dst: &Path) -> Result<()> {
    // async implementation
}
```

**Go:**
```go
func CopyDir(src, dst string) error {
    // synchronous implementation
}
```

## Testing Philosophy

- Unit tests for all public functions
- Table-driven tests for multiple scenarios
- Test coverage target: ≥80%
- Integration tests in separate files
- Use `testify` for assertions where helpful

## Notes

- Rust async operations converted to synchronous Go code (Go handles concurrency differently)
- Go's error handling is more explicit than Rust's `Result<T, E>` type
- Package naming follows Go conventions (lowercase, no underscores)
- Exported functions use PascalCase, unexported use camelCase
