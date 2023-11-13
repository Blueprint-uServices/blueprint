---
title: blueprint/pkg/coreplugins/service
---
# blueprint/pkg/coreplugins/service
```go
package service // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
```

## TYPES

```go
type Method interface {
	GetName() string
	GetArguments() []Variable
	GetReturns() []Variable
}
```
```go
type ServiceInterface interface {
	GetName() string
	GetMethods() []Method
}
```
```go
type ServiceNode interface {
```
Any IR node that represents a callable service should implement this
interface.
```go
	// Returns the interface of this service
	GetInterface(ctx ir.BuildContext) (ServiceInterface, error)
}
```
```go
type Variable interface {
	GetName() string
	GetType() string // a "well-known" type
}
```

