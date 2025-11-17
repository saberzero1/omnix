package common

import (
	"os"
	"testing"
)

func TestNixInstalled(t *testing.T) {
	// This test will pass or fail depending on the system
	// We just verify the function doesn't panic
	result := NixInstalled()
	t.Logf("NixInstalled() = %v", result)
}

func TestWhichStrict(t *testing.T) {
	tests := []struct {
		name       string
		binary     string
		wantEmpty  bool
		shouldFind bool
	}{
		{
			name:       "find sh",
			binary:     "sh",
			wantEmpty:  false,
			shouldFind: true,
		},
		{
			name:       "nonexistent binary",
			binary:     "this-binary-should-never-exist-12345",
			wantEmpty:  true,
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WhichStrict(tt.binary)
			isEmpty := got == ""

			if isEmpty != tt.wantEmpty {
				t.Errorf("WhichStrict(%q) returned %q, isEmpty=%v, want isEmpty=%v",
					tt.binary, got, isEmpty, tt.wantEmpty)
			}

			if tt.shouldFind && isEmpty {
				t.Errorf("WhichStrict(%q) expected to find binary but got empty string", tt.binary)
			}
		})
	}
}

func TestWhichStrictPanic(t *testing.T) {
	// We can't easily test the panic case without mocking exec.LookPath
	// This is just a placeholder to document the expected panic behavior
	t.Skip("Skipping panic test - requires mocking")
}

func TestWhichStrictWithGo(t *testing.T) {
	// Try to find the go binary which should be available in this test environment
	goPath := WhichStrict("go")
	if goPath == "" {
		t.Skip("Go binary not found in PATH, skipping test")
	}

	// Verify the path exists
	if _, err := os.Stat(goPath); err != nil {
		t.Errorf("WhichStrict returned path %q that doesn't exist: %v", goPath, err)
	}
}
