package store

import (
	"path/filepath"
	"strings"
)

// Path represents a path in the Nix store.
// See: https://zero-to-nix.com/concepts/nix-store#store-paths
type Path struct {
	path  string
	isDrv bool
}

// NewPath creates a new store path from the given path string.
// It automatically detects if the path is a derivation (.drv file).
func NewPath(path string) Path {
	base := filepath.Base(path)
	isDrv := strings.HasSuffix(base, ".drv")
	return Path{
		path:  path,
		isDrv: isDrv,
	}
}

// String returns the string representation of the store path.
func (p Path) String() string {
	return p.path
}

// IsDrv returns true if this is a derivation path (ends with .drv).
func (p Path) IsDrv() bool {
	return p.isDrv
}

// IsOutput returns true if this is an output path (not a derivation).
func (p Path) IsOutput() bool {
	return !p.isDrv
}

// AsPath returns the underlying path string.
func (p Path) AsPath() string {
	return p.path
}

// Base returns the base name of the path (last element).
func (p Path) Base() string {
	return filepath.Base(p.path)
}

// Dir returns the directory containing the path.
func (p Path) Dir() string {
	return filepath.Dir(p.path)
}
