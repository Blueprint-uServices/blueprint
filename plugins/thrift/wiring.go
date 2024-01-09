// Package thrift implements a Blueprint plugin that enables any Golang service to be deployed using a Thrift server.
//
// To use the plugin in a Blueprint wiring spec, import this package and use the [Deploy] method, i.e.
//
//	import "github.com/blueprint-uservices/blueprint/plugins/thrift"
//	thrift.Deploy(spec, "my_service")
//
// See the documentation for [Deploy] for more information about its behavior.
//
// The plugin implements thrift code generation, as well as generating a server-side handler
// and a client-side library that calls the server.
// This is implemented within the [thriftcodegen] pacakge.
//
// To use this plugin, the thrift compiler and version-matching go bindings are required to be installed on the machine that is compiling the Blueprint wiring spec.
// Installation instructions can be found: https://thrift.apache.org/download
package thrift

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Deploys `serviceName` as a Thrift server.
//
// Typically serviceName should be the name of a workflow service that was initially
// defined using [workflow.Define].
//
// Like many other modifiers, Thrift modifies the service at the golang level, by generating
// server-side handler code and a client-side library.
// However, Thrift should be the last golang-level modifier
// applied to a service, because thereafter communication between
// the client and server is no longer at the golang level, but at the network level.
//
// Deploying a service with Thrift increases the visibility of the service within the application.
// By default, any other service running in any other container or namespace can now contact this service.
func Deploy(spec wiring.WiringSpec, serviceName string) {
	// The nodes that we are defining
	thrift_client := serviceName + ".thrift_client"
	thrift_server := serviceName + ".thrift_server"
	thrift_addr := serviceName + ".thrift.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using Thrift as it is not a pointer")
		return
	}

	// Define the address that will be used by clients and the server
	address.Define[*golangThriftServer](spec, thrift_addr, thrift_server)

	// Add the client-side modifier
	//
	// The client-side modifier creates a Thrift client and dials the server address.
	// It assumes the next src modifier node will be a golangThriftServer address.
	clientNext := ptr.AddSrcModifier(spec, thrift_client)
	spec.Define(thrift_client, &golangThriftClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*golangThriftServer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("Thrift client %s expected %s to be an address, but encountered %s", thrift_client, clientNext, err)
		}
		return newGolangThriftClient(thrift_client, addr)
	})

	// Add the server-side modifier, which is an address that PointsTo the grpcServer
	serverNext := ptr.AddAddrModifier(spec, thrift_addr)
	spec.Define(thrift_server, &golangThriftServer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := namespace.Get(serverNext, &wrapped); err != nil {
			return nil, err
		}

		server, err := newGolangThriftServer(thrift_server, wrapped)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*golangThriftServer](namespace, thrift_addr, server, &server.Bind)
		return server, err
	})
}
