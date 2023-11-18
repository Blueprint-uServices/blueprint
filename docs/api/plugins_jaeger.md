---
title: plugins/jaeger
---
# plugins/jaeger
```go
package jaeger // import "gitlab.mpi-sws.org/cld/blueprint/plugins/jaeger"
```

## FUNCTIONS

## func DefineJaegerCollector
```go
func DefineJaegerCollector(spec wiring.WiringSpec, collectorName string) string
```
Defines the Jaeger collector as a process node. Also creates a pointer to
the collector and a client node that are used by clients.


## TYPES

```go
type JaegerCollectorClient struct {
	golang.Node
	golang.Instantiable
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
func (node *JaegerCollectorClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *JaegerCollectorClient) AddInterfaces(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *JaegerCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *JaegerCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *JaegerCollectorClient) ImplementsGolangNode()
```

## func 
```go
func (node *JaegerCollectorClient) ImplementsOTCollectorClient()
```

## func 
```go
func (node *JaegerCollectorClient) Name() string
```

## func 
```go
func (node *JaegerCollectorClient) String() string
```

```go
type JaegerCollectorContainer struct {
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
func (node *JaegerCollectorContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error
```

## func 
```go
func (node *JaegerCollectorContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *JaegerCollectorContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *JaegerCollectorContainer) Name() string
```

## func 
```go
func (node *JaegerCollectorContainer) String() string
```

```go
type JaegerInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (j *JaegerInterface) GetMethods() []service.Method
```

## func 
```go
func (j *JaegerInterface) GetName() string
```


