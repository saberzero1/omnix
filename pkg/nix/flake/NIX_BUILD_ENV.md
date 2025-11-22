# Nix Build Environment Integration

This document explains how the Go build integrates with the Nix build environment to access flake-related resources.

## Environment Variables

The Nix build system injects the following environment variables into the Go binary at compile time:

- `DEFAULT_FLAKE_SCHEMAS`: Path to the flake-schemas flake used for analyzing flake outputs
- `INSPECT_FLAKE`: Path to the inspect flake for querying flake metadata

These are defined in `nix/envs/default.nix` and injected via ldflags in `nix/modules/flake/go.nix`.

## Go API

The `pkg/nix/flake` package provides functions to access these values:

```go
import "github.com/saberzero1/omnix/pkg/nix/flake"

// Check if running with Nix-built binary
if flake.HasNixBuildEnvironment() {
    schemas := flake.GetDefaultFlakeSchemas()  // Path to flake-schemas
    inspect := flake.GetInspectFlake()         // Path to inspect flake
    
    // Use these paths to analyze flakes...
} else {
    // Running with `go build` - these features unavailable
}
```

## Building

### With Nix (recommended)
```bash
nix build
# The binary will have environment variables injected
```

### Without Nix (development)
```bash
go build ./cmd/om
# The binary will work but HasNixBuildEnvironment() returns false
```

## Development Shell

When using `nix develop`, the environment variables are also exported:

```bash
nix develop
echo $DEFAULT_FLAKE_SCHEMAS
echo $INSPECT_FLAKE
```

This allows Go code to use these values at runtime when needed.

## Future Work

The `FromNix()` method on the `Flake` type can now be implemented using:
1. `GetInspectFlake()` to get the inspect flake path
2. `GetDefaultFlakeSchemas()` to get the flake-schemas path
3. The existing `Eval()` function to query the inspect flake
4. Parse the results into `FlakeOutputs`

This requires implementing the FlakeSchemas parsing logic, which involves handling the inventory format returned by the inspect flake.
