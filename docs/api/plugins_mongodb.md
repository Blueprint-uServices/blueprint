---
title: plugins/mongodb
---
# plugins/mongodb
```go
package mongodb // import "gitlab.mpi-sws.org/cld/blueprint/plugins/mongodb"
```

## FUNCTIONS

## func PrebuiltContainer
```go
func PrebuiltContainer(spec wiring.WiringSpec, dbName string) string
```

## TYPES

```go
type MongoDBContainer struct {
	docker.Container
	backend.NoSQLDB
```
```go
	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}
```
## func 
```go
func (node *MongoDBContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error
```

## func 
```go
func (node *MongoDBContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (m *MongoDBContainer) GenerateArtifacts(outdir string) error
```

## func 
```go
func (m *MongoDBContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (m *MongoDBContainer) Name() string
```

## func 
```go
func (m *MongoDBContainer) String() string
```

```go
type MongoDBGoClient struct {
	golang.Service
	backend.NoSQLDB
	InstanceName string
	Addr         *address.DialConfig
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}
```
## func 
```go
func (n *MongoDBGoClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (n *MongoDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (n *MongoDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```

## func 
```go
func (n *MongoDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *MongoDBGoClient) ImplementsGolangNode()
```

## func 
```go
func (node *MongoDBGoClient) ImplementsGolangService()
```

## func 
```go
func (m *MongoDBGoClient) Name() string
```

## func 
```go
func (m *MongoDBGoClient) String() string
```

```go
type MongoInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}
```
## func 
```go
func (m *MongoInterface) GetMethods() []service.Method
```

## func 
```go
func (m *MongoInterface) GetName() string
```


