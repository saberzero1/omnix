# Package `develop` - Development Shell Management

The `develop` package provides development shell management for Nix projects.

## Features

- **Pre-Shell Health Checks**: Validates Nix environment before entering dev shell
- **Automatic Cache Setup**: Configures cachix caches automatically
- **Post-Shell Welcome**: Displays project README after shell activation
- **Direnv Integration**: Works seamlessly with direnv

## Usage

### Basic Development Setup

```go
import (
    "context"
    "github.com/saberzero1/omnix/pkg/develop"
    "github.com/saberzero1/omnix/pkg/nix"
)

// Load configuration
config, _ := develop.LoadConfig("om.yaml")

// Parse flake URL
flake, _ := nix.ParseFlakeURL(".")

// Create project
project, _ := develop.NewProject(context.Background(), flake, config)

// Run develop workflow
err := develop.Run(context.Background(), project)
```

### Custom README Configuration

```go
config := develop.Config{
    Readme: develop.ReadmeConfig{
        File:   "DEVELOPMENT.md",
        Enable: true,
    },
}
```

## Configuration

Example `om.yaml` configuration:

```yaml
develop:
  readme:
    file: "README.md"
    enable: true
```

## Workflow

The development workflow consists of two phases:

### 1. Pre-Shell Phase

Runs before entering the development shell:

- Checks Nix version compatibility
- Validates system configuration (Rosetta, max-jobs)
- Configures required caches
- Ensures environment is properly set up

### 2. Post-Shell Phase

Runs after shell activation:

- Displays project README (markdown rendered)
- Shows welcome message
- Provides context for developers

## Health Checks

The package runs relevant health checks from `pkg/health`:

- **Nix Version**: Ensures minimum version requirements
- **Rosetta**: Checks Rosetta 2 on macOS (for x86_64 emulation)
- **Max Jobs**: Validates `max-jobs` configuration

### Cachix Integration

If cachix is available, the package can automatically configure caches:

```go
if develop.IsCachixAvailable() {
    _ = develop.UseCachixCache(ctx, "nixpkgs")
}
```

## Project Structure

### Project
Represents a Nix project being developed:

```go
type Project struct {
    Dir    *string          // Local directory (nil for remote)
    Flake  nix.FlakeURL     // Flake URL
    Config Config           // Development configuration
}
```

### Config
Development configuration from `om.yaml`:

```go
type Config struct {
    Readme ReadmeConfig
}

type ReadmeConfig struct {
    File   string  // Path to README file
    Enable bool    // Whether to display README
}
```

## Migration from Rust

This package replaces the `omnix-develop` Rust crate. Key changes:

- **Error Handling**: Rust's `Result<T, E>` → Go's `(T, error)`
- **Async**: Rust's `async fn` → Go's synchronous functions
- **Health Checks**: Integration with `pkg/health` package
- **Markdown**: Uses `pkg/common` for markdown rendering

## Examples

### Remote Flake

```go
flake, _ := nix.ParseFlakeURL("github:saberzero1/omnix")
project, _ := develop.NewProject(ctx, flake, config)
// project.Dir will be nil for remote flakes
```

### Local Flake

```go
flake, _ := nix.ParseFlakeURL(".")
project, _ := develop.NewProject(ctx, flake, config)
// project.Dir will be set to absolute path
```

### Custom Health Checks

Future enhancement - ability to customize which health checks run:

```go
// TODO: Add support for custom health check selection
```

## Testing

Run tests:
```bash
go test ./pkg/develop/...
```

With coverage:
```bash
go test -coverprofile=coverage.out ./pkg/develop
go tool cover -html=coverage.out
```

Integration tests (requires Nix):
```bash
go test ./pkg/develop/...
```

## Future Enhancements

- [ ] Actually invoke Nix devShell (currently just shows warning)
- [ ] Support for multiple development shells
- [x] ~~Custom health check selection~~ (Completed)
- [x] ~~Automatic direnv setup~~ (Completed)
- [ ] Shell hook customization
- [ ] Development environment templates
