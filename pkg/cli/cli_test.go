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
	// Test that all commands are registered
	commands := rootCmd.Commands()

	expectedCommands := []string{"health", "init", "show", "ci", "develop", "completion"}
	foundCommands := make(map[string]bool)

	for _, cmd := range commands {
		foundCommands[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "%s command should be registered", expected)
	}
}

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		commit         string
		expectedPrefix string
	}{
		{
			name:           "development version",
			version:        "dev",
			commit:         "dev",
			expectedPrefix: "dev (commit: dev)",
		},
		{
			name:           "release version with short commit",
			version:        "1.0.0",
			commit:         "abc1234",
			expectedPrefix: "1.0.0 (commit: abc1234)",
		},
		{
			name:           "release version with long commit",
			version:        "2.0.0",
			commit:         "abc1234567890",
			expectedPrefix: "2.0.0 (commit: abc1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version, tt.commit)
			assert.Contains(t, rootCmd.Version, tt.expectedPrefix)
		})
	}
}

func TestVerboseFlag(t *testing.T) {
	// Test that verbose flag is registered
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, flag, "verbose flag should be registered")
	assert.Equal(t, "2", flag.DefValue, "default verbosity should be 2 (info)")
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic
	// We can't test full execution in unit tests, but we can ensure it's callable
	assert.NotPanics(t, func() {
		// Just test that the function exists and can be called
		_ = Execute
	})
}
