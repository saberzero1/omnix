package health

import (
	"testing"

	"github.com/juspay/omnix/pkg/health/checks"
	"github.com/stretchr/testify/assert"
)

func TestAllChecksResult_SummaryMessage(t *testing.T) {
	tests := []struct {
		name     string
		result   AllChecksResult
		expected string
	}{
		{
			name:     "Pass",
			result:   Pass,
			expected: "✅ All checks passed",
		},
		{
			name:     "PassSomeFail",
			result:   PassSomeFail,
			expected: "✅ Required checks passed, but some non-required checks failed",
		},
		{
			name:     "Fail",
			result:   Fail,
			expected: "❌ Some required checks failed",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.result.SummaryMessage()
			assert.Equal(t, tt.expected, msg)
		})
	}
}

func TestAllChecksResult_ExitCode_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		result   AllChecksResult
		expected int
	}{
		{"Pass", Pass, 0},
		{"PassSomeFail", PassSomeFail, 0},
		{"Fail", Fail, 1},
		{"Invalid", AllChecksResult(999), 1}, // Unknown state defaults to 1
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ExitCode())
		})
	}
}

func TestPrintCheckResult(t *testing.T) {
	tests := []struct {
		name  string
		check checks.NamedCheck
	}{
		{
			name: "Green check",
			check: checks.NamedCheck{
				Name: "test-green",
				Check: checks.Check{
					Title:    "Test Green",
					Info:     "This is a green check",
					Result:   checks.GreenResult{},
					Required: true,
				},
			},
		},
		{
			name: "Red check",
			check: checks.NamedCheck{
				Name: "test-red",
				Check: checks.Check{
					Title: "Test Red",
					Info:  "This is a red check",
					Result: checks.RedResult{
						Message:    "Something went wrong",
						Suggestion: "Fix it this way",
					},
					Required: false,
				},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			err := PrintCheckResult(tt.check)
			assert.NoError(t, err)
		})
	}
}
