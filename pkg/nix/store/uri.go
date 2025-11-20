package store

import (
	"fmt"
	"net/url"
	"strings"
)

// URI represents a Nix store URI.
// Currently supports SSH stores.
type URI struct {
	scheme  string
	sshURI  *SSHURI
	options Options
}

// SSHURI represents a remote SSH store URI.
type SSHURI struct {
	User string
	Host string
}

// Options contains user-passed options for a store URI.
type Options struct {
	// CopyInputs determines whether to copy all flake inputs recursively.
	// If disabled, we copy only the flake source itself.
	// Enabling this option is useful when there are private Git inputs
	// but the target machine does not have access to them.
	CopyInputs bool
}

// ParseURI parses a Nix store URI string.
// Currently only supports ssh:// scheme.
func ParseURI(uriStr string) (*URI, error) {
	u, err := url.Parse(uriStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %w", err)
	}

	switch u.Scheme {
	case "ssh":
		return parseSSHURI(u)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}

func parseSSHURI(u *url.URL) (*URI, error) {
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}

	sshURI := &SSHURI{
		Host: u.Host,
	}

	if u.User != nil {
		sshURI.User = u.User.Username()
	}

	// Parse query parameters for options
	opts := Options{}
	query := u.Query()
	if query.Get("copy-inputs") == "true" {
		opts.CopyInputs = true
	}

	return &URI{
		scheme:  "ssh",
		sshURI:  sshURI,
		options: opts,
	}, nil
}

// String returns the string representation of the URI.
func (u *URI) String() string {
	switch u.scheme {
	case "ssh":
		return u.sshString()
	default:
		return ""
	}
}

func (u *URI) sshString() string {
	if u.sshURI == nil {
		return ""
	}
	
	var builder strings.Builder
	builder.WriteString("ssh://")
	
	if u.sshURI.User != "" {
		builder.WriteString(u.sshURI.User)
		builder.WriteString("@")
	}
	
	builder.WriteString(u.sshURI.Host)
	
	return builder.String()
}

// GetOptions returns the options for this store URI.
func (u *URI) GetOptions() Options {
	return u.options
}

// IsSSH returns true if this is an SSH store URI.
func (u *URI) IsSSH() bool {
	return u.scheme == "ssh"
}

// GetSSHURI returns the SSH URI details if this is an SSH store.
func (u *URI) GetSSHURI() *SSHURI {
	return u.sshURI
}

// String returns the string representation of an SSH URI.
func (s *SSHURI) String() string {
	if s.User != "" {
		return fmt.Sprintf("%s@%s", s.User, s.Host)
	}
	return s.Host
}
