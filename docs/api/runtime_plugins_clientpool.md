---
title: runtime/plugins/clientpool
---
# runtime/plugins/clientpool
```go
package clientpool // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/clientpool"
```

## TYPES

```go
type ClientPool[T any] struct {
	// Has unexported fields.
}
```
## func NewClientPool[T
```go
func NewClientPool[T any](maxClients int64, fn func() (T, error)) *ClientPool[T]
```

## func 
```go
func (this *ClientPool[T]) Pop(ctx context.Context) (client T, err error)
```

## func 
```go
func (this *ClientPool[T]) Push(client T)
```


