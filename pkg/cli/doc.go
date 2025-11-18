// Package cli provides the command-line interface for omnix.
//
// This package implements the Cobra-based CLI framework for the omnix tool,
// providing user-facing commands for Nix operations.
//
// # Overview
//
// The CLI package defines:
//   - Root command (om) with version and help
//   - Subcommands for health checks (om health)
//   - Subcommands for project initialization (om init)
//   - Command-line argument parsing and validation
//
// # Usage
//
// The CLI is invoked through the Execute function:
//
//	package main
//	
//	import (
//	    "fmt"
//	    "os"
//	    
//	    "github.com/juspay/omnix/pkg/cli"
//	)
//	
//	func main() {
//	    if err := cli.Execute(); err != nil {
//	        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//	        os.Exit(1)
//	    }
//	}
//
// # Commands
//
// ## om health
//
// Checks the health of a Nix installation:
//
//	om health              # Run all health checks
//	om health --json       # Output results in JSON format
//
// ## om init
//
// Initialize a new project from a template:
//
//	om init --template ./template output/
//	om init --template ./template --param name=myapp output/
//
// # Architecture
//
// The CLI package is organized into:
//   - pkg/cli/root.go: Root command definition
//   - pkg/cli/cmd/: Individual command implementations
//   - pkg/cli/cli_test.go: CLI tests
//
// Commands delegate to the appropriate packages (health, init, etc.) for
// actual functionality, keeping the CLI layer thin and focused on user
// interaction.
package cli
