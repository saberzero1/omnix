package flake

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Cmd represents a Nix command executor interface for flake operations.
type Cmd interface {
	Run(ctx context.Context, args ...string) (string, error)
}

// FlakeOptions represents options for flake commands.
type FlakeOptions struct {
	// OverrideInputs maps input names to flake URLs to override
	OverrideInputs map[string]string
	// Impure enables impure evaluation
	Impure bool
	// Refresh refreshes cached flake data
	Refresh bool
}

// Eval runs `nix eval <url> --json` and parses the result into the provided type.
func Eval[T any](ctx context.Context, cmd Cmd, opts *FlakeOptions, url string) (T, error) {
	return eval[T](ctx, cmd, opts, url, false)
}

// EvalMaybe is like Eval but returns nil if the attribute is missing.
func EvalMaybe[T any](ctx context.Context, cmd Cmd, opts *FlakeOptions, url string) (*T, error) {
	result, err := eval[T](ctx, cmd, opts, url, true)
	if err != nil {
		if isMissingAttributeError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

// eval is the internal implementation for evaluation.
func eval[T any](ctx context.Context, cmd Cmd, opts *FlakeOptions, url string, captureStderr bool) (T, error) {
	var result T
	
	args := []string{"eval", "--json"}
	
	// Add flake options
	if opts != nil {
		if opts.Impure {
			args = append(args, "--impure")
		}
		if opts.Refresh {
			args = append(args, "--refresh")
		}
		for input, flakeURL := range opts.OverrideInputs {
			args = append(args, "--override-input", input, flakeURL)
		}
	}
	
	args = append(args, url)
	
	// Suppress logs from --override-input (requires double --quiet)
	args = append(args, "--quiet", "--quiet")
	
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return result, err
	}
	
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return result, nil
}

// isMissingAttributeError checks if an error is due to a missing attribute.
func isMissingAttributeError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "does not provide attribute")
}

// EvalExpr evaluates a Nix expression and returns the result.
func EvalExpr[T any](ctx context.Context, cmd Cmd, expr string) (T, error) {
	var result T
	
	args := []string{"eval", "--json", "--expr", expr}
	
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return result, err
	}
	
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return result, nil
}

// EvalImpureExpr evaluates an impure Nix expression and returns the result.
func EvalImpureExpr[T any](ctx context.Context, cmd Cmd, expr string) (T, error) {
	var result T
	
	args := []string{"eval", "--impure", "--json", "--expr", expr}
	
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		return result, err
	}
	
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return result, nil
}
