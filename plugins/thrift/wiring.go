package thrift

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

func Deploy(spec wiring.WiringSpec, serviceName string) {
	thrift_client := serviceName + ".thrift_client"
	thrift_server := serviceName + ".thrift_server"
	thrift_addr := serviceName + ".thrift.addr"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using Thrift as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, thrift_client)

	spec.Define(thrift_client, &GolangThriftClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*GolangThriftServer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("Thrift client %s expected %s to be an address, but encountered %s", thrift_client, clientNext, err)
		}
		return newGolangThriftClient(thrift_client, addr)
	})

	serverNext := ptr.AddDstModifier(spec, thrift_server)

	spec.Define(thrift_server, &GolangThriftServer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*GolangThriftServer](namespace, thrift_addr)
		if err != nil {
			return nil, blueprint.Errorf("Thrift server %s expected %s to be an address, but encountered %s", thrift_server, thrift_addr, err)
		}

		var wrapped golang.Service
		if err := namespace.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Thrift server %s expected %s to be a golang.Service, but encountered %s", thrift_server, serverNext, err)
		}

		return newGolangThriftServer(thrift_server, addr, wrapped)
	})

	address.Define[*GolangThriftServer](spec, thrift_addr, thrift_server, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, thrift_addr)
}
