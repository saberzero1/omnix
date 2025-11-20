package store

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StoreCmd represents the nix-store command.
type StoreCmd struct{}

// NewStoreCmd creates a new StoreCmd instance.
func NewStoreCmd() *StoreCmd {
	return &StoreCmd{}
}

// runNixStore executes nix-store with the given arguments.
func runNixStore(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "nix-store", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("nix-store failed: %s: %w", string(exitErr.Stderr), err)
		}
		return "", err
	}
	return string(output), nil
}

// QueryDeriver returns the derivations used to build the given build outputs.
func (s *StoreCmd) QueryDeriver(ctx context.Context, paths []Path) ([]string, error) {
	args := []string{"--query", "--valid-derivers"}
	for _, p := range paths {
		args = append(args, p.String())
	}
	
	output, err := runNixStore(ctx, args...)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Filter out "unknown-deriver"
	var derivers []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "unknown-deriver" {
			derivers = append(derivers, line)
		}
	}
	
	if len(derivers) == 0 && len(lines) > 0 {
		// All derivers were unknown
		return nil, fmt.Errorf("unknown deriver for provided paths")
	}
	
	return derivers, nil
}

// QueryRequisites recursively queries and returns all dependencies of the given derivation paths.
func (s *StoreCmd) QueryRequisites(ctx context.Context, drvPaths []string, includeOutputs bool) ([]Path, error) {
	args := []string{"--query", "--requisites"}
	if includeOutputs {
		args = append(args, "--include-outputs")
	}
	args = append(args, drvPaths...)
	
	output, err := runNixStore(ctx, args...)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var paths []Path
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, NewPath(line))
		}
	}
	
	return paths, nil
}

// FetchAllDeps fetches all build and runtime dependencies of given derivation outputs.
func (s *StoreCmd) FetchAllDeps(ctx context.Context, outPaths []Path) ([]Path, error) {
	// First, get the derivers
	derivers, err := s.QueryDeriver(ctx, outPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to query derivers: %w", err)
	}
	
	if len(derivers) == 0 {
		return []Path{}, nil
	}
	
	// Then get all requisites with outputs
	allDeps, err := s.QueryRequisites(ctx, derivers, true)
	if err != nil {
		return nil, fmt.Errorf("failed to query requisites: %w", err)
	}
	
	return allDeps, nil
}

// Add adds a path to the Nix store and returns the store path.
func (s *StoreCmd) Add(ctx context.Context, path string) (Path, error) {
	output, err := runNixStore(ctx, "--add", path)
	if err != nil {
		return Path{}, err
	}
	
	storePath := strings.TrimSpace(output)
	return NewPath(storePath), nil
}

// AddRoot creates a GC root for the given store paths at the specified symlink location.
func (s *StoreCmd) AddRoot(ctx context.Context, symlink string, paths []Path) error {
	for _, path := range paths {
		if _, err := runNixStore(ctx, "--add-root", symlink, "--realise", path.String()); err != nil {
			return fmt.Errorf("failed to add root for %s: %w", path, err)
		}
	}
	
	return nil
}

// AddFilePermanently creates a file in the Nix store such that it escapes garbage collection.
// Returns the store path added.
func (s *StoreCmd) AddFilePermanently(ctx context.Context, symlinkPath string, contents string) (Path, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "omnix-*")
	if err != nil {
		return Path{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Best effort cleanup
	}()
	
	// Write contents to a temporary file
	tempFile := filepath.Join(tempDir, "file")
	if err := os.WriteFile(tempFile, []byte(contents), 0644); err != nil {
		return Path{}, fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Add to store
	storePath, err := s.Add(ctx, tempFile)
	if err != nil {
		return Path{}, fmt.Errorf("failed to add to store: %w", err)
	}
	
	// Create GC root
	if err := s.AddRoot(ctx, symlinkPath, []Path{storePath}); err != nil {
		return Path{}, fmt.Errorf("failed to add root: %w", err)
	}
	
	return storePath, nil
}
