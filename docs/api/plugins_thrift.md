---
title: plugins/thrift
---
# plugins/thrift
```go
package thrift // import "gitlab.mpi-sws.org/cld/blueprint/plugins/thrift"
```

## FUNCTIONS

## func Deploy
```go
func Deploy(spec wiring.WiringSpec, serviceName string)
```

## TYPES

```go
type GolangThriftClient struct {
	golang.Node
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	ServerAddr   *address.Address[*GolangThriftServer]
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangThriftClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangThriftClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *GolangThriftClient) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *GolangThriftClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangThriftClient) ImplementsGolangNode()
```

## func 
```go
func (node *GolangThriftClient) ImplementsGolangService()
```

## func 
```go
func (n *GolangThriftClient) Name() string
```

## func 
```go
func (n *GolangThriftClient) String() string
```

```go
type GolangThriftServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Addr         *address.Address[*GolangThriftServer]
	Wrapped      golang.Service
```
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *GolangThriftServer) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *GolangThriftServer) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *GolangThriftServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *GolangThriftServer) ImplementsGolangNode()
```

## func 
```go
func (n *GolangThriftServer) Name() string
```

## func 
```go
func (n *GolangThriftServer) String() string
```

```go
type ThriftInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (thrift *ThriftInterface) GetMethods() []service.Method
```

## func 
```go
func (thrift *ThriftInterface) GetName() string
```


