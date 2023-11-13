---
title: plugins/opentelemetry
---
# plugins/opentelemetry
```go
package opentelemetry // import "gitlab.mpi-sws.org/cld/blueprint/plugins/opentelemetry"
```

## VARIABLES

```go
var DefaultOpenTelemetryCollectorName = "ot_collector"
```
## FUNCTIONS

## func DefineOpenTelemetryCollector
```go
func DefineOpenTelemetryCollector(spec wiring.WiringSpec, collectorName string) string
```
Defines the OpenTelemetry collector as a process node

# Also creates a pointer to the collector and a client node that are used by
OT clients

This doesn't need to be explicitly called, although it can if users want to
control the placement of the opentelemetry collector

## func Instrument
```go
func Instrument(spec wiring.WiringSpec, serviceName string)
```
Instruments `serviceName` with OpenTelemetry. This can only be done if
`serviceName` is a pointer from Golang nodes to Golang nodes.

This call will also define the OpenTelemetry collector.

Instrumenting `serviceName` will add both src and dst-side modifiers to the
pointer.

## func InstrumentUsingCustomCollector
```go
func InstrumentUsingCustomCollector(spec wiring.WiringSpec, serviceName string, collectorName string)
```
This is the same as the Instrument function, but uses `collectorName` as
the OpenTelemetry collector and does not attempt to define or redefine the
collector.


## TYPES

```go
type OTCollectorInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (xt *OTCollectorInterface) GetMethods() []service.Method
```

## func 
```go
func (xt *OTCollectorInterface) GetName() string
```

```go
type OpenTelemetryClientWrapper struct {
	golang.Service
	golang.GeneratesFuncs
```
```go
	WrapperName string
```
```go
	Wrapped   golang.Service
	Collector OpenTelemetryCollectorInterface
	// Has unexported fields.
}
```
## func 
```go
func (node *OpenTelemetryClientWrapper) AddInstantiation(builder golang.GraphBuilder) error
```
Part of code generation compilation pass; provides instantiation snippet

## func 
```go
func (node *OpenTelemetryClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error
```
Part of code generation compilation pass; creates the interface definition
code for the wrapper, and any new generated structs that are exposed and can
be used by other IRNodes

## func 
```go
func (node *OpenTelemetryClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error
```
Part of code generation compilation pass; provides implementation of
interfaces from GenerateInterfaces

## func 
```go
func (node *OpenTelemetryClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *OpenTelemetryClientWrapper) ImplementsGolangNode()
```

## func 
```go
func (node *OpenTelemetryClientWrapper) ImplementsGolangService()
```

## func 
```go
func (node *OpenTelemetryClientWrapper) Name() string
```

## func 
```go
func (node *OpenTelemetryClientWrapper) String() string
```

```go
type OpenTelemetryCollector struct {
	docker.Container
```
```go
	CollectorName string
	BindAddr      *address.BindConfig
	Iface         *goparser.ParsedInterface
}
```
## func 
```go
func (node *OpenTelemetryCollector) AddContainerArtifacts(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *OpenTelemetryCollector) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *OpenTelemetryCollector) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *OpenTelemetryCollector) Name() string
```

## func 
```go
func (node *OpenTelemetryCollector) String() string
```

```go
type OpenTelemetryCollectorClient struct {
	golang.Node
	golang.Instantiable
```
```go
	ClientName string
	ServerDial *address.DialConfig
```
```go
	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}
```
## func 
```go
func (node *OpenTelemetryCollectorClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *OpenTelemetryCollectorClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *OpenTelemetryCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *OpenTelemetryCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *OpenTelemetryCollectorClient) ImplementsGolangNode()
```

## func 
```go
func (node *OpenTelemetryCollectorClient) ImplementsOTCollectorClient()
```

## func 
```go
func (node *OpenTelemetryCollectorClient) Name() string
```

## func 
```go
func (node *OpenTelemetryCollectorClient) String() string
```

```go
type OpenTelemetryCollectorInterface interface {
	golang.Node
	golang.Instantiable
	ImplementsOTCollectorClient()
}
```
```go
type OpenTelemetryServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs
```
```go
	WrapperName string
```
```go
	Wrapped   golang.Service
	Collector OpenTelemetryCollectorInterface
	// Has unexported fields.
}
```
## func 
```go
func (node *OpenTelemetryServerWrapper) AddInstantiation(builder golang.GraphBuilder) error
```
Part of code generation compilation pass; provides instantiation snippet

## func 
```go
func (node *OpenTelemetryServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error
```
Part of code generation compilation pass; creates the interface definition
code for the wrapper, and any new generated structs that are exposed and can
be used by other IRNodes

## func 
```go
func (node *OpenTelemetryServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error
```
Part of code generation compilation pass; provides implementation of
interfaces from GenerateInterfaces

## func 
```go
func (node *OpenTelemetryServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *OpenTelemetryServerWrapper) ImplementsGolangNode()
```

## func 
```go
func (node *OpenTelemetryServerWrapper) ImplementsGolangService()
```

## func 
```go
func (node *OpenTelemetryServerWrapper) Name() string
```

## func 
```go
func (node *OpenTelemetryServerWrapper) String() string
```


