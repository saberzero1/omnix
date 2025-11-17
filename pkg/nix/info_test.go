package nix

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGetInfo(t *testing.T) {
	// This is an integration test requiring Nix
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	info, err := GetInfo(ctx)
	if err != nil {
		// If nix is not installed, skip
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("GetInfo() error = %v", err)
	}
	
	// Verify info is populated
	if info == nil {
		t.Fatal("GetInfo() returned nil")
	}
	
	// Check version
	if info.Version.Major == 0 && info.Version.Minor == 0 && info.Version.Patch == 0 {
		t.Error("GetInfo() returned zero version")
	}
	
	// Check env
	if info.Env == nil {
		t.Fatal("GetInfo() Env is nil")
	}
	
	if info.Env.OS.Type == "" {
		t.Error("GetInfo() Env.OS.Type is empty")
	}
	
	t.Logf("Nix Info: %s", info.String())
	t.Logf("Version: %s", info.Version)
	t.Logf("OS: %s", info.Env.OS)
	t.Logf("User: %s", info.Env.CurrentUser)
	t.Logf("Groups: %v", info.Env.CurrentUserGroups)
}

func TestInfoString(t *testing.T) {
	tests := []struct {
		name    string
		info    *Info
		wantVer string
		wantOS  string
	}{
		{
			name: "Linux system",
			info: &Info{
				Version: Version{Major: 2, Minor: 13, Patch: 0},
				Env: &Env{
					OS: OSType{
						Type: "linux",
					},
				},
			},
			wantVer: "2.13.0",
			wantOS:  "Linux",
		},
		{
			name: "NixOS system",
			info: &Info{
				Version: Version{Major: 2, Minor: 18, Patch: 1},
				Env: &Env{
					OS: OSType{
						Type:    "linux",
						IsNixOS: true,
					},
				},
			},
			wantVer: "2.18.1",
			wantOS:  "NixOS",
		},
		{
			name: "macOS system",
			info: &Info{
				Version: Version{Major: 2, Minor: 20, Patch: 0},
				Env: &Env{
					OS: OSType{
						Type: "darwin",
					},
				},
			},
			wantVer: "2.20.0",
			wantOS:  "macOS",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.String()
			if !strings.Contains(result, tt.wantVer) {
				t.Errorf("Info.String() = %v, should contain version %s", result, tt.wantVer)
			}
			if !strings.Contains(result, tt.wantOS) {
				t.Errorf("Info.String() = %v, should contain OS %s", result, tt.wantOS)
			}
		})
	}
}

func TestGetInfoWithCanceledContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err := GetInfo(ctx)
	if err == nil {
		t.Error("GetInfo() with canceled context should return error")
	}
	
	t.Logf("GetInfo() with canceled context error = %v (expected)", err)
}
