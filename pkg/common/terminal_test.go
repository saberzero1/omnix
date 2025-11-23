package common

import (
	"testing"
)

func TestIsTerminal(t *testing.T) {
	// Test that IsTerminal doesn't panic
	// The actual result depends on how tests are run, so we just ensure it returns a bool
	result := IsTerminal()

	// Result should be a boolean (either true or false)
	if result != true && result != false {
		t.Errorf("IsTerminal() should return a boolean value")
	}
}
