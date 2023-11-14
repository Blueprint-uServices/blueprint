---
title: blueprint/pkg/coreplugins/backend
---
# blueprint/pkg/coreplugins/backend
```go
package backend // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/backend"
```
```go
Package backend provides IR node interfaces for common backend components.
```
## TYPES

```go
type Cache interface {
	ir.IRNode
	service.ServiceNode
}
```
```go
type NoSQLDB interface {
	ir.IRNode
	service.ServiceNode
}
```

