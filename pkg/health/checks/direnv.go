package checks

import (
	"context"
	"os/exec"

	"github.com/saberzero1/omnix/pkg/nix"
)

// Direnv checks for direnv installation
type Direnv struct{}

// Check verifies that direnv is installed
func (d *Direnv) Check(_ context.Context, _ *nix.Info) []NamedCheck {
	_, err := exec.LookPath("direnv")
	hasDirenv := err == nil

	var result CheckResult
	if hasDirenv {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message:    "direnv is not installed",
			Suggestion: "Install direnv from https://direnv.net/",
		}
	}

	check := Check{
		Title:    "Direnv",
		Info:     "direnv provides automatic directory-specific environment management",
		Result:   result,
		Required: false, // Optional but recommended
	}

	return []NamedCheck{
		{Name: "direnv", Check: check},
	}
}
