package http

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Deploys `serviceName` as a HTTP server. This can only be done if `serviceName` is a pointer from Golang nodes to Golang nodes.

This call adds both src and dst side modifiers to `serviceName`. After this, the pointer will be from addr to addr and can no longer modified with golang nodes.
*/
func Deploy(spec wiring.WiringSpec, serviceName string) {
	// The nodes that we are defining
	httpClient := serviceName + ".http_client"
	httpServer := serviceName + ".http_server"
	httpAddr := serviceName + ".http.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using HTTP as it not a pointer")
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(spec, httpClient)

	// Define the client wrapper
	spec.Define(httpClient, &GolangHttpClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*GolangHttpServer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("HTTP client %s expected %s to be an address, but encountered %s", httpClient, clientNext, err)
		}
		return newGolangHttpClient(httpClient, addr)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(spec, httpServer)

	// Define the server
	spec.Define(httpServer, &GolangHttpServer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*GolangHttpServer](ns, httpAddr)
		if err != nil {
			return nil, blueprint.Errorf("HTTP server %s expected %s to be an address, but encountered %s", httpServer, httpAddr, err)
		}

		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("HTTP server %s expected %s to be a golang.Service, but encountered %s", httpServer, serverNext, err)
		}

		return newGolangHttpServer(httpServer, addr, wrapped)
	})

	// Define the address and add it to the pointer dst
	address.Define[*GolangHttpServer](spec, httpAddr, httpServer, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, httpAddr)
}
