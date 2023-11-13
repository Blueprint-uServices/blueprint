---
title: blueprint/pkg/ir/
---
# blueprint/pkg/ir/
```go
package ir // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
```
```go
Package ir provides the basic interfaces for Blueprint's Internal Representation
(IR) and for subsequently generating application artifacts such as code and
container images.
```
```go
An application's IR representation is produced by constructing and then building
a wiring spec using methods from the wiring package and from wiring extensions
provided by plugins.
```
## FUNCTIONS

## func CleanName
```go
func CleanName(name string) string
```
Returns name with only alphanumeric characters and all other symbols
converted to underscores.

CleanName is primarily used by plugins to convert user-defined service
names into names that are valid as e.g. environment variables, command line
arguments, etc.

## func Filter[T
```go
func Filter[T any](nodes []IRNode) []T
```
Returns a slice containing only nodes of type T

## func RegisterDefaultBuilder[T
```go
func RegisterDefaultBuilder[T IRNode](name string, buildFunc func(outputDir string, node IRNode) error)
```
When building an application, any IR nodes of type T that reside within
the top-level application will be built using the specified buildFunc. The
buildFunc will only be invoked if there isn't a default namespace registered
for nodes of type T.

## func RegisterDefaultNamespace[T
```go
func RegisterDefaultNamespace[T IRNode](name string, buildFunc func(outputDir string, nodes []IRNode) error)
```
When building an application, any IR nodes of type T that reside within the
top-level application will be built using the specified buildFunc.


## TYPES

```go
type ApplicationNode struct {
	IRNode
	ArtifactGenerator
```
The IR Node that represents the whole application. Building a wiring spec
will return an ApplicationNode. An ApplicationNode can be built with the
GenerateArtifacts method.
```go
	ApplicationName string
	Children        []IRNode
}
```
## func 
```go
func (app *ApplicationNode) GenerateArtifacts(dir string) error
```

## func 
```go
func (node *ApplicationNode) Name() string
```

## func 
```go
func (node *ApplicationNode) String() string
```
Print the IR graph

```go
type ArtifactGenerator interface {
```
Most IRNodes can generate code artifacts but they do so in the context
of some BuildContext. A few IRNodes, however, can generate artifacts
independent of any external context. Those IRNodes implement the
ArtifactGenerator interface. Typically these are namespace nodes such as
golang processes, linux containers, or docker deployments.
```go
	// Generate all artifacts for this node to the specified dir on the local filesystem.
	GenerateArtifacts(dir string) error
}
```
All artifact generation occurs in the context of some BuildContext.
```go
type BuildContext interface {
	VisitTracker
	ImplementsBuildContext()
}
```
Plugins that control the artifact generation process should implement this
interface.

```go
type IRConfig interface {
	IRNode
	Optional() bool
	// At various points during the build process, an IRConfig node might have a concrete value
	// set, or it might be left unbound.
	HasValue() bool
```
IRConfig is an IR node that represents a configured or configurable
variable. In a generated application, IRConfig nodes typically map down
to things like environment variables or command line arguments, and can be
passed all the way into specific application-level instances. IRConfig is
also used for addressing.
```go
	// Returns the current value of the config node if it has been set.  Config values
	// are always strings.
	Value() string
	ImplementsIRConfig()
}
```
Metadata is an IR node that exists in the IR of an application but does not
build any artifacts or provide configuration or anything like that.
```go
type IRMetadata interface {
	ImplementsIRMetadata()
}
```
All nodes implement the IRNode interface
```go
type IRNode interface {
	Name() string
	String() string
}
```
## func FilterNodes[T
```go
func FilterNodes[T any](nodes []IRNode) []IRNode
```
Returns a slice containing only nodes of type T

## func Remove[T
```go
func Remove[T any](nodes []IRNode) []IRNode
```
Returns a slice containing all nodes except those of type T

A Blueprint application can potentially have multiple IR node instances
spread across the application that generate the same code.
```go
type VisitTracker interface {
	// Returns false on the first invocation of name; true on subsequent invocations
	Visited(name string) bool
}
```
Visit tracker is a utility method used during artifact generation to prevent
nodes from unnecessarily generating the same artifact repeatedly, when once
will suffice.

Basic implementation of the VisitTracker interface
```go
type VisitTrackerImpl struct {
	// Has unexported fields.
}
```
## func 
```go
func (tracker *VisitTrackerImpl) Visited(name string) bool
```


