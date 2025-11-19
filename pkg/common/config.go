package common

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// OmConfig represents the omnix configuration with additional metadata
// about the flake URL and reference.
type OmConfig struct {
	// FlakeURL is the flake URL used to load this configuration
	FlakeURL string `json:"flake_url" yaml:"flake_url"`

	// Reference is the (nested) key reference into the flake config.
	// This is the part of the flake URL after `#`
	Reference []string `json:"reference" yaml:"reference"`

	// Config is the configuration tree
	Config *OmConfigTree `json:"config" yaml:"config"`
}

// OmConfigTree represents the whole configuration for omnix parsed from JSON/YAML
type OmConfigTree struct {
	data map[string]map[string]json.RawMessage
}

// NewOmConfigTree creates a new empty config tree
func NewOmConfigTree() *OmConfigTree {
	return &OmConfigTree{
		data: make(map[string]map[string]json.RawMessage),
	}
}

// UnmarshalYAML implements yaml.Unmarshaler
func (t *OmConfigTree) UnmarshalYAML(value *yaml.Node) error {
	var temp map[string]map[string]interface{}
	if err := value.Decode(&temp); err != nil {
		return err
	}

	t.data = make(map[string]map[string]json.RawMessage)
	for k1, v1 := range temp {
		t.data[k1] = make(map[string]json.RawMessage)
		for k2, v2 := range v1 {
			jsonBytes, err := json.Marshal(v2)
			if err != nil {
				return err
			}
			t.data[k1][k2] = jsonBytes
		}
	}
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (t *OmConfigTree) MarshalYAML() (interface{}, error) {
	result := make(map[string]map[string]interface{})
	for k1, v1 := range t.data {
		result[k1] = make(map[string]interface{})
		for k2, v2 := range v1 {
			var temp interface{}
			if err := json.Unmarshal(v2, &temp); err != nil {
				return nil, err
			}
			result[k1][k2] = temp
		}
	}
	return result, nil
}

// Get retrieves all configs of type T for a given sub-config key
// Returns nil if key doesn't exist
func (t *OmConfigTree) Get(key string, result interface{}) error {
	subConfig, exists := t.data[key]
	if !exists {
		return nil // Key doesn't exist, return nil
	}

	// Convert the sub-config map to JSON and then unmarshal into result
	jsonData, err := json.Marshal(subConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// GetSubConfigUnder gets the user referenced (per reference) sub-tree under the given root key.
//
// get_sub_config_under("ci") will return `ci.default` (or default value) without a reference.
// Otherwise, it will use the reference to find the correct sub-tree.
func (c *OmConfig) GetSubConfigUnder(rootKey string, _ interface{}, result interface{}) ([]string, error) {
	// Create a map to hold the sub-config
	subConfigMap := make(map[string]json.RawMessage)

	// Get the config map
	if err := c.Config.Get(rootKey, &subConfigMap); err != nil {
		return nil, err
	}

	// If no config found, return default
	if len(subConfigMap) == 0 {
		// Use reflection or type assertion to set result to defaultValue
		return []string{}, nil
	}

	// Determine which key to use
	key := "default"
	rest := []string{}

	if len(c.Reference) > 0 {
		key = c.Reference[0]
		rest = c.Reference[1:]
	}

	// Get the value for the key
	rawValue, exists := subConfigMap[key]
	if !exists {
		return nil, fmt.Errorf("missing configuration attribute: %s", key)
	}

	// Unmarshal into result
	if err := json.Unmarshal(rawValue, result); err != nil {
		return nil, fmt.Errorf("failed to parse config for key %s: %w", key, err)
	}

	return rest, nil
}

// ParseYAMLConfig parses a YAML configuration file
func ParseYAMLConfig(yamlContent string) (*OmConfigTree, error) {
	var tree OmConfigTree
	if err := yaml.Unmarshal([]byte(yamlContent), &tree); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &tree, nil
}

// ParseJSONConfig parses a JSON configuration
func ParseJSONConfig(jsonContent string) (*OmConfigTree, error) {
	tree := NewOmConfigTree()
	if err := json.Unmarshal([]byte(jsonContent), &tree.data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return tree, nil
}
