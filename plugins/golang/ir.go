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

	// [WorkspaceBuilder] is used during Blueprint's compilation process to enable [Node] implementations
	// to generate or copy Golang code modules into the output workspace.  A [Node] must also implement the
	// [ProvidesModule] interface if it wishes to make use of the [WorkspaceBuilder].
	//
	// The main use case for [WorkspaceBuilder] is to copy existing golang modules from the local filesystem
	// into the output workspace.  A specialized use case is for generating golang modules.  However, although
	// most plugins generate golang code, most would not need to be in control of generating an entire module,
	// and would probably want to instead use the [ModuleBuilder] to contribute code to a shared module.
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

	// Metadata about a golang module that resides within a golang workspace
	ModuleInfo struct {
		Name    string // Fully-qualified module name being built
		Version string // Version of the module being built
		Path    string // The path on the filesystem to the directory containing the module
	}

	// Metadata about a package within a golang module
	PackageInfo struct {
		ShortName   string // Shortname of the package
		Name        string // Fully package name within the module
		PackageName string // Fully qualified package name including module name
		Path        string // Fully-qualified path to the package
	}

	// [ModuleBuilder] is used during Blueprint's compilation process by [Node] implementations
	// that generate code.  A [Node] must also implement the [ProvidesInterface] or [GeneratesFuncs]
	// interfaces if it wishes to make use of the [ModuleBuilder].
	//
	// [ModuleBuilder] enables plugins to create packages within a shared module.  After creating a package,
	// the plugin should then generate and place its code inside that package.
	//
	// go mod tidy will be invoked later as part of the compilation.  Packages that reside locally within the
	// workspace will be resolved locally.  Any remaining required packages will be automatically resolved.
	ModuleBuilder interface {
		ir.BuildContext

		/*
			Metadata into about the module being built
		*/
		Info() ModuleInfo

		// This creates a package within the module, with the specified package name.
		// It will create the necessary output directories and returns information
		// about the created package.  The provided packageName should take the form a/b/c
		// This call will succeed even if the package already exists on the filesystem.
		CreatePackage(packageName string) (PackageInfo, error)

		// Enables a plugin to add a 'require' statement to the go.mod file for the generated
		// module.  Typically this is not necessary because a subsequent `go mod tidy` will
		// automatically pick up module dependencies.  However, if a plugin wishes to explicitly
		// control the dependency version, it can use this method.
		Require(moduleName string, version string) error

		/*
			Gets the WorkspaceBuilder that contains this ModuleBuilder
		*/
		Workspace() WorkspaceBuilder
	}

	// Metadata about a namespace code file being generated
	NamespaceInfo struct {
		Package  PackageInfo
		FileName string // Name of the file within the pacakge
		FilePath string // Fully-qualified path to the file on the filesystem
		FuncName string // Name of the function that builds the namespace
	}

	// [NamespaceBuilder] is used during Blueprint's compilation process by [Node] implementations that
	// want to instantiate code.  A [Node] must also implement the [Instantiable] interface if it wishes to
	// make use of the [NamespaceBuilder].
	//
	// During compilation, the [NamespaceBuilder] will accumulate instantiation code from [Node] instances
	// that reside within the namespace.  Some nodes might have dependencies on other nodes; these dependencies
	// are handled automatically by the [NamespaceBuilder].  Some nodes might require arguments be passed in
	// to the namespace (e.g. command-line arguments like addresses); these are also checked by the [NamespaceBuilder].
	//
	// After all [Node] instances have added their declarations to the NamespaceBuilder, the NamespaceBuilder will
	// generate a golang file with methods for instantiating the namespace.  The namespace instantiation function
	// will invoke the instantiation code for all relevant nodes, and makes use of additional helper classes
	// defined in [runtime/plugins/golang].
	//
	// The main NamespaceBuilder implementation is in [gogen/namespacebuilder.go]
	//
	// [gogen/namespacebuilder.go]: https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/golang/gogen/namespacebuilder.go
	// [runtime/plugins/golang]: https://github.com/Blueprint-uServices/blueprint/blob/main/runtime/plugins/golang
	NamespaceBuilder interface {
		ir.BuildContext

		// Metadata info about the namespace being built
		Info() NamespaceInfo

		// Adds an import statement to the generated file; this is necessary for any types
		// declared in other packages that are going to be used in a DI declaration.
		//
		// This method returns the type alias that should be used in the generated code.
		// By default the type alias is just the package name, but if there are multiple
		// different imports with the same package name, then aliases will be created
		Import(packageName string) string

		// If the provided type is a user type or a builtin type, adds an import statement
		// similar to the `Import` method.
		//
		// Returns the name that should be used in code for the type.  For example, if it's
		// a type from an imported package, then would return mypackage.Foo.
		ImportType(typeName gocode.TypeName) string

		// Declares buildFuncSrc, the golang source code that should be invoked at runtime
		// to instantiate instanceName.  Most plugins will probably want to use [DeclareConstructor]
		// rather than [Declare].
		//
		// # buildFuncSrc
		//
		// buildFuncSrc should be a function with the following signature:
		//
		// 	func(n *golang.Namespace) (any, error)
		//
		// The first return value of the function should be the instance.  An error can be
		// returned if anything went wrong when creating the instance.
		//
		// # golang.Namespace
		//
		// The golang.Namespace argument in the buildFunc method signature is defined in
		// the [runtime/plugins/golang] package.  This argument enables buildFunc to
		// get other nodes' instances by name, using the method Get.
		//
		// # Example
		//
		//     func(n *golang.Namespace) (any, error) {
		// 	      var cart_db backend.NoSQLDatabase
		//     	  if err := n.Get("cart_db", &cart_db); err != nil {
		// 	    	  return nil, err
		//     	  }
		// 	      return cart.NewCartService(n.Context(), cart_db)
		//     }
		//
		// [runtime/plugins/golang]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/golang
		Declare(instanceName string, buildFuncSrc string) error

		// [DeclareConstructor] is a simpler version of [Declare] that does not require the
		// caller manually construct buildFunc source code.
		//
		// By invoking [DeclareConstructor], the caller specifies constructor, the func to use
		// to build name.
		//
		// The [NamespaceBuilder] will generate code that instantiates each node in args,
		// then passes those instances to constructor.
		DeclareConstructor(name string, constructor *gocode.Constructor, args []ir.IRNode) error

		// Specify nodes needed by this namespace that exist in a parent namespace
		// and will be passed as runtime arguments
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
