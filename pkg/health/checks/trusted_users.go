package checks

import (
	"context"
	"fmt"

	"github.com/saberzero1/omnix/pkg/nix"
)

// TrustedUsers checks that the current user is in trusted-users
type TrustedUsers struct {
	// Enable controls whether this check runs (disabled by default for security)
	// See https://github.com/saberzero1/omnix/issues/409
	Enable bool `yaml:"enable" json:"enable"`
}

// Check verifies that the current user is a trusted user
func (tu *TrustedUsers) Check(_ context.Context, nixInfo *nix.Info) []NamedCheck {
	// Skip if disabled
	if !tu.Enable {
		return []NamedCheck{}
	}

	// Use the new User field directly (preferred over deprecated CurrentUser() method)
	currentUser := nixInfo.Env.User
	userGroups := make(map[string]bool)
	for _, group := range nixInfo.Env.Groups {
		userGroups[group] = true
	}

	// TODO: Parse trusted-users from nix config
	// For now, we'll assume it's not set
	isTrusted := false
	trustedUsersStr := "root" // Placeholder

	var result CheckResult
	if isTrusted {
		result = GreenResult{}
	} else {
		msg := fmt.Sprintf("User '%s' not present in trusted_users", currentUser)

		var suggestion string
		if configLabel := nixInfo.Env.OS.NixSystemConfigLabel(); configLabel != "" {
			suggestion = fmt.Sprintf(
				`Add 'nix.trustedUsers = [ "root" "%s" ];' to your %s`,
				currentUser, configLabel,
			)
		} else {
			suggestion = fmt.Sprintf(
				"Set 'trusted-users = root %s' in /etc/nix/nix.conf and then restart the Nix daemon using `sudo pkill nix-daemon`",
				currentUser,
			)
		}

		result = RedResult{
			Message:    msg,
			Suggestion: suggestion,
		}
	}

	check := Check{
		Title:    "Trusted Users",
		Info:     fmt.Sprintf("trusted-users = %s", trustedUsersStr),
		Result:   result,
		Required: true,
	}

	return []NamedCheck{
		{Name: "trusted-users", Check: check},
	}
}
