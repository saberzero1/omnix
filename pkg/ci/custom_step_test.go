package ci

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomStep_CanRunOn(t *testing.T) {
	tests := []struct {
		name     string
		step     CustomStep
		systems  []string
		expected bool
	}{
		{
			name:     "no systems whitelist - can run on any system",
			step:     CustomStep{Type: CustomStepTypeApp},
			systems:  []string{"x86_64-linux"},
			expected: true,
		},
		{
			name: "systems whitelist matches",
			step: CustomStep{
				Type:    CustomStepTypeApp,
				Systems: []string{"x86_64-linux", "aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux"},
			expected: true,
		},
		{
			name: "systems whitelist doesn't match",
			step: CustomStep{
				Type:    CustomStepTypeApp,
				Systems: []string{"aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux"},
			expected: false,
		},
		{
			name: "multiple systems - one matches",
			step: CustomStep{
				Type:    CustomStepTypeApp,
				Systems: []string{"aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux", "aarch64-darwin"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.step.CanRunOn(tt.systems)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomStep_Types(t *testing.T) {
	appStep := CustomStep{
		Type: CustomStepTypeApp,
		Name: "my-app",
		Args: []string{"arg1", "arg2"},
	}
	assert.Equal(t, CustomStepTypeApp, appStep.Type)
	assert.Equal(t, "my-app", appStep.Name)
	assert.Equal(t, []string{"arg1", "arg2"}, appStep.Args)

	devshellStep := CustomStep{
		Type:    CustomStepTypeDevShell,
		Name:    "my-shell",
		Command: []string{"echo", "hello"},
	}
	assert.Equal(t, CustomStepTypeDevShell, devshellStep.Type)
	assert.Equal(t, "my-shell", devshellStep.Name)
	assert.Equal(t, []string{"echo", "hello"}, devshellStep.Command)
}

func TestLoadConfig_WithCustomSteps(t *testing.T) {
	// Create a temporary config file with custom steps
	tmpDir := t.TempDir()
	configPath := tmpDir + "/om.yaml"

	configContent := `
ci:
  default:
    omnix:
      dir: .
      steps:
        custom:
          om-show:
            type: app
            args:
              - show
              - .
          binary-size-is-small:
            type: app
            name: check-closure-size
            systems:
              - x86_64-linux
          cargo-tests:
            type: devshell
            command:
              - just
              - cargo-test
            systems:
              - x86_64-linux
              - aarch64-darwin
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.NoError(t, err)

	// Check that custom steps are loaded correctly
	assert.Contains(t, config.Default, "omnix")
	omnix := config.Default["omnix"]

	assert.Len(t, omnix.Steps.Custom, 3)

	// Check om-show step
	omShow, ok := omnix.Steps.Custom["om-show"]
	assert.True(t, ok)
	assert.Equal(t, CustomStepTypeApp, omShow.Type)
	assert.Equal(t, []string{"show", "."}, omShow.Args)

	// Check binary-size-is-small step
	binarySize, ok := omnix.Steps.Custom["binary-size-is-small"]
	assert.True(t, ok)
	assert.Equal(t, CustomStepTypeApp, binarySize.Type)
	assert.Equal(t, "check-closure-size", binarySize.Name)
	assert.Equal(t, []string{"x86_64-linux"}, binarySize.Systems)

	// Check cargo-tests step
	cargoTests, ok := omnix.Steps.Custom["cargo-tests"]
	assert.True(t, ok)
	assert.Equal(t, CustomStepTypeDevShell, cargoTests.Type)
	assert.Equal(t, []string{"just", "cargo-test"}, cargoTests.Command)
	assert.Equal(t, []string{"x86_64-linux", "aarch64-darwin"}, cargoTests.Systems)
}
