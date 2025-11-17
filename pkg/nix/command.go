package nix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	
	"github.com/juspay/omnix/pkg/common"
	"go.uber.org/zap"
)

// CommandError represents an error from running a Nix command.
type CommandError struct {
	Command  string
	Args     []string
	ExitCode int
	Stderr   string
	Err      error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("nix command failed: %s %s (exit %d): %s",
		e.Command, strings.Join(e.Args, " "), e.ExitCode, e.Stderr)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// Cmd represents a Nix command executor.
type Cmd struct {
	// ExtraArgs are additional arguments to pass to all nix commands
	ExtraArgs []string
}

// NewCmd creates a new Nix command executor.
func NewCmd() *Cmd {
	return &Cmd{
		ExtraArgs: []string{},
	}
}

// RunVersion executes `nix --version` and returns the parsed version.
func (c *Cmd) RunVersion(ctx context.Context) (Version, error) {
	output, err := c.runReturningStdout(ctx, []string{"--version"})
	if err != nil {
		return Version{}, err
	}
	
	return ParseVersion(strings.TrimSpace(string(output)))
}

// RunJSON executes a nix command and parses the JSON output into the provided type.
func (c *Cmd) RunJSON(ctx context.Context, result interface{}, args ...string) error {
	output, err := c.runReturningStdout(ctx, args)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal(output, result); err != nil {
		return fmt.Errorf("failed to parse nix command JSON output: %w", err)
	}
	
	return nil
}

// Run executes a nix command and returns the stdout as a string.
func (c *Cmd) Run(ctx context.Context, args ...string) (string, error) {
	output, err := c.runReturningStdout(ctx, args)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// runReturningStdout executes a nix command and returns stdout as bytes.
func (c *Cmd) runReturningStdout(ctx context.Context, args []string) ([]byte, error) {
	// Combine extra args with command args
	allArgs := append(c.ExtraArgs, args...)
	
	// Log the command
	logger := common.Logger()
	logger.Debug("executing nix command",
		zap.String("command", "nix"),
		zap.Strings("args", allArgs))
	
	// Create the command
	cmd := exec.CommandContext(ctx, "nix", allArgs...)
	
	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		
		return nil, &CommandError{
			Command:  "nix",
			Args:     allArgs,
			ExitCode: exitCode,
			Stderr:   stderr.String(),
			Err:      err,
		}
	}
	
	return stdout.Bytes(), nil
}
