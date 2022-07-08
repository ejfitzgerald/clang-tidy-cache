package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindInParents searches for a file named filename in searchDir or any of it's
// parents.
// Returns the full path to the file when found, starting the search in
// searchDir and moving up to the parent directories.
func FindInParents(searchDir string, filename string) (string, error) {
	return findInParentsOrig(searchDir, searchDir, filename)
}

func findInParentsOrig(origSearchDir string, searchDir string, filename string) (string, error) {
	currentPath := filepath.Join(searchDir, filename)
	if _, err := os.Stat(currentPath); err == nil {
		return currentPath, nil
	}
	// File does not exists, try any parent directories recursively
	parentDir := filepath.Dir(searchDir)
	if parentDir == searchDir {
		return "", fmt.Errorf("Failed to find %v in %v or any of the parent directories", filename, origSearchDir)
	}
	return findInParentsOrig(origSearchDir, parentDir, filename)
}
