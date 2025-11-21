package ci

import (
	"context"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomStepExecution tests that custom steps are executed correctly
func TestCustomStepExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create a config with custom steps
	config := Config{
		Default: map[string]SubflakeConfig{
			"test": {
				Dir:  ".",
				Skip: false,
				Steps: StepsConfig{
					Build:      BuildStep{Enable: false}, // Disable build to speed up test
					Lockfile:   LockfileStep{Enable: false},
					FlakeCheck: FlakeCheckStep{Enable: false},
					Custom: map[string]CustomStep{
						// Test app step - om show is available in the current flake
						"show-test": {
							Type: CustomStepTypeApp,
							Name: "default", // Use default app (om)
							Args: []string{"--version"},
						},
					},
				},
			},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	// Get current system
	info, err := nix.GetInfo(ctx)
	require.NoError(t, err)

	opts := RunOptions{
		Systems: []string{info.Config.System.Value},
	}

	results, err := Run(ctx, flake, config, opts)
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "test", result.Subflake)

	// Check that the custom step ran
	assert.Contains(t, result.Steps, "custom:show-test")
	stepResult := result.Steps["custom:show-test"]

	// The step should succeed
	assert.True(t, stepResult.Success, "custom step should succeed")
	assert.Empty(t, stepResult.Error, "custom step should have no error")
}
