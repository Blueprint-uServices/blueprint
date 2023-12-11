// Package xtrace provides two plugins:
// (i)  a plugin to generate and include an xtrace instance in a Blueprint application.
// (ii) provides a modifier plugin to wrap the service with an XTrace wrapper to generate XTrace compatible traces/logs.
//
// The package provides a built-in xtrace container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications muse use a backend.XTracer (runtime/core/backend) as the interface in the workflow.
package xtrace

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments the service with an entry + exit point xtrace wrapper to generate xtrace compatible logs.
// Usage:
//   Instrument(spec, "serviceA")
func Instrument(spec wiring.WiringSpec, serviceName string) {
	DefineXTraceServerContainer(spec)
	clientWrapper := serviceName + ".client.xtrace"
	serverWrapper := serviceName + ".server.xtrace"
	xtrace_server := "xtrace_server"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using XTrace as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)
	slog.Info("Next client is ", clientNext)

	spec.Define(clientWrapper, &XtraceClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("XTrace client %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		var xtraceClient *XTraceClient
		err := ns.Get(xtrace_server, &xtraceClient)
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
		err := ns.Get(xtrace_server, &xtraceClient)
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
func DefineXTraceServerContainer(spec wiring.WiringSpec) {
	xtrace_server := "xtrace_server"
	xtrace_addr := xtrace_server + ".addr"
	xtraceClient := xtrace_server + ".client"
	xtraceProc := xtrace_server + ".proc"
	xtraceDst := xtrace_server + ".dst"

	spec.Define(xtraceProc, &XTraceServerContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*XTraceServerContainer](ns, xtrace_addr)
		if err != nil {
			return nil, err
		}

		return newXTraceServerContainer(xtraceProc, addr.Bind)
	})

	spec.Alias(xtraceDst, xtraceProc)
	pointer.RequireUniqueness(spec, xtraceDst, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, xtrace_server, &XTraceClient{}, xtraceDst)
	ptr := pointer.GetPointer(spec, xtrace_server)
	ptr.AddDstModifier(spec, xtrace_addr)

	clientNext := ptr.AddSrcModifier(spec, xtraceClient)

	spec.Define(xtraceClient, &XTraceClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*XTraceServerContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newXTraceClient(xtraceClient, addr.Dial)
	})

	address.Define[*XTraceServerContainer](spec, xtrace_addr, xtraceProc, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, xtrace_addr)
}
