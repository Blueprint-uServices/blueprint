package address

import (
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

type (
	/*
		IRNode representing an address, used during the build process.
		Contains metadata about the address and the node it points to
	*/
	Node interface {
		blueprint.IRNode
		blueprint.IRConfig
		Name() string
		GetDestination() blueprint.IRNode
		SetDestination(blueprint.IRNode) error
		ImplementsAddressNode()
	}
)

type (
	/* Basic generic implementation of address.Node */
	Address[ServerType blueprint.IRNode] struct {
		AddrName string
		Server   ServerType
	}
)

/* The address of a server, used by a client */
type ServerAddress struct {
	blueprint.IRConfig
	Hostname string
	Port     string
}

/* The address that a server binds to */
type BindAddress struct {
	blueprint.IRConfig
	Hostname string
	Port     string
}

func (addr *Address[ServerType]) Name() string {
	return addr.AddrName
}

func (addr *Address[ServerType]) String() string {
	return addr.AddrName + " = ServerAddress()"
}

func (addr *Address[ServerType]) GetDestination() blueprint.IRNode {
	if reflect.ValueOf(addr.Server).IsNil() {
		return nil
	}
	return addr.Server
}

func (addr *Address[ServerType]) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(ServerType)
	if !isServer {
		return blueprint.Errorf("address %v points to invalid server type %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *Address[ServerType]) ImplementsAddressNode() {}
func (addr *Address[ServerType]) ImplementsIRConfig()    {}
