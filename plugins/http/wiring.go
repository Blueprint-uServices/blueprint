package http

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"golang.org/x/exp/slog"
)

/*
Deploys `serviceName` as a HTTP server. This can only be done if `serviceName` is a pointer from Golang nodes to Golang nodes.

This call adds both src and dst side modifiers to `serviceName`. After this, the pointer will be from addr to addr and can no longer modified with golang nodes.
*/
func Deploy(wiring blueprint.WiringSpec, serviceName string) {
	// The nodes that we are defining
	httpClient := serviceName + ".http_client"
	httpServer := serviceName + ".http_server"
	httpAddr := serviceName + ".http.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using HTTP as it not a pointer")
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(wiring, httpClient)

	// Define the client wrapper
	wiring.Define(httpClient, &GolangHttpClient{}, func(scope blueprint.Namespace) (blueprint.IRNode, error) {
		server, err := scope.Get(clientNext)
		if err != nil {
			return nil, err
		}
		return newGolangHttpClient(httpClient, server)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(wiring, httpServer)

	// Define the server
	wiring.Define(httpServer, &GolangHttpServer{}, func(scope blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := scope.Get(httpAddr)
		if err != nil {
			return nil, err
		}

		wrapped, err := scope.Get(serverNext)
		if err != nil {
			return nil, err
		}

		return newGolangHttpServer(httpServer, addr, wrapped)
	})

	// Define the address and add it to the pointer dst
	address.Define(wiring, httpAddr, httpServer, &blueprint.ApplicationNode{}, func(scope blueprint.Namespace) (address.Address, error) {
		addr := &GolangHttpServerAddress{
			AddrName: httpAddr,
			Server:   nil,
		}
		return addr, nil
	})
	ptr.AddDstModifier(wiring, httpAddr)
}
