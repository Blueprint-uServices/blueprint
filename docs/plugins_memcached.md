---
title: plugins/memcached
---
# plugins/memcached
```go
package memcached // import "gitlab.mpi-sws.org/cld/blueprint/plugins/memcached"
```

## FUNCTIONS

## func PrebuiltContainer
```go
func PrebuiltContainer(spec wiring.WiringSpec, cacheName string) string
```
Defines a cache called `cacheName` that uses the pre-built memcached process
image


## TYPES

```go
type MemcachedContainer struct {
	backend.Cache
	docker.Container
```
```go
	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}
```
## func 
```go
func (node *MemcachedContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *MemcachedContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *MemcachedContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (n *MemcachedContainer) Name() string
```

## func 
```go
func (n *MemcachedContainer) String() string
```

```go
type MemcachedGoClient struct {
	golang.Service
	backend.Cache
```
```go
	InstanceName string
	DialAddr     *address.DialConfig
```
```go
	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}
```
## func 
```go
func (node *MemcachedGoClient) AddInstantiation(builder golang.GraphBuilder) error
```
Part of code generation compilation pass; provides instantiation snippet

## func 
```go
func (node *MemcachedGoClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *MemcachedGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (n *MemcachedGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *MemcachedGoClient) ImplementsGolangNode()
```

## func 
```go
func (node *MemcachedGoClient) ImplementsGolangService()
```

## func 
```go
func (n *MemcachedGoClient) Name() string
```

## func 
```go
func (n *MemcachedGoClient) String() string
```

```go
type MemcachedInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (m *MemcachedInterface) GetMethods() []service.Method
```

## func 
```go
func (m *MemcachedInterface) GetName() string
```


