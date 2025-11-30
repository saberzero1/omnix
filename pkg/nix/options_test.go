package nix

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestGlobalOptions_DefaultEmpty(t *testing.T) {
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
			_ = got.ToArgs() // Use the result
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
	copied.ExtraArgs[0] = "modified"

	// Original should be unchanged (deep copy verification)
	got := GetGlobalOptions()
	assert.Equal(t, "--verbose", got.ExtraArgs[0], "original should be unchanged after modifying copy")
}
