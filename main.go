package main

import (
	"encoding/json"
	"fmt"
	"github.com/ejfitzgerald/clang-tidy-cache/caches"
	"github.com/ejfitzgerald/clang-tidy-cache/clang"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
)

const VERSION = "0.3.0"

type Configuration struct {
	ClangTidyPath string                   `json:"clang_tidy_path"`
	BaseDir       string                   `json:"base_dir"`
	GcsConfig     *caches.GcsConfiguration `json:"gcs,omitempty"`
}

func readConfigFile(cfg *Configuration) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	// define the configuration path
	configPath := path.Join(usr.HomeDir, ".ctcache", "config.json")

	// missing config file is fine: we simply use the defaults or env vars
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	// open the configuration file
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read the contents
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		return err
	}

	return nil
}

func readConfigEnv(cfg *Configuration) {
	if envPath := os.Getenv("CLANG_TIDY_CACHE_BINARY"); len(envPath) > 0 {
		cfg.ClangTidyPath = envPath
	}
	if envBaseDir := os.Getenv("CLANG_TIDY_CACHE_BASEDIR"); len(envBaseDir) > 0 {
		cfg.BaseDir = filepath.Clean(envBaseDir)
	}
}

func loadConfiguration() (*Configuration, error) {
	// lowest priority: built-in defaults
	cfg := Configuration{ClangTidyPath: "clang-tidy"}

	// higher priority: config file
	err := readConfigFile(&cfg)
	if err != nil {
		return nil, err
	}

	// highest priority: environment variables
	readConfigEnv(&cfg)

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

	// stream out the output of the command
	go streamOutput(os.Stdout, stdout)
	go streamOutput(os.Stderr, stderr)

	err = cmd.Start()
	if err != nil {
		return err
	}

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
		computedFingerPrint, err := caches.ComputeFingerPrint(cfg.ClangTidyPath, cfg.BaseDir, invocation, wd, args)
		if err != nil {
			return err
		}
		fingerPrint = computedFingerPrint

		// evaluate if this function is has already been completed
		cacheContent, err := cache.FindEntry(fingerPrint)
		if err != nil {
			return err
		}
		if invocation.ExportFile != nil {
			f, err := os.Create(*invocation.ExportFile)
			if err != nil {
				return err
			}
			defer f.Close()
			f.Write(cacheContent)
		}

		// this is "hopefully" the general case where we get a cache hit and this means that we need to do nothing
		// further
		if cacheContent != nil {
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
		content := []byte{}
		if invocation.ExportFile != nil {
			content, err = ioutil.ReadFile(*invocation.ExportFile)
			if err != nil {
				return err
			}
		}
		err = cache.SaveEntry(fingerPrint, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// we are only interested in the arguments for the command
	args := os.Args[1:]

	// handle version
	if len(args) == 1 && args[0] == "version" {
		fmt.Printf("clang-tidy-cache %s\n", VERSION)
		os.Exit(1)
	}

	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// find the working directory
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	// attempt to load the Google Cloud cache
	var cache caches.Cacher
	if cfg.GcsConfig != nil {
		candidate, err := caches.NewGcsCache(cfg.GcsConfig)
		if err == nil {
			cache = candidate
		}
	}

	// if no other cache is configured then default to the FS cache
	if cache == nil {
		cache = caches.NewFsCache()
	}

	// evaluate the clang tidy command
	err = evaluateTidyCommand(cfg, wd, args, cache)
	if err != nil {
		fmt.Printf("Failed to get commands: %v\n", err)
		os.Exit(1)
	}
}
