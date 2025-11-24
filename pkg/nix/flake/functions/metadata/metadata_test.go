package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlakeMetadataFn_FlakeURL(t *testing.T) {
	fn := FlakeMetadataFn{}
	url := fn.FlakeURL()
	assert.NotEmpty(t, url)
	// Should contain "metadata" in some form
	assert.Contains(t, url, "metadata")
}

func TestFlakeMetadataFn_Init(t *testing.T) {
	fn := FlakeMetadataFn{}
	out := &Output{
		Flake: "/nix/store/abc-flake",
	}
	// Init should not panic and should not modify the output
	fn.Init(out)
	assert.Equal(t, "/nix/store/abc-flake", out.Flake)
}

func TestOutput_FindInput(t *testing.T) {
	tests := []struct {
		name      string
		output    Output
		inputName string
		wantErr   bool
		wantInput *FlakeInput
	}{
		{
			name: "input found",
			output: Output{
				Inputs: []FlakeInput{
					{Name: "nixpkgs", Path: "/nix/store/abc-nixpkgs"},
					{Name: "flake-utils", Path: "/nix/store/def-flake-utils"},
				},
			},
			inputName: "nixpkgs",
			wantInput: &FlakeInput{Name: "nixpkgs", Path: "/nix/store/abc-nixpkgs"},
		},
		{
			name: "input not found",
			output: Output{
				Inputs: []FlakeInput{
					{Name: "nixpkgs", Path: "/nix/store/abc-nixpkgs"},
				},
			},
			inputName: "nonexistent",
			wantErr:   true,
		},
		{
			name:      "no inputs available",
			output:    Output{},
			inputName: "nixpkgs",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := tt.output.FindInput(tt.inputName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantInput, input)
			}
		})
	}
}

func TestOutput_InputPaths(t *testing.T) {
	tests := []struct {
		name     string
		output   Output
		expected map[string]string
	}{
		{
			name: "with inputs",
			output: Output{
				Inputs: []FlakeInput{
					{Name: "nixpkgs", Path: "/nix/store/abc-nixpkgs"},
					{Name: "flake-utils", Path: "/nix/store/def-flake-utils"},
				},
			},
			expected: map[string]string{
				"nixpkgs":     "/nix/store/abc-nixpkgs",
				"flake-utils": "/nix/store/def-flake-utils",
			},
		},
		{
			name:     "no inputs",
			output:   Output{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.InputPaths()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOutput_AllPaths(t *testing.T) {
	output := Output{
		Flake: "/nix/store/abc-flake",
		Inputs: []FlakeInput{
			{Name: "nixpkgs", Path: "/nix/store/def-nixpkgs"},
			{Name: "flake-utils", Path: "/nix/store/ghi-flake-utils"},
		},
	}

	paths := output.AllPaths()
	assert.Len(t, paths, 3)
	assert.Equal(t, "/nix/store/abc-flake", paths[0])
	assert.Contains(t, paths, "/nix/store/def-nixpkgs")
	assert.Contains(t, paths, "/nix/store/ghi-flake-utils")
}
