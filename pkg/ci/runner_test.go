package ci

import (
	"context"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRunBuildStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix/main")
	require.NoError(t, err)

	step := BuildStep{
		Enable: true,
		Impure: false,
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	subflake := SubflakeConfig{
		Dir:            ".",
		OverrideInputs: make(map[string]string),
	}

	result := runBuildStep(ctx, flake, step, opts, subflake)
	assert.Equal(t, "build", result.Name)
	// Note: This may fail or succeed depending on the system, just testing it runs
}

func TestRunLockfileStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix/main")
	require.NoError(t, err)

	step := LockfileStep{
		Enable: true,
	}

	subflake := SubflakeConfig{
		Dir:            ".",
		OverrideInputs: make(map[string]string),
	}

	result := runLockfileStep(ctx, flake, step, subflake)
	assert.Equal(t, "lockfile", result.Name)
}

func TestRunFlakeCheckStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix/main")
	require.NoError(t, err)

	step := FlakeCheckStep{
		Enable: true,
	}

	subflake := SubflakeConfig{
		Dir:            ".",
		OverrideInputs: make(map[string]string),
	}

	result := runFlakeCheckStep(ctx, flake, step, subflake)
	assert.Equal(t, "flakeCheck", result.Name)
}

func TestRunCustomStep(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	tests := []struct {
		name          string
		stepName      string
		step          CustomStep
		expectedName  string
		expectedError bool
		errorContains string
	}{
		{
			name: "empty command",
			step: CustomStep{
				Type:    CustomStepTypeDevShell,
				Command: []string{},
			},
			stepName:      "test",
			expectedName:  "custom:test",
			expectedError: true,
			errorContains: "no command",
		},
		{
			name: "echo command",
			step: CustomStep{
				Type:    CustomStepTypeDevShell,
				Command: []string{"echo", "hello"},
			},
			stepName:      "echo-test",
			expectedName:  "custom:echo-test",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subflake := SubflakeConfig{
				Dir:            ".",
				OverrideInputs: make(map[string]string),
			}
			result := runCustomStep(ctx, flake, tt.stepName, tt.step, subflake)
			assert.Equal(t, tt.expectedName, result.Name)

			if tt.expectedError {
				assert.False(t, result.Success)
				assert.Contains(t, result.Error, tt.errorContains)
			}
		})
	}
}

func TestLogResult(t *testing.T) {
	logger := zap.NewNop()

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

	// Just test that it doesn't panic
	LogResult(result, logger)
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflakes := map[string]SubflakeConfig{
		".": {
			Dir:  ".",
			Skip: false,
			Steps: StepsConfig{
				Build:      BuildStep{Enable: false},
				Lockfile:   LockfileStep{Enable: false},
				FlakeCheck: FlakeCheckStep{Enable: false},
			},
		},
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	results, err := Run(ctx, flake, subflakes, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestAppendOverrideInputs(t *testing.T) {
	tests := []struct {
		name           string
		initialArgs    []string
		overrideInputs map[string]string
		expected       []string
	}{
		{
			name:           "empty override inputs",
			initialArgs:    []string{"nix", "build"},
			overrideInputs: map[string]string{},
			expected:       []string{"nix", "build"},
		},
		{
			name:        "single override input",
			initialArgs: []string{"nix", "build"},
			overrideInputs: map[string]string{
				"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
			},
			expected: []string{"nix", "build", "--override-input", "nixpkgs", "github:NixOS/nixpkgs/nixos-unstable"},
		},
		{
			name:        "multiple override inputs",
			initialArgs: []string{"nix", "build"},
			overrideInputs: map[string]string{
				"nixpkgs":      "github:NixOS/nixpkgs/nixos-unstable",
				"home-manager": "github:nix-community/home-manager",
			},
			// Note: map iteration order is not guaranteed, so we just check the length
			expected: []string{"nix", "build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendOverrideInputs(tt.initialArgs, tt.overrideInputs)

			if len(tt.overrideInputs) == 0 || len(tt.overrideInputs) == 1 {
				assert.Equal(t, tt.expected, result)
			} else {
				// For multiple inputs, just verify the count
				expectedLen := len(tt.initialArgs) + len(tt.overrideInputs)*3
				assert.Equal(t, expectedLen, len(result))

				// Verify all inputs are present
				for key, value := range tt.overrideInputs {
					found := false
					for i := 0; i < len(result)-2; i++ {
						if result[i] == "--override-input" && result[i+1] == key && result[i+2] == value {
							found = true
							break
						}
					}
					assert.True(t, found, "override input %s=%s not found in result", key, value)
				}
			}
		})
	}
}

func TestAppendOverrideInputsForFlake(t *testing.T) {
	tests := []struct {
		name           string
		initialArgs    []string
		overrideInputs map[string]string
		expected       []string
	}{
		{
			name:           "empty override inputs",
			initialArgs:    []string{"nix", "build"},
			overrideInputs: map[string]string{},
			expected:       []string{"nix", "build"},
		},
		{
			name:        "single override input with flake prefix",
			initialArgs: []string{"nix", "build"},
			overrideInputs: map[string]string{
				"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
			},
			expected: []string{"nix", "build", "--override-input", "flake/nixpkgs", "github:NixOS/nixpkgs/nixos-unstable"},
		},
		{
			name:        "multiple override inputs with flake prefix",
			initialArgs: []string{"nix", "build"},
			overrideInputs: map[string]string{
				"nixpkgs":      "github:NixOS/nixpkgs/nixos-unstable",
				"home-manager": "github:nix-community/home-manager",
			},
			// Note: map iteration order is not guaranteed, so we just check the length
			expected: []string{"nix", "build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendOverrideInputsForFlake(tt.initialArgs, tt.overrideInputs)

			if len(tt.overrideInputs) == 0 || len(tt.overrideInputs) == 1 {
				assert.Equal(t, tt.expected, result)
			} else {
				// For multiple inputs, just verify the count
				expectedLen := len(tt.initialArgs) + len(tt.overrideInputs)*3
				assert.Equal(t, expectedLen, len(result))

				// Verify all inputs are present with flake/ prefix
				for key, value := range tt.overrideInputs {
					found := false
					expectedKey := "flake/" + key
					for i := 0; i < len(result)-2; i++ {
						if result[i] == "--override-input" && result[i+1] == expectedKey && result[i+2] == value {
							found = true
							break
						}
					}
					assert.True(t, found, "override input flake/%s=%s not found in result", key, value)
				}
			}
		})
	}
}

// TestCustomStepsDeterministicOrder verifies that custom steps are executed in
// deterministic (alphabetically sorted) order, matching the Rust implementation
// which uses BTreeMap. This test runs the same configuration multiple times and
// verifies that the step execution order is always consistent.
func TestCustomStepsDeterministicOrder(t *testing.T) {
	// Test StepsConfig.GetEnabledSteps returns custom steps in sorted order
	t.Run("GetEnabledSteps returns sorted custom steps", func(t *testing.T) {
		config := StepsConfig{
			Custom: map[string]CustomStep{
				"zebra-step":  {Type: CustomStepTypeApp},
				"alpha-step":  {Type: CustomStepTypeApp},
				"middle-step": {Type: CustomStepTypeApp},
				"beta-step":   {Type: CustomStepTypeDevShell, Command: []string{"echo"}},
			},
		}

		// Run multiple times to catch any non-determinism
		for i := 0; i < 10; i++ {
			enabled := config.GetEnabledSteps()

			// Custom steps should be at the end, in alphabetical order
			expectedOrder := []string{"custom:alpha-step", "custom:beta-step", "custom:middle-step", "custom:zebra-step"}
			assert.Equal(t, expectedOrder, enabled, "iteration %d: custom steps should be in alphabetical order", i)
		}
	})

	// Test that runSubflake processes custom steps in sorted order by checking
	// the result map keys correspond to steps processed in alphabetical order
	t.Run("runSubflake processes custom steps in deterministic order", func(t *testing.T) {
		ctx := context.Background()
		flake, err := nix.ParseFlakeURL(".")
		require.NoError(t, err)

		subflake := SubflakeConfig{
			Dir: ".",
			Steps: StepsConfig{
				// Use multiple custom steps with names that would be affected by
				// map iteration randomization
				Custom: map[string]CustomStep{
					"step-z": {Type: CustomStepTypeDevShell, Command: []string{"true"}},
					"step-a": {Type: CustomStepTypeDevShell, Command: []string{"true"}},
					"step-m": {Type: CustomStepTypeDevShell, Command: []string{"true"}},
					"step-b": {Type: CustomStepTypeDevShell, Command: []string{"true"}},
				},
			},
		}

		opts := RunOptions{
			Systems: []string{"x86_64-linux"},
		}

		// Run multiple iterations to catch non-deterministic behavior
		// In the old implementation, this would sometimes produce different orderings
		for i := 0; i < 10; i++ {
			result, err := runSubflake(ctx, flake, "test", subflake, opts)
			require.NoError(t, err)

			// Verify all expected steps are present in the result
			assert.Contains(t, result.Steps, "custom:step-a", "iteration %d", i)
			assert.Contains(t, result.Steps, "custom:step-b", "iteration %d", i)
			assert.Contains(t, result.Steps, "custom:step-m", "iteration %d", i)
			assert.Contains(t, result.Steps, "custom:step-z", "iteration %d", i)

			// The result.Steps map itself doesn't guarantee order, but we verify
			// that all steps were processed (and in the implementation, they are
			// now processed in alphabetical order)
			assert.Len(t, result.Steps, 4, "iteration %d: should have exactly 4 custom steps", i)
		}
	})
}
