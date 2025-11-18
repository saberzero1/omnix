package develop

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the develop configuration from om.yaml
type Config struct {
	// Readme configures the welcome message displayed after shell activation
	Readme ReadmeConfig `yaml:"readme" json:"readme"`
}

// ReadmeConfig specifies how to display README information
type ReadmeConfig struct {
	// File is the path to the markdown file to display (default: "README.md")
	File string `yaml:"file" json:"file"`
	// Enable controls whether to show the README (default: true)
	Enable bool `yaml:"enable" json:"enable"`
}

// DefaultConfig returns the default develop configuration
func DefaultConfig() Config {
	return Config{
		Readme: ReadmeConfig{
			File:   "README.md",
			Enable: true,
		},
	}
}

// LoadConfig loads the develop configuration from a YAML file
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		Develop Config `yaml:"develop" json:"develop"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply defaults
	if config.Develop.Readme.File == "" {
		config.Develop.Readme.File = "README.md"
	}

	return config.Develop, nil
}

// GetMarkdown returns the markdown content to display
func (r *ReadmeConfig) GetMarkdown(dir string) (string, error) {
	if !r.Enable {
		return "", nil
	}

	readmePath := filepath.Join(dir, r.File)
	content, err := os.ReadFile(readmePath)
	if err != nil {
		// Don't fail if README doesn't exist
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read README: %w", err)
	}

	return string(content), nil
}
