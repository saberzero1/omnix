package nix

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version represents a Nix version parsed from `nix --version`.
// The version format is typically "nix (Nix) X.Y.Z" or just "X.Y.Z".
type Version struct {
	Major uint32
	Minor uint32
	Patch uint32
}

// String returns the string representation of the version (e.g., "2.13.0").
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion parses a version string from `nix --version` output.
// It accepts formats like:
// - "nix (Nix) 2.13.0"
// - "2.13.0"
// - "nix (Determinate Nix 3.6.6) 2.29.0"
func ParseVersion(s string) (Version, error) {
	// Lenient regex that matches the version number at the end
	re := regexp.MustCompile(`(?:nix \((?:Nix|Determinate Nix [^\)]+)\) )?(\d+)\.(\d+)\.(\d+)$`)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return Version{}, fmt.Errorf("failed to parse nix version from: %s", s)
	}

	major, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse major version: %w", err)
	}

	minor, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse minor version: %w", err)
	}

	patch, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse patch version: %w", err)
	}

	return Version{
		Major: uint32(major),
		Minor: uint32(minor),
		Patch: uint32(patch),
	}, nil
}

// Compare returns:
//   - -1 if v < other
//   - 0 if v == other
//   - 1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// LessThan returns true if v < other.
func (v Version) LessThan(other Version) bool {
	return v.Compare(other) < 0
}

// GreaterThan returns true if v > other.
func (v Version) GreaterThan(other Version) bool {
	return v.Compare(other) > 0
}

// Equal returns true if v == other.
func (v Version) Equal(other Version) bool {
	return v.Compare(other) == 0
}
