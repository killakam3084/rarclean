package mover

import (
	"fmt"
	"os"
	"path/filepath"
)

// Mover handles safe directory relocation operations
type Mover struct {
	dryRun bool
}

// New creates a new Mover instance
func New(dryRun bool) *Mover {
	return &Mover{
		dryRun: dryRun,
	}
}

// MoveDirectory moves a directory from source to destination
// Creates parent directory if needed
func (m *Mover) MoveDirectory(source, destination string) error {
	// Validate source exists
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source directory does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", source)
	}

	// Validate destination parent exists
	destParent := filepath.Dir(destination)
	if _, err := os.Stat(destParent); err != nil {
		if os.IsNotExist(err) {
			// Create parent directory if it doesn't exist
			if !m.dryRun {
				if err := os.MkdirAll(destParent, 0755); err != nil {
					return fmt.Errorf("failed to create destination parent directory: %w", err)
				}
			}
		} else {
			return fmt.Errorf("failed to check destination parent: %w", err)
		}
	}

	if m.dryRun {
		fmt.Printf("[DRY RUN] Would move: %s -> %s\n", source, destination)
		return nil
	}

	// Perform the move
	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("failed to move directory: %w", err)
	}

	return nil
}

// ValidateMoveOperation checks if a move operation is safe
// Returns true if the destination is writable and source exists
func (m *Mover) ValidateMoveOperation(source, destination string) error {
	// Check source exists
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source does not exist: %w", err)
	}

	// Check destination parent directory
	destParent := filepath.Dir(destination)
	if _, err := os.Stat(destParent); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check destination parent: %w", err)
		}
		// Parent doesn't exist - we can create it
	}

	// Check destination doesn't already exist
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	return nil
}

// CleanupEmptyDirs removes empty directories recursively from rootPath
// Only removes directories that are empty
func (m *Mover) CleanupEmptyDirs(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if !info.IsDir() || path == rootPath {
			return nil
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		if len(entries) == 0 {
			if m.dryRun {
				fmt.Printf("[DRY RUN] Would remove empty directory: %s\n", path)
			} else {
				if err := os.Remove(path); err != nil {
					// Continue on errors, not all directories can be removed
					return nil
				}
			}
		}

		return nil
	})
}
