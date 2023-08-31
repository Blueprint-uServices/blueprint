package golang

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow/parser"
)

/*
The golang plugin extends Blueprint's IR as follows:

It defines the following IR interfaces:

 - golang.Node is the base interface for any node that lives within a golang process
 - golang.Service is a golang node that has methods that can be directly called by other golang nodes

The golang plugin also defines the following new IR node implementations:

 - golang.Process is a node that represents a runnable Golang process.  It can contain any number of
   other golang.Node IRNodes.  When it's compiled, the golang.Process will generate a go module with
   a runnable main method that instantiates and initializes the contained go nodes.  To achieve this,
   the golang.Process also collects module dependencies from its contained nodes.

To support golang code generation, the following IR interfaces are provided, as defined in ir_codegen.go.
The golang.Process depends on these interfaces for collecting and packaging code, however, usage of these interfaces
is not intended to be private to just the golang.Process plugin.  Other plugins are permitted to
use these interfaces.

 - golang.Instantiable is for golang nodes that can generate instantiation source code snippets
 - golang.RequiresPackage is for golang nodes that generate source files and have module dependencies
 - golang.ProvidesModule is for golang nodes that generate or otherwise provide the full source code of modules
*/

/*
golang.Node is the base IRNode interface that should be implemented by any IRNode that
wishes to exist within a Golang namespace.
*/
type Node interface {
	blueprint.IRNode
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
type Service interface {
	Node
	Instantiable // Services must implement the Instantiable interface in order to create instances
	service.ServiceNode
	ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
}

/*
// Representation of a golang service interface, which extends the service.Service interface
// to include module and package info for all method arguments, and constructor info
// */
// type GolangServiceInterface struct {
// 	service.ServiceInterface
// 	ServiceName string

// 	MethodImpls []*GolangMethod
// }

// type ServiceInterface interface {
// 	Name() string
// 	Methods() []MethodSignature
// }

// type MethodSignature interface {
// 	Name() string
// 	Arguments() []Variable
// 	Returns() []Variable
// }

// type Variable interface {
// 	Name() string
// 	Type() string // a "well-known" type
// }

// Code location and interfaces of a service
type GolangServiceDetails struct {
	Interface        service.ServiceInterface         // The interface that is implemented
	InterfacePackage *parser.PackageInfo              // The package containing the constructor method
	ImplName         string                           // The type name of the implementing struct
	ImplConstructor  service.ServiceMethodDeclaration // The constructor method for the implementing struct
	ImplPackage      *parser.PackageInfo              // The package containing the constructor method
}

func (d GolangServiceDetails) String() string {
	var b strings.Builder
	b.WriteString("import \"" + d.InterfacePackage.ImportName + "\"\n")
	b.WriteString("var service " + d.InterfacePackage.ShortName + "." + d.Interface.Name + "\n")
	b.WriteString("service = " + d.ImplConstructor.Name)
	var constructorArgs []string
	for _, arg := range d.ImplConstructor.Args {
		constructorArgs = append(constructorArgs, arg.Name)
	}
	b.WriteString("(")
	b.WriteString(strings.Join(constructorArgs, ", "))
	b.WriteString(")")

	return b.String()
}
