package flake

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryItem_UnmarshalJSON_Val(t *testing.T) {
	jsonData := `{
		"what": "package",
		"derivationName": "hello",
		"shortDescription": "A friendly greeting"
	}`

	var item InventoryItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)

	assert.True(t, item.isLeaf)
	require.NotNil(t, item.leaf)
	require.NotNil(t, item.leaf.Val)
	assert.Equal(t, TypePackage, item.leaf.Val.Type_)
	assert.Equal(t, "hello", *item.leaf.Val.DerivationName)
	assert.Equal(t, "A friendly greeting", *item.leaf.Val.ShortDescription)
}

func TestInventoryItem_UnmarshalJSON_Doc(t *testing.T) {
	jsonData := `"This is a documentation string"`

	var item InventoryItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)

	assert.True(t, item.isLeaf)
	require.NotNil(t, item.leaf)
	require.NotNil(t, item.leaf.Doc)
	assert.Equal(t, "This is a documentation string", *item.leaf.Doc)
}

func TestInventoryItem_UnmarshalJSON_Attrset(t *testing.T) {
	jsonData := `{
		"x86_64-linux": {
			"what": "package",
			"derivationName": "hello"
		},
		"aarch64-darwin": {
			"what": "package",
			"derivationName": "hello"
		}
	}`

	var item InventoryItem
	err := json.Unmarshal([]byte(jsonData), &item)
	require.NoError(t, err)

	assert.False(t, item.isLeaf)
	require.NotNil(t, item.attrset)
	assert.Len(t, item.attrset, 2)

	// Check x86_64-linux
	x86Item, ok := item.attrset["x86_64-linux"]
	assert.True(t, ok)
	assert.True(t, x86Item.isLeaf)
	require.NotNil(t, x86Item.leaf)
	require.NotNil(t, x86Item.leaf.Val)
	assert.Equal(t, TypePackage, x86Item.leaf.Val.Type_)
}

func TestFlakeSchemas_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"inventory": {
			"packages": {
				"x86_64-linux": {
					"default": {
						"what": "package",
						"derivationName": "my-package"
					}
				}
			},
			"devShells": {
				"x86_64-linux": {
					"default": {
						"what": "development environment",
						"derivationName": "my-shell"
					}
				}
			}
		}
	}`

	var schemas FlakeSchemas
	err := json.Unmarshal([]byte(jsonData), &schemas)
	require.NoError(t, err)

	assert.Len(t, schemas.Inventory, 2)

	// Check packages
	packages, ok := schemas.Inventory["packages"]
	assert.True(t, ok)
	assert.False(t, packages.isLeaf)

	// Check devShells
	devShells, ok := schemas.Inventory["devShells"]
	assert.True(t, ok)
	assert.False(t, devShells.isLeaf)
}

func TestFlakeSchemas_ToFlakeOutputs(t *testing.T) {
	const helloPkgName = "hello"
	
	// Create a simple schema structure
	hello := helloPkgName
	pkg := Val{
		Type_:            TypePackage,
		DerivationName:   &hello,
		ShortDescription: nil,
	}

	schemas := &FlakeSchemas{
		Inventory: map[string]InventoryItem{
			"packages": {
				isLeaf: false,
				attrset: map[string]InventoryItem{
					"x86_64-linux": {
						isLeaf: false,
						attrset: map[string]InventoryItem{
							"default": {
								isLeaf: true,
								leaf:   &Leaf{Val: &pkg},
							},
						},
					},
				},
			},
		},
	}

	outputs := schemas.ToFlakeOutputs()
	require.NotNil(t, outputs)
	assert.True(t, outputs.IsAttrset())

	// Navigate to the package
	result := outputs.GetByPath([]string{"packages", "x86_64-linux", "default"})
	require.NotNil(t, result)
	assert.True(t, result.IsVal())
	assert.Equal(t, TypePackage, result.GetVal().Type_)
	assert.Equal(t, helloPkgName, *result.GetVal().DerivationName)
}

func TestFlakeSchemas_ToFlakeOutputs_WithChildren(t *testing.T) {
	const helloPkgName = "hello"
	
	// Test the special "children" key handling
	hello := helloPkgName
	pkg := Val{
		Type_:            TypePackage,
		DerivationName:   &hello,
		ShortDescription: nil,
	}

	schemas := &FlakeSchemas{
		Inventory: map[string]InventoryItem{
			"packages": {
				isLeaf: false,
				attrset: map[string]InventoryItem{
					"children": {
						isLeaf: false,
						attrset: map[string]InventoryItem{
							"x86_64-linux": {
								isLeaf: false,
								attrset: map[string]InventoryItem{
									"default": {
										isLeaf: true,
										leaf:   &Leaf{Val: &pkg},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	outputs := schemas.ToFlakeOutputs()
	require.NotNil(t, outputs)

	// The "children" key should be unwrapped
	result := outputs.GetByPath([]string{"packages", "x86_64-linux", "default"})
	require.NotNil(t, result)
	assert.True(t, result.IsVal())
	assert.Equal(t, TypePackage, result.GetVal().Type_)
}

func TestFlakeSchemas_ToFlakeOutputs_FiltersDocs(t *testing.T) {
	const helloPkgName = "hello"
	
	// Doc strings should be filtered out
	doc := "This is documentation"
	hello := helloPkgName
	pkg := Val{
		Type_:            TypePackage,
		DerivationName:   &hello,
		ShortDescription: nil,
	}

	schemas := &FlakeSchemas{
		Inventory: map[string]InventoryItem{
			"packages": {
				isLeaf: false,
				attrset: map[string]InventoryItem{
					"doc": {
						isLeaf: true,
						leaf:   &Leaf{Doc: &doc},
					},
					"x86_64-linux": {
						isLeaf: false,
						attrset: map[string]InventoryItem{
							"default": {
								isLeaf: true,
								leaf:   &Leaf{Val: &pkg},
							},
						},
					},
				},
			},
		},
	}

	outputs := schemas.ToFlakeOutputs()
	require.NotNil(t, outputs)

	// Should only have packages (doc filtered out)
	attrset := outputs.GetAttrset()
	require.NotNil(t, attrset)
	packages, ok := attrset["packages"]
	assert.True(t, ok)

	// packages should not have "doc" key
	packagesAttrset := packages.GetAttrset()
	require.NotNil(t, packagesAttrset)
	_, hasDoc := packagesAttrset["doc"]
	assert.False(t, hasDoc, "Doc strings should be filtered out")

	// Should have x86_64-linux
	_, hasX86 := packagesAttrset["x86_64-linux"]
	assert.True(t, hasX86)
}

func TestFromNix_RequiresNixEnvironment(t *testing.T) {
	// Mock command
	mockCmd := &mockCmd{
		output: "{}",
		err:    nil,
	}

	// When not built with Nix (normal test environment)
	if !HasNixBuildEnvironment() {
		_, err := FromNix(context.Background(), mockCmd, ".", SystemLinuxX86_64)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires binary built with Nix")
	}
}

func TestGetFlakeSchemas_RequiresNixEnvironment(t *testing.T) {
	// Mock command
	mockCmd := &mockCmd{
		output: "{}",
		err:    nil,
	}

	// When not built with Nix, GetFlakeSchemas should fail early
	// because GetInspectFlake() returns empty string
	if !HasNixBuildEnvironment() {
		// The function should fail due to empty inspect flake path
		// (the actual error might vary based on how the empty path is handled)
		_, err := GetFlakeSchemas(context.Background(), mockCmd, ".", SystemLinuxX86_64)
		// We expect some error, though the exact message depends on implementation
		assert.Error(t, err)
	}
}
