package address

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

type (
	/*
		Metadata IRNode representing an address, used during the build process.
		Contains metadata about the address and the node it points to
	*/
	Node interface {
		ir.IRNode
		ir.IRMetadata
		Name() string
		GetDestination() ir.IRNode
		SetDestination(ir.IRNode) error
		ImplementsAddressNode()
	}

	addressConfig struct {
		ir.IRConfig
		AddressName string // The name of the address metadata node
		Key         string
		Hostname    string
		Port        uint16
	}

	/* A configuration parameter representing the address for a server to bind to */
	BindConfig struct {
		addressConfig
		PreferredPort uint16
	}

	/* A configuration parameter representing the address for a client to dial */
	DialConfig struct {
		addressConfig
	}
)

type (
	/* Basic generic implementation of address.Node */
	Address[ServerType ir.IRNode] struct {
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

func (addr *Address[ServerType]) GetDestination() ir.IRNode {
	if reflect.ValueOf(addr.Server).IsNil() {
		return nil
	}
	return addr.Server
}

func (addr *Address[ServerType]) SetDestination(node ir.IRNode) error {
	server, isServer := node.(ServerType)
	if !isServer {
		return blueprint.Errorf("address %v points to invalid server type %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *Address[ServerType]) ImplementsAddressNode() {}
func (addr *Address[ServerType]) ImplementsIRMetadata()  {}

func (conf *addressConfig) Name() string {
	return conf.Key
}

func (conf *addressConfig) String() string {
	return conf.Key + " = AddressConfig()"
}

func (conf *addressConfig) Optional() bool {
	return false
}

func (conf *addressConfig) HasValue() bool {
	return conf.Hostname != "" && conf.Port != 0
}

func (conf *addressConfig) Value() string {
	return fmt.Sprintf("%v:%v", conf.Hostname, conf.Port)
}

func (conf *addressConfig) ImplementsIRConfig() {}
func (conf *BindConfig) ImplementsBindConfig()  {}
func (conf *DialConfig) ImplementsDialConfig()  {}
