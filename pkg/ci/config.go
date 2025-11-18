package ci

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the CI configuration from om.yaml
type Config struct {
	// Default contains the default subflake configurations
	Default map[string]SubflakeConfig `yaml:"default" json:"default"`
}

// SubflakeConfig represents configuration for a sub-flake
type SubflakeConfig struct {
	// Skip controls whether to skip this subflake
	Skip bool `yaml:"skip" json:"skip"`

	// Dir is the subdirectory where the flake lives
	Dir string `yaml:"dir" json:"dir"`

	// OverrideInputs specifies inputs to override via --override-input
	OverrideInputs map[string]string `yaml:"overrideInputs" json:"overrideInputs"`

	// Systems is an optional whitelist of systems to build on
	Systems []string `yaml:"systems" json:"systems"`

	// Steps defines which CI steps to run
	Steps StepsConfig `yaml:"steps" json:"steps"`
}

// StepsConfig defines the CI steps to run
type StepsConfig struct {
	// Build controls the build step
	Build BuildStep `yaml:"build" json:"build"`

	// Lockfile controls the lockfile check step
	Lockfile LockfileStep `yaml:"lockfile" json:"lockfile"`

	// FlakeCheck controls the flake check step
	FlakeCheck FlakeCheckStep `yaml:"flakeCheck" json:"flakeCheck"`

	// Custom defines custom steps
	Custom []CustomStep `yaml:"custom" json:"custom"`
}

// BuildStep configures the build step
type BuildStep struct {
	// Enable controls whether this step is enabled
	Enable bool `yaml:"enable" json:"enable"`

	// Impure controls whether to pass --impure to nix build
	Impure bool `yaml:"impure" json:"impure"`
}

// LockfileStep configures the lockfile check step
type LockfileStep struct {
	// Enable controls whether this step is enabled
	Enable bool `yaml:"enable" json:"enable"`
}

// FlakeCheckStep configures the flake check step
type FlakeCheckStep struct {
	// Enable controls whether this step is enabled
	Enable bool `yaml:"enable" json:"enable"`
}

// CustomStep defines a custom CI step
type CustomStep struct {
	// Name of the custom step
	Name string `yaml:"name" json:"name"`

	// Command to execute
	Command []string `yaml:"command" json:"command"`

	// Enable controls whether this step is enabled
	Enable bool `yaml:"enable" json:"enable"`
}

// DefaultConfig returns the default CI configuration
func DefaultConfig() Config {
	return Config{
		Default: map[string]SubflakeConfig{
			".": {
				Skip: false,
				Dir:  ".",
				Steps: StepsConfig{
					Build: BuildStep{
						Enable: true,
						Impure: false,
					},
					Lockfile: LockfileStep{
						Enable: true,
					},
					FlakeCheck: FlakeCheckStep{
						Enable: true,
					},
					Custom: []CustomStep{},
				},
			},
		},
	}
}

// LoadConfig loads the CI configuration from a YAML file
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var wrapper struct {
		CI Config `yaml:"ci" json:"ci"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return Config{}, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply defaults for each subflake
	for name, subflake := range wrapper.CI.Default {
		if subflake.Dir == "" {
			subflake.Dir = "."
		}
		wrapper.CI.Default[name] = subflake
	}

	return wrapper.CI, nil
}

// CanRunOn checks if this subflake can run on any of the given systems
func (s *SubflakeConfig) CanRunOn(systems []string) bool {
	// If no systems whitelist, can run on any system
	if len(s.Systems) == 0 {
		return true
	}

	// Check if any of the requested systems is in the whitelist
	for _, sys := range systems {
		for _, allowed := range s.Systems {
			if sys == allowed {
				return true
			}
		}
	}

	return false
}

// GetEnabledSteps returns a list of enabled step names
func (s *StepsConfig) GetEnabledSteps() []string {
	var enabled []string

	if s.Build.Enable {
		enabled = append(enabled, "build")
	}
	if s.Lockfile.Enable {
		enabled = append(enabled, "lockfile")
	}
	if s.FlakeCheck.Enable {
		enabled = append(enabled, "flakeCheck")
	}
	for _, custom := range s.Custom {
		if custom.Enable {
			enabled = append(enabled, "custom:"+custom.Name)
		}
	}

	return enabled
}
