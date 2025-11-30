package nix

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testModifiedValue = "modified"

func TestGlobalOptions_ToArgs(t *testing.T) {
	tests := []struct {
		name string
		opts *GlobalOptions
		want []string
	}{
		{
			name: "nil options",
			opts: nil,
			want: []string{},
		},
		{
			name: "empty options",
			opts: &GlobalOptions{},
			want: []string{},
		},
		{
			name: "accept-flake-config only",
			opts: &GlobalOptions{
				AcceptFlakeConfig: true,
			},
			want: []string{"--accept-flake-config"},
		},
		{
			name: "extra args only",
			opts: &GlobalOptions{
				ExtraArgs: []string{"--verbose", "--option", "foo", "bar"},
			},
			want: []string{"--verbose", "--option", "foo", "bar"},
		},
		{
			name: "all options",
			opts: &GlobalOptions{
				AcceptFlakeConfig: true,
				ExtraArgs:         []string{"--verbose"},
			},
			want: []string{"--accept-flake-config", "--verbose"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.ToArgs()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetGetGlobalOptions(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Set options
	opts := &GlobalOptions{
		AcceptFlakeConfig: true,
		ExtraArgs:         []string{"--verbose"},
	}
	SetGlobalOptions(opts)

	// Get options
	got := GetGlobalOptions()
	assert.Equal(t, opts.AcceptFlakeConfig, got.AcceptFlakeConfig)
	assert.Equal(t, opts.ExtraArgs, got.ExtraArgs)
}

func TestNewCmdWithOptions(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Test with accept-flake-config
	SetGlobalOptions(&GlobalOptions{
		AcceptFlakeConfig: true,
	})

	cmd := NewCmdWithOptions()
	assert.Contains(t, cmd.ExtraArgs, "--accept-flake-config")
}

func TestNewCmd_UsesGlobalOptions(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Set global options
	SetGlobalOptions(&GlobalOptions{
		AcceptFlakeConfig: true,
		ExtraArgs:         []string{"--verbose"},
	})

	// NewCmd should use global options
	cmd := NewCmd()
	assert.Contains(t, cmd.ExtraArgs, "--accept-flake-config")
	assert.Contains(t, cmd.ExtraArgs, "--verbose")
}

func TestGlobalOptions_ResetToEmpty(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Reset to empty
	SetGlobalOptions(&GlobalOptions{})

	opts := GetGlobalOptions()
	assert.False(t, opts.AcceptFlakeConfig)
	assert.Empty(t, opts.ExtraArgs)

	// NewCmd should have no extra args
	cmd := NewCmd()
	assert.Empty(t, cmd.ExtraArgs)
}

func TestGlobalOptions_Concurrency(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Simulate concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(val int) {
			opts := &GlobalOptions{
				AcceptFlakeConfig: val%2 == 0,
				ExtraArgs:         []string{fmt.Sprintf("arg-%d", val)},
			}
			SetGlobalOptions(opts)
			got := GetGlobalOptions()
			args := got.ToArgs()

			// Verify returned args are valid
			assert.NotNil(t, args, "ToArgs should not return nil")

			// Verify deep copy by modifying returned value
			if len(got.ExtraArgs) > 0 {
				got.ExtraArgs[0] = testModifiedValue
				// Modification should not affect global state
				newGot := GetGlobalOptions()
				if len(newGot.ExtraArgs) > 0 {
					assert.NotEqual(t, testModifiedValue, newGot.ExtraArgs[0], "deep copy should prevent modifications")
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSetGlobalOptions_NilHandling(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Set options first
	SetGlobalOptions(&GlobalOptions{
		AcceptFlakeConfig: true,
		ExtraArgs:         []string{"--verbose"},
	})

	// Now set nil - should reset to empty options, not panic
	SetGlobalOptions(nil)

	opts := GetGlobalOptions()
	assert.False(t, opts.AcceptFlakeConfig)
	assert.Empty(t, opts.ExtraArgs)
}

func TestGetGlobalOptions_DeepCopy(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Set options with ExtraArgs
	original := &GlobalOptions{
		AcceptFlakeConfig: true,
		ExtraArgs:         []string{"--verbose", "--option", "foo"},
	}
	SetGlobalOptions(original)

	// Get a copy
	copied := GetGlobalOptions()

	// Modify the copy's ExtraArgs
	copied.ExtraArgs[0] = testModifiedValue

	// Original should be unchanged (deep copy verification)
	got := GetGlobalOptions()
	assert.Equal(t, "--verbose", got.ExtraArgs[0], "original should be unchanged after modifying copy")
}

func TestSetGlobalOptions_DoesNotRetainReference(t *testing.T) {
	// Reset global options after test
	defer SetGlobalOptions(&GlobalOptions{})

	// Create options and set them
	opts := &GlobalOptions{
		AcceptFlakeConfig: true,
		ExtraArgs:         []string{"--verbose"},
	}
	SetGlobalOptions(opts)

	// Modify the original after setting (simulating caller mutation)
	opts.AcceptFlakeConfig = false
	opts.ExtraArgs[0] = testModifiedValue

	// Global state should be unchanged (deep copy verification)
	got := GetGlobalOptions()
	assert.True(t, got.AcceptFlakeConfig, "modifying input after Set should not affect global state")
	assert.Equal(t, "--verbose", got.ExtraArgs[0], "modifying input slice should not affect global state")
}
