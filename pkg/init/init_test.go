package init

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceAction_HasValue(t *testing.T) {
	tests := []struct {
		name     string
		action   ReplaceAction
		expected bool
	}{
		{
			name:     "nil value",
			action:   ReplaceAction{Placeholder: "FOO", Value: nil},
			expected: false,
		},
		{
			name:     "empty value",
			action:   ReplaceAction{Placeholder: "FOO", Value: strPtr("")},
			expected: true,
		},
		{
			name:     "non-empty value",
			action:   ReplaceAction{Placeholder: "FOO", Value: strPtr("bar")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.HasValue())
		})
	}
}

func TestReplaceAction_String(t *testing.T) {
	tests := []struct {
		name     string
		action   ReplaceAction
		expected string
	}{
		{
			name:     "disabled",
			action:   ReplaceAction{Placeholder: "FOO", Value: nil},
			expected: "replace [disabled]",
		},
		{
			name:     "enabled",
			action:   ReplaceAction{Placeholder: "FOO", Value: strPtr("bar")},
			expected: "replace [FOO => bar]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.String())
		})
	}
}

func TestReplaceAction_Apply(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("Hello PLACEHOLDER world"), 0644))

	placeholderFile := filepath.Join(tmpDir, "PLACEHOLDER.txt")
	require.NoError(t, os.WriteFile(placeholderFile, []byte("content"), 0644))

	// Apply replace action
	action := ReplaceAction{
		Placeholder: "PLACEHOLDER",
		Value:       strPtr("REPLACED"),
	}

	err := action.Apply(ctx, tmpDir)
	require.NoError(t, err)

	// Check that file content was replaced
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Hello REPLACED world", string(content))

	// Check that file was renamed
	_, err = os.Stat(placeholderFile)
	assert.True(t, os.IsNotExist(err), "Original file should be renamed")

	renamedFile := filepath.Join(tmpDir, "REPLACED.txt")
	_, err = os.Stat(renamedFile)
	assert.NoError(t, err, "Renamed file should exist")
}

func TestRetainAction_HasValue(t *testing.T) {
	tests := []struct {
		name     string
		action   RetainAction
		expected bool
	}{
		{
			name:     "nil value",
			action:   RetainAction{Paths: []string{"*.txt"}, Value: nil},
			expected: false,
		},
		{
			name:     "false value",
			action:   RetainAction{Paths: []string{"*.txt"}, Value: boolPtr(false)},
			expected: true,
		},
		{
			name:     "true value",
			action:   RetainAction{Paths: []string{"*.txt"}, Value: boolPtr(true)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.HasValue())
		})
	}
}

func TestRetainAction_Apply(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create test files
	txtFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(txtFile, []byte("content"), 0644))

	mdFile := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(mdFile, []byte("readme"), 0644))

	// Apply retain action to delete *.txt files
	action := RetainAction{
		Paths: []string{"*.txt"},
		Value: boolPtr(false),
	}

	err := action.Apply(ctx, tmpDir)
	require.NoError(t, err)

	// Check that txt file was deleted
	_, err = os.Stat(txtFile)
	assert.True(t, os.IsNotExist(err), "txt file should be deleted")

	// Check that md file still exists
	_, err = os.Stat(mdFile)
	assert.NoError(t, err, "md file should still exist")
}

func TestActionPriority(t *testing.T) {
	retain := RetainAction{}
	replace := ReplaceAction{}

	assert.Less(t, ActionPriority(&retain), ActionPriority(&replace),
		"Retain should have higher priority (lower value) than Replace")
}

func TestTemplate_SetParamValues(t *testing.T) {
	replaceVal := "bar"
	template := &Template{
		Params: []Param{
			{
				Name: "foo",
				Action: &ReplaceAction{
					Placeholder: "FOO",
				},
			},
			{
				Name: "keep",
				Action: &RetainAction{
					Paths: []string{"*.txt"},
				},
			},
		},
	}

	values := map[string]interface{}{
		"foo":  "bar",
		"keep": true,
	}

	template.SetParamValues(values)

	// Check that replace action got the value
	replaceAction := template.Params[0].Action.(*ReplaceAction)
	assert.Equal(t, &replaceVal, replaceAction.Value)

	// Check that retain action got the value
	retainAction := template.Params[1].Action.(*RetainAction)
	assert.Equal(t, boolPtr(true), retainAction.Value)
}

func TestParam_String(t *testing.T) {
	param := Param{
		Name:        "test",
		Description: "Test parameter",
		Action: ReplaceAction{
			Placeholder: "FOO",
			Value:       strPtr("bar"),
		},
	}

	str := param.String()
	assert.Contains(t, str, "test")
	assert.Contains(t, str, "replace")
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "Simple basename match",
			pattern: "*.txt",
			path:    "file.txt",
			want:    true,
		},
		{
			name:    "Basename no match",
			pattern: "*.txt",
			path:    "file.md",
			want:    false,
		},
		{
			name:    "Doublestar prefix",
			pattern: "**/*.go",
			path:    "pkg/main.go",
			want:    true,
		},
		{
			name:    "Doublestar suffix",
			pattern: "src/**",
			path:    "src/file.txt",
			want:    true,
		},
		{
			name:    "Directory match",
			pattern: ".github/**",
			path:    ".github/workflows/ci.yml",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchGlob(tt.pattern, tt.path); got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
