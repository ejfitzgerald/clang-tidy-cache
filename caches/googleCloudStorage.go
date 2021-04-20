package caches

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/hex"
	"io/ioutil"
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

func (c *GoogleCloudStorageCache) FindEntry(digest []byte) ([]byte, error) {
	objectName := hex.EncodeToString(digest)

	// attempt to read the entry from the bucket
	source, err := c.client.Bucket(c.cfg.BucketId).Object(objectName).NewReader(c.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, nil
		} else {
			return nil, err
		}
	}
	defer source.Close()

	return ioutil.ReadAll(source)
}

func (c *GoogleCloudStorageCache) SaveEntry(digest []byte, content []byte) error {
	objectName := hex.EncodeToString(digest)

	wc := c.client.Bucket(c.cfg.BucketId).Object(objectName).NewWriter(c.ctx)
	_, err := wc.Write(content)
	if err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}
