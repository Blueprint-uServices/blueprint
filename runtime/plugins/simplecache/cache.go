package simplecache

import (
	"context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

/* A simple map-based cache that implements the cache interface */
type SimpleCache struct {
	backend.Cache
	values map[string]any
}

func NewSimpleCache(ctx context.Context) (*SimpleCache, error) {
	cache := &SimpleCache{}
	cache.values = make(map[string]any)
	return cache, nil
}

func (cache *SimpleCache) Put(ctx context.Context, key string, value interface{}) error {
	cache.values[key] = value
	return nil
}

func (cache *SimpleCache) Get(ctx context.Context, key string, val interface{}) error {
	if v, exists := cache.values[key]; exists {
		return backend.CopyResult(v, val)
	} else {
		return backend.SetZero(val)
	}
}

func (cache *SimpleCache) Mset(ctx context.Context, keys []string, values []interface{}) error {
	if len(keys) != len(values) {
		return fmt.Errorf("mset received %v keys but %v values", len(keys), len(values))
	}

	for i, key := range keys {
		err := cache.Put(ctx, key, values[i])
		if err != nil {
			return err
		}
	}

	return nil
}
func (cache *SimpleCache) Mget(ctx context.Context, keys []string, values []interface{}) error {
	if len(keys) != len(values) {
		return fmt.Errorf("mget received %v keys but %v values", len(keys), len(values))
	}

	for i, key := range keys {
		err := cache.Get(ctx, key, values[i])
		if err != nil {
			return err
		}
	}

	return nil
}
func (cache *SimpleCache) Delete(ctx context.Context, key string) error {
	delete(cache.values, key)
	return nil
}

func (cache *SimpleCache) Incr(ctx context.Context, key string) (int64, error) {
	cur := int64(0)
	err := cache.Get(ctx, key, &cur)
	if err != nil {
		return cur, err
	}
	cur += 1
	return cur, cache.Put(ctx, key, cur)
}
