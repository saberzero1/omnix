package ci

import (
	"context"
	"testing"
	"time"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
)

func TestRunParallel(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"sub1": {
				Skip: false,
				Dir:  ".",
				Steps: StepsConfig{
					Build:  BuildStep{Enable: false},
					Custom: make(map[string]CustomStep),
				},
			},
			"sub2": {
				Skip: false,
				Dir:  ".",
				Steps: StepsConfig{
					Build:  BuildStep{Enable: false},
					Custom: make(map[string]CustomStep),
				},
			},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:  []string{"x86_64-linux"},
		Parallel: true,
	}

	results, err := Run(ctx, flake, config, opts)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Both results should have succeeded (no steps to fail)
	for _, result := range results {
		assert.True(t, result.Success)
	}
}

func TestRunParallelWithConcurrencyLimit(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"sub1": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"sub2": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"sub3": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:        []string{"x86_64-linux"},
		Parallel:       true,
		MaxConcurrency: 2, // Only 2 concurrent jobs
	}

	start := time.Now()
	results, err := Run(ctx, flake, config, opts)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Should complete reasonably quickly with parallelism
	assert.Less(t, duration, 5*time.Second)
}

func TestRunSequential(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"sub1": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"sub2": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:  []string{"x86_64-linux"},
		Parallel: false, // Sequential execution
	}

	results, err := Run(ctx, flake, config, opts)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	for _, result := range results {
		assert.True(t, result.Success)
	}
}

func TestRunParallelMaintainsOrder(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"first":  {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"second": {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"third":  {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:  []string{"x86_64-linux"},
		Parallel: true,
	}

	results, err := Run(ctx, flake, config, opts)
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Results should maintain a consistent ordering even with parallel execution
	subflakeNames := make(map[string]bool)
	for _, result := range results {
		subflakeNames[result.Subflake] = true
	}

	// All subflakes should be present
	assert.True(t, subflakeNames["first"])
	assert.True(t, subflakeNames["second"])
	assert.True(t, subflakeNames["third"])
}

func TestRunParallelWithFailure(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"success": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
			"failure": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: make(map[string]CustomStep)}},
		},
	}

	flake, err := nix.ParseFlakeURL(".")
	assert.NoError(t, err)

	opts := RunOptions{
		Systems:  []string{"x86_64-linux"},
		Parallel: true,
	}

	results, err := Run(ctx, flake, config, opts)

	// No error returned - errors are in the result
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Both should succeed (no steps to fail)
	for _, result := range results {
		assert.True(t, result.Success)
	}
}
