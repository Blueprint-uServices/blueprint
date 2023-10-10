package xtrace

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type GolangXTraceAddress struct {
	address.Address
	service.ServiceNode
	AddrName string
	Server   *XTraceServer
}

func (addr *GolangXTraceAddress) Name() string {
	return addr.AddrName
}

func (addr *GolangXTraceAddress) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *GolangXTraceAddress) SetDestination(node blueprint.IRNode) error {
	xtrace_server, isServer := node.(*XTraceServer)
	if !isServer {
		return blueprint.Errorf("address %v should point to an XTraceServer but got %v", addr.AddrName, node)
	}
	addr.Server = xtrace_server
	return nil
}

func (addr *GolangXTraceAddress) String() string {
	return addr.AddrName + " = GolangXTraceAddress()"
}

func (addr *GolangXTraceAddress) ImplementsAddressNode() {}
