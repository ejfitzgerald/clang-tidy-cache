package main

import (
	"encoding/json"
	"github.com/ejfitzgerald/clang-tidy-cache/caches"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
)

type Configuration struct {
	ClangTidyPath string `json:"clang_tidy_path"`
}

func loadConfiguration() (*Configuration, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// define the configuration path
	configPath := path.Join(usr.HomeDir, ".ctcache", "config.json")

	// open the configuration file
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read the contents
	bytes, err := ioutil.ReadAll(jsonFile)

	var cfg Configuration
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func streamOutput(file *os.File, closer io.ReadCloser) {
	defer closer.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := closer.Read(buffer)
		if err != nil {
			break
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			break
		}
	}
}

func runClangTidyCommand(cfg *Configuration, args []string) error {
	cmd := exec.Command(cfg.ClangTidyPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	// stream out the output of the command
	go streamOutput(os.Stdout, stdout)
	go streamOutput(os.Stderr, stderr)

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func shouldBypassCache(args []string) bool {
	for _, arg := range args {
		if arg == "-list-checks" || arg == "--version" {
			return true
		}
	}

	return false
}

func evaluateTidyCommand(cfg *Configuration, wd string, args []string, cache caches.Cacher) error {
	bypassCache := shouldBypassCache(args)

	// fingerprint
	var fingerPrint []byte = nil
	var invocation *clang.TidyInvocation = nil

	if !bypassCache {

		// evaluate the commands that have been provided
		other, err := clang.ParseTidyCommand(args)
		if err != nil {
			return err
		}
		invocation = other

		// compute the finger print for the file
		computedFingerPrint, err := caches.ComputeFingerPrint(invocation, wd, args)
		if err != nil {
			return err
		}
		fingerPrint = computedFingerPrint

		// evaluate if this function is has already been completed
		cacheHit, err := cache.FindEntry(fingerPrint, invocation.ExportFile)
		if err != nil {
			return err
		}

		// this is "hopefully" the general case where we get a cache hit and this means that we need to do nothing
		// further
		if cacheHit {
			return nil
		}
	}

	// we need to run the command
	err := runClangTidyCommand(cfg, args)
	if err != nil {
		return err
	}

	// if the file was clean then we should record this fact into the cache
	if !bypassCache && fingerPrint != nil && invocation != nil {
		err = cache.SaveEntry(fingerPrint, invocation.ExportFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	cfg, err := loadConfiguration()
	if err != nil {
		os.Exit(1)
	}

	// we are only interested in the arguments for the command
	args := os.Args[1:]

	// find the working directory
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	// cache creation
	cache := caches.NewFsCache()

	// evaluate the clang tidy command
	err = evaluateTidyCommand(cfg, wd, args, cache)
	if err != nil {
		os.Exit(1)
	}
}
