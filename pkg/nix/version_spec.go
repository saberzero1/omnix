package nix

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionSpec represents an individual component of a version requirement
type VersionSpec struct {
	operator string
	version  Version
}

// VersionSpecType represents the type of version comparison
type VersionSpecType int

const (
	// VersionSpecGt requires version greater than specified
	VersionSpecGt VersionSpecType = iota
	// VersionSpecGte requires version greater than or equal to specified
	VersionSpecGte
	// VersionSpecLt requires version less than specified
	VersionSpecLt
	// VersionSpecLte requires version less than or equal to specified
	VersionSpecLte
	// VersionSpecNeq requires version not equal to specified
	VersionSpecNeq
)

// NewVersionSpec creates a new version specification
func NewVersionSpec(op VersionSpecType, version Version) *VersionSpec {
	opStr := ""
	switch op {
	case VersionSpecGt:
		opStr = ">"
	case VersionSpecGte:
		opStr = ">="
	case VersionSpecLt:
		opStr = "<"
	case VersionSpecLte:
		opStr = "<="
	case VersionSpecNeq:
		opStr = "!="
	}
	return &VersionSpec{
		operator: opStr,
		version:  version,
	}
}

// Matches checks if a given Nix version satisfies this version specification
func (s *VersionSpec) Matches(version Version) bool {
	switch s.operator {
	case ">":
		return version.GreaterThan(s.version)
	case ">=":
		return version.GreaterThan(s.version) || version.Equal(s.version)
	case "<":
		return version.LessThan(s.version)
	case "<=":
		return version.LessThan(s.version) || version.Equal(s.version)
	case "!=":
		return !version.Equal(s.version)
	default:
		return false
	}
}

// String returns the string representation of the version spec
func (s *VersionSpec) String() string {
	return fmt.Sprintf("%s%s", s.operator, s.version.String())
}

// ParseVersionSpec parses a version specification string like ">=2.8"
func ParseVersionSpec(s string) (*VersionSpec, error) {
	re := regexp.MustCompile(`^(>=|<=|>|<|!=)(\d+)(?:\.(\d+))?(?:\.(\d+))?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid version spec format: %s", s)
	}

	op := matches[1]
	major, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	var minor, patch uint64
	if matches[3] != "" {
		minor, err = strconv.ParseUint(matches[3], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %w", err)
		}
	}
	if matches[4] != "" {
		patch, err = strconv.ParseUint(matches[4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %w", err)
		}
	}

	version := Version{
		Major: uint32(major),
		Minor: uint32(minor),
		Patch: uint32(patch),
	}

	return &VersionSpec{
		operator: op,
		version:  version,
	}, nil
}

// VersionReq represents a version requirement for Nix
// Example: ">=2.8, <2.14, !=2.13.4"
type VersionReq struct {
	Specs []*VersionSpec
}

// ParseVersionReq parses a version requirement string
func ParseVersionReq(s string) (*VersionReq, error) {
	parts := strings.Split(s, ",")
	specs := make([]*VersionSpec, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		spec, err := ParseVersionSpec(part)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return &VersionReq{Specs: specs}, nil
}

// Matches checks if a version satisfies all specifications in the requirement
func (r *VersionReq) Matches(version Version) bool {
	for _, spec := range r.Specs {
		if !spec.Matches(version) {
			return false
		}
	}
	return true
}

// String returns the string representation of the version requirement
func (r *VersionReq) String() string {
	strs := make([]string, len(r.Specs))
	for i, spec := range r.Specs {
		strs[i] = spec.String()
	}
	return strings.Join(strs, ", ")
}
