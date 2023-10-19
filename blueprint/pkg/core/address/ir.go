package address

import (
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

type (
	/*
		Metadata IRNode representing an address, used during the build process.
		Contains metadata about the address and the node it points to
	*/
	Node interface {
		blueprint.IRNode
		blueprint.IRMetadata
		Name() string
		GetDestination() blueprint.IRNode
		SetDestination(blueprint.IRNode) error
		ImplementsAddressNode()
	}

	/* Config IRNode representing an address */
	AddressConfig struct {
		blueprint.IRConfig
		Key      string
		Hostname string
		Port     string
	}
)

type (
	/* Basic generic implementation of address.Node */
	Address[ServerType blueprint.IRNode] struct {
		AddrName string
		Server   ServerType
	}
)

func (addr *Address[ServerType]) Name() string {
	return addr.AddrName
}

func (addr *Address[ServerType]) String() string {
	return addr.AddrName
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
func (addr *Address[ServerType]) ImplementsIRMetadata()  {}

func (conf *AddressConfig) Name() string {
	return conf.Key
}

func (conf *AddressConfig) String() string {
	return conf.Key + " = AddressConfig()"
}

func (conf *AddressConfig) ImplementsIRConfig() {}
