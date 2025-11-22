package flake

// Val represents a terminal value of a flake output.
type Val struct {
	// Type_ represents the type of the flake output
	Type_ Type `json:"what"`
	// DerivationName is the name of the derivation if the output is a derivation
	DerivationName *string `json:"derivationName,omitempty"`
	// ShortDescription is a short description from meta.description
	ShortDescription *string `json:"shortDescription,omitempty"`
}

// Type represents the type of a flake output.
// These types are based on flake-schemas.
type Type string

const (
	// TypeNixosModule represents a NixOS module
	TypeNixosModule Type = "NixOS module"
	// TypeNixosConfiguration represents a NixOS configuration
	TypeNixosConfiguration Type = "NixOS configuration"
	// TypeDarwinConfiguration represents a nix-darwin configuration
	TypeDarwinConfiguration Type = "nix-darwin configuration"
	// TypePackage represents a package
	TypePackage Type = "package"
	// TypeDevShell represents a development environment
	TypeDevShell Type = "development environment"
	// TypeCheck represents a CI test
	TypeCheck Type = "CI test"
	// TypeApp represents an app
	TypeApp Type = "app"
	// TypeTemplate represents a template
	TypeTemplate Type = "template"
	// TypeUnknown represents an unknown type
	TypeUnknown Type = ""
)

// ToIcon returns the icon for this type.
func (t Type) ToIcon() string {
	switch t {
	case TypeNixosModule:
		return "❄️"
	case TypeNixosConfiguration:
		return "🔧"
	case TypeDarwinConfiguration:
		return "🍎"
	case TypePackage:
		return "📦"
	case TypeDevShell:
		return "🐚"
	case TypeCheck:
		return "🧪"
	case TypeApp:
		return "📱"
	case TypeTemplate:
		return "🏗️"
	default:
		return "❓"
	}
}

// FlakeOutputs represents the outputs of a flake.
// It can be either a terminal Val or an Attrset of nested FlakeOutputs.
type FlakeOutputs struct {
	val     *Val
	attrset map[string]*FlakeOutputs
}

// NewValOutput creates a FlakeOutputs with a terminal value.
func NewValOutput(val Val) *FlakeOutputs {
	return &FlakeOutputs{val: &val}
}

// NewAttrsetOutput creates a FlakeOutputs with an attrset.
func NewAttrsetOutput(attrset map[string]*FlakeOutputs) *FlakeOutputs {
	return &FlakeOutputs{attrset: attrset}
}

// IsVal returns true if this is a terminal value.
func (f *FlakeOutputs) IsVal() bool {
	return f.val != nil
}

// IsAttrset returns true if this is an attrset.
func (f *FlakeOutputs) IsAttrset() bool {
	return f.attrset != nil
}

// GetVal returns the terminal value if this is a Val, otherwise nil.
func (f *FlakeOutputs) GetVal() *Val {
	return f.val
}

// GetAttrset returns the attrset if this is an Attrset, otherwise nil.
func (f *FlakeOutputs) GetAttrset() map[string]*FlakeOutputs {
	return f.attrset
}

// GetAttrsetOfVal returns a slice of key-value pairs where the values are terminal Vals.
// Only terminal values are included in the result.
func (f *FlakeOutputs) GetAttrsetOfVal() []struct {
	Key string
	Val Val
} {
	result := []struct {
		Key string
		Val Val
	}{}

	if f.attrset == nil {
		return result
	}

	for k, v := range f.attrset {
		if v.val != nil {
			result = append(result, struct {
				Key string
				Val Val
			}{Key: k, Val: *v.val})
		}
	}

	return result
}

// GetByPath looks up the given path in the output tree, returning the value if it exists.
// For example: GetByPath([]string{"aarch64-darwin", "default"})
func (f *FlakeOutputs) GetByPath(path []string) *FlakeOutputs {
	current := f
	for _, key := range path {
		if current.attrset == nil {
			return nil
		}
		next, ok := current.attrset[key]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}
