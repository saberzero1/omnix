package ci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetConfigByName tests retrieving configs by name
func TestGetConfigByName(t *testing.T) {
	config := Config{
		Configs: map[string]map[string]SubflakeConfig{
			"default": {
				"main": {Dir: ".", Skip: false},
			},
			"switch": {
				"custom": {Dir: "custom", Skip: false},
			},
		},
	}

	tests := []struct {
		name          string
		configName    string
		expectError   bool
		expectedCount int
		expectedKey   string
	}{
		{
			name:          "get default config with empty string",
			configName:    "",
			expectError:   false,
			expectedCount: 1,
			expectedKey:   "main",
		},
		{
			name:          "get default config explicitly",
			configName:    "default",
			expectError:   false,
			expectedCount: 1,
			expectedKey:   "main",
		},
		{
			name:          "get switch config",
			configName:    "switch",
			expectError:   false,
			expectedCount: 1,
			expectedKey:   "custom",
		},
		{
			name:        "get non-existent config",
			configName:  "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subflakes, err := config.GetConfigByName(tt.configName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Len(t, subflakes, tt.expectedCount)
				assert.Contains(t, subflakes, tt.expectedKey)
			}
		})
	}
}

// TestLoadConfig_MultipleConfigs tests loading configs with multiple named configurations
func TestLoadConfig_MultipleConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
ci:
  default:
    main:
      dir: "."
      skip: false
      steps:
        build:
          enable: true
    tests:
      dir: "tests"
      skip: false
  switch:
    custom:
      dir: "custom"
      skip: false
      systems:
        - x86_64-linux
      steps:
        build:
          enable: true
          impure: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Check default config
	assert.Contains(t, config.Configs, "default")
	defaultConfig := config.Configs["default"]
	assert.Contains(t, defaultConfig, "main")
	assert.Contains(t, defaultConfig, "tests")

	mainConfig := defaultConfig["main"]
	assert.Equal(t, ".", mainConfig.Dir)
	assert.True(t, mainConfig.Steps.Build.Enable)

	// Check switch config
	assert.Contains(t, config.Configs, "switch")
	switchConfig := config.Configs["switch"]
	assert.Contains(t, switchConfig, "custom")

	customConfig := switchConfig["custom"]
	assert.Equal(t, "custom", customConfig.Dir)
	assert.Equal(t, []string{"x86_64-linux"}, customConfig.Systems)
	assert.True(t, customConfig.Steps.Build.Enable)
	assert.True(t, customConfig.Steps.Build.Impure)
}

// TestGetConfigByName_DefaultBehavior tests that empty config name defaults to "default"
func TestGetConfigByName_DefaultBehavior(t *testing.T) {
	config := Config{
		Configs: map[string]map[string]SubflakeConfig{
			"default": {
				"main": {Dir: ".", Skip: false},
			},
		},
	}

	// Test with empty string - should use "default"
	subflakes1, err := config.GetConfigByName("")
	require.NoError(t, err)
	assert.Contains(t, subflakes1, "main")

	// Test with explicit "default"
	subflakes2, err := config.GetConfigByName("default")
	require.NoError(t, err)
	assert.Contains(t, subflakes2, "main")

	// Both should return the same result
	assert.Equal(t, subflakes1, subflakes2)
}
