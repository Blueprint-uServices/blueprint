---
title: plugins/http
---
# plugins/http
```go
package http // import "gitlab.mpi-sws.org/cld/blueprint/plugins/http"
```

## FUNCTIONS

## func Deploy
```go
func Deploy(spec wiring.WiringSpec, serviceName string)
```
Deploys `serviceName` as a HTTP server. This can only be done if
`serviceName` is a pointer from Golang nodes to Golang nodes.

This call adds both src and dst side modifiers to `serviceName`. After this,
the pointer will be from addr to addr and can no longer modified with golang
nodes.


## TYPES

```go
type GolangHttpClient struct {
	golang.Node
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	ServerAddr   *address.Address[*GolangHttpServer]
```
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangHttpClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangHttpClient) AddInterfaces(builder golang.ModuleBuilder) error
```
Just makes sure that the interface exposed by the server is included in the
built module

## func 
```go
func (node *GolangHttpClient) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *GolangHttpClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangHttpClient) ImplementsGolangNode()
```

## func 
```go
func (node *GolangHttpClient) ImplementsGolangService()
```

## func 
```go
func (n *GolangHttpClient) Name() string
```

## func 
```go
func (n *GolangHttpClient) String() string
```

```go
type GolangHttpServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Addr         *address.Address[*GolangHttpServer]
	Wrapped      golang.Service
```
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangHttpServer) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangHttpServer) GenerateFuncs(builder golang.ModuleBuilder) error
```
Generates the HTTP Server handler

## func 
```go
func (node *GolangHttpServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangHttpServer) ImplementsGolangNode()
```

## func 
```go
func (n *GolangHttpServer) Name() string
```

## func 
```go
func (n *GolangHttpServer) String() string
```

Represents a service that is exposed over HTTP
```go
type HttpInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (i *HttpInterface) GetMethods() []service.Method
```

## func 
```go
func (i *HttpInterface) GetName() string
```


