---
title: blueprint/pkg/wiring/
---
# blueprint/pkg/wiring/
```go
package wiring // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
```
```go
Package wiring provides the entry point for a Blueprint application to create
and configure a wiring spec; that wiring spec can enriched and extended by
plugins; ultimately it is used by applications to generate concrete application
instances.
```
```go
The starting point for a Blueprint application is the NewWiringSpec function.
Subsequently, Blueprint applications should typically not need to directly
invoke methods on the WiringSpec instance; instead the applications should
invoke plugins, passing the WiringSpec instance to those plugins.
```
## FUNCTIONS

## func BuildApplicationIR
```go
func BuildApplicationIR(spec WiringSpec, name string, nodesToInstantiate ...string) (*ir.ApplicationNode, error)
```
Builds the IR of an application using the definitions of the provided spec.
Returns an ir.ApplicationNode of the application.

Callers should typically provide nodesToInstantiate to specify which nodes
should be instantiated in the application. This method will recursively
instantiate any dependencies.

If nodesToInstantiate is empty, all nodes will be instantiated, but this
might not result in an application with the desired topology. Hence the
recommended approach is to explicitly specify which nodes to instantiate.


## TYPES

Creates an IR node within the provided namespace or within a new child
namespace. Other named IR nodes can be fetched from the provided Namespace
by invoking [Namespace.Get] or other Namespace methods.
```go
type BuildFunc func(Namespace) (ir.IRNode, error)
```
```go
type DefaultNamespaceHandler struct {
	SimpleNamespaceHandler
	Namespace *SimpleNamespace
```
A basic SimpleNamespaceHandler implementation that accepts nodes of all
types.
```go
	Nodes []ir.IRNode
	Edges []ir.IRNode
}
```
## func 
```go
func (handler *DefaultNamespaceHandler) Accepts(nodeType any) bool
```

## func 
```go
func (handler *DefaultNamespaceHandler) AddEdge(name string, node ir.IRNode) error
```

## func 
```go
func (handler *DefaultNamespaceHandler) AddNode(name string, node ir.IRNode) error
```

## func 
```go
func (handler *DefaultNamespaceHandler) Init(namespace *SimpleNamespace)
```

## func 
```go
func (handler *DefaultNamespaceHandler) LookupDef(name string) (*WiringDef, error)
```

```go
type Namespace interface {
	// Returns the name of this namespace
	Name() string
```
```go
	// Gets an IRNode with the specified name from this namespace, placing the result in the pointer dst.
	// dst should typically be a pointer to an IRNode type.
	// If the node has already been built, it returns the existing built node.
	// If the node hasn't yet been built, the node's [BuildFunc] will be called and the result will be
	// cached and returned.
	// This call might recursively call [Get] on a parent namespace depending on the [nodeType] registered
	// for name.
	Get(name string, dst any) error
```
```go
	// The same as [Get] but without creating a depending (an edge) into the current namespace.  Most
	// plugins should use [Get] instead.
	Instantiate(name string, dst any) error
```
```go
	// Gets a property from the wiring spec; dst should be a pointer to a value
	GetProperty(name string, key string, dst any) error
```
```go
	// Gets a slice of properties from the wiring spec; dst should be a pointer to a slice
	GetProperties(name string, key string, dst any) error
```
```go
	// Puts a node into this namespace
	Put(name string, node ir.IRNode) error
```
```go
	// Enqueue a function to be executed after all currently-queued functions have finished executing.
	// Most plugins should not need to use this.
	Defer(f func() error)
```
```go
	// Log an info-level message
	Info(message string, args ...any)
```
```go
	// Log a warn-level message
	Warn(message string, args ...any)
```
Namespace is a dependency injection container used by Blueprint plugins
during Blueprint's IR construction process. Namespaces instantiate and store
IRNodes. A root Blueprint application is itself a Namespace.
```go
	// Log an error-level message
	Error(message string, args ...any) error
}
```
A Namespace argument is passed to the BuildFunc when an IRNode is being
built. An IRNode can potentially be built multiple times, in different
namespaces.

If an IRNode depends on other IRNodes, those others can be fetched by
calling [Namespace.Get]. If those IRNodes haven't yet been built, then their
BuildFuncs will also be invoked, recursively. Conversely, if those IRNodes
are already built, then the built instance is re-used.

Namespaces are hierarchical and a namespace implementation can choose
to only support a subset of IRNodes. In this case, [Namespace.Get] on an
unsupported IRNode will recursively get the node on the parent namespace.
Namespaces inspect the nodeType argument of [WiringSpec.Define] to make this
decision.

```go
type SimpleNamespace struct {
	Namespace
```
```go
	NamespaceName   string                 // A name for this namespace
	NamespaceType   string                 // The type of this namespace
	ParentNamespace Namespace              // The parent namespace that created this namespace; can be nil
	Wiring          WiringSpec             // The wiring spec
	Handler         SimpleNamespaceHandler // User-provided handler
	Seen            map[string]ir.IRNode   // Cache of built nodes
	Added           map[string]any         // Nodes that have been passed to the handler
	Deferred        []func() error         // Deferred functions to execute
```
SimpleNamespace is a base implementation of a Namespace that provides
implementations of most methods.
```go
	// Has unexported fields.
}
```
Most plugins that want to implement a Namespace will want to use
SimpleNamespace and only provide a SimpleNamespaceHandler implementation for
a few of the custom namespace logics.

See the documentation of SimpleNamespaceHandler for methods to implement.

## func 
```go
func (namespace *SimpleNamespace) Debug(message string, args ...any)
```

## func 
```go
func (namespace *SimpleNamespace) Defer(f func() error)
```

## func 
```go
func (namespace *SimpleNamespace) Error(message string, args ...any) error
```

## func 
```go
func (namespace *SimpleNamespace) Get(name string, dst any) error
```

## func 
```go
func (namespace *SimpleNamespace) GetProperties(name string, key string, dst any) error
```

## func 
```go
func (namespace *SimpleNamespace) GetProperty(name string, key string, dst any) error
```

## func 
```go
func (namespace *SimpleNamespace) Info(message string, args ...any)
```

## func 
```go
func (namespace *SimpleNamespace) Init(name, namespacetype string, parent Namespace, wiring WiringSpec, handler SimpleNamespaceHandler)
```
Initializes a SimpleNamespace. To do so, a parent namespace, wiring spec,
and SimpleNamespaceHandler implementation must be provided.

## func 
```go
func (namespace *SimpleNamespace) Instantiate(name string, dst any) error
```

## func 
```go
func (namespace *SimpleNamespace) Name() string
```

## func 
```go
func (namespace *SimpleNamespace) Put(name string, node ir.IRNode) error
```

```go
type SimpleNamespaceHandler interface {
	// Initialize the handler with a namespace
	Init(*SimpleNamespace)
```
```go
	// Look up a [WiringDef] from the wiring spec.
	LookupDef(string) (*WiringDef, error)
```
```go
	// Reports true if this namespace can build nodes of the specified node type.
	//
	// For some node type T, if Accepts(T) returns false, then nodes of type T will
	// not be built in this namespace and instead the parent namespace will be called.
	Accepts(any) bool
```
```go
	// After a node has been gotten from the parent namespace, AddEdge will be
	// called to inform the handler that the node should be passed to this namespace
	// as an argument.
	AddEdge(string, ir.IRNode) error
```
SimpleNamespaceHandler is an interface intended for use by any Blueprint
plugin that wants to provide a custom namespace.
```go
	// After a node has been built in this namespace, AddNode will be called
	// to enable the handler to save the built node.
	AddNode(string, ir.IRNode) error
}
```
The plugin should implement the methods of this handler and then create a
SimpleNamespace and call SimpleNamespace.Init

```go
type WiringDef struct {
	Name       string
	NodeType   any
	Build      BuildFunc
	Properties map[string][]any
}
```
## func 
```go
func (def *WiringDef) AddProperty(key string, value any)
```

## func 
```go
func (def *WiringDef) GetProperties(key string, dst any) error
```

## func 
```go
func (def *WiringDef) GetProperty(key string, dst any) error
```

## func 
```go
func (def *WiringDef) String() string
```

```go
type WiringError struct {
	Errors []error
}
```
## func 
```go
func (e WiringError) Error() string
```

```go
type WiringSpec interface {
	Define(name string, nodeType any, build BuildFunc) // Adds a named node definition to the spec that can be built with the provided build function
	GetDef(name string) *WiringDef                     // For use by plugins to access the defined build functions and metadata
	Defs() []string                                    // Returns names of all defined nodes
```
```go
	Alias(name string, pointsto string)   // Defines an alias to another defined node; these can be recursive
	GetAlias(alias string) (string, bool) // Gets the value of the specified alias, if it exists
```
```go
	SetProperty(name string, key string, value any)       // Sets a static property value in the wiring spec, replacing any existing value specified
	AddProperty(name string, key string, value any)       // Adds a static property value in the wiring spec
	GetProperty(name string, key string, dst any) error   // Gets a static property value from the wiring spec
	GetProperties(name string, key string, dst any) error // Gets all static property values from the wiring spec
```
```go
	String() string // Returns a string representation of everything that has been defined
```
```go
	// Errors while building a wiring spec are accumulated within the wiring spec, rather than as return values to calls
	AddError(err error) // Used by plugins to signify an error; the error will be returned by a call to Err or GetBlueprint
	Err() error         // Gets an error if there is currently one
```
```go
	BuildIR(nodesToInstantiate ...string) (*ir.ApplicationNode, error) // After defining everything, this builds the IR for the specified named nodes (implicitly including dependencies of those nodes)
}
```
## func NewWiringSpec
```go
func NewWiringSpec(name string) WiringSpec
```


