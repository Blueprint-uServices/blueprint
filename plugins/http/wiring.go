// Package http implements a Blueprint plugin that enables any Golang service to be deployed using a http server.

// To use the plugin in a Blueprint wiring spec, import this package and use the [Deploy] method, i.e.
//
//	import "github.com/blueprint-uservices/blueprint/plugins/http"
//	http.Deploy(spec, "my_service")
//
// See the documentation for [Deploy] for more information about its behavior.
//
// The plugin implements a server-side handler and client-side
// library that calls the server. This is implemented within the [httpcodegen] package.
package http

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

//Deploys `serviceName` as a HTTP server.

// Typcially serviceName should be the name of a workflow service that was initially defined using [workflow.Define].
//
// Like many other modifiers, HTTP modifier the service at the golang level, by generating
// server-side handler code and a client-side library. However, HTTP
// should be the last golang-level modifier applied to a service, because
// thereafter communication between the client and server
// is no longer at the golang level, but at the network level.
//
// Deploying a service with HTTP increases the visibility of the service within the application.
// By default, any other service running in any other container or namespace can now contact this service.
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

	// Define the address that will be used by clients and the server
	address.Define[*golangHttpServer](spec, httpAddr, httpServer)

	// Add the client-side modifier
	//
	// The client-side modifier creates an HTTP client and dials the server address.
	// It assumes that the next src modifier node will be a golangHttpServer address.
	clientNext := ptr.AddSrcModifier(spec, httpClient)
	spec.Define(httpClient, &GolangHttpClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*golangHttpServer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("HTTP client %s expected %s to be an address, but encountered %s", httpClient, clientNext, err)
		}
		return newGolangHttpClient(httpClient, addr)
	})

	// Add the server-side modifier, which is an address that PointsTo the grpcServer
	serverNext := ptr.AddAddrModifier(spec, httpAddr)
	spec.Define(httpServer, &golangHttpServer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("HTTP server %s expected %s to be a golang.Service, but encountered %s", httpServer, serverNext, err)
		}

		server, err := newGolangHttpServer(httpServer, wrapped)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*golangHttpServer](ns, httpAddr, server, &server.Bind)
		return server, err
	})
}
