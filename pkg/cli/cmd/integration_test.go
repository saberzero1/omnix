package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for CLI commands
// These tests exercise the actual command execution paths

func TestRunInit_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "template")
	outputDir := filepath.Join(tmpDir, "output")

	// Create a minimal template
	require.NoError(t, os.MkdirAll(templateDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templateDir, "README.md"),
		[]byte("# PROJECT_NAME"),
		0644,
	))

	cmd := NewInitCmd()
	cmd.SetArgs([]string{
		"--template", templateDir,
		"--param", "project_name=MyProject",
		outputDir,
	})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify the output directory was created
	_, err = os.Stat(outputDir)
	assert.NoError(t, err, "output directory should exist")
}

func TestRunShow_ErrorHandling(t *testing.T) {
	cmd := NewShowCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test with invalid flake path
	cmd.SetArgs([]string{"/nonexistent/path/to/flake"})

	err := cmd.Execute()
	// Should error because the flake doesn't exist
	assert.Error(t, err)
}

func TestRunHealth_ErrorHandling(t *testing.T) {
	cmd := NewHealthCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test help works
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "health")
}

func TestRunDevelop_Structure(t *testing.T) {
	cmd := NewDevelopCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "develop [flake-url]", cmd.Use)

	// Test help
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "development")
}

func TestCIRunCmd_Structure(t *testing.T) {
	cmd := newCIRunCmd() // Use the internal function directly

	require.NotNil(t, cmd)

	// Test help
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "CI")
}

func TestCIGHMatrixCmd_Structure(t *testing.T) {
	cmd := newCIGHMatrixCmd() // Use the internal function directly

	require.NotNil(t, cmd)

	// Test help
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "matrix")
}
