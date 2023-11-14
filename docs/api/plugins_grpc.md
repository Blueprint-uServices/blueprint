---
title: plugins/grpc
---
# plugins/grpc
```go
package grpc // import "gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
```

## FUNCTIONS

## func Deploy
```go
func Deploy(spec wiring.WiringSpec, serviceName string)
```
Deploys `serviceName` as a GRPC server. This can only be done if
`serviceName` is a pointer from Golang nodes to Golang nodes.

This call adds both src and dst-side modifiers to `serviceName`. After this,
the pointer will be from addr to addr and can no longer be modified with
golang nodes.


## TYPES

Represents a service that is exposed over GRPC
```go
type GRPCInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (grpc *GRPCInterface) GetMethods() []service.Method
```

## func 
```go
func (grpc *GRPCInterface) GetName() string
```

```go
type GolangClient struct {
	golang.Service
	golang.GeneratesFuncs
```
```go
	InstanceName string
	ServerAddr   *address.Address[*GolangServer]
```
IRNode representing a client to a Golang server. This node does not
introduce any new runtime interfaces or types that can be used by other
IRNodes GRPC code generation happens during the ModuleBuilder GenerateFuncs
pass
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangClient) AddInterfaces(builder golang.ModuleBuilder) error
```
Just makes sure that the interface exposed by the server is included in the
built module

## func 
```go
func (node *GolangClient) GenerateFuncs(builder golang.ModuleBuilder) error
```
Generates proto files and the RPC client

## func 
```go
func (node *GolangClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangClient) ImplementsGolangNode()
```

## func 
```go
func (node *GolangClient) ImplementsGolangService()
```

## func 
```go
func (n *GolangClient) Name() string
```

## func 
```go
func (n *GolangClient) String() string
```

```go
type GolangServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Addr         *address.Address[*GolangServer]
	Wrapped      golang.Service
```
IRNode representing a Golang GPRC server. This node does not introduce any
new runtime interfaces or types that can be used by other IRNodes GRPC code
generation happens during the ModuleBuilder GenerateFuncs pass
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangServer) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangServer) GenerateFuncs(builder golang.ModuleBuilder) error
```
Generates proto files and the RPC server handler

## func 
```go
func (node *GolangServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangServer) ImplementsGolangNode()
```

## func 
```go
func (n *GolangServer) Name() string
```

## func 
```go
func (n *GolangServer) String() string
```


