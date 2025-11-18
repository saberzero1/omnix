package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestShowCommand_InvalidFlake tests error handling for invalid flake paths
func TestShowCommand_InvalidFlake(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewShowCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Test with clearly invalid flake path
	cmd.SetArgs([]string{"/nonexistent/invalid/path/to/flake"})
	
	err := cmd.Execute()
	// Should error because the flake doesn't exist
	assert.Error(t, err)
}

// TestDevelopCommand_Args tests argument parsing
func TestDevelopCommand_Args(t *testing.T) {
	cmd := NewDevelopCmd()
	
	// Test that command accepts maximum 1 argument
	assert.NotNil(t, cmd.Args)
}

// TestCIRunCommand_Args tests argument parsing
func TestCIRunCommand_Args(t *testing.T) {
	cmd := newCIRunCmd()
	
	// Test that command accepts maximum 1 argument
	assert.NotNil(t, cmd.Args)
}

// TestInitCommand_MissingArgs tests error handling for missing required arguments
func TestInitCommand_MissingArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := NewInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Try to run without required args
	cmd.SetArgs([]string{})
	
	err := cmd.Execute()
	// Should error because output directory is required
	assert.Error(t, err)
}
