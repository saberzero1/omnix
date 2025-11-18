// Package ci provides CI/CD automation for Nix projects.
//
// This package implements the "om ci" command functionality, which runs
// comprehensive CI/CD pipelines for Nix flakes. It includes:
//   - Building all flake outputs
//   - Checking flake lock files
//   - Running flake checks
//   - Custom step execution
//   - GitHub Actions matrix generation
//   - Remote build support (SSH)
//   - Results JSON output
//
// Example usage:
//
//	import (
//	    "context"
//	    "github.com/juspay/omnix/pkg/ci"
//	    "github.com/juspay/omnix/pkg/nix"
//	)
//
//	// Load configuration
//	config, _ := ci.LoadConfig("om.yaml")
//
//	// Run CI for a flake
//	flake, _ := nix.ParseFlakeUrl(".")
//	result, _ := ci.Run(ctx, flake, config, options)
//
// The package supports running CI steps in parallel and can generate
// GitHub Actions matrix configurations for cross-platform testing.
package ci
