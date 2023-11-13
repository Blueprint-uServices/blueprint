---
title: runtime/plugins/jaeger
---
# runtime/plugins/jaeger
```go
package jaeger // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/jaeger"
```

## TYPES

```go
type JaegerTracer struct {
	// Has unexported fields.
}
```
## func NewJaegerTracer
```go
func NewJaegerTracer(ctx context.Context, addr string) (*JaegerTracer, error)
```

## func 
```go
func (t *JaegerTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
```


