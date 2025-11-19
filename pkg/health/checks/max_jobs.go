package checks

import (
	"context"

	"github.com/saberzero1/omnix/pkg/nix"
)

// MaxJobs checks the max-jobs configuration
type MaxJobs struct {
	// TODO: Add configurable minimum or recommended values
}

// Check verifies the max-jobs setting
func (mj *MaxJobs) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	// TODO: Implement max-jobs check
	// This would check nixInfo.Config for max-jobs setting
	// For now, skip this check
	return []NamedCheck{}
}
