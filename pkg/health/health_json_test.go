package health

import (
	"encoding/json"
	"testing"

	"github.com/saberzero1/omnix/pkg/health/checks"
	"github.com/saberzero1/omnix/pkg/nix"
)

func TestResultsToJSON(t *testing.T) {
	// Create mock nix info
	nixInfo := &nix.Info{
		Version: nix.Version{Major: 2, Minor: 18, Patch: 1},
		Env: &nix.Env{
			OS: nix.OSType{
				Type: "linux",
			},
		},
	}

	// Create mock check results
	results := []checks.NamedCheck{
		{
			Name: "test-check-pass",
			Check: checks.Check{
				Title:    "Test Check Pass",
				Info:     "This is a passing test",
				Result:   checks.GreenResult{},
				Required: true,
			},
		},
		{
			Name: "test-check-fail",
			Check: checks.Check{
				Title:    "Test Check Fail",
				Info:     "This is a failing test",
				Result:   checks.RedResult{Message: "Failed", Suggestion: "Fix it"},
				Required: false,
			},
		},
	}

	status := EvaluateResults(results)

	jsonOutput, err := ResultsToJSON(results, status, nixInfo)
	if err != nil {
		t.Fatalf("ResultsToJSON() failed: %v", err)
	}

	// Verify it's valid JSON
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify expected fields
	systemStr, ok := output["system"].(string)
	if !ok || systemStr == "" {
		t.Errorf("Expected non-empty system string, got '%v'", output["system"])
	}

	if output["nix_version"] != "2.18.1" {
		t.Errorf("Expected nix_version '2.18.1', got '%v'", output["nix_version"])
	}

	if output["status"] != "pass_with_warnings" {
		t.Errorf("Expected status 'pass_with_warnings', got '%v'", output["status"])
	}

	if output["passed_count"] != float64(1) {
		t.Errorf("Expected passed_count 1, got %v", output["passed_count"])
	}

	if output["failed_count"] != float64(1) {
		t.Errorf("Expected failed_count 1, got %v", output["failed_count"])
	}

	// Verify checks array
	checksArray, ok := output["checks"].([]interface{})
	if !ok {
		t.Fatal("checks is not an array")
	}

	if len(checksArray) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(checksArray))
	}
}

func TestResultsToJSON_AllPass(t *testing.T) {
	nixInfo := &nix.Info{
		Version: nix.Version{Major: 2, Minor: 19, Patch: 0},
		Env: &nix.Env{
			OS: nix.OSType{
				Type: "darwin",
			},
		},
	}

	results := []checks.NamedCheck{
		{
			Name: "test-check-1",
			Check: checks.Check{
				Title:    "Test Check 1",
				Info:     "All good",
				Result:   checks.GreenResult{},
				Required: true,
			},
		},
	}

	status := EvaluateResults(results)

	jsonOutput, err := ResultsToJSON(results, status, nixInfo)
	if err != nil {
		t.Fatalf("ResultsToJSON() failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if output["status"] != "pass" {
		t.Errorf("Expected status 'pass', got '%v'", output["status"])
	}

	if output["exit_code"] != float64(0) {
		t.Errorf("Expected exit_code 0, got %v", output["exit_code"])
	}
}
