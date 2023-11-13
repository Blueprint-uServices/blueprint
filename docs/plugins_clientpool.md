---
title: plugins/clientpool
---
# plugins/clientpool
```go
package clientpool // import "gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
```

## FUNCTIONS

## func Create
```go
func Create(spec wiring.WiringSpec, serviceName string, n int)
```
Wraps the client side of a service with a client pool with N client
instances


## TYPES

```go
type ClientPool struct {
	golang.Service
	golang.GeneratesFuncs
```
```go
	PoolName       string
	N              int
	Client         golang.Service
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
}
```
## func 
```go
func (pool *ClientPool) AddArg(argnode ir.IRNode)
```

## func 
```go
func (pool *ClientPool) AddChild(child ir.IRNode) error
```

## func 
```go
func (pool *ClientPool) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (pool *ClientPool) AddInterfaces(module golang.ModuleBuilder) error
```

## func 
```go
func (pool *ClientPool) GenerateFuncs(module golang.ModuleBuilder) error
```

## func 
```go
func (pool *ClientPool) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *ClientPool) Name() string
```

## func 
```go
func (node *ClientPool) String() string
```

```go
type ClientpoolNamespace struct {
	wiring.SimpleNamespace
```
```go
	// Has unexported fields.
}
```
## func NewClientPoolNamespace
```go
func NewClientPoolNamespace(parent wiring.Namespace, spec wiring.WiringSpec, name string, n int) *ClientpoolNamespace
```


