// Package redis implements a key-value [backend.Cache] client interface to a vanilla redis implementation.
package redis

import (
	"context"
	"encoding/json"

	redis_impl "github.com/go-redis/redis/v8"
)

// A redis client wrapper that implements the [backend.Cache] interface
type RedisCache struct {
	client *redis_impl.Client
}

// Instantiates a new redis client to a memcached instance running at `serverAddress`
func NewRedisCacheClient(ctx context.Context, addr string) (*RedisCache, error) {
	conn_addr := addr
	client := redis_impl.NewClient(&redis_impl.Options{
		Addr:     conn_addr,
		Password: "",
		DB:       0,
	})
	return &RedisCache{client: client}, nil
}

// Implements the backend.Cache interface
func (r *RedisCache) Put(ctx context.Context, key string, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	val_str := string(val)
	return r.client.Set(ctx, key, val_str, 0).Err()
}

// Implements the backend.Cache interface
func (r *RedisCache) Get(ctx context.Context, key string, value interface{}) (bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis_impl.Nil {
		// Key doesn't exist
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(val), value)
}

// Implements the backend.Cache interface
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Implements the backend.Cache interface
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Implements the backend.Cache interface
func (r *RedisCache) Mget(ctx context.Context, keys []string, values []interface{}) error {
	result, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}
	for idx, res := range result {
		err := json.Unmarshal([]byte(res.(string)), values[idx])
		if err != nil {
			return err
		}
	}
	return nil
}

// Implements the backend.Cache interface
func (r *RedisCache) Mset(ctx context.Context, keys []string, values []interface{}) error {
	kv_map := make(map[string]string)
	for idx, key := range keys {
		val, err := json.Marshal(values[idx])
		if err != nil {
			return err
		}
		kv_map[key] = string(val)
	}
	return r.client.MSet(ctx, kv_map).Err()
}
