package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	// Test that version variables are set
	assert.NotEmpty(t, Version, "Version should be set")
	assert.NotEmpty(t, Commit, "Commit should be set")
}

func TestMain_Variables(t *testing.T) {
	// Ensure default values
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "dev", Commit)
}
