package develop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/juspay/omnix/pkg/nix"
)

// Project represents a Nix project that can be developed locally
type Project struct {
	// Dir is the local directory of the project (nil for remote flakes)
	Dir *string
	// Flake is the flake URL
	Flake nix.FlakeURL
	// Config is the develop configuration
	Config Config
}

// NewProject creates a new Project instance
func NewProject(ctx context.Context, flake nix.FlakeURL, config Config) (*Project, error) {
	var dir *string

	// If it's a local path, canonicalize it
	localPath := flake.AsLocalPath()
	if localPath != "" {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		dir = &absPath
	}

	return &Project{
		Dir:    dir,
		Flake:  flake,
		Config: config,
	}, nil
}

// GetWorkingDir returns the working directory for the project
// Returns current directory if project is remote
func (p *Project) GetWorkingDir() (string, error) {
	if p.Dir != nil {
		return *p.Dir, nil
	}
	return os.Getwd()
}
