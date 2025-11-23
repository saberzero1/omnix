package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/saberzero1/omnix/pkg/ci"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCIRun_FlakeAttributeHandling tests that flake URL attributes are properly
// extracted and used to select the correct config from om.yaml
func TestCIRun_FlakeAttributeHandling(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create an om.yaml with multiple configs
	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `
ci:
  default:
    main:
      dir: "."
      skip: false
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false
  switch:
    custom:
      dir: "."
      skip: false
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false
  production:
    prod-main:
      dir: "."
      skip: false
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load the config
	config, err := ci.LoadConfig(configPath)
	require.NoError(t, err)

	tests := []struct {
		name              string
		flakeURL          string
		expectedConfig    string
		expectedSubflakes []string
		expectError       bool
	}{
		{
			name:              "no attribute - defaults to 'default' config",
			flakeURL:          ".",
			expectedConfig:    "", // Empty means default
			expectedSubflakes: []string{"main"},
			expectError:       false,
		},
		{
			name:              "explicit default attribute",
			flakeURL:          ".#default",
			expectedConfig:    "default",
			expectedSubflakes: []string{"main"},
			expectError:       false,
		},
		{
			name:              "switch config via attribute",
			flakeURL:          ".#switch",
			expectedConfig:    "switch",
			expectedSubflakes: []string{"custom"},
			expectError:       false,
		},
		{
			name:              "production config via attribute",
			flakeURL:          ".#production",
			expectedConfig:    "production",
			expectedSubflakes: []string{"prod-main"},
			expectError:       false,
		},
		{
			name:           "non-existent config",
			flakeURL:       ".#nonexistent",
			expectedConfig: "nonexistent",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse flake URL
			flake, err := nix.ParseFlakeURL(tt.flakeURL)
			require.NoError(t, err)

			// Extract config name from attribute
			attr := flake.GetAttr()
			configName := ""
			attrList := attr.AsList()
			if len(attrList) > 0 {
				configName = attrList[0]
			}

			// Verify we extracted the right config name
			assert.Equal(t, tt.expectedConfig, configName)

			// Get the config by name
			subflakes, err := config.GetConfigByName(configName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)

				// Verify we got the right subflakes
				assert.Len(t, subflakes, len(tt.expectedSubflakes))
				for _, expectedSubflake := range tt.expectedSubflakes {
					assert.Contains(t, subflakes, expectedSubflake,
						"Expected subflake '%s' not found in config", expectedSubflake)
				}
			}
		})
	}
}

// TestCIRun_NestedAttributes tests that nested attributes (e.g., "switch.subconfig")
// only use the first part for config selection
func TestCIRun_NestedAttributes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an om.yaml with a config
	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `
ci:
  dev:
    test:
      dir: "test"
      skip: false
      steps:
        build:
          enable: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := ci.LoadConfig(configPath)
	require.NoError(t, err)

	// Test with nested attribute - should only use first part
	flake, err := nix.ParseFlakeURL(".#dev.extra.nested")
	require.NoError(t, err)

	attr := flake.GetAttr()
	attrList := attr.AsList()

	// Should have 3 parts: ["dev", "extra", "nested"]
	require.Len(t, attrList, 3)

	// But we only use the first part for config name
	configName := attrList[0]
	assert.Equal(t, "dev", configName)

	// Should successfully get the dev config
	subflakes, err := config.GetConfigByName(configName)
	require.NoError(t, err)
	assert.Contains(t, subflakes, "test")
}

// TestCIRun_EmptyAttribute tests that an empty attribute defaults to "default"
func TestCIRun_EmptyAttribute(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `
ci:
  default:
    main:
      dir: "."
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := ci.LoadConfig(configPath)
	require.NoError(t, err)

	// Test various URLs that should all use default config
	testURLs := []string{
		".",
		"./some/path",
		"github:org/repo",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			flake, err := nix.ParseFlakeURL(url)
			require.NoError(t, err)

			attr := flake.GetAttr()
			configName := ""
			attrList := attr.AsList()
			if len(attrList) > 0 {
				configName = attrList[0]
			}

			// All should have empty config name
			assert.Equal(t, "", configName)

			// All should successfully get the default config
			subflakes, err := config.GetConfigByName(configName)
			require.NoError(t, err)
			assert.Contains(t, subflakes, "main")
		})
	}
}

// TestCIRun_Integration is a high-level integration test that simulates
// the actual ci run command flow
func TestCIRun_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a test om.yaml
	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `
ci:
  default:
    test1:
      dir: "."
      skip: false
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false
  alt:
    test2:
      dir: "."
      skip: false
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test running with default config
	t.Run("default config", func(t *testing.T) {
		config, err := ci.LoadConfig(configPath)
		require.NoError(t, err)

		flake, err := nix.ParseFlakeURL(".")
		require.NoError(t, err)

		configName := ""
		subflakes, err := config.GetConfigByName(configName)
		require.NoError(t, err)

		ctx := context.Background()
		opts := ci.RunOptions{
			Systems: []string{"x86_64-linux"},
		}

		results, err := ci.Run(ctx, flake, subflakes, opts)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "test1", results[0].Subflake)
		assert.True(t, results[0].Success)
	})

	// Test running with alt config
	t.Run("alt config", func(t *testing.T) {
		config, err := ci.LoadConfig(configPath)
		require.NoError(t, err)

		flake, err := nix.ParseFlakeURL(".#alt")
		require.NoError(t, err)

		attr := flake.GetAttr()
		attrList := attr.AsList()
		require.Len(t, attrList, 1, "Expected attribute list to have exactly 1 element for .#alt")
		configName := attrList[0]

		subflakes, err := config.GetConfigByName(configName)
		require.NoError(t, err)

		ctx := context.Background()
		opts := ci.RunOptions{
			Systems: []string{"x86_64-linux"},
		}

		baseFlake := flake.WithoutAttr()
		results, err := ci.Run(ctx, baseFlake, subflakes, opts)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "test2", results[0].Subflake)
		assert.True(t, results[0].Success)
	})
}
