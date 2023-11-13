---
title: plugins/redis
---
# plugins/redis
```go
package redis // import "gitlab.mpi-sws.org/cld/blueprint/plugins/redis"
```

## FUNCTIONS

## func PrebuiltContainer
```go
func PrebuiltContainer(spec wiring.WiringSpec, cacheName string) string
```
Defines a cache called `cacheName` that uses the pre-built redis image


## TYPES

```go
type RedisContainer struct {
	docker.Container
	backend.Cache
```
```go
	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}
```
## func 
```go
func (node *RedisContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *RedisContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (r *RedisContainer) GenerateArtifacts(outputDir string) error
```

## func 
```go
func (node *RedisContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (r *RedisContainer) Name() string
```

## func 
```go
func (r *RedisContainer) String() string
```

```go
type RedisGoClient struct {
	golang.Service
	backend.Cache
	InstanceName string
	Addr         *address.DialConfig
```
```go
	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}
```
## func 
```go
func (n *RedisGoClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (n *RedisGoClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (n *RedisGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (n *RedisGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *RedisGoClient) ImplementsGolangNode()
```

## func 
```go
func (node *RedisGoClient) ImplementsGolangService()
```

## func 
```go
func (n *RedisGoClient) Name() string
```

## func 
```go
func (n *RedisGoClient) String() string
```

```go
type RedisInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (r *RedisInterface) GetMethods() []service.Method
```

## func 
```go
func (r *RedisInterface) GetName() string
```


