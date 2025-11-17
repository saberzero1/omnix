package checks

import (
	"context"
	"fmt"

	"github.com/juspay/omnix/pkg/nix"
)

// FlakeEnabled checks that experimental features include flakes
type FlakeEnabled struct{}

// Check verifies that flakes are enabled in the Nix configuration
func (f *FlakeEnabled) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	features := nixInfo.Config.ExperimentalFeatures.Value
	
	hasFlakes := false
	hasNixCommand := false
	
	for _, feature := range features {
		if feature == "flakes" {
			hasFlakes = true
		}
		if feature == "nix-command" {
			hasNixCommand = true
		}
	}
	
	var result CheckResult
	if hasFlakes && hasNixCommand {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message:    "Nix flakes are not enabled",
			Suggestion: "See https://nixos.wiki/wiki/Flakes#Enable_flakes",
		}
	}
	
	check := Check{
		Title:    "Flakes Enabled",
		Info:     fmt.Sprintf("experimental-features = %v", features),
		Result:   result,
		Required: true,
	}
	
	return []NamedCheck{
		{Name: "flake-enabled", Check: check},
	}
}
