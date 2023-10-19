package xtrace

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments the service with an entry + exit point xtrace wrapper to generate xtrace compatible logs
func Instrument(wiring blueprint.WiringSpec, serviceName string) {
	DefineXTraceServer(wiring)
	clientWrapper := serviceName + ".client.xtrace"
	serverWrapper := serviceName + ".server.xtrace"
	xtrace_server := "xtrace_server"

	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using XTrace as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(wiring, clientWrapper)
	slog.Info("Next client is ", clientNext)

	wiring.Define(clientWrapper, &XtraceClientWrapper{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
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

	serverNext := ptr.AddDstModifier(wiring, serverWrapper)

	wiring.Define(serverWrapper, &XtraceServerWrapper{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
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

func DefineXTraceServer(wiring blueprint.WiringSpec) {
	xtrace_server := "xtrace_server"
	xtrace_addr := xtrace_server + ".addr"
	xtraceClient := xtrace_server + ".client"
	xtraceProc := xtrace_server + ".proc"
	xtraceDst := xtrace_server + ".dst"

	wiring.Define(xtraceProc, &XTraceServer{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*XTraceServer](ns, xtrace_addr)
		if err != nil {
			return nil, err
		}

		return newXTraceServer(xtraceProc, addr)
	})

	wiring.Alias(xtraceDst, xtraceProc)
	pointer.RequireUniqueness(wiring, xtraceDst, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, xtrace_server, &XTraceClient{}, xtraceDst)
	ptr := pointer.GetPointer(wiring, xtrace_server)
	ptr.AddDstModifier(wiring, xtrace_addr)

	clientNext := ptr.AddSrcModifier(wiring, xtraceClient)

	wiring.Define(xtraceClient, &XTraceClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*XTraceServer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newXTraceClient(xtraceClient, addr)
	})

	address.Define[*XTraceServer](wiring, xtrace_addr, xtraceProc, &blueprint.ApplicationNode{})
	ptr.AddDstModifier(wiring, xtrace_addr)
}
