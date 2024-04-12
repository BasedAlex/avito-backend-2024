package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	redis *redis.Client
}

func New(connStr string) (Cache, error) {
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return Cache{}, err
	}
	return Cache{redis.NewClient(opts)}, nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	strcom := c.redis.Get(ctx, key)
	return strcom.Val(), strcom.Err()
}

func (c *Cache) Set(ctx context.Context, key, value string) error {
	return c.redis.Set(ctx, key, value, time.Minute).Err()
}