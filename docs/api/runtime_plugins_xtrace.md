---
title: runtime/plugins/xtrace
---
# runtime/plugins/xtrace
```go
package xtrace // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/xtrace"
```

## TYPES

```go
type XTracerImpl struct {
	backend.XTracer
}
```
## func NewXTracerImpl
```go
func NewXTracerImpl(ctx context.Context, addr string) (*XTracerImpl, error)
```

## func 
```go
func (xt *XTracerImpl) Get(ctx context.Context) (tracingplane.BaggageContext, error)
```

## func 
```go
func (xt *XTracerImpl) IsTracing(ctx context.Context) (bool, error)
```

## func 
```go
func (xt *XTracerImpl) Log(ctx context.Context, msg string) (context.Context, error)
```

## func 
```go
func (xt *XTracerImpl) LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error)
```

## func 
```go
func (xt *XTracerImpl) Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error)
```

## func 
```go
func (xt *XTracerImpl) Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error)
```

## func 
```go
func (xt *XTracerImpl) StartTask(ctx context.Context, tags ...string) (context.Context, error)
```

## func 
```go
func (xt *XTracerImpl) StopTask(ctx context.Context) (context.Context, error)
```


