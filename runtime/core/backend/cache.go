package backend

import "context"

type Cache interface {
	Put(ctx context.Context, key string, value interface{}) error
	// val is the pointer to which the value will be stored
	Get(ctx context.Context, key string, val interface{}) (bool, error)
	Mset(ctx context.Context, keys []string, values []interface{}) error
	// values is the array of pointers to which the value will be stored
	Mget(ctx context.Context, keys []string, values []interface{}) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
}
