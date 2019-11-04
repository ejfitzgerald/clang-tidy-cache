package caches

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/hex"
	"io"
	"os"
)

type GcsConfiguration struct {
	BucketId string `json:"bucket_id"`
}

type GoogleCloudStorageCache struct {
	cfg    *GcsConfiguration
	ctx    context.Context
	client *storage.Client
}

func NewGcsCache(cfg *GcsConfiguration) (*GoogleCloudStorageCache, error) {

	// create the context
	ctx := context.Background()

	// create the client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	// create the cache
	cache := &GoogleCloudStorageCache{
		cfg:    cfg,
		ctx:    ctx,
		client: client,
	}

	return cache, nil
}

func (c *GoogleCloudStorageCache) FindEntry(digest []byte, outputFile string) (bool, error) {
	objectName := hex.EncodeToString(digest)

	// attempt to read the entry from the bucket
	source, err := c.client.Bucket(c.cfg.BucketId).Object(objectName).NewReader(c.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		} else {
			return false, err
		}
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

func (c *GoogleCloudStorageCache) SaveEntry(digest []byte, inputFile string) error {
	objectName := hex.EncodeToString(digest)

	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	wc := c.client.Bucket(c.cfg.BucketId).Object(objectName).NewWriter(c.ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}
