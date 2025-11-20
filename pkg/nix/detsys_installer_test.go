package nix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInstallerVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    InstallerVersion
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "0.16.1",
			want:  InstallerVersion{Major: 0, Minor: 16, Patch: 1},
		},
		{
			name:  "version with prefix",
			input: "nix-installer 0.16.1",
			want:  InstallerVersion{Major: 0, Minor: 16, Patch: 1},
		},
		{
			name:  "version with suffix",
			input: "0.16.1 (some build info)",
			want:  InstallerVersion{Major: 0, Minor: 16, Patch: 1},
		},
		{
			name:  "version at start",
			input: "1.2.3",
			want:  InstallerVersion{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:    "invalid - no version",
			input:   "no version here",
			wantErr: true,
		},
		{
			name:    "invalid - partial version",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInstallerVersion(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInstallerVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version InstallerVersion
		want    string
	}{
		{
			name:    "normal version",
			version: InstallerVersion{Major: 0, Minor: 16, Patch: 1},
			want:    "0.16.1",
		},
		{
			name:    "major version",
			version: InstallerVersion{Major: 1, Minor: 0, Patch: 0},
			want:    "1.0.0",
		},
		{
			name:    "all zeros",
			version: InstallerVersion{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetSysInstaller_String(t *testing.T) {
	installer := DetSysInstaller{
		Version: InstallerVersion{Major: 0, Minor: 16, Patch: 1},
	}
	want := "DetSys nix-installer (0.16.1)"
	got := installer.String()
	assert.Equal(t, want, got)
}

func TestDetectDetSysInstaller(t *testing.T) {
	// This test only runs if we're in a testing environment
	// In most cases, /nix/nix-installer won't exist
	installer, err := DetectDetSysInstaller()
	require.NoError(t, err)
	
	if installer != nil {
		// If the installer is detected, verify it has a valid version
		assert.Greater(t, installer.Version.Major, uint32(0), "Major version should be > 0")
		t.Logf("Detected: %s", installer.String())
	} else {
		t.Log("DetSys nix-installer not installed (expected on most systems)")
	}
}
