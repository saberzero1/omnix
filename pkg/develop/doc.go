// Package develop provides development shell management for Nix projects.
//
// This package implements the "om develop" command functionality, which helps
// set up and manage development environments for Nix flakes. It includes:
//   - Pre-shell health checks
//   - Automatic cache configuration (via cachix)
//   - Post-shell welcome messages
//   - Direnv integration
//
// Example usage:
//
//	import (
//	    "context"
//	    "github.com/juspay/omnix/pkg/develop"
//	    "github.com/juspay/omnix/pkg/nix"
//	)
//
//	// Create a project for development
//	flake, _ := nix.ParseFlakeURL(".")
//	project, _ := develop.NewProject(ctx, flake, config)
//
//	// Run development shell setup
//	err := develop.Run(ctx, project)
//
// The package performs health checks before entering the shell and displays
// helpful README information after shell activation.
package develop
