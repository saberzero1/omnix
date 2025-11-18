package checks

import (
	"context"
	"fmt"

	"github.com/juspay/omnix/pkg/nix"
)

// NixVersion checks that the Nix version meets minimum requirements
type NixVersion struct {
	// MinVersion specifies the minimum supported version (default: 2.16.0)
	MinVersion nix.Version `yaml:"min-version" json:"min-version"`
}

// DefaultNixVersion returns a NixVersion with the default minimum version
func DefaultNixVersion() NixVersion {
	return NixVersion{
		MinVersion: nix.Version{Major: 2, Minor: 16, Patch: 0},
	}
}

// Check verifies that the installed Nix version is supported
func (nv *NixVersion) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	currentVersion := nixInfo.Version

	var result CheckResult
	if currentVersion.GreaterThan(nv.MinVersion) || currentVersion.Equal(nv.MinVersion) {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message: fmt.Sprintf(
				"Your Nix version (%s) doesn't satisfy the supported bounds: >=%s",
				currentVersion.String(),
				nv.MinVersion.String(),
			),
			Suggestion: "To use a specific version of Nix, see <https://nixos.asia/en/howto/nix-package>",
		}
	}

	check := Check{
		Title:    "Nix Version is supported",
		Info:     fmt.Sprintf("nix version = %s", currentVersion.String()),
		Result:   result,
		Required: true,
	}

	return []NamedCheck{
		{Name: "supported-nix-versions", Check: check},
	}
}
