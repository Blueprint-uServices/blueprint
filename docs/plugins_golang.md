---
title: plugins/golang
---
# plugins/golang
```go
package golang // import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
```

## FUNCTIONS

## func AddRuntimeModule
```go
func AddRuntimeModule(workspace WorkspaceBuilder) error
```
## func GetGoInterface
```go
func GetGoInterface(ctx ir.BuildContext, node ir.IRNode) (*gocode.ServiceInterface, error)
```
Helper method that does typecasting on builder and service.

Assumes builder is a golang module builder, and service is a golang module;
if so, gets the golang service interface for the service. If not, returns an
error.


## TYPES

```go
type GeneratesFuncs interface {
	//		   GenerateFuncs will be invoked during compilation to enable an IRNode to write generated code files to
	//		   an output module, containing interface and struct definitions
```
Some IRNodes generate implementations of service interfaces. This interface
should be used to do so. This is separate from the GeneratesTypes interface
because all structs and interfaces need to be declared before method
bodies can be written, because method bodies might need to use structs and
interfaces defined in different packages
```go
	GenerateFuncs(ModuleBuilder) error
}
```
There are three interfaces for contributing source code to generated
modules: 1. RequiresPackage for adding module dependencies to the generated
module's go.mod file 2. ProvidesInterface for generating struct and
interface type declarations 3. GeneratesFuncs for generating functions and
struct method bodies

```go
type GraphBuilder interface {
	ir.BuildContext
```
```go
	//			Metadata info about the graph being built
```
```go
	Info() GraphInfo
```
```go
	//			Adds an import statement to the generated file; this is necessary for any types
	//			declared in other packages that are going to be used in a DI declaration.
	//
	//			This method returns the type alias that should be used in the generated code.
	//			By default the type alias is just the package name, but if there are multiple
	//			different imports with the same package name, then aliases will be created
```
```go
	Import(packageName string) string
```
```go
	//			If the provided type is a user type or a builtin type, adds an import statement
	//			similar to the `Import` method.
	//
	//			Returns the name that should be used in code for the type.  For example, if it's
	//			a type from an imported package, then would return mypackage.Foo.
```
```go
	ImportType(typeName gocode.TypeName) string
```
```go
	//			Provides the source code of a buildFunc that will be invoked at runtime by the
	//			generated code, to build the named instance
```
```go
	Declare(instanceName string, buildFuncSrc string) error
```
```go
	//			This is like Declare, but instead of having to manually construct the source
	//			code, the GraphBuilder will automatically create the build func src code,
	//			invoking the specified constructor and passing the provided nodes as args
```
```go
	DeclareConstructor(name string, constructor *gocode.Constructor, args []ir.IRNode) error
```
```go
	//			Gets the ModuleBuilder that contains this GraphBuilder
```
GraphBuilder is used by IRNodes that implement the Instantiable interface.
The GraphBuilder provides the following methods that can be used by plugins
to provide instantiation code:
```go
	Module() ModuleBuilder
}
```
  - `Import` declares that a particular package should be imported, as it
    will be used by the instantiation code

  - `Declare` provides a buildFunc as a string that will be inserted into
    the output file; buildFunc is used at runtime to create the instance

In the generated golang code, instances are declared and created using
a simple dependency injection style. The runtime dependency injection
interface is defined in runtime/plugins/golang/di.go

The basic requirement of an instantiable node is that it can provide a
buildFunc definition that will be invoked at runtime to create the instance.
A buildFunc has method signature:

    func(ctr golang.Container) (any, error)

The buildFunc will instantiate and return an instance or an error. If the
node needs to be able to call other instances, it can acquire the instances
through the golang.Container's Get method. For example, the following
pseudocode for a tracing wrapper class would get the underlying handler then
return the wrapper class:

    func(ctr golang.Container) (any, error) {
    	handler, err := ctr.Get("serviceA.handler")
    	if err != nil {
    		return nil, err
    	}

    	serviceA, isValid := handler.(ServiceA)
    	if !isValid {
    		return nil, blueprint.Errorf("serviceA.handler does not implement ServiceA interface")
    	}

    	return newServiceATracingWrapper(serviceA), nil
    }

The above code makes reference to names like `serviceA.handler`; rarely
should these names be hard-coded, instead they would typically be provided
by calling or inspecting the IR dependencies of this node.

APIs used by the above IR nodes when they are generating code.
```go
type GraphInfo struct {
	Package  PackageInfo
	FileName string // Name of the file within the pacakge
	FilePath string // Fully-qualified path to the file on the filesystem
	FuncName string // Name of the function that builds the graph
}
```
The main implementation of these interfaces is in the
[goprocess](../goprocess) plugin

```go
type Instantiable interface {
	//		   AddInstantion will be invoked during compilation to enable an IRNode to add its instantiation code
	//		   to a generated golang file.
```
This is an interface for IRNodes that can be used by plugins that want to be
able to instantiate things in the generated code. For example, a service,
or a tracing wrapper, will want to instantiate code.
```go
	AddInstantiation(GraphBuilder) error
}
```
All Services are Instantiable, but not all Instantiable are Services.
For example, a Golang GRPC server is instantiable but it does not expose
methods that can be directly invoked at the application level. In constract
a Golang GRPC client is a service and is instantiable because it does expose
methods that can be directly invoked.

The GraphBuilder struct provides functionality for plugins to declare:

    (a) how to instantiate instances of things
    (b) relevant types and method signatures for use by other plugins

```go
type ModuleBuilder interface {
	ir.BuildContext
```
```go
	//			Metadata into about the module being built
```
```go
	Info() ModuleInfo
```
```go
	//			This creates a package within the module, with the specified package name.
	//			It will create the necessary output directories and returns information
	//			about the created package.  The provided packageName should take the form a/b/c
	//			This call will succeed even if the package already exists on the filesystem.
```
```go
	CreatePackage(packageName string) (PackageInfo, error)
```
```go
	//			Gets the WorkspaceBuilder that contains this ModuleBuilder
```
ModuleBuilder is used by IRNodes for plugins that want to generate Golang
code and collect it into a module.
```go
	Workspace() WorkspaceBuilder
}
```
An IRNode must implement the RequiresPackages interface; then during
compilation, `AddToModule` will be called, enabling the IRNode to add
its dependencies and code to the output module using the methods on
`ModuleBuilder`.

After creating a module builder, plugins can directly create directories and
copy files into the ModuleDir. Any go dependencies should be added with the
Require function.

When finished building the module, plugins should call Finish to finish
building the go.mod file

APIs used by the above IR nodes when they are generating code.
```go
type ModuleInfo struct {
	Name    string // Fully-qualified module name being built
	Version string // Version of the module being built
	Path    string // The path on the filesystem to the directory containing the module
}
```
The main implementation of these interfaces is in the
[goprocess](../goprocess) plugin

golang.Node is the base IRNode interface that should be implemented by any
IRNode that wishes to exist within a Golang namespace.
```go
type Node interface {
	ir.IRNode
	ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
}
```
APIs used by the above IR nodes when they are generating code.
```go
type PackageInfo struct {
	ShortName   string // Shortname of the package
	Name        string // Fully package name within the module
	PackageName string // Fully qualified package name including module name
	Path        string // Fully-qualified path to the package
}
```
The main implementation of these interfaces is in the
[goprocess](../goprocess) plugin

```go
type ProvidesInterface interface {
	//			AddInterfaces will be invoked during compilation to enable an IRNode to include the code where its
	//			interface is defined, e.g. in an external module (like a workflow spec node) or by auto-generating
	//			code that defines the interface (like a modifier node).
	//
	//			AddInterfaces might be called in situations where ONLY the interfaces are needed, but not the
	//			implementation.  For example, the client-side of an RPC call needs to know the remote interfaces,
	//			but does not need to know the server-side implementation.  That logic belongs in the GenerateFuncs
	//			interface and method
```
Service nodes need to include interface definitions that callers of the
code depend on. The most basic example is that a workflow service needs to
include the code where it's defined.
```go
	AddInterfaces(ModuleBuilder) error
}
```
Some nodes, primarily modifier nodes, will also generate new interfaces by
extending the interfaces of other IRNodes. For example, tracing nodes might
add context arguments; the healthchecker node might add new functions.

Here is where interfaces should be included or generated.

There are three interfaces for contributing source code to generated
modules: 1. RequiresPackage for adding module dependencies to the generated
module's go.mod file 2. ProvidesInterface for generating struct and
interface type declarations 3. GeneratesFuncs for generating functions and
struct method bodies

```go
type ProvidesModule interface {
	//			AddToWorkspace will be invoked during compilation to enable an IRNode to copy a local Go module
	//			directly into the output workspace directory
```
This is an interface for IRNodes for plugins that want to include standalone
modules in the output workspace. The most straightforward example is the
workflow spec, which will be copied into the output workspace using this
interface.
```go
	AddToWorkspace(WorkspaceBuilder) error
}
```
The IRNode must implement the `AddToWorkspace` method, to interact with the
`WorkspaceBuilder` to copy relevant modules.

golang.Service is a golang.Node that exposes an interface that can be
directly invoked by other golang.Nodes.
```go
type Service interface {
	Node
	Instantiable      // Services must implement the Instantiable interface in order to create instances
	ProvidesInterface // Services must include any interfaces that they implement or define
	service.ServiceNode
	ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
}
```
For example, services within a workflow spec are represented by
golang.Service nodes because they have invokable methods. Similarly plugins
such as tracing, which wrap service nodes, are themselves also service
nodes, because they have invokable methods.

golang.Service extends the golang.Instantiable interface, which is part
of the codegen process. Thus any plugin that provides IRNodes that extend
golang.Service must implement the code generation methods defined by the
golang.Instantiable interface.

```go
type WorkspaceBuilder interface {
	ir.BuildContext
```
```go
	//			Metadata into about the workspace being built
```
```go
	Info() WorkspaceInfo
```
```go
	//		   This method is used by plugins if they want to copy a locally-defined module into the generated workspace.
	//
	//		   The specified moduleSrcPath must point to a valid Go module with a go.mod file.
```
```go
	AddLocalModule(shortName string, moduleSrcPath string) error
```
```go
	//			This is a variant of `AddLocalMethod` provided for convenience; instead of an absolute filesystem path, the
	//			specified path is relative to the caller
```
```go
	AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) error
```
```go
	//			This method is used by plugins if they want to create a module in the workspace to then generate code into.
	//
	//			The specified moduleName must be a golang style module name.
	//
	//			This will create the directory for the module and an empty go.mod file.
	//
	//			Returns the path to the module in the output directory
```
```go
	CreateModule(moduleName string, moduleVersion string) (string, error)
```
```go
	//		   If the specified module exists locally within the workspace, gets the subdirectory within the workspace that it exists in, the module
	//		   version, and returns true.
	//
	//		   Returns "", false otherwise
```
WorkspaceBuilder is used by plugins if they want to collect and combine
Golang code and modules.
```go
	GetLocalModule(modulePath string) (string, bool)
}
```
An IRNode must implement the ProvidesModule interface; then during
compilation, `AddToWorkspace` will be called, enabling the IRNode to add
its code and modules to the output workspace directory using the methods on
`WorkspaceBuilder`.

APIs used by the above IR nodes when they are generating code.
```go
type WorkspaceInfo struct {
	Path string // fully-qualified path on the filesystem to this workspace
}
```
The main implementation of these interfaces is in the
[goprocess](../goprocess) plugin


