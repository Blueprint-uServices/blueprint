---
title: plugins/workload
---
# plugins/workload
```go
package workload // import "gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
```

## FUNCTIONS

## func GenerateWorkloadgenCode
```go
func GenerateWorkloadgenCode(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
Generates the workload generator client

## func Generator
```go
func Generator(spec wiring.WiringSpec, service string) string
```
Creates a workload generator process that will invoke the specified service


## TYPES

```go
type WorkloadgenClient struct {
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Wrapped      golang.Service
```
Golang-level client that will make calls to a service
```go
	// Has unexported fields.
}
```
## func NewWorkloadGenerator
```go
func NewWorkloadGenerator(name string, node ir.IRNode) (*WorkloadgenClient, error)
```

## func 
```go
func (node *WorkloadgenClient) AddInstantiation(builder golang.GraphBuilder) error
```
Provides the golang code to instantiate the workloadgen client

## func 
```go
func (node *WorkloadgenClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *WorkloadgenClient) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *WorkloadgenClient) ImplementsGolangNode()
```

## func 
```go
func (workloadgen *WorkloadgenClient) Name() string
```

## func 
```go
func (workloadgen *WorkloadgenClient) String() string
```


