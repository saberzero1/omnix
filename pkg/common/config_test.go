package common

import (
	"encoding/json"
	"testing"
)

func TestParseYAMLConfig(t *testing.T) {
	yamlContent := `
ci:
  default:
    build: true
    test: true
  custom:
    build: false
health:
  default:
    checks: 5
`

	tree, err := ParseYAMLConfig(yamlContent)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() failed: %v", err)
	}

	if tree == nil {
		t.Fatal("ParseYAMLConfig() returned nil tree")
	}

	// Test Get method
	var ciConfig map[string]interface{}
	if err := tree.Get("ci", &ciConfig); err != nil {
		t.Errorf("Get('ci') failed: %v", err)
	}

	if len(ciConfig) == 0 {
		t.Error("Get('ci') returned empty map")
	}
}

func TestOmConfigTreeGet(t *testing.T) {
	yamlContent := `
ci:
  default:
    build: true
    steps: ["check", "build"]
  production:
    build: false
    steps: ["deploy"]
`

	tree, err := ParseYAMLConfig(yamlContent)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() failed: %v", err)
	}

	// Test getting existing key
	var ciConfig map[string]json.RawMessage
	if err := tree.Get("ci", &ciConfig); err != nil {
		t.Errorf("Get('ci') failed: %v", err)
	}

	if len(ciConfig) != 2 {
		t.Errorf("Get('ci') returned %d items, want 2", len(ciConfig))
	}

	// Test getting non-existing key
	var missingConfig map[string]json.RawMessage
	if err := tree.Get("nonexistent", &missingConfig); err != nil {
		t.Errorf("Get('nonexistent') should not fail: %v", err)
	}

	if missingConfig != nil {
		t.Error("Get('nonexistent') should return nil for missing key")
	}
}

func TestGetSubConfigUnder(t *testing.T) {
	yamlContent := `
ci:
  default:
    enabled: true
  custom:
    enabled: false
`

	tree, err := ParseYAMLConfig(yamlContent)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() failed: %v", err)
	}

	tests := []struct {
		name      string
		reference []string
		rootKey   string
		wantErr   bool
		wantRest  int
	}{
		{
			name:      "no reference defaults to 'default'",
			reference: []string{},
			rootKey:   "ci",
			wantErr:   false,
			wantRest:  0,
		},
		{
			name:      "with reference 'custom'",
			reference: []string{"custom"},
			rootKey:   "ci",
			wantErr:   false,
			wantRest:  0,
		},
		{
			name:      "with nested reference",
			reference: []string{"custom", "extra"},
			rootKey:   "ci",
			wantErr:   false,
			wantRest:  1,
		},
		{
			name:      "missing reference",
			reference: []string{"missing"},
			rootKey:   "ci",
			wantErr:   true,
			wantRest:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OmConfig{
				FlakeURL:  ".",
				Reference: tt.reference,
				Config:    tree,
			}

			var result map[string]interface{}
			rest, err := config.GetSubConfigUnder(tt.rootKey, map[string]interface{}{}, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubConfigUnder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(rest) != tt.wantRest {
				t.Errorf("GetSubConfigUnder() rest length = %d, want %d", len(rest), tt.wantRest)
			}
		})
	}
}

func TestParseJSONConfig(t *testing.T) {
	jsonContent := `{
		"ci": {
			"default": {
				"build": true
			}
		}
	}`

	tree, err := ParseJSONConfig(jsonContent)
	if err != nil {
		t.Fatalf("ParseJSONConfig() failed: %v", err)
	}

	if tree == nil {
		t.Fatal("ParseJSONConfig() returned nil tree")
	}

	var ciConfig map[string]interface{}
	if err := tree.Get("ci", &ciConfig); err != nil {
		t.Errorf("Get('ci') failed: %v", err)
	}
}

func TestOmConfigTreeMarshalYAML(t *testing.T) {
	yamlContent := `
ci:
  default:
    build: true
`

	tree, err := ParseYAMLConfig(yamlContent)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() failed: %v", err)
	}

	// Try to marshal back to YAML
	data, err := tree.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() failed: %v", err)
	}

	if data == nil {
		t.Error("MarshalYAML() returned nil")
	}
}

func TestEmptyConfig(t *testing.T) {
	yamlContent := ``

	tree, err := ParseYAMLConfig(yamlContent)
	if err != nil {
		t.Fatalf("ParseYAMLConfig() failed on empty content: %v", err)
	}

	// Get should return nil for any key
	var result map[string]interface{}
	if err := tree.Get("anything", &result); err != nil {
		t.Errorf("Get() on empty config should not error: %v", err)
	}

	if result != nil {
		t.Error("Get() on empty config should return nil")
	}
}
