---
title: runtime/plugins/zipkin
---
# runtime/plugins/zipkin
```go
package zipkin // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/zipkin"
```

## TYPES

```go
type ZipkinTracer struct {
	// Has unexported fields.
}
```
## func NewZipkinTracer
```go
func NewZipkinTracer(ctx context.Context, addr string) (*ZipkinTracer, error)
```

## func 
```go
func (t *ZipkinTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
```


