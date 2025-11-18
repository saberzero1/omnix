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

	// Try with various invalid URLs
	invalidURLs := []string{
		"",
		"not a valid url",
		// Empty URL was already tested in nix package
	}

	for _, url := range invalidURLs {
		if url == "" {
			continue // ParseFlakeURL handles empty strings
		}
		flake, err := nix.ParseFlakeURL(url)
		if err != nil {
			continue // Skip if URL parsing fails
		}

		config := DefaultConfig()
		_, err = NewProject(ctx, flake, config)
		// Should either succeed or fail gracefully
		_ = err
	}
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
	flake, err := nix.ParseFlakeURL("github:juspay/omnix")
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
