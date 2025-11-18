package ci

import (
	"context"
	"testing"

	"github.com/juspay/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRunBuildStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:juspay/omnix/main")
	require.NoError(t, err)

	step := BuildStep{
		Enable: true,
		Impure: false,
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	result := runBuildStep(ctx, flake, step, opts)
	assert.Equal(t, "build", result.Name)
	// Note: This may fail or succeed depending on the system, just testing it runs
}

func TestRunLockfileStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:juspay/omnix/main")
	require.NoError(t, err)

	step := LockfileStep{
		Enable: true,
	}

	result := runLockfileStep(ctx, flake, step)
	assert.Equal(t, "lockfile", result.Name)
}

func TestRunFlakeCheckStep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL("github:juspay/omnix/main")
	require.NoError(t, err)

	step := FlakeCheckStep{
		Enable: true,
	}

	result := runFlakeCheckStep(ctx, flake, step)
	assert.Equal(t, "flakeCheck", result.Name)
}

func TestRunCustomStep(t *testing.T) {
	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	tests := []struct {
		name          string
		step          CustomStep
		expectedName  string
		expectedError bool
		errorContains string
	}{
		{
			name: "empty command",
			step: CustomStep{
				Name:    "test",
				Command: []string{},
				Enable:  true,
			},
			expectedName:  "custom:test",
			expectedError: true,
			errorContains: "no command",
		},
		{
			name: "echo command",
			step: CustomStep{
				Name:    "echo-test",
				Command: []string{"echo", "hello"},
				Enable:  true,
			},
			expectedName:  "custom:echo-test",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCustomStep(ctx, flake, tt.step)
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

	config := Config{
		Default: map[string]SubflakeConfig{
			".": {
				Dir:  ".",
				Skip: false,
				Steps: StepsConfig{
					Build:      BuildStep{Enable: false},
					Lockfile:   LockfileStep{Enable: false},
					FlakeCheck: FlakeCheckStep{Enable: false},
				},
			},
		},
	}

	opts := RunOptions{
		Systems: []string{"x86_64-linux"},
	}

	results, err := Run(ctx, flake, config, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}
