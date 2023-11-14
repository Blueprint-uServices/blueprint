---
title: plugins/linuxcontainer
---
# plugins/linuxcontainer
```go
package linuxcontainer // import "gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer"
```

## FUNCTIONS

## func AddProcessToContainer
```go
func AddProcessToContainer(spec wiring.WiringSpec, containerName, childName string)
```
Adds a process to an existing container

## func CreateContainer
```go
func CreateContainer(spec wiring.WiringSpec, containerName string, children ...string) string
```
Adds a container that will explicitly instantiate all of the named child
processes The container will also implicitly instantiate any of the
dependencies of the children

## func RegisterBuilders
```go
func RegisterBuilders()
```
to trigger module initialization and register builders

## func NewDockerWorkspace
```go
func NewDockerWorkspace(name string, dir string) *dockerWorkspaceImpl
```
## func NewBasicWorkspace
```go
func NewBasicWorkspace(name string, dir string) *filesystemWorkspace
```
Creates a BasicWorkspace, which is the simplest process workspace that can
write processes to an output directory


## TYPES

```go
type Container struct {
	ir.IRNode
```
```go
	InstanceName   string
	ImageName      string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
	// Has unexported fields.
}
```
## func 
```go
func (node *Container) AddArg(argnode ir.IRNode)
```

## func 
```go
func (node *Container) AddChild(child ir.IRNode) error
```

## func 
```go
func (node *Container) AddContainerArtifacts(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *Container) AddContainerInstance(target docker.ContainerWorkspace) error
```

## func 
```go
func (node *Container) GenerateArtifacts(dir string) error
```
From the core.ArtifactGenerator interface

This is the starting point for generating process workspace artifacts.

Collects process artifacts into a directory on the local filesystem and
generates a build.sh and run.sh script.

The output processes will be runnable in the local environment.

## func 
```go
func (node *Container) ImplementsDockerContainer()
```

## func 
```go
func (node *Container) Name() string
```

## func 
```go
func (node *Container) String() string
```

Used during building to accumulate linux process nodes Non-linux process
nodes will just be recursively fetched from the parent namespace
```go
type LinuxNamespace struct {
	wiring.SimpleNamespace
	// Has unexported fields.
}
```

