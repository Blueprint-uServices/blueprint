---
title: plugins/goproc
---
# plugins/goproc
```go
package goproc // import "gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
```

## FUNCTIONS

## func AddChildToProcess
```go
func AddChildToProcess(spec wiring.WiringSpec, procName, childName string)
```
Adds a child node to an existing process

## func CreateClientProcess
```go
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string
```
Creates a process that contains clients to the specified children. This is
for convenience in serving as a starting point to write a custom client

## func CreateProcess
```go
func CreateProcess(spec wiring.WiringSpec, procName string, children ...string) string
```
Adds a process that explicitly instantiates all of the children provided.
The process will also implicitly instantiate any of the dependencies of the
children

## func RegisterDefaultBuilders
```go
func RegisterDefaultBuilders()
```

## TYPES

An IRNode representing a golang process. This is Blueprint's main
implementation of Golang processes
```go
type Process struct {
	InstanceName   string
	ProcName       string
	ModuleName     string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
	// Has unexported fields.
}
```
## func 
```go
func (node *Process) AddArg(argnode ir.IRNode)
```

## func 
```go
func (node *Process) AddChild(child ir.IRNode) error
```

## func 
```go
func (node *Process) AddProcessArtifacts(builder linux.ProcessWorkspace) error
```
From process.ProvidesProcessArtifacts

## func 
```go
func (node *Process) AddProcessInstance(builder linux.ProcessWorkspace) error
```
From process.InstantiableProcess

## func 
```go
func (node *Process) GenerateArtifacts(workspaceDir string) error
```
Generates a golang process to a directory on the local filesystem.

This will collect and package all of the code for the contained Golang nodes
and generate a main.go method.

The output code will be runnable on the local filesystem, assuming the user
has configured the appropriate environment

## func 
```go
func (node *Process) ImplementsLinuxProcess()
```

## func 
```go
func (node *Process) Name() string
```

## func 
```go
func (node *Process) String() string
```

Used during building to accumulate golang application-level nodes Non-golang
nodes will just be recursively fetched from the parent namespace
```go
type ProcessNamespace struct {
	wiring.SimpleNamespace
	// Has unexported fields.
}
```

