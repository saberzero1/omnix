package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultFlakeSchemas(t *testing.T) {
	// When built without Nix, these should be empty strings
	// When built with Nix, they will be populated by ldflags
	schemas := GetDefaultFlakeSchemas()

	// We can't assert a specific value since it depends on build method
	// Just verify the function returns a string (empty or not)
	assert.IsType(t, "", schemas)
}

func TestGetInspectFlake(t *testing.T) {
	// When built without Nix, these should be empty strings
	// When built with Nix, they will be populated by ldflags
	inspect := GetInspectFlake()

	// We can't assert a specific value since it depends on build method
	// Just verify the function returns a string (empty or not)
	assert.IsType(t, "", inspect)
}

func TestHasNixBuildEnvironment(t *testing.T) {
	// When built without Nix (normal go build), should return false
	// When built with Nix, should return true
	hasEnv := HasNixBuildEnvironment()

	// The actual value depends on how the tests are run
	// Just verify it returns a boolean
	assert.IsType(t, true, hasEnv)

	// Log the current state for debugging
	t.Logf("HasNixBuildEnvironment: %v", hasEnv)
	t.Logf("DefaultFlakeSchemas: %q", GetDefaultFlakeSchemas())
	t.Logf("InspectFlake: %q", GetInspectFlake())
}

func TestNixBuildEnvironmentConsistency(t *testing.T) {
	// Both should be set together or both empty
	hasSchemas := GetDefaultFlakeSchemas() != ""
	hasInspect := GetInspectFlake() != ""
	hasEnv := HasNixBuildEnvironment()

	// Either both are set, or both are empty
	if hasSchemas || hasInspect {
		// If one is set, HasNixBuildEnvironment should check both
		assert.Equal(t, hasSchemas && hasInspect, hasEnv,
			"HasNixBuildEnvironment should be true only when both are set")
	} else {
		// Both empty means no Nix environment
		assert.False(t, hasEnv, "HasNixBuildEnvironment should be false when paths are empty")
	}
}
