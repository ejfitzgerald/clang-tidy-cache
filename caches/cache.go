package caches

import (
	"crypto/sha256"
	"fmt"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"io"
	"os"
	"path/filepath"
)

type Cacher interface {
	// Find contents of cache entry specified by digest.
	FindEntry(digest []byte) ([]byte, error)
	// Store contents into a cache entry specified by digest.
	SaveEntry(digest []byte, content []byte) error
}

func findInParents(searchDir string, filename string) (string, error) {
	currentPath := filepath.Join(searchDir, filename)
	if _, err := os.Stat(currentPath); err == nil {
		return currentPath, nil
	}
	// File does not exists, try any parent directories recursively
	parentDir := filepath.Dir(searchDir)
	if parentDir == searchDir {
		return "", fmt.Errorf("Failed to find %v in %v or any of the parent directories", filename, searchDir)
	}
	return findInParents(parentDir, filename)
}

func computeDigestForConfigFile(projectRoot string) ([]byte, error) {
	configFilePath, err := findInParents(projectRoot, ".clang-tidy")
	if err != nil {
		return nil, err
	}

	// compute the SHA of the configuration file
	// read the contents of the file am hash it
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}

	// compute the final digest
	digest := hasher.Sum(nil)

	return digest, nil
}

func ComputeFingerPrint(invocation *clang.TidyInvocation, wd string, args []string) ([]byte, error) {

	// extract the compilation target command flags from the database
	targetFlags, err := clang.ExtractCompilationTarget(invocation.DatabaseRoot, invocation.TargetPath)
	if err != nil {
		return nil, err
	}

	// parse the main clang flags
	compileCommand, err := clang.ParseClangCommandString(targetFlags.Command)
	if err != nil {
		return nil, err
	}

	// main part of the fingerprint check generate the preprocessed output file and create a SHA256 of it
	preProcessedDigest, err := clang.EvaluatePreprocessedFile(targetFlags.Directory, compileCommand)
	if err != nil {
		return nil, err
	}

	// generate a digest for the full configuration
	configDigest, err := computeDigestForConfigFile(wd)
	if err != nil {
		return nil, err
	}

	// combine all the digests to generate a unique fingerprint
	hasher := sha256.New()
	hasher.Write(preProcessedDigest)
	hasher.Write(configDigest)
	fingerPrint := hasher.Sum(nil)

	return fingerPrint, nil
}
