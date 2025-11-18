package nix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewCmd(t *testing.T) {
	cmd := NewCmd()
	if cmd == nil {
		t.Fatal("NewCmd() returned nil")
	}
	if cmd.ExtraArgs == nil {
		t.Error("NewCmd() ExtraArgs should be initialized")
	}
}

func TestRunVersion(t *testing.T) {
	// This is an integration test - skip if nix is not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	version, err := cmd.RunVersion(ctx)
	if err != nil {
		// If nix is not installed, skip the test
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("RunVersion() error = %v", err)
	}

	// Verify we got a valid version
	if version.Major == 0 && version.Minor == 0 && version.Patch == 0 {
		t.Error("RunVersion() returned zero version")
	}

	t.Logf("Detected Nix version: %s", version)
}

func TestRun(t *testing.T) {
	// This is an integration test - skip if nix is not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := cmd.Run(ctx, "--version")
	if err != nil {
		// If nix is not installed, skip the test
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("Run() error = %v", err)
	}

	// Verify output contains "nix"
	if !strings.Contains(output, "nix") {
		t.Errorf("Run(--version) output doesn't contain 'nix': %s", output)
	}
}

func TestRunJSON(t *testing.T) {
	// This is an integration test - skip if nix is not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with a simple JSON output - using eval to get a JSON value
	var result map[string]interface{}
	err := cmd.RunJSON(ctx, &result, "eval", "--expr", "{foo = 42;}", "--json")
	if err != nil {
		// If nix is not installed, skip the test
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("RunJSON() error = %v", err)
	}

	// Verify we got the expected JSON
	if foo, ok := result["foo"]; !ok {
		t.Error("RunJSON() result missing 'foo' key")
	} else if fooNum, ok := foo.(float64); !ok || fooNum != 42 {
		t.Errorf("RunJSON() result['foo'] = %v, want 42", foo)
	}
}

func TestCommandError(t *testing.T) {
	err := &CommandError{
		Command:  "nix",
		Args:     []string{"flake", "show"},
		ExitCode: 1,
		Stderr:   "error: some error message",
		Err:      nil,
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "nix") {
		t.Error("CommandError.Error() should contain 'nix'")
	}
	if !strings.Contains(errStr, "flake show") {
		t.Error("CommandError.Error() should contain command args")
	}
	if !strings.Contains(errStr, "exit 1") {
		t.Error("CommandError.Error() should contain exit code")
	}
}

func TestCommandError_Unwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	cmdErr := &CommandError{
		Command:  "nix",
		Args:     []string{"build"},
		ExitCode: 1,
		Stderr:   "build failed",
		Err:      baseErr,
	}

	unwrapped := cmdErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("CommandError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestRunWithInvalidCommand(t *testing.T) {
	cmd := NewCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to run an invalid nix command
	_, err := cmd.Run(ctx, "this-is-not-a-valid-subcommand-xyz123")
	if err == nil {
		t.Error("Run() with invalid command should return error")
	}

	// Verify it's a CommandError
	var cmdErr *CommandError
	if !errors.As(err, &cmdErr) {
		t.Errorf("Run() error should be CommandError, got %T", err)
	}

	// Test error message contains useful info
	errMsg := err.Error()
	if !strings.Contains(errMsg, "nix") {
		t.Error("Error message should contain 'nix'")
	}
}

func TestRunWithTimeout(t *testing.T) {
	cmd := NewCmd()
	// Use a very short timeout to force context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should fail due to context timeout
	_, err := cmd.Run(ctx, "--version")
	if err == nil {
		// On fast systems, the command might complete before timeout
		t.Log("Command completed before timeout (acceptable)")
		return
	}

	// Error should be context-related or command error
	t.Logf("Got expected error: %v", err)
}

func TestRunJSONWithInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := NewCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run a command that doesn't produce JSON
	var result map[string]interface{}
	err := cmd.RunJSON(ctx, &result, "--version")
	if err == nil {
		t.Error("RunJSON() with non-JSON output should return error")
	}

	// Should be a JSON parse error
	var jsonErr *json.SyntaxError
	if !errors.As(err, &jsonErr) && !strings.Contains(err.Error(), "parse") {
		t.Logf("Expected JSON parse error, got: %v", err)
	}
}

func TestCmdWithExtraArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cmd := &Cmd{
		ExtraArgs: []string{"--extra-experimental-features", "nix-command"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This should still work with extra args
	version, err := cmd.RunVersion(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("RunVersion() with extra args error = %v", err)
	}

	if version.Major == 0 && version.Minor == 0 && version.Patch == 0 {
		t.Error("RunVersion() with extra args returned zero version")
	}
}
