package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantIsDrv bool
	}{
		{
			name:      "derivation path",
			input:     "/nix/store/abc123-hello-2.10.drv",
			wantIsDrv: true,
		},
		{
			name:      "output path",
			input:     "/nix/store/xyz789-hello-2.10",
			wantIsDrv: false,
		},
		{
			name:      "bin path",
			input:     "/nix/store/xyz789-hello-2.10/bin/hello",
			wantIsDrv: false,
		},
		{
			name:      "drv in middle of path",
			input:     "/nix/store/some.drv/output",
			wantIsDrv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPath(tt.input)
			assert.Equal(t, tt.input, path.String())
			assert.Equal(t, tt.wantIsDrv, path.IsDrv())
			assert.Equal(t, !tt.wantIsDrv, path.IsOutput())
		})
	}
}

func TestPath_AsPath(t *testing.T) {
	path := NewPath("/nix/store/abc-foo")
	assert.Equal(t, "/nix/store/abc-foo", path.AsPath())
}

func TestPath_Base(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple path",
			input: "/nix/store/abc-hello",
			want:  "abc-hello",
		},
		{
			name:  "drv path",
			input: "/nix/store/xyz-hello.drv",
			want:  "xyz-hello.drv",
		},
		{
			name:  "nested path",
			input: "/nix/store/abc-hello/bin/hello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPath(tt.input)
			assert.Equal(t, tt.want, path.Base())
		})
	}
}

func TestPath_Dir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "store path",
			input: "/nix/store/abc-hello",
			want:  "/nix/store",
		},
		{
			name:  "nested path",
			input: "/nix/store/abc-hello/bin/hello",
			want:  "/nix/store/abc-hello/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewPath(tt.input)
			assert.Equal(t, tt.want, path.Dir())
		})
	}
}

func TestPath_String(t *testing.T) {
	input := "/nix/store/abc-hello-2.10.drv"
	path := NewPath(input)
	assert.Equal(t, input, path.String())
}

func TestPath_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		path     Path
		expected string
	}{
		{
			name:     "output path",
			path:     NewPath("/nix/store/abc123-hello-2.10"),
			expected: `"/nix/store/abc123-hello-2.10"`,
		},
		{
			name:     "derivation path",
			path:     NewPath("/nix/store/xyz789-hello.drv"),
			expected: `"/nix/store/xyz789-hello.drv"`,
		},
		{
			name:     "nested path",
			path:     NewPath("/nix/store/abc-hello/bin/hello"),
			expected: `"/nix/store/abc-hello/bin/hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.path.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestPath_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		expected  string
		wantIsDrv bool
		wantError bool
	}{
		{
			name:      "output path",
			json:      `"/nix/store/abc123-hello-2.10"`,
			expected:  "/nix/store/abc123-hello-2.10",
			wantIsDrv: false,
			wantError: false,
		},
		{
			name:      "derivation path",
			json:      `"/nix/store/xyz789-hello.drv"`,
			expected:  "/nix/store/xyz789-hello.drv",
			wantIsDrv: true,
			wantError: false,
		},
		{
			name:      "nested path",
			json:      `"/nix/store/abc-hello/bin/hello"`,
			expected:  "/nix/store/abc-hello/bin/hello",
			wantIsDrv: false,
			wantError: false,
		},
		{
			name:      "invalid json - not a string",
			json:      `123`,
			wantError: true,
		},
		{
			name:      "invalid json - malformed",
			json:      `"incomplete`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path Path
			err := path.UnmarshalJSON([]byte(tt.json))

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, path.String())
			assert.Equal(t, tt.wantIsDrv, path.IsDrv())
		})
	}
}

func TestPath_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "output path",
			path: "/nix/store/abc123-hello-2.10",
		},
		{
			name: "derivation path",
			path: "/nix/store/xyz789-hello.drv",
		},
		{
			name: "nested path",
			path: "/nix/store/abc-hello/bin/hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewPath(tt.path)

			// Marshal to JSON
			data, err := original.MarshalJSON()
			assert.NoError(t, err)

			// Unmarshal back
			var unmarshaled Path
			err = unmarshaled.UnmarshalJSON(data)
			assert.NoError(t, err)

			// Should be identical
			assert.Equal(t, original.String(), unmarshaled.String())
			assert.Equal(t, original.IsDrv(), unmarshaled.IsDrv())
		})
	}
}
