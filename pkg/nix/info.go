package nix

import (
	"context"
	"fmt"
)

// Info represents all information about the user's Nix installation.
type Info struct {
	// Version is the Nix version
	Version Version
	// Env is the environment in which Nix operates
	Env *Env
}

// GetInfo gathers all Nix installation information.
func GetInfo(ctx context.Context) (*Info, error) {
	// Get Nix version
	cmd := NewCmd()
	version, err := cmd.RunVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Nix version: %w", err)
	}
	
	// Detect environment
	env, err := DetectEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}
	
	return &Info{
		Version: version,
		Env:     env,
	}, nil
}

// String returns a human-readable string representation of the Nix info.
func (i *Info) String() string {
	return fmt.Sprintf("Nix %s on %s", i.Version, i.Env.OS)
}
