package flake

import (
	"context"
	"encoding/json"
	"fmt"
)

// FlakeSchemas represents the schema of a given flake evaluated using inspect-flake.
type FlakeSchemas struct {
	// Inventory maps output names to inventory items
	// Each key represents either a top-level flake output or other metadata (e.g. "docs")
	Inventory map[string]InventoryItem `json:"inventory"`
}

// InventoryItem represents a tree-like structure for each flake output or metadata.
// It can be either a Leaf (terminal node) or an Attrset (non-terminal node).
type InventoryItem struct {
	// IsLeaf indicates if this is a terminal value
	isLeaf bool
	// Leaf data (only set if IsLeaf is true)
	leaf *Leaf
	// Attrset data (only set if IsLeaf is false)
	attrset map[string]InventoryItem
}

// Leaf represents a terminal value of a flake schema.
type Leaf struct {
	// Val is the value data (if this is a Val type)
	Val *Val
	// Doc is the documentation string (if this is a Doc type)
	Doc *string
}

// MarshalJSON implements custom JSON marshaling for InventoryItem.
func (i InventoryItem) MarshalJSON() ([]byte, error) {
	if i.isLeaf && i.leaf != nil {
		// For leaf nodes, we need to marshal either Val or Doc
		if i.leaf.Val != nil {
			return json.Marshal(i.leaf.Val)
		}
		if i.leaf.Doc != nil {
			return json.Marshal(*i.leaf.Doc)
		}
		return []byte("null"), nil
	}
	// For attrset nodes
	return json.Marshal(i.attrset)
}

// UnmarshalJSON implements custom JSON unmarshaling for InventoryItem.
func (i *InventoryItem) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as Val first
	var val Val
	if err := json.Unmarshal(data, &val); err == nil && val.Type_ != TypeUnknown {
		i.isLeaf = true
		i.leaf = &Leaf{Val: &val}
		return nil
	}

	// Try to unmarshal as string (Doc)
	var doc string
	if err := json.Unmarshal(data, &doc); err == nil && doc != "" {
		i.isLeaf = true
		i.leaf = &Leaf{Doc: &doc}
		return nil
	}

	// Try to unmarshal as attrset
	var attrset map[string]InventoryItem
	if err := json.Unmarshal(data, &attrset); err == nil {
		i.isLeaf = false
		i.attrset = attrset
		return nil
	}

	return fmt.Errorf("failed to unmarshal InventoryItem")
}

// ToFlakeOutputs converts FlakeSchemas to FlakeOutputs.
func (fs *FlakeSchemas) ToFlakeOutputs() *FlakeOutputs {
	outputMap := make(map[string]*FlakeOutputs)

	for k, v := range fs.Inventory {
		if out := inventoryItemToFlakeOutputs(&v); out != nil {
			outputMap[k] = out
		}
	}

	if len(outputMap) == 0 {
		return nil
	}

	return NewAttrsetOutput(outputMap)
}

// inventoryItemToFlakeOutputs converts an InventoryItem to FlakeOutputs.
func inventoryItemToFlakeOutputs(item *InventoryItem) *FlakeOutputs {
	if item.isLeaf && item.leaf != nil {
		// Terminal node - only convert if it's a Val (not Doc)
		if item.leaf.Val != nil {
			return NewValOutput(*item.leaf.Val)
		}
		return nil
	}

	// Non-terminal node
	if item.attrset != nil {
		// Check for special "children" key
		if children, ok := item.attrset["children"]; ok {
			return inventoryItemToFlakeOutputs(&children)
		}

		// Regular attrset
		outputMap := make(map[string]*FlakeOutputs)
		for k, v := range item.attrset {
			if out := inventoryItemToFlakeOutputs(&v); out != nil {
				outputMap[k] = out
			}
		}

		if len(outputMap) == 0 {
			return nil
		}

		return NewAttrsetOutput(outputMap)
	}

	return nil
}

// GetKnownSystemFlakeURL returns the nix-systems flake URL for a known system.
// Returns empty string if the system is not known.
func getKnownSystemFlakeURL(sys System) string {
	// Map system strings to known nix-systems flake URLs
	nixSystemsMap := map[string]string{
		"aarch64-linux":  "github:nix-systems/aarch64-linux",
		"x86_64-linux":   "github:nix-systems/x86_64-linux",
		"x86_64-darwin":  "github:nix-systems/x86_64-darwin",
		"aarch64-darwin": "github:nix-systems/aarch64-darwin",
	}

	if url, ok := nixSystemsMap[sys.String()]; ok {
		return url
	}
	return ""
}

// FromNix constructs a Flake from a URL by using the inspect-flake to analyze its outputs.
// This requires the binary to be built with Nix (HasNixBuildEnvironment() must return true).
func FromNix(ctx context.Context, cmd Cmd, flakeURL string, system System) (*Flake, error) {
	// Check if we have the necessary environment
	if !HasNixBuildEnvironment() {
		return nil, fmt.Errorf("FromNix requires binary built with Nix (DEFAULT_FLAKE_SCHEMAS and INSPECT_FLAKE not available)")
	}

	// Get the schemas for this flake
	schemas, err := GetFlakeSchemas(ctx, cmd, flakeURL, system)
	if err != nil {
		return nil, fmt.Errorf("failed to get flake schemas: %w", err)
	}

	// Convert schemas to outputs
	outputs := schemas.ToFlakeOutputs()

	return NewFlake(flakeURL, outputs), nil
}

// GetFlakeSchemas retrieves the FlakeSchemas for a given flake URL.
// This uses the inspect flake and default flake-schemas paths injected at build time.
func GetFlakeSchemas(ctx context.Context, cmd Cmd, flakeURL string, system System) (*FlakeSchemas, error) {
	// Check if we have the necessary environment
	if !HasNixBuildEnvironment() {
		return nil, fmt.Errorf("GetFlakeSchemas requires binary built with Nix (DEFAULT_FLAKE_SCHEMAS and INSPECT_FLAKE not available)")
	}

	// Construct the inspect flake URL with the appropriate attribute
	// Using excludingOutputPaths for faster evaluation (see Rust implementation)
	inspectURL := GetInspectFlake() + "#contents.excludingOutputPaths"

	// Get the systems flake for this system
	systemsURL := getKnownSystemFlakeURL(system)
	if systemsURL == "" {
		return nil, fmt.Errorf("unsupported system: %s", system.String())
	}

	// Build the flake options with override inputs
	opts := &FlakeOptions{
		NoWriteLockFile: true,
		OverrideInputs: map[string]string{
			"flake-schemas": GetDefaultFlakeSchemas(),
			"flake":         flakeURL,
			"systems":       systemsURL,
		},
	}

	// Evaluate the inspect flake
	schemas, err := Eval[FlakeSchemas](ctx, cmd, opts, inspectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate inspect flake: %w", err)
	}

	return &schemas, nil
}
