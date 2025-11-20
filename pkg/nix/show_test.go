package nix

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFlakeOutputs_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVal bool
		wantSet bool
		wantErr bool
	}{
		{
			name:    "terminal value",
			input:   `{"type":"derivation","shortDescription":"A test package"}`,
			wantVal: true,
			wantSet: false,
		},
		{
			name:    "attribute set",
			input:   `{"foo":{"type":"derivation"},"bar":{"type":"derivation"}}`,
			wantVal: false,
			wantSet: true,
		},
		{
			name:    "nested attribute set",
			input:   `{"x86_64-linux":{"default":{"type":"derivation"}}}`,
			wantVal: false,
			wantSet: true,
		},
		{
			name:    "invalid json",
			input:   `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputs FlakeOutputs
			err := json.Unmarshal([]byte(tt.input), &outputs)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			if tt.wantVal && outputs.Val == nil {
				t.Error("expected Val to be non-nil")
			}
			if !tt.wantVal && outputs.Val != nil {
				t.Error("expected Val to be nil")
			}
			if tt.wantSet && outputs.Attrset == nil {
				t.Error("expected Attrset to be non-nil")
			}
			if !tt.wantSet && outputs.Attrset != nil {
				t.Error("expected Attrset to be nil")
			}
		})
	}
}

func TestFlakeOutputs_GetByPath(t *testing.T) {
	// Create a test structure manually
	outputs := &FlakeOutputs{
		Attrset: map[string]*FlakeOutputs{
			"packages": {
				Attrset: map[string]*FlakeOutputs{
					"x86_64-linux": {
						Attrset: map[string]*FlakeOutputs{
							"default": {
								Val: &FlakeVal{
									ShortDescription: "Default package",
									Type:             "derivation",
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name    string
		path    []string
		wantNil bool
	}{
		{
			name:    "empty path",
			path:    []string{},
			wantNil: false,
		},
		{
			name:    "single level",
			path:    []string{"packages"},
			wantNil: false,
		},
		{
			name:    "multi level",
			path:    []string{"packages", "x86_64-linux", "default"},
			wantNil: false,
		},
		{
			name:    "non-existent path",
			path:    []string{"packages", "aarch64-darwin"},
			wantNil: true,
		},
		{
			name:    "invalid path",
			path:    []string{"does-not-exist"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := outputs.GetByPath(tt.path...)
			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestFlakeOutputs_GetAttrsetOfVal(t *testing.T) {
	tests := []struct {
		name      string
		outputs   *FlakeOutputs
		wantCount int
	}{
		{
			name: "empty attrset",
			outputs: &FlakeOutputs{
				Attrset: map[string]*FlakeOutputs{},
			},
			wantCount: 0,
		},
		{
			name: "attrset with values",
			outputs: &FlakeOutputs{
				Attrset: map[string]*FlakeOutputs{
					"foo": {
						Val: &FlakeVal{ShortDescription: "Foo package"},
					},
					"bar": {
						Val: &FlakeVal{ShortDescription: "Bar package"},
					},
				},
			},
			wantCount: 2,
		},
		{
			name: "mixed attrset",
			outputs: &FlakeOutputs{
				Attrset: map[string]*FlakeOutputs{
					"foo": {
						Val: &FlakeVal{ShortDescription: "Foo package"},
					},
					"nested": {
						Attrset: map[string]*FlakeOutputs{
							"bar": {
								Val: &FlakeVal{ShortDescription: "Bar package"},
							},
						},
					},
				},
			},
			wantCount: 1, // Only terminal values at this level
		},
		{
			name: "nil attrset",
			outputs: &FlakeOutputs{
				Val: &FlakeVal{ShortDescription: "Single value"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.outputs.GetAttrsetOfVal()
			if len(result) != tt.wantCount {
				t.Errorf("GetAttrsetOfVal() returned %d values, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestFlakeShow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	cmd := NewCmd()

	// Test with a known flake
	flakeURL := NewFlakeURL("github:saberzero1/omnix")

	metadata, err := cmd.FlakeShow(ctx, flakeURL)
	if err != nil {
		t.Skipf("Failed to show flake (might not have network access): %v", err)
	}

	if metadata == nil {
		t.Fatal("expected non-nil metadata")
	}

	// Outputs may or may not be present depending on the flake and nix version
	// Just verify that the metadata was returned successfully
	t.Logf("Flake show succeeded, outputs present: %v", metadata.Outputs != nil)
}

func TestFlakeShow_Local(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	cmd := NewCmd()

	// Test with current directory if it's a flake
	flakeURL := NewFlakeURL(".")

	metadata, err := cmd.FlakeShow(ctx, flakeURL)
	if err != nil {
		// It's okay if the current directory is not a flake
		t.Skipf("Current directory is not a flake: %v", err)
	}

	if metadata == nil {
		t.Fatal("expected non-nil metadata")
	}

	// Just verify we can access outputs without error
	if metadata.Outputs != nil {
		_ = metadata.Outputs.GetByPath("packages")
	}
}

func TestFlakeMetadata_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOutputs bool
		wantErr     bool
	}{
		{
			name:        "metadata with outputs",
			input:       `{"description":"A test flake","outputs":{"packages":{"x86_64-linux":{"default":{"type":"derivation"}}}}}`,
			wantOutputs: true,
			wantErr:     false,
		},
		{
			name:        "metadata without outputs",
			input:       `{"description":"A test flake"}`,
			wantOutputs: false,
			wantErr:     false,
		},
		{
			name:        "empty metadata",
			input:       `{}`,
			wantOutputs: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metadata FlakeMetadata
			err := json.Unmarshal([]byte(tt.input), &metadata)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantOutputs && metadata.Outputs == nil {
					t.Error("expected Outputs to be non-nil")
				}
				if !tt.wantOutputs && metadata.Outputs != nil {
					t.Error("expected Outputs to be nil")
				}
			}
		})
	}
}
