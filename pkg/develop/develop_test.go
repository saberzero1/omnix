package develop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "README.md", config.Readme.File)
	assert.True(t, config.Readme.Enable)
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
develop:
  readme:
    file: "CUSTOM.md"
    enable: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "CUSTOM.md", config.Readme.File)
	assert.True(t, config.Readme.Enable)
}

func TestLoadConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
develop:
  readme:
    enable: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "README.md", config.Readme.File) // Default applied
	assert.False(t, config.Readme.Enable)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	err := os.WriteFile(configPath, []byte("invalid: [yaml"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config YAML")
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/om.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestReadmeConfig_GetMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")
	content := "# Test README\n\nThis is a test."

	err := os.WriteFile(readmePath, []byte(content), 0644)
	require.NoError(t, err)

	readme := ReadmeConfig{
		File:   "README.md",
		Enable: true,
	}

	markdown, err := readme.GetMarkdown(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, content, markdown)
}

func TestReadmeConfig_GetMarkdown_Disabled(t *testing.T) {
	readme := ReadmeConfig{
		File:   "README.md",
		Enable: false,
	}

	markdown, err := readme.GetMarkdown("/any/dir")
	require.NoError(t, err)
	assert.Equal(t, "", markdown)
}

func TestReadmeConfig_GetMarkdown_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	readme := ReadmeConfig{
		File:   "MISSING.md",
		Enable: true,
	}

	// Should not error if file doesn't exist
	markdown, err := readme.GetMarkdown(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "", markdown)
}

func TestNewProject_LocalPath(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	assert.NotNil(t, project.Dir)
	assert.Equal(t, flake, project.Flake)
	assert.Equal(t, config, project.Config)
}

func TestNewProject_RemoteFlake(t *testing.T) {
	ctx := context.Background()

	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix")
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	assert.Nil(t, project.Dir)
	assert.Equal(t, flake, project.Flake)
}

func TestProject_GetWorkingDir_Local(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	dir, err := project.GetWorkingDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestProject_GetWorkingDir_Remote(t *testing.T) {
	ctx := context.Background()

	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix")
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	// Should return current directory
	dir, err := project.GetWorkingDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestIsCachixAvailable(_ *testing.T) {
	// Just test that it doesn't panic
	_ = IsCachixAvailable()
}
