# pkg/cli

Command-line interface for omnix using Cobra.

## Overview

The `cli` package provides the user-facing command-line interface for omnix. It uses the Cobra framework to implement commands and argument parsing.

## Commands

### om

Root command that provides help and version information.

```bash
om --version    # Show version
om --help       # Show help
```

### om health

Check the health of your Nix installation.

```bash
om health              # Run all health checks with detailed output
om health --json       # Output results in JSON format
```

**Features:**
- Validates Nix version compatibility
- Checks if flakes are enabled
- Verifies cache configuration
- Platform-specific checks (Rosetta, Homebrew, etc.)
- Exit code 0 if all required checks pass, 1 otherwise

### om init

Initialize a new project from a template.

```bash
om init --template ./my-template output-dir/
om init --template ./template --param name=myapp --param author=me output/
om init --template ./template --non-interactive --param name=app output/
```

**Features:**
- Copy template directory structure
- Apply parameter substitutions (with --param)
- Interactive parameter prompting (when not in --non-interactive mode)
- Preserve file permissions and symlinks

**Flags:**
- `--template`: Path to template directory (required)
- `--param key=value`: Set template parameters
- `--non-interactive`: Disable interactive prompts (all params must be provided)

## Usage Example

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/saberzero1/omnix/pkg/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Architecture

```
pkg/cli/
├── root.go        # Root command definition and setup
├── cli_test.go    # CLI integration tests
├── doc.go         # Package documentation
├── README.md      # This file
└── cmd/           # Individual command implementations
    ├── health.go  # Health check command
    └── init.go    # Init command
```

## Command Flow

1. **Root Command** (`root.go`):
   - Initializes Cobra root command
   - Registers subcommands
   - Provides version and help

2. **Subcommands** (`cmd/`):
   - Parse command-line arguments
   - Validate inputs
   - Delegate to appropriate packages (health, init, etc.)
   - Format and display output
   - Return appropriate exit codes

3. **Package Integration**:
   - Health command → `pkg/health` package
   - Init command → `pkg/init` package
   - Nix info → `pkg/nix` package

## Testing

```bash
# Run CLI tests
go test ./pkg/cli/...

# Run all tests
go test ./pkg/...

# Build the binary
go build -o om ./cmd/om

# Test the binary
./om --version
./om health
./om init --help
```

## Exit Codes

- **0**: Success (all required checks passed for health, operation completed for init)
- **1**: Failure (required checks failed, or error during operation)

## Future Work

- [x] ~~Add JSON output for init command~~ (Completed - see ScaffoldAtWithResult)
- [ ] Implement registry support for template discovery
- [x] ~~Add interactive parameter prompting~~ (Available in pkg/init)
- [ ] Support loading templates from flakes
- [ ] Add progress indicators for long operations
- [ ] Implement remaining commands (show, ci, develop)
