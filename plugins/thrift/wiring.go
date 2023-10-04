package thrift

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"golang.org/x/exp/slog"
)

func Deploy(wiring blueprint.WiringSpec, serviceName string) {
	thrift_client := serviceName + ".thrift_client"
	thrift_server := serviceName + ".thrift_server"
	thrift_addr := serviceName + ".thrift.addr"

	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using Thrift as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(wiring, thrift_client)

	wiring.Define(thrift_client, &GolangThriftClient{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		server, err := namespace.Get(clientNext)
		if err != nil {
			return nil, err
		}

		return newGolangThriftClient(thrift_client, server)
	})

	serverNext := ptr.AddDstModifier(wiring, thrift_server)

	wiring.Define(thrift_server, &GolangThriftServer{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := namespace.Get(thrift_addr)
		if err != nil {
			return nil, err
		}

		wrapped, err := namespace.Get(serverNext)
		if err != nil {
			return nil, err
		}

		return newGolangThriftServer(thrift_server, addr, wrapped)
	})

	address.Define(wiring, thrift_addr, thrift_server, &blueprint.ApplicationNode{}, func(namespace blueprint.Namespace) (address.Address, error) {
		addr := &GolangThriftServerAddress{
			AddrName: thrift_addr,
			Server:   nil,
		}
		return addr, nil
	})

	ptr.AddDstModifier(wiring, thrift_addr)
}
