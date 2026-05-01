package extractor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Extractor handles RAR file discovery and extraction
type Extractor struct {
	directory string
}

// RARFile represents a RAR archive
type RARFile struct {
	Path      string // Full path to first part
	Directory string // Directory containing the RAR files
	BaseName  string // Base name of the RAR (without extensions)
}

// New creates a new Extractor for the given directory
func New(directory string) *Extractor {
	return &Extractor{
		directory: directory,
	}
}

// FindRARFiles finds all RAR archives in the directory
// Returns a list of RARFile structures representing unique archives
func (e *Extractor) FindRARFiles() ([]RARFile, error) {
	var rarFiles []RARFile
	seen := make(map[string]bool)

	err := filepath.Walk(e.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is a RAR archive
		if !isRARFile(path) {
			return nil
		}

		// Get the base name without part suffixes
		baseName := getRARBaseName(path)

		// Skip if we've already processed this archive
		if seen[baseName] {
			return nil
		}

		// Find the first part of the archive
		firstPart := findFirstPart(filepath.Dir(path), baseName)
		if firstPart == "" {
			// If no first part found, use the current file
			firstPart = path
		}

		seen[baseName] = true
		rarFiles = append(rarFiles, RARFile{
			Path:      firstPart,
			Directory: filepath.Dir(firstPart),
			BaseName:  baseName,
		})

		return nil
	})

	return rarFiles, err
}

// Extract extracts a RAR file to its directory
func (e *Extractor) Extract(rar RARFile) error {
	// Verify 7z is available
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z command not found: %w", err)
	}

	// Run extraction with 7z
	// -y: assume yes to all prompts
	// -o: output directory (same as RAR location)
	cmd := exec.Command("7z", "x", "-y", fmt.Sprintf("-o%s", rar.Directory), rar.Path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract %s: %w", rar.Path, err)
	}

	return nil
}

// isRARFile checks if a file has a RAR extension
func isRARFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".rar" {
		return true
	}

	// Check for .partXXX.rar pattern
	base := filepath.Base(path)
	if strings.Contains(base, ".part") && strings.HasSuffix(strings.ToLower(base), ".rar") {
		return true
	}

	return false
}

// getRARBaseName extracts the base name from a RAR file path
// Handles: archive.rar, archive.part01.rar, archive.part001.rar
func getRARBaseName(path string) string {
	base := filepath.Base(path)
	base = strings.ToLower(base)

	// Remove .rar extension
	base = strings.TrimSuffix(base, ".rar")

	// Remove .partXXX suffix
	if idx := strings.LastIndex(base, ".part"); idx != -1 {
		return base[:idx]
	}

	return base
}

// findFirstPart finds the first part of a multi-part RAR archive
// Looks for .rar, .part01.rar, .part001.rar in order
func findFirstPart(directory, baseName string) string {
	candidates := []string{
		fmt.Sprintf("%s.rar", baseName),
		fmt.Sprintf("%s.part01.rar", baseName),
		fmt.Sprintf("%s.part001.rar", baseName),
	}

	for _, candidate := range candidates {
		fullPath := filepath.Join(directory, candidate)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}
