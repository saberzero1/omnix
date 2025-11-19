// Package checks provides individual health checks for Nix installations
package checks

import (
	"context"
	"fmt"

	"github.com/saberzero1/omnix/pkg/nix"
)

// CheckResult represents the result of a health check
type CheckResult interface {
	// IsGreen returns true if the check passed
	IsGreen() bool
	// String returns a string representation of the result
	String() string
}

// GreenResult indicates a passed check
type GreenResult struct{}

func (g GreenResult) IsGreen() bool  { return true }
func (g GreenResult) String() string { return "✅ Passed" }

// RedResult indicates a failed check
type RedResult struct {
	Message    string // Problem description
	Suggestion string // How to fix the problem
}

func (r RedResult) IsGreen() bool { return false }
func (r RedResult) String() string {
	return fmt.Sprintf("❌ Failed: %s. Fix: %s", r.Message, r.Suggestion)
}

// Check represents a single health check
type Check struct {
	// Title is a user-facing title of this check
	Title string `json:"title"`

	// Info contains user-facing information used to conduct this check
	Info string `json:"info"`

	// Result is the result of running this check
	Result CheckResult `json:"result"`

	// Required indicates whether this check is mandatory
	// Failures are considered non-critical if this is false
	Required bool `json:"required"`
}

// NamedCheck is a Check with a unique identifier
type NamedCheck struct {
	Name  string
	Check Check
}

// Checkable is the interface for types that can perform health checks
type Checkable interface {
	// Check runs the health check and returns zero or more named checks
	// Returning an empty slice indicates that the check is skipped on this environment
	Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck
}
