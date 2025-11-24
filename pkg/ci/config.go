package ci

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// Config represents the CI configuration from om.yaml
// It contains named configurations, each with a set of subflakes
type Config struct {
	// Configs is a map of config name to subflake configurations
	// The key is the config name (e.g., "default", "switch")
	// The value is a map of subflake name to SubflakeConfig
	Configs map[string]map[string]SubflakeConfig `yaml:",inline" json:",inline"`
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
	// Supports both "flakeCheck" (Go style) and "flake-check" (Rust style) YAML keys
	FlakeCheck FlakeCheckStep `yaml:"flakeCheck" json:"flakeCheck"`

	// Custom defines custom steps (map of step name to CustomStep)
	Custom map[string]CustomStep `yaml:"custom" json:"custom"`
}

// UnmarshalYAML implements custom unmarshaling to support both "flakeCheck" and "flake-check" keys.
// The "flake-check" key (Rust style) takes precedence over "flakeCheck" (Go style) if both are present.
func (s *StepsConfig) UnmarshalYAML(value *yaml.Node) error {
	// First decode into a raw map to check which keys are present
	var rawMap map[string]yaml.Node
	if err := value.Decode(&rawMap); err != nil {
		return err
	}

	// Define a temporary struct with explicit field mapping
	type stepsConfigAlias struct {
		Build      BuildStep             `yaml:"build"`
		Lockfile   LockfileStep          `yaml:"lockfile"`
		FlakeCheck FlakeCheckStep        `yaml:"flakeCheck"`
		Custom     map[string]CustomStep `yaml:"custom"`
	}

	var temp stepsConfigAlias
	if err := value.Decode(&temp); err != nil {
		return err
	}

	// If "flake-check" key exists (Rust style), use it to override flakeCheck
	// This provides backwards compatibility while preferring the Rust style
	if flakeCheckNode, ok := rawMap["flake-check"]; ok {
		var flakeCheck FlakeCheckStep
		if err := flakeCheckNode.Decode(&flakeCheck); err != nil {
			return err
		}
		temp.FlakeCheck = flakeCheck
	}

	// Copy values to the actual struct
	s.Build = temp.Build
	s.Lockfile = temp.Lockfile
	s.FlakeCheck = temp.FlakeCheck
	s.Custom = temp.Custom

	return nil
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

// CustomStepType represents the type of custom step
type CustomStepType string

const (
	// CustomStepTypeApp runs a flake app
	CustomStepTypeApp CustomStepType = "app"
	// CustomStepTypeDevShell runs a command in a devshell
	CustomStepTypeDevShell CustomStepType = "devshell"
)

// CustomStep defines a custom CI step
type CustomStep struct {
	// Type of the custom step (app or devshell)
	Type CustomStepType `yaml:"type" json:"type"`

	// Name of the app or devshell to use (defaults to "default")
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Args to pass to the app (only for app type)
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`

	// Command to execute in devshell (only for devshell type)
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`

	// Systems is an optional whitelist of systems to run on
	Systems []string `yaml:"systems,omitempty" json:"systems,omitempty"`
}

// CanRunOn checks if this custom step can run on any of the given systems
func (c *CustomStep) CanRunOn(systems []string) bool {
	// If no systems whitelist, can run on any system
	if len(c.Systems) == 0 {
		return true
	}

	// Check if any of the requested systems is in the whitelist
	for _, sys := range systems {
		for _, allowed := range c.Systems {
			if sys == allowed {
				return true
			}
		}
	}

	return false
}

// DefaultConfig returns the default CI configuration
func DefaultConfig() Config {
	return Config{
		Configs: map[string]map[string]SubflakeConfig{
			"default": {
				".": {
					Skip: false,
					Dir:  ".",
					Steps: StepsConfig{
						Build: BuildStep{
							Enable: false,
							Impure: false,
						},
						Lockfile: LockfileStep{
							Enable: false,
						},
						FlakeCheck: FlakeCheckStep{
							Enable: false,
						},
						Custom: make(map[string]CustomStep),
					},
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

	// Apply defaults for each subflake in all configs
	for configName, subflakes := range wrapper.CI.Configs {
		for name, subflake := range subflakes {
			if subflake.Dir == "" {
				subflake.Dir = "."
			}
			wrapper.CI.Configs[configName][name] = subflake
		}
	}

	return wrapper.CI, nil
}

// GetConfigByName returns the subflakes configuration for a specific config name.
// If the config name is not found, returns an error.
// If configName is empty, defaults to "default".
func (c *Config) GetConfigByName(configName string) (map[string]SubflakeConfig, error) {
	if configName == "" {
		configName = "default"
	}

	subflakes, ok := c.Configs[configName]
	if !ok {
		return nil, fmt.Errorf("config '%s' not found in om.yaml", configName)
	}

	return subflakes, nil
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

// GetEnabledSteps returns a list of enabled step names in deterministic order
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

	// Sort custom step names for deterministic order
	customNames := make([]string, 0, len(s.Custom))
	for name := range s.Custom {
		customNames = append(customNames, name)
	}
	sort.Strings(customNames)

	for _, name := range customNames {
		// Custom steps are always enabled if they exist in the config
		enabled = append(enabled, "custom:"+name)
	}

	return enabled
}
