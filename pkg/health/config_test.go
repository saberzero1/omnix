package health

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test config file
	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `health:
  nix-version:
    min-version: "2.18.0"
  caches:
    required:
      - "https://cache.nixos.org"
      - "https://my-cache.cachix.org"
  trusted-users:
    enable: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify NixVersion config
	if config.NixVersion == nil {
		t.Error("Expected NixVersion config to be set")
	} else if config.NixVersion.MinVersion != "2.18.0" {
		t.Errorf("Expected MinVersion to be '2.18.0', got '%s'", config.NixVersion.MinVersion)
	}

	// Verify Caches config
	if config.Caches == nil {
		t.Error("Expected Caches config to be set")
	} else {
		if len(config.Caches.Required) != 2 {
			t.Errorf("Expected 2 required caches, got %d", len(config.Caches.Required))
		}
		if config.Caches.Required[0] != "https://cache.nixos.org" {
			t.Errorf("Expected first cache to be 'https://cache.nixos.org', got '%s'", config.Caches.Required[0])
		}
	}

	// Verify TrustedUsers config
	if config.TrustedUsers == nil {
		t.Error("Expected TrustedUsers config to be set")
	} else if !config.TrustedUsers.Enable {
		t.Error("Expected TrustedUsers to be enabled")
	}
}

func TestLoadConfigNonexistent(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/om.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent config file")
	}
}

func TestApplyConfig(t *testing.T) {
	// Create a default NixHealth
	h := Default()

	// Create a config
	minVersion := "2.20.0"
	config := Config{
		NixVersion: &NixVersionConfig{
			MinVersion: minVersion,
		},
		Caches: &CachesConfig{
			Required: []string{
				"https://custom-cache.example.com",
			},
		},
		TrustedUsers: &TrustedUsersConfig{
			Enable: true,
		},
	}

	// Apply the config
	config.ApplyConfig(h)

	// Verify the config was applied
	if h.NixVersion.MinVersion.String() != "2.20.0" {
		t.Errorf("Expected MinVersion to be '2.20.0', got '%s'", h.NixVersion.MinVersion.String())
	}

	if len(h.Caches.Required) != 1 {
		t.Errorf("Expected 1 required cache, got %d", len(h.Caches.Required))
	}

	if !h.TrustedUsers.Enable {
		t.Error("Expected TrustedUsers to be enabled")
	}
}

func TestNewFromConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test with nonexistent config (should use defaults)
	h, err := NewFromConfig(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("Expected no error with nonexistent config, got: %v", err)
	}

	// Verify defaults were used
	if h.NixVersion.MinVersion.String() == "" {
		t.Error("Expected default MinVersion to be set")
	}

	// Test with valid config
	configPath := filepath.Join(tmpDir, "om.yaml")
	configContent := `health:
  nix-version:
    min-version: "2.19.0"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	h, err = NewFromConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to create NixHealth from config: %v", err)
	}

	if h.NixVersion.MinVersion.String() != "2.19.0" {
		t.Errorf("Expected MinVersion to be '2.19.0', got '%s'", h.NixVersion.MinVersion.String())
	}
}

func TestDefaultCachesConfig(t *testing.T) {
	caches := DefaultCachesConfig()
	
	// Verify it returns the expected default caches
	if len(caches) != 1 {
		t.Errorf("Expected 1 default cache, got %d", len(caches))
	}
	
	if len(caches) > 0 && caches[0] != "https://cache.nixos.org" {
		t.Errorf("Expected default cache to be 'https://cache.nixos.org', got '%s'", caches[0])
	}
}
