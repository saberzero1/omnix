package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDirAll(t *testing.T) {
	// Create a temporary source directory
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create test structure
	testFile := filepath.Join(srcDir, "test.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(srcDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Test copying
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDirAll() failed: %v", err)
	}

	// Verify files were copied
	copiedFile := filepath.Join(dstDir, "test.txt")
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Errorf("Failed to read copied file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Copied file content = %q, want %q", string(content), testContent)
	}

	// Verify subdirectory was copied
	copiedSubFile := filepath.Join(dstDir, "subdir", "subfile.txt")
	if _, err := os.Stat(copiedSubFile); err != nil {
		t.Errorf("Subdirectory file not copied: %v", err)
	}
}

func TestCopyDirAllWithSymlink(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a file and a symlink to it
	testFile := filepath.Join(srcDir, "original.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	linkPath := filepath.Join(srcDir, "link.txt")
	if err := os.Symlink("original.txt", linkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Copy directory
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDirAll() failed: %v", err)
	}

	// Verify symlink was preserved
	copiedLink := filepath.Join(dstDir, "link.txt")
	info, err := os.Lstat(copiedLink)
	if err != nil {
		t.Fatalf("Failed to stat copied symlink: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Symlink was not preserved during copy")
	}
}

func TestCopyDirAllWithExistingSymlink(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a file and a symlink to it in source
	testFile := filepath.Join(srcDir, "original.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	linkPath := filepath.Join(srcDir, "link.txt")
	if err := os.Symlink("original.txt", linkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// First copy - should succeed
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("First CopyDirAll() failed: %v", err)
	}

	// Verify symlink was created
	copiedLink := filepath.Join(dstDir, "link.txt")
	target, err := os.Readlink(copiedLink)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if target != "original.txt" {
		t.Errorf("Symlink target = %q, want %q", target, "original.txt")
	}

	// Modify source to point to a different target
	if err := os.Remove(linkPath); err != nil {
		t.Fatalf("Failed to remove original symlink: %v", err)
	}
	testFile2 := filepath.Join(srcDir, "other.txt")
	if err := os.WriteFile(testFile2, []byte("other content"), 0644); err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}
	if err := os.Symlink("other.txt", linkPath); err != nil {
		t.Fatalf("Failed to create new symlink: %v", err)
	}

	// Second copy - should succeed and update the symlink
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("Second CopyDirAll() failed: %v", err)
	}

	// Verify symlink was updated
	target, err = os.Readlink(copiedLink)
	if err != nil {
		t.Fatalf("Failed to read updated symlink: %v", err)
	}
	if target != "other.txt" {
		t.Errorf("Updated symlink target = %q, want %q", target, "other.txt")
	}
}

func TestCopyDirAllTypeTransition(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a file and a symlink to it in source
	testFile := filepath.Join(srcDir, "original.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	linkPath := filepath.Join(srcDir, "link.txt")
	if err := os.Symlink("original.txt", linkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// First copy - creates symlink at dst/link.txt
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("First CopyDirAll() failed: %v", err)
	}

	// Verify symlink was created
	copiedLink := filepath.Join(dstDir, "link.txt")
	info, err := os.Lstat(copiedLink)
	if err != nil {
		t.Fatalf("Failed to stat copied path: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("Expected symlink, got regular file")
	}

	// Change source: replace symlink with regular file
	if err := os.Remove(linkPath); err != nil {
		t.Fatalf("Failed to remove symlink: %v", err)
	}
	if err := os.WriteFile(linkPath, []byte("now a regular file"), 0644); err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	// Second copy - should replace symlink with regular file
	if err := CopyDirAll(srcDir, dstDir); err != nil {
		t.Fatalf("Second CopyDirAll() failed: %v", err)
	}

	// Verify it's now a regular file, not a symlink
	info, err = os.Lstat(copiedLink)
	if err != nil {
		t.Fatalf("Failed to stat updated path: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("Expected regular file, but got symlink")
	}
	if !info.Mode().IsRegular() {
		t.Errorf("Expected regular file, got mode %v", info.Mode())
	}

	// Verify content
	content, err := os.ReadFile(copiedLink)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "now a regular file" {
		t.Errorf("Content = %q, want %q", string(content), "now a regular file")
	}
}

func TestCopyDirAllDirectoryToSymlink(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a symlink in source
	testFile := filepath.Join(srcDir, "original.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	linkPath := filepath.Join(srcDir, "item")
	if err := os.Symlink("original.txt", linkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create a directory at dst/item with files inside (simulating user modification)
	dstItem := filepath.Join(dstDir, "item")
	if err := os.Mkdir(dstItem, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstItem, "userfile.txt"), []byte("user data"), 0644); err != nil {
		t.Fatalf("Failed to create user file: %v", err)
	}

	// Copy should fail because we refuse to replace directories
	err := CopyDirAll(srcDir, dstDir)
	if err == nil {
		t.Fatal("Expected error when trying to replace directory with symlink, but got nil")
	}

	// Verify the directory still exists with its contents (no data loss)
	if _, err := os.Stat(filepath.Join(dstItem, "userfile.txt")); err != nil {
		t.Errorf("User file should still exist after failed copy: %v", err)
	}
}

func TestCopyDirAllDirectoryToFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a regular file in source
	srcFile := filepath.Join(srcDir, "item")
	if err := os.WriteFile(srcFile, []byte("file content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create a directory at dst/item with files inside (simulating user modification)
	dstItem := filepath.Join(dstDir, "item")
	if err := os.Mkdir(dstItem, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstItem, "userfile.txt"), []byte("user data"), 0644); err != nil {
		t.Fatalf("Failed to create user file: %v", err)
	}

	// Copy should fail because we refuse to replace directories
	err := CopyDirAll(srcDir, dstDir)
	if err == nil {
		t.Fatal("Expected error when trying to replace directory with file, but got nil")
	}

	// Verify the directory still exists with its contents (no data loss)
	if _, err := os.Stat(filepath.Join(dstItem, "userfile.txt")); err != nil {
		t.Errorf("User file should still exist after failed copy: %v", err)
	}
}

func TestFindPaths(t *testing.T) {
	dir := t.TempDir()

	// Create test structure
	file1 := filepath.Join(dir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	subDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Find paths
	paths, err := FindPaths(dir)
	if err != nil {
		t.Fatalf("FindPaths() failed: %v", err)
	}

	// Verify we found the expected paths (including the root ".")
	if len(paths) < 3 {
		t.Errorf("FindPaths() found %d paths, want at least 3", len(paths))
	}
}

func TestRemoveAll(t *testing.T) {
	t.Run("remove file", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		if err := RemoveAll(file); err != nil {
			t.Fatalf("RemoveAll() failed: %v", err)
		}

		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Error("File still exists after RemoveAll()")
		}
	})

	t.Run("remove directory", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		if err := RemoveAll(subDir); err != nil {
			t.Fatalf("RemoveAll() failed: %v", err)
		}

		if _, err := os.Stat(subDir); !os.IsNotExist(err) {
			t.Error("Directory still exists after RemoveAll()")
		}
	})
}

func TestPathExists(t *testing.T) {
	dir := t.TempDir()
	existingFile := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: existingFile,
			want: true,
		},
		{
			name: "existing directory",
			path: dir,
			want: true,
		},
		{
			name: "non-existing path",
			path: filepath.Join(dir, "does-not-exist.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathExists(tt.path)
			if got != tt.want {
				t.Errorf("PathExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "directory",
			path:    dir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "file",
			path:    file,
			want:    false,
			wantErr: false,
		},
		{
			name:    "non-existing",
			path:    filepath.Join(dir, "does-not-exist"),
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestReadWriteFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	content := "test content\nwith newline"

	// Test write
	if err := WriteFile(file, content); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Test read
	got, err := ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if got != content {
		t.Errorf("ReadFile() = %q, want %q", got, content)
	}
}

func TestReadFile_Error(t *testing.T) {
	// Test reading non-existent file
	_, err := ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFile() expected error for non-existent file, got nil")
	}
}

func TestWriteFile_Error(t *testing.T) {
	// Test writing to invalid path (directory doesn't exist)
	err := WriteFile("/nonexistent/dir/file.txt", "content")
	if err == nil {
		t.Error("WriteFile() expected error for invalid path, got nil")
	}
}
