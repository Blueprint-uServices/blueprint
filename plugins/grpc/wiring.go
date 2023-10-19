package grpc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Deploys `serviceName` as a GRPC server.  This can only be done if `serviceName` is a
pointer from Golang nodes to Golang nodes.

This call adds both src and dst-side modifiers to `serviceName`.  After this, the
pointer will be from addr to addr and can no longer be modified with golang nodes.
*/
func Deploy(wiring blueprint.WiringSpec, serviceName string) {
	// The nodes that we are defining
	grpcClient := serviceName + ".grpc_client"
	grpcServer := serviceName + ".grpc_server"
	grpcAddr := serviceName + ".grpc.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using GRPC as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(wiring, grpcClient)

	// Define the client wrapper
	wiring.Define(grpcClient, &GolangClient{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*GolangServer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("GRPC client %s expected %s to be an address, but encountered %s", grpcClient, clientNext, err)
		}

		return newGolangClient(grpcClient, addr)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(wiring, grpcServer)

	// Define the server
	wiring.Define(grpcServer, &GolangServer{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*GolangServer](namespace, grpcAddr)
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
	address.Define[*GolangServer](wiring, grpcAddr, grpcServer, &blueprint.ApplicationNode{})
	ptr.AddDstModifier(wiring, grpcAddr)

}
