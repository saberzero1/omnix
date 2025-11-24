package functions

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToOverrideInputs(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "string values",
			input: struct {
				Flake  string `json:"flake"`
				System string `json:"system"`
			}{
				Flake:  "github:nixos/nixpkgs",
				System: "x86_64-linux",
			},
			expected: map[string]string{
				"flake":  "github:nixos/nixpkgs",
				"system": "x86_64-linux",
			},
		},
		{
			name: "boolean true",
			input: struct {
				IncludeInputs bool `json:"include-inputs"`
			}{
				IncludeInputs: true,
			},
			expected: map[string]string{
				"include-inputs": TrueFlakeURL(),
			},
		},
		{
			name: "boolean false",
			input: struct {
				IncludeInputs bool `json:"include-inputs"`
			}{
				IncludeInputs: false,
			},
			expected: map[string]string{
				"include-inputs": FalseFlakeURL(),
			},
		},
		{
			name: "mixed types",
			input: struct {
				Flake         string `json:"flake"`
				IncludeInputs bool   `json:"include-inputs"`
				System        string `json:"system,omitempty"`
			}{
				Flake:         "github:nixos/nixpkgs",
				IncludeInputs: true,
			},
			expected: map[string]string{
				"flake":          "github:nixos/nixpkgs",
				"include-inputs": TrueFlakeURL(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toOverrideInputs(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTransformOverrideInputs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty args",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no override-input",
			input:    []string{"--impure", "--no-link"},
			expected: []string{"--impure", "--no-link"},
		},
		{
			name:     "single override-input",
			input:    []string{"--override-input", "nixpkgs", "github:nixos/nixpkgs"},
			expected: []string{"--override-input", "flake/nixpkgs", "github:nixos/nixpkgs"},
		},
		{
			name: "multiple override-inputs",
			input: []string{
				"--override-input", "nixpkgs", "github:nixos/nixpkgs",
				"--override-input", "flake-utils", "github:numtide/flake-utils",
			},
			expected: []string{
				"--override-input", "flake/nixpkgs", "github:nixos/nixpkgs",
				"--override-input", "flake/flake-utils", "github:numtide/flake-utils",
			},
		},
		{
			name: "mixed args",
			input: []string{
				"--impure",
				"--override-input", "nixpkgs", "github:nixos/nixpkgs",
				"--no-link",
			},
			expected: []string{
				"--impure",
				"--override-input", "flake/nixpkgs", "github:nixos/nixpkgs",
				"--no-link",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformOverrideInputs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrueFlakeURL(t *testing.T) {
	// Save original env var
	original := os.Getenv("TRUE_FLAKE")
	defer func() {
		if original != "" {
			_ = os.Setenv("TRUE_FLAKE", original)
		} else {
			_ = os.Unsetenv("TRUE_FLAKE")
		}
	}()

	t.Run("with env var", func(t *testing.T) {
		expected := "path:/custom/true/flake"
		_ = os.Setenv("TRUE_FLAKE", expected)
		assert.Equal(t, expected, TrueFlakeURL())
	})

	t.Run("without env var", func(t *testing.T) {
		_ = os.Unsetenv("TRUE_FLAKE")
		url := TrueFlakeURL()
		assert.Contains(t, url, "github:boolean-option/true")
	})
}

func TestFalseFlakeURL(t *testing.T) {
	// Save original env var
	original := os.Getenv("FALSE_FLAKE")
	defer func() {
		if original != "" {
			_ = os.Setenv("FALSE_FLAKE", original)
		} else {
			_ = os.Unsetenv("FALSE_FLAKE")
		}
	}()

	t.Run("with env var", func(t *testing.T) {
		expected := "path:/custom/false/flake"
		_ = os.Setenv("FALSE_FLAKE", expected)
		assert.Equal(t, expected, FalseFlakeURL())
	})

	t.Run("without env var", func(t *testing.T) {
		_ = os.Unsetenv("FALSE_FLAKE")
		url := FalseFlakeURL()
		assert.Contains(t, url, "github:boolean-option/false")
	})
}

func TestFlakeMetadataURL(t *testing.T) {
	// Save original env var
	original := os.Getenv("FLAKE_METADATA")
	defer func() {
		if original != "" {
			_ = os.Setenv("FLAKE_METADATA", original)
		} else {
			_ = os.Unsetenv("FLAKE_METADATA")
		}
	}()

	t.Run("with env var", func(t *testing.T) {
		expected := "path:/nix/store/abc-metadata"
		_ = os.Setenv("FLAKE_METADATA", expected)
		assert.Equal(t, expected, FlakeMetadataURL())
	})

	t.Run("without env var", func(t *testing.T) {
		_ = os.Unsetenv("FLAKE_METADATA")
		url := FlakeMetadataURL()
		assert.Contains(t, url, "metadata")
	})
}

func TestFlakeAddStringContextURL(t *testing.T) {
	// Save original env var
	original := os.Getenv("FLAKE_ADDSTRINGCONTEXT")
	defer func() {
		if original != "" {
			_ = os.Setenv("FLAKE_ADDSTRINGCONTEXT", original)
		} else {
			_ = os.Unsetenv("FLAKE_ADDSTRINGCONTEXT")
		}
	}()

	t.Run("with env var", func(t *testing.T) {
		expected := "path:/nix/store/abc-addstringcontext"
		_ = os.Setenv("FLAKE_ADDSTRINGCONTEXT", expected)
		assert.Equal(t, expected, FlakeAddStringContextURL())
	})

	t.Run("without env var", func(t *testing.T) {
		_ = os.Unsetenv("FLAKE_ADDSTRINGCONTEXT")
		url := FlakeAddStringContextURL()
		assert.Contains(t, url, "addstringcontext")
	})
}
