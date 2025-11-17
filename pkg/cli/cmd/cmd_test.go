package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthCmd(t *testing.T) {
	cmd := NewHealthCmd()
	
	assert.NotNil(t, cmd)
	assert.Equal(t, "health", cmd.Use)
	assert.Contains(t, cmd.Short, "Check the health")
	
	// Test flags are registered
	jsonFlag := cmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag, "json flag should be registered")
}

func TestHealthCommand_Help(t *testing.T) {
	cmd := NewHealthCmd()
	
	// Capture help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "health")
	assert.Contains(t, output, "Nix")
}

func TestNewInitCmd(t *testing.T) {
	cmd := NewInitCmd()
	
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "init")
	assert.Contains(t, cmd.Short, "Initialize")
	
	// Test flags are registered
	templateFlag := cmd.Flags().Lookup("template")
	assert.NotNil(t, templateFlag, "template flag should be registered")
	
	paramFlag := cmd.Flags().Lookup("param")
	assert.NotNil(t, paramFlag, "param flag should be registered")
	
	nonInteractiveFlag := cmd.Flags().Lookup("non-interactive")
	assert.NotNil(t, nonInteractiveFlag, "non-interactive flag should be registered")
}

func TestInitCommand_Help(t *testing.T) {
	cmd := NewInitCmd()
	
	// Capture help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "init")
	assert.Contains(t, output, "template")
}

func TestInitCommand_MissingTemplate(t *testing.T) {
	cmd := NewInitCmd()
	
	// Create temp dir for output
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	
	// Try to run without template flag
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{outputDir})
	
	err := cmd.Execute()
	assert.Error(t, err, "should fail without template flag")
}

func TestInitCommand_TemplateNotExists(t *testing.T) {
	cmd := NewInitCmd()
	
	// Create temp dir for output
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	
	// Try to run with non-existent template
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"--template", "/nonexistent/path",
		outputDir,
	})
	
	err := cmd.Execute()
	assert.Error(t, err, "should fail with non-existent template")
	assert.Contains(t, err.Error(), "does not exist")
}

func TestInitCommand_OutputDirExists(t *testing.T) {
	cmd := NewInitCmd()
	
	// Create temp dirs
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "template")
	outputDir := filepath.Join(tmpDir, "output")
	
	// Create template and output dirs
	require.NoError(t, os.MkdirAll(templateDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))
	
	// Try to run with existing output dir
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"--template", templateDir,
		outputDir,
	})
	
	err := cmd.Execute()
	assert.Error(t, err, "should fail when output dir exists")
	assert.Contains(t, err.Error(), "already exists")
}

func TestInitCommand_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a new command for each test to avoid state issues
	cmd := NewInitCmd()
	
	// Create temp dirs
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "template")
	outputDir := filepath.Join(tmpDir, "output")
	
	// Create a simple template
	require.NoError(t, os.MkdirAll(templateDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templateDir, "README.md"),
		[]byte("# Test Project"),
		0644,
	))
	
	// Set args
	cmd.SetArgs([]string{
		"--template", templateDir,
		outputDir,
	})
	
	// Execute command
	err := cmd.Execute()
	assert.NoError(t, err)
	
	// Verify file was copied - this is the important part
	_, err = os.Stat(filepath.Join(outputDir, "README.md"))
	assert.NoError(t, err, "README.md should be copied")
}
