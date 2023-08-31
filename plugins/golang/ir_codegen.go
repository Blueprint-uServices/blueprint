package golang

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"

/*
This file contains IR interfaces for use by other plugins that wish to generate application-level Golang code.
*/

/*
This is an interface for IRNodes that can be used by plugins that want to be able to instantiate things in
the generated code.  For example, a service, or a tracing wrapper, will want to instantiate code.

All Services are Instantiable, but not all Instantiable are Services.  For example, a Golang GRPC
server is instantiable but it does not expose methods that can be directly invoked at the application
level.  In constract a Golang GRPC client is a service and is instantiable because it does expose methods
that can be directly invoked.

The DICodeBuilder struct provides functionality for plugins to declare:

	(a) how to instantiate instances of things
	(b) relevant types and method signatures for use by other plugins
*/
type Instantiable interface {
	/*
	   AddInstantion will be invoked during compilation to enable an IRNode to add its instantiation code
	   to a generated golang file.
	*/
	AddInstantiation(DICodeBuilder) error
}

/*
This is an interface for IRNodes for plugins that want to generate source code files rather than entire modules.
The code for this IRNode, as well as other IRNodes, gets accumulated by a `ModuleBuilder` which will then
package that code into a single Golang module.

The main concern of the ModuleBuilder is to enable plugins to correctly add any dependencies to the module
being built.
*/
type RequiresPackages interface {
	/*
	   AddModule will be invoked during compilation to enable an IRNode to write generated code files to
	   an output module and to add module dependencies
	*/
	AddToModule(ModuleBuilder) error
}

/*
This is an interface for IRNodes for plugins that want to include standalone modules in the output workspace.
The most straightforward example is the workflow spec, which will be copied into the output workspace
using this interface.

The IRNode must implement the `AddToWorkspace` method, to interact with the `WorkspaceBuilder` to copy
relevant modules.
*/
type ProvidesModule interface {
	/*
		AddToWorkspace will be invoked during compimlation to enable an IRNode to copy a local Go module
		directly into the output workspace directory
	*/
	AddToWorkspace(WorkspaceBuilder) error
}

/*
WorkspaceBuilder is used by plugins if they want to collect and combine Golang code and modules.

An IRNode must implement the ProvidesModule interface; then during compilation, `AddToWorkspace`
will be called, enabling the IRNode to add its code and modules to the output workspace directory
using the methods on `WorkspaceBuilder`.
*/
type WorkspaceBuilder interface {
	irutil.VisitTracker

	/*
	   This method is used by plugins if they want to copy a locally-defined module into the generated workspace.

	   The specified moduleSrcPath must point to a valid Go module with a go.mod file.
	*/
	AddLocalModule(shortName string, moduleSrcPath string) error

	/*
		This is a variant of `AddLocalMethod` provided for convenience; instead of an absolute filesystem path, the
		specified path is relative to the caller
	*/
	AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) error

	/*
	   If the specified module exists locally within the workspace, gets the subdirectory within the workspace that it exists in, the module
	   version, and returns true.

	   Returns "", false otherwise
	*/
	GetLocalModule(modulePath string) (string, bool)
}

/*
ModuleBuilder is used by IRNodes for plugins that want to generate Golang code and collect it into a module.

An IRNode must implement the RequiresPackages interface; then during compilation, `AddToModule`
will be called, enabling the IRNode to add its dependencies and code to the output module using the
methods on `ModuleBuilder`.

After creating a module builder, plugins can directly create directories and copy files into
the ModuleDir.  Any go dependencies should be added with the Require function.

When finished building the module, plugins should call Finish to finish building the go.mod
file
*/
type ModuleBuilder interface {
	irutil.VisitTracker

	/*
		Adds a dependency to a given module and version.  The module can be an external dependency, or it can be
		a local module that resides within the same workspace.

		These dependencies will be added to the go.mod file that gets generated.

		For any local modules that reside within the same workspace, the go.mod will use `replace` directives
		to point the dependency directly to those local modules.
	*/
	Require(dependencyName string, version string) error

	/*
		Gets the WorkspaceBuilder that contains this ModuleBuilder
	*/
	Workspace() WorkspaceBuilder
}

/*
DICodeBuilder is used by IRNodes that implement the Instantiable interface.  The DICodeBuilder provides
the following methods that can be used by plugins to provide instantiation code:

  - `Import` declares that a particular package should be imported, as it will be used by the
    instantiation code

  - `Declare` provides a buildFunc as a string that will be inserted into the output file; buildFunc
    is used at runtime to create the instance

In the generated golang code, instances are declared and created using a simple dependency injection
style.  The runtime dependency injection interface is defined in runtime/plugins/golang/di.go

The basic requirement of an instantiable node is that it can provide a buildFunc definition that
will be invoked at runtime to create the instance.  A buildFunc has method signature:

	func(ctr golang.Container) (any, error)

The buildFunc will instantiate and return an instance or an error.  If the node needs to be
able to call other instances, it can acquire the instances through the golang.Container's Get
method.  For example, the following pseudocode for a tracing wrapper class would get the
underlying handler then return the wrapper class:

	func(ctr golang.Container) (any, error) {
		handler, err := ctr.Get("serviceA.handler")
		if err != nil {
			return nil, err
		}

		serviceA, isValid := handler.(ServiceA)
		if !isValid {
			return nil, fmt.Errorf("serviceA.handler does not implement ServiceA interface")
		}

		return newServiceATracingWrapper(serviceA), nil
	}

The above code makes reference to names like `serviceA.handler`; rarely should these names
be hard-coded, instead they would typically be provided by calling or inspecting the IR
dependencies of this node.
*/
type DICodeBuilder interface {
	irutil.VisitTracker

	/*
		Adds an import statement to the generated file; this is necessary for any types
		declared in other packages that are going to be used in a DI declaration.

		This method returns the type alias that should be used in the generated code.
		By default the type alias is just the package name, but if there are multiple
		different imports with the same package name, then aliases will be created
	*/
	Import(packageName string) string

	/*
		Provides the source code of a buildFunc that will be invoked at runtime by the
		generated code, to build the named instance
	*/
	Declare(instanceName string, buildFuncSrc string) error

	/*
		Gets the ModuleBuilder that contains this DICodeBuilder
	*/
	Module() ModuleBuilder
}
