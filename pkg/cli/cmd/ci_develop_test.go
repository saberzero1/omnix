package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCIRunCommand(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
ci:
  default:
    ".":
      dir: "."
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

	// Test the command can be created
	cmd := newCIRunCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "run")
}

func TestCIRunCommand_Help(t *testing.T) {
	cmd := newCIRunCmd()
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "run")
}

func TestCIGHMatrixCommand(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
ci:
  default:
    ".":
      dir: "."
      skip: false
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test the command can be created
	cmd := newCIGHMatrixCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "gh-matrix", cmd.Use)
}

func TestCIGHMatrixCommand_Help(t *testing.T) {
	cmd := newCIGHMatrixCmd()
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "matrix")
}

func TestDevelopCommand(t *testing.T) {
	// Test the command can be created
	cmd := NewDevelopCmd()
	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "develop")
}

func TestDevelopCommand_Help(t *testing.T) {
	cmd := NewDevelopCmd()
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "develop")
}

func TestCICommandStructure(t *testing.T) {
	// Test parent ci command exists
	cmd := NewCICmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "ci", cmd.Use)

	// Test it has subcommands
	assert.Greater(t, len(cmd.Commands()), 0)

	// Test subcommands are registered
	var hasRun, hasMatrix bool
	for _, subcmd := range cmd.Commands() {
		if subcmd.Name() == "run" {
			hasRun = true
		}
		if subcmd.Name() == "gh-matrix" {
			hasMatrix = true
		}
	}

	assert.True(t, hasRun, "ci run command should be registered")
	assert.True(t, hasMatrix, "ci gh-matrix command should be registered")
}

func TestCIFlags(t *testing.T) {
	// Test ci run flags
	runCmd := newCIRunCmd()

	systemsFlag := runCmd.Flags().Lookup("systems")
	assert.NotNil(t, systemsFlag)

	githubFlag := runCmd.Flags().Lookup("github-output")
	assert.NotNil(t, githubFlag)

	configFlag := runCmd.Flags().Lookup("config")
	assert.NotNil(t, configFlag)

	// Test ci gh-matrix flags
	matrixCmd := newCIGHMatrixCmd()

	systemsFlag = matrixCmd.Flags().Lookup("systems")
	assert.NotNil(t, systemsFlag)
}

func TestDevelopFlags(t *testing.T) {
	cmd := NewDevelopCmd()

	configFlag := cmd.Flags().Lookup("config")
	assert.NotNil(t, configFlag)
}

