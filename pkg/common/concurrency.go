package common

import "runtime"

const (
	// DefaultMaxConcurrency is the default maximum number of concurrent workers.
	// This value balances parallelism with avoiding overwhelming external services
	// (like GitHub API) when building flakes with many outputs or GitHub-based inputs.
	DefaultMaxConcurrency = 4
)

// CalculateMaxConcurrency determines the optimal concurrency level for parallel operations.
//
// When maxConcurrency is 0 or negative, it returns a sensible default based on CPU count
// (at least DefaultMaxConcurrency) but never more than itemCount to avoid idle workers.
//
// This helps prevent overwhelming external APIs (like GitHub) with too many concurrent
// requests while still providing good parallelism for typical workloads.
//
// Parameters:
//   - maxConcurrency: The user-specified max concurrency (0 or negative means auto)
//   - itemCount: The total number of items to process
//
// Returns:
//   - The calculated concurrency level to use
func CalculateMaxConcurrency(maxConcurrency, itemCount int) int {
	if maxConcurrency > 0 {
		return maxConcurrency
	}

	// Use the greater of DefaultMaxConcurrency or NumCPU
	cpuBased := runtime.NumCPU()
	if cpuBased < DefaultMaxConcurrency {
		cpuBased = DefaultMaxConcurrency
	}

	// Never more than the number of items to avoid idle workers
	if cpuBased > itemCount {
		cpuBased = itemCount
	}

	return cpuBased
}
