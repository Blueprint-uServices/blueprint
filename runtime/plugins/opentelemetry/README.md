<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# opentelemetry

```go
import "github.com/blueprint-uservices/blueprint/runtime/plugins/opentelemetry"
```

## Index

- [type OTTraceLogger](<#OTTraceLogger>)
  - [func NewOTTraceLogger\(ctx context.Context\) \(\*OTTraceLogger, error\)](<#NewOTTraceLogger>)
  - [func \(l \*OTTraceLogger\) Debug\(ctx context.Context, format string, args ...any\) \(context.Context, error\)](<#OTTraceLogger.Debug>)
  - [func \(l \*OTTraceLogger\) Error\(ctx context.Context, format string, args ...any\) \(context.Context, error\)](<#OTTraceLogger.Error>)
  - [func \(l \*OTTraceLogger\) Info\(ctx context.Context, format string, args ...any\) \(context.Context, error\)](<#OTTraceLogger.Info>)
  - [func \(l \*OTTraceLogger\) Logf\(ctx context.Context, opts backend.LogOptions, format string, args ...any\) \(context.Context, error\)](<#OTTraceLogger.Logf>)
  - [func \(l \*OTTraceLogger\) Warn\(ctx context.Context, format string, args ...any\) \(context.Context, error\)](<#OTTraceLogger.Warn>)
- [type StdoutMetricCollector](<#StdoutMetricCollector>)
  - [func NewStdoutMetricCollector\(ctx context.Context\) \(\*StdoutMetricCollector, error\)](<#NewStdoutMetricCollector>)
  - [func \(s \*StdoutMetricCollector\) GetMetricProvider\(ctx context.Context\) \(metric.MeterProvider, error\)](<#StdoutMetricCollector.GetMetricProvider>)
- [type StdoutTracer](<#StdoutTracer>)
  - [func NewStdoutTracer\(ctx context.Context, addr string\) \(\*StdoutTracer, error\)](<#NewStdoutTracer>)
  - [func \(t \*StdoutTracer\) GetTracerProvider\(ctx context.Context\) \(trace.TracerProvider, error\)](<#StdoutTracer.GetTracerProvider>)


<a name="OTTraceLogger"></a>
## type [OTTraceLogger](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L16-L18>)

Implementation of the \[backend.Logger\] interface for backend.Tracer This logger converts each log statement into an event which is added to a current span. Note: This logger should only be used in conjunction with a backend.Tracer. Using this logger without using a backend.Tracer would result in no\-op logging behavior. Note: This implementation will not be the same as a future OpenTelemetry.Logger which is in beta\-testing for select languages \(not including Go\).

```go
type OTTraceLogger struct {
    backend.Logger
}
```

<a name="NewOTTraceLogger"></a>
### func [NewOTTraceLogger](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L21>)

```go
func NewOTTraceLogger(ctx context.Context) (*OTTraceLogger, error)
```

Returns a new OTTraceLogger object

<a name="OTTraceLogger.Debug"></a>
### func \(\*OTTraceLogger\) [Debug](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L28>)

```go
func (l *OTTraceLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error)
```

Implements backend.Logger

<a name="OTTraceLogger.Error"></a>
### func \(\*OTTraceLogger\) [Error](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L61>)

```go
func (l *OTTraceLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error)
```

Implements backend.Logger

<a name="OTTraceLogger.Info"></a>
### func \(\*OTTraceLogger\) [Info](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L39>)

```go
func (l *OTTraceLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error)
```

Implements backend.Logger

<a name="OTTraceLogger.Logf"></a>
### func \(\*OTTraceLogger\) [Logf](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L72>)

```go
func (l *OTTraceLogger) Logf(ctx context.Context, opts backend.LogOptions, format string, args ...any) (context.Context, error)
```

Implements backend.Logger

<a name="OTTraceLogger.Warn"></a>
### func \(\*OTTraceLogger\) [Warn](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/log.go#L50>)

```go
func (l *OTTraceLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error)
```

Implements backend.Logger

<a name="StdoutMetricCollector"></a>
## type [StdoutMetricCollector](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/metric.go#L14-L16>)



```go
type StdoutMetricCollector struct {
    // contains filtered or unexported fields
}
```

<a name="NewStdoutMetricCollector"></a>
### func [NewStdoutMetricCollector](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/metric.go#L22>)

```go
func NewStdoutMetricCollector(ctx context.Context) (*StdoutMetricCollector, error)
```



<a name="StdoutMetricCollector.GetMetricProvider"></a>
### func \(\*StdoutMetricCollector\) [GetMetricProvider](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/metric.go#L18>)

```go
func (s *StdoutMetricCollector) GetMetricProvider(ctx context.Context) (metric.MeterProvider, error)
```



<a name="StdoutTracer"></a>
## type [StdoutTracer](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/trace.go#L11-L13>)



```go
type StdoutTracer struct {
    // contains filtered or unexported fields
}
```

<a name="NewStdoutTracer"></a>
### func [NewStdoutTracer](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/trace.go#L15>)

```go
func NewStdoutTracer(ctx context.Context, addr string) (*StdoutTracer, error)
```



<a name="StdoutTracer.GetTracerProvider"></a>
### func \(\*StdoutTracer\) [GetTracerProvider](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/opentelemetry/trace.go#L28>)

```go
func (t *StdoutTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
