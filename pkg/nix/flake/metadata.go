package flake

import (
	"context"
	"encoding/json"
	"fmt"
)

// Metadata represents flake metadata information.
// See https://nix.dev/manual/nix/2.18/command-ref/new-cli/nix3-flake-metadata
type Metadata struct {
	// Description of the flake
	Description string `json:"description,omitempty"`
	
	// LastModified timestamp
	LastModified int64 `json:"lastModified,omitempty"`
	
	// Locked metadata
	Locked *LockedMetadata `json:"locked,omitempty"`
	
	// Original URL
	Original *FlakeRef `json:"original,omitempty"`
	
	// Resolved URL
	Resolved *FlakeRef `json:"resolved,omitempty"`
	
	// URL of the flake
	URL string `json:"url,omitempty"`
	
	// Path to the flake in the store
	Path string `json:"path,omitempty"`
	
	// Revision (git commit hash for git-based flakes)
	Revision string `json:"revision,omitempty"`
	
	// RevCount (number of commits for git-based flakes)
	RevCount int `json:"revCount,omitempty"`
}

// LockedMetadata contains locked flake reference information.
type LockedMetadata struct {
	// LastModified timestamp
	LastModified int64 `json:"lastModified,omitempty"`
	
	// NarHash of the flake
	NarHash string `json:"narHash,omitempty"`
	
	// Owner (for GitHub flakes)
	Owner string `json:"owner,omitempty"`
	
	// Repo (for GitHub flakes)
	Repo string `json:"repo,omitempty"`
	
	// Rev (git revision)
	Rev string `json:"rev,omitempty"`
	
	// Type of the flake reference
	Type string `json:"type,omitempty"`
}

// FlakeRef represents a flake reference (original or resolved).
type FlakeRef struct {
	// Owner (for GitHub flakes)
	Owner string `json:"owner,omitempty"`
	
	// Repo (for GitHub flakes)
	Repo string `json:"repo,omitempty"`
	
	// Type of the reference
	Type string `json:"type,omitempty"`
	
	// Dir (subdirectory in the flake)
	Dir string `json:"dir,omitempty"`
	
	// Ref (branch/tag name for git flakes)
	Ref string `json:"ref,omitempty"`
}

// GetMetadata retrieves metadata for a flake.
func GetMetadata(ctx context.Context, cmd Cmd, flakeURL string) (*Metadata, error) {
	args := []string{"flake", "metadata", "--json", flakeURL}
	
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get flake metadata: %w", err)
	}
	
	var metadata Metadata
	if err := json.Unmarshal([]byte(output), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}
	
	return &metadata, nil
}

// GetMetadataWithOptions retrieves metadata for a flake with custom options.
func GetMetadataWithOptions(ctx context.Context, cmd Cmd, flakeURL string, opts *FlakeOptions) (*Metadata, error) {
	args := []string{"flake", "metadata", "--json"}
	
	if opts != nil {
		if opts.Refresh {
			args = append(args, "--refresh")
		}
		for input, url := range opts.OverrideInputs {
			args = append(args, "--override-input", input, url)
		}
	}
	
	args = append(args, flakeURL)
	
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get flake metadata: %w", err)
	}
	
	var metadata Metadata
	if err := json.Unmarshal([]byte(output), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}
	
	return &metadata, nil
}

// MetadataInput represents input options for fetching flake metadata.
type MetadataInput struct {
	// Flake URL to query
	FlakeURL string
	
	// IncludeInputs transitively includes flake inputs in the result
	// NOTE: This makes evaluation more expensive
	IncludeInputs bool
}

// Lock represents a flake lock operation.
func Lock(ctx context.Context, cmd Cmd, flakeURL string, updateInputs []string) error {
	args := []string{"flake", "lock", flakeURL}
	
	for _, input := range updateInputs {
		args = append(args, "--update-input", input)
	}
	
	_, err := cmd.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to lock flake: %w", err)
	}
	
	return nil
}

// Update updates all inputs of a flake.
func Update(ctx context.Context, cmd Cmd, flakeURL string) error {
	args := []string{"flake", "update", flakeURL}
	
	_, err := cmd.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to update flake: %w", err)
	}
	
	return nil
}
