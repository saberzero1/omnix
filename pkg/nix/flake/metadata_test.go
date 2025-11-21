package flake

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadata_JSONParsing(t *testing.T) {
	testJSON := `{
		"description": "Test flake",
		"lastModified": 1234567890,
		"url": "github:owner/repo",
		"path": "/nix/store/test",
		"revision": "abc123",
		"revCount": 42
	}`

	var metadata Metadata
	err := json.Unmarshal([]byte(testJSON), &metadata)
	require.NoError(t, err)

	assert.Equal(t, "Test flake", metadata.Description)
	assert.Equal(t, int64(1234567890), metadata.LastModified)
	assert.Equal(t, "github:owner/repo", metadata.URL)
	assert.Equal(t, "/nix/store/test", metadata.Path)
	assert.Equal(t, "abc123", metadata.Revision)
	assert.Equal(t, 42, metadata.RevCount)
}

func TestLockedMetadata_JSONParsing(t *testing.T) {
	testJSON := `{
		"lastModified": 1234567890,
		"narHash": "sha256-abc123",
		"owner": "testowner",
		"repo": "testrepo",
		"rev": "abc123def",
		"type": "github"
	}`

	var locked LockedMetadata
	err := json.Unmarshal([]byte(testJSON), &locked)
	require.NoError(t, err)

	assert.Equal(t, int64(1234567890), locked.LastModified)
	assert.Equal(t, "sha256-abc123", locked.NarHash)
	assert.Equal(t, "testowner", locked.Owner)
	assert.Equal(t, "testrepo", locked.Repo)
	assert.Equal(t, "abc123def", locked.Rev)
	assert.Equal(t, "github", locked.Type)
}

func TestFlakeRef_JSONParsing(t *testing.T) {
	testJSON := `{
		"owner": "NixOS",
		"repo": "nixpkgs",
		"type": "github",
		"ref": "nixos-unstable"
	}`

	var ref FlakeRef
	err := json.Unmarshal([]byte(testJSON), &ref)
	require.NoError(t, err)

	assert.Equal(t, "NixOS", ref.Owner)
	assert.Equal(t, "nixpkgs", ref.Repo)
	assert.Equal(t, "github", ref.Type)
	assert.Equal(t, "nixos-unstable", ref.Ref)
}

func TestMetadata_CompleteStructure(t *testing.T) {
	testJSON := `{
		"description": "Complete test",
		"lastModified": 1234567890,
		"locked": {
			"narHash": "sha256-test",
			"owner": "test",
			"repo": "test",
			"rev": "abc",
			"type": "github"
		},
		"original": {
			"owner": "test",
			"repo": "test",
			"type": "github"
		},
		"resolved": {
			"owner": "test",
			"repo": "test",
			"type": "github",
			"ref": "main"
		},
		"url": "github:test/test",
		"path": "/nix/store/test"
	}`

	var metadata Metadata
	err := json.Unmarshal([]byte(testJSON), &metadata)
	require.NoError(t, err)

	assert.Equal(t, "Complete test", metadata.Description)
	assert.NotNil(t, metadata.Locked)
	assert.NotNil(t, metadata.Original)
	assert.NotNil(t, metadata.Resolved)
	assert.Equal(t, "github:test/test", metadata.URL)
	assert.Equal(t, "/nix/store/test", metadata.Path)
}

func TestGetMetadata_APIStructure(t *testing.T) {
	// Test that the API is correctly structured
	ctx := context.Background()

	t.Run("GetMetadata exists", func(t *testing.T) {
		cmd := &mockCmd{output: `{"description":"test"}`}
		metadata, err := GetMetadata(ctx, cmd, ".")
		require.NoError(t, err)
		assert.NotNil(t, metadata)
	})

	t.Run("GetMetadataWithOptions exists", func(t *testing.T) {
		cmd := &mockCmd{output: `{"description":"test"}`}
		opts := &FlakeOptions{
			Refresh: true,
		}
		metadata, err := GetMetadataWithOptions(ctx, cmd, ".", opts)
		require.NoError(t, err)
		assert.NotNil(t, metadata)
	})
}

func TestLock_APIStructure(t *testing.T) {
	ctx := context.Background()

	t.Run("Lock with no updates", func(t *testing.T) {
		cmd := &mockCmd{output: ""}
		err := Lock(ctx, cmd, ".", []string{})
		require.NoError(t, err)
	})

	t.Run("Lock with updates", func(t *testing.T) {
		cmd := &mockCmd{output: ""}
		err := Lock(ctx, cmd, ".", []string{"nixpkgs"})
		require.NoError(t, err)
	})
}

func TestUpdate_APIStructure(t *testing.T) {
	ctx := context.Background()

	cmd := &mockCmd{output: ""}
	err := Update(ctx, cmd, ".")
	require.NoError(t, err)
}

func TestMetadataInput(t *testing.T) {
	input := MetadataInput{
		FlakeURL:      "github:NixOS/nixpkgs",
		IncludeInputs: true,
	}

	assert.Equal(t, "github:NixOS/nixpkgs", input.FlakeURL)
	assert.True(t, input.IncludeInputs)
}

func TestFlakeOptions_WithMetadata(t *testing.T) {
	opts := &FlakeOptions{
		Refresh: true,
		OverrideInputs: map[string]string{
			"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
		},
	}

	assert.True(t, opts.Refresh)
	assert.Len(t, opts.OverrideInputs, 1)
	assert.Equal(t, "github:NixOS/nixpkgs/nixos-unstable", opts.OverrideInputs["nixpkgs"])
}
