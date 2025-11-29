package nix

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix/flake"
	"github.com/saberzero1/omnix/pkg/nix/store"
	"go.uber.org/zap"
)

// OutputCategory represents a category of flake outputs
type OutputCategory string

const (
	// OutputCategoryPackages represents packages output
	OutputCategoryPackages OutputCategory = "packages"
	// OutputCategoryChecks represents checks output
	OutputCategoryChecks OutputCategory = "checks"
	// OutputCategoryDevShells represents devShells output
	OutputCategoryDevShells OutputCategory = "devShells"
	// OutputCategoryApps represents apps output
	OutputCategoryApps OutputCategory = "apps"
	// OutputCategoryLegacyPackages represents legacyPackages output
	OutputCategoryLegacyPackages OutputCategory = "legacyPackages"
	// OutputCategoryNixosConfigurations represents nixosConfigurations output
	OutputCategoryNixosConfigurations OutputCategory = "nixosConfigurations"
	// OutputCategoryDarwinConfigurations represents darwinConfigurations output
	OutputCategoryDarwinConfigurations OutputCategory = "darwinConfigurations"
)

// AllPerSystemCategories returns all per-system output categories
func AllPerSystemCategories() []OutputCategory {
	return []OutputCategory{
		OutputCategoryPackages,
		OutputCategoryChecks,
		OutputCategoryDevShells,
		OutputCategoryApps,
		OutputCategoryLegacyPackages,
	}
}

// AllFlakeLevelCategories returns all flake-level output categories (not per-system)
func AllFlakeLevelCategories() []OutputCategory {
	return []OutputCategory{
		OutputCategoryNixosConfigurations,
		OutputCategoryDarwinConfigurations,
	}
}

// FlakeShowOutput represents the output of `nix flake show --json`
type FlakeShowOutput struct {
	Packages             map[string]map[string]FlakeShowVal `json:"packages,omitempty"`
	Checks               map[string]map[string]FlakeShowVal `json:"checks,omitempty"`
	DevShells            map[string]map[string]FlakeShowVal `json:"devShells,omitempty"`
	Apps                 map[string]map[string]FlakeShowVal `json:"apps,omitempty"`
	LegacyPackages       map[string]map[string]interface{}  `json:"legacyPackages,omitempty"`
	NixosConfigurations  map[string]FlakeShowVal            `json:"nixosConfigurations,omitempty"`
	DarwinConfigurations map[string]FlakeShowVal            `json:"darwinConfigurations,omitempty"`
}

// FlakeShowVal represents a terminal value in `nix flake show` output
type FlakeShowVal struct {
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// FlakeOutput represents a single buildable flake output
type FlakeOutput struct {
	// Category is the output category (packages, checks, devShells, etc.)
	Category OutputCategory
	// System is the system for per-system outputs (empty for flake-level outputs)
	System string
	// Name is the attribute name
	Name string
	// FlakeRef is the full flake reference to build (e.g., ".#packages.x86_64-linux.default")
	FlakeRef string
}

// EnumerateFlakeOutputs enumerates all buildable outputs of a flake
func EnumerateFlakeOutputs(ctx context.Context, flakeURL FlakeURL, systems []string) ([]FlakeOutput, error) {
	cmd := NewCmd()

	// Get flake show output
	args := []string{"flake", "show", "--json", flakeURL.String()}
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to show flake: %w", err)
	}

	var showOutput FlakeShowOutput
	if err := json.Unmarshal([]byte(output), &showOutput); err != nil {
		return nil, fmt.Errorf("failed to parse flake show output: %w", err)
	}

	// Filter systems if specified
	systemSet := make(map[string]bool)
	if len(systems) > 0 {
		for _, sys := range systems {
			systemSet[sys] = true
		}
	}

	var outputs []FlakeOutput

	// Enumerate per-system outputs
	for _, category := range AllPerSystemCategories() {
		var attrsBySystem map[string]map[string]FlakeShowVal
		switch category {
		case OutputCategoryPackages:
			attrsBySystem = showOutput.Packages
		case OutputCategoryChecks:
			attrsBySystem = showOutput.Checks
		case OutputCategoryDevShells:
			attrsBySystem = showOutput.DevShells
		case OutputCategoryApps:
			attrsBySystem = showOutput.Apps
		case OutputCategoryLegacyPackages:
			// Skip legacyPackages for now - they may have complex nested structure
			continue
		}

		for sys, attrs := range attrsBySystem {
			// Filter by system if specified
			if len(systemSet) > 0 && !systemSet[sys] {
				continue
			}

			for name, val := range attrs {
				// Only include derivations (things that can be built)
				if val.Type == "derivation" || category == OutputCategoryApps {
					outputs = append(outputs, FlakeOutput{
						Category: category,
						System:   sys,
						Name:     name,
						FlakeRef: fmt.Sprintf("%s#%s.%s.%s", flakeURL.String(), category, sys, name),
					})
				}
			}
		}
	}

	// Enumerate flake-level outputs (nixosConfigurations, darwinConfigurations)
	for cfgName := range showOutput.NixosConfigurations {
		// NixOS configurations are for Linux systems
		if len(systemSet) > 0 {
			hasLinux := false
			for sys := range systemSet {
				if strings.Contains(sys, "linux") {
					hasLinux = true
					break
				}
			}
			if !hasLinux {
				continue
			}
		}
		outputs = append(outputs, FlakeOutput{
			Category: OutputCategoryNixosConfigurations,
			Name:     cfgName,
			FlakeRef: fmt.Sprintf("%s#nixosConfigurations.%s.config.system.build.toplevel", flakeURL.String(), cfgName),
		})
	}

	for cfgName := range showOutput.DarwinConfigurations {
		// Darwin configurations are for Darwin systems
		if len(systemSet) > 0 {
			hasDarwin := false
			for sys := range systemSet {
				if strings.Contains(sys, "darwin") {
					hasDarwin = true
					break
				}
			}
			if !hasDarwin {
				continue
			}
		}
		outputs = append(outputs, FlakeOutput{
			Category: OutputCategoryDarwinConfigurations,
			Name:     cfgName,
			FlakeRef: fmt.Sprintf("%s#darwinConfigurations.%s.config.system.build.toplevel", flakeURL.String(), cfgName),
		})
	}

	// Sort outputs for deterministic ordering
	sort.Slice(outputs, func(i, j int) bool {
		if outputs[i].Category != outputs[j].Category {
			return outputs[i].Category < outputs[j].Category
		}
		if outputs[i].System != outputs[j].System {
			return outputs[i].System < outputs[j].System
		}
		return outputs[i].Name < outputs[j].Name
	})

	return outputs, nil
}

// BuildResult represents the result of building a flake output
type BuildResult struct {
	// Output is the flake output that was built
	Output FlakeOutput
	// StorePath is the built store path (empty if build failed)
	StorePath store.Path
	// Error is the error if the build failed
	Error error
}

// BuildOutput builds a single flake output and returns the store path
func BuildOutput(ctx context.Context, output FlakeOutput, impure bool, overrideInputs map[string]string) (store.Path, error) {
	cmd := NewCmd()

	args := []string{
		"build",
		output.FlakeRef,
		"-L",
		"--no-link",
		"--print-out-paths",
	}

	if impure {
		args = append(args, "--impure")
	}

	// Add override inputs
	inputNames := make([]string, 0, len(overrideInputs))
	for name := range overrideInputs {
		inputNames = append(inputNames, name)
	}
	sort.Strings(inputNames)

	for _, name := range inputNames {
		args = append(args, "--override-input", name, overrideInputs[name])
	}

	result, err := cmd.Run(ctx, args...)
	if err != nil {
		return store.Path{}, fmt.Errorf("failed to build %s: %w", output.FlakeRef, err)
	}

	// Parse the output path
	pathStr := strings.TrimSpace(result)
	if pathStr == "" {
		return store.Path{}, fmt.Errorf("build returned empty output for %s", output.FlakeRef)
	}

	// Handle multiple paths (some builds may have multiple outputs)
	paths := strings.Split(pathStr, "\n")
	return store.NewPath(strings.TrimSpace(paths[0])), nil
}

// BuildAllOutputsOptions contains options for BuildAllOutputs
type BuildAllOutputsOptions struct {
	// Impure enables impure builds
	Impure bool
	// OverrideInputs maps input names to flake URLs to override
	OverrideInputs map[string]string
	// Parallel enables parallel builds
	Parallel bool
	// MaxConcurrency limits the number of parallel builds (0 = unlimited)
	MaxConcurrency int
}

// BuildAllOutputs builds all provided flake outputs
func BuildAllOutputs(ctx context.Context, outputs []FlakeOutput, opts BuildAllOutputsOptions) []BuildResult {
	if len(outputs) == 0 {
		return nil
	}

	if !opts.Parallel {
		return buildOutputsSequential(ctx, outputs, opts)
	}

	return buildOutputsParallel(ctx, outputs, opts)
}

// buildOutputsSequential builds outputs one at a time
func buildOutputsSequential(ctx context.Context, outputs []FlakeOutput, opts BuildAllOutputsOptions) []BuildResult {
	results := make([]BuildResult, len(outputs))

	for i, output := range outputs {
		path, err := BuildOutput(ctx, output, opts.Impure, opts.OverrideInputs)
		results[i] = BuildResult{
			Output:    output,
			StorePath: path,
			Error:     err,
		}
	}

	return results
}

// buildOutputsParallel builds outputs concurrently
func buildOutputsParallel(ctx context.Context, outputs []FlakeOutput, opts BuildAllOutputsOptions) []BuildResult {
	results := make([]BuildResult, len(outputs))

	// Determine concurrency limit
	maxConcurrency := opts.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(outputs)
	}

	// Create work channel
	type job struct {
		index  int
		output FlakeOutput
	}

	jobs := make(chan job, len(outputs))
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < maxConcurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				path, err := BuildOutput(ctx, j.output, opts.Impure, opts.OverrideInputs)
				results[j.index] = BuildResult{
					Output:    j.output,
					StorePath: path,
					Error:     err,
				}
			}
		}()
	}

	// Queue jobs
	for i, output := range outputs {
		jobs <- job{index: i, output: output}
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	return results
}

// DevourFlakeNative builds all outputs of a flake using native Go implementation
// This is an alternative to DevourFlake that doesn't rely on the external devour-flake flake
func DevourFlakeNative(ctx context.Context, flakeURL FlakeURL, systems []string, impure bool, overrideInputs map[string]string, parallel bool, maxConcurrency int) (*DevourFlakeOutput, error) {
	// Enumerate all outputs
	outputs, err := EnumerateFlakeOutputs(ctx, flakeURL, systems)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate flake outputs: %w", err)
	}

	if len(outputs) == 0 {
		return &DevourFlakeOutput{
			OutPaths: []store.Path{},
			ByName:   make(map[string]store.Path),
		}, nil
	}

	// Build all outputs
	buildResults := BuildAllOutputs(ctx, outputs, BuildAllOutputsOptions{
		Impure:         impure,
		OverrideInputs: overrideInputs,
		Parallel:       parallel,
		MaxConcurrency: maxConcurrency,
	})

	// Collect results
	var outPaths []store.Path
	byName := make(map[string]store.Path)
	logger := common.Logger()

	for _, result := range buildResults {
		if result.Error != nil {
			// Log error but continue - some outputs may fail
			logger.Warn("failed to build flake output",
				zap.String("flakeRef", result.Output.FlakeRef),
				zap.Error(result.Error))
			continue
		}

		outPaths = append(outPaths, result.StorePath)

		// Use the output name as key
		name := result.Output.Name
		if name != "" {
			byName[name] = result.StorePath
		}
	}

	// Remove duplicates
	outPaths = uniquePaths(outPaths)

	return &DevourFlakeOutput{
		OutPaths: outPaths,
		ByName:   byName,
	}, nil
}

// GetCurrentSystem returns the current system string
func GetCurrentSystem() string {
	return flake.GetCurrentSystem().String()
}
