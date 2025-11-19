package nix

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestGetCurrentUser(t *testing.T) {
	// Save original env vars
	origUser := os.Getenv("USER")
	origUsername := os.Getenv("USERNAME")
	defer func() {
		_ = os.Setenv("USER", origUser)
		_ = os.Setenv("USERNAME", origUsername)
	}()

	tests := []struct {
		name         string
		userEnv      string
		usernameEnv  string
		wantNonEmpty bool
	}{
		{
			name:         "USER env var set",
			userEnv:      "testuser",
			usernameEnv:  "",
			wantNonEmpty: true,
		},
		{
			name:         "USERNAME env var set",
			userEnv:      "",
			usernameEnv:  "testuser",
			wantNonEmpty: true,
		},
		{
			name:         "both env vars set, USER takes precedence",
			userEnv:      "user1",
			usernameEnv:  "user2",
			wantNonEmpty: true,
		},
		{
			name:         "no env vars set",
			userEnv:      "",
			usernameEnv:  "",
			wantNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("USER", tt.userEnv)
			_ = os.Setenv("USERNAME", tt.usernameEnv)

			user := getCurrentUser()
			if tt.wantNonEmpty && user == "" {
				t.Error("getCurrentUser() returned empty string, want non-empty")
			}
			if !tt.wantNonEmpty && user != "" {
				t.Errorf("getCurrentUser() = %v, want empty string", user)
			}

			// If userEnv is set, it should be returned
			if tt.userEnv != "" && user != tt.userEnv {
				t.Errorf("getCurrentUser() = %v, want %v", user, tt.userEnv)
			}
		})
	}

	// Test actual behavior
	user := getCurrentUser()
	// Should return something on most systems
	// Can be empty in some edge cases, so we just check it doesn't panic
	t.Logf("Current user: %s", user)
}

func TestGetCurrentUserGroups(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	groups, err := getCurrentUserGroups(ctx)
	if err != nil {
		t.Fatalf("getCurrentUserGroups() error = %v", err)
	}

	// Groups can be empty on some systems, so we just verify no error
	t.Logf("User groups: %v", groups)
}

func TestDetectOS(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	osType, err := detectOS(ctx)
	if err != nil {
		t.Fatalf("detectOS() error = %v", err)
	}

	// Verify basic fields are set
	if osType.Type == "" {
		t.Error("detectOS() Type is empty")
	}
	if osType.Arch == "" {
		t.Error("detectOS() Arch is empty")
	}

	// Type should match runtime.GOOS
	if osType.Type != runtime.GOOS {
		t.Errorf("detectOS() Type = %v, want %v", osType.Type, runtime.GOOS)
	}

	// Arch should match runtime.GOARCH
	if osType.Arch != runtime.GOARCH {
		t.Errorf("detectOS() Arch = %v, want %v", osType.Arch, runtime.GOARCH)
	}

	t.Logf("Detected OS: %s", osType.String())
	t.Logf("IsNixOS: %v, IsNixDarwin: %v", osType.IsNixOS, osType.IsNixDarwin)

	// Additional checks based on detected OS
	if osType.IsNixOS {
		if osType.Type != "linux" {
			t.Error("NixOS should have Type=linux")
		}
	}

	if osType.IsNixDarwin {
		if osType.Type != "darwin" {
			t.Error("nix-darwin should have Type=darwin")
		}
	}
}

func TestOSTypeString(t *testing.T) {
	tests := []struct {
		name string
		os   OSType
		want string
	}{
		{
			name: "NixOS",
			os: OSType{
				Type:    "linux",
				IsNixOS: true,
			},
			want: "NixOS",
		},
		{
			name: "nix-darwin",
			os: OSType{
				Type:        "darwin",
				IsNixDarwin: true,
			},
			want: "macOS (nix-darwin)",
		},
		{
			name: "plain macOS",
			os: OSType{
				Type: "darwin",
			},
			want: "macOS",
		},
		{
			name: "plain Linux",
			os: OSType{
				Type: "linux",
			},
			want: "Linux",
		},
		{
			name: "other OS",
			os: OSType{
				Type: "freebsd",
			},
			want: "freebsd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.os.String(); got != tt.want {
				t.Errorf("OSType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOSTypeNixConfigLabel(t *testing.T) {
	tests := []struct {
		name string
		os   OSType
		want string
	}{
		{
			name: "NixOS",
			os: OSType{
				IsNixOS: true,
			},
			want: "nixos configuration",
		},
		{
			name: "nix-darwin",
			os: OSType{
				IsNixDarwin: true,
			},
			want: "nix-darwin configuration",
		},
		{
			name: "other",
			os: OSType{
				Type: "linux",
			},
			want: "/etc/nix/nix.conf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.os.NixConfigLabel(); got != tt.want {
				t.Errorf("OSType.NixConfigLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectEnv(t *testing.T) {
	// This is more of an integration test
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env, err := DetectEnv(ctx)
	if err != nil {
		t.Fatalf("DetectEnv() error = %v", err)
	}

	// Verify env is populated
	if env == nil {
		t.Fatal("DetectEnv() returned nil")
	}

	// CurrentUser might be empty in some environments, so we just log it
	t.Logf("Current user: %s", env.User)
	t.Logf("User groups: %v", env.Groups)
	t.Logf("OS: %s", env.OS.String())
	t.Logf("OS Type: %s, Arch: %s", env.OS.Type, env.OS.Arch)
}

func TestDetectEnvWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := DetectEnv(ctx)
	// Should still work because we don't use context for most operations
	// The context is only passed to getCurrentUserGroups which handles errors
	if err != nil {
		t.Logf("DetectEnv() with canceled context error = %v (expected)", err)
	}
}
