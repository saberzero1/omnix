package init

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChmodAction_Apply(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(testFile, []byte("#!/bin/bash\necho test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Apply chmod action
	trueVal := true
	action := ChmodAction{
		Paths: []string{"*.sh"},
		Mode:  0755,
		Value: &trueVal,
	}

	ctx := context.Background()
	if err := action.Apply(ctx, tmpDir); err != nil {
		t.Fatalf("ChmodAction.Apply() failed: %v", err)
	}

	// Verify permissions changed
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected mode 0755, got %o", info.Mode().Perm())
	}
}

func TestChmodAction_Disabled(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Disabled action
	falseVal := false
	action := ChmodAction{
		Paths: []string{"*.sh"},
		Mode:  0755,
		Value: &falseVal,
	}

	ctx := context.Background()
	if err := action.Apply(ctx, tmpDir); err != nil {
		t.Fatalf("ChmodAction.Apply() failed: %v", err)
	}

	// Verify permissions unchanged
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0644 {
		t.Errorf("Expected mode 0644 (unchanged), got %o", info.Mode().Perm())
	}
}

func TestMoveAction_Apply(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "old.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Apply move action
	trueVal := true
	action := MoveAction{
		From:  "old.txt",
		To:    "new.txt",
		Value: &trueVal,
	}

	ctx := context.Background()
	if err := action.Apply(ctx, tmpDir); err != nil {
		t.Fatalf("MoveAction.Apply() failed: %v", err)
	}

	// Verify file moved
	dstFile := filepath.Join(tmpDir, "new.txt")
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should not exist after move")
	}

	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("Expected 'content', got '%s'", string(content))
	}
}

func TestMoveAction_Disabled(t *testing.T) {
	tmpDir := t.TempDir()

	srcFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Disabled action
	falseVal := false
	action := MoveAction{
		From:  "file.txt",
		To:    "moved.txt",
		Value: &falseVal,
	}

	ctx := context.Background()
	if err := action.Apply(ctx, tmpDir); err != nil {
		t.Fatalf("MoveAction.Apply() failed: %v", err)
	}

	// Verify file not moved
	if _, err := os.Stat(srcFile); err != nil {
		t.Error("Source file should still exist when action is disabled")
	}
}

func TestActionPriority_NewActions(t *testing.T) {
	trueVal := true

	tests := []struct {
		name     string
		action   Action
		expected int
	}{
		{"RetainAction", &RetainAction{}, 0},
		{"ReplaceAction", &ReplaceAction{}, 1},
		{"ChmodAction", &ChmodAction{Value: &trueVal}, 2},
		{"MoveAction", &MoveAction{Value: &trueVal}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := ActionPriority(tt.action)
			if priority != tt.expected {
				t.Errorf("ActionPriority() = %d, want %d", priority, tt.expected)
			}
		})
	}
}

func TestMoveAction_MultipleFilesError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files that match the pattern
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to move multiple files to single destination (should fail)
	trueVal := true
	action := MoveAction{
		From:  "*.txt",
		To:    "output.txt",
		Value: &trueVal,
	}

	ctx := context.Background()
	err := action.Apply(ctx, tmpDir)

	// Should get an error about multiple matches
	if err == nil {
		t.Fatal("Expected error when moving multiple files to single destination, got nil")
	}

	if !strings.Contains(err.Error(), "matched 2 files") {
		t.Errorf("Expected error message about 2 files, got: %v", err)
	}
}

func TestChmodAction_HasValue(t *testing.T) {
	tests := []struct {
		name     string
		action   ChmodAction
		expected bool
	}{
		{
			name:     "nil value",
			action:   ChmodAction{Value: nil},
			expected: false,
		},
		{
			name:     "false value",
			action:   ChmodAction{Value: boolPtr(false)},
			expected: false,
		},
		{
			name:     "true value",
			action:   ChmodAction{Value: boolPtr(true)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.HasValue(); got != tt.expected {
				t.Errorf("ChmodAction.HasValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestChmodAction_String(t *testing.T) {
	tests := []struct {
		name     string
		action   ChmodAction
		contains string
	}{
		{
			name:     "disabled action (nil)",
			action:   ChmodAction{Value: nil, Paths: []string{"*.sh"}, Mode: 0755},
			contains: "disabled",
		},
		{
			name:     "disabled action (false)",
			action:   ChmodAction{Value: boolPtr(false), Paths: []string{"*.sh"}, Mode: 0755},
			contains: "disabled",
		},
		{
			name:     "enabled action",
			action:   ChmodAction{Value: boolPtr(true), Paths: []string{"*.sh"}, Mode: 0755},
			contains: "chmod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.action.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("ChmodAction.String() = %v, want to contain %v", got, tt.contains)
			}
		})
	}
}

func TestMoveAction_HasValue(t *testing.T) {
	tests := []struct {
		name     string
		action   MoveAction
		expected bool
	}{
		{
			name:     "nil value",
			action:   MoveAction{Value: nil},
			expected: false,
		},
		{
			name:     "false value",
			action:   MoveAction{Value: boolPtr(false)},
			expected: false,
		},
		{
			name:     "true value",
			action:   MoveAction{Value: boolPtr(true)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.HasValue(); got != tt.expected {
				t.Errorf("MoveAction.HasValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMoveAction_String(t *testing.T) {
	tests := []struct {
		name     string
		action   MoveAction
		contains string
	}{
		{
			name:     "disabled action (nil)",
			action:   MoveAction{Value: nil, From: "a.txt", To: "b.txt"},
			contains: "disabled",
		},
		{
			name:     "disabled action (false)",
			action:   MoveAction{Value: boolPtr(false), From: "a.txt", To: "b.txt"},
			contains: "disabled",
		},
		{
			name:     "enabled action",
			action:   MoveAction{Value: boolPtr(true), From: "a.txt", To: "b.txt"},
			contains: "move",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.action.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("MoveAction.String() = %v, want to contain %v", got, tt.contains)
			}
		})
	}
}
