package caches

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"io"
	"os"
	"path"
)

type Cacher interface {
	FindEntry(digest []byte, outputFile string) (bool, error)
	SaveEntry(digest []byte, inputFile string) error
}

func computeDigestForConfigFile(projectRoot string) ([]byte, error) {
	configFilePath := path.Join(projectRoot, ".clang-tidy")

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
	fmt.Println("WD: ", wd)

	// extract the compilation target command flags from the database
	targetFlags, err := clang.ExtractCompilationTarget(invocation.DatabaseRoot, invocation.TargetPath)
	if err != nil {
		fmt.Println("FINGERPRINT: Failed to extract compilation target")
		return nil, err
	}

	// parse the main clang flags
	compileCommand, err := clang.ParseClangCommandString(targetFlags.Command)
	if err != nil {
		fmt.Println("FINGERPRINT: Failed to parse compiler commands")
		return nil, err
	}

	fmt.Println("BuildROOT: ", invocation.DatabaseRoot)

	// main part of the fingerprint check generate the preprocessed output file and create a SHA256 of it
	preProcessedDigest, err := clang.EvaluatePreprocessedFile(targetFlags.Directory, compileCommand)
	if err != nil {
		fmt.Println("FINGERPRINT: Failed to evaluate pre-processor output")
		return nil, err
	}

	// generate a digest for the full configuration
	configDigest, err := computeDigestForConfigFile(wd)
	if err != nil {
		fmt.Println("FINGERPRINT: Failed to generate digest for configuration file")
		return nil, err
	}

	// combine all the digests to generate a unique fingerprint
	hasher := sha256.New()
	hasher.Write(preProcessedDigest)
	hasher.Write(configDigest)
	fingerPrint := hasher.Sum(nil)

	fmt.Println("Digest: ", hex.EncodeToString(fingerPrint))

	return fingerPrint, nil
}
