// Package address provides IR nodes to represent addressing, particularly between clients and servers.
//
// This plugin is primarily for use by other Blueprint plugins.  It is not expected that a wiring spec
// needs to directly call methods from this package.
//
// The main usage by other Blueprint plugins is the Define method, which will define an address that points
// to an IR node of a specified type.  Within a buildfunc, plugins can directly get the address by name,
// or use the helper methods Bind and Dial to get only the relevant configuration nodes.
//
// To implement addressing, several concerns are addressed:
//   - when a client node is instantiated it usually wants to access the server node in order to discover
//     the server's address and its interface.  But the client doesn't want to accidentally instantiate
//     the server in the wrong place.  An Address node adds one layer of indirection to prevent this.
//   - an address node takes care of configuration variables for binding and dialling the address.
package address

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

type (
	// IR metadata node representing an address.
	//
	// The main purpose of this node is to enable client nodes to link to server
	// nodes lazily without inadvertently instantiating the server nodes in the
	// wrong namespace.
	//
	// During the build process, the destination node of an address will be stored
	// on this node.  That enables clients, later, to call methods on the server
	// node, e.g. to get the interface that the server node exposes.
	//
	// The main implementation of this interface is [Address]
	Node interface {
		ir.IRNode
		ir.IRMetadata

		// Returns the server-side of an address if it has been instantiated; nil otherwise
		GetDestination() ir.IRNode

		// Sets the server-side of the address to be the provided node.
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

	// IR config node representing an address that a server should bind to.
	BindConfig struct {
		addressConfig
		PreferredPort uint16
	}

	// IR config node representing an address that a client should dial.
	DialConfig struct {
		addressConfig
	}
)

type (
	// The main implementation of the [Node] interface.
	//
	// In addition to storing the destination IR node, the address
	// also comes with two configuration IR nodes: a [BindConfig] that
	// is the bind address of the destination node, and a [DialConfig] that is the
	// address callers should dial.
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
