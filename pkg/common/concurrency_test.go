package common

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateMaxConcurrency(t *testing.T) {
	numCPU := runtime.NumCPU()

	tests := []struct {
		name           string
		maxConcurrency int
		itemCount      int
		expected       int
	}{
		{
			name:           "explicit concurrency is used",
			maxConcurrency: 10,
			itemCount:      100,
			expected:       10,
		},
		{
			name:           "zero concurrency uses CPU-based default",
			maxConcurrency: 0,
			itemCount:      100,
			expected:       max(DefaultMaxConcurrency, numCPU),
		},
		{
			name:           "negative concurrency uses CPU-based default",
			maxConcurrency: -1,
			itemCount:      100,
			expected:       max(DefaultMaxConcurrency, numCPU),
		},
		{
			name:           "concurrency capped by item count",
			maxConcurrency: 0,
			itemCount:      2,
			expected:       2,
		},
		{
			name:           "explicit concurrency not capped by item count",
			maxConcurrency: 10,
			itemCount:      5,
			expected:       10, // Explicit values are not capped
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateMaxConcurrency(tc.maxConcurrency, tc.itemCount)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateMaxConcurrency_MinimumDefault(t *testing.T) {
	// Ensure that even on single-core systems, we get at least DefaultMaxConcurrency
	result := CalculateMaxConcurrency(0, 100)
	assert.GreaterOrEqual(t, result, DefaultMaxConcurrency)
}
