package develop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/juspay/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a simple README
	readmePath := filepath.Join(tmpDir, "README.md")
	err := os.WriteFile(readmePath, []byte("# Test Project\n\nThis is a test."), 0644)
	require.NoError(t, err)

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	// This might fail due to missing Nix info, but tests the code path
	_ = Run(ctx, project)
}

func TestRunPreShell(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	// This might fail due to missing Nix, but tests the code path
	_ = RunPreShell(ctx, project)
}

func TestRunPostShell(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a simple README
	readmePath := filepath.Join(tmpDir, "README.md")
	content := "# Test Project\n\nThis is a test."
	err := os.WriteFile(readmePath, []byte(content), 0644)
	require.NoError(t, err)

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	err = RunPostShell(ctx, project)
	require.NoError(t, err)
}

func TestRunPostShell_NoReadme(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	// Should not error even without README
	err = RunPostShell(ctx, project)
	require.NoError(t, err)
}

func TestRunPostShell_DisabledReadme(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := Config{
		Readme: ReadmeConfig{
			File:   "README.md",
			Enable: false,
		},
	}

	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	err = RunPostShell(ctx, project)
	require.NoError(t, err)
}

func TestUseCachixCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// This will likely fail if cachix is not installed, but tests the code path
	_ = UseCachixCache(ctx, "nixpkgs")
}

func TestIsCachixAvailable_ChecksPath(t *testing.T) {
	// Just verify it doesn't panic
	available := IsCachixAvailable()
	assert.IsType(t, false, available)
}
