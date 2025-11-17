package nix

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Env represents the environment in which Nix operates.
type Env struct {
	// CurrentUser is the current user ($USER)
	CurrentUser string
	// CurrentUserGroups are the current user's groups
	CurrentUserGroups []string
	// OS is the underlying operating system
	OS OSType
}

// OSType represents the operating system type.
type OSType struct {
	// Type is the OS type (e.g., "darwin", "linux")
	Type string
	// IsNixOS indicates if this is NixOS
	IsNixOS bool
	// IsNixDarwin indicates if this is nix-darwin on macOS
	IsNixDarwin bool
	// Arch is the architecture (e.g., "amd64", "arm64")
	Arch string
}

// String returns a human-readable string representation of the OS.
func (o OSType) String() string {
	switch {
	case o.IsNixOS:
		return "NixOS"
	case o.IsNixDarwin:
		return "macOS (nix-darwin)"
	case o.Type == "darwin":
		return "macOS"
	case o.Type == "linux":
		return "Linux"
	default:
		return o.Type
	}
}

// NixConfigLabel returns the label for where Nix is configured.
func (o OSType) NixConfigLabel() string {
	if o.IsNixOS {
		return "nixos configuration"
	}
	if o.IsNixDarwin {
		return "nix-darwin configuration"
	}
	return "/etc/nix/nix.conf"
}

// DetectEnv detects the Nix environment on the current system.
func DetectEnv(ctx context.Context) (*Env, error) {
	currentUser := getCurrentUser()
	groups, err := getCurrentUserGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	
	osType, err := detectOS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}
	
	return &Env{
		CurrentUser:       currentUser,
		CurrentUserGroups: groups,
		OS:                osType,
	}, nil
}

// getCurrentUser returns the current username.
func getCurrentUser() string {
	// Try USER environment variable first
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	// Try USERNAME on Windows
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	// Fallback to empty string
	return ""
}

// getCurrentUserGroups returns the current user's groups.
func getCurrentUserGroups(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "groups")
	output, err := cmd.Output()
	if err != nil {
		// If groups command fails, return empty list rather than error
		// This can happen on some systems
		return []string{}, nil
	}
	
	groupsStr := strings.TrimSpace(string(output))
	if groupsStr == "" {
		return []string{}, nil
	}
	
	groups := strings.Fields(groupsStr)
	return groups, nil
}

// detectOS detects the operating system type.
func detectOS(ctx context.Context) (OSType, error) {
	osType := OSType{
		Type: runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	
	// Check for NixOS
	if _, err := os.Stat("/etc/NIXOS"); err == nil {
		osType.IsNixOS = true
		return osType, nil
	}
	
	// Check for nix-darwin on macOS
	if runtime.GOOS == "darwin" {
		// Check if /etc/nix/nix.conf is a symlink (managed by nix-darwin)
		info, err := os.Lstat("/etc/nix/nix.conf")
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			osType.IsNixDarwin = true
		}
	}
	
	return osType, nil
}
