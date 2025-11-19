// Package ci provides CI/CD automation for Nix projects.
//
// This package implements the "om ci" command functionality, which runs
// comprehensive CI/CD pipelines for Nix flakes. It includes:
//   - Building all flake outputs
//   - Checking flake lock files
//   - Running flake checks
//   - Custom step execution
//   - GitHub Actions matrix generation
//   - Parallel subflake execution
//   - Remote build support via SSH
//   - Results JSON output
//
// Example usage:
//
//	import (
//	    "context"
//	    "github.com/saberzero1/omnix/pkg/ci"
//	    "github.com/saberzero1/omnix/pkg/nix"
//	)
//
//	// Load configuration
//	config, _ := ci.LoadConfig("om.yaml")
//
//	// Run CI for a flake
//	flake, _ := nix.ParseFlakeURL(".")
//	opts := ci.RunOptions{
//	    Systems:    []string{"x86_64-linux"},
//	    Parallel:   true,                    // Run in parallel
//	    RemoteHost: "user@remote.host",      // Optional: remote builds
//	}
//	result, _ := ci.Run(ctx, flake, config, opts)
//
// The package supports running CI steps in parallel for improved performance
// and can execute builds on remote hosts via SSH. It also generates
// GitHub Actions matrix configurations for cross-platform testing.
package ci
