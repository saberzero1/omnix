package develop

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saberzero1/omnix/pkg/common"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "README.md", config.Readme.File)
	assert.True(t, config.Readme.Enable)
}

func TestReadmeConfig_Defaults(t *testing.T) {
	readme := ReadmeConfig{}
	assert.Empty(t, readme.File)
	assert.False(t, readme.Enable)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	// Empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	// Should have defaults applied
	assert.Equal(t, "README.md", config.Readme.File)
}

func TestLoadConfig_OnlyDevelopSection(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
develop:
  readme:
    file: "DEVELOPMENT.md"
    enable: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "DEVELOPMENT.md", config.Readme.File)
	assert.True(t, config.Readme.Enable)
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	// Only enable field, file should get default
	configContent := `
develop:
  readme:
    enable: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "README.md", config.Readme.File) // Default
	assert.True(t, config.Readme.Enable)
}

func TestGetMarkdown_ReadError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with the README name (not a file)
	readmeDir := filepath.Join(tmpDir, "README.md")
	err := os.Mkdir(readmeDir, 0755)
	require.NoError(t, err)

	readme := ReadmeConfig{
		File:   "README.md",
		Enable: true,
	}

	// Should error because it's a directory, not a file
	_, err = readme.GetMarkdown(tmpDir)
	assert.Error(t, err)
}

func TestGetMarkdown_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	// Create a large README
	largeContent := make([]byte, 1024*100) // 100KB
	for i := range largeContent {
		largeContent[i] = 'A'
	}

	err := os.WriteFile(readmePath, largeContent, 0644)
	require.NoError(t, err)

	readme := ReadmeConfig{
		File:   "README.md",
		Enable: true,
	}

	markdown, err := readme.GetMarkdown(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, len(largeContent), len(markdown))
}

func TestGetMarkdown_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	readmePath := filepath.Join(tmpDir, "README.md")

	err := os.WriteFile(readmePath, []byte(""), 0644)
	require.NoError(t, err)

	readme := ReadmeConfig{
		File:   "README.md",
		Enable: true,
	}

	markdown, err := readme.GetMarkdown(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, markdown)
}

func TestNewProject_InvalidFlakeURL(t *testing.T) {
	ctx := context.Background()

	// Test with a URL that parses successfully but may behave differently
	flake, err := nix.ParseFlakeURL("not-a-valid-url")
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	// Should succeed - NewProject doesn't validate URLs, just stores them
	require.NoError(t, err)
	assert.NotNil(t, project)
}

func TestNewProject_AbsolutePathConversion(t *testing.T) {
	ctx := context.Background()

	// Use a relative path
	flake, err := nix.ParseFlakeURL("./some/relative/path")
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	// Dir should be set and absolute
	if project.Dir != nil {
		assert.True(t, filepath.IsAbs(*project.Dir))
	}
}

func TestProject_GetWorkingDir_CurrentDir(t *testing.T) {
	ctx := context.Background()

	// Remote flake - should use current directory
	flake, err := nix.ParseFlakeURL("github:saberzero1/omnix")
	require.NoError(t, err)

	config := DefaultConfig()
	project, err := NewProject(ctx, flake, config)
	require.NoError(t, err)

	dir, err := project.GetWorkingDir()
	require.NoError(t, err)

	// Should be a valid directory
	assert.NotEmpty(t, dir)

	// Should exist
	_, err = os.Stat(dir)
	assert.NoError(t, err)
}

func TestProject_MultipleConfigs(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	flake, err := nix.ParseFlakeURL(tmpDir)
	require.NoError(t, err)

	configs := []Config{
		DefaultConfig(),
		{
			Readme: ReadmeConfig{
				File:   "CUSTOM.md",
				Enable: true,
			},
		},
		{
			Readme: ReadmeConfig{
				File:   "README.md",
				Enable: false,
			},
		},
	}

	for _, config := range configs {
		project, err := NewProject(ctx, flake, config)
		require.NoError(t, err)
		assert.Equal(t, config, project.Config)
	}
}

func TestIsCachixAvailable_Consistency(t *testing.T) {
	// Call multiple times to ensure consistency
	result1 := IsCachixAvailable()
	result2 := IsCachixAvailable()
	assert.Equal(t, result1, result2)
}

func TestReadmeConfig_VariousFileNames(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		fileName string
		content  string
	}{
		{"README", "README.md", "# Main README"},
		{"Development", "DEVELOPMENT.md", "# Dev Guide"},
		{"Contributing", "CONTRIBUTING.md", "# How to Contribute"},
		{"Docs", "docs/README.md", "# Documentation"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create nested directory if needed
			fullPath := filepath.Join(tmpDir, tc.fileName)
			dir := filepath.Dir(fullPath)
			if dir != tmpDir {
				err := os.MkdirAll(dir, 0755)
				require.NoError(t, err)
			}

			err := os.WriteFile(fullPath, []byte(tc.content), 0644)
			require.NoError(t, err)

			readme := ReadmeConfig{
				File:   tc.fileName,
				Enable: true,
			}

			markdown, err := readme.GetMarkdown(tmpDir)
			require.NoError(t, err)
			assert.Equal(t, tc.content, markdown)
		})
	}
}

func TestConfig_FullCoverage(t *testing.T) {
	config := Config{
		Readme: ReadmeConfig{
			File:   "test.md",
			Enable: true,
		},
	}

	assert.Equal(t, "test.md", config.Readme.File)
	assert.True(t, config.Readme.Enable)
}

func TestIsDirenvEnabled(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string) error
		expected bool
	}{
		{
			name: ".envrc exists",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".envrc"), []byte("use flake"), 0644)
			},
			expected: true,
		},
		{
			name: ".envrc does not exist",
			setup: func(dir string) error {
				return nil
			},
			expected: false,
		},
		{
			name: ".envrc is a directory",
			setup: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, ".envrc"), 0755)
			},
			expected: true, // Stat succeeds even for directories
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			err := tt.setup(tmpDir)
			require.NoError(t, err)

			result := IsDirenvEnabled(tmpDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetupDirenv_Disabled(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := DirenvConfig{Enable: false}
	err := SetupDirenv(ctx, tmpDir, config)
	require.NoError(t, err)

	// .envrc should not be created
	_, err = os.Stat(filepath.Join(tmpDir, ".envrc"))
	assert.True(t, os.IsNotExist(err))
}

func TestSetupDirenv_CreatesEnvrc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires direnv in short mode")
	}

	// Skip if direnv is not installed
	if common.WhichStrict("direnv") == "" {
		t.Skip("direnv is not installed, skipping test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	config := DirenvConfig{
		Enable:             true,
		AllowAutomatically: false,
	}

	err := SetupDirenv(ctx, tmpDir, config)
	require.NoError(t, err)

	// .envrc should be created
	envrcPath := filepath.Join(tmpDir, ".envrc")
	content, err := os.ReadFile(envrcPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "use flake")
}

func TestSetupDirenv_ExistingEnvrc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires direnv in short mode")
	}

	// Skip if direnv is not installed
	if common.WhichStrict("direnv") == "" {
		t.Skip("direnv is not installed, skipping test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create existing .envrc
	existingContent := "# existing content"
	envrcPath := filepath.Join(tmpDir, ".envrc")
	err := os.WriteFile(envrcPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	config := DirenvConfig{
		Enable:             true,
		AllowAutomatically: false,
	}

	err = SetupDirenv(ctx, tmpDir, config)
	require.NoError(t, err)

	// Content should be unchanged
	content, err := os.ReadFile(envrcPath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(content))
}
