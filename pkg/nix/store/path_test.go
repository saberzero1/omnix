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
