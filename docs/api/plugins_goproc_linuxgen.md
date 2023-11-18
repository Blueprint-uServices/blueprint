---
title: plugins/goproc/linuxgen
---
# plugins/goproc/linuxgen
```go
package linuxgen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/goproc/linuxgen"
```

## FUNCTIONS

## func GenerateBinaryRunFunc
```go
func GenerateBinaryRunFunc(procName string, args ...ir.IRNode) (string, error)
```
Generates command-line function to run a goproc that has been built to a
binary using `go build`

## func GenerateDockerfileBuildCommands
```go
func GenerateDockerfileBuildCommands(goProcName string) (string, error)
```
If the goproc is being deployed to Docker, we can provide some custom build
commands to add to the Dockerfile

## func GenerateRunFunc
```go
func GenerateRunFunc(procName string, args ...ir.IRNode) (string, error)
```
Generates command-line function to run a goproc


