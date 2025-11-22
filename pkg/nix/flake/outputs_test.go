package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_ToIcon(t *testing.T) {
	tests := []struct {
		name string
		t    Type
		want string
	}{
		{name: "NixOS module", t: TypeNixosModule, want: "❄️"},
		{name: "NixOS configuration", t: TypeNixosConfiguration, want: "🔧"},
		{name: "Darwin configuration", t: TypeDarwinConfiguration, want: "🍎"},
		{name: "Package", t: TypePackage, want: "📦"},
		{name: "Dev shell", t: TypeDevShell, want: "🐚"},
		{name: "Check", t: TypeCheck, want: "🧪"},
		{name: "App", t: TypeApp, want: "📱"},
		{name: "Template", t: TypeTemplate, want: "🏗️"},
		{name: "Unknown", t: TypeUnknown, want: "❓"},
		{name: "Other", t: Type("other"), want: "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.ToIcon()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewValOutput(t *testing.T) {
	val := Val{Type_: TypePackage}
	output := NewValOutput(val)

	assert.True(t, output.IsVal())
	assert.False(t, output.IsAttrset())
	assert.NotNil(t, output.GetVal())
	assert.Nil(t, output.GetAttrset())
	assert.Equal(t, TypePackage, output.GetVal().Type_)
}

func TestNewAttrsetOutput(t *testing.T) {
	attrset := map[string]*FlakeOutputs{
		"default": NewValOutput(Val{Type_: TypePackage}),
	}
	output := NewAttrsetOutput(attrset)

	assert.False(t, output.IsVal())
	assert.True(t, output.IsAttrset())
	assert.Nil(t, output.GetVal())
	assert.NotNil(t, output.GetAttrset())
	assert.Len(t, output.GetAttrset(), 1)
}

func TestFlakeOutputs_GetAttrsetOfVal(t *testing.T) {
	// Create a nested structure
	packageVal := Val{Type_: TypePackage}
	devShellVal := Val{Type_: TypeDevShell}

	attrset := map[string]*FlakeOutputs{
		"default":  NewValOutput(packageVal),
		"devShell": NewValOutput(devShellVal),
		"nested": NewAttrsetOutput(map[string]*FlakeOutputs{
			"inner": NewValOutput(Val{Type_: TypeCheck}),
		}),
	}
	output := NewAttrsetOutput(attrset)

	// Get terminal values
	vals := output.GetAttrsetOfVal()

	// Should only include terminal values, not nested attrsets
	assert.Len(t, vals, 2)

	// Check that we have the expected values
	var foundDefault, foundDevShell bool
	for _, v := range vals {
		if v.Key == "default" && v.Val.Type_ == TypePackage {
			foundDefault = true
		}
		if v.Key == "devShell" && v.Val.Type_ == TypeDevShell {
			foundDevShell = true
		}
	}
	assert.True(t, foundDefault, "Should find default package")
	assert.True(t, foundDevShell, "Should find devShell")
}

func TestFlakeOutputs_GetByPath(t *testing.T) {
	// Create a nested structure:
	// {
	//   "x86_64-linux": {
	//     "default": Val(Package),
	//     "hello": Val(Package)
	//   },
	//   "aarch64-darwin": {
	//     "default": Val(Package)
	//   }
	// }
	defaultPkg := Val{Type_: TypePackage}
	helloPkg := Val{Type_: TypePackage}

	x86Attrset := map[string]*FlakeOutputs{
		"default": NewValOutput(defaultPkg),
		"hello":   NewValOutput(helloPkg),
	}

	aarch64Attrset := map[string]*FlakeOutputs{
		"default": NewValOutput(defaultPkg),
	}

	root := NewAttrsetOutput(map[string]*FlakeOutputs{
		"x86_64-linux":   NewAttrsetOutput(x86Attrset),
		"aarch64-darwin": NewAttrsetOutput(aarch64Attrset),
	})

	tests := []struct {
		name     string
		path     []string
		wantNil  bool
		wantType Type
	}{
		{
			name:     "find x86_64-linux default",
			path:     []string{"x86_64-linux", "default"},
			wantNil:  false,
			wantType: TypePackage,
		},
		{
			name:     "find x86_64-linux hello",
			path:     []string{"x86_64-linux", "hello"},
			wantNil:  false,
			wantType: TypePackage,
		},
		{
			name:     "find aarch64-darwin default",
			path:     []string{"aarch64-darwin", "default"},
			wantNil:  false,
			wantType: TypePackage,
		},
		{
			name:    "nonexistent system",
			path:    []string{"nonexistent", "default"},
			wantNil: true,
		},
		{
			name:    "nonexistent package",
			path:    []string{"x86_64-linux", "nonexistent"},
			wantNil: true,
		},
		{
			name:    "empty path returns root",
			path:    []string{},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := root.GetByPath(tt.path)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if len(tt.path) > 0 && result != nil && result.IsVal() {
					assert.Equal(t, tt.wantType, result.GetVal().Type_)
				}
			}
		})
	}
}

func TestFlakeOutputs_GetByPath_IntermediateAttrset(t *testing.T) {
	// Test getting an intermediate attrset, not just terminal values
	root := NewAttrsetOutput(map[string]*FlakeOutputs{
		"packages": NewAttrsetOutput(map[string]*FlakeOutputs{
			"x86_64-linux": NewAttrsetOutput(map[string]*FlakeOutputs{
				"default": NewValOutput(Val{Type_: TypePackage}),
			}),
		}),
	})

	// Get the intermediate "packages" attrset
	packages := root.GetByPath([]string{"packages"})
	assert.NotNil(t, packages)
	assert.True(t, packages.IsAttrset())
	assert.NotNil(t, packages.GetAttrset()["x86_64-linux"])

	// Get the x86_64-linux attrset
	x86 := root.GetByPath([]string{"packages", "x86_64-linux"})
	assert.NotNil(t, x86)
	assert.True(t, x86.IsAttrset())
	assert.NotNil(t, x86.GetAttrset()["default"])
}

func TestFlakeOutputs_GetAttrsetOfVal_EmptyAttrset(t *testing.T) {
	// Test with an empty attrset
	output := NewAttrsetOutput(map[string]*FlakeOutputs{})
	vals := output.GetAttrsetOfVal()
	assert.Empty(t, vals)
}

func TestFlakeOutputs_GetAttrsetOfVal_OnlyNested(t *testing.T) {
	// Test with only nested attrsets, no terminal values
	output := NewAttrsetOutput(map[string]*FlakeOutputs{
		"packages": NewAttrsetOutput(map[string]*FlakeOutputs{
			"default": NewValOutput(Val{Type_: TypePackage}),
		}),
	})
	vals := output.GetAttrsetOfVal()
	assert.Empty(t, vals, "Should not include nested attrsets")
}

func TestFlakeOutputs_GetAttrsetOfVal_FromVal(t *testing.T) {
	// Test calling GetAttrsetOfVal on a Val (should return empty)
	output := NewValOutput(Val{Type_: TypePackage})
	vals := output.GetAttrsetOfVal()
	assert.Empty(t, vals)
}
