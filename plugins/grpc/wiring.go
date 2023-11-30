// Package grpc implements a Blueprint plugin that enables any Golang service to
// be deployed using a gRPC server.
//
// To use the plugin in a Blueprint wiring spec, import this package and use the [Deploy]
// method, ie.
//
//	import "gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
//	grpc.Deploy(spec, "my_service")
//
// See the documentation for [Deploy] for more information about its behavior.
//
// The plugin implements gRPC code generation, as well as generating a server-side
// handler and client-side library that calls the server.  This is implemented within
// the [grpccodegen] package.
//
// To use this plugin requires the protocol buffers and grpc compilers are installed
// on the machine that is compiling the Blueprint wiring spec.  Installation instructions
// can be found on the [gRPC Quick Start].
//
// [gRPC quick Start]: https://grpc.io/docs/languages/go/quickstart/
package grpc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Deploys `serviceName` as a GRPC server.
//
// Typically serviceName should be the name of a workflow service that was initially
// defined using [workflow.Define].
//
// Like many other modifiers, GRPC modifies the service at the golang level, by generating
// server-side handler code and a client-side library.  However, GRPC should be the last
// golang-level modifier applied to a service, because thereafter communication between
// the client and server is no longer at the golang level, but at the network level.
//
// Deploying a service with GRPC increases the visibility of the service within the application.
// By default, any other service running in any other container or namespace can now contact
// this service.
func Deploy(spec wiring.WiringSpec, serviceName string) {
	// The nodes that we are defining
	grpcClient := serviceName + ".grpc_client"
	grpcServer := serviceName + ".grpc_server"
	grpcAddr := serviceName + ".grpc.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using GRPC as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(spec, grpcClient)

	// Define the client wrapper
	spec.Define(grpcClient, &golangClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*golangServer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("GRPC client %s expected %s to be an address, but encountered %s", grpcClient, clientNext, err)
		}

		return newGolangClient(grpcClient, addr)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(spec, grpcServer)

	// Define the server
	spec.Define(grpcServer, &golangServer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*golangServer](namespace, grpcAddr)
		if err != nil {
			return nil, blueprint.Errorf("GRPC server %s expected %s to be an address, but encountered %s", grpcServer, grpcAddr, err)
		}

		var wrapped golang.Service
		if err := namespace.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GRPC server %s expected %s to be a golang.Service, but encountered %s", grpcServer, serverNext, err)
		}

		return newGolangServer(grpcServer, addr, wrapped)
	})

	// Define the address and add it to the pointer dst
	address.Define[*golangServer](spec, grpcAddr, grpcServer, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, grpcAddr)

}
