package nix

import (
	"context"
	"encoding/json"
	"fmt"
)

// Config represents Nix configuration from `nix show-config --json`.
type Config struct {
	// ExperimentalFeatures are the experimental features currently enabled
	ExperimentalFeatures ConfigValue[[]string] `json:"experimental-features"`
	// System is the current system architecture
	System ConfigValue[string] `json:"system"`
	// Substituters are the cache substituters
	Substituters ConfigValue[[]string] `json:"substituters"`
	// MaxJobs is the maximum number of build jobs to run in parallel
	MaxJobs ConfigValue[int] `json:"max-jobs"`
	// Cores is the number of CPU cores used for builds
	Cores ConfigValue[int] `json:"cores"`
}

// ConfigValue represents a configuration value with its metadata.
type ConfigValue[T any] struct {
	// Value is the current value in use
	Value T `json:"value"`
	// DefaultValue is the default value by Nix
	DefaultValue T `json:"defaultValue"`
	// Description describes this config item
	Description string `json:"description"`
}

// GetConfig retrieves Nix configuration using `nix show-config --json`.
func GetConfig(ctx context.Context) (*Config, error) {
	cmd := NewCmd()
	
	var config Config
	err := cmd.RunJSON(ctx, &config, "show-config", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to get nix config: %w", err)
	}
	
	return &config, nil
}

// IsFlakesEnabled checks if flakes are enabled in the configuration.
func (c *Config) IsFlakesEnabled() bool {
	for _, feature := range c.ExperimentalFeatures.Value {
		if feature == "flakes" || feature == "nix-command" {
			return true
		}
	}
	return false
}

// HasFeature checks if a specific experimental feature is enabled.
func (c *Config) HasFeature(feature string) bool {
	for _, f := range c.ExperimentalFeatures.Value {
		if f == feature {
			return true
		}
	}
	return false
}

// UnmarshalJSON implements custom JSON unmarshaling for Config.
// This handles the fact that nix show-config outputs a complex nested structure.
func (c *Config) UnmarshalJSON(data []byte) error {
	// Define a type alias to avoid infinite recursion
	type Alias Config
	aux := (*Alias)(c)
	
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	
	return nil
}
