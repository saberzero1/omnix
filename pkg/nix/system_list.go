package nix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saberzero1/omnix/pkg/nix/flake"
)

// SystemsList represents a list of Nix systems.
type SystemsList struct {
	Systems []flake.System
}

// SystemsListFlakeRef represents a flake referencing a SystemsList.
type SystemsListFlakeRef struct {
	URL FlakeURL
}

// Well-known nix-systems flake URLs
// See: https://github.com/nix-systems
var nixSystemsMap = map[string]string{
	"aarch64-linux":  "github:nix-systems/aarch64-linux",
	"x86_64-linux":   "github:nix-systems/x86_64-linux",
	"x86_64-darwin":  "github:nix-systems/x86_64-darwin",
	"aarch64-darwin": "github:nix-systems/aarch64-darwin",
	"default":        "github:nix-systems/default",
	"default-linux":  "github:nix-systems/default-linux",
	"default-darwin": "github:nix-systems/default-darwin",
}

// ParseSystemsListFlakeRef parses a system or flake URL into a SystemsListFlakeRef.
func ParseSystemsListFlakeRef(s string) SystemsListFlakeRef {
	sys := flake.ParseSystem(s)
	
	// Check if there's a known flake for this system
	if knownURL, ok := nixSystemsMap[sys.String()]; ok {
		return SystemsListFlakeRef{URL: NewFlakeURL(knownURL)}
	}
	
	// Otherwise use the input as a flake URL
	return SystemsListFlakeRef{URL: NewFlakeURL(s)}
}

// FromKnownSystem creates a SystemsListFlakeRef from a known system.
// Returns nil if the system is not known.
func FromKnownSystem(sys flake.System) *SystemsListFlakeRef {
	if knownURL, ok := nixSystemsMap[sys.String()]; ok {
		return &SystemsListFlakeRef{URL: NewFlakeURL(knownURL)}
	}
	return nil
}

// LoadSystemsList loads the list of systems from a SystemsListFlakeRef.
func LoadSystemsList(ctx context.Context, cmd *Cmd, ref SystemsListFlakeRef) (*SystemsList, error) {
	// First check if this is a known flake we can handle without network
	if systems := systemsListFromKnownFlake(ref); systems != nil {
		return systems, nil
	}
	
	// Otherwise, evaluate the flake
	return loadSystemsListFromRemoteFlake(ctx, cmd, ref)
}

// systemsListFromKnownFlake returns a SystemsList for known nix-systems flakes
// without requiring network access.
func systemsListFromKnownFlake(ref SystemsListFlakeRef) *SystemsList {
	url := ref.URL.String()
	
	// Map known URLs to their corresponding systems
	switch url {
	case "github:nix-systems/aarch64-linux":
		return &SystemsList{Systems: []flake.System{flake.SystemLinuxAarch64}}
	case "github:nix-systems/x86_64-linux":
		return &SystemsList{Systems: []flake.System{flake.SystemLinuxX86_64}}
	case "github:nix-systems/x86_64-darwin":
		return &SystemsList{Systems: []flake.System{flake.SystemDarwinX86_64}}
	case "github:nix-systems/aarch64-darwin":
		return &SystemsList{Systems: []flake.System{flake.SystemDarwinAarch64}}
	case "github:nix-systems/default":
		return &SystemsList{Systems: []flake.System{
			flake.SystemLinuxX86_64,
			flake.SystemLinuxAarch64,
			flake.SystemDarwinX86_64,
			flake.SystemDarwinAarch64,
		}}
	case "github:nix-systems/default-linux":
		return &SystemsList{Systems: []flake.System{
			flake.SystemLinuxX86_64,
			flake.SystemLinuxAarch64,
		}}
	case "github:nix-systems/default-darwin":
		return &SystemsList{Systems: []flake.System{
			flake.SystemDarwinX86_64,
			flake.SystemDarwinAarch64,
		}}
	default:
		return nil
	}
}

// loadSystemsListFromRemoteFlake loads a SystemsList by evaluating a flake.
func loadSystemsListFromRemoteFlake(ctx context.Context, cmd *Cmd, ref SystemsListFlakeRef) (*SystemsList, error) {
	// First get the flake path
	flakePath, err := nixEvalImpureExpr(ctx, cmd, fmt.Sprintf(`builtins.getFlake "%s"`, ref.URL.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to get flake: %w", err)
	}
	
	// Then import and evaluate it
	systemsJSON, err := nixEvalImpureExpr(ctx, cmd, fmt.Sprintf("import %s", flakePath))
	if err != nil {
		return nil, fmt.Errorf("failed to import flake: %w", err)
	}
	
	// Parse the JSON result
	var systemStrings []string
	if err := json.Unmarshal([]byte(systemsJSON), &systemStrings); err != nil {
		return nil, fmt.Errorf("failed to parse systems: %w", err)
	}
	
	// Convert to System objects
	systems := make([]flake.System, len(systemStrings))
	for i, s := range systemStrings {
		systems[i] = flake.ParseSystem(s)
	}
	
	return &SystemsList{Systems: systems}, nil
}

// nixEvalImpureExpr evaluates a Nix expression and returns the JSON result as a string.
func nixEvalImpureExpr(ctx context.Context, cmd *Cmd, expr string) (string, error) {
	args := []string{"eval", "--impure", "--json", "--expr", expr}
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return "", err
	}
	return output, nil
}
