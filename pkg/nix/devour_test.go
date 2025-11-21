package nix

import (
	"encoding/json"
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

func TestDevourFlakeOutput_UnmarshalJSON(t *testing.T) {
	// This tests the actual JSON format returned by devour-flake
	// as seen in the issue: https://github.com/saberzero1/omnix/issues/XXX
	jsonInput := `{
		"byName": {
			"activate": "/nix/store/99cqaq1rv0w6n2y1xjvx588lcsmjaq4s-activate",
			"darwin-system-25.11.3bda9f6": "/nix/store/01snjwkpdpsa8x1ssmzj2z1f2kb4qyik-darwin-system-25.11.3bda9f6",
			"home-manager-generation": "/nix/store/76mb9q6mw65iwqpsqk8qcfwvp7ni3q8j-home-manager-generation",
			"nixos-unified-template-shell": "/nix/store/26mn8iga5ry5k5dlaxchkpz9y6p57vn2-nixos-unified-template-shell",
			"update-main-flake-inputs": "/nix/store/h9arwv5lvsdqznx386026cf40bar5f93-update-main-flake-inputs"
		},
		"outPaths": [
			"/nix/store/01snjwkpdpsa8x1ssmzj2z1f2kb4qyik-darwin-system-25.11.3bda9f6",
			"/nix/store/ny03lhhc2ll1ag2396gpyypch0m0r4p4-activate-home/bin/activate-home",
			"/nix/store/26mn8iga5ry5k5dlaxchkpz9y6p57vn2-nixos-unified-template-shell",
			"/nix/store/76mb9q6mw65iwqpsqk8qcfwvp7ni3q8j-home-manager-generation",
			"/nix/store/99cqaq1rv0w6n2y1xjvx588lcsmjaq4s-activate",
			"/nix/store/99cqaq1rv0w6n2y1xjvx588lcsmjaq4s-activate",
			"/nix/store/h9arwv5lvsdqznx386026cf40bar5f93-update-main-flake-inputs"
		]
	}`

	var output DevourFlakeOutput
	err := json.Unmarshal([]byte(jsonInput), &output)

	assert.NoError(t, err, "Should successfully unmarshal devour-flake JSON")

	// Verify byName map
	assert.Equal(t, 5, len(output.ByName), "byName should have 5 entries")
	assert.Equal(t, "/nix/store/99cqaq1rv0w6n2y1xjvx588lcsmjaq4s-activate",
		output.ByName["activate"].String())
	assert.Equal(t, "/nix/store/01snjwkpdpsa8x1ssmzj2z1f2kb4qyik-darwin-system-25.11.3bda9f6",
		output.ByName["darwin-system-25.11.3bda9f6"].String())

	// Verify outPaths array (before deduplication)
	assert.Equal(t, 7, len(output.OutPaths), "outPaths should have 7 entries")
	assert.Equal(t, "/nix/store/01snjwkpdpsa8x1ssmzj2z1f2kb4qyik-darwin-system-25.11.3bda9f6",
		output.OutPaths[0].String())

	// Test deduplication
	output.OutPaths = uniquePaths(output.OutPaths)
	assert.Equal(t, 6, len(output.OutPaths), "After deduplication should have 6 unique paths")
}

func TestDevourFlakeOutput_MarshalJSON(t *testing.T) {
	// Test that we can marshal back to JSON
	output := DevourFlakeOutput{
		ByName: map[string]store.Path{
			"test1": store.NewPath("/nix/store/abc-test1"),
			"test2": store.NewPath("/nix/store/def-test2"),
		},
		OutPaths: []store.Path{
			store.NewPath("/nix/store/abc-test1"),
			store.NewPath("/nix/store/def-test2"),
		},
	}

	data, err := json.Marshal(output)
	assert.NoError(t, err)

	// Unmarshal back to verify round-trip
	var output2 DevourFlakeOutput
	err = json.Unmarshal(data, &output2)
	assert.NoError(t, err)

	assert.Equal(t, len(output.ByName), len(output2.ByName))
	assert.Equal(t, len(output.OutPaths), len(output2.OutPaths))
	assert.Equal(t, output.ByName["test1"].String(), output2.ByName["test1"].String())
}
