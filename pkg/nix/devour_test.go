package nix

import (
	"testing"

	"github.com/saberzero1/omnix/pkg/nix/store"
	"github.com/stretchr/testify/assert"
)

func TestDevourFlakeURL(t *testing.T) {
	url := DevourFlakeURL()
	assert.NotEmpty(t, url, "DevourFlakeURL should not be empty")
	// Should either be from environment or fallback
	assert.Contains(t, url, "devour-flake")
}

func TestUniquePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    []store.Path
		expected int
	}{
		{
			name:     "no duplicates",
			input:    []store.Path{store.NewPath("/nix/store/abc-foo"), store.NewPath("/nix/store/def-bar")},
			expected: 2,
		},
		{
			name:     "with duplicates",
			input:    []store.Path{store.NewPath("/nix/store/abc-foo"), store.NewPath("/nix/store/abc-foo"), store.NewPath("/nix/store/def-bar")},
			expected: 2,
		},
		{
			name:     "empty list",
			input:    []store.Path{},
			expected: 0,
		},
		{
			name:     "all duplicates",
			input:    []store.Path{store.NewPath("/nix/store/abc-foo"), store.NewPath("/nix/store/abc-foo")},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniquePaths(tt.input)
			assert.Equal(t, tt.expected, len(result))

			// Verify no duplicates in result
			seen := make(map[string]bool)
			for _, path := range result {
				pathStr := path.String()
				assert.False(t, seen[pathStr], "found duplicate path: %s", pathStr)
				seen[pathStr] = true
			}
		})
	}
}
