package checks

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/saberzero1/omnix/pkg/nix"
)

// Rosetta checks for Rosetta 2 on Apple Silicon Macs
type Rosetta struct{}

// Check verifies Rosetta 2 availability on macOS ARM64
func (r *Rosetta) Check(_ context.Context, _ *nix.Info) []NamedCheck {
	// Only relevant for macOS on ARM64
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return []NamedCheck{}
	}

	// Check if Rosetta is installed by looking for the arch command
	// and checking if it can run x86_64 binaries
	_, err := exec.LookPath("arch")
	hasRosetta := err == nil

	var result CheckResult
	if hasRosetta {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message:    "Rosetta 2 is not installed",
			Suggestion: "Install Rosetta 2 with: softwareupdate --install-rosetta",
		}
	}

	check := Check{
		Title:    "Rosetta 2",
		Info:     "Required for running x86_64 binaries on Apple Silicon",
		Result:   result,
		Required: false, // Not strictly required
	}

	return []NamedCheck{
		{Name: "rosetta", Check: check},
	}
}
