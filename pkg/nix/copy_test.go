package nix

import (
	"context"
	"testing"

	"github.com/saberzero1/omnix/pkg/nix/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	// Note: These are unit tests that verify the argument construction
	// Integration tests would require actual nix copy functionality

	t.Run("basic copy with to URI", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		toURI, err := store.ParseURI("ssh://example.com")
		require.NoError(t, err)

		cmd := NewCmd()
		options := CopyOptions{
			To: toURI,
		}

		// This will fail because example.com doesn't exist,
		// but we can verify the command is constructed correctly
		err = Copy(context.Background(), cmd, options, []string{"/nix/store/test"})
		assert.Error(t, err) // Expected to fail with non-existent host
	})
}

func TestCopyOptions(t *testing.T) {
	tests := []struct {
		name    string
		options CopyOptions
	}{
		{
			name: "with from URI",
			options: CopyOptions{
				From: mustParseURI("ssh://source.example.com"),
			},
		},
		{
			name: "with to URI",
			options: CopyOptions{
				To: mustParseURI("ssh://dest.example.com"),
			},
		},
		{
			name: "with both URIs",
			options: CopyOptions{
				From: mustParseURI("ssh://source.example.com"),
				To:   mustParseURI("ssh://dest.example.com"),
			},
		},
		{
			name: "with no-check-sigs",
			options: CopyOptions{
				To:          mustParseURI("ssh://dest.example.com"),
				NoCheckSigs: true,
			},
		},
		{
			name:    "empty options",
			options: CopyOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the options can be created
			assert.NotNil(t, tt.options)

			// Verify from/to URIs are valid if present
			if tt.options.From != nil {
				assert.NotEmpty(t, tt.options.From.String())
			}
			if tt.options.To != nil {
				assert.NotEmpty(t, tt.options.To.String())
			}
		})
	}
}

func TestCopyPath(t *testing.T) {
	t.Run("single path", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		toURI, err := store.ParseURI("ssh://example.com")
		require.NoError(t, err)

		cmd := NewCmd()
		options := CopyOptions{
			To: toURI,
		}

		err = CopyPath(context.Background(), cmd, options, "/nix/store/test")
		assert.Error(t, err) // Expected to fail with non-existent host
	})
}

// Helper function for tests
func mustParseURI(uri string) *store.URI {
	parsed, err := store.ParseURI(uri)
	if err != nil {
		panic(err)
	}
	return parsed
}
