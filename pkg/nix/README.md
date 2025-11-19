# pkg/nix

Go package for interacting with the Nix package manager.

## Overview

The `nix` package provides a comprehensive Go API for working with Nix. It includes functionality for:

- **Version Detection**: Parse and compare Nix versions
- **Command Execution**: Run Nix commands with context support
- **Flake URLs**: Parse and manipulate Nix flake URLs
- **Environment Detection**: Identify OS, user, and system configuration
- **Configuration**: Read and parse Nix configuration
- **Installation Info**: Aggregate system and Nix installation information

## Installation

```bash
go get github.com/saberzero1/omnix/pkg/nix
```

## Quick Start

### Get Nix Version

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/saberzero1/omnix/pkg/nix"
)

func main() {
    cmd := nix.NewCmd()
    version, err := cmd.RunVersion(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Nix version: %s\n", version)
}
```

### Execute Nix Commands

```go
cmd := nix.NewCmd()
output, err := cmd.Run(context.Background(), "flake", "show", ".")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output)
```

### Parse JSON Output

```go
var result map[string]interface{}
err := cmd.RunJSON(context.Background(), &result, "eval", "--expr", "{foo = 42;}", "--json")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Result: %v\n", result)
```

### Work with Flake URLs

```go
// Create a flake URL
url := nix.NewFlakeURL(".")

// Add an attribute
withAttr := url.WithAttr("packages.x86_64-linux.default")
fmt.Println(withAttr) // Output: .#packages.x86_64-linux.default

// Check if it's a local path
if url.IsLocal() {
    fmt.Printf("Local path: %s\n", url.AsLocalPath())
}

// Split base and attribute
base, attr := url.SplitAttr()
fmt.Printf("Base: %s, Attr: %s\n", base, attr)
```

### Detect Environment

```go
env, err := nix.DetectEnv(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User: %s\n", env.CurrentUser)
fmt.Printf("OS: %s\n", env.OS)
fmt.Printf("Groups: %v\n", env.CurrentUserGroups)

// Check OS type
if env.OS.IsNixOS {
    fmt.Println("Running on NixOS")
} else if env.OS.IsNixDarwin {
    fmt.Println("Running on nix-darwin")
}
```

### Get Configuration

```go
config, err := nix.GetConfig(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("System: %s\n", config.System.Value)
fmt.Printf("Max jobs: %d\n", config.MaxJobs.Value)

// Check if flakes are enabled
if config.IsFlakesEnabled() {
    fmt.Println("Flakes are enabled")
}

// Check for specific features
if config.HasFeature("nix-command") {
    fmt.Println("nix-command is enabled")
}
```

### Get Complete Installation Info

```go
info, err := nix.GetInfo(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Println(info) // Output: Nix 2.13.0 on Linux
fmt.Printf("Version: %s\n", info.Version)
fmt.Printf("OS: %s\n", info.Env.OS)
fmt.Printf("User: %s\n", info.Env.CurrentUser)
```

## API Reference

### Types

#### `Version`
Represents a Nix version with major, minor, and patch numbers.

Methods:
- `String() string` - Format as "major.minor.patch"
- `Compare(other Version) int` - Compare versions (-1, 0, 1)
- `LessThan(other Version) bool`
- `GreaterThan(other Version) bool`
- `Equal(other Version) bool`

#### `Cmd`
Nix command executor with context support.

Methods:
- `NewCmd() *Cmd` - Create a new command executor
- `RunVersion(ctx context.Context) (Version, error)` - Get Nix version
- `Run(ctx context.Context, args ...string) (string, error)` - Run command, return text
- `RunJSON(ctx context.Context, result interface{}, args ...string) error` - Run command, parse JSON

#### `FlakeURL`
Represents a Nix flake URL.

Methods:
- `NewFlakeURL(url string) FlakeURL` - Create from string
- `String() string` - Get URL as string
- `AsLocalPath() string` - Get local path (empty if not local)
- `IsLocal() bool` - Check if URL is a local path
- `WithAttr(attr string) FlakeURL` - Add/replace attribute
- `SplitAttr() (string, string)` - Split into base URL and attribute
- `Clean() FlakeURL` - Clean/normalize the URL

#### `Env`
Environment information where Nix operates.

Fields:
- `CurrentUser string` - Current user name
- `CurrentUserGroups []string` - User's groups
- `OS OSType` - Operating system information

#### `OSType`
Operating system type and configuration.

Fields:
- `Type string` - OS type ("darwin", "linux")
- `IsNixOS bool` - Running on NixOS
- `IsNixDarwin bool` - Running on nix-darwin
- `Arch string` - Architecture ("amd64", "arm64")

Methods:
- `String() string` - Human-readable OS description
- `NixConfigLabel() string` - Configuration location label

#### `Config`
Nix configuration from `nix show-config`.

Fields:
- `ExperimentalFeatures ConfigValue[[]string]`
- `System ConfigValue[string]`
- `Substituters ConfigValue[[]string]`
- `MaxJobs ConfigValue[int]`
- `Cores ConfigValue[int]`

Methods:
- `GetConfig(ctx context.Context) (*Config, error)` - Retrieve configuration
- `IsFlakesEnabled() bool` - Check if flakes are enabled
- `HasFeature(feature string) bool` - Check for specific experimental feature

#### `ConfigValue[T]`
Generic type for configuration values with metadata.

Fields:
- `Value T` - Current value
- `DefaultValue T` - Default value
- `Description string` - Configuration description

#### `Info`
Complete Nix installation information.

Fields:
- `Version Version` - Nix version
- `Env *Env` - Environment information

Methods:
- `GetInfo(ctx context.Context) (*Info, error)` - Get all info
- `String() string` - Format as "Nix X.Y.Z on OS"

### Functions

#### Version Parsing
```go
func ParseVersion(s string) (Version, error)
```

Parse a Nix version string. Supports multiple formats:
- Standard: "nix (Nix) 2.13.0"
- Simple: "2.13.0"
- Determinate: "nix (Determinate Nix 3.6.6) 2.29.0"

#### Environment Detection
```go
func DetectEnv(ctx context.Context) (*Env, error)
```

Detect the current Nix environment including user, groups, and OS type.

#### Flake URL Parsing
```go
func ParseFlakeURL(s string) (FlakeURL, error)
```

Parse a string into a FlakeURL.

## Error Handling

The package uses standard Go error handling patterns:

```go
cmd := nix.NewCmd()
output, err := cmd.Run(ctx, "flake", "show")
if err != nil {
    var cmdErr *nix.CommandError
    if errors.As(err, &cmdErr) {
        fmt.Printf("Command failed with exit code %d\n", cmdErr.ExitCode)
        fmt.Printf("Stderr: %s\n", cmdErr.Stderr)
    }
    return err
}
```

## Context Support

All command execution functions accept a `context.Context` for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

version, err := cmd.RunVersion(ctx)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Command timed out")
    }
    return err
}
```

## Testing

The package includes comprehensive unit tests and integration tests. Unit tests can be run without Nix installed:

```bash
go test -short ./pkg/nix
```

Integration tests require Nix to be installed:

```bash
go test ./pkg/nix
```

## Coverage

The package maintains high test coverage (>75%) with comprehensive edge case testing.

## Migration from Rust

This package is part of the omnix Rust-to-Go migration. It provides equivalent functionality to the Rust `nix_rs` crate with idiomatic Go patterns:

- Rust `async fn` → Go synchronous functions with context
- Rust `Result<T, E>` → Go `(T, error)` tuple returns
- Rust `Option<T>` → Go zero values or pointers
- Rust tokio → Go context.Context for cancellation

## License

AGPL-3.0
