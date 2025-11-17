package common

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyDirAll copies a directory recursively from src to dst.
// The target directory will always be user readable & writable.
func CopyDirAll(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return copyEntry(src, path, d, dst)
	})
}

func copyEntry(srcBase, srcPath string, entry fs.DirEntry, dstBase string) error {
	relPath, err := filepath.Rel(srcBase, srcPath)
	if err != nil {
		return err
	}
	target := filepath.Join(dstBase, relPath)

	if entry.IsDir() {
		// Handle directories
		return os.MkdirAll(target, 0755)
	}

	// Ensure parent directory exists
	if parent := filepath.Dir(target); parent != "" {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return err
		}
	}

	// Check if it's a symlink
	info, err := entry.Info()
	if err != nil {
		return err
	}

	if info.Mode()&fs.ModeSymlink != 0 {
		// Handle symlinks as is (preserving relative symlink targets)
		linkTarget, err := os.Readlink(srcPath)
		if err != nil {
			return err
		}
		return os.Symlink(linkTarget, target)
	}

	// Handle regular files
	if err := copyFile(srcPath, target); err != nil {
		return err
	}

	// Because we are copying from the Nix store, the source paths will be read-only.
	// So, make the target writeable by the owner.
	return makeOwnerWriteable(target)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func makeOwnerWriteable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	// Add read/write for owner (0600)
	mode := info.Mode() | 0600
	return os.Chmod(path, mode)
}

// FindPaths recursively finds all paths under a directory.
// Returned list of files or directories are relative to the given directory.
func FindPaths(dir string) ([]string, error) {
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		paths = append(paths, relPath)
		return nil
	})

	return paths, err
}

// RemoveAll recursively deletes the path
func RemoveAll(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// ReadFile reads a file and returns its contents as a string
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(data), nil
}

// WriteFile writes a string to a file
func WriteFile(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}
