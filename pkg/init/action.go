// Package init provides template initialization functionality for Nix projects
package init

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saberzero1/omnix/pkg/common"
)

// Action represents an action to perform on a template
type Action interface {
	// HasValue returns true if this action has a value set
	HasValue() bool
	// Apply applies the action to the given directory
	Apply(ctx context.Context, outDir string) error
	// String returns a string representation of the action
	String() string
}

// ReplaceAction replaces a placeholder with a value
type ReplaceAction struct {
	Placeholder string
	Value       *string
}

// HasValue returns true if the replace action has a value set
func (r ReplaceAction) HasValue() bool {
	return r.Value != nil
}

func (r ReplaceAction) String() string {
	if r.Value == nil {
		return "replace [disabled]"
	}
	return fmt.Sprintf("replace [%s => %s]", r.Placeholder, *r.Value)
}

// Apply performs the replace action on files in the output directory
func (r ReplaceAction) Apply(_ context.Context, outDir string) error {
	if r.Value == nil {
		return nil
	}

	// Find all files in the directory
	files, err := findAllPaths(outDir)
	if err != nil {
		return fmt.Errorf("failed to find paths: %w", err)
	}

	// Sort in reverse order to process files before their parent directories get renamed
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, relPath := range files {
		filePath := filepath.Join(outDir, relPath)

		// Replace in file content
		if info, err := os.Stat(filePath); err == nil && info.Mode().IsRegular() {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			contentStr := string(content)
			if strings.Contains(contentStr, r.Placeholder) {
				fmt.Printf("   ✍️ %s\n", relPath)
				newContent := strings.ReplaceAll(contentStr, r.Placeholder, *r.Value)
				if err := os.WriteFile(filePath, []byte(newContent), info.Mode()); err != nil {
					return fmt.Errorf("failed to write file %s: %w", filePath, err)
				}
			}
		}

		// Rename path if necessary
		fileName := filepath.Base(relPath)
		if strings.Contains(fileName, r.Placeholder) {
			newName := strings.ReplaceAll(fileName, r.Placeholder, *r.Value)
			newPath := filepath.Join(filepath.Dir(filePath), newName)
			if filePath != newPath {
				fmt.Printf("   ✏️ %s => %s\n", relPath, newName)
				if err := os.Rename(filePath, newPath); err != nil {
					return fmt.Errorf("failed to rename %s to %s: %w", filePath, newPath, err)
				}
			}
		}
	}

	return nil
}

// RetainAction deletes paths matching glob patterns if value is false
type RetainAction struct {
	Paths []string // Glob patterns
	Value *bool
}

// HasValue returns true if the retain action has a value set
func (r RetainAction) HasValue() bool {
	return r.Value != nil
}

func (r RetainAction) String() string {
	if r.Value == nil || *r.Value {
		return "prune [disabled]"
	}
	return fmt.Sprintf("prune [%s]", strings.Join(r.Paths, ", "))
}

// Apply performs the retain action by pruning files not matching the retain patterns
func (r RetainAction) Apply(_ context.Context, outDir string) error {
	if r.Value == nil || *r.Value {
		return nil
	}

	// Find all files
	files, err := findAllPaths(outDir)
	if err != nil {
		return fmt.Errorf("failed to find paths: %w", err)
	}

	// Match files against glob patterns
	var filesToDelete []string
	for _, file := range files {
		for _, pattern := range r.Paths {
			if matchGlob(pattern, file) {
				filesToDelete = append(filesToDelete, file)
				break
			}
		}
	}

	if len(filesToDelete) == 0 {
		// No files matched the glob patterns; nothing to delete.
		// This is not an error: RetainAction is designed to work for template features that may not exist in all cases,
		// regardless of whether the patterns are marked as optional.
		return nil
	}

	// Sort in reverse to delete children before parents
	sort.Sort(sort.Reverse(sort.StringSlice(filesToDelete)))

	for _, relPath := range filesToDelete {
		path := filepath.Join(outDir, relPath)
		fmt.Printf("   ❌ %s\n", relPath)
		if err := common.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// matchGlob matches a path against a glob pattern with enhanced ** support
func matchGlob(pattern, path string) bool {
	// Convert backslashes to forward slashes for consistency
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ** patterns (match zero or more directories)
	if strings.Contains(pattern, "**") {
		// Split pattern into parts
		parts := strings.Split(pattern, "**")

		// For simple cases like "dir/**" or "**/file.txt"
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			// Remove leading/trailing slashes
			prefix = strings.TrimSuffix(prefix, "/")
			suffix = strings.TrimPrefix(suffix, "/")

			// Check prefix
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			// Check suffix
			if suffix != "" {
				// Match the suffix as a regular glob
				if suffix == "" {
					return true
				}
				// Match any component ending with suffix
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				if matched {
					return true
				}
				// Also check if the full suffix matches
				if strings.HasSuffix(path, suffix) {
					return true
				}
			} else {
				// Pattern is like "dir/**", match anything under dir
				return true
			}

			return false
		}
	}

	// Fall back to basic glob matching
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	return matched
}

// findAllPaths returns all file and directory paths relative to the root
func findAllPaths(root string) ([]string, error) {
	var paths []string

	err := filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip the root itself
		if relPath == "." {
			return nil
		}

		paths = append(paths, relPath)
		return nil
	})

	return paths, err
}

// ActionPriority returns the priority for sorting actions
// Retain actions should come before Replace actions
func ActionPriority(a Action) int {
	switch a.(type) {
	case RetainAction, *RetainAction:
		return 0
	case ReplaceAction, *ReplaceAction:
		return 1
	case ChmodAction, *ChmodAction:
		return 2
	case MoveAction, *MoveAction:
		return 3
	default:
		return 4
	}
}

// ChmodAction changes file permissions
type ChmodAction struct {
	Paths []string    // Glob patterns for files to change
	Mode  os.FileMode // File permissions mode
	Value *bool       // Enable/disable this action
}

// HasValue returns true if the chmod action has a value set
func (c ChmodAction) HasValue() bool {
	return c.Value != nil && *c.Value
}

func (c ChmodAction) String() string {
	if c.Value == nil || !*c.Value {
		return "chmod [disabled]"
	}
	return fmt.Sprintf("chmod [%s => %o]", strings.Join(c.Paths, ", "), c.Mode)
}

// Apply performs the chmod action on matching files
func (c ChmodAction) Apply(_ context.Context, outDir string) error {
	if c.Value == nil || !*c.Value {
		return nil
	}

	// Find all files
	files, err := findAllPaths(outDir)
	if err != nil {
		return fmt.Errorf("failed to find paths: %w", err)
	}

	// Match files against glob patterns
	var filesToChmod []string
	for _, file := range files {
		for _, pattern := range c.Paths {
			if matchGlob(pattern, file) {
				filesToChmod = append(filesToChmod, file)
				break
			}
		}
	}

	// Change permissions on matched files
	for _, relPath := range filesToChmod {
		path := filepath.Join(outDir, relPath)
		fmt.Printf("   🔒 %s (mode: %o)\n", relPath, c.Mode)
		if err := os.Chmod(path, c.Mode); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", path, err)
		}
	}

	return nil
}

// MoveAction moves or renames files
type MoveAction struct {
	From  string  // Source path pattern
	To    string  // Destination path
	Value *bool   // Enable/disable this action
}

// HasValue returns true if the move action has a value set
func (m MoveAction) HasValue() bool {
	return m.Value != nil && *m.Value
}

func (m MoveAction) String() string {
	if m.Value == nil || !*m.Value {
		return "move [disabled]"
	}
	return fmt.Sprintf("move [%s => %s]", m.From, m.To)
}

// Apply performs the move action
func (m MoveAction) Apply(_ context.Context, outDir string) error {
	if m.Value == nil || !*m.Value {
		return nil
	}

	// Find all files
	files, err := findAllPaths(outDir)
	if err != nil {
		return fmt.Errorf("failed to find paths: %w", err)
	}

	// Match files against the pattern
	var filesToMove []string
	for _, file := range files {
		if matchGlob(m.From, file) {
			filesToMove = append(filesToMove, file)
		}
	}

	// Move each matched file
	for _, relPath := range filesToMove {
		srcPath := filepath.Join(outDir, relPath)
		dstPath := filepath.Join(outDir, m.To)

		// Create destination directory if needed
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
		}

		fmt.Printf("   📦 %s => %s\n", relPath, m.To)
		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", srcPath, dstPath, err)
		}
	}

	return nil
}
