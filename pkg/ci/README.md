# Package `ci` - CI/CD for Nix Projects

The `ci` package provides comprehensive CI/CD automation for Nix flakes.

## Features

- **Build Step**: Builds all flake outputs
- **Lockfile Check**: Verifies `flake.lock` is up to date
- **Flake Check**: Runs `nix flake check`
- **Custom Steps**: Execute arbitrary commands
- **GitHub Matrix Generation**: Generate GitHub Actions matrix configurations
- **Multi-System Support**: Build for multiple systems
- **Results Output**: JSON results for integration with CI systems
- **Parallel Execution**: Run subflakes in parallel for faster CI
- **Remote Builds**: Execute builds on remote hosts via SSH

## Usage

### Basic CI Run

```go
import (
    "context"
    "github.com/juspay/omnix/pkg/ci"
    "github.com/juspay/omnix/pkg/nix"
)

// Load configuration
config, _ := ci.LoadConfig("om.yaml")

// Parse flake URL
flake, _ := nix.ParseFlakeURL(".")

// Configure options
opts := ci.RunOptions{
    Systems:      []string{"x86_64-linux", "aarch64-darwin"},
    GitHubOutput: false,
}

// Run CI
results, _ := ci.Run(context.Background(), flake, config, opts)

// Check results
for _, result := range results {
    fmt.Printf("Subflake %s: success=%v\n", result.Subflake, result.Success)
}
```

### Parallel Execution

```go
opts := ci.RunOptions{
    Systems:        []string{"x86_64-linux"},
    Parallel:       true,          // Enable parallel execution
    MaxConcurrency: 4,              // Limit to 4 concurrent builds
}

results, _ := ci.Run(ctx, flake, config, opts)
```

### Remote Builds via SSH

```go
opts := ci.RunOptions{
    Systems:    []string{"x86_64-linux"},
    RemoteHost: "user@builder.example.com",  // Execute on remote host
}

results, _ := ci.Run(ctx, flake, config, opts)
```

### Generate GitHub Actions Matrix

```go
// Define systems to build for
systems := []string{"x86_64-linux", "aarch64-darwin"}

// Generate matrix
matrix := ci.GenerateMatrix(systems, config)

// Output as JSON
json, _ := matrix.ToJSON()
fmt.Println(json)
```

## Configuration

Example `om.yaml` configuration:

```yaml
ci:
  default:
    ".":
      dir: "."
      steps:
        build:
          enable: true
          impure: false
        lockfile:
          enable: true
        flakeCheck:
          enable: true
        custom:
          - name: "test"
            command: ["nix", "run", ".#test"]
            enable: true
    "tests":
      dir: "tests"
      systems:
        - "x86_64-linux"
      steps:
        build:
          enable: true
```

## CI Steps

### Build Step
Builds all flake outputs using `nix build`. Optionally supports `--impure` flag.

### Lockfile Step
Checks if `flake.lock` is up to date by running `nix flake lock --no-update-lock-file`.

### Flake Check Step
Runs `nix flake check` to validate the flake.

### Custom Steps
Execute custom commands. Useful for running tests, linters, or other tools.

## GitHub Actions Integration

The package generates matrix configurations compatible with GitHub Actions:

```yaml
jobs:
  ci:
    strategy:
      matrix:
        include:
          - system: x86_64-linux
            subflake: .
          - system: aarch64-darwin
            subflake: .
    runs-on: ${{ matrix.system }}
    steps:
      - uses: actions/checkout@v4
      - run: om ci run --systems ${{ matrix.system }}
```

## Migration from Rust

This package replaces the `omnix-ci` Rust crate. Key changes:

- **Error Handling**: Rust's `Result<T, E>` → Go's `(T, error)`
- **Concurrency**: Rust's async/await → Go's goroutines (future enhancement)
- **Config**: YAML parsing uses `gopkg.in/yaml.v3`
- **Nix Commands**: Uses `pkg/nix.Cmd` for command execution

## Testing

Run tests:
```bash
go test ./pkg/ci/...
```

With coverage:
```bash
go test -coverprofile=coverage.out ./pkg/ci
go tool cover -html=coverage.out
```

Integration tests (requires Nix):
```bash
go test ./pkg/ci/...
```
