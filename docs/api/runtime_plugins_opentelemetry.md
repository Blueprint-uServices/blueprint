---
title: runtime/plugins/opentelemetry
---
# runtime/plugins/opentelemetry
```go
package opentelemetry // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/opentelemetry"
```

## TYPES

```go
type StdoutTracer struct {
	// Has unexported fields.
}
```
## func NewStdoutTracer
```go
func NewStdoutTracer(ctx context.Context, addr string) (*StdoutTracer, error)
```

## func 
```go
func (t *StdoutTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
```


