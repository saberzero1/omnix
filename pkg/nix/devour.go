package nix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// DevourFlakeOutput represents the output of devour-flake
type DevourFlakeOutput struct {
	// OutPaths is the list of built store paths
	OutPaths []store.Path `json:"outPaths"`
	// ByName is a map of output names to store paths
	ByName map[string]store.Path `json:"byName"`
}

// DevourFlake builds all outputs of a flake using devour-flake
func DevourFlake(ctx context.Context, flake FlakeURL, systems []string, impure bool) (*DevourFlakeOutput, error) {
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

	// Add systems filtering if specified
	if len(systems) > 0 {
		systemsFlakeURL, err := GetSystemsFlakeURL(systems)
		if err != nil {
			return nil, fmt.Errorf("failed to get systems flake URL: %w", err)
		}
		args = append(args,
			"--override-input", "systems", systemsFlakeURL.String(),
		)
	}

	// Run nix build
	cmd := NewCmd()
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("devour-flake failed: %w", err)
	}

	// The output is a store path containing JSON
	// Read the JSON file
	storePath := strings.TrimSpace(output)
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

// GetSystemsFlakeURL converts a list of system strings to a nix-systems flake URL.
// If there's a single system, it uses the corresponding nix-systems flake.
// If there are multiple systems, it uses default or returns an error.
func GetSystemsFlakeURL(systems []string) (FlakeURL, error) {
	if len(systems) == 0 {
		return FlakeURL{}, fmt.Errorf("no systems specified")
	}

	// If single system, try to find a matching nix-systems flake
	if len(systems) == 1 {
		ref := ParseSystemsListFlakeRef(systems[0])
		return ref.URL, nil
	}

	// For multiple systems, use default if it matches the common patterns
	// This is a simplified approach - the Rust version may handle this differently
	return NewFlakeURL("github:nix-systems/default"), nil
}
