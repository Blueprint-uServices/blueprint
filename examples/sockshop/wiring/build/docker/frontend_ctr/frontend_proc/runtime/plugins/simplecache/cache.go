// Package simplecache implements a key-value [backend.Cache] using a golang map.
package simplecache

import (
	"context"
	"fmt"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

// A simple map-based cache that implements the [backend.Cache] interface
type SimpleCache struct {
	backend.Cache
	sync.RWMutex
	values map[string]any
}

// Instantiates a map-based [SimpleCache]
func NewSimpleCache(ctx context.Context) (*SimpleCache, error) {
	cache := &SimpleCache{}
	cache.values = make(map[string]any)
	return cache, nil
}

func (cache *SimpleCache) Put(ctx context.Context, key string, value interface{}) error {
	cache.Lock()
	defer cache.Unlock()
	cache.values[key] = value
	return nil
}

func (cache *SimpleCache) Get(ctx context.Context, key string, val interface{}) (bool, error) {
	if v, exists := cache.values[key]; exists {
		return true, backend.CopyResult(v, val)
	}
	return false, nil
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
		_, err := cache.Get(ctx, key, values[i])
		if err != nil {
			return err
		}
	}

	return nil
}
func (cache *SimpleCache) Delete(ctx context.Context, key string) error {
	cache.Lock()
	defer cache.Unlock()
	delete(cache.values, key)
	return nil
}

func (cache *SimpleCache) Incr(ctx context.Context, key string) (int64, error) {
	cur := int64(0)
	_, err := cache.Get(ctx, key, &cur)
	if err != nil {
		return cur, err
	}
	cur += 1
	return cur, cache.Put(ctx, key, cur)
}
