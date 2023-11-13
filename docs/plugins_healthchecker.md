---
title: plugins/healthchecker
---
# plugins/healthchecker
```go
package healthchecker // import "gitlab.mpi-sws.org/cld/blueprint/plugins/healthchecker"
```

## FUNCTIONS

## func AddHealthCheckAPI
```go
func AddHealthCheckAPI(spec wiring.WiringSpec, serviceName string)
```

## TYPES

```go
type HealthCheckerServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Wrapped      golang.Service
```
```go
	// Has unexported fields.
}
```
## func 
```go
func (node *HealthCheckerServerWrapper) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *HealthCheckerServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *HealthCheckerServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *HealthCheckerServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *HealthCheckerServerWrapper) ImplementsGolangNode()
```

## func 
```go
func (node *HealthCheckerServerWrapper) Name() string
```

## func 
```go
func (node *HealthCheckerServerWrapper) String() string
```


