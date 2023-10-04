package thrift

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type GolangThriftServerAddress struct {
	address.Address
	service.ServiceNode
	AddrName string
	Server   *GolangThriftServer
}

func (addr *GolangThriftServerAddress) Name() string {
	return addr.AddrName
}

func (addr *GolangThriftServerAddress) String() string {
	return addr.AddrName + " = GolangThriftServerAddress()"
}

func (addr *GolangThriftServerAddress) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *GolangThriftServerAddress) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(*GolangThriftServer)
	if !isServer {
		return blueprint.Errorf("address %v should point to a Golang server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *GolangThriftServerAddress) GetInterface() service.ServiceInterface {
	return addr.Server.GetInterface()
}

func (addr *GolangThriftServerAddress) ImplementsAddressNode() {}
