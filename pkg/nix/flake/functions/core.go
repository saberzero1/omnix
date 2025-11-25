package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/saberzero1/omnix/pkg/nix"
)

// Build-time injectable variables.
// These are set via ldflags during Nix build to embed absolute paths to flake functions.
// Example: -X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeMetadata=path:/nix/store/...
var (
	// flakeMetadata is the path to the metadata flake function.
	// Injected via: -X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeMetadata=...
	flakeMetadata string

	// flakeAddStringContext is the path to the addstringcontext flake function.
	// Injected via: -X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeAddStringContext=...
	flakeAddStringContext string
)

var (
	// repoRoot is the root directory of the repository, determined at init time
	repoRoot     string
	repoRootOnce sync.Once
)

// getRepoRoot returns the repository root directory.
// It's determined once and cached for subsequent calls.
//
// The function walks up the directory tree from the current working directory
// looking for go.mod or flake.nix as indicators of the repository root.
// If these markers are not found, it falls back to the current working directory.
//
// Note: This fallback may not be correct in all scenarios (e.g., tests or when
// running from outside the repository). In such cases, the FLAKE_METADATA and
// FLAKE_ADDSTRINGCONTEXT environment variables should be set explicitly.
func getRepoRoot() string {
	repoRootOnce.Do(func() {
		// Try to get from current working directory first
		if cwd, err := os.Getwd(); err == nil {
			// Walk up the directory tree looking for go.mod or flake.nix
			dir := cwd
			for {
				// Check for go.mod or flake.nix as indicators of repo root
				if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
					repoRoot = dir
					return
				}
				if _, err := os.Stat(filepath.Join(dir, "flake.nix")); err == nil {
					repoRoot = dir
					return
				}

				// Move up one directory
				parent := filepath.Dir(dir)
				if parent == dir {
					// Reached root without finding markers
					break
				}
				dir = parent
			}
		}

		// Fallback to current working directory
		// This may not be correct in all scenarios but is better than failing
		if cwd, err := os.Getwd(); err == nil {
			repoRoot = cwd
		} else {
			repoRoot = "."
		}
	})
	return repoRoot
}

// FlakeFn defines the interface for Nix flake functions.
// This interface allows calling Nix flakes as functions with typed inputs and outputs.
type FlakeFn[Input, Output any] interface {
	// FlakeURL returns the flake URL referencing this function
	FlakeURL() string

	// Init is called after reading from Nix build to initialize/modify the output
	Init(out *Output)
}

// CallOptions configures how a flake function is called
type CallOptions struct {
	// Impure enables --impure flag for nix build
	Impure bool

	// WorkDir sets the working directory for the nix build command
	WorkDir string

	// OutLink specifies the output link path. If empty, --no-link is used
	OutLink string

	// ExtraArgs are additional arguments to pass to nix build
	// Note: --override-input arguments are treated specially (see transformOverrideInputs)
	ExtraArgs []string
}

// Call executes a flake function with the given input and returns the output.
//
// The function:
//  1. Serializes the input struct to override-input arguments
//  2. Runs `nix build` with the flake function URL
//  3. Reads the JSON output from the built store path
//  4. Deserializes the JSON into the output struct
//  5. Calls Init on the output
//
// Arguments with boolean values are converted to TRUE_FLAKE or FALSE_FLAKE URLs.
// Override-input arguments in ExtraArgs are prefixed with "flake/" to account for
// nested flake inputs.
func Call[Input, Output any](
	ctx context.Context,
	fn FlakeFn[Input, Output],
	opts *CallOptions,
	input Input,
) (storePath string, output Output, err error) {
	if opts == nil {
		opts = &CallOptions{}
	}

	// Build nix build command
	args := []string{"build", fn.FlakeURL(), "-L", "--print-out-paths"}

	if opts.Impure {
		args = append(args, "--impure")
	}

	if opts.OutLink != "" {
		args = append(args, "--out-link", opts.OutLink)
	} else {
		args = append(args, "--no-link")
	}

	// Convert input struct to override-input arguments
	overrideInputs, err := toOverrideInputs(input)
	if err != nil {
		return "", output, fmt.Errorf("failed to convert input to override-inputs: %w", err)
	}

	for k, v := range overrideInputs {
		args = append(args, "--override-input", k, v)
	}

	// Transform and add extra args (with special handling for --override-input)
	args = append(args, transformOverrideInputs(opts.ExtraArgs)...)

	// Create command
	cmd := nix.NewCmd()

	// Run nix build
	// Note: os.Chdir affects the entire process and is NOT goroutine-safe.
	// This implementation should not be called concurrently with different WorkDir values.
	// We restore the original directory in a defer to minimize the window of side effects.
	// Since nix.Cmd doesn't support setting the working directory directly,
	// we must use os.Chdir for WorkDir support.
	if opts.WorkDir != "" {
		originalDir, err := os.Getwd()
		if err != nil {
			return "", output, fmt.Errorf("failed to get current directory: %w", err)
		}
		if err := os.Chdir(opts.WorkDir); err != nil {
			return "", output, fmt.Errorf("failed to change directory: %w", err)
		}
		defer func() {
			_ = os.Chdir(originalDir)
		}()
	}

	cmdOutput, err := cmd.Run(ctx, args...)
	if err != nil {
		return "", output, fmt.Errorf("nix build failed: %w", err)
	}

	// Parse store path from output
	storePath = strings.TrimSpace(cmdOutput)
	if storePath == "" {
		return "", output, fmt.Errorf("nix build returned empty output")
	}

	// Read JSON from store path
	data, err := os.ReadFile(storePath)
	if err != nil {
		return "", output, fmt.Errorf("failed to read output from %s: %w", storePath, err)
	}

	// Unmarshal JSON into output
	if err := json.Unmarshal(data, &output); err != nil {
		return "", output, fmt.Errorf("failed to unmarshal output JSON: %w", err)
	}

	// Initialize output
	fn.Init(&output)

	return storePath, output, nil
}

// toOverrideInputs converts a struct to a map of override-input arguments.
// Fields are serialized to JSON and then converted:
//   - String values are used as-is
//   - Boolean true becomes TrueFlakeURL()
//   - Boolean false becomes FalseFlakeURL()
//   - Null/nil values are skipped
func toOverrideInputs(input any) (map[string]string, error) {
	// Serialize to JSON to get field names and values
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	result := make(map[string]string)
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			result[k] = val
		case bool:
			if val {
				result[k] = TrueFlakeURL()
			} else {
				result[k] = FalseFlakeURL()
			}
		case nil:
			// Skip null values
			continue
		default:
			// For other types, we skip them
			// The original Rust implementation only handles String and Bool
			continue
		}
	}

	return result, nil
}

// transformOverrideInputs transforms --override-input arguments to use "flake/" prefix.
//
// This is necessary because when calling a flake function that itself takes a flake as input,
// we need to distinguish between override inputs for the function's flake vs the nested flake.
//
// For example, if the function flake has an input named "flake" (common pattern), and we want
// to override an input of that nested flake, we use "flake/input-name" as the override key.
func transformOverrideInputs(args []string) []string {
	result := make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		result = append(result, args[i])
		if args[i] == "--override-input" && i+1 < len(args) {
			i++
			// Prefix the input name with "flake/"
			result = append(result, "flake/"+args[i])
		}
		i++
	}
	return result
}

// TrueFlakeURL returns the URL to the TRUE_FLAKE
// This is set by the Nix build environment
func TrueFlakeURL() string {
	if url := os.Getenv("TRUE_FLAKE"); url != "" {
		return url
	}
	// Fallback to stable commit from the boolean-option GitHub organization.
	// This commit SHA must match the one in nix/envs/default.nix (TRUE_FLAKE, line 55).
	// These are minimal flakes maintained specifically for boolean inputs in flake functions.
	// Pinned to a specific commit to ensure reproducibility.
	return "github:boolean-option/true/6ecb49143ca31b140a5273f1575746ba93c3f698"
}

// FalseFlakeURL returns the URL to the FALSE_FLAKE
// This is set by the Nix build environment
func FalseFlakeURL() string {
	if url := os.Getenv("FALSE_FLAKE"); url != "" {
		return url
	}
	// Fallback to stable commit from the boolean-option GitHub organization.
	// This commit SHA must match the one in nix/envs/default.nix (FALSE_FLAKE, line 49).
	// These are minimal flakes maintained specifically for boolean inputs in flake functions.
	// Pinned to a specific commit to ensure reproducibility.
	return "github:boolean-option/false/d06b4794a134686c70a1325df88a6e6768c6b212"
}

// FlakeMetadataURL returns the URL to the FLAKE_METADATA flake function.
//
// Priority order:
// 1. Build-time injected value (flakeMetadata variable, set via ldflags during Nix build)
// 2. Runtime environment variable (FLAKE_METADATA)
// 3. Fallback: derive from repository root (only works when running from within the repo)
func FlakeMetadataURL() string {
	// First, check build-time injected value (most reliable for Nix builds)
	if flakeMetadata != "" {
		return flakeMetadata + "#default"
	}
	// Then check runtime environment variable
	if url := os.Getenv("FLAKE_METADATA"); url != "" {
		return url + "#default"
	}
	// Fallback: use absolute path to avoid issues with WorkDir changes
	root := getRepoRoot()
	return fmt.Sprintf("path:%s/pkg/nix/flake/functions/metadata#default", root)
}

// FlakeAddStringContextURL returns the URL to the FLAKE_ADDSTRINGCONTEXT flake function.
//
// Priority order:
// 1. Build-time injected value (flakeAddStringContext variable, set via ldflags during Nix build)
// 2. Runtime environment variable (FLAKE_ADDSTRINGCONTEXT)
// 3. Fallback: derive from repository root (only works when running from within the repo)
func FlakeAddStringContextURL() string {
	// First, check build-time injected value (most reliable for Nix builds)
	if flakeAddStringContext != "" {
		return flakeAddStringContext + "#default"
	}
	// Then check runtime environment variable
	if url := os.Getenv("FLAKE_ADDSTRINGCONTEXT"); url != "" {
		return url + "#default"
	}
	// Fallback: use absolute path to avoid issues with WorkDir changes
	root := getRepoRoot()
	return fmt.Sprintf("path:%s/pkg/nix/flake/functions/addstringcontext#default", root)
}
