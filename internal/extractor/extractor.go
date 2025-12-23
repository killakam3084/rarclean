package extractor

import (
	"fmt"
	"log"
)

// Extractor handles RAR archive extraction
type Extractor struct {
	// Configuration fields will be added here
}

// New creates a new Extractor instance
func New() *Extractor {
	return &Extractor{}
}

// Extract extracts a RAR file to the specified destination
func (e *Extractor) Extract(rarPath, destPath string) error {
	log.Printf("Extracting %s to %s\n", rarPath, destPath)
	// TODO: Implement RAR extraction logic
	return fmt.Errorf("not implemented")
}

// IsRARFile checks if a file is a valid RAR archive
func (e *Extractor) IsRARFile(filePath string) bool {
	// TODO: Implement RAR file detection
	return false
}
