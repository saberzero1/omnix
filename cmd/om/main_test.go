package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_DefaultValues(t *testing.T) {
	// Test that version variables have default values
	assert.Equal(t, "dev", Version, "Version should default to 'dev'")
	assert.Equal(t, "dev", Commit, "Commit should default to 'dev'")
}
