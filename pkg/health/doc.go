// Package health provides health checks for Nix installations.
//
// This package implements various health checks to validate a Nix environment's
// configuration and ensure it meets recommended or required standards.
//
// # Overview
//
// The health package defines:
//   - Check interfaces and result types for health validation
//   - Individual health check implementations (flakes, version, caches, etc.)
//   - Aggregation and reporting of health check results
//
// # Usage
//
// Basic health check execution:
//
//	ctx := context.Background()
//	nixInfo, _ := nix.GetInfo(ctx)
//
//	healthChecks := health.Default()
//	results := healthChecks.RunAllChecks(ctx, nixInfo)
//
//	exitCode := health.EvaluateResults(results)
//	fmt.Println(exitCode.SummaryMessage())
//
// # Health Checks
//
// The package includes the following checks:
//   - FlakeEnabled: Verifies that Nix flakes are enabled
//   - NixVersion: Validates the Nix version meets minimum requirements
//   - Caches: Checks that required binary caches are configured
//   - TrustedUsers: Validates trusted user configuration
//   - Rosetta: Checks for Rosetta 2 on Apple Silicon Macs
//   - Direnv: Verifies direnv installation
//   - Homebrew: Checks for Homebrew on macOS
//   - Shell: Validates shell configuration
//
// # Check Results
//
// Each check returns a result that is either Green (passed) or Red (failed).
// Failed checks include a message describing the problem and a suggestion
// for how to fix it.
//
// # Architecture
//
// The package is organized into:
//   - pkg/health/checks: Individual check implementations and types
//   - pkg/health: Aggregation, evaluation, and reporting
//
// Check types are defined in the checks subpackage to avoid import cycles.
package health
