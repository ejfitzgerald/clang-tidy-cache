package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindInParents searches for a file named filename in searchDir or any of it's
// parents.
// Returns the full path to the file when found, starting the search in
// searchDir and moving up to the parent directories.
func FindInParents(searchDir string, filename string) (string, error) {
	currentPath := filepath.Join(searchDir, filename)
	if _, err := os.Stat(currentPath); err == nil {
		return currentPath, nil
	}
	// File does not exists, try any parent directories recursively
	parentDir := filepath.Dir(searchDir)
	if parentDir == searchDir {
		return "", fmt.Errorf("Failed to find %v in %v or any of the parent directories", filename, searchDir)
	}
	return FindInParents(parentDir, filename)
}

// Converts given path to Posix (replacing \ with /)
//
// @param {string} givenPath Path to convert
//
// @returns {string} Converted filepath
//
func PosixifyPath(givenPath string) string {
	return strings.ReplaceAll(givenPath, "\\", "/")
}
