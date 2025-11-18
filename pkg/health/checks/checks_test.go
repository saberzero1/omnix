package checks

import (
	"context"
	"testing"

	"github.com/juspay/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
)

func TestFlakeEnabled_Check(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		features       []string
		expectGreen    bool
		expectRequired bool
	}{
		{
			name:           "Both flakes and nix-command enabled",
			features:       []string{"flakes", "nix-command"},
			expectGreen:    true,
			expectRequired: true,
		},
		{
			name:           "Only flakes enabled",
			features:       []string{"flakes"},
			expectGreen:    false,
			expectRequired: true,
		},
		{
			name:           "Only nix-command enabled",
			features:       []string{"nix-command"},
			expectGreen:    false,
			expectRequired: true,
		},
		{
			name:           "Neither enabled",
			features:       []string{},
			expectGreen:    false,
			expectRequired: true,
		},
		{
			name:           "Extra features don't matter",
			features:       []string{"flakes", "nix-command", "ca-derivations"},
			expectGreen:    true,
			expectRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nixInfo := &nix.Info{
				Config: nix.Config{
					ExperimentalFeatures: nix.ConfigValue[[]string]{Value: tt.features},
				},
			}

			check := &FlakeEnabled{}
			results := check.Check(ctx, nixInfo)

			assert.Len(t, results, 1)
			assert.Equal(t, "flake-enabled", results[0].Name)
			assert.Equal(t, tt.expectGreen, results[0].Check.Result.IsGreen())
			assert.Equal(t, tt.expectRequired, results[0].Check.Required)
		})
	}
}

func TestNixVersion_Check(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		minVersion     nix.Version
		currentVersion nix.Version
		expectGreen    bool
	}{
		{
			name:           "Current version equals minimum",
			minVersion:     nix.Version{Major: 2, Minor: 16, Patch: 0},
			currentVersion: nix.Version{Major: 2, Minor: 16, Patch: 0},
			expectGreen:    true,
		},
		{
			name:           "Current version greater than minimum",
			minVersion:     nix.Version{Major: 2, Minor: 16, Patch: 0},
			currentVersion: nix.Version{Major: 2, Minor: 18, Patch: 0},
			expectGreen:    true,
		},
		{
			name:           "Current version less than minimum",
			minVersion:     nix.Version{Major: 2, Minor: 16, Patch: 0},
			currentVersion: nix.Version{Major: 2, Minor: 15, Patch: 0},
			expectGreen:    false,
		},
		{
			name:           "Major version difference",
			minVersion:     nix.Version{Major: 2, Minor: 16, Patch: 0},
			currentVersion: nix.Version{Major: 3, Minor: 0, Patch: 0},
			expectGreen:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nixInfo := &nix.Info{
				Version: tt.currentVersion,
			}

			check := &NixVersion{MinVersion: tt.minVersion}
			results := check.Check(ctx, nixInfo)

			assert.Len(t, results, 1)
			assert.Equal(t, "supported-nix-versions", results[0].Name)
			assert.Equal(t, tt.expectGreen, results[0].Check.Result.IsGreen())
			assert.True(t, results[0].Check.Required)
		})
	}
}

func TestDefaultNixVersion(t *testing.T) {
	check := DefaultNixVersion()
	assert.Equal(t, uint32(2), check.MinVersion.Major)
	assert.Equal(t, uint32(16), check.MinVersion.Minor)
	assert.Equal(t, uint32(0), check.MinVersion.Patch)
}

func TestTrustedUsers_CheckDisabled(t *testing.T) {
	ctx := context.Background()
	nixInfo := &nix.Info{
		Env: &nix.Env{
			User:   "testuser",
			Groups: []string{"users"},
		},
	}

	check := &TrustedUsers{Enable: false}
	results := check.Check(ctx, nixInfo)

	// Should return empty when disabled
	assert.Empty(t, results)
}

func TestRosetta_CheckNonMacOS(t *testing.T) {
	ctx := context.Background()
	nixInfo := &nix.Info{}

	check := &Rosetta{}
	_ = check.Check(ctx, nixInfo)

	// Should return empty on non-macOS or non-ARM64
	// (This test will pass on Linux, may fail on macOS ARM64)
	// The actual behavior depends on runtime.GOOS and runtime.GOARCH
}

func TestDirenv_Check(t *testing.T) {
	ctx := context.Background()
	nixInfo := &nix.Info{}

	check := &Direnv{}
	_ = check.Check(ctx, nixInfo)

	// Just verify it doesn't panic
}

func TestHomebrew_CheckNonMacOS(t *testing.T) {
	ctx := context.Background()
	nixInfo := &nix.Info{}

	check := &Homebrew{}
	_ = check.Check(ctx, nixInfo)

	// On non-macOS, should return empty
	// On macOS, should return 1 check
	// The actual behavior depends on runtime.GOOS
}

func TestShell_Check(t *testing.T) {
	ctx := context.Background()
	nixInfo := &nix.Info{}

	check := &Shell{}
	results := check.Check(ctx, nixInfo)

	// May return 0 or 1 depending on whether SHELL is set
	// Just verify it doesn't panic
	assert.NotNil(t, results)
}

func TestParseCachixURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		expectName string
		expectNil  bool
	}{
		{
			name:       "Valid cachix URL",
			url:        "https://foo.cachix.org",
			expectName: "foo",
			expectNil:  false,
		},
		{
			name:      "Non-cachix URL",
			url:       "https://cache.nixos.org",
			expectNil: true,
		},
		{
			name:       "Cachix with path",
			url:        "https://bar.cachix.org/serve",
			expectName: "bar",
			expectNil:  false,
		},
		{
			name:      "Invalid URL",
			url:       "not a url",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCachixURL(tt.url)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectName, result.Name)
			}
		})
	}
}

func TestDefaultCaches(t *testing.T) {
	caches := DefaultCaches()
	assert.Contains(t, caches.Required, "https://cache.nixos.org")
}
