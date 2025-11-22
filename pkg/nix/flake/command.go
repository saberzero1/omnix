package flake

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// OutPath represents a path built by nix, as returned by --print-out-paths.
type OutPath struct {
	// DrvPath is the derivation that built these outputs
	DrvPath string `json:"drvPath"`
	// Outputs are the build outputs
	Outputs map[string]string `json:"outputs"`
}

// FirstOutput returns the first build output, if any.
func (o *OutPath) FirstOutput() *string {
	for _, path := range o.Outputs {
		return &path
	}
	return nil
}

// CommandOptions represents options for flake commands.
type CommandOptions struct {
	// OverrideInputs maps input names to flake URLs to override
	OverrideInputs map[string]string
	// NoWriteLockFile passes --no-write-lock-file
	NoWriteLockFile bool
	// CurrentDir is the directory from which to run the command
	CurrentDir string
}

// applyOptions applies command options to the args slice.
func applyOptions(args []string, opts *CommandOptions) []string {
	if opts == nil {
		return args
	}

	for name, url := range opts.OverrideInputs {
		args = append(args, "--override-input", name, url)
	}

	if opts.NoWriteLockFile {
		args = append(args, "--no-write-lock-file")
	}

	return args
}

// Run executes `nix run` on the given flake app.
func Run(ctx context.Context, cmd Cmd, opts *CommandOptions, url string, appArgs []string) error {
	args := []string{"run"}
	args = applyOptions(args, opts)
	args = append(args, url, "--")
	args = append(args, appArgs...)

	_, err := cmd.Run(ctx, args...)
	return err
}

// Develop executes `nix develop` on the given flake devshell.
func Develop(ctx context.Context, cmd Cmd, opts *CommandOptions, url string, command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command cannot be empty")
	}

	args := []string{"develop"}
	args = applyOptions(args, opts)
	args = append(args, url, "-c")
	args = append(args, command...)

	_, err := cmd.Run(ctx, args...)
	return err
}

// Build executes `nix build` and returns the output paths.
func Build(ctx context.Context, cmd Cmd, opts *CommandOptions, url string) ([]OutPath, error) {
	args := []string{"build", "--no-link", "--json"}
	args = applyOptions(args, opts)
	args = append(args, url)

	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, err
	}

	var outPaths []OutPath
	if err := json.Unmarshal([]byte(output), &outPaths); err != nil {
		return nil, fmt.Errorf("failed to parse build output: %w", err)
	}

	return outPaths, nil
}

// FlakeLock executes `nix flake lock` with additional options.
// Use this for advanced lock operations. For simple locking, use the Lock function from metadata.go.
func FlakeLock(ctx context.Context, cmd Cmd, opts *CommandOptions, url string, extraArgs []string) error {
	args := []string{"flake", "lock", url}
	args = applyOptions(args, opts)
	args = append(args, extraArgs...)

	_, err := cmd.Run(ctx, args...)
	return err
}

// Check executes `nix flake check`.
func Check(ctx context.Context, cmd Cmd, opts *CommandOptions, url string) error {
	args := []string{"flake", "check", url}
	args = applyOptions(args, opts)

	_, err := cmd.Run(ctx, args...)
	return err
}

// Show executes `nix flake show` and returns the output.
func Show(ctx context.Context, cmd Cmd, opts *CommandOptions, url string) (string, error) {
	args := []string{"flake", "show"}
	args = applyOptions(args, opts)
	args = append(args, url)

	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}
