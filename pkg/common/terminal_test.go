package common

import (
	"testing"
)

func TestIsTerminal(t *testing.T) {
	// Test that IsTerminal doesn't panic
	// The actual result depends on how tests are run, so we just ensure it executes
	_ = IsTerminal()
}
