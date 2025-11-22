package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFlake(t *testing.T) {
	outputs := NewAttrsetOutput(map[string]*FlakeOutputs{
		"packages": NewAttrsetOutput(map[string]*FlakeOutputs{
			"x86_64-linux": NewAttrsetOutput(map[string]*FlakeOutputs{
				"default": NewValOutput(Val{Type_: TypePackage}),
			}),
		}),
	})

	flake := NewFlake("github:nixos/nixpkgs", outputs)

	assert.Equal(t, "github:nixos/nixpkgs", flake.URL)
	assert.NotNil(t, flake.Outputs)
	assert.True(t, flake.Outputs.IsAttrset())

	// Navigate to packages.x86_64-linux.default
	result := flake.Outputs.GetByPath([]string{"packages", "x86_64-linux", "default"})
	assert.NotNil(t, result)
	assert.True(t, result.IsVal())
	assert.Equal(t, TypePackage, result.GetVal().Type_)
}

func TestFlake_Structure(t *testing.T) {
	const helloPkgName = "hello"
	
	// Test that we can create a complete flake structure
	name1 := "my-dev-shell"
	desc1 := "Development environment"
	devShell := NewValOutput(Val{
		Type_:            TypeDevShell,
		DerivationName:   &name1,
		ShortDescription: &desc1,
	})

	name2 := helloPkgName
	desc2 := "Hello world program"
	defaultPkg := NewValOutput(Val{
		Type_:            TypePackage,
		DerivationName:   &name2,
		ShortDescription: &desc2,
	})

	outputs := NewAttrsetOutput(map[string]*FlakeOutputs{
		"devShells": NewAttrsetOutput(map[string]*FlakeOutputs{
			"x86_64-linux": NewAttrsetOutput(map[string]*FlakeOutputs{
				"default": devShell,
			}),
		}),
		"packages": NewAttrsetOutput(map[string]*FlakeOutputs{
			"x86_64-linux": NewAttrsetOutput(map[string]*FlakeOutputs{
				"default": defaultPkg,
			}),
		}),
	})

	flake := NewFlake(".", outputs)

	// Test devShell
	devShellResult := flake.Outputs.GetByPath([]string{"devShells", "x86_64-linux", "default"})
	assert.NotNil(t, devShellResult)
	assert.Equal(t, TypeDevShell, devShellResult.GetVal().Type_)
	assert.Equal(t, "my-dev-shell", *devShellResult.GetVal().DerivationName)

	// Test package
	pkgResult := flake.Outputs.GetByPath([]string{"packages", "x86_64-linux", "default"})
	assert.NotNil(t, pkgResult)
	assert.Equal(t, TypePackage, pkgResult.GetVal().Type_)
	assert.Equal(t, helloPkgName, *pkgResult.GetVal().DerivationName)
}
