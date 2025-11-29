package nix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllPerSystemCategories(t *testing.T) {
	categories := AllPerSystemCategories()

	assert.Contains(t, categories, OutputCategoryPackages)
	assert.Contains(t, categories, OutputCategoryChecks)
	assert.Contains(t, categories, OutputCategoryDevShells)
	assert.Contains(t, categories, OutputCategoryApps)
	assert.Contains(t, categories, OutputCategoryLegacyPackages)
	assert.Len(t, categories, 5)
}

func TestAllFlakeLevelCategories(t *testing.T) {
	categories := AllFlakeLevelCategories()

	assert.Contains(t, categories, OutputCategoryNixosConfigurations)
	assert.Contains(t, categories, OutputCategoryDarwinConfigurations)
	assert.Len(t, categories, 2)
}

func TestOutputCategory_Constants(t *testing.T) {
	assert.Equal(t, OutputCategory("packages"), OutputCategoryPackages)
	assert.Equal(t, OutputCategory("checks"), OutputCategoryChecks)
	assert.Equal(t, OutputCategory("devShells"), OutputCategoryDevShells)
	assert.Equal(t, OutputCategory("apps"), OutputCategoryApps)
	assert.Equal(t, OutputCategory("legacyPackages"), OutputCategoryLegacyPackages)
	assert.Equal(t, OutputCategory("nixosConfigurations"), OutputCategoryNixosConfigurations)
	assert.Equal(t, OutputCategory("darwinConfigurations"), OutputCategoryDarwinConfigurations)
}

func TestFlakeOutput_FlakeRef(t *testing.T) {
	output := FlakeOutput{
		Category: OutputCategoryPackages,
		System:   "x86_64-linux",
		Name:     "default",
		FlakeRef: ".#packages.x86_64-linux.default",
	}

	assert.Equal(t, OutputCategoryPackages, output.Category)
	assert.Equal(t, "x86_64-linux", output.System)
	assert.Equal(t, "default", output.Name)
	assert.Equal(t, ".#packages.x86_64-linux.default", output.FlakeRef)
}

func TestBuildAllOutputsOptions_Defaults(t *testing.T) {
	opts := BuildAllOutputsOptions{}

	assert.False(t, opts.Impure)
	assert.Nil(t, opts.OverrideInputs)
	assert.False(t, opts.Parallel)
	assert.Zero(t, opts.MaxConcurrency)
}

func TestBuildResult_Success(t *testing.T) {
	output := FlakeOutput{
		Category: OutputCategoryPackages,
		System:   "x86_64-linux",
		Name:     "hello",
		FlakeRef: ".#packages.x86_64-linux.hello",
	}

	result := BuildResult{
		Output: output,
		Error:  nil,
	}

	assert.NoError(t, result.Error)
	assert.Equal(t, "hello", result.Output.Name)
}

func TestGetCurrentSystem(t *testing.T) {
	sys := GetCurrentSystem()

	// Should return a valid system string
	sysStr := sys
	assert.NotEmpty(t, sysStr)
}

func TestBuildAllOutputs_Empty(t *testing.T) {
	results := BuildAllOutputs(context.Background(), nil, BuildAllOutputsOptions{})
	assert.Nil(t, results)
}

func TestBuildAllOutputsOptions_WithOverrideInputs(t *testing.T) {
	opts := BuildAllOutputsOptions{
		Impure: true,
		OverrideInputs: map[string]string{
			"nixpkgs": "github:NixOS/nixpkgs/nixos-24.11",
		},
		Parallel:       true,
		MaxConcurrency: 4,
	}

	assert.True(t, opts.Impure)
	assert.Equal(t, "github:NixOS/nixpkgs/nixos-24.11", opts.OverrideInputs["nixpkgs"])
	assert.True(t, opts.Parallel)
	assert.Equal(t, 4, opts.MaxConcurrency)
}

// TestEnumerateFlakeOutputs_Integration tests the actual enumeration of flake outputs
// This is an integration test that requires Nix to be available
func TestEnumerateFlakeOutputs_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flakeURL := NewFlakeURL(".")

	// Get current system
	info, err := GetInfo(ctx)
	if err != nil {
		t.Skip("Nix not available: ", err)
	}

	systems := []string{info.Config.System.Value}

	outputs, err := EnumerateFlakeOutputs(ctx, flakeURL, systems)

	// Should not return error
	assert.NoError(t, err, "EnumerateFlakeOutputs should not return error")

	// Should return at least some outputs (this repo has packages)
	assert.NotEmpty(t, outputs, "Should have at least some flake outputs")

	// Log the outputs for debugging
	t.Logf("Found %d outputs", len(outputs))
	for _, output := range outputs[:min(5, len(outputs))] {
		t.Logf("  - %s: %s", output.Category, output.FlakeRef)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
