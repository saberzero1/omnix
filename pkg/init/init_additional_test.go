package init

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceAction_String_Detailed(t *testing.T) {
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
			name:     "with value",
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

func TestRetainAction_String_Detailed(t *testing.T) {
	tests := []struct {
		name     string
		action   RetainAction
		expected string
	}{
		{
			name:     "disabled - nil",
			action:   RetainAction{Paths: []string{"*.txt"}, Value: nil},
			expected: "prune [disabled]",
		},
		{
			name:     "disabled - true",
			action:   RetainAction{Paths: []string{"*.txt"}, Value: boolPtr(true)},
			expected: "prune [disabled]",
		},
		{
			name:     "enabled",
			action:   RetainAction{Paths: []string{"*.txt", "*.md"}, Value: boolPtr(false)},
			expected: "prune [*.txt, *.md]",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.action.String())
		})
	}
}

func TestActionPriority_Ordering(t *testing.T) {
	retain := &RetainAction{}
	replace := &ReplaceAction{}
	
	retainPriority := ActionPriority(retain)
	replacePriority := ActionPriority(replace)
	
	assert.Less(t, retainPriority, replacePriority, 
		"Retain actions should have higher priority (lower number) than Replace")
	assert.Equal(t, 0, retainPriority, "Retain should be priority 0")
	assert.Equal(t, 1, replacePriority, "Replace should be priority 1")
}

func TestReplaceAction_ApplyToFileContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create a test file with placeholder
	testFile := filepath.Join(tmpDir, "config.txt")
	content := "Server: HOSTNAME\nPort: 8080"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))
	
	// Apply replacement
	action := ReplaceAction{
		Placeholder: "HOSTNAME",
		Value:       strPtr("example.com"),
	}
	
	err := action.Apply(ctx, tmpDir)
	require.NoError(t, err)
	
	// Verify content was replaced
	newContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Server: example.com\nPort: 8080", string(newContent))
}

func TestReplaceAction_RenameDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create a directory with placeholder in name
	placeholderDir := filepath.Join(tmpDir, "MYPROJECT-src")
	require.NoError(t, os.MkdirAll(placeholderDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(placeholderDir, "main.go"), []byte("package main"), 0644))
	
	// Apply replacement
	action := ReplaceAction{
		Placeholder: "MYPROJECT",
		Value:       strPtr("awesome-app"),
	}
	
	err := action.Apply(ctx, tmpDir)
	require.NoError(t, err)
	
	// Verify directory was renamed
	_, err = os.Stat(placeholderDir)
	assert.True(t, os.IsNotExist(err), "Original directory should be renamed")
	
	newDir := filepath.Join(tmpDir, "awesome-app-src")
	info, err := os.Stat(newDir)
	assert.NoError(t, err, "Renamed directory should exist")
	assert.True(t, info.IsDir())
	
	// Verify file still exists in renamed directory
	_, err = os.Stat(filepath.Join(newDir, "main.go"))
	assert.NoError(t, err, "File should exist in renamed directory")
}

func TestRetainAction_DeleteMultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "keep.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "delete.txt"), []byte("delete me"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "also-delete.txt"), []byte("delete me too"), 0644))
	
	// Apply retain action to delete .txt files
	action := RetainAction{
		Paths: []string{"*.txt"},
		Value: boolPtr(false),
	}
	
	err := action.Apply(ctx, tmpDir)
	require.NoError(t, err)
	
	// Verify .txt files were deleted
	_, err = os.Stat(filepath.Join(tmpDir, "delete.txt"))
	assert.True(t, os.IsNotExist(err), "delete.txt should be deleted")
	
	_, err = os.Stat(filepath.Join(tmpDir, "also-delete.txt"))
	assert.True(t, os.IsNotExist(err), "also-delete.txt should be deleted")
	
	// Verify .go file still exists
	_, err = os.Stat(filepath.Join(tmpDir, "keep.go"))
	assert.NoError(t, err, "keep.go should still exist")
}

func TestTemplate_ScaffoldAt_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	tmpDir := t.TempDir()
	
	// Create template directory
	templateDir := filepath.Join(tmpDir, "template")
	require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "src"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "README.md"), []byte("# APPNAME"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "src", "main.go"), []byte("package main"), 0644))
	
	// Create template with parameters
	welcomeText := "Welcome to your new project!"
	template := &Template{
		Path:        templateDir,
		WelcomeText: &welcomeText,
		Params: []Param{
			{
				Name:        "appname",
				Description: "Application name",
				Action: &ReplaceAction{
					Placeholder: "APPNAME",
					Value:       strPtr("MyAwesomeApp"),
				},
			},
		},
	}
	
	// Scaffold template
	outputDir := filepath.Join(tmpDir, "output")
	outPath, err := template.ScaffoldAt(ctx, outputDir)
	require.NoError(t, err)
	assert.NotEmpty(t, outPath)
	
	// Verify files were copied
	_, err = os.Stat(filepath.Join(outputDir, "README.md"))
	assert.NoError(t, err, "README.md should be copied")
	
	_, err = os.Stat(filepath.Join(outputDir, "src", "main.go"))
	assert.NoError(t, err, "src/main.go should be copied")
	
	// Verify replacement was applied
	content, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# MyAwesomeApp", string(content))
}

func TestTemplate_SetParamValues_MultipleTypes(t *testing.T) {
	template := &Template{
		Params: []Param{
			{
				Name: "name",
				Action: &ReplaceAction{
					Placeholder: "NAME",
				},
			},
			{
				Name: "include-tests",
				Action: &RetainAction{
					Paths: []string{"tests/**"},
				},
			},
		},
	}
	
	values := map[string]interface{}{
		"name":          "my-app",
		"include-tests": false,
	}
	
	template.SetParamValues(values)
	
	// Verify replace action got string value
	replaceAction := template.Params[0].Action.(*ReplaceAction)
	assert.NotNil(t, replaceAction.Value)
	assert.Equal(t, "my-app", *replaceAction.Value)
	
	// Verify retain action got bool value
	retainAction := template.Params[1].Action.(*RetainAction)
	assert.NotNil(t, retainAction.Value)
	assert.False(t, *retainAction.Value)
}

func TestParam_SetValue_WrongType(t *testing.T) {
	// Test that SetValue handles wrong types gracefully
	param := &Param{
		Name: "test",
		Action: &ReplaceAction{
			Placeholder: "TEST",
		},
	}
	
	// Try to set a non-string value to ReplaceAction
	param.SetValue(123) // Should not panic
	
	// Value should remain nil
	replaceAction := param.Action.(*ReplaceAction)
	assert.Nil(t, replaceAction.Value)
}

func TestFindAllPaths_Recursive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	tmpDir := t.TempDir()
	
	// Create nested directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "dir1", "dir2"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "dir1", "file2.txt"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "dir1", "dir2", "file3.txt"), []byte("content"), 0644))
	
	paths, err := findAllPaths(tmpDir)
	require.NoError(t, err)
	
	// Should find all files and directories (but not root ".")
	assert.Contains(t, paths, "file1.txt")
	assert.Contains(t, paths, "dir1")
	assert.Contains(t, paths, filepath.Join("dir1", "file2.txt"))
	assert.Contains(t, paths, filepath.Join("dir1", "dir2"))
	assert.Contains(t, paths, filepath.Join("dir1", "dir2", "file3.txt"))
}
