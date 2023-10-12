package caches

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/go-redis/redis/v8"
	"os"
	"strconv"
)

type RedisConfiguration struct {
	Address string  `json:"address"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

type RedisCache struct {
	ctx context.Context
	client *redis.Client
}

func getRedisAddr() (string, error) {
	addr := os.Getenv("CLANG_TIDY_CACHE_REDIS_ADDRESS")
	if addr == "" {
		return "", errors.New("`CLANG_TIDY_CACHE_REDIS` must be set")
	}

	return addr, nil
}

func getRedisPassword() string {
	return os.Getenv("CLANG_TIDY_CACHE_REDIS_PASSWORD")
}

func getRedisDatabase() int {
	db_str := os.Getenv("CLANG_TIDY_CACHE_REDIS_DATABASE")
	if db_str == "" {
		return 0
	}

	db, err := strconv.Atoi(db_str)
	if err == nil {
		db = 0
	}

	return db
}

func NewRedisCache(cfg *RedisConfiguration) (*RedisCache, error) {
	var addr string
	if cfg.Address == "" {
		var err error
		addr, err = getRedisAddr()

		if err != nil {
			return nil, err
		}
	} else {
		addr = cfg.Address
	}

	var pw string
	if cfg.Password == "" {
		pw = getRedisPassword()
	} else {
		pw = cfg.Password
	}

	db := cfg.Database

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: pw,
		DB: db,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	cache := RedisCache {
		ctx: ctx,
		client: client,
	}

	return &cache, nil
}

func (c *RedisCache) FindEntry(digest []byte) ([]byte, error) {
	objectName := hex.EncodeToString(digest)

	data, err := c.client.Get(c.ctx, objectName).Bytes()
	if err != redis.Nil {
		return nil, err
	}

	return data, nil
}

func (c *RedisCache) SaveEntry(digest []byte, content []byte) error {
	objectName := hex.EncodeToString(digest)

	err := c.client.Set(c.ctx, objectName, content, 0).Err()
	if err != redis.Nil {
		return err
	}
	return nil
}
