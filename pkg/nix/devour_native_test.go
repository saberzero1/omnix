package nix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidStorePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid store path with simple name",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello",
			expected: true,
		},
		{
			name:     "valid store path with version",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello-1.0.0",
			expected: true,
		},
		{
			name:     "valid store path with nested path",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/bin/hello",
			expected: true,
		},
		{
			name:     "valid store path with dots and underscores",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-my_app.v1.0",
			expected: true,
		},
		{
			name:     "invalid - not starting with /nix/store/",
			path:     "/tmp/nix/store/abc123def456ghi789jkl012mno345pq-hello",
			expected: false,
		},
		{
			name:     "invalid - hash too short",
			path:     "/nix/store/abc123-hello",
			expected: false,
		},
		{
			name:     "invalid - hash with uppercase",
			path:     "/nix/store/ABC123def456ghi789jkl012mno345pq-hello",
			expected: false,
		},
		{
			name:     "invalid - empty string",
			path:     "",
			expected: false,
		},
		{
			name:     "invalid - relative path",
			path:     "nix/store/abc123def456ghi789jkl012mno345pq-hello",
			expected: false,
		},
		{
			name:     "invalid - missing name after hash",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq",
			expected: false,
		},
		{
			name:     "invalid - trailing slash",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/",
			expected: false,
		},
		{
			name:     "invalid - multiple consecutive slashes",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello//bin",
			expected: false,
		},
		{
			name:     "invalid - empty path segment",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/bin//app",
			expected: false,
		},
		{
			name:     "invalid - directory traversal with ..",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/../etc/passwd",
			expected: false,
		},
		{
			name:     "invalid - directory traversal at end",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/bin/..",
			expected: false,
		},
		{
			name:     "invalid - single dot traversal",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/./bin",
			expected: false,
		},
		{
			name:     "valid - dot in segment name",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/.envrc",
			expected: true,
		},
		{
			name:     "valid - dots in segment name",
			path:     "/nix/store/abc123def456ghi789jkl012mno345pq-hello/file..txt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStorePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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

// TestBuildOutput_AppWithStorePath tests that apps with store paths are returned directly
func TestBuildOutput_AppWithStorePath(t *testing.T) {
	ctx := context.Background()

	// Create an app output with a valid store path (simulating what enumerateApps produces)
	// The hash must be exactly 32 lowercase alphanumeric characters
	output := FlakeOutput{
		Category: OutputCategoryApps,
		System:   "x86_64-linux",
		Name:     "test-app",
		FlakeRef: "/nix/store/abc123def456ghi789jkl012mno345pq-test-app/bin/test-app",
	}

	// BuildOutput should return the store path directly for apps
	path, err := BuildOutput(ctx, output, false, nil)

	// Should succeed without actually calling nix build
	assert.NoError(t, err)
	assert.Equal(t, "/nix/store/abc123def456ghi789jkl012mno345pq-test-app/bin/test-app", path.String())
}

// TestBuildOutput_AppWithFlakeRef tests that apps with invalid FlakeRef (not a store path)
// fall through to the normal build path
func TestBuildOutput_AppWithFlakeRef(t *testing.T) {
	ctx := context.Background()

	// Create an app output with a flake ref instead of a store path
	output := FlakeOutput{
		Category: OutputCategoryApps,
		System:   "x86_64-linux",
		Name:     "test-app",
		FlakeRef: ".#apps.x86_64-linux.test-app", // Not a store path
	}

	// This should attempt to build (and fail since we're not in a real flake)
	_, err := BuildOutput(ctx, output, false, nil)

	// Expected to fail since we're not in a real flake environment
	assert.Error(t, err)
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

func TestEnumerateLegacyPackages(t *testing.T) {
	flakeURL := NewFlakeURL(".")

	tests := []struct {
		name           string
		legacyPackages map[string]map[string]interface{}
		systemSet      map[string]bool
		expectedCount  int
	}{
		{
			name:           "empty legacyPackages",
			legacyPackages: map[string]map[string]interface{}{},
			systemSet:      map[string]bool{},
			expectedCount:  0,
		},
		{
			name: "legacyPackages without homeConfigurations",
			legacyPackages: map[string]map[string]interface{}{
				"x86_64-linux": {
					"somePackage": map[string]interface{}{},
				},
			},
			systemSet:     map[string]bool{},
			expectedCount: 0,
		},
		{
			name: "legacyPackages with homeConfigurations",
			legacyPackages: map[string]map[string]interface{}{
				"x86_64-linux": {
					"homeConfigurations": map[string]interface{}{
						"user@host": map[string]interface{}{
							"activationPackage": "/nix/store/...",
						},
					},
				},
			},
			systemSet:     map[string]bool{},
			expectedCount: 1,
		},
		{
			name: "legacyPackages with multiple homeConfigurations",
			legacyPackages: map[string]map[string]interface{}{
				"x86_64-linux": {
					"homeConfigurations": map[string]interface{}{
						"user1": map[string]interface{}{},
						"user2": map[string]interface{}{},
					},
				},
				"aarch64-darwin": {
					"homeConfigurations": map[string]interface{}{
						"user3": map[string]interface{}{},
					},
				},
			},
			systemSet:     map[string]bool{},
			expectedCount: 3,
		},
		{
			name: "legacyPackages filtered by system",
			legacyPackages: map[string]map[string]interface{}{
				"x86_64-linux": {
					"homeConfigurations": map[string]interface{}{
						"user1": map[string]interface{}{},
					},
				},
				"aarch64-darwin": {
					"homeConfigurations": map[string]interface{}{
						"user2": map[string]interface{}{},
					},
				},
			},
			systemSet:     map[string]bool{"x86_64-linux": true},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := enumerateLegacyPackages(flakeURL, tt.legacyPackages, tt.systemSet)
			assert.Len(t, outputs, tt.expectedCount)

			// Check that all outputs have correct category
			for _, output := range outputs {
				assert.Equal(t, OutputCategoryLegacyPackages, output.Category)
				assert.Contains(t, output.FlakeRef, "homeConfigurations")
				assert.Contains(t, output.FlakeRef, "activationPackage")
			}
		})
	}
}

// TestBuildAllOutputs_ContextCancellation tests that context cancellation is handled properly
func TestBuildAllOutputs_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	outputs := []FlakeOutput{
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test1", FlakeRef: ".#packages.x86_64-linux.test1"},
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test2", FlakeRef: ".#packages.x86_64-linux.test2"},
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test3", FlakeRef: ".#packages.x86_64-linux.test3"},
	}

	// Run parallel builds with cancelled context
	results := BuildAllOutputs(ctx, outputs, BuildAllOutputsOptions{
		Parallel:       true,
		MaxConcurrency: 2,
	})

	// Should return results for all outputs
	assert.Len(t, results, len(outputs))

	// All results should have errors (context cancelled)
	for i, result := range results {
		assert.NotEmpty(t, result.Output.FlakeRef, "Result %d should have FlakeRef set", i)
		assert.Error(t, result.Error, "Result %d should have an error due to context cancellation", i)
	}
}

// TestBuildAllOutputs_AllResultsHaveFlakeRef tests that all results have FlakeRef set
func TestBuildAllOutputs_AllResultsHaveFlakeRef(t *testing.T) {
	ctx := context.Background()

	outputs := []FlakeOutput{
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test1", FlakeRef: ".#packages.x86_64-linux.test1"},
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test2", FlakeRef: ".#packages.x86_64-linux.test2"},
	}

	// Run sequential builds (will fail because flake doesn't exist, but that's OK)
	results := BuildAllOutputs(ctx, outputs, BuildAllOutputsOptions{
		Parallel: false,
	})

	// Should return results for all outputs
	assert.Len(t, results, len(outputs))

	// All results should have FlakeRef set (even if they failed)
	for i, result := range results {
		assert.Equal(t, outputs[i].FlakeRef, result.Output.FlakeRef, "Result %d should have correct FlakeRef", i)
	}
}

// TestBuildAllOutputs_ParallelAllResultsHaveFlakeRef tests parallel mode results
func TestBuildAllOutputs_ParallelAllResultsHaveFlakeRef(t *testing.T) {
	ctx := context.Background()

	outputs := []FlakeOutput{
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test1", FlakeRef: ".#packages.x86_64-linux.test1"},
		{Category: OutputCategoryPackages, System: "x86_64-linux", Name: "test2", FlakeRef: ".#packages.x86_64-linux.test2"},
	}

	// Run parallel builds (will fail because flake doesn't exist, but that's OK)
	results := BuildAllOutputs(ctx, outputs, BuildAllOutputsOptions{
		Parallel:       true,
		MaxConcurrency: 2,
	})

	// Should return results for all outputs
	assert.Len(t, results, len(outputs))

	// All results should have FlakeRef set (even if they failed)
	for i, result := range results {
		assert.NotEmpty(t, result.Output.FlakeRef, "Result %d should have FlakeRef set", i)
	}
}
