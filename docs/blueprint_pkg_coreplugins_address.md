---
title: blueprint/pkg/coreplugins/address
---
# blueprint/pkg/coreplugins/address
```go
package address // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
```
```go
Package address provides IR nodes to represent addressing, particularly between
clients and servers.
```
```go
This plugin is primarily for use by other Blueprint plugins. It is not expected
that a wiring spec needs to directly call methods from this package.
```
```go
The main usage by other Blueprint plugins is the Define method, which will
define an address that points to an IR node of a specified type. Within a
buildfunc, plugins can directly get the address by name, or use the helper
methods Bind and Dial to get only the relevant configuration nodes.
```
node in order to discover the server's address and its interface. But the
client doesn't want to accidentally instantiate the server in the wrong
place. An Address node adds one layer of indirection to prevent this.
dialling the address.
```go
To implement addressing, several concerns are addressed:
  - when a client node is instantiated it usually wants to access the server
  - an address node takes care of configuration variables for binding and
```
## FUNCTIONS

## func AssignPorts
```go
func AssignPorts(hostname string, nodes []ir.IRNode) error
```
AssignPorts is a helper method intended for use by namespace nodes when
they are compiling code and concrete ports must be assigned to BindConfig IR
nodes.

The provided nodes can be any IR nodes; this method will filter out only the
BindConfig nodes.

Some of the provided nodes might already be assigned to a particular port.
This method will not change those port assignments, though it will return an
error if two nodes are already pre-assigned to the same port.

Ports will be assigned either ascending from port 2000, or ascending from a
node's preferred port if a preference was specified.

After calling this method, any provided BindConfig IR nodes will have their
hostname and port set.

## func CheckPorts
```go
func CheckPorts(nodes []ir.IRNode) error
```
Returns an error if there are BindConfig nodes in the provided list that
haven't been allocated a port.

## func Define[ServerType
```go
func Define[ServerType ir.IRNode](spec wiring.WiringSpec, addressName string, pointsTo string, reachability any)
```
Defines an address called addressName whose server-side node has name
pointsTo.

This method is primarily intended for use by other Blueprint plugins.

The type parameter ServerType should correspond to the node type of
pointsTo.

Reachability of an address defines how far up the parent namespaces the
address should exist and be reachable. By default most addresses will want
to use ir.ApplicationNode as the reachability to indicate that the address
can be reached by any node anywhere in the application.

## func PointsTo
```go
func PointsTo(namespace wiring.Namespace, addressName string) (string, error)
```
Returns the value of pointsTo that was provided when addressName was
defined.

Used by the pointer plugin.

## func ResetPorts
```go
func ResetPorts(nodes []ir.IRNode)
```
Clears the hostname and port from any BindConfig node.

This is used by namespace nodes when performing address translation, e.g.
between ports within a container vs. external to a container.


## TYPES

The main implementation of the Node interface.
```go
type Address[ServerType ir.IRNode] struct {
	AddrName string
	Server   ServerType
	Bind     *BindConfig // Configuration value for the bind address
	Dial     *DialConfig // Configuration value for the dial address
}
```
In addition to storing the destination IR node, the address also comes with
two configuration IR nodes: a BindConfig that is the bind address of the
destination node, and a DialConfig that is the address callers should dial.

## func Bind[ServerType
```go
func Bind[ServerType ir.IRNode](namespace wiring.Namespace, addressName string) (*Address[ServerType], error)
```
Gets the BindConfig configuration node of addressName from the namespace.

This method is intended for use by other Blueprint plugins within their own
BuildFuncs.

This is a convenience method for use when only the dial address is needed.
It is equivalent to getting addressName directly from namespace and then
reading then [Address.Bind] field.

## func Dial[ServerType
```go
func Dial[ServerType ir.IRNode](namespace wiring.Namespace, addressName string) (*Address[ServerType], error)
```
Gets the DialConfig configuration node of addressName from the namespace.

This method is intended for use by other Blueprint plugins within their own
BuildFuncs.

This is a convenience method for use when only the dial address is needed.
It is equivalent to getting addressName directly from namespace and then
reading then [Address.Dial] field.

## func 
```go
func (addr *Address[ServerType]) GetDestination() ir.IRNode
```

## func 
```go
func (addr *Address[ServerType]) ImplementsAddressNode()
```

## func 
```go
func (addr *Address[ServerType]) ImplementsIRMetadata()
```

## func 
```go
func (addr *Address[ServerType]) Name() string
```

## func 
```go
func (addr *Address[ServerType]) SetDestination(node ir.IRNode) error
```

## func 
```go
func (addr *Address[ServerType]) String() string
```

IR config node representing an address that a server should bind to.
```go
type BindConfig struct {
	PreferredPort uint16
	// Has unexported fields.
}
```
## func 
```go
func (conf *BindConfig) HasValue() bool
```

## func 
```go
func (conf *BindConfig) ImplementsBindConfig()
```

## func 
```go
func (conf *BindConfig) ImplementsIRConfig()
```

## func 
```go
func (conf *BindConfig) Name() string
```

## func 
```go
func (conf *BindConfig) Optional() bool
```

## func 
```go
func (conf *BindConfig) String() string
```

## func 
```go
func (conf *BindConfig) Value() string
```

IR config node representing an address that a client should dial.
```go
type DialConfig struct {
	// Has unexported fields.
}
```
## func 
```go
func (conf *DialConfig) HasValue() bool
```

## func 
```go
func (conf *DialConfig) ImplementsDialConfig()
```

## func 
```go
func (conf *DialConfig) ImplementsIRConfig()
```

## func 
```go
func (conf *DialConfig) Name() string
```

## func 
```go
func (conf *DialConfig) Optional() bool
```

## func 
```go
func (conf *DialConfig) String() string
```

## func 
```go
func (conf *DialConfig) Value() string
```

```go
type Node interface {
	ir.IRNode
	ir.IRMetadata
```
```go
	// Returns the server-side of an address if it has been instantiated; nil otherwise
	GetDestination() ir.IRNode
```
```go
	// Sets the server-side of the address to be the provided node.
	SetDestination(ir.IRNode) error
```
IR metadata node representing an address.
```go
	ImplementsAddressNode()
}
```
The main purpose of this node is to enable client nodes to link to server
nodes lazily without inadvertently instantiating the server nodes in the
wrong namespace.

During the build process, the destination node of an address will be stored
on this node. That enables clients, later, to call methods on the server
node, e.g. to get the interface that the server node exposes.

The main implementation of this interface is Address


