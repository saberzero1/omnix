package ci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/saberzero1/omnix/pkg/nix/flake/functions/metadata"
	"github.com/saberzero1/omnix/pkg/nix/store"
)

// RemoteRunOptions contains options for running CI remotely with flake caching
type RemoteRunOptions struct {
	// StoreURI is the remote store URI (e.g., "ssh://user@host")
	StoreURI *store.URI

	// CopyInputs determines whether to copy all flake inputs transitively
	CopyInputs bool

	// OutLink is the path for the symlink to results (if requested)
	OutLink string
}

// RunOnRemoteStore executes CI on a remote store with flake metadata caching.
//
// This function:
//  1. Caches the flake locally using metadata functions (with optional input inclusion)
//  2. Copies the flake closure and omnix source to the remote store
//  3. Executes CI on the remote using the cached flake via SSH
//  4. Optionally copies results back if out-link is requested
//
// This matches the Rust implementation in crates/omnix-ci/src/command/run_remote.rs
func RunOnRemoteStore(
	ctx context.Context,
	flake nix.FlakeURL,
	subflakesConfig map[string]SubflakeConfig,
	opts RunOptions,
	remoteOpts RemoteRunOptions,
) ([]Result, error) {
	if !remoteOpts.StoreURI.IsSSH() {
		return nil, fmt.Errorf("only SSH store URIs are supported")
	}

	sshURI := remoteOpts.StoreURI.GetSSHURI()
	if sshURI == nil {
		return nil, fmt.Errorf("failed to get SSH URI from store URI")
	}

	// Cache the flake locally with optional inputs
	flakeClosure, cachedFlakeURL, err := cacheFlake(ctx, flake, remoteOpts.CopyInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to cache flake: %w", err)
	}

	// Get omnix source path
	omnixSource, err := getOmnixSourcePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get omnix source: %w", err)
	}

	// Paths to copy to remote
	pathsToCopy := []string{omnixSource, flakeClosure}

	// Copy flake and omnix source to remote store
	if err := copyToRemote(ctx, remoteOpts.StoreURI, pathsToCopy); err != nil {
		return nil, fmt.Errorf("failed to copy to remote: %w", err)
	}

	// Handle out-link if requested
	if remoteOpts.OutLink != "" {
		return runRemoteWithOutLink(ctx, sshURI, cachedFlakeURL, subflakesConfig, opts, omnixSource, remoteOpts)
	}

	// Run CI on remote without out-link
	return runRemoteWithoutOutLink(ctx, sshURI, cachedFlakeURL, subflakesConfig, opts, omnixSource)
}

// cacheFlake caches a flake locally using metadata functions and returns the
// closure path and the cached flake URL.
func cacheFlake(ctx context.Context, flake nix.FlakeURL, includeInputs bool) (string, nix.FlakeURL, error) {
	input := metadata.Input{
		Flake:         flake.String(),
		IncludeInputs: includeInputs,
	}

	_, output, err := metadata.GetMetadata(ctx, input)
	if err != nil {
		return "", nix.FlakeURL{}, fmt.Errorf("failed to get flake metadata: %w", err)
	}

	// The flake itself is in the metadata output
	cachedFlakeURL, err := nix.ParseFlakeURL(output.Flake)
	if err != nil {
		return "", nix.FlakeURL{}, fmt.Errorf("failed to parse cached flake URL: %w", err)
	}

	// Restore the attribute from the original flake URL if any
	if attr := flake.GetAttr(); !attr.IsNone() {
		cachedFlakeURL = cachedFlakeURL.WithAttr(attr.String())
	}

	return output.Flake, cachedFlakeURL, nil
}

// omnixSourcePath is injected at build time by Nix via -ldflags -X
// See nix/modules/flake/go.nix for the injection
var omnixSourcePath string

// getOmnixSourcePath returns the path to the omnix source.
// This is set during build time via ldflags.
func getOmnixSourcePath() (string, error) {
	// Check if injected at build time
	if omnixSourcePath != "" {
		return omnixSourcePath, nil
	}

	// Check if OMNIX_SOURCE env var is set (alternative for development)
	if omnixSource := os.Getenv("OMNIX_SOURCE"); omnixSource != "" {
		return omnixSource, nil
	}

	// Fallback: try to find omnix binary in PATH and use its store path
	omnixPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get omnix executable path: %w", err)
	}

	// If the binary is in /nix/store, use its store path
	if strings.HasPrefix(omnixPath, "/nix/store/") {
		// Extract store path (everything up to the first / after /nix/store/hash-name)
		rel := omnixPath[len("/nix/store/"):]
		idx := strings.Index(rel, "/")
		if idx != -1 {
			storePath := "/nix/store/" + rel[:idx]
			return storePath, nil
		}
		// If there is no further slash, the binary is directly in the store path
		return omnixPath, nil
	}

	return "", fmt.Errorf("OMNIX_SOURCE not set and omnix not in /nix/store")
}

// copyToRemote copies paths to a remote Nix store
func copyToRemote(ctx context.Context, storeURI *store.URI, paths []string) error {
	cmd := nix.NewCmd()
	opts := nix.CopyOptions{
		To:          storeURI,
		NoCheckSigs: true,
	}

	return nix.Copy(ctx, cmd, opts, paths)
}

// copyFromRemote copies paths from a remote Nix store
func copyFromRemote(ctx context.Context, storeURI *store.URI, paths []string) error {
	cmd := nix.NewCmd()
	opts := nix.CopyOptions{
		From:        storeURI,
		NoCheckSigs: true,
	}

	return nix.Copy(ctx, cmd, opts, paths)
}

// runRemoteWithoutOutLink runs CI on remote without creating a local out-link
func runRemoteWithoutOutLink(
	ctx context.Context,
	sshURI *store.SSHURI,
	cachedFlake nix.FlakeURL,
	subflakesConfig map[string]SubflakeConfig,
	opts RunOptions,
	omnixSource string,
) ([]Result, error) {
	// Build the remote om ci run command
	args := buildRemoteCICommand(omnixSource, cachedFlake, subflakesConfig, opts, "")

	// Execute via SSH
	output, err := executeRemoteCommand(ctx, sshURI.String(), args)
	if err != nil {
		return nil, fmt.Errorf("remote CI failed: %w\nOutput: %s", err, output)
	}

	// Since we don't have out-link, we can't retrieve detailed results
	// Return a basic success result
	return []Result{
		{
			Subflake: "remote",
			Success:  true,
			Steps:    make(map[string]StepResult),
		},
	}, nil
}

// runRemoteWithOutLink runs CI on remote and copies results back for local out-link.
//
// Note: Currently returns a generic success result rather than parsing the JSON.
// The detailed results are available in the out-link file on disk.
func runRemoteWithOutLink(
	ctx context.Context,
	sshURI *store.SSHURI,
	cachedFlake nix.FlakeURL,
	subflakesConfig map[string]SubflakeConfig,
	opts RunOptions,
	omnixSource string,
	remoteOpts RemoteRunOptions,
) ([]Result, error) {
	// Create a temporary directory on the remote to hold results
	tmpDirCmd := []string{"mktemp", "-d", "-t", "om.json.XXXXXX"}
	tmpDirOutput, err := runNixpkgsCommand(ctx, sshURI.String(), "coreutils", tmpDirCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote temp dir: %w", err)
	}

	tmpDir := strings.TrimSpace(tmpDirOutput)
	omJSONPath := filepath.Join(tmpDir, "om.json")

	// Build the remote om ci run command with out-link
	args := buildRemoteCICommand(omnixSource, cachedFlake, subflakesConfig, opts, omJSONPath)

	// Execute via SSH
	output, err := executeRemoteCommand(ctx, sshURI.String(), args)
	if err != nil {
		return nil, fmt.Errorf("remote CI failed: %w\nOutput: %s", err, output)
	}

	// Get the out-link store path
	readlinkCmd := []string{"readlink", omJSONPath}
	resultPathOutput, err := runNixpkgsCommand(ctx, sshURI.String(), "coreutils", readlinkCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote out-link: %w", err)
	}

	resultPath := strings.TrimSpace(resultPathOutput)

	// Copy results back to local store
	if err := copyFromRemote(ctx, remoteOpts.StoreURI, []string{resultPath}); err != nil {
		return nil, fmt.Errorf("failed to copy results from remote: %w", err)
	}

	// Create local GC root
	storeCmd := store.NewCmd()
	resultStorePath := store.NewPath(resultPath)
	if err := storeCmd.AddRoot(ctx, remoteOpts.OutLink, []store.Path{resultStorePath}); err != nil {
		return nil, fmt.Errorf("failed to create local out-link: %w", err)
	}

	// Since we have the results, return a basic success
	// TODO: Parse the JSON results file to return actual results
	return []Result{
		{
			Subflake: "remote",
			Success:  true,
			Steps:    make(map[string]StepResult),
		},
	}, nil
}

// buildRemoteCICommand builds the om ci run command to execute on remote
func buildRemoteCICommand(
	omnixSource string,
	cachedFlake nix.FlakeURL,
	subflakesConfig map[string]SubflakeConfig,
	opts RunOptions,
	outLink string,
) []string {
	// Use nix run to execute omnix from source
	omnixFlake := omnixSource + "#default"
	args := []string{
		"nix",
		"--accept-flake-config",
		"run",
		omnixFlake,
		"--",
		"ci",
		"run",
		cachedFlake.String(),
	}

	// Add systems if specified
	if len(opts.Systems) > 0 {
		args = append(args, "--systems", strings.Join(opts.Systems, ","))
	}

	// Add out-link if specified
	if outLink != "" {
		args = append(args, "--out-link", outLink)
	} else {
		args = append(args, "--no-link")
	}

	// Add other options
	if opts.GitHubOutput {
		args = append(args, "--github-output")
	}

	if opts.IncludeAllDependencies {
		args = append(args, "--include-all-dependencies")
	}

	if opts.Parallel {
		args = append(args, "--parallel")
	}

	if opts.MaxConcurrency > 0 {
		args = append(args, "--max-concurrency", fmt.Sprintf("%d", opts.MaxConcurrency))
	}

	// Disable UI for remote execution
	args = append(args, "--no-ui")

	return args
}

// runNixpkgsCommand runs a command from nixpkgs on the remote host
func runNixpkgsCommand(ctx context.Context, host string, pkg string, cmd []string) (string, error) {
	args := []string{
		"nix",
		"shell",
		fmt.Sprintf("nixpkgs#%s", pkg),
		"-c",
	}
	args = append(args, cmd...)

	return executeRemoteCommand(ctx, host, args)
}
