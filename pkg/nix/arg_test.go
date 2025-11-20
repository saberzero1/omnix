package nix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewArgs(t *testing.T) {
	args := NewArgs()
	assert.NotNil(t, args)
	assert.Empty(t, args.ExtraExperimentalFeatures)
	assert.Empty(t, args.ExtraAccessTokens)
	assert.Empty(t, args.ExtraNixArgs)
}

func TestArgs_ToArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        *Args
		subcommands []string
		want        []string
	}{
		{
			name: "empty args",
			args: NewArgs(),
			want: []string{},
		},
		{
			name: "experimental features only",
			args: &Args{
				ExtraExperimentalFeatures: []string{"flakes", "nix-command"},
			},
			want: []string{"--extra-experimental-features", "flakes nix-command"},
		},
		{
			name: "access tokens only",
			args: &Args{
				ExtraAccessTokens: []string{"github.com=token123"},
			},
			want: []string{"--extra-access-tokens", "github.com=token123"},
		},
		{
			name: "extra nix args only",
			args: &Args{
				ExtraNixArgs: []string{"--verbose", "--option", "foo", "bar"},
			},
			want: []string{"--verbose", "--option", "foo", "bar"},
		},
		{
			name: "all arguments",
			args: &Args{
				ExtraExperimentalFeatures: []string{"flakes"},
				ExtraAccessTokens:         []string{"token1"},
				ExtraNixArgs:              []string{"--verbose"},
			},
			want: []string{
				"--extra-experimental-features", "flakes",
				"--extra-access-tokens", "token1",
				"--verbose",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.ToArgs(tt.subcommands...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestArgs_WithFlakes(t *testing.T) {
	args := NewArgs()
	result := args.WithFlakes()
	
	assert.Same(t, args, result, "WithFlakes should return the same instance")
	assert.Contains(t, args.ExtraExperimentalFeatures, "nix-command")
	assert.Contains(t, args.ExtraExperimentalFeatures, "flakes")
}

func TestArgs_WithNixCommand(t *testing.T) {
	args := NewArgs()
	result := args.WithNixCommand()
	
	assert.Same(t, args, result, "WithNixCommand should return the same instance")
	assert.Contains(t, args.ExtraExperimentalFeatures, "nix-command")
}

func TestRemoveNonsenseArgs(t *testing.T) {
	tests := []struct {
		name        string
		subcommands []string
		input       []string
		want        []string
	}{
		{
			name:        "no filtering for unknown subcommand",
			subcommands: []string{"show"},
			input:       []string{"--rebuild", "--verbose"},
			want:        []string{"--rebuild", "--verbose"},
		},
		{
			name:        "remove --rebuild for eval",
			subcommands: []string{"eval"},
			input:       []string{"--rebuild", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove --override-input for eval",
			subcommands: []string{"eval"},
			input:       []string{"--override-input", "foo", "bar", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove both for eval",
			subcommands: []string{"eval"},
			input:       []string{"--rebuild", "--override-input", "foo", "bar", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove --rebuild for flake check",
			subcommands: []string{"flake", "check"},
			input:       []string{"--rebuild", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove --rebuild for develop",
			subcommands: []string{"develop"},
			input:       []string{"--rebuild", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove --rebuild for run",
			subcommands: []string{"run"},
			input:       []string{"--rebuild", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "remove both for flake lock",
			subcommands: []string{"flake", "lock"},
			input:       []string{"--rebuild", "--override-input", "nixpkgs", "path", "--verbose"},
			want:        []string{"--verbose"},
		},
		{
			name:        "multiple --override-input",
			subcommands: []string{"eval"},
			input:       []string{"--override-input", "a", "b", "--override-input", "c", "d", "--verbose"},
			want:        []string{"--verbose"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test case
			input := make([]string, len(tt.input))
			copy(input, tt.input)
			
			removeNonsenseArgs(tt.subcommands, &input)
			assert.Equal(t, tt.want, input)
		})
	}
}

func TestRemoveArgument(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		arg      string
		argCount int
		want     []string
	}{
		{
			name:     "remove flag with no args",
			args:     []string{"--rebuild", "--verbose"},
			arg:      "--rebuild",
			argCount: 0,
			want:     []string{"--verbose"},
		},
		{
			name:     "remove flag with 2 args",
			args:     []string{"--override-input", "foo", "bar", "--verbose"},
			arg:      "--override-input",
			argCount: 2,
			want:     []string{"--verbose"},
		},
		{
			name:     "remove multiple occurrences",
			args:     []string{"--rebuild", "--verbose", "--rebuild"},
			arg:      "--rebuild",
			argCount: 0,
			want:     []string{"--verbose"},
		},
		{
			name:     "arg not present",
			args:     []string{"--verbose"},
			arg:      "--rebuild",
			argCount: 0,
			want:     []string{"--verbose"},
		},
		{
			name:     "empty list",
			args:     []string{},
			arg:      "--rebuild",
			argCount: 0,
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			
			removeArgument(&args, tt.arg, tt.argCount)
			assert.Equal(t, tt.want, args)
		})
	}
}
