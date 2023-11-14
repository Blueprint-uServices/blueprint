---
title: plugins/workflow
---
# plugins/workflow
```go
package workflow // import "gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
```

## FUNCTIONS

## func Define
```go
func Define(spec wiring.WiringSpec, serviceName, serviceType string, serviceArgs ...string) string
```
This adds a service to the application, using a definition that was provided
in the workflow spec.

`serviceType` must refer to a named service that was defined in the workflow
spec. If the service doesn't exist, then this will result in a build error.

`serviceArgs` can be zero or more other named nodes that are provided as
arguments to the service.

This call creates several definitions within the wiring spec. In particular,
`serviceName` is defined as a pointer to the actual service, and can thus be
modified and

## func Init
```go
func Init(srcModulePaths ...string)
```
The Golang workflow plugin must be initialized in the wiring spec with the
location of the workflow spec modules.

Workflow specs can be included from more than one source module.

The provided paths should be to the root of a go module (containing a go.mod
file). The arguments are assumed to be **relative** to the calling file.

This can be called more than once, which will concatenate all provided
srcModulePaths

## func Reset
```go
func Reset()
```

## TYPES

```go
type WorkflowService struct {
	// IR node types
	golang.Service
```
```go
	InstanceName string // Name of this instance
	ServiceType  string // The short-name serviceType used to initialize this workflow service
```
```go
	// Details of the service, including its interface and constructor
	ServiceInfo *WorkflowSpecService
```
```go
	// The workflow spec where this service originated
	Spec *WorkflowSpec
```
This Node represents a Golang Workflow spec service in the Blueprint IR.
```go
	// IR Nodes of arguments that will be passed in to the generated code
	Args []ir.IRNode
}
```
## func 
```go
func (node *WorkflowService) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *WorkflowService) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *WorkflowService) AddToWorkspace(builder golang.WorkspaceBuilder) error
```
Part of artifact generation. In addition to the interfaces, adds the
constructor to the workspace. Most likely the constructor resides in the
same module as the interfaces, but in case it doesn't, it will add the
correct module

## func 
```go
func (node *WorkflowService) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *WorkflowService) ImplementsGolangNode()
```

## func 
```go
func (node *WorkflowService) ImplementsGolangService()
```

## func 
```go
func (node *WorkflowService) Name() string
```

## func 
```go
func (n *WorkflowService) String() string
```

Representation of a workflow spec.
```go
type WorkflowSpec struct {
	Parsed *goparser.ParsedModuleSet
}
```
This code makes heavy use of the Golang code parser defined in the Golang
plugin. That code parser extracts structs, interfaces, and function
definitions from a set of golang modules.

This code adds functionality that:
  - Identifies valid service interfaces
  - Matches structs to interfaces that they implement
  - Finds constructors of structs

## func GetSpec
```go
func GetSpec() (*WorkflowSpec, error)
```
Static initialization of the workflow spec

## func NewWorkflowSpec
```go
func NewWorkflowSpec(srcModuleDirs ...string) (*WorkflowSpec, error)
```
Parses the specified module directories and loads workflow specs from there.

This will return an error if *any* of the provided srcModuleDirs are not
valid Go modules

## func 
```go
func (spec *WorkflowSpec) Get(name string) (*WorkflowSpecService, error)
```
Looks up the named service in the workflow spec. When a wiring spec
instantiates a workflow spec service, this method will ultimately get
called.

Returns the service and a constructor

```go
type WorkflowSpecService struct {
	Iface       *goparser.ParsedInterface
	Constructor *goparser.ParsedFunc
}
```

