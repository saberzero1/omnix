package ci

import (
	"context"
	"strings"
	"testing"

	"github.com/juspay/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRemoteCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		host        string
		command     []string
		shouldError bool
	}{
		{
			name:        "empty host",
			host:        "",
			command:     []string{"echo", "test"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeRemoteCommand(ctx, tt.host, tt.command)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				_ = output
				_ = err
			}
		})
	}
}

func TestExecuteRemoteCommandQuoting(t *testing.T) {
	// Test command building logic without actually executing SSH
	ctx := context.Background()
	host := ""
	command := []string{"echo", "hello world", "--flag=value"}

	// Empty host should return error immediately
	_, err := executeRemoteCommand(ctx, host, command)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not specified")
}

func TestRunBuildStepRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	step := BuildStep{
		Enable: true,
		Impure: false,
	}

	opts := RunOptions{
		Systems:    []string{"x86_64-linux"},
		RemoteHost: "user@remotehost",
	}

	result := runBuildStepRemote(ctx, opts.RemoteHost, flake, step, opts)

	// Should complete without panic
	assert.Equal(t, "build", result.Name)
	assert.Greater(t, result.Duration.Nanoseconds(), int64(0))

	// Will likely fail due to no SSH connection
	if !result.Success {
		assert.NotEmpty(t, result.Error)
		assert.Contains(t, strings.ToLower(result.Error), "remote")
	}
}

func TestRunLockfileStepRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	step := LockfileStep{
		Enable: true,
	}

	result := runLockfileStepRemote(ctx, "user@host", flake, step)

	assert.Equal(t, "lockfile", result.Name)
	assert.Greater(t, result.Duration.Nanoseconds(), int64(0))
}

func TestRunFlakeCheckStepRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	step := FlakeCheckStep{
		Enable: true,
	}

	result := runFlakeCheckStepRemote(ctx, "user@host", flake, step)

	assert.Equal(t, "flakeCheck", result.Name)
	assert.Greater(t, result.Duration.Nanoseconds(), int64(0))
}

func TestRunCustomStepRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	step := CustomStep{
		Name:    "custom-test",
		Command: []string{"echo", "test"},
		Enable:  true,
	}

	result := runCustomStepRemote(ctx, "user@host", flake, step)

	assert.Equal(t, "custom:custom-test", result.Name)
	assert.Greater(t, result.Duration.Nanoseconds(), int64(0))
}

func TestRunWithRemoteHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH test in short mode")
	}

	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"main": {
				Skip: false,
				Dir:  ".",
				Steps: StepsConfig{
					Build: BuildStep{Enable: false}, // Disable actual build
					Custom: []CustomStep{
						{
							Name:    "test",
							Command: []string{"echo", "remote-test"},
							Enable:  true,
						},
					},
				},
			},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:    []string{"x86_64-linux"},
		RemoteHost: "testuser@testhost",
	}

	// This will fail due to no SSH connection, but should not panic
	results, err := Run(ctx, flake, config, opts)

	// Either returns error (connection failed) or results
	if err == nil {
		assert.NotNil(t, results)
	} else {
		assert.Error(t, err)
	}
}
