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

This allows developers to inspect these values for debugging purposes.

## Implemented Features

The following features are now fully implemented:

1. **FlakeSchemas** (`pkg/nix/flake/schema.go`):
   - Complete schema inventory representation
   - Custom JSON marshaling/unmarshaling
   - Conversion to FlakeOutputs with "children" unwrapping

2. **GetFlakeSchemas()**:
   - Fetches schemas using inspect-flake
   - Uses compile-time paths from Nix build
   - Supports all known systems

3. **FromNix()**:
   - Constructs complete Flake from URL automatically
   - Single function call for full flake analysis
   - Example: `flake.FromNix(ctx, cmd, ".", flake.SystemLinuxX86_64)`
