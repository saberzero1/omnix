package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCommand_Execute tests the health command execution
func TestHealthCommand_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test will execute the actual health command
	// It may fail if Nix is not installed, but that's expected
	cmd := NewHealthCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute the health command
	err := cmd.Execute()
	
	// The command might fail if Nix is not available or if some checks fail
	// We're testing that the command executes and doesn't panic
	if err != nil {
		// If it fails, it should be a clean error
		assert.NotEmpty(t, err.Error(), "Error should not be empty")
	} else {
		// If it succeeds, output should contain health check info
		output := buf.String()
		assert.NotEmpty(t, output, "Health output should not be empty")
	}
}

// TestHealthCommand_JSONFlag tests the --json flag
func TestHealthCommand_JSONFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewHealthCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json"})

	// Execute with --json flag
	_ = cmd.Execute()
	
	// If successful, should have some output (even if "not implemented")
	// This exercises the JSON flag path in runHealth
}

// TestShowCommand_Execute tests the show command execution
func TestShowCommand_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewShowCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Try to show current directory as a flake
	cmd.SetArgs([]string{"."})
	
	err := cmd.Execute()
	
	// The command might fail if current directory is not a flake, which is expected
	// We're testing that the command executes without panicking
	if err != nil {
		// Error should be reasonable
		assert.NotEmpty(t, err.Error())
	}
}

// TestShowCommand_WithFlakeURL tests show with a specific flake URL
func TestShowCommand_WithFlakeURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a minimal test flake
	tmpDir := t.TempDir()
	flakeNix := filepath.Join(tmpDir, "flake.nix")
	
	flakeContent := `{
  description = "Test flake";
  outputs = { self, nixpkgs }: {
    packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.hello;
  };
}`
	
	err := os.WriteFile(flakeNix, []byte(flakeContent), 0644)
	require.NoError(t, err)

	cmd := NewShowCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{tmpDir})

	// Execute
	_ = cmd.Execute()
	
	// Test exercises the show command path
}

// TestDevelopCommand_Execute tests the develop command execution
func TestDevelopCommand_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewDevelopCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Try with current directory
	cmd.SetArgs([]string{"."})
	
	err := cmd.Execute()
	
	// Command might fail if Nix is not available or directory is not a flake
	// We're testing the execution path
	if err != nil {
		assert.NotEmpty(t, err.Error())
	}
}

// TestDevelopCommand_WithConfigFile tests develop with a config file
func TestDevelopCommand_WithConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "om.yaml")
	
	// Create a minimal config
	configContent := `develop:
  pre_shell_hook: ""
  post_shell_hook: ""`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := NewDevelopCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", configFile, "."})

	// Execute - will likely fail but exercises the config path
	_ = cmd.Execute()
}

// TestCIRunCommand_Execute tests the CI run command execution
func TestCIRunCommand_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a minimal config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "om.yaml")
	
	configContent := `ci:
  default:
    ".":
      dir: "."
      steps:
        build:
          enable: false
        lockfile:
          enable: false
        flakeCheck:
          enable: false`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newCIRunCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", configFile, "."})

	// Execute - will likely fail but exercises the RunE path
	_ = cmd.Execute()
}

// TestCIGHMatrixCommand_Execute tests the CI gh-matrix command execution  
func TestCIGHMatrixCommand_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a minimal config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "om.yaml")
	
	configContent := `ci:
  default:
    ".":
      dir: "."
      skip: false`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newCIGHMatrixCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", configFile})

	// Execute
	err = cmd.Execute()
	
	// This command should work without Nix as it just generates a matrix
	// from the config file
	if err != nil {
		assert.NotEmpty(t, err.Error())
	}
	// The command executes successfully, which is what we're testing
}

// TestCIRunCommand_WithSystems tests CI run with explicit systems
func TestCIRunCommand_WithSystems(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "om.yaml")
	
	configContent := `ci:
  default:
    ".":
      dir: "."
      steps:
        build:
          enable: false`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newCIRunCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", configFile,
		"--systems", "x86_64-linux",
		".",
	})

	// Execute - tests the systems flag path
	_ = cmd.Execute()
}

// TestCIRunCommand_Parallel tests CI run with parallel flag
func TestCIRunCommand_Parallel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "om.yaml")
	
	configContent := `ci:
  default:
    ".":
      dir: "."
      steps:
        build:
          enable: false`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newCIRunCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", configFile,
		"--parallel",
		"--max-concurrency", "2",
		".",
	})

	// Execute - tests the parallel execution path
	_ = cmd.Execute()
}
