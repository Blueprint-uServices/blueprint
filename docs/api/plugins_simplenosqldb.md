---
title: plugins/simplenosqldb
---
# plugins/simplenosqldb
```go
package simplenosqldb // import "gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
```

## FUNCTIONS

## func Define
```go
func Define(spec wiring.WiringSpec, dbName string) string
```
Creates a simple nosqldb instance with the specified name


## TYPES

```go
type SimpleNoSQLDB struct {
	golang.Service
	backend.NoSQLDB
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
	Iface       *goparser.ParsedInterface // The NoSQLDB interface
	Constructor *gocode.Constructor       // Constructor for this SimpleNoSQLDB implementation
}
```
## func 
```go
func (node *SimpleNoSQLDB) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *SimpleNoSQLDB) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *SimpleNoSQLDB) AddToWorkspace(builder golang.WorkspaceBuilder) error
```
The nosqldb interface and SimpleNoSQLDB implementation exist in the runtime
package

## func 
```go
func (node *SimpleNoSQLDB) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *SimpleNoSQLDB) ImplementsGolangNode()
```

## func 
```go
func (node *SimpleNoSQLDB) ImplementsGolangService()
```

## func 
```go
func (node *SimpleNoSQLDB) Name() string
```

## func 
```go
func (node *SimpleNoSQLDB) String() string
```


