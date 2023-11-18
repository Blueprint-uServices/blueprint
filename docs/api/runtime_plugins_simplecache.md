---
title: runtime/plugins/simplecache
---
# runtime/plugins/simplecache
```go
package simplecache // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
```

## TYPES

A simple map-based cache that implements the cache interface
```go
type SimpleCache struct {
	backend.Cache
	// Has unexported fields.
}
```
## func NewSimpleCache
```go
func NewSimpleCache(ctx context.Context) (*SimpleCache, error)
```

## func 
```go
func (cache *SimpleCache) Delete(ctx context.Context, key string) error
```

## func 
```go
func (cache *SimpleCache) Get(ctx context.Context, key string, val interface{}) error
```

## func 
```go
func (cache *SimpleCache) Incr(ctx context.Context, key string) (int64, error)
```

## func 
```go
func (cache *SimpleCache) Mget(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (cache *SimpleCache) Mset(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (cache *SimpleCache) Put(ctx context.Context, key string, value interface{}) error
```


