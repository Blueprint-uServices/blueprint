---
title: plugins/zipkin
---
# plugins/zipkin
```go
package zipkin // import "gitlab.mpi-sws.org/cld/blueprint/plugins/zipkin"
```

## FUNCTIONS

## func DefineZipkinCollector
```go
func DefineZipkinCollector(spec wiring.WiringSpec, collectorName string) string
```
Defines the Zipkin collector as a process node. Also creates a pointer to
the collector and a client node that are used by clients.


## TYPES

```go
type ZipkinCollectorClient struct {
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
func (node *ZipkinCollectorClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *ZipkinCollectorClient) AddInterfaces(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *ZipkinCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (node *ZipkinCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *ZipkinCollectorClient) ImplementsGolangNode()
```

## func 
```go
func (node *ZipkinCollectorClient) ImplementsOTCollectorClient()
```

## func 
```go
func (node *ZipkinCollectorClient) Name() string
```

## func 
```go
func (node *ZipkinCollectorClient) String() string
```

```go
type ZipkinCollectorContainer struct {
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
func (node *ZipkinCollectorContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error
```

## func 
```go
func (node *ZipkinCollectorContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *ZipkinCollectorContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *ZipkinCollectorContainer) Name() string
```

## func 
```go
func (node *ZipkinCollectorContainer) String() string
```

```go
type ZipkinInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (j *ZipkinInterface) GetMethods() []service.Method
```

## func 
```go
func (j *ZipkinInterface) GetName() string
```


