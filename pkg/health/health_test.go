package health

import (
	"context"
	"testing"

	"github.com/saberzero1/omnix/pkg/health/checks"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
)

func TestGreenResult(t *testing.T) {
	result := checks.GreenResult{}
	assert.True(t, result.IsGreen())
	assert.Contains(t, result.String(), "Passed")
}

func TestRedResult(t *testing.T) {
	result := checks.RedResult{
		Message:    "Test failed",
		Suggestion: "Fix it",
	}
	assert.False(t, result.IsGreen())
	assert.Contains(t, result.String(), "Test failed")
	assert.Contains(t, result.String(), "Fix it")
}

func TestDefault(t *testing.T) {
	health := Default()
	assert.NotNil(t, health)
	assert.False(t, health.TrustedUsers.Enable) // Should be disabled by default
}

func TestAllChecksResult_RegisterFailure(t *testing.T) {
	tests := []struct {
		name            string
		initialState    AllChecksResult
		failureRequired bool
		expectedState   AllChecksResult
	}{
		{
			name:            "Pass to Fail on required failure",
			initialState:    Pass,
			failureRequired: true,
			expectedState:   Fail,
		},
		{
			name:            "Pass to PassSomeFail on non-required failure",
			initialState:    Pass,
			failureRequired: false,
			expectedState:   PassSomeFail,
		},
		{
			name:            "PassSomeFail to Fail on required failure",
			initialState:    PassSomeFail,
			failureRequired: true,
			expectedState:   Fail,
		},
		{
			name:            "PassSomeFail stays on non-required failure",
			initialState:    PassSomeFail,
			failureRequired: false,
			expectedState:   PassSomeFail,
		},
		{
			name:            "Fail stays Fail",
			initialState:    Fail,
			failureRequired: false,
			expectedState:   Fail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initialState
			result.RegisterFailure(tt.failureRequired)
			assert.Equal(t, tt.expectedState, result)
		})
	}
}

func TestAllChecksResult_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		result   AllChecksResult
		expected int
	}{
		{"Pass returns 0", Pass, 0},
		{"PassSomeFail returns 0", PassSomeFail, 0},
		{"Fail returns 1", Fail, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ExitCode())
		})
	}
}

func TestEvaluateResults(t *testing.T) {
	tests := []struct {
		name     string
		checks   []checks.NamedCheck
		expected AllChecksResult
	}{
		{
			name: "All green checks",
			checks: []checks.NamedCheck{
				{
					Name: "check1",
					Check: checks.Check{
						Title:    "Test 1",
						Result:   checks.GreenResult{},
						Required: true,
					},
				},
				{
					Name: "check2",
					Check: checks.Check{
						Title:    "Test 2",
						Result:   checks.GreenResult{},
						Required: false,
					},
				},
			},
			expected: Pass,
		},
		{
			name: "Non-required red check",
			checks: []checks.NamedCheck{
				{
					Name: "check1",
					Check: checks.Check{
						Title:    "Test 1",
						Result:   checks.GreenResult{},
						Required: true,
					},
				},
				{
					Name: "check2",
					Check: checks.Check{
						Title:    "Test 2",
						Result:   checks.RedResult{Message: "fail", Suggestion: "fix"},
						Required: false,
					},
				},
			},
			expected: PassSomeFail,
		},
		{
			name: "Required red check",
			checks: []checks.NamedCheck{
				{
					Name: "check1",
					Check: checks.Check{
						Title:    "Test 1",
						Result:   checks.RedResult{Message: "fail", Suggestion: "fix"},
						Required: true,
					},
				},
			},
			expected: Fail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateResults(tt.checks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunAllChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Get real nix info
	nixInfo, err := nix.GetInfo(ctx)
	if err != nil {
		t.Skipf("Nix not available: %v", err)
	}

	health := Default()
	checkResults := health.RunAllChecks(ctx, nixInfo)

	// Should have at least some checks
	assert.NotEmpty(t, checkResults)

	// Each check should have a name and check
	for _, nc := range checkResults {
		assert.NotEmpty(t, nc.Name)
		assert.NotEmpty(t, nc.Check.Title)
	}
}

func TestPrintCheckResultMarkdown(t *testing.T) {
	// Test with a green result
	greenCheck := checks.NamedCheck{
		Name: "test-check",
		Check: checks.Check{
			Title:    "Test Check",
			Info:     "Test info",
			Result:   checks.GreenResult{},
			Required: true,
		},
	}
	
	err := PrintCheckResultMarkdown(greenCheck)
	assert.NoError(t, err)
	
	// Test with a red result
	redCheck := checks.NamedCheck{
		Name: "fail-check",
		Check: checks.Check{
			Title:    "Failing Check",
			Info:     "Fail info",
			Result:   checks.RedResult{Message: "Failed", Suggestion: "Fix it"},
			Required: false,
		},
	}
	
	err = PrintCheckResultMarkdown(redCheck)
	assert.NoError(t, err)
}
