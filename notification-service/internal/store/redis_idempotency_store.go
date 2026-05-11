package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(addr string) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (s *RedisIdempotencyStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisIdempotencyStore) IsProcessed(ctx context.Context, key string) (bool, error) {
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (s *RedisIdempotencyStore) MarkProcessed(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Set(ctx, key, "processed", ttl).Err()
}

func (s *RedisIdempotencyStore) Close() error {
	return s.client.Close()
}
