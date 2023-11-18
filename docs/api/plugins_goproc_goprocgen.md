---
title: plugins/goproc/goprocgen
---
# plugins/goproc/goprocgen
```go
package goprocgen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/goproc/goprocgen"
```

## FUNCTIONS

## func GenerateMain
```go
func GenerateMain(
```
Generates a main.go file in the provided module. The main method will call
the graphConstructor provided to create and instantiate nodes.
```go
	name string,
	argNodes []ir.IRNode,
	nodesToInstantiate []ir.IRNode,
	module golang.ModuleBuilder,
	graphPackage string,
	graphConstructor string) error
```

