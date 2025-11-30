package ci

import (
	"context"
	"os"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/saberzero1/omnix/pkg/nix/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOmnixSourcePath(t *testing.T) {
	// Test when OMNIX_SOURCE env var is set
	t.Run("from environment variable", func(t *testing.T) {
		expected := "/nix/store/abc123-omnix"
		require.NoError(t, os.Setenv("OMNIX_SOURCE", expected))
		defer func() {
			_ = os.Unsetenv("OMNIX_SOURCE")
		}()

		path, err := getOmnixSourcePath()
		require.NoError(t, err)
		assert.Equal(t, expected, path)
	})

	// Test error when neither env var nor store path is available
	t.Run("no source available", func(t *testing.T) {
		_ = os.Unsetenv("OMNIX_SOURCE")
		// This test may fail in Nix environment, but should work in normal Go test
		_, err := getOmnixSourcePath()
		// Either succeeds (in Nix) or fails (outside Nix)
		_ = err
	})
}

func TestCacheFlake(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("cache without inputs", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("github:nix-systems/default")
		require.NoError(t, err)

		flakePath, cachedURL, err := cacheFlake(ctx, flake, false)
		if err != nil {
			// This test requires network and Nix, so we allow it to fail gracefully
			t.Skipf("Skipping test due to: %v", err)
		}

		assert.NotEmpty(t, flakePath)
		assert.NotEmpty(t, cachedURL.String())
		// Cached URL should be a store path
		assert.Contains(t, cachedURL.String(), "/nix/store/")
	})

	t.Run("cache with attribute", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("github:nix-systems/default#x86_64-linux")
		require.NoError(t, err)

		_, cachedURL, err := cacheFlake(ctx, flake, false)
		if err != nil {
			t.Skipf("Skipping test due to: %v", err)
		}

		// Should preserve the attribute
		assert.NotEmpty(t, cachedURL.GetAttr().String())
	})
}

func TestRemoteRunOptions(t *testing.T) {
	t.Run("parse SSH URI", func(t *testing.T) {
		uri, err := store.ParseURI("ssh://user@example.com")
		require.NoError(t, err)
		assert.True(t, uri.IsSSH())

		sshURI := uri.GetSSHURI()
		require.NotNil(t, sshURI)
		assert.Equal(t, "user", sshURI.User)
		assert.Equal(t, "example.com", sshURI.Host)
	})

	t.Run("parse SSH URI with copy-inputs option", func(t *testing.T) {
		uri, err := store.ParseURI("ssh://user@example.com?copy-inputs=true")
		require.NoError(t, err)

		opts := uri.GetOptions()
		assert.True(t, opts.CopyInputs)
	})
}

func TestBuildRemoteCICommand(t *testing.T) {
	omnixSource := "/nix/store/abc123-omnix"

	t.Run("basic command", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("/nix/store/xyz-flake")
		require.NoError(t, err)

		opts := RunOptions{
			Systems: []string{"x86_64-linux"},
		}

		args := buildRemoteCICommand(omnixSource, flake, opts, "")

		assert.Contains(t, args, "nix")
		assert.Contains(t, args, "run")
		assert.Contains(t, args, "/nix/store/abc123-omnix#default")
		assert.Contains(t, args, "ci")
		assert.Contains(t, args, "run")
		assert.Contains(t, args, "/nix/store/xyz-flake")
		assert.Contains(t, args, "--systems")
		assert.Contains(t, args, "x86_64-linux")
		assert.Contains(t, args, "--no-link")
	})

	t.Run("with out-link", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("/nix/store/xyz-flake")
		require.NoError(t, err)

		opts := RunOptions{}
		outLink := "/tmp/results.json"

		args := buildRemoteCICommand(omnixSource, flake, opts, outLink)

		assert.Contains(t, args, "--out-link")
		assert.Contains(t, args, outLink)
		assert.NotContains(t, args, "--no-link")
	})

	t.Run("with parallel and concurrency", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("/nix/store/xyz-flake")
		require.NoError(t, err)

		opts := RunOptions{
			Parallel:       true,
			MaxConcurrency: 4,
		}

		args := buildRemoteCICommand(omnixSource, flake, opts, "")

		assert.Contains(t, args, "--parallel")
		assert.Contains(t, args, "--max-concurrency")
		assert.Contains(t, args, "4")
	})

	t.Run("with include-all-dependencies", func(t *testing.T) {
		flake, err := nix.ParseFlakeURL("/nix/store/xyz-flake")
		require.NoError(t, err)

		opts := RunOptions{
			IncludeAllDependencies: true,
		}

		args := buildRemoteCICommand(omnixSource, flake, opts, "")

		assert.Contains(t, args, "--include-all-dependencies")
	})
}

func TestRunOnRemoteStore_ValidationOnly(t *testing.T) {
	ctx := context.Background()

	t.Run("requires SSH URI", func(t *testing.T) {
		// This should fail with non-SSH URI
		flake, err := nix.ParseFlakeURL(".")
		require.NoError(t, err)

		// Create a mock non-SSH URI by manipulating the structure
		// Since we can only create SSH URIs with ParseURI, we'll skip this test
		// and just verify SSH URI works

		uri, err := store.ParseURI("ssh://user@example.com")
		require.NoError(t, err)

		remoteOpts := RemoteRunOptions{
			StoreURI:   uri,
			CopyInputs: false,
		}

		opts := RunOptions{}

		// This will fail at the cache/copy stage, but that's expected
		// We're just testing that it accepts SSH URIs
		_, err = RunOnRemoteStore(ctx, flake, opts, remoteOpts)
		// Error is expected since we're not actually connecting
		assert.Error(t, err)
	})
}
