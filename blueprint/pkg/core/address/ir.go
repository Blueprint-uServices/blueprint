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

	/* A configuration parameter representing the address for a server to bind to */
	BindConfig struct {
		blueprint.IRConfig
		Key           string
		Interface     string
		Port          uint16
		PreferredPort uint16
	}

	/* A configuration parameter representing the address for a client to dial */
	DialConfig struct {
		blueprint.IRConfig
		Key      string
		Hostname string
		Port     uint16
	}
)

type (
	/* Basic generic implementation of address.Node */
	Address[ServerType blueprint.IRNode] struct {
		AddrName string
		Server   ServerType
		Bind     *BindConfig // Configuration value for the bind address
		Dial     *DialConfig // Configuration value for the dial address
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

func (conf *BindConfig) Name() string {
	return conf.Key
}

func (conf *BindConfig) String() string {
	return conf.Key + " = DialConfig()"
}

func (conf *BindConfig) ImplementsIRConfig() {}

func (conf *DialConfig) Name() string {
	return conf.Key
}

func (conf *DialConfig) String() string {
	return conf.Key + " = DialConfig()"
}

func (conf *DialConfig) ImplementsIRConfig() {}
