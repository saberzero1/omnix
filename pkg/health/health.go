package health

import (
	"context"
	"encoding/json"
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

// Health check result constants
const (
	// Pass indicates all checks passed
	Pass AllChecksResult = iota
	// PassSomeFail indicates required checks passed, some non-required failed
	PassSomeFail
	// Fail indicates some required checks failed
	Fail
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

// PrintCheckResult prints a single check result with markdown formatting
func PrintCheckResult(nc checks.NamedCheck) error {
	// Build markdown output
	var md string
	if nc.Check.Required {
		md = fmt.Sprintf("### %s (Required)\n\n", nc.Check.Title)
	} else {
		md = fmt.Sprintf("### %s\n\n", nc.Check.Title)
	}

	md += fmt.Sprintf("%s\n\n", nc.Check.Info)
	md += fmt.Sprintf("**Result**: %s\n", nc.Check.Result.String())

	// Print without markdown rendering for now (simpler output)
	fmt.Println(md)
	return nil
}

// PrintCheckResultMarkdown prints a single check result with rendered markdown
// Note: This function is currently identical to PrintCheckResult.
// Actual markdown rendering using common.RenderMarkdown would create import cycles.
// For future enhancement, consider restructuring packages to enable proper rendering.
func PrintCheckResultMarkdown(nc checks.NamedCheck) error {
	return PrintCheckResult(nc)
}

// ResultsToJSON converts health check results to JSON format
func ResultsToJSON(checkList []checks.NamedCheck, result AllChecksResult, nixInfo *nix.Info) (string, error) {
	type CheckJSON struct {
		Name     string `json:"name"`
		Title    string `json:"title"`
		Info     string `json:"info"`
		Required bool   `json:"required"`
		Success  bool   `json:"success"`
		Message  string `json:"message,omitempty"`
	}

	type OutputJSON struct {
		System      string      `json:"system"`
		NixVersion  string      `json:"nix_version"`
		Status      string      `json:"status"`
		ExitCode    int         `json:"exit_code"`
		Summary     string      `json:"summary"`
		Checks      []CheckJSON `json:"checks"`
		PassedCount int         `json:"passed_count"`
		FailedCount int         `json:"failed_count"`
	}

	var output OutputJSON
	output.System = nixInfo.Env.OS.String()
	output.NixVersion = nixInfo.Version.String()
	output.ExitCode = result.ExitCode()
	output.Summary = result.SummaryMessage()

	switch result {
	case Pass:
		output.Status = "pass"
	case PassSomeFail:
		output.Status = "pass_with_warnings"
	case Fail:
		output.Status = "fail"
	}

	output.Checks = make([]CheckJSON, 0, len(checkList))
	passedCount := 0
	failedCount := 0

	for _, nc := range checkList {
		checkJSON := CheckJSON{
			Name:     nc.Name,
			Title:    nc.Check.Title,
			Info:     nc.Check.Info,
			Required: nc.Check.Required,
			Success:  nc.Check.Result.IsGreen(),
		}

		if !nc.Check.Result.IsGreen() {
			checkJSON.Message = nc.Check.Result.String()
			failedCount++
		} else {
			passedCount++
		}

		output.Checks = append(output.Checks, checkJSON)
	}

	output.PassedCount = passedCount
	output.FailedCount = failedCount

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}
