package ci

import (
	"context"
	"testing"
	"time"

	"github.com/juspay/omnix/pkg/nix"
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
					Build: BuildStep{Enable: false},
					Custom: []CustomStep{
						{
							Name:    "test1",
							Command: []string{"echo", "hello1"},
							Enable:  true,
						},
					},
				},
			},
			"sub2": {
				Skip: false,
				Dir:  ".",
				Steps: StepsConfig{
					Build: BuildStep{Enable: false},
					Custom: []CustomStep{
						{
							Name:    "test2",
							Command: []string{"echo", "hello2"},
							Enable:  true,
						},
					},
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

	// Both results should have succeeded
	for _, result := range results {
		assert.True(t, result.Success)
	}
}

func TestRunParallelWithConcurrencyLimit(t *testing.T) {
	ctx := context.Background()

	config := Config{
		Default: map[string]SubflakeConfig{
			"sub1": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "1"}, Enable: true}}}},
			"sub2": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "2"}, Enable: true}}}},
			"sub3": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "3"}, Enable: true}}}},
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
			"sub1": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "1"}, Enable: true}}}},
			"sub2": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "2"}, Enable: true}}}},
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
			"first":  {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "first"}, Enable: true}}}},
			"second": {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "second"}, Enable: true}}}},
			"third":  {Dir: ".", Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "third"}, Enable: true}}}},
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
			"success": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"echo", "ok"}, Enable: true}}}},
			"failure": {Steps: StepsConfig{Build: BuildStep{Enable: false}, Custom: []CustomStep{{Name: "test", Command: []string{"false"}, Enable: true}}}},
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

	// One should have failed
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount, "Expected exactly one successful result")
}
