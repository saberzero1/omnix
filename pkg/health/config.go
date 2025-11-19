package health

import (
	"fmt"
	"os"

	"github.com/saberzero1/omnix/pkg/nix"
	"gopkg.in/yaml.v3"
)

// Config represents the health configuration from om.yaml
type Config struct {
	// NixVersion configures the Nix version check
	NixVersion *NixVersionConfig `yaml:"nix-version,omitempty" json:"nix-version,omitempty"`

	// Caches configures the cache check
	Caches *CachesConfig `yaml:"caches,omitempty" json:"caches,omitempty"`

	// TrustedUsers configures the trusted users check
	TrustedUsers *TrustedUsersConfig `yaml:"trusted-users,omitempty" json:"trusted-users,omitempty"`

	// FlakeEnabled configures the flake enabled check
	FlakeEnabled *FlakeEnabledConfig `yaml:"flake-enabled,omitempty" json:"flake-enabled,omitempty"`

	// MaxJobs configures the max jobs check
	MaxJobs *MaxJobsConfig `yaml:"max-jobs,omitempty" json:"max-jobs,omitempty"`

	// Rosetta configures the Rosetta check
	Rosetta *RosettaConfig `yaml:"rosetta,omitempty" json:"rosetta,omitempty"`

	// Direnv configures the direnv check
	Direnv *DirenvConfig `yaml:"direnv,omitempty" json:"direnv,omitempty"`

	// Homebrew configures the homebrew check
	Homebrew *HomebrewConfig `yaml:"homebrew,omitempty" json:"homebrew,omitempty"`

	// Shell configures the shell check
	Shell *ShellConfig `yaml:"shell,omitempty" json:"shell,omitempty"`
}

// NixVersionConfig configures the Nix version check
type NixVersionConfig struct {
	// MinVersion is the minimum required Nix version
	MinVersion string `yaml:"min-version,omitempty" json:"min-version,omitempty"`
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// CachesConfig configures the cache check
type CachesConfig struct {
	// Required lists the required caches
	Required []string `yaml:"required,omitempty" json:"required,omitempty"`
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// TrustedUsersConfig configures the trusted users check
type TrustedUsersConfig struct {
	// Enable controls whether this check is enabled (disabled by default for security)
	Enable bool `yaml:"enable" json:"enable"`
}

// FlakeEnabledConfig configures the flake enabled check
type FlakeEnabledConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// MaxJobsConfig configures the max jobs check
type MaxJobsConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// RosettaConfig configures the Rosetta check
type RosettaConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// DirenvConfig configures the direnv check
type DirenvConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// HomebrewConfig configures the homebrew check
type HomebrewConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// ShellConfig configures the shell check
type ShellConfig struct {
	// Enable controls whether this check is enabled
	Enable *bool `yaml:"enable,omitempty" json:"enable,omitempty"`
}

// LoadConfig loads the health configuration from a YAML file
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var wrapper struct {
		Health Config `yaml:"health" json:"health"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return Config{}, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return wrapper.Health, nil
}

// ApplyConfig applies configuration to a NixHealth instance
func (c *Config) ApplyConfig(h *NixHealth) {
	// Apply NixVersion config
	if c.NixVersion != nil {
		if c.NixVersion.MinVersion != "" {
			// Parse the version string
			version, err := nix.ParseVersion(c.NixVersion.MinVersion)
			if err == nil {
				h.NixVersion.MinVersion = version
			}
		}
	}

	// Apply Caches config
	if c.Caches != nil {
		if len(c.Caches.Required) > 0 {
			h.Caches.Required = c.Caches.Required
		}
	}

	// Apply TrustedUsers config
	if c.TrustedUsers != nil {
		h.TrustedUsers.Enable = c.TrustedUsers.Enable
	}
}

// NewFromConfig creates a NixHealth instance with configuration applied
func NewFromConfig(configPath string) (*NixHealth, error) {
	// Start with defaults
	h := Default()

	// Try to load config
	config, err := LoadConfig(configPath)
	if err != nil {
		// If config file doesn't exist, just use defaults
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			return h, nil
		}
		return nil, err
	}

	// Apply config
	config.ApplyConfig(h)

	return h, nil
}

// DefaultCaches returns the default cache configuration
func DefaultCachesConfig() []string {
	return []string{
		"https://cache.nixos.org",
	}
}

// boolPtr is a helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}
