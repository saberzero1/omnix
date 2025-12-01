package nix

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix/flake"
	"github.com/saberzero1/omnix/pkg/nix/store"
	"go.uber.org/zap"
)

// storePathRegex validates Nix store paths
// Format: /nix/store/<32-char-hash>-<name>[/optional/path]
// The pattern ensures:
// - No trailing slashes
// - No multiple consecutive slashes
// - No empty path segments
var storePathRegex = regexp.MustCompile(`^/nix/store/[a-z0-9]{32}-[a-zA-Z0-9._+-]+(/[a-zA-Z0-9._+-]+)*$`)

// dirTraversalRegex detects directory traversal patterns
var dirTraversalRegex = regexp.MustCompile(`(^|/)\.\.?(/|$)`)

// isValidStorePath checks if a string is a valid Nix store path
// and rejects directory traversal attempts
func isValidStorePath(path string) bool {
	// First check basic format
	if !storePathRegex.MatchString(path) {
		return false
	}
	// Reject directory traversal patterns (. or .. as path segments)
	if dirTraversalRegex.MatchString(path) {
		return false
	}
	return true
}

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
	// OutputCategoryHomeConfigurations represents homeConfigurations output
	OutputCategoryHomeConfigurations OutputCategory = "homeConfigurations"
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
		OutputCategoryHomeConfigurations,
	}
}

// FlakeShowOutput represents the output of `nix flake show --json`
type FlakeShowOutput struct {
	Packages             map[string]map[string]FlakeShowVal `json:"packages,omitempty"`
	Checks               map[string]map[string]FlakeShowVal `json:"checks,omitempty"`
	DevShells            map[string]map[string]FlakeShowVal `json:"devShells,omitempty"`
	Apps                 map[string]map[string]FlakeShowVal `json:"apps,omitempty"`
	LegacyPackages       map[string]map[string]interface{}  `json:"legacyPackages,omitempty"`
	NixosConfigurations  map[string]interface{}             `json:"nixosConfigurations,omitempty"`
	DarwinConfigurations map[string]interface{}             `json:"darwinConfigurations,omitempty"`
	HomeConfigurations   map[string]interface{}             `json:"homeConfigurations,omitempty"`
}

// FlakeShowVal represents a terminal value in `nix flake show` output
type FlakeShowVal struct {
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// flakeShowMetadataKeys contains keys that nix flake show adds at the top level
// of configuration objects (nixosConfigurations, darwinConfigurations, homeConfigurations)
// when they cannot be fully evaluated. These should be skipped when enumerating configurations.
var flakeShowMetadataKeys = map[string]bool{
	"type":        true,
	"name":        true,
	"description": true,
}

// isFlakeShowMetadataKey returns true if the key is a metadata key from nix flake show output
func isFlakeShowMetadataKey(key string) bool {
	return flakeShowMetadataKeys[key]
}

// hasSystemMatching checks if systemSet contains any system matching the substring
func hasSystemMatching(systemSet map[string]bool, substring string) bool {
	if len(systemSet) == 0 {
		return true
	}
	for sys := range systemSet {
		if strings.Contains(sys, substring) {
			return true
		}
	}
	return false
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
			// Apps need special handling - enumerate them using nix eval to get the .program store path
			outputs = append(outputs, enumerateApps(ctx, flakeURL, showOutput.Apps, systemSet)...)
			continue
		case OutputCategoryLegacyPackages:
			// Handle legacyPackages specially - look for homeConfigurations
			outputs = append(outputs, enumerateLegacyPackages(flakeURL, showOutput.LegacyPackages, systemSet)...)
			continue
		}

		for sys, attrs := range attrsBySystem {
			// Filter by system if specified
			if len(systemSet) > 0 && !systemSet[sys] {
				continue
			}

			for name, val := range attrs {
				// Only include derivations (things that can be built)
				if val.Type == "derivation" {
					flakeRef := fmt.Sprintf("%s#%s.%s.%s", flakeURL.String(), category, sys, name)
					outputs = append(outputs, FlakeOutput{
						Category: category,
						System:   sys,
						Name:     name,
						FlakeRef: flakeRef,
					})
				}
			}
		}
	}

	// Enumerate flake-level outputs (nixosConfigurations, darwinConfigurations)
	for cfgName := range showOutput.NixosConfigurations {
		// Skip metadata keys that nix flake show adds at the top level when configs can't be evaluated
		if isFlakeShowMetadataKey(cfgName) {
			continue
		}
		// NixOS configurations are for Linux systems
		if !hasSystemMatching(systemSet, "linux") {
			continue
		}
		outputs = append(outputs, FlakeOutput{
			Category: OutputCategoryNixosConfigurations,
			Name:     cfgName,
			FlakeRef: fmt.Sprintf("%s#nixosConfigurations.%s.config.system.build.toplevel", flakeURL.String(), cfgName),
		})
	}

	for cfgName := range showOutput.DarwinConfigurations {
		// Skip metadata keys that nix flake show adds at the top level when configs can't be evaluated
		if isFlakeShowMetadataKey(cfgName) {
			continue
		}
		// Darwin configurations are for Darwin systems
		if !hasSystemMatching(systemSet, "darwin") {
			continue
		}
		outputs = append(outputs, FlakeOutput{
			Category: OutputCategoryDarwinConfigurations,
			Name:     cfgName,
			FlakeRef: fmt.Sprintf("%s#darwinConfigurations.%s.config.system.build.toplevel", flakeURL.String(), cfgName),
		})
	}

	// Enumerate flake-level homeConfigurations
	// Home Manager configurations expose .activationPackage for the activation script
	for cfgName := range showOutput.HomeConfigurations {
		// Skip metadata keys that nix flake show adds at the top level when configs can't be evaluated
		if isFlakeShowMetadataKey(cfgName) {
			continue
		}
		outputs = append(outputs, FlakeOutput{
			Category: OutputCategoryHomeConfigurations,
			Name:     cfgName,
			FlakeRef: fmt.Sprintf("%s#homeConfigurations.%s.activationPackage", flakeURL.String(), cfgName),
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

// enumerateLegacyPackages handles legacyPackages which may contain homeConfigurations
// devour-flake uses legacyPackages.${system}.homeConfigurations.<name>.activationPackage
// for Home Manager configurations since outputs.homeConfigurations doesn't specify which architecture to build for
func enumerateLegacyPackages(flakeURL FlakeURL, legacyPackages map[string]map[string]interface{}, systemSet map[string]bool) []FlakeOutput {
	var outputs []FlakeOutput
	logger := common.Logger()

	for sys, attrs := range legacyPackages {
		// Filter by system if specified
		if len(systemSet) > 0 && !systemSet[sys] {
			continue
		}

		// Look for homeConfigurations within legacyPackages.${system}
		homeConfigs, ok := attrs["homeConfigurations"]
		if !ok {
			continue
		}

		// homeConfigurations is a map of config names to config objects
		configsMap, ok := homeConfigs.(map[string]interface{})
		if !ok {
			logger.Warn("unexpected type for homeConfigurations",
				zap.String("system", sys),
				zap.String("type", fmt.Sprintf("%T", homeConfigs)))
			continue
		}

		for configName := range configsMap {
			// Build the activationPackage for each Home Manager configuration
			// The flake ref is: legacyPackages.${system}.homeConfigurations.${name}.activationPackage
			outputs = append(outputs, FlakeOutput{
				Category: OutputCategoryLegacyPackages,
				System:   sys,
				Name:     configName,
				FlakeRef: fmt.Sprintf("%s#legacyPackages.%s.homeConfigurations.%s.activationPackage", flakeURL.String(), sys, configName),
			})
		}
	}

	return outputs
}

// enumerateApps handles apps by evaluating app.program to get the store path
// Apps are metadata that reference programs from other outputs. Unlike derivations,
// they can't be built directly. Instead, we use nix eval to get the .program store path
// which is a direct reference to the executable in the Nix store.
func enumerateApps(ctx context.Context, flakeURL FlakeURL, apps map[string]map[string]FlakeShowVal, systemSet map[string]bool) []FlakeOutput {
	var outputs []FlakeOutput
	logger := common.Logger()
	cmd := NewCmd()

	for sys, attrs := range apps {
		// Filter by system if specified
		if len(systemSet) > 0 && !systemSet[sys] {
			continue
		}

		for name, val := range attrs {
			// Only include apps (type: "app")
			if val.Type != "app" {
				continue
			}

			// Use nix eval to get the .program store path
			// The app.program attribute contains the direct path to the executable
			appRef := fmt.Sprintf("%s#apps.%s.%s.program", flakeURL.String(), sys, name)
			evalOutput, err := cmd.Run(ctx, "eval", "--raw", appRef)
			if err != nil {
				logger.Warn("failed to evaluate app.program",
					zap.String("app", appRef),
					zap.Error(err))
				continue
			}

			storePath := strings.TrimSpace(evalOutput)
			if storePath == "" {
				logger.Warn("app.program evaluated to empty string",
					zap.String("app", appRef))
				continue
			}

			// Validate the store path to prevent security issues
			if !isValidStorePath(storePath) {
				logger.Warn("app.program is not a valid store path",
					zap.String("app", appRef),
					zap.String("storePath", storePath))
				continue
			}

			// The store path from app.program is a direct store path, not a flake ref
			// We create a special FlakeOutput that uses the store path directly
			outputs = append(outputs, FlakeOutput{
				Category: OutputCategoryApps,
				System:   sys,
				Name:     name,
				FlakeRef: storePath, // This is the store path, not a flake ref
			})
		}
	}

	return outputs
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

// BuildOutput builds a single flake output and returns the store path.
// If impure is true, passes --impure to the build command.
// overrideInputs maps input names to flake URLs that override the flake's inputs.
// When a build produces multiple outputs (e.g., out, dev, doc), only the first path is returned.
// Returns an error if the build fails or produces no output.
// For apps, the FlakeRef is a pre-evaluated store path, so we return it directly.
func BuildOutput(ctx context.Context, output FlakeOutput, impure bool, overrideInputs map[string]string) (store.Path, error) {
	// For apps, the FlakeRef is already a store path (evaluated via nix eval)
	// We don't need to build it, just return the path directly
	if output.Category == OutputCategoryApps && isValidStorePath(output.FlakeRef) {
		return store.NewPath(output.FlakeRef), nil
	}

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

	// Handle multiple paths (some builds may have multiple outputs like out, dev, doc).
	// We only return the first path as that is typically the main output.
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
	processed := make([]bool, len(outputs)) // Track which jobs were processed
	var mu sync.Mutex                       // Protects writes to results and processed slices

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
				// Check if context is cancelled before building
				if ctx.Err() != nil {
					mu.Lock()
					results[j.index] = BuildResult{
						Output: j.output,
						Error:  fmt.Errorf("build cancelled: %w", ctx.Err()),
					}
					processed[j.index] = true
					mu.Unlock()
					continue
				}

				path, err := BuildOutput(ctx, j.output, opts.Impure, opts.OverrideInputs)
				result := BuildResult{
					Output:    j.output,
					StorePath: path,
					Error:     err,
				}
				mu.Lock()
				results[j.index] = result
				processed[j.index] = true
				mu.Unlock()
			}
		}()
	}

	// Queue all jobs - do not exit early on context cancellation
	// This ensures all job slots are properly handled by workers
	for i, output := range outputs {
		jobs <- job{index: i, output: output}
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()

	// Mark any unprocessed jobs as cancelled (this should not happen with the new design,
	// but is a safety measure)
	for i, output := range outputs {
		if !processed[i] {
			results[i] = BuildResult{
				Output: output,
				Error:  fmt.Errorf("internal error: job was queued but never processed by worker"),
			}
		}
	}

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
	var failedCount int

	for _, result := range buildResults {
		// Check for zero-valued results (should not happen with fixed buildOutputsParallel,
		// but this is a safety check for any unprocessed jobs)
		if result.Output.FlakeRef == "" {
			failedCount++
			logger.Warn("skipping build result with empty FlakeRef (indicates unprocessed job)")
			continue
		}

		if result.Error != nil {
			failedCount++
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

	// Return error if all builds failed
	if failedCount > 0 && failedCount == len(buildResults) {
		return nil, fmt.Errorf("all %d builds failed", failedCount)
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
