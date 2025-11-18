package ci

import (
	"encoding/json"
	"fmt"
)

// GitHubMatrixRow represents a single row in the GitHub Actions matrix
type GitHubMatrixRow struct {
	// System to build on (e.g., "x86_64-linux", "aarch64-darwin")
	System string `json:"system"`

	// Subflake to build (e.g., ".", "tests")
	Subflake string `json:"subflake"`
}

// GitHubMatrix represents a GitHub Actions matrix configuration
type GitHubMatrix struct {
	// Include contains all matrix rows
	Include []GitHubMatrixRow `json:"include"`
}

// GenerateMatrix creates a GitHub Actions matrix from systems and subflakes
func GenerateMatrix(systems []string, config Config) GitHubMatrix {
	var include []GitHubMatrixRow

	for _, system := range systems {
		for name, subflake := range config.Default {
			// Skip if this subflake is marked to skip
			if subflake.Skip {
				continue
			}

			// Skip if this subflake can't run on this system
			if !subflake.CanRunOn([]string{system}) {
				continue
			}

			include = append(include, GitHubMatrixRow{
				System:   system,
				Subflake: name,
			})
		}
	}

	return GitHubMatrix{Include: include}
}

// ToJSON converts the matrix to JSON format
func (m *GitHubMatrix) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal matrix to JSON: %w", err)
	}
	return string(data), nil
}

// Count returns the number of matrix rows
func (m *GitHubMatrix) Count() int {
	return len(m.Include)
}
