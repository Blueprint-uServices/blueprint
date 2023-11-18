---
title: runtime/plugins/memcached
---
# runtime/plugins/memcached
```go
package memcached // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/memcached"
```

## TYPES

```go
type Memcached struct {
	backend.Cache
	Client *memcache.Client
}
```
## func NewMemcachedClient
```go
func NewMemcachedClient(ctx context.Context, serverAddress string) (*Memcached, error)
```

## func 
```go
func (m *Memcached) Delete(ctx context.Context, key string) error
```

## func 
```go
func (m *Memcached) Get(ctx context.Context, key string, value interface{}) error
```

## func 
```go
func (m *Memcached) Incr(ctx context.Context, key string) (int64, error)
```

## func 
```go
func (m *Memcached) Mget(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (m *Memcached) Mset(ctx context.Context, keys []string, values []interface{}) error
```

## func 
```go
func (m *Memcached) Put(ctx context.Context, key string, value interface{}) error
```


