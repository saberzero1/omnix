package init

import (
	"context"
	"os"
	"path/filepath"
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
