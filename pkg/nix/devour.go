package nix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/saberzero1/omnix/pkg/nix/store"
)

// DevourFlakeURL returns the URL of the devour-flake flake
// This is set by the Nix build environment
func DevourFlakeURL() string {
	// Try to get from environment variable (set by Nix during build)
	if url := os.Getenv("DEVOUR_FLAKE"); url != "" {
		return url
	}
	// Fallback to the GitHub repository
	return "github:srid/devour-flake/9fe4db872c107ea217c13b24527b68d9e4a4c01b"
}

// DevourFlakeInput represents the input to devour-flake
type DevourFlakeInput struct {
	// Flake is the flake URL to build all outputs for
	Flake FlakeURL `json:"flake"`
	// Systems is an optional list of systems to build for
	// An empty list means all allowed systems
	Systems *FlakeURL `json:"systems,omitempty"`
}

// DevourFlakeOutput represents the output of devour-flake
type DevourFlakeOutput struct {
	// OutPaths is the list of built store paths
	OutPaths []store.Path `json:"outPaths"`
	// ByName is a map of output names to store paths
	ByName map[string]store.Path `json:"byName"`
}

// DevourFlake builds all outputs of a flake using devour-flake
func DevourFlake(ctx context.Context, flake FlakeURL, systems []string, impure bool) (*DevourFlakeOutput, error) {
	// Prepare input
	input := DevourFlakeInput{
		Flake: flake,
	}

	// Convert systems to flake URL if provided
	if len(systems) > 0 {
		// devour-flake expects systems as a flake URL
		// We'll use the NIX_SYSTEMS environment variable approach
		// For now, we'll pass nil and let devour-flake use all systems
		// TODO: Implement proper systems filtering
		input.Systems = nil
	}

	// Build the devour-flake command
	devourURL := DevourFlakeURL() + "#json"

	args := []string{
		"build",
		devourURL,
		"-L",
		"--no-link",
		"--print-out-paths",
	}

	if impure {
		args = append(args, "--impure")
	}

	// Add override-input for the flake to build
	args = append(args,
		"--override-input", "flake", flake.String(),
	)

	// Run nix build
	cmd := NewCmd()
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("devour-flake failed: %w", err)
	}

	// The output is a store path containing JSON
	// Read the JSON file
	storePath := output
	if storePath == "" {
		return nil, fmt.Errorf("devour-flake returned empty output")
	}

	// Read the JSON output
	data, err := os.ReadFile(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read devour-flake output: %w", err)
	}

	// Parse the JSON
	var result DevourFlakeOutput
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse devour-flake output: %w", err)
	}

	// Remove duplicates
	result.OutPaths = uniquePaths(result.OutPaths)

	return &result, nil
}

// uniquePaths removes duplicate paths from the list
func uniquePaths(paths []store.Path) []store.Path {
	seen := make(map[string]bool)
	result := make([]store.Path, 0, len(paths))

	for _, path := range paths {
		pathStr := path.String()
		if !seen[pathStr] {
			seen[pathStr] = true
			result = append(result, path)
		}
	}

	return result
}
