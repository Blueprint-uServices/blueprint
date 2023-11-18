---
title: plugins/http/httpcodegen
---
# plugins/http/httpcodegen
```go
package httpcodegen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/http/httpcodegen"
```

## FUNCTIONS

## func GenerateClient
```go
func GenerateClient(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
## func GenerateServerHandler
```go
func GenerateServerHandler(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
This function is used by the HTTP plugin to generate the server-side HTTP
service.


