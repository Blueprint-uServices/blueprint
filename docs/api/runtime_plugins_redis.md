---
title: runtime/plugins/redis
---
# runtime/plugins/redis
```go
package redis // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/redis"
```

## TYPES

```go
type RedisCache struct {
	// Has unexported fields.
}
```
## func NewRedisCacheClient
```go
func NewRedisCacheClient(ctx context.Context, addr string) (*RedisCache, error)
```

## func 
```go
func (r *RedisCache) Delete(ctx context.Context, key string) error
```

## func 
```go
func (r *RedisCache) Get(ctx context.Context, key string, value interface{}) error
```

## func 
```go
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error)
```

## func 
```go
func (r *RedisCache) Mget(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (r *RedisCache) Mset(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (r *RedisCache) Put(ctx context.Context, key string, value interface{}) error
```


