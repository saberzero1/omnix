package nix

import (
	"testing"

	"github.com/saberzero1/omnix/pkg/nix/flake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSystemsListFlakeRef(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantURL string
	}{
		{
			name:    "known system x86_64-linux",
			input:   "x86_64-linux",
			wantURL: "github:nix-systems/x86_64-linux",
		},
		{
			name:    "known system aarch64-darwin",
			input:   "aarch64-darwin",
			wantURL: "github:nix-systems/aarch64-darwin",
		},
		{
			name:    "custom flake URL",
			input:   "github:myorg/my-systems",
			wantURL: "github:myorg/my-systems",
		},
		{
			name:    "default",
			input:   "default",
			wantURL: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := ParseSystemsListFlakeRef(tt.input)
			// Just verify it creates a valid ref
			assert.NotNil(t, ref.URL)
		})
	}
}

func TestFromKnownSystem(t *testing.T) {
	tests := []struct {
		name       string
		system     flake.System
		wantNil    bool
		wantURLStr string
	}{
		{
			name:       "x86_64-linux",
			system:     flake.SystemLinuxX86_64,
			wantNil:    false,
			wantURLStr: "github:nix-systems/x86_64-linux",
		},
		{
			name:       "aarch64-linux",
			system:     flake.SystemLinuxAarch64,
			wantNil:    false,
			wantURLStr: "github:nix-systems/aarch64-linux",
		},
		{
			name:       "x86_64-darwin",
			system:     flake.SystemDarwinX86_64,
			wantNil:    false,
			wantURLStr: "github:nix-systems/x86_64-darwin",
		},
		{
			name:       "aarch64-darwin",
			system:     flake.SystemDarwinAarch64,
			wantNil:    false,
			wantURLStr: "github:nix-systems/aarch64-darwin",
		},
		{
			name:    "unknown system",
			system:  flake.ParseSystem("riscv64-linux"),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := FromKnownSystem(tt.system)
			if tt.wantNil {
				assert.Nil(t, ref)
			} else {
				require.NotNil(t, ref)
				assert.Equal(t, tt.wantURLStr, ref.URL.String())
			}
		})
	}
}

func TestSystemsListFromKnownFlake(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantNil     bool
		wantSystems []flake.System
	}{
		{
			name:        "aarch64-linux",
			url:         "github:nix-systems/aarch64-linux",
			wantSystems: []flake.System{flake.SystemLinuxAarch64},
		},
		{
			name:        "x86_64-linux",
			url:         "github:nix-systems/x86_64-linux",
			wantSystems: []flake.System{flake.SystemLinuxX86_64},
		},
		{
			name:        "x86_64-darwin",
			url:         "github:nix-systems/x86_64-darwin",
			wantSystems: []flake.System{flake.SystemDarwinX86_64},
		},
		{
			name:        "aarch64-darwin",
			url:         "github:nix-systems/aarch64-darwin",
			wantSystems: []flake.System{flake.SystemDarwinAarch64},
		},
		{
			name: "default",
			url:  "github:nix-systems/default",
			wantSystems: []flake.System{
				flake.SystemLinuxX86_64,
				flake.SystemLinuxAarch64,
				flake.SystemDarwinX86_64,
				flake.SystemDarwinAarch64,
			},
		},
		{
			name: "default-linux",
			url:  "github:nix-systems/default-linux",
			wantSystems: []flake.System{
				flake.SystemLinuxX86_64,
				flake.SystemLinuxAarch64,
			},
		},
		{
			name: "default-darwin",
			url:  "github:nix-systems/default-darwin",
			wantSystems: []flake.System{
				flake.SystemDarwinX86_64,
				flake.SystemDarwinAarch64,
			},
		},
		{
			name:    "unknown flake",
			url:     "github:custom/systems",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := SystemsListFlakeRef{URL: NewFlakeURL(tt.url)}
			systems := systemsListFromKnownFlake(ref)

			if tt.wantNil {
				assert.Nil(t, systems)
			} else {
				require.NotNil(t, systems)
				assert.Equal(t, len(tt.wantSystems), len(systems.Systems))
				for i, sys := range tt.wantSystems {
					assert.Equal(t, sys.String(), systems.Systems[i].String())
				}
			}
		})
	}
}

func TestSystemsList(t *testing.T) {
	// Test creating a SystemsList
	systems := &SystemsList{
		Systems: []flake.System{
			flake.SystemLinuxX86_64,
			flake.SystemDarwinAarch64,
		},
	}

	assert.Len(t, systems.Systems, 2)
	assert.Equal(t, "x86_64-linux", systems.Systems[0].String())
	assert.Equal(t, "aarch64-darwin", systems.Systems[1].String())
}
