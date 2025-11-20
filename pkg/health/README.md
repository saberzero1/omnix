# pkg/health

Health checks for Nix installations.

## Overview

The `health` package provides a comprehensive set of health checks to validate a Nix environment's configuration and ensure it meets recommended or required standards. This is the Go implementation of the Rust `omnix-health` crate.

## Features

- **Multiple Health Checks**: Validates various aspects of a Nix installation
- **Configurable**: Checks can be enabled/disabled and configured
- **Clear Results**: Each check provides clear pass/fail status with actionable suggestions
- **Exit Codes**: Aggregated results provide appropriate exit codes for CI/CD

## Health Checks

| Check | Required | Description |
|-------|----------|-------------|
| FlakeEnabled | Yes | Verifies that Nix flakes and nix-command are enabled |
| NixVersion | Yes | Validates the Nix version meets minimum requirements (≥2.16.0) |
| Caches | Yes | Checks that required binary caches are configured |
| TrustedUsers | No* | Validates that the current user is in trusted-users |
| MaxJobs | No | Checks max-jobs configuration |
| Rosetta | No | Checks for Rosetta 2 on Apple Silicon Macs |
| Direnv | No | Verifies direnv is installed |
| Homebrew | No | Checks for Homebrew on macOS |
| Shell | No | Validates shell configuration |

*TrustedUsers check is disabled by default for security reasons

## Usage

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/saberzero1/omnix/pkg/health"
    "github.com/saberzero1/omnix/pkg/nix"
)

func main() {
    ctx := context.Background()
    
    // Get Nix installation info
    nixInfo, err := nix.GetInfo(ctx)
    if err != nil {
        panic(err)
    }
    
    // Create health checks with default configuration
    healthChecks := health.Default()
    
    // Run all checks
    results := healthChecks.RunAllChecks(ctx, nixInfo)
    
    // Print results
    for _, result := range results {
        health.PrintCheckResult(result)
    }
    
    // Evaluate and get exit code
    status := health.EvaluateResults(results)
    fmt.Println(status.SummaryMessage())
}
```

## Configuration

Health checks can be configured via `om.yaml`:

```yaml
health:
  nix-version:
    min-version: "2.18.0"
  caches:
    required:
      - "https://cache.nixos.org"
      - "https://my-cache.cachix.org"
  trusted-users:
    enable: true  # Enable the check (disabled by default)
```

## Test Coverage

- **Coverage**: 81.1% ✅ (exceeds 80% target)
- **Status**: All unit tests passing
- **Integration Tests**: Available (skipped in short mode)

## Architecture

The package is organized to avoid import cycles:

```
pkg/health/
├── health.go           # Aggregation and result evaluation
├── health_test.go      # Aggregation tests
├── doc.go             # Package documentation
└── checks/            # Individual check implementations
    ├── types.go       # Shared types (Check, CheckResult, Checkable)
    ├── flake_enabled.go
    ├── nix_version.go
    ├── caches.go
    ├── trusted_users.go
    ├── max_jobs.go
    ├── rosetta.go
    ├── direnv.go
    ├── homebrew.go
    ├── shell.go
    └── checks_test.go
```

## Migration Notes

Migrated from Rust `omnix-health` crate:
- **Rust LOC**: ~1,283
- **Go LOC**: ~1,006 (21% reduction)
- **Key Changes**:
  - No async/await complexity (synchronous checks)
  - Simpler error handling (no Result types)
  - Interface-based check system
  - Type definitions moved to checks package to avoid cycles

## Future Work

- [x] ~~Implement full Config parsing for all check types~~ (Completed)
- [ ] Add integration tests with real Nix
- [x] ~~Improve test coverage to ≥80%~~ (Completed - 81.1%)
- [x] ~~Add markdown rendering for check output~~ (Completed)
- [x] ~~Support loading configuration from om.yaml~~ (Completed)
