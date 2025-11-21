package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunCmd(t *testing.T) {
	cmd := NewRunCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "run [name]", cmd.Use)
	assert.Contains(t, cmd.Short, "Run tasks")
	assert.Contains(t, cmd.Long, "om/default.yaml")
}

func TestRunCommand_Help(t *testing.T) {
	cmd := NewRunCmd()

	// Capture help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "run")
	assert.Contains(t, output, "om/")
	assert.Contains(t, output, "simplified")
}

func TestGetConfigPath_Default(t *testing.T) {
	// Create temp directory with om/default.yaml
	tmpDir := t.TempDir()
	omDir := filepath.Join(tmpDir, "om")
	require.NoError(t, os.MkdirAll(omDir, 0755))

	configFile := filepath.Join(omDir, "default.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("dir: .\n"), 0644))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	require.NoError(t, os.Chdir(tmpDir))

	// Test
	path, err := getConfigPath(".", "default")
	assert.NoError(t, err)
	assert.Contains(t, path, "om/default.yaml")
}

func TestGetConfigPath_Named(t *testing.T) {
	// Create temp directory with om/test.yaml
	tmpDir := t.TempDir()
	omDir := filepath.Join(tmpDir, "om")
	require.NoError(t, os.MkdirAll(omDir, 0755))

	configFile := filepath.Join(omDir, "test.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("dir: .\n"), 0644))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	require.NoError(t, os.Chdir(tmpDir))

	// Test
	path, err := getConfigPath(".", "test")
	assert.NoError(t, err)
	assert.Contains(t, path, "om/test.yaml")
}

func TestGetConfigPath_NotFound(t *testing.T) {
	// Create temp directory without config
	tmpDir := t.TempDir()

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	require.NoError(t, os.Chdir(tmpDir))

	// Test
	_, err = getConfigPath(".", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoadRunConfig_Valid(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `dir: /tmp/test
steps:
  test-step:
    type: app
    name: test-app
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Test
	config, err := loadRunConfig(configFile)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "/tmp/test", config.Dir)
	assert.NotNil(t, config.Steps)
}

func TestLoadRunConfig_DefaultDir(t *testing.T) {
	// Create temp config file without dir field
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `steps:
  test-step:
    type: app
    name: test-app
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Test
	config, err := loadRunConfig(configFile)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, ".", config.Dir) // Should default to "."
}

func TestLoadRunConfig_InvalidYAML(t *testing.T) {
	// Create temp config file with invalid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	require.NoError(t, os.WriteFile(configFile, []byte("invalid: yaml: content:"), 0644))

	// Test
	_, err := loadRunConfig(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestLoadRunConfig_FileNotFound(t *testing.T) {
	// Test with non-existent file
	_, err := loadRunConfig("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read")
}

func TestRunAppStep_InvalidName(t *testing.T) {
	// Test app step without name field
	stepMap := map[string]interface{}{
		"type": "app",
	}

	err := runAppStep(nil, "test-step", stepMap, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid 'name'")
}

func TestRunAppStep_InvalidArgs(t *testing.T) {
	// Test app step with non-string args
	stepMap := map[string]interface{}{
		"type": "app",
		"name": "test-app",
		"args": []interface{}{
			"valid-arg",
			123, // Invalid: not a string
		},
	}

	err := runAppStep(nil, "test-step", stepMap, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
	assert.Contains(t, err.Error(), "expected string")
}

func TestRunDevshellStep_InvalidCommand(t *testing.T) {
	// Test devshell step with non-string command items
	stepMap := map[string]interface{}{
		"type": "devshell",
		"command": []interface{}{
			"echo",
			123, // Invalid: not a string
		},
	}

	err := runDevshellStep(nil, "test-step", stepMap, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid command argument")
	assert.Contains(t, err.Error(), "expected string")
}

func TestRunDevshellStep_MissingCommand(t *testing.T) {
	// Test devshell step without command field
	stepMap := map[string]interface{}{
		"type": "devshell",
	}

	err := runDevshellStep(nil, "test-step", stepMap, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing 'command'")
}

func TestRunDevshellStep_EmptyCommand(t *testing.T) {
	// Test devshell step with empty command
	stepMap := map[string]interface{}{
		"type":    "devshell",
		"command": []interface{}{},
	}

	err := runDevshellStep(nil, "test-step", stepMap, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty 'command'")
}

func TestRunCustomStep_UnknownType(t *testing.T) {
	// Test step with unknown type
	stepConfig := map[string]interface{}{
		"type": "unknown-type",
	}

	err := runCustomStep(nil, "test-step", stepConfig, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown step type")
}

func TestRunCustomStep_MissingType(t *testing.T) {
	// Test step without type field
	stepConfig := map[string]interface{}{
		"name": "test",
	}

	err := runCustomStep(nil, "test-step", stepConfig, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid 'type'")
}

func TestRunCustomStep_InvalidConfig(t *testing.T) {
	// Test with invalid step config (not a map)
	stepConfig := "invalid-config-string"

	err := runCustomStep(nil, "test-step", stepConfig, ".", ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step config")
}

// Integration test - only runs when not in short mode
func TestRunCommand_MissingConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewRunCmd()

	// Create temp directory without om directory
	tmpDir := t.TempDir()

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	require.NoError(t, os.Chdir(tmpDir))

	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"default"})

	// Execute command
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
