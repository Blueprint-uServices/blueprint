package grpc

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// IRNode representing an address to a grpc server
type GolangServerAddress struct {
	address.Address
	AddrName string
	Server   *GolangServer
}

// IRNode representing a Golang server that wraps a golang service
type GolangServer struct {
	golang.Node

	InstanceName string
	Addr         *GolangServerAddress
	Wrapped      golang.Service
}

func newGolangServer(name string, serverAddr blueprint.IRNode, wrapped blueprint.IRNode) (*GolangServer, error) {
	addr, is_addr := serverAddr.(*GolangServerAddress)
	if !is_addr {
		return nil, fmt.Errorf("GRPC server %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, fmt.Errorf("GRPC server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &GolangServer{}
	node.InstanceName = name
	node.Addr = addr
	node.Wrapped = service
	return node, nil
}

func (n *GolangServer) String() string {
	return n.InstanceName + " = GRPCServer(" + n.Wrapped.Name() + ", " + n.Addr.Name() + ")"
}

func (n *GolangServer) Name() string {
	return n.InstanceName
}

func (node *GolangServer) AddInstantiation(builder golang.DICodeBuilder) error {
	// TODO
	return nil
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
	fmt.Printf("Setting destination of %v to %v\n", addr.AddrName, node.Name())
	server, isServer := node.(*GolangServer)
	if !isServer {
		return fmt.Errorf("address %v should point to a Golang server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (node *GolangServer) ImplementsGolangNode()         {}
func (addr *GolangServerAddress) ImplementsAddressNode() {}
