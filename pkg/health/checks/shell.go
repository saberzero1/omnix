package checks

import (
	"context"
	"os"

	"github.com/juspay/omnix/pkg/nix"
)

// Shell checks shell configuration
type Shell struct{}

// Check verifies shell configuration
func (s *Shell) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	shell := os.Getenv("SHELL")

	if shell == "" {
		// No shell set, skip check
		return []NamedCheck{}
	}

	// For now, just report the shell being used
	// TODO: Add more sophisticated shell checks (e.g., nix integration)
	result := GreenResult{}

	check := Check{
		Title:    "Shell Configuration",
		Info:     "SHELL = " + shell,
		Result:   result,
		Required: false,
	}

	return []NamedCheck{
		{Name: "shell", Check: check},
	}
}
