package clang

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ejfitzgerald/clang-tidy-cache/utils"
)

type DatabaseEntry struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

type Database = []DatabaseEntry

func ExtractCompilationTarget(databaseRootPath string, target string) (*DatabaseEntry, error) {
	compilationDbPath, err := utils.FindInParents(databaseRootPath, "compile_commands.json")
	if err != nil {
		return nil, err
	}

	jsonFile, err := os.Open(compilationDbPath)
	if err != nil {
		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var db Database
	err = json.Unmarshal(bytes, &db)
	if err != nil {
		return nil, err
	}

	for _, entry := range db {
		entry.File = utils.PosixifyPath(entry.File)
		entry.Directory = utils.PosixifyPath(entry.Directory)
		target = utils.PosixifyPath(target)

		if entry.File == target || entry.File == filepath.Join(entry.Directory, target) {
			return &entry, nil
		}
	}

	return nil, errors.New("Unable to locate the compiler definition")
}
