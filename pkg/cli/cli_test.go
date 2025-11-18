package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	// Test that root command is properly initialized
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "om", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "omnix")
}

func TestCommandsRegistered(t *testing.T) {
	// Test that health and init commands are registered
	commands := rootCmd.Commands()
	
	var healthFound, initFound bool
	for _, cmd := range commands {
		if cmd.Name() == "health" {
			healthFound = true
		}
		if cmd.Name() == "init" {
			initFound = true
		}
	}
	
	assert.True(t, healthFound, "health command should be registered")
	assert.True(t, initFound, "init command should be registered")
}
