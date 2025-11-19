package checks

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/saberzero1/omnix/pkg/nix"
)

// Homebrew checks for Homebrew on macOS
type Homebrew struct{}

// Check verifies Homebrew installation on macOS
func (h *Homebrew) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	// Only relevant for macOS
	if runtime.GOOS != "darwin" {
		return []NamedCheck{}
	}

	_, err := exec.LookPath("brew")
	hasHomebrew := err == nil

	var result CheckResult
	if hasHomebrew {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message:    "Homebrew is not installed",
			Suggestion: "Install Homebrew from https://brew.sh/",
		}
	}

	check := Check{
		Title:    "Homebrew",
		Info:     "Homebrew package manager for macOS",
		Result:   result,
		Required: false, // Optional
	}

	return []NamedCheck{
		{Name: "homebrew", Check: check},
	}
}
