---
title: plugins/xtrace
---
# plugins/xtrace
```go
package xtrace // import "gitlab.mpi-sws.org/cld/blueprint/plugins/xtrace"
```

## FUNCTIONS

## func DefineXTraceServerContainer
```go
func DefineXTraceServerContainer(spec wiring.WiringSpec)
```
## func Instrument
```go
func Instrument(spec wiring.WiringSpec, serviceName string)
```
Instruments the service with an entry + exit point xtrace wrapper to
generate xtrace compatible logs


## TYPES

```go
type XTraceClient struct {
	golang.Node
	golang.Instantiable
```
```go
	ClientName     string
	ServerDialAddr *address.DialConfig
```
```go
	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}
```
## func 
```go
func (node *XTraceClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *XTraceClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *XTraceClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *XTraceClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *XTraceClient) ImplementsGolangNode()
```

## func 
```go
func (node *XTraceClient) Name() string
```

## func 
```go
func (node *XTraceClient) String() string
```

```go
type XTraceInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (xt *XTraceInterface) GetMethods() []service.Method
```

## func 
```go
func (xt *XTraceInterface) GetName() string
```

```go
type XTraceServerContainer struct {
	docker.Container
```
```go
	ServerName string
	BindAddr   *address.BindConfig
	Iface      *goparser.ParsedInterface
}
```
## func 
```go
func (node *XTraceServerContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *XTraceServerContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *XTraceServerContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *XTraceServerContainer) Name() string
```

## func 
```go
func (node *XTraceServerContainer) String() string
```

```go
type XtraceClientWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs
```
```go
	InstanceName string
```
```go
	Wrapped  golang.Service
	XTClient *XTraceClient
	// Has unexported fields.
}
```
## func 
```go
func (node *XtraceClientWrapper) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *XtraceClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *XtraceClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *XtraceClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *XtraceClientWrapper) ImplementsGolangNode()
```

## func 
```go
func (node *XtraceClientWrapper) ImplementsGolangService()
```

## func 
```go
func (node *XtraceClientWrapper) Name() string
```

## func 
```go
func (node *XtraceClientWrapper) String() string
```

```go
type XtraceServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs
```
```go
	InstanceName string
```
```go
	Wrapped  golang.Service
	XTClient *XTraceClient
	// Has unexported fields.
}
```
## func 
```go
func (node *XtraceServerWrapper) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *XtraceServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *XtraceServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *XtraceServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *XtraceServerWrapper) ImplementsGolangNode()
```

## func 
```go
func (node *XtraceServerWrapper) ImplementsGolangService()
```

## func 
```go
func (node *XtraceServerWrapper) Name() string
```

## func 
```go
func (node *XtraceServerWrapper) String() string
```


