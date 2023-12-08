package backend

import "context"

// Represents a key-value cache.
type Cache interface {
	// Store a key-value pair in the cache
	Put(ctx context.Context, key string, value interface{}) error

	// Retrieves a value from the cache.
	// val should be a pointer in which the value will be stored, e.g.
	//
	//   var value interface{}
	//   cache.Get(ctx, "key", &value)
	//
	// Reports whether the key existed in the cache
	Get(ctx context.Context, key string, val interface{}) (bool, error)

	// Store multiple key-value pairs in the cache.
	// keys and values must have the same length or an error will be returned
	Mset(ctx context.Context, keys []string, values []interface{}) error

	// Retrieve the values for multiple keys from the cache.
	// keys and values must have the same length or an error will be returned
	// values is an array of pointers to which the values will be stored, e.g.
	//
	//   var a string
	//   var b int64
	//   cache.Mget(ctx, []string{"a", "b"}, []any{&a, &b})
	Mget(ctx context.Context, keys []string, values []interface{}) error

	// Delete from the cache
	Delete(ctx context.Context, key string) error

	// Treats the value mapped to key as an integer, and increments it
	Incr(ctx context.Context, key string) (int64, error)
}
