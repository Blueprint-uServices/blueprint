package http

import (
	"errors"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

// IRNode representing an address to a http web server
type GolangHttpServerAddress struct {
	address.Address
	service.ServiceNode
	AddrName string
	Server   *GolangHttpServer
}

func (addr *GolangHttpServerAddress) Name() string {
	return addr.AddrName
}

func (addr *GolangHttpServerAddress) String() string {
	return addr.AddrName + " = GolangHttpServerAddress()"
}

func (addr *GolangHttpServerAddress) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *GolangHttpServerAddress) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(*GolangHttpServer)
	if !isServer {
		return errors.New(fmt.Sprintf("address %v should point to a Golang Http server but got %v", addr.AddrName, node))
	}
	addr.Server = server
	return nil
}

func (addr *GolangHttpServerAddress) GetInterface() service.ServiceInterface {
	return addr.Server.GetInterface()
}

func (addr *GolangHttpServerAddress) ImplementsAddressNode() {}
