package thrift

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
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
		var addr *GolangThriftServerAddress
		if err := namespace.Get(clientNext, &addr); err != nil {
			return nil, blueprint.Errorf("Thrift client %s expected %s to be an address, but encountered %s", thrift_client, clientNext, err)
		}
		return newGolangThriftClient(thrift_client, addr)
	})

	serverNext := ptr.AddDstModifier(wiring, thrift_server)

	wiring.Define(thrift_server, &GolangThriftServer{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *GolangThriftServerAddress
		if err := namespace.Get(thrift_addr, &addr); err != nil {
			return nil, blueprint.Errorf("Thrift server %s expected %s to be an address, but encountered %s", thrift_server, thrift_addr, err)
		}

		var wrapped golang.Service
		if err := namespace.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Thrift server %s expected %s to be a golang.Service, but encountered %s", thrift_server, serverNext, err)
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
