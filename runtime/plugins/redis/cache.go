package redis

import (
	"context"
	"encoding/json"

	redis_impl "github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis_impl.Client
}

func NewRedisCacheClient(ctx context.Context, addr string) (*RedisCache, error) {
	conn_addr := addr
	client := redis_impl.NewClient(&redis_impl.Options{
		Addr:     conn_addr,
		Password: "",
		DB:       0,
	})
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Put(ctx context.Context, key string, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	val_str := string(val)
	return r.client.Set(ctx, key, val_str, 0).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), value)
}

func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

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
