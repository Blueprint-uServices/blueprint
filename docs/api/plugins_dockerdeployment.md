---
title: plugins/dockerdeployment
---
# plugins/dockerdeployment
```go
package dockerdeployment // import "gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment"
```

## FUNCTIONS

## func AddContainerToDeployment
```go
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string)
```
Adds a child node to an existing container deployment

## func NewDeployment
```go
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string
```
Adds a deployment that explicitly instantiates all of the containers
provided. The deployment will also implicitly instantiate any of the
dependencies of the containers

## func RegisterBuilders
```go
func RegisterBuilders()
```
to trigger module initialization and register builders

## func NewDockerComposeWorkspace
```go
func NewDockerComposeWorkspace(name string, dir string) *dockerComposeWorkspace
```

## TYPES

A deployment is a collection of containers
```go
type Deployment struct {
	DeploymentName string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
	// Has unexported fields.
}
```
## func 
```go
func (node *Deployment) AddArg(argnode ir.IRNode)
```

## func 
```go
func (node *Deployment) AddChild(child ir.IRNode) error
```

## func 
```go
func (node *Deployment) GenerateArtifacts(dir string) error
```

## func 
```go
func (node *Deployment) Name() string
```

## func 
```go
func (node *Deployment) String() string
```

Used during building to accumulate docker container nodes Non-container
nodes will just be recursively fetched from the parent namespace
```go
type DockerComposeNamespace struct {
	wiring.SimpleNamespace
	// Has unexported fields.
}
```

