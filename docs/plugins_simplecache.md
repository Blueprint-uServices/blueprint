---
title: plugins/simplecache
---
# plugins/simplecache
```go
package simplecache // import "gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
```

## FUNCTIONS

## func Define
```go
func Define(spec wiring.WiringSpec, cacheName string) string
```
Creates a simple cache instance with the specified name


## TYPES

```go
type SimpleCache struct {
	golang.Service
	backend.Cache
```
```go
	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.Instantiable
```
```go
	InstanceName string
```
```go
	Iface       *goparser.ParsedInterface // The Cache interface
	Constructor *gocode.Constructor       // Constructor for this Cache implementation
}
```
## func 
```go
func (node *SimpleCache) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *SimpleCache) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *SimpleCache) AddToWorkspace(builder golang.WorkspaceBuilder) error
```
The cache interface and simplecache implementation exist in the runtime
package

## func 
```go
func (node *SimpleCache) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *SimpleCache) ImplementsGolangNode()
```

## func 
```go
func (node *SimpleCache) ImplementsGolangService()
```

## func 
```go
func (node *SimpleCache) Name() string
```

## func 
```go
func (node *SimpleCache) String() string
```


