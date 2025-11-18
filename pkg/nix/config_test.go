package nix

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestConfigValue_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"value": ["flakes", "nix-command"],
		"defaultValue": [],
		"description": "Experimental features to enable"
	}`

	var cv ConfigValue[[]string]
	err := json.Unmarshal([]byte(jsonData), &cv)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if len(cv.Value) != 2 {
		t.Errorf("ConfigValue.Value length = %d, want 2", len(cv.Value))
	}

	if cv.Value[0] != "flakes" {
		t.Errorf("ConfigValue.Value[0] = %s, want flakes", cv.Value[0])
	}

	if cv.Description == "" {
		t.Error("ConfigValue.Description is empty")
	}
}

func TestConfigValue_Int(t *testing.T) {
	jsonData := `{
		"value": 4,
		"defaultValue": 0,
		"description": "Number of CPU cores"
	}`

	var cv ConfigValue[int]
	err := json.Unmarshal([]byte(jsonData), &cv)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if cv.Value != 4 {
		t.Errorf("ConfigValue.Value = %d, want 4", cv.Value)
	}

	if cv.DefaultValue != 0 {
		t.Errorf("ConfigValue.DefaultValue = %d, want 0", cv.DefaultValue)
	}
}

func TestConfig_IsFlakesEnabled(t *testing.T) {
	tests := []struct {
		name     string
		features []string
		want     bool
	}{
		{
			name:     "flakes enabled",
			features: []string{"flakes", "nix-command"},
			want:     true,
		},
		{
			name:     "only nix-command",
			features: []string{"nix-command"},
			want:     true,
		},
		{
			name:     "no flakes",
			features: []string{"repl-flake"},
			want:     false,
		},
		{
			name:     "empty features",
			features: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ExperimentalFeatures: ConfigValue[[]string]{
					Value: tt.features,
				},
			}

			if got := config.IsFlakesEnabled(); got != tt.want {
				t.Errorf("Config.IsFlakesEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_HasFeature(t *testing.T) {
	config := &Config{
		ExperimentalFeatures: ConfigValue[[]string]{
			Value: []string{"flakes", "nix-command", "repl-flake"},
		},
	}

	tests := []struct {
		feature string
		want    bool
	}{
		{"flakes", true},
		{"nix-command", true},
		{"repl-flake", true},
		{"non-existent", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := config.HasFeature(tt.feature); got != tt.want {
				t.Errorf("Config.HasFeature(%s) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	// This is an integration test requiring Nix
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := GetConfig(ctx)
	if err != nil {
		// If nix is not installed, skip
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("nix command not found - skipping integration test")
		}
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Verify config is populated
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// System should be set
	if config.System.Value == "" {
		t.Error("GetConfig() System.Value is empty")
	}

	t.Logf("System: %s", config.System.Value)
	t.Logf("Experimental features: %v", config.ExperimentalFeatures.Value)
	t.Logf("Flakes enabled: %v", config.IsFlakesEnabled())
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"experimental-features": {
			"value": ["flakes"],
			"defaultValue": [],
			"description": "Experimental features"
		},
		"system": {
			"value": "x86_64-linux",
			"defaultValue": "x86_64-linux",
			"description": "System type"
		},
		"substituters": {
			"value": ["https://cache.nixos.org"],
			"defaultValue": [],
			"description": "Binary caches"
		},
		"max-jobs": {
			"value": 4,
			"defaultValue": 1,
			"description": "Max parallel jobs"
		},
		"cores": {
			"value": 0,
			"defaultValue": 0,
			"description": "CPU cores per job"
		}
	}`

	var config Config
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify parsed values
	if config.System.Value != "x86_64-linux" {
		t.Errorf("Config.System.Value = %s, want x86_64-linux", config.System.Value)
	}

	if len(config.ExperimentalFeatures.Value) != 1 || config.ExperimentalFeatures.Value[0] != "flakes" {
		t.Errorf("Config.ExperimentalFeatures.Value = %v, want [flakes]", config.ExperimentalFeatures.Value)
	}

	if config.MaxJobs.Value != 4 {
		t.Errorf("Config.MaxJobs.Value = %d, want 4", config.MaxJobs.Value)
	}

	if len(config.Substituters.Value) != 1 {
		t.Errorf("Config.Substituters.Value length = %d, want 1", len(config.Substituters.Value))
	}
}
