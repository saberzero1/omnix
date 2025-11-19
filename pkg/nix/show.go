package nix

import (
	"context"
	"encoding/json"
	"fmt"
)

// FlakeOutputs represents the outputs of a Nix flake.
// It can be either a terminal value or a nested attribute set.
type FlakeOutputs struct {
	// Val is the terminal value if this is not an attribute set
	Val *FlakeVal `json:"-"`
	// Attrset is the nested attribute set
	Attrset map[string]*FlakeOutputs `json:"-"`
}

// FlakeVal represents a terminal flake output value
type FlakeVal struct {
	// ShortDescription is a brief description of the output
	ShortDescription string `json:"shortDescription,omitempty"`
	// Type is the type of the output (e.g., "app", "package", "devShell")
	Type string `json:"type,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for FlakeOutputs
func (f *FlakeOutputs) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as an attribute set first
	var attrset map[string]*FlakeOutputs
	if err := json.Unmarshal(data, &attrset); err == nil {
		// Check if it has keys that indicate it's a value, not an attrset
		// Look for specific keys that indicate this is a terminal value
		if _, hasType := attrset["type"]; hasType {
			// This is actually a terminal value
			var val FlakeVal
			if err := json.Unmarshal(data, &val); err != nil {
				return err
			}
			f.Val = &val
			return nil
		}
		f.Attrset = attrset
		return nil
	}

	// Try to unmarshal as a terminal value
	var val FlakeVal
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	f.Val = &val
	return nil
}

// GetByPath looks up the given path in the flake outputs, returning the value if it exists.
func (f *FlakeOutputs) GetByPath(path ...string) *FlakeOutputs {
	current := f
	for _, key := range path {
		if current.Attrset == nil {
			return nil
		}
		next, ok := current.Attrset[key]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

// GetAttrsetOfVal returns all terminal values in the attribute set as a slice.
// Non-terminal nested attribute sets are skipped.
func (f *FlakeOutputs) GetAttrsetOfVal() []struct {
	Name string
	Val  FlakeVal
} {
	if f.Attrset == nil {
		return nil
	}

	var result []struct {
		Name string
		Val  FlakeVal
	}

	for name, output := range f.Attrset {
		if output.Val != nil {
			result = append(result, struct {
				Name string
				Val  FlakeVal
			}{
				Name: name,
				Val:  *output.Val,
			})
		}
	}

	return result
}

// FlakeMetadata represents metadata about a Nix flake
type FlakeMetadata struct {
	// Description of the flake
	Description string `json:"description,omitempty"`
	// Outputs from the flake
	Outputs *FlakeOutputs `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling for FlakeMetadata
func (f *FlakeMetadata) UnmarshalJSON(data []byte) error {
	type Alias FlakeMetadata
	aux := &struct {
		Outputs json.RawMessage `json:"outputs,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Outputs != nil {
		var outputs FlakeOutputs
		if err := json.Unmarshal(aux.Outputs, &outputs); err != nil {
			return err
		}
		f.Outputs = &outputs
	}

	return nil
}

// FlakeShow returns the metadata and outputs of a flake
func (c *Cmd) FlakeShow(ctx context.Context, flakeURL FlakeURL) (*FlakeMetadata, error) {
	var metadata FlakeMetadata
	err := c.RunJSON(ctx, &metadata, "flake", "show", flakeURL.String(), "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to show flake %s: %w", flakeURL, err)
	}
	return &metadata, nil
}
