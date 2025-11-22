package nix

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/saberzero1/omnix/pkg/nix/flake"
)

// FlakeURL represents a Nix flake URL.
// See https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake.html#url-like-syntax
type FlakeURL struct {
	url string
}

// NewFlakeURL creates a new FlakeURL from a string.
func NewFlakeURL(url string) FlakeURL {
	return FlakeURL{url: url}
}

// String returns the string representation of the flake URL.
func (f FlakeURL) String() string {
	return f.url
}

// AsLocalPath returns the local path if the flake URL is a local path.
// Applicable only if the flake URL uses the Path-like syntax.
// Returns empty string if not a local path.
func (f FlakeURL) AsLocalPath() string {
	s := f.url

	// Strip "path:" prefix if present
	s = strings.TrimPrefix(s, "path:")

	// Check if it's a local path (starts with . or /)
	if !strings.HasPrefix(s, ".") && !strings.HasPrefix(s, "/") {
		return ""
	}

	// Strip query parameters (?...) and attributes (#...)
	if idx := strings.IndexByte(s, '?'); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.IndexByte(s, '#'); idx != -1 {
		s = s[:idx]
	}

	return s
}

// IsLocal returns true if the flake URL points to a local path.
func (f FlakeURL) IsLocal() bool {
	return f.AsLocalPath() != ""
}

// WithAttr returns a new FlakeURL with the given attribute appended.
// For example, WithAttr("packages.x86_64-linux.default")
func (f FlakeURL) WithAttr(attr string) FlakeURL {
	// Remove existing attribute if present
	url := f.url
	if idx := strings.IndexByte(url, '#'); idx != -1 {
		url = url[:idx]
	}

	// Append new attribute
	if attr != "" {
		url = url + "#" + attr
	}

	return FlakeURL{url: url}
}

// SplitAttr splits the flake URL into the base URL and attribute.
// Returns (baseURL, attr) where attr may be empty.
func (f FlakeURL) SplitAttr() (string, string) {
	if idx := strings.IndexByte(f.url, '#'); idx != -1 {
		return f.url[:idx], f.url[idx+1:]
	}
	return f.url, ""
}

// GetAttr returns the Attr part of the FlakeURL.
func (f FlakeURL) GetAttr() flake.Attr {
	_, attrStr := f.SplitAttr()
	if attrStr == "" {
		return flake.NoneAttr()
	}
	return flake.NewAttr(attrStr)
}

// WithoutAttr returns a new FlakeURL without the attribute part.
func (f FlakeURL) WithoutAttr() FlakeURL {
	base, _ := f.SplitAttr()
	return FlakeURL{url: base}
}

// SubFlakeURL returns a FlakeURL pointing to a sub-flake at the given directory.
// For local paths, it joins the directory path.
// For non-local URLs, it appends a "?dir=" query parameter.
// If dir is ".", returns the current FlakeURL unchanged.
func (f FlakeURL) SubFlakeURL(dir string) FlakeURL {
	if dir == "." {
		return f
	}

	if localPath := f.AsLocalPath(); localPath != "" {
		// Local path: join the directory
		joined := filepath.Join(localPath, dir)
		return FlakeURL{url: joined}
	}

	// Non-path URL: append dir query parameter
	url := f.url
	if strings.Contains(url, "?") {
		url += "&dir=" + dir
	} else {
		url += "?dir=" + dir
	}
	return FlakeURL{url: url}
}

// Clean returns a cleaned version of the flake URL.
// For local paths, it resolves relative paths.
func (f FlakeURL) Clean() FlakeURL {
	if localPath := f.AsLocalPath(); localPath != "" {
		// Clean the local path
		cleaned := filepath.Clean(localPath)

		// Preserve the attribute if it exists
		_, attr := f.SplitAttr()
		if attr != "" {
			cleaned = cleaned + "#" + attr
		}

		return FlakeURL{url: cleaned}
	}
	return f
}

// ParseFlakeURL parses a string into a FlakeURL.
// This is a simple wrapper around NewFlakeURL for now.
func ParseFlakeURL(s string) (FlakeURL, error) {
	if s == "" {
		return FlakeURL{}, fmt.Errorf("empty flake URL")
	}
	return NewFlakeURL(s), nil
}
