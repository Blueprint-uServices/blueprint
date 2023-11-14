---
title: plugins/retries
---
# plugins/retries
```go
package retries // import "gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
```

## FUNCTIONS

## func AddRetries
```go
func AddRetries(spec wiring.WiringSpec, serviceName string, max_retries int64)
```
Modifies the given service such that all clients to that service retry
`max_retries` number of times on error.


## TYPES

```go
type RetrierClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Wrapped      golang.Service
```
```go
	Max int64
	// Has unexported fields.
}
```
## func 
```go
func (node *RetrierClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *RetrierClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *RetrierClient) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *RetrierClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *RetrierClient) ImplementsGolangNode()
```

## func 
```go
func (node *RetrierClient) Name() string
```

## func 
```go
func (node *RetrierClient) String() string
```


