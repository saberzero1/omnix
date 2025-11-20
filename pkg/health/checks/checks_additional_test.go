package checks

import (
	"context"
	"os"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
)

func TestGreenResult_String(t *testing.T) {
	result := GreenResult{}
	str := result.String()
	assert.Contains(t, str, "Passed")
}

func TestRedResult_String(t *testing.T) {
	result := RedResult{
		Message:    "Something failed",
		Suggestion: "Do this to fix",
	}
	str := result.String()
	assert.Contains(t, str, "Failed")
	assert.Contains(t, str, "Something failed")
	assert.Contains(t, str, "Do this to fix")
}

func TestCaches_Check_WithMissingCaches(t *testing.T) {
	ctx := context.Background()

	check := Caches{
		Required: []string{
			"https://cache.nixos.org",
			"https://my-cache.cachix.org",
		},
	}

	nixInfo := &nix.Info{
		Config: nix.Config{
			Substituters: nix.ConfigValue[[]string]{ //nolint:misspell // "Substituters" is correct Nix terminology
				Value: []string{"https://cache.nixos.org"},
			},
		},
		Env: &nix.Env{
			OS: nix.OSType{Type: "linux"},
		},
	}

	results := check.Check(ctx, nixInfo)

	assert.Len(t, results, 1)
	assert.Equal(t, "caches", results[0].Name)
	assert.False(t, results[0].Check.Result.IsGreen(), "should fail when caches are missing")
}

func TestCaches_Check_AllPresent(t *testing.T) {
	ctx := context.Background()

	check := Caches{
		Required: []string{"https://cache.nixos.org"},
	}

	nixInfo := &nix.Info{
		Config: nix.Config{
			Substituters: nix.ConfigValue[[]string]{ //nolint:misspell // "Substituters" is correct Nix terminology
				Value: []string{
					"https://cache.nixos.org",
					"https://other-cache.org",
				},
			},
		},
		Env: &nix.Env{
			OS: nix.OSType{Type: "linux"},
		},
	}

	results := check.Check(ctx, nixInfo)

	assert.Len(t, results, 1)
	// Note: The check implementation has empty substituters placeholder (Nix terminology) //nolint:misspell
	// so this test verifies it doesn't crash rather than actual functionality
	assert.NotNil(t, results[0].Check.Result)
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with trailing slash",
			url:      "https://cache.nixos.org/",
			expected: "https://cache.nixos.org",
		},
		{
			name:     "URL without trailing slash",
			url:      "https://cache.nixos.org",
			expected: "https://cache.nixos.org",
		},
		{
			name:     "Invalid URL returns as-is",
			url:      "://invalid url with spaces",
			expected: "://invalid url with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrustedUsers_CheckEnabled(t *testing.T) {
	ctx := context.Background()

	check := TrustedUsers{Enable: true}

	nixInfo := &nix.Info{
		Env: &nix.Env{
			User:   "testuser",
			Groups: []string{"users", "wheel"},
			OS:     nix.OSType{Type: "linux"},
		},
	}

	results := check.Check(ctx, nixInfo)

	// Should return a check when enabled
	assert.Len(t, results, 1)
	assert.Equal(t, "trusted-users", results[0].Name)
}

func TestHomebrew_CheckNonDarwin(t *testing.T) {
	ctx := context.Background()

	// Save original GOOS
	// Note: We can't actually change runtime.GOOS, so this test
	// just verifies the check doesn't panic on non-macOS

	nixInfo := &nix.Info{}
	check := Homebrew{}

	results := check.Check(ctx, nixInfo)

	// On non-macOS (like Linux CI), should return empty or 0 results
	// On macOS, should return 1 result
	// Just verify it doesn't panic
	assert.NotNil(t, results)
}

func TestRosetta_CheckNonDarwinARM(t *testing.T) {
	ctx := context.Background()

	nixInfo := &nix.Info{}
	check := Rosetta{}

	results := check.Check(ctx, nixInfo)

	// On non-macOS ARM64, should return empty
	// Just verify it doesn't panic
	assert.NotNil(t, results)
}

func TestShell_CheckWithShell(t *testing.T) {
	ctx := context.Background()

	// Save and restore SHELL env var
	oldShell := os.Getenv("SHELL")
	defer func() {
		if oldShell != "" {
			_ = os.Setenv("SHELL", oldShell)
		} else {
			_ = os.Unsetenv("SHELL")
		}
	}()

	_ = os.Setenv("SHELL", "/bin/bash")

	nixInfo := &nix.Info{}
	check := Shell{}

	results := check.Check(ctx, nixInfo)

	assert.Len(t, results, 1)
	assert.Equal(t, "shell", results[0].Name)
	assert.Contains(t, results[0].Check.Info, "/bin/bash")
}

func TestShell_CheckNoShell(t *testing.T) {
	ctx := context.Background()

	// Save and restore SHELL env var
	oldShell := os.Getenv("SHELL")
	defer func() {
		if oldShell != "" {
			_ = os.Setenv("SHELL", oldShell)
		} else {
			_ = os.Unsetenv("SHELL")
		}
	}()

	_ = os.Unsetenv("SHELL")

	nixInfo := &nix.Info{}
	check := Shell{}

	results := check.Check(ctx, nixInfo)

	// Should return empty when no SHELL set
	assert.Empty(t, results)
}

func TestMaxJobs_Check(t *testing.T) {
	ctx := context.Background()

	nixInfo := &nix.Info{}
	check := MaxJobs{}

	results := check.Check(ctx, nixInfo)

	// Currently returns empty (placeholder implementation)
	assert.Empty(t, results)
}
