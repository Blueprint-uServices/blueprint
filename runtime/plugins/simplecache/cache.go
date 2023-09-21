package simplecache

import (
	"context"
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

/* A simple map-based cache that implements the cache interface */
type SimpleCache struct {
	backend.Cache
	values map[string]any
}

func NewSimpleCache() (*SimpleCache, error) {
	cache := &SimpleCache{}
	cache.values = make(map[string]any)
	return cache, nil
}

func (cache *SimpleCache) Put(ctx context.Context, key string, value interface{}) error {
	cache.values[key] = value
	return nil
}

func (cache *SimpleCache) Get(ctx context.Context, key string, val interface{}) error {
	receiver_ptr := reflect.ValueOf(val)
	if receiver_ptr.Kind() != reflect.Pointer || receiver_ptr.IsNil() {
		return fmt.Errorf("invalid cache Get type %v", reflect.TypeOf(val))
	}
	receiver_value := reflect.Indirect(receiver_ptr)

	v, exists := cache.values[key]
	if !exists {
		receiver_value.SetZero()
		return nil
	}

	actual_value := reflect.ValueOf(v)
	if !actual_value.Type().AssignableTo(receiver_value.Type()) {
		return fmt.Errorf("cache Get %v received incompatible types %v and %v", key, actual_value.Type(), receiver_value.Type())
	}
	receiver_value.Set(actual_value)
	return nil
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
