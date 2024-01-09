// Package xtrace provides two plugins:
// (i)  a plugin to generate and include an xtrace instance in a Blueprint application.
// (ii) provides a modifier plugin to wrap the service with an XTrace wrapper to generate XTrace compatible traces/logs.
//
// The package provides a built-in xtrace container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications must use a backend.XTracer (runtime/core/backend) as the interface in the workflow.
package xtrace

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/blueprint"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/wiring"
	"github.com/Blueprint-uServices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

var default_xtrace_server_name = "xtrace_server"

// Instruments the service with an entry + exit point xtrace wrapper to generate xtrace compatible logs.
// Usage:
//
//	Instrument(spec, "serviceA")
func Instrument(spec wiring.WiringSpec, serviceName string) {
	xtraceServer := DefineXTraceServerContainer(spec, default_xtrace_server_name)
	clientWrapper := serviceName + ".client.xtrace"
	serverWrapper := serviceName + ".server.xtrace"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using XTrace as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)
	spec.Define(clientWrapper, &XtraceClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("XTrace client %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		var xtraceClient *XTraceClient
		err := ns.Get(xtraceServer, &xtraceClient)
		if err != nil {
			return nil, err
		}

		return newXtraceClientWrapper(clientWrapper, wrapped, xtraceClient)
	})

	serverNext := ptr.AddDstModifier(spec, serverWrapper)
	spec.Define(serverWrapper, &XtraceServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("XTrace server %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, wrapped)
		}

		var xtraceClient *XTraceClient
		err := ns.Get(xtraceServer, &xtraceClient)
		if err != nil {
			return nil, err
		}

		return newXtraceServerWrapper(serverWrapper, wrapped, xtraceClient)
	})
}

// Generates the IRNodes for a xtrace docker container that uses the latest xtrace image
// and the clients needed by the generated application to communicate with the server.
//
// The generated container has the name `serviceName`.
func DefineXTraceServerContainer(spec wiring.WiringSpec, serverName string) string {
	// The nodes that we are defining
	xtraceAddr := serverName + ".addr"
	xtraceClient := serverName + ".client"
	xtraceCtr := serverName + ".ctr"

	// Define the X-Trace server container
	spec.Define(xtraceCtr, &XTraceServerContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		xtrace, err := newXTraceServerContainer(xtraceCtr)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*XTraceServerContainer](ns, xtraceAddr, xtrace, &xtrace.BindAddr)
		return xtrace, err
	})

	// Create a pointer to the server
	ptr := pointer.CreatePointer[*XTraceClient](spec, serverName, xtraceCtr)

	// Define the address that points to the X-Trace collector
	address.Define[*XTraceServerContainer](spec, xtraceAddr, xtraceCtr)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, xtraceAddr)

	// Define the X-Trace client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, xtraceClient)
	spec.Define(xtraceClient, &XTraceClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*XTraceServerContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newXTraceClient(xtraceClient, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the X-Trace server should do so through the pointer
	return serverName
}
