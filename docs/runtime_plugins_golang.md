---
title: runtime/plugins/golang
---
# runtime/plugins/golang
```go
package golang // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/golang"
```

## TYPES

```go
type BuildFunc func(ctr Container) (any, error)
```
```go
type Container interface {
	Get(name string, receiver any) error
	Context() context.Context // In case the buildfunc wants to start background goroutines
	CancelFunc() context.CancelFunc
	WaitGroup() *sync.WaitGroup // Waitgroup used by this container; plugins can call Add if they create goroutines
}
```
```go
type Graph interface {
	Define(name string, build BuildFunc) error
	Build() Container
}
```
## func NewGraph
```go
func NewGraph(ctx context.Context, cancel context.CancelFunc, parent Container, name string) Graph
```

For nodes that want to run background goroutines
```go
type Runnable interface {
	Run(ctx context.Context) error
}
```

