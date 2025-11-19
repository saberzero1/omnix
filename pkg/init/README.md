# pkg/init

Template initialization and scaffolding for Nix projects.

## Overview

The `init` package provides template-based project initialization from Nix flake templates. This is the Go implementation of the Rust `omnix-init` crate.

## Features

- **Template Scaffolding**: Copy template directories with full structure preservation
- **Parameter Substitution**: Replace placeholders in file contents and names
- **Conditional Pruning**: Delete files/directories based on parameters
- **Action System**: Flexible action-based template customization
- **Glob Patterns**: Support for glob pattern matching

## Usage

### Basic Template Scaffolding

```go
package main

import (
    "context"
    
    "github.com/saberzero1/omnix/pkg/init"
)

func main() {
    ctx := context.Background()
    
    // Helper function for string pointers
    strPtr := func(s string) *string { return &s }
    
    // Create a template
    template := &init.Template{
        Path: "./my-template",
        Params: []init.Param{
            {
                Name:        "project-name",
                Description: "The name of your project",
                Action: &init.ReplaceAction{
                    Placeholder: "MYPROJECT",
                    Value:       strPtr("awesome-app"),
                },
            },
        },
    }
    
    // Scaffold the template
    outPath, err := template.ScaffoldAt(ctx, "./my-new-project")
    if err != nil {
        panic(err)
    }
    
    println("Project created at:", outPath)
}
```

### Parameter Types

#### Replace Action

Replaces placeholder text in file contents and filenames:

```go
{
    Name:        "author",
    Description: "Project author name",
    Action: &init.ReplaceAction{
        Placeholder: "AUTHOR_NAME",
        Value:       strPtr("John Doe"),
    },
}
```

This will:
1. Replace all occurrences of "AUTHOR_NAME" in file contents
2. Rename any files/directories containing "AUTHOR_NAME"

#### Retain Action

Conditionally keeps or deletes files matching glob patterns:

```go
{
    Name:        "include-ci",
    Description: "Include CI configuration",
    Action: &init.RetainAction{
        Paths: []string{".github/**", "*.yml"},
        Value: boolPtr(false),  // false = delete matching paths
    },
}
```

When `Value` is `false`, all matching files are deleted.

### Setting Parameter Values

```go
template := &init.Template{...}

// Set values from a map
values := map[string]interface{}{
    "project-name": "my-app",
    "include-ci":   true,
}
template.SetParamValues(values)

// Scaffold with the values
outPath, _ := template.ScaffoldAt(ctx, "./output")
```

## Action System

### Action Interface

All actions implement the `Action` interface:

```go
type Action interface {
    HasValue() bool
    Apply(ctx context.Context, outDir string) error
    String() string
}
```

### Action Ordering

Actions are applied in priority order:
1. **Retain** actions (pruning) run first
2. **Replace** actions run second

This ensures files are deleted before text replacement occurs, preventing unnecessary work on files that will be deleted.

## Examples

### Full Example with Multiple Parameters

```go
boolPtr := func(b bool) *bool { return &b }
strPtr := func(s string) *string { return &s }

template := &init.Template{
    Path:        "./templates/rust-app",
    Description: strPtr("Rust application template"),
    Params: []init.Param{
        {
            Name:        "name",
            Description: "Project name",
            Action: &init.ReplaceAction{
                Placeholder: "PROJECT_NAME",
                Value:       strPtr("my-rust-app"),
            },
        },
        {
            Name:        "author",
            Description: "Author name",
            Action: &init.ReplaceAction{
                Placeholder: "AUTHOR",
                Value:       strPtr("Jane Developer"),
            },
        },
        {
            Name:        "use-actix",
            Description: "Include Actix web framework",
            Action: &init.RetainAction{
                Paths: []string{"src/actix/**"},
                Value: boolPtr(true),
            },
        },
        {
            Name:        "use-tokio",
            Description: "Include async/tokio support",
            Action: &init.RetainAction{
                Paths: []string{"src/async/**"},
                Value: boolPtr(false),  // Remove async code
            },
        },
    },
}

outPath, err := template.ScaffoldAt(context.Background(), "./my-project")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Created project at:", outPath)
```

## Test Coverage

- **Coverage**: 66.3%
- **Status**: All tests passing
- **Integration Tests**: Available (test actual file operations)

## Architecture

```
pkg/init/
├── action.go      # Action interface and implementations
├── template.go    # Template and scaffolding logic
├── init_test.go   # Comprehensive tests
├── doc.go        # Package documentation
└── README.md     # This file
```

## Migration Notes

Migrated from Rust `omnix-init` crate:
- **Rust LOC**: ~656
- **Go LOC**: ~515 (21% reduction)
- **Key Changes**:
  - No async file operations (synchronous in Go)
  - Interface-based action system
  - Simplified glob matching
  - Direct filesystem operations

## Limitations

Current implementation:
- ✅ Basic template scaffolding
- ✅ Replace and Retain actions
- ✅ Glob pattern support
- ⚠️ No Nix flake integration (registry)
- ⚠️ No interactive prompting
- ⚠️ No test execution support

## Future Work

- [ ] Add registry support for template discovery
- [x] ~~Implement interactive parameter prompting~~ (Completed - see prompt.go)
- [ ] Add template test execution
- [ ] Support for loading templates from flakes
- [x] ~~Improve glob pattern matching (full globset support)~~ (Enhanced with ** support)
- [ ] Add more sophisticated file operation support
