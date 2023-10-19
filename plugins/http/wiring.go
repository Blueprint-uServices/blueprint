package http

import (
	"errors"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
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
	wiring.Define(httpClient, &GolangHttpClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *address.Address[*GolangHttpServer]
		if err := ns.Get(clientNext, &addr); err != nil {
			return nil, errors.New(fmt.Sprintf("HTTP client %s expected %s to be an address, but encountered %s", httpClient, clientNext, err))
		}
		return newGolangHttpClient(httpClient, addr)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(wiring, httpServer)

	// Define the server
	wiring.Define(httpServer, &GolangHttpServer{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *address.Address[*GolangHttpServer]
		if err := ns.Get(httpAddr, &addr); err != nil {
			return nil, errors.New(fmt.Sprintf("HTTP server %s expected %s to be an address, but encountered %s", httpServer, httpAddr, err))
		}

		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, errors.New(fmt.Sprintf("HTTP server %s expected %s to be a golang.Service, but encountered %s", httpServer, serverNext, err))
		}

		return newGolangHttpServer(httpServer, addr, wrapped)
	})

	// Define the address and add it to the pointer dst
	address.Define(wiring, httpAddr, httpServer, &blueprint.ApplicationNode{}, func(ns blueprint.Namespace) (address.Node, error) {
		addr := &address.Address[*GolangHttpServer]{
			AddrName: httpAddr,
			Server:   nil,
		}
		return addr, nil
	})
	ptr.AddDstModifier(wiring, httpAddr)
}
