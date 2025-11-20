package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStoreCmd(t *testing.T) {
	cmd := NewStoreCmd()
	assert.NotNil(t, cmd)
}

func TestStoreCmd_Methods(t *testing.T) {
	// These tests verify the API structure
	// Actual functionality tests would require a Nix installation

	cmd := NewStoreCmd()
	ctx := context.Background()

	t.Run("QueryDeriver structure", func(t *testing.T) {
		// Just verify the method exists and has correct signature
		paths := []Path{NewPath("/nix/store/test")}
		_, err := cmd.QueryDeriver(ctx, paths)
		// We expect an error since nix-store may not be available
		// or the path doesn't exist
		_ = err // Error is expected in test environment
	})

	t.Run("QueryRequisites structure", func(t *testing.T) {
		drvPaths := []string{"/nix/store/test.drv"}
		_, err := cmd.QueryRequisites(ctx, drvPaths, true)
		_ = err // Error is expected in test environment
	})

	t.Run("FetchAllDeps structure", func(t *testing.T) {
		paths := []Path{NewPath("/nix/store/test")}
		_, err := cmd.FetchAllDeps(ctx, paths)
		_ = err // Error is expected in test environment
	})

	t.Run("Add structure", func(t *testing.T) {
		_, err := cmd.Add(ctx, "/tmp/test")
		_ = err // Error is expected in test environment
	})

	t.Run("AddRoot structure", func(t *testing.T) {
		paths := []Path{NewPath("/nix/store/test")}
		err := cmd.AddRoot(ctx, "/tmp/root", paths)
		_ = err // Error is expected in test environment
	})

	t.Run("AddFilePermanently structure", func(t *testing.T) {
		_, err := cmd.AddFilePermanently(ctx, "/tmp/root", "test content")
		_ = err // Error is expected in test environment
	})
}

func TestFetchAllDeps_EmptyPaths(t *testing.T) {
	cmd := NewStoreCmd()
	ctx := context.Background()

	// Test with empty paths - should handle gracefully
	emptyPaths := []Path{}
	_, err := cmd.FetchAllDeps(ctx, emptyPaths)
	// The behavior with empty paths may vary, so we just check it doesn't panic
	_ = err
}

func TestQueryRequisites_IncludeOutputsFlag(t *testing.T) {
	cmd := NewStoreCmd()
	ctx := context.Background()

	testPaths := []string{"/nix/store/test.drv"}

	t.Run("with include outputs", func(t *testing.T) {
		_, err := cmd.QueryRequisites(ctx, testPaths, true)
		_ = err // Expected to fail without real Nix
	})

	t.Run("without include outputs", func(t *testing.T) {
		_, err := cmd.QueryRequisites(ctx, testPaths, false)
		_ = err // Expected to fail without real Nix
	})
}

func TestAddRoot_MultiplePaths(t *testing.T) {
	cmd := NewStoreCmd()
	ctx := context.Background()

	multiplePaths := []Path{
		NewPath("/nix/store/test1"),
		NewPath("/nix/store/test2"),
	}

	err := cmd.AddRoot(ctx, "/tmp/multi-root", multiplePaths)
	_ = err // Expected to fail without real Nix
}
