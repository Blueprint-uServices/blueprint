package grpc

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

// IRNode representing an address to a grpc server
type GolangServerAddress struct {
	address.Address
	service.ServiceNode
	AddrName string
	Server   *GolangServer
}

func (addr *GolangServerAddress) Name() string {
	return addr.AddrName
}

func (addr *GolangServerAddress) String() string {
	return addr.AddrName + " = GolangServerAddress()"
}

func (addr *GolangServerAddress) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *GolangServerAddress) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(*GolangServer)
	if !isServer {
		return fmt.Errorf("address %v should point to a Golang server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *GolangServerAddress) GetInterface() service.ServiceInterface {
	return addr.Server.GetInterface()
}

func (addr *GolangServerAddress) ImplementsAddressNode() {}
