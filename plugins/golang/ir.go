package golang

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
The golang plugin extends Blueprint's IR as follows:

It defines the following IR interfaces:

  - golang.Node is the base interface for any node that lives within a golang process
  - golang.Service is a golang node that has methods that can be directly called by other golang nodes

To support golang code generation, the following IR interfaces are provided.
The golang.Process depends on these interfaces for collecting and packaging code, however, usage of these interfaces
is not intended to be private to just the golang.Process plugin.  Other plugins are permitted to
use these interfaces.

  - golang.Instantiable is for golang nodes that can generate instantiation source code snippets
  - golang.ProvidesInterface is for golang nodes that include interfaces or generate new ones
  - golang.GeneratesFuncs is for golang nodes that generate function implementations
  - golang.ProvidesModule is for golang nodes that generate or otherwise provide the full source code of modules
*/
type (
	/*
		golang.Node is the base IRNode interface that should be implemented by any IRNode that
		wishes to exist within a Golang namespace.
	*/
	Node interface {
		ir.IRNode
		ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
	}

	/*
		golang.Service is a golang.Node that exposes an interface that can be directly invoked
		by other golang.Nodes.

		For example, services within a workflow spec are represented by golang.Service nodes
		because they have invokable methods.  Similarly plugins such as tracing, which
		wrap service nodes, are themselves also service nodes, because they have invokable methods.

		golang.Service extends the golang.Instantiable interface, which is part of the codegen
		process.  Thus any plugin that provides IRNodes that extend golang.Service must implement
		the code generation methods defined by the golang.Instantiable interface.
	*/
	Service interface {
		Node
		Instantiable      // Services must implement the Instantiable interface in order to create instances
		ProvidesInterface // Services must include any interfaces that they implement or define
		service.ServiceNode
		ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
	}
)

/*
Golang code-generation interfaces that IR nodes should implement if they expect to generate code or artifacts
*/
type (
	/*
	   This is an interface for IRNodes that can be used by plugins that want to be able to instantiate things in
	   the generated code.  For example, a service, or a tracing wrapper, will want to instantiate code.

	   All Services are Instantiable, but not all Instantiable are Services.  For example, a Golang GRPC
	   server is instantiable but it does not expose methods that can be directly invoked at the application
	   level.  In constract a Golang GRPC client is a service and is instantiable because it does expose methods
	   that can be directly invoked.

	   The NamespaceBuilder struct provides functionality for plugins to declare:

	   	(a) how to instantiate instances of things
	   	(b) relevant types and method signatures for use by other plugins
	*/
	Instantiable interface {
		/*
		   AddInstantion will be invoked during compilation to enable an IRNode to add its instantiation code
		   to a generated golang file.
		*/
		AddInstantiation(NamespaceBuilder) error
	}

	/*
		Service nodes need to include interface definitions that callers of the code depend on.  The most basic
		example is that a workflow service needs to include the code where it's defined.

		Some nodes, primarily modifier nodes, will also generate new interfaces by extending the interfaces
		of other IRNodes.  For example, tracing nodes might add context arguments; the healthchecker node
		might add new functions.

		Here is where interfaces should be included or generated.

		There are three interfaces for contributing source code to generated modules:
		1. RequiresPackage for adding module dependencies to the generated module's go.mod file
		2. ProvidesInterface for generating struct and interface type declarations
		3. GeneratesFuncs for generating functions and struct method bodies
	*/
	ProvidesInterface interface {
		/*
			AddInterfaces will be invoked during compilation to enable an IRNode to include the code where its
			interface is defined, e.g. in an external module (like a workflow spec node) or by auto-generating
			code that defines the interface (like a modifier node).

			AddInterfaces might be called in situations where ONLY the interfaces are needed, but not the
			implementation.  For example, the client-side of an RPC call needs to know the remote interfaces,
			but does not need to know the server-side implementation.  That logic belongs in the GenerateFuncs
			interface and method
		*/
		AddInterfaces(ModuleBuilder) error
	}

	/*
		Some IRNodes generate implementations of service interfaces.  This interface should be used to do so.  This is
		separate from the GeneratesTypes interface because all structs and interfaces need to be declared before
		method bodies can be written, because method bodies might need to use structs and interfaces defined in different
		packages

		There are three interfaces for contributing source code to generated modules:
		1. RequiresPackage for adding module dependencies to the generated module's go.mod file
		2. ProvidesInterface for generating struct and interface type declarations
		3. GeneratesFuncs for generating functions and struct method bodies
	*/
	GeneratesFuncs interface {
		/*
		   GenerateFuncs will be invoked during compilation to enable an IRNode to write generated code files to
		   an output module, containing interface and struct definitions
		*/
		GenerateFuncs(ModuleBuilder) error
	}

	/*
	   This is an interface for IRNodes for plugins that want to include standalone modules in the output workspace.
	   The most straightforward example is the workflow spec, which will be copied into the output workspace
	   using this interface.

	   The IRNode must implement the `AddToWorkspace` method, to interact with the `WorkspaceBuilder` to copy
	   relevant modules.
	*/
	ProvidesModule interface {
		/*
			AddToWorkspace will be invoked during compilation to enable an IRNode to copy a local Go module
			directly into the output workspace directory
		*/
		AddToWorkspace(WorkspaceBuilder) error
	}
)

/*
APIs used by the above IR nodes when they are generating code.

The main implementation of these interfaces is in the [goprocess](../goprocess) plugin
*/
type (
	WorkspaceInfo struct {
		Path string // fully-qualified path on the filesystem to this workspace
	}

	/*
	   WorkspaceBuilder is used by plugins if they want to collect and combine Golang code and modules.

	   An IRNode must implement the ProvidesModule interface; then during compilation, `AddToWorkspace`
	   will be called, enabling the IRNode to add its code and modules to the output workspace directory
	   using the methods on `WorkspaceBuilder`.
	*/
	WorkspaceBuilder interface {
		ir.BuildContext

		/*
			Metadata into about the workspace being built
		*/
		Info() WorkspaceInfo

		/*
			   This method is used by plugins if they want to copy a locally-defined module into the generated workspace.

			   The specified moduleSrcPath must point to a valid Go module with a go.mod file.

				Returns the path to the module in the output directory
		*/
		AddLocalModule(shortName string, moduleSrcPath string) (string, error)

		/*
			This is a variant of `AddLocalMethod` provided for convenience; instead of an absolute filesystem path, the
			specified path is relative to the caller

			Returns the path to the module in the output directory
		*/
		AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) (string, error)

		/*
			This method is used by plugins if they want to create a module in the workspace to then generate code into.

			The specified moduleName must be a golang style module name.

			This will create the directory for the module and an empty go.mod file.

			Returns the path to the module in the output directory
		*/
		CreateModule(moduleName string, moduleVersion string) (string, error)

		/*
		   If the specified module exists locally within the workspace, gets the subdirectory within the workspace that it exists in, the module
		   version, and returns true.

		   Returns "", false otherwise
		*/
		GetLocalModule(modulePath string) (string, bool)
	}

	ModuleInfo struct {
		Name    string // Fully-qualified module name being built
		Version string // Version of the module being built
		Path    string // The path on the filesystem to the directory containing the module
	}

	PackageInfo struct {
		ShortName   string // Shortname of the package
		Name        string // Fully package name within the module
		PackageName string // Fully qualified package name including module name
		Path        string // Fully-qualified path to the package
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
	ModuleBuilder interface {
		ir.BuildContext

		/*
			Metadata into about the module being built
		*/
		Info() ModuleInfo

		/*
			This creates a package within the module, with the specified package name.
			It will create the necessary output directories and returns information
			about the created package.  The provided packageName should take the form a/b/c
			This call will succeed even if the package already exists on the filesystem.
		*/
		CreatePackage(packageName string) (PackageInfo, error)

		/*
			Gets the WorkspaceBuilder that contains this ModuleBuilder
		*/
		Workspace() WorkspaceBuilder
	}

	NamespaceInfo struct {
		Package  PackageInfo
		FileName string // Name of the file within the pacakge
		FilePath string // Fully-qualified path to the file on the filesystem
		FuncName string // Name of the function that builds the namespace
	}

	/*
	   NamespaceBuilder is used by IRNodes that implement the Instantiable interface.  The NamespaceBuilder provides
	   the following methods that can be used by plugins to provide instantiation code:

	     - `Import` declares that a particular package should be imported, as it will be used by the
	       instantiation code

	     - `Declare` provides a buildFunc as a string that will be inserted into the output file; buildFunc
	       is used at runtime to create the instance

	   In the generated golang code, instances are declared and created using a simple dependency injection
	   style.  The runtime dependency injection interface is defined in runtime/plugins/golang/di.go

	   The basic requirement of an instantiable node is that it can provide a buildFunc definition that
	   will be invoked at runtime to create the instance.  A buildFunc has method signature:

	   	func(n *golang.Namespace) (any, error)

	   The buildFunc will instantiate and return an instance or an error.  If the node needs to be
	   able to call other instances, it can acquire the instances through the golang.Namespace's Get
	   method.  For example, the following pseudocode for a tracing wrapper class would get the
	   underlying handler then return the wrapper class:

	   	func(n *golang.Namespace) (any, error) {
	   		handler, err := n.Get("serviceA.handler")
	   		if err != nil {
	   			return nil, err
	   		}

	   		serviceA, isValid := handler.(ServiceA)
	   		if !isValid {
	   			return nil, blueprint.Errorf("serviceA.handler does not implement ServiceA interface")
	   		}

	   		return newServiceATracingWrapper(serviceA), nil
	   	}

	   The above code makes reference to names like `serviceA.handler`; rarely should these names
	   be hard-coded, instead they would typically be provided by calling or inspecting the IR
	   dependencies of this node.
	*/
	NamespaceBuilder interface {
		ir.BuildContext

		/*
			Metadata info about the namespace being built
		*/
		Info() NamespaceInfo

		/*
			Adds an import statement to the generated file; this is necessary for any types
			declared in other packages that are going to be used in a DI declaration.

			This method returns the type alias that should be used in the generated code.
			By default the type alias is just the package name, but if there are multiple
			different imports with the same package name, then aliases will be created
		*/
		Import(packageName string) string

		/*
			If the provided type is a user type or a builtin type, adds an import statement
			similar to the `Import` method.

			Returns the name that should be used in code for the type.  For example, if it's
			a type from an imported package, then would return mypackage.Foo.
		*/
		ImportType(typeName gocode.TypeName) string

		/*
			Provides the source code of a buildFunc that will be invoked at runtime by the
			generated code, to build the named instance
		*/
		Declare(instanceName string, buildFuncSrc string) error

		/*
			This is like Declare, but instead of having to manually construct the source
			code, the NamespaceBuilder will automatically create the build func src code,
			invoking the specified constructor and passing the provided nodes as args
		*/
		DeclareConstructor(name string, constructor *gocode.Constructor, args []ir.IRNode) error

		/*
			Specify nodes needed by this namespace that exist in a parent namespace
			and will be passed as runtime arguments
		*/
		RequiredArg(name, description string)

		/*
			Specify optional arguments to this namespace.  If an optional argument is not
			present then a runtime error can occur if a node needs to access the missing
			argument
		*/
		OptionalArg(name, description string)

		/*
			Specify nodes that should be immediately built when the namespace is instantiated
		*/
		Instantiate(name string)

		/*
			Specify priority nodes that should be built before any other nodes are built
		*/
		PriorityInstantiate(name string)

		/*
			Gets the ModuleBuilder that contains this NamespaceBuilder
		*/
		Module() ModuleBuilder
	}
)
