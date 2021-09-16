package caches

import (
	"crypto/sha256"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"github.com/ejfitzgerald/clang-tidy-cache/utils"
	"io"
	"os"
	"os/exec"
)

type Cacher interface {
	// Find contents of cache entry specified by digest.
	FindEntry(digest []byte) ([]byte, error)
	// Store contents into a cache entry specified by digest.
	SaveEntry(digest []byte, content []byte) error
}

func computeFileDigest(path string) ([]byte, error) {
	f, err := os.Open(path)
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

func computeDigestForConfigFile(projectRoot string) ([]byte, error) {
	configFilePath, err := utils.FindInParents(projectRoot, ".clang-tidy")
	if err != nil {
		return nil, err
	}

	return computeFileDigest(configFilePath)
}

func computeDigestForClangTidyBinary(clangTidyPath string) ([]byte, error) {
	// resolve to a full path: e.g. `clang-tidy` -> `/usr/local/bin/clang-tidy`
	path, err := exec.LookPath(clangTidyPath)
	if err != nil {
		return nil, err
	}

	return computeFileDigest(path)
}

func ComputeFingerPrint(clangTidyPath string, baseDir string, invocation *clang.TidyInvocation,
	wd string, args []string) ([]byte, error) {

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
	preProcessedDigest, err := clang.EvaluatePreprocessedFile(targetFlags.Directory, baseDir, compileCommand)
	if err != nil {
		return nil, err
	}

	// generate a digest for the full configuration
	configDigest, err := computeDigestForConfigFile(wd)
	if err != nil {
		return nil, err
	}

	// we also need to include the clang-tidy binary since different version have different output
	binaryDigest, err := computeDigestForClangTidyBinary(clangTidyPath)
	if err != nil {
		return nil, err
	}

	// combine all the digests to generate a unique fingerprint
	hasher := sha256.New()
	hasher.Write(preProcessedDigest)
	hasher.Write(configDigest)
	hasher.Write(binaryDigest)
	fingerPrint := hasher.Sum(nil)

	return fingerPrint, nil
}
