package ci

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_EmptyConfig(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflakes := map[string]SubflakeConfig{}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	results, err := Run(ctx, flake, subflakes, opts)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRun_SkippedSubflake(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflakes := map[string]SubflakeConfig{
		"test": {
			Dir:  "test",
			Skip: true, // This should be skipped
		},
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	results, err := Run(ctx, flake, subflakes, opts)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRun_SystemMismatch(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflakes := map[string]SubflakeConfig{
		"test": {
			Dir:     "test",
			Skip:    false,
			Systems: []string{"aarch64-darwin"}, // Only darwin
		},
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"}, // Requesting linux
	}

	results, err := Run(ctx, flake, subflakes, opts)
	require.NoError(t, err)
	assert.Empty(t, results) // Should skip due to system mismatch
}

func TestRunSubflake_InvalidFlakeURL(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflake := SubflakeConfig{
		Dir:  "../invalid/../path",
		Skip: false,
		Steps: StepsConfig{
			Build:      BuildStep{Enable: false},
			Lockfile:   LockfileStep{Enable: false},
			FlakeCheck: FlakeCheckStep{Enable: false},
		},
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	// This should handle the error gracefully or succeed
	result, err := runSubflake(ctx, flake, "test", subflake, opts)
	// Verify it doesn't panic - either succeeds or fails gracefully
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.Equal(t, "test", result.Subflake)
	}
}

func TestStepResult_JSONSerialization(t *testing.T) {
	result := StepResult{
		Name:    "test-step",
		Success: true,
		Output:  "test output",
		Error:   "",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-step")
	assert.Contains(t, string(data), "test output")

	var decoded StepResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, result.Name, decoded.Name)
	assert.Equal(t, result.Success, decoded.Success)
}

func TestResult_JSONSerialization(t *testing.T) {
	result := Result{
		Subflake: "test",
		Success:  true,
		Steps: map[string]StepResult{
			"build": {
				Name:    "build",
				Success: true,
			},
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test")
	assert.Contains(t, string(data), "build")

	var decoded Result
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, result.Subflake, decoded.Subflake)
	assert.Equal(t, result.Success, decoded.Success)
}

func TestRunOptions_DefaultValues(t *testing.T) {
	opts := RunOptions{}
	assert.Empty(t, opts.Systems)
	assert.False(t, opts.GitHubOutput)
	assert.False(t, opts.IncludeAllDependencies)
}

func TestBuildStep_DefaultValues(t *testing.T) {
	step := BuildStep{}
	assert.False(t, step.Enable)
	assert.False(t, step.Impure)
}

func TestLockfileStep_DefaultValues(t *testing.T) {
	step := LockfileStep{}
	assert.False(t, step.Enable)
}

func TestFlakeCheckStep_DefaultValues(t *testing.T) {
	step := FlakeCheckStep{}
	assert.False(t, step.Enable)
}

func TestCustomStep_DefaultValues(t *testing.T) {
	step := CustomStep{}
	assert.Empty(t, step.Name)
	assert.Empty(t, step.Command)
	assert.Empty(t, step.Type)
}

func TestStepsConfig_NoStepsEnabled(t *testing.T) {
	config := StepsConfig{
		Build:      BuildStep{Enable: false},
		Lockfile:   LockfileStep{Enable: false},
		FlakeCheck: FlakeCheckStep{Enable: false},
		Custom:     make(map[string]CustomStep),
	}

	enabled := config.GetEnabledSteps()
	assert.Empty(t, enabled)
}

func TestStepsConfig_MultipleCustomSteps(t *testing.T) {
	config := StepsConfig{
		Build:      BuildStep{Enable: false},
		Lockfile:   LockfileStep{Enable: false},
		FlakeCheck: FlakeCheckStep{Enable: false},
		Custom: map[string]CustomStep{
			"test1": {Type: CustomStepTypeApp},
			"test2": {Type: CustomStepTypeDevShell},
			"test3": {Type: CustomStepTypeApp},
		},
	}

	enabled := config.GetEnabledSteps()
	// Custom steps are sorted alphabetically for deterministic order
	assert.Equal(t, []string{"custom:test1", "custom:test2", "custom:test3"}, enabled)
}

func TestGitHubMatrix_EmptyInclude(t *testing.T) {
	matrix := GitHubMatrix{
		Include: []GitHubMatrixRow{},
	}

	assert.Equal(t, 0, matrix.Count())

	json, err := matrix.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, json, "include")
}

func TestGitHubMatrix_LargeMatrix(t *testing.T) {
	systems := []string{"x86_64-linux", "aarch64-linux", "x86_64-darwin", "aarch64-darwin"}
	config := Config{
		Configs: map[string]map[string]SubflakeConfig{
			"default": {
				".":      {Dir: ".", Skip: false},
				"tests":  {Dir: "tests", Skip: false},
				"docs":   {Dir: "docs", Skip: false},
				"extras": {Dir: "extras", Skip: false},
			},
		},
	}

	matrix, err := GenerateMatrix(systems, config)
	require.NoError(t, err)

	// 4 systems × 4 subflakes = 16 rows
	assert.Equal(t, 16, matrix.Count())

	json, err := matrix.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, json, "x86_64-linux")
	assert.Contains(t, json, "aarch64-darwin")
}

func TestToJSON_ErrorHandling(t *testing.T) {
	// Create a matrix with valid data
	matrix := GitHubMatrix{
		Include: []GitHubMatrixRow{
			{System: "x86_64-linux", Subflake: "."},
		},
	}

	jsonStr, err := matrix.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Verify it's valid JSON
	var decoded GitHubMatrix
	err = json.Unmarshal([]byte(jsonStr), &decoded)
	require.NoError(t, err)
}

func TestGenerateMatrix_MissingDefaultConfig(t *testing.T) {
	systems := []string{"x86_64-linux"}
	config := Config{
		Configs: map[string]map[string]SubflakeConfig{
			"production": {
				".": {Dir: ".", Skip: false},
			},
		},
	}

	// Should error because there's no "default" config
	_, err := GenerateMatrix(systems, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default")
}

func TestOverrideInputs_LockfileSkipped(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	step := LockfileStep{
		Enable: true,
	}

	// Subflake with override inputs should skip lockfile check
	subflake := SubflakeConfig{
		Dir: ".",
		OverrideInputs: map[string]string{
			"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
		},
	}

	result := runLockfileStep(ctx, flake, step, subflake)
	assert.Equal(t, "lockfile", result.Name)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Skipped")
}

func TestOverrideInputs_NoSkipWhenEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix/main")
	require.NoError(t, err)

	step := LockfileStep{
		Enable: true,
	}

	// Subflake without override inputs should run lockfile check
	subflake := SubflakeConfig{
		Dir:            ".",
		OverrideInputs: make(map[string]string),
	}

	result := runLockfileStep(ctx, flake, step, subflake)
	assert.Equal(t, "lockfile", result.Name)
	// Result may succeed or fail, but it should not be skipped
	assert.NotContains(t, result.Output, "Skipped")
}

// TestGenerateMatrix_DeterministicOrder verifies that the GitHub matrix is generated
// with subflakes in deterministic (alphabetically sorted) order, matching the Rust
// implementation which uses BTreeMap.
func TestGenerateMatrix_DeterministicOrder(t *testing.T) {
	systems := []string{"x86_64-linux", "aarch64-darwin"}
	config := Config{
		Configs: map[string]map[string]SubflakeConfig{
			"default": {
				"zebra":  {Dir: "zebra", Skip: false},
				"alpha":  {Dir: "alpha", Skip: false},
				"middle": {Dir: "middle", Skip: false},
				"beta":   {Dir: "beta", Skip: false},
			},
		},
	}

	// Run multiple iterations to verify deterministic ordering
	for i := 0; i < 10; i++ {
		matrix, err := GenerateMatrix(systems, config)
		require.NoError(t, err)

		// 2 systems × 4 subflakes = 8 rows
		require.Equal(t, 8, matrix.Count())

		// For each system, subflakes should be in alphabetical order
		// First system (x86_64-linux) rows 0-3
		// Second system (aarch64-darwin) rows 4-7
		expectedSubflakeOrder := []string{"alpha", "beta", "middle", "zebra"}

		// Check first system's rows
		for j := 0; j < 4; j++ {
			assert.Equal(t, "x86_64-linux", matrix.Include[j].System,
				"iteration %d, row %d: expected system x86_64-linux", i, j)
			assert.Equal(t, expectedSubflakeOrder[j], matrix.Include[j].Subflake,
				"iteration %d, row %d: expected subflake %s, got %s",
				i, j, expectedSubflakeOrder[j], matrix.Include[j].Subflake)
		}

		// Check second system's rows
		for j := 0; j < 4; j++ {
			assert.Equal(t, "aarch64-darwin", matrix.Include[j+4].System,
				"iteration %d, row %d: expected system aarch64-darwin", i, j+4)
			assert.Equal(t, expectedSubflakeOrder[j], matrix.Include[j+4].Subflake,
				"iteration %d, row %d: expected subflake %s, got %s",
				i, j+4, expectedSubflakeOrder[j], matrix.Include[j+4].Subflake)
		}
	}
}
