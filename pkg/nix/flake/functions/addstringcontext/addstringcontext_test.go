package addstringcontext

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddStringContextFn_FlakeURL(t *testing.T) {
	fn := AddStringContextFn{}
	url := fn.FlakeURL()
	assert.NotEmpty(t, url)
	// Should contain "addstringcontext" in some form
	assert.Contains(t, url, "addstringcontext")
}

func TestAddStringContextFn_Init(t *testing.T) {
	fn := AddStringContextFn{}
	var out json.RawMessage
	// Init should not panic
	fn.Init(&out)
}

// Note: Full integration tests for AddStringContext would require:
// 1. A working Nix environment with the flake function built
// 2. A sample JSON file with store paths
// 3. The ability to verify string context was added
//
// These are better suited for integration tests that run in a Nix environment.
// The unit tests here verify the basic structure and configuration.
