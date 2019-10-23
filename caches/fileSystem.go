package caches

import (
	"encoding/hex"
	"io"
	"os"
	"os/user"
	"path"
)

type FileSystemCache struct {
	root string
}

func NewFsCache() *FileSystemCache {
	usr, _ := user.Current()
	return &FileSystemCache{
		root: path.Join(usr.HomeDir, ".ctcache", "cache"),
	}
}

func (c *FileSystemCache) FindEntry(digest []byte, outputFile string) (bool, error) {
	_, entryPath := defineEntryPath(c.root, digest)
	_, err := os.Stat(entryPath)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	source, err := os.Open(entryPath)
	if err != nil {
		return false, err
	}
	defer source.Close()

	destination, err := os.Create(outputFile)
	if err != nil {
		return false, err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *FileSystemCache) SaveEntry(digest []byte, inputFile string) error {
	entryRoot, entryPath := defineEntryPath(c.root, digest)

	err := os.MkdirAll(entryRoot, 0755)
	if err != nil {
		return err
	}

	source, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(entryPath)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return nil
}

func defineEntryPath(root string, digest []byte) (string, string) {
	encodedDigest := hex.EncodeToString(digest)
	entryRoot := path.Join(root, encodedDigest[0:2], encodedDigest[2:4])
	entryPath := path.Join(entryRoot, encodedDigest[4:])
	return entryRoot, entryPath
}