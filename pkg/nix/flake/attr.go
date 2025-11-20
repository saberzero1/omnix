package flake

import "strings"

// Attr represents the (optional) attribute output part of a FlakeURL.
// Example: "foo" in ".#foo" or "packages.x86_64-linux.hello" in ".#packages.x86_64-linux.hello"
type Attr struct {
	value *string
}

// NewAttr creates a new Attr with the given attribute string.
func NewAttr(attr string) Attr {
	return Attr{value: &attr}
}

// NoneAttr returns an Attr with no attribute set.
func NoneAttr() Attr {
	return Attr{value: nil}
}

// GetName returns the attribute name.
// If no attribute is set, returns "default".
func (a Attr) GetName() string {
	if a.value == nil {
		return "default"
	}
	return *a.value
}

// IsNone returns true if no explicit attribute is set.
func (a Attr) IsNone() bool {
	return a.value == nil
}

// AsList returns nested attributes if the attribute is separated by '.'.
// For example, "packages.x86_64-linux.hello" returns ["packages", "x86_64-linux", "hello"].
// Returns an empty slice if no attribute is set.
func (a Attr) AsList() []string {
	if a.value == nil {
		return []string{}
	}
	return strings.Split(*a.value, ".")
}

// String returns the string representation of the attribute.
// Returns empty string if no attribute is set.
func (a Attr) String() string {
	if a.value == nil {
		return ""
	}
	return *a.value
}
