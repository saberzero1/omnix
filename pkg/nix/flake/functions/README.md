# Flake Functions

This package provides a framework for calling Nix functions defined in flakes from Go, implementing a Go equivalent of the Rust `FlakeFn` trait.

## Overview

The flake functions pattern allows you to define "functions" in Nix flakes that can be called from external processes. The flake's package derivation acts as the function "body", with its inputs acting as function "arguments", and the built output acting as the function's "return value".

This generalizes the [devour-flake](https://github.com/srid/devour-flake) pattern to be able to define arbitrary functions.

## Architecture

### Core Framework (`core.go`)

The core module provides:
- `FlakeFn[Input, Output]` interface for defining flake functions
- `Call()` function to execute flake functions
- `CallOptions` for configuring function execution
- Utilities for handling boolean values (TRUE_FLAKE/FALSE_FLAKE)
- Environment variable resolution for function URLs

### Available Functions

#### Metadata Function (`metadata/`)

Retrieves metadata for a flake using Nix evaluation. This is an alternative to `nix flake metadata` that can include transitive inputs.

**Nix Implementation**: `metadata/flake.nix`

**Go API**:
```go
import "github.com/saberzero1/omnix/pkg/nix/flake/functions/metadata"

// Get flake path only
path, err := metadata.GetFlakePathOnly(ctx, "github:nixos/nixpkgs")

// Get flake with all transitive inputs
output, err := metadata.GetFlakeWithInputs(ctx, ".")
```

**Input**:
- `flake`: Flake URL to query
- `include-inputs`: Boolean - include transitive inputs

**Output**:
- `flake`: Store path to the flake
- `inputs`: List of input names and paths (if include-inputs is true)

#### AddStringContext Function (`addstringcontext/`)

Transforms a JSON file with Nix store paths so that the resultant JSON file will track those paths as dependencies. Requires `--impure`.

**Nix Implementation**: `addstringcontext/flake.nix`

**Go API**:
```go
import "github.com/saberzero1/omnix/pkg/nix/flake/functions/addstringcontext"

// Transform a JSON file
storePath, err := addstringcontext.TransformJSONFile(ctx, "./output.json")

// Transform JSON data directly
data := map[string]interface{}{
    "outPaths": []string{"/nix/store/..."},
}
storePath, err := addstringcontext.AddStringContextToData(ctx, data, "result")
```

**Input**:
- `jsonfile`: Path to JSON file (as flake URL, e.g., "path:./file.json")

**Output**: JSON file with string context added to `outPaths` and `allDeps` fields

## How It Works

1. **Input Serialization**: Input structs are serialized to `--override-input` arguments
   - String values are used as-is
   - Boolean `true` becomes `TRUE_FLAKE` URL
   - Boolean `false` becomes `FALSE_FLAKE` URL

2. **Nix Build**: Runs `nix build` with the flake function URL and override inputs

3. **Output Parsing**: Reads JSON from the built store path and deserializes into output struct

4. **Initialization**: Calls `Init()` on the output for any post-processing

## Environment Variables

These are set by the Nix build environment (see `nix/envs/default.nix`):

- `FLAKE_METADATA`: URL to the metadata flake function
- `FLAKE_ADDSTRINGCONTEXT`: URL to the addstringcontext flake function  
- `TRUE_FLAKE`: URL to a flake representing boolean true
- `FALSE_FLAKE`: URL to a flake representing boolean false

Fallback values are provided when these are not set (e.g., during development).

## Creating New Flake Functions

1. Create a new directory under `functions/` (e.g., `myfunction/`)
2. Add a `flake.nix` that implements your function logic
3. Create a Go file implementing `FlakeFn[Input, Output]`
4. Add an environment variable in `nix/envs/default.nix` pointing to your flake
5. Add tests and documentation

Example:

```go
package myfunction

import (
	"context"
	"os"

	"github.com/saberzero1/omnix/pkg/nix/flake/functions"
)

type MyFunctionFn struct{}

func (f MyFunctionFn) FlakeURL() string {
	if url := os.Getenv("MY_FUNCTION"); url != "" {
		return url
	}
	return "path:./pkg/nix/flake/functions/myfunction#default"
}

func (f MyFunctionFn) Init(out *MyOutput) {
	// Optional post-processing
}

type MyInput struct {
	Arg1 string `json:"arg1"`
	Arg2 bool   `json:"arg2"`
}

type MyOutput struct {
	Result string `json:"result"`
}

func CallMyFunction(ctx context.Context, input MyInput) (MyOutput, error) {
	fn := MyFunctionFn{}
	_, output, err := functions.Call(ctx, fn, nil, input)
	return output, err
}
```

## References

- [devour-flake](https://github.com/srid/devour-flake) - Original implementation
- [Nix String Context](https://nix.dev/manual/nix/2.23/language/string-context)
- Rust implementation: `crates/nix_rs/src/flake/functions/`
