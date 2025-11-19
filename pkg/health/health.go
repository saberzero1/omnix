package health

import (
	"context"
	"fmt"

	"github.com/saberzero1/omnix/pkg/health/checks"
	"github.com/saberzero1/omnix/pkg/nix"
)

// NixHealth contains all health checks for a Nix installation
type NixHealth struct {
	FlakeEnabled checks.FlakeEnabled `yaml:"flake-enabled" json:"flake-enabled"`
	NixVersion   checks.NixVersion   `yaml:"nix-version" json:"nix-version"`
	TrustedUsers checks.TrustedUsers `yaml:"trusted-users" json:"trusted-users"`
	Caches       checks.Caches       `yaml:"caches" json:"caches"`
	MaxJobs      checks.MaxJobs      `yaml:"max-jobs" json:"max-jobs"`
	Rosetta      checks.Rosetta      `yaml:"rosetta" json:"rosetta"`
	Direnv       checks.Direnv       `yaml:"direnv" json:"direnv"`
	Homebrew     checks.Homebrew     `yaml:"homebrew" json:"homebrew"`
	Shell        checks.Shell        `yaml:"shell" json:"shell"`
}

// Default returns a NixHealth with default check configurations
func Default() *NixHealth {
	return &NixHealth{
		FlakeEnabled: checks.FlakeEnabled{},
		NixVersion:   checks.DefaultNixVersion(),
		TrustedUsers: checks.TrustedUsers{Enable: false}, // Disabled by default for security
		Caches:       checks.DefaultCaches(),
		MaxJobs:      checks.MaxJobs{},
		Rosetta:      checks.Rosetta{},
		Direnv:       checks.Direnv{},
		Homebrew:     checks.Homebrew{},
		Shell:        checks.Shell{},
	}
}

// RunAllChecks executes all health checks and returns the results
func (h *NixHealth) RunAllChecks(ctx context.Context, nixInfo *nix.Info) []checks.NamedCheck {
	var results []checks.NamedCheck

	// Collect all checks in order
	checkables := []checks.Checkable{
		&h.FlakeEnabled,
		&h.NixVersion,
		&h.Rosetta,
		&h.MaxJobs,
		&h.TrustedUsers,
		&h.Caches,
		&h.Direnv,
		&h.Homebrew,
		&h.Shell,
	}

	// Run each check and collect results
	for _, checkable := range checkables {
		checkResults := checkable.Check(ctx, nixInfo)
		results = append(results, checkResults...)
	}

	return results
}

// AllChecksResult aggregates check results and provides summary reporting
type AllChecksResult int

const (
	Pass         AllChecksResult = iota // All checks passed
	PassSomeFail                        // Required checks passed, some non-required failed
	Fail                                // Some required checks failed
)

// RegisterFailure updates the result based on a failed check
func (r *AllChecksResult) RegisterFailure(required bool) {
	if required {
		*r = Fail
	} else if *r == Pass {
		*r = PassSomeFail
	}
}

// ExitCode returns the appropriate exit code for the result
func (r AllChecksResult) ExitCode() int {
	switch r {
	case Pass, PassSomeFail:
		return 0
	case Fail:
		return 1
	default:
		return 1
	}
}

// SummaryMessage returns a human-readable summary message
func (r AllChecksResult) SummaryMessage() string {
	switch r {
	case Pass:
		return "✅ All checks passed"
	case PassSomeFail:
		return "✅ Required checks passed, but some non-required checks failed"
	case Fail:
		return "❌ Some required checks failed"
	default:
		return "❌ Unknown result"
	}
}

// EvaluateResults processes all checks and returns the aggregated result
func EvaluateResults(checkList []checks.NamedCheck) AllChecksResult {
	result := Pass
	for _, nc := range checkList {
		if !nc.Check.Result.IsGreen() {
			result.RegisterFailure(nc.Check.Required)
		}
	}
	return result
}

// PrintCheckResult prints a single check result (placeholder for now)
func PrintCheckResult(nc checks.NamedCheck) error {
	// TODO: Add markdown rendering support
	fmt.Printf("%s: %s\n", nc.Check.Title, nc.Check.Info)
	fmt.Printf("  Result: %s\n", nc.Check.Result.String())
	return nil
}
