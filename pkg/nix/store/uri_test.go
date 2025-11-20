package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantScheme  string
		wantUser    string
		wantHost    string
		wantOptions Options
		wantErr     bool
	}{
		{
			name:        "ssh with user",
			input:       "ssh://user@example.com",
			wantScheme:  "ssh",
			wantUser:    "user",
			wantHost:    "example.com",
			wantOptions: Options{CopyInputs: false},
		},
		{
			name:        "ssh without user",
			input:       "ssh://example.com",
			wantScheme:  "ssh",
			wantUser:    "",
			wantHost:    "example.com",
			wantOptions: Options{CopyInputs: false},
		},
		{
			name:        "ssh with copy-inputs option",
			input:       "ssh://user@example.com?copy-inputs=true",
			wantScheme:  "ssh",
			wantUser:    "user",
			wantHost:    "example.com",
			wantOptions: Options{CopyInputs: true},
		},
		{
			name:    "unsupported scheme",
			input:   "http://example.com",
			wantErr: true,
		},
		{
			name:    "missing host",
			input:   "ssh://",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			input:   "not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURI(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantScheme, got.scheme)
			if got.IsSSH() {
				assert.NotNil(t, got.sshURI)
				assert.Equal(t, tt.wantUser, got.sshURI.User)
				assert.Equal(t, tt.wantHost, got.sshURI.Host)
			}
			assert.Equal(t, tt.wantOptions, got.GetOptions())
		})
	}
}

func TestURI_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ssh with user",
			input: "ssh://user@example.com",
			want:  "ssh://user@example.com",
		},
		{
			name:  "ssh without user",
			input: "ssh://example.com",
			want:  "ssh://example.com",
		},
		{
			name:  "ssh with options (options not in string output)",
			input: "ssh://user@example.com?copy-inputs=true",
			want:  "ssh://user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, uri.String())
		})
	}
}

func TestURI_IsSSH(t *testing.T) {
	uri, err := ParseURI("ssh://example.com")
	require.NoError(t, err)
	assert.True(t, uri.IsSSH())
}

func TestURI_GetSSHURI(t *testing.T) {
	uri, err := ParseURI("ssh://user@example.com")
	require.NoError(t, err)

	sshURI := uri.GetSSHURI()
	require.NotNil(t, sshURI)
	assert.Equal(t, "user", sshURI.User)
	assert.Equal(t, "example.com", sshURI.Host)
}

func TestSSHURI_String(t *testing.T) {
	tests := []struct {
		name string
		uri  SSHURI
		want string
	}{
		{
			name: "with user",
			uri:  SSHURI{User: "user", Host: "example.com"},
			want: "user@example.com",
		},
		{
			name: "without user",
			uri:  SSHURI{Host: "example.com"},
			want: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.uri.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestURI_GetOptions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Options
	}{
		{
			name:  "default options",
			input: "ssh://example.com",
			want:  Options{CopyInputs: false},
		},
		{
			name:  "copy-inputs enabled",
			input: "ssh://example.com?copy-inputs=true",
			want:  Options{CopyInputs: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, uri.GetOptions())
		})
	}
}
