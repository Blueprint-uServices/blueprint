// Package golang defines compiler interfaces for use by plugins that generate and instantiate golang code.
// The package does not provide any wiring spec functionality and it is not directly used by Blueprint applications;
// only by other Blueprint plugins.
//
// # Getting Started
//
// The golang package is primarily of interest to plugin developers.  Since the majority of Blueprint plugins
// operate at the application level and generate golang code, most plugins want to make use of the interfaces
// and methods defined in this package.  The recommended way of familiarizing yourself with the golang plugin
// is simply by browsing the code of other plugins, to see how they implement the interfaces.
// Recommended places to start include:
//   - the [simple] plugin, which provides simple implementations of backends such as Databases, Caches, etc.
//     During compilation, all the plugin needs to do is copy code from [runtime/plugins] to the output directory.
//   - the [retries] plugin, which generates a client-side wrapper to a service.  It contains a relatively
//     simple implementation of code generation and client instantiation that doesn't extend the service
//     interface in any way
//   - the [healthchecker] plugin, which extends a service interface to add a 'healthcheck' API call
//   - the [grpc] plugin is the most 'heavyweight' example, that includes code parsing and generating client
//     and server wrapper.
//
// # IRNodes
//   - [Node] is an interface for any IRNode that lives within a golang process.  If an IRNode implements
//     this interface then it will ultimately reside within a [goproc].
//   - [Service] is an extension of [Node] that represents a callable interface and instance.
//   - A [Node] can optionally generate code, instantiate objects, extend existing code, and more.
//     These tasks form a four-stage code-generation lifecycle.  A [Node] can hook into each of these
//     stages by implementing the [Instantiable], [ProvidesInterface], [GeneratesFuncs], and/or
//     [ProvidesModule] interfaces.
//
// # Code generation
//
// A [Node] can implement one or more of the following interfaces to hook in to the golang code-generation
// step of compilation:
//   - [ProvidesInterface] allows a [Node] to define new service interfaces or extend existing ones.  This
//     is often used by wrapper classes, which might add arguments or methods to a wrapped service.
//   - [GeneratesFuncs] allows a [Node] to generate an implementation of an interface.  This is also often
//     used by wrapper classes around services.
//   - [ProvidesModule] allows a [Node] to copy golang modules into the output workspace, typically if
//     the plugin has some helper code that it wants to use.
//   - [Instantiable] allows a [Node] to specify the runtime instances that should be created; typically these
//     will invoke code gathered using [GeneratesFuncs] or [ProvidesModule].
//
// # Code parsing
//
// Several plugins are interested in parsing golang code -- principally the workflow spec of an application.
// The purpose of this parsing is primarily to identify service interfaces, constructors, and methods.
// Parsing code is implemented in [goparser] and used by plugins such as [workflow] and [gotests].
//
// [simple]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/simple
// [retries]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/retries
// [healthchecker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/healthchecker
// [grpc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/grpc
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
// [goparser]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/goparser
// [workflow]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/workflow
// [gotests]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/gotests
// [runtime/plugins]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins
//
// [gogen]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/gogen
package golang

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

type (
	// Node should be implemented by any IRNode that ought to exist within a Golang namespace.
	Node interface {
		ir.IRNode
		ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
	}

	// Service is a [Node] that represents a callable service with an interface, constructor, and methods.
	// For example, services within a workflow spec are represented by Service nodes because they have
	// invokable methods.  Similarly plugins such as tracing, which wrap service nodes, are themselves
	// also service nodes, because they have invokable methods.
	//
	// Service nodes must implement the [Instantiable] and [ProvidesInterface] interfaces.
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
	// A [Node] should implement Instantiable if it wants to instantiate objects in the generated
	// golang namespace at runtime.  For example, a service node needs to actually call the service constructor
	// at runtime, to instantiate the service.
	//
	// All [Service] are Instantiable, but not all Instantiable are Services.  For example, a Golang GRPC
	// server is instantiable but it does not expose methods that can be directly invoked at the application
	// level.  In constract a Golang GRPC client is a [Service] and is Instantiable because it does expose methods
	// that can be directly invoked.
	Instantiable interface {

		// AddInstantiation is invoked during compilation to allow the callee to provide code snippets for instantiating
		// a golang object, using the provided [NamespaceBuilder]
		AddInstantiation(NamespaceBuilder) error
	}

	// A [Node] should implement ProvidesInterface if it wants to modify or extend any service interfaces,
	// particularly those that are defined by other nodes.  For example, a tracing plugin might extend all methods
	// of an interface to add trace contexts.
	ProvidesInterface interface {

		// AddInterfaces is invoked during compilation to allow the callee to add interfaces to the generated output,
		// using the provided [ModuleBuilder].  Interfaces could come from an external module, or they might need to
		// be auto-generated (e.g. by inspecting and modifying the interface of a wrapped [Service].
		//
		// Plugins typically also provide implementations of their interfaces.  However, code for the implementations
		// should be provided via [GeneratesFuncs] rather than [ProvidesInterface].
		AddInterfaces(ModuleBuilder) error
	}

	// A [Node] should implement GeneratesFuncs if it wants to implement any service interfaces, whether defined
	// by the node itself, or from some other node (e.g. if it wraps a service, the service's interface might be
	// used unmodified).
	GeneratesFuncs interface {
		// GenerateFuncs is invoked during compilation to allow the callee to provide interface implementations and
		// constructors.  The generated code will be included in the output.
		//
		// GenerateFuncs should be used in conjunction with [ProvidesInterface].  [ProvidesInterface] should be used
		// to declare the interface code, and GeneratesFuncs should be used to declare an interface implementation
		// and constructor.
		GenerateFuncs(ModuleBuilder) error
	}

	// A [Node] should implement ProvidesModule if it uses off-the-shelf code implemented in a golang module, and wants
	// to copy that code directly into the output.
	ProvidesModule interface {
		// AddToWorkspace is invoked during compilation to allow the callee to copy golang modules into the output
		// workspace.  Those modules might contain interface definitions or implementations, that can then be used
		// by the node, e.g. they can be instantiated if the node is [Instantiable].
		AddToWorkspace(WorkspaceBuilder) error
	}
)

type (
	// Metadata about a golang workspace
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
			Gets the ModuleBuilder that contains this NamespaceBuilder
		*/
		Module() ModuleBuilder
	}
)
