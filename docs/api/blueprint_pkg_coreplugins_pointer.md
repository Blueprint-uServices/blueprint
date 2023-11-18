---
title: blueprint/pkg/coreplugins/pointer
---
# blueprint/pkg/coreplugins/pointer
```go
package pointer // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
```
```go
Package pointer provides the core functionality for wrapping clients and servers
with modifier nodes such as tracing, RPC, and many more.
```
```go
The functionality in this package is primarily intended for use by other
Blueprint plugins.
```
```go
When a plugin declares an IR node, if that node represents something like
a service whose interfaces can be further wrapped by other plugins, then in
addition to the original IR node definition, the plugin should also declare a
pointer to that IR node.
```
```go
Once a pointer is declared, other plugins can wrap the client or server side of
the pointer. Many plugins *only* apply to nodes that have pointers declared,
and will get compilation errors for non-pointer nodes.
```
```go
Other Blueprint plugins apply client and server modifications by calling
[AddSrcModifier] or [AddDstModifier].
```
```go
Internally the pointer keeps track of the modifier nodes that have been applied
to the pointer.
```
```go
Typically there is an address node internally that separates the client side and
server side of the pointer. This is not implemented here directly -- a plugin
should explicitly add an address node if it has separate client and server
processes. For example, a service will only add an address node when it is
deployed over RPC.
```
```go
The server side of a pointer is usually instantiated lazily, because typically
the server does not want to live in the same namespace as the client.
```
```go
The method [InstantiateDst] is used by other Blueprint plugins that wish to
explicitly instantiate the server side of a pointer in a particular namespace.
```
## FUNCTIONS

## func RequireUniqueness
```go
func RequireUniqueness(spec wiring.WiringSpec, alias string, visibility any)
```
A uniqueness check can be applied to any aliased node.

It requires that the specified node must be unique up to a certain
granularity.

This is independent of whether it can be addressed by any node within that
granularity.

The name argument should be an alias that this call will redefine.


## TYPES

A PointerDef provides methods for plugins to add client or server side
modifiers to a pointer.
```go
type PointerDef struct {
	// Has unexported fields.
}
```
Stored as metadata within a wiring spec.

## func CreatePointer
```go
func CreatePointer(spec wiring.WiringSpec, name string, ptrType any, dst string) *PointerDef
```
Creates a pointer called name that points to the specified node dst. ptrType
is the node type of dst.

Any plugin that defines client and server nodes should typically declare a
pointer to the server node. This will provide some useful functionality:

First, declaring a pointer will enable other plugins to apply client or
server modifiers to the pointer.

Second, pointers provide functionality for lazily instantiating the server
side of the pointer if the server is not explicitly instantiated by the
wiring spec.

## func GetPointer
```go
func GetPointer(spec wiring.WiringSpec, name string) *PointerDef
```
Gets the PointerDef metadata for a pointer name that was defined using
CreatePointer

## func 
```go
func (ptr *PointerDef) AddDstModifier(spec wiring.WiringSpec, modifierName string) string
```
Appends a modifier node called modifierName to the server side modifiers of
a pointer.

Plugins use this method if they want to wrap the server side of a service,
for example to add functionality like tracing, or to deploy a service with
RPC.

A pointer can have multiple modifiers applied to it. They will be applied in
the order that AddDstModifier was caleld.

The return value of AddDstModifier is the name of the _previous_ server side
modifier. This can be used within the BuildFunc of modifierName.

## func 
```go
func (ptr *PointerDef) AddSrcModifier(spec wiring.WiringSpec, modifierName string) string
```
Appends a modifier node called modifierName to the client side modifiers of
a pointer.

Plugins use this method if they want to wrap the client side of a service,
for example to add functionality like tracing, or to make calls over RPC.

A pointer can have multiple modifiers applied to it. They will be applied in
the order that AddSrcModifier was called.

The return value of AddSrcModifier is the name of the _next_ client side
modifier. This can be used within the BuildFunc of modifierName.

## func 
```go
func (ptr *PointerDef) InstantiateDst(namespace wiring.Namespace) (ir.IRNode, error)
```
If any pointer modifiers are addresses, this will instantiate the server
side of the addresses.

This is primarily used by namespace plugins.

## func 
```go
func (ptr PointerDef) String() string
```


