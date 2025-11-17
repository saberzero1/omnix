// Package nix provides Go bindings for interacting with the Nix package manager.
//
// This package offers functionality for:
//   - Version detection and parsing
//   - Command execution with context support
//   - Flake URL handling and manipulation
//   - Environment detection (OS, user, groups)
//   - Installation information aggregation
//
// # Basic Usage
//
// Get Nix version:
//
//	cmd := nix.NewCmd()
//	version, err := cmd.RunVersion(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Nix version: %s\n", version)
//
// Execute a Nix command:
//
//	output, err := cmd.Run(context.Background(), "flake", "show", ".")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parse and manipulate flake URLs:
//
//	url := nix.NewFlakeURL(".")
//	withAttr := url.WithAttr("packages.x86_64-linux.default")
//	fmt.Println(withAttr) // .#packages.x86_64-linux.default
//
// Detect environment:
//
//	env, err := nix.DetectEnv(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("OS: %s, User: %s\n", env.OS, env.CurrentUser)
//
// Get complete Nix installation info:
//
//	info, err := nix.GetInfo(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(info) // Nix 2.13.0 on Linux
package nix
