package caches

import (
	"crypto/sha256"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"github.com/ejfitzgerald/clang-tidy-cache/utils"
	"io"
	"os"
)

type Cacher interface {
	// Find contents of cache entry specified by digest.
	FindEntry(digest []byte) ([]byte, error)
	// Store contents into a cache entry specified by digest.
	SaveEntry(digest []byte, content []byte) error
}

func computeDigestForConfigFile(projectRoot string) ([]byte, error) {
	configFilePath, err := utils.FindInParents(projectRoot, ".clang-tidy")
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
