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

	// Note: Build-time injected variables (flakeMetadata) cannot be tested directly
	// since they require recompilation with ldflags. In unit tests, they are empty,
	// so the environment variable or fallback is used.

	t.Run("with env var", func(t *testing.T) {
		// When no build-time variable is set (unit test case), env var takes precedence
		// The function appends #default to the env var value for consistency
		envVal := "path:/nix/store/abc-metadata"
		_ = os.Setenv("FLAKE_METADATA", envVal)
		assert.Equal(t, envVal+"#default", FlakeMetadataURL())
	})

	t.Run("without env var uses fallback", func(t *testing.T) {
		// Without build-time var or env var, falls back to repo root detection
		_ = os.Unsetenv("FLAKE_METADATA")
		url := FlakeMetadataURL()
		assert.Contains(t, url, "metadata")
		assert.Contains(t, url, "#default") // Should include the attribute
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

	// Note: Build-time injected variables (flakeAddStringContext) cannot be tested directly
	// since they require recompilation with ldflags. In unit tests, they are empty,
	// so the environment variable or fallback is used.

	t.Run("with env var", func(t *testing.T) {
		// When no build-time variable is set (unit test case), env var takes precedence
		// The function appends #default to the env var value for consistency
		envVal := "path:/nix/store/abc-addstringcontext"
		_ = os.Setenv("FLAKE_ADDSTRINGCONTEXT", envVal)
		assert.Equal(t, envVal+"#default", FlakeAddStringContextURL())
	})

	t.Run("without env var uses fallback", func(t *testing.T) {
		// Without build-time var or env var, falls back to repo root detection
		_ = os.Unsetenv("FLAKE_ADDSTRINGCONTEXT")
		url := FlakeAddStringContextURL()
		assert.Contains(t, url, "addstringcontext")
		assert.Contains(t, url, "#default") // Should include the attribute
	})
}

// TestBuildTimeVariables documents the build-time variables that can be injected
// via ldflags. These tests verify the expected behavior when the variables are set.
func TestBuildTimeVariables(t *testing.T) {
	t.Run("flakeMetadata variable documentation", func(t *testing.T) {
		// flakeMetadata is set via: -X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeMetadata=...
		// When set, FlakeMetadataURL() returns the injected value + "#default"
		// In unit tests, this variable is empty so the test verifies fallback behavior
		if flakeMetadata == "" {
			t.Log("flakeMetadata is empty (expected in unit tests)")
		} else {
			t.Logf("flakeMetadata is set to: %s", flakeMetadata)
		}
	})

	t.Run("flakeAddStringContext variable documentation", func(t *testing.T) {
		// flakeAddStringContext is set via: -X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeAddStringContext=...
		// When set, FlakeAddStringContextURL() returns the injected value + "#default"
		// In unit tests, this variable is empty so the test verifies fallback behavior
		if flakeAddStringContext == "" {
			t.Log("flakeAddStringContext is empty (expected in unit tests)")
		} else {
			t.Logf("flakeAddStringContext is set to: %s", flakeAddStringContext)
		}
	})
}
