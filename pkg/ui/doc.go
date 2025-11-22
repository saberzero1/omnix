// Package ui provides shared UI rendering components for omnix.
//
// This package offers reusable UI primitives for both terminal user interfaces (TUI)
// and command-line interfaces (CLI) in omnix. It centralizes UI logic and ensures
// a consistent look and feel across commands and tools.
//
// # Available Components
//
//   - Health Rendering: Functions for displaying health check results with color-coded icons (✓/✗)
//   - Spinners: Animated indicators for ongoing operations, suitable for both TUI and CLI contexts
//   - Progress Bars: Visual progress indicators for long-running tasks with percentage tracking
//
// # When to Use This Package
//
// Use pkg/ui when you need:
//   - Consistent, reusable UI elements across omnix commands
//   - UI components that work in both TUI and CLI environments
//   - Styled UI elements with color support and animations
//
// Use pkg/common for:
//   - Low-level progress indicators or logging utilities
//   - Scenarios where minimal UI formatting is required
//
// # Usage Examples
//
// Rendering health check results:
//
//	results := []health.Check{...}
//	output := ui.RenderHealthChecks(results, 80)
//	fmt.Println(output)
//
// Using a spinner for long operations:
//
//	err := ui.RunWithSpinner("Checking environment...", func() error {
//	    return performCheck()
//	})
//
// Creating a progress bar:
//
//	progress := ui.NewProgress("Building packages", 32)
//	// Update with: progress.Update(ui.ProgressMsg{Current: 6})
package ui
