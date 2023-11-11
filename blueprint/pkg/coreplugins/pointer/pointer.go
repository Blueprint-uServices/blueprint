// Package pointer provides the core functionality for wrapping clients and servers
// with modifier nodes such as tracing, RPC, and many more.
//
// The functionality in this package is primarily intended for use by other Blueprint plugins.
//
// When a plugin declares an IR node, if that node represents something like a service
// whose interfaces can be further wrapped by other plugins, then in addition to the
// original IR node definition, the plugin should also declare a pointer to that IR node.
//
// Once a pointer is declared, other plugins can wrap the client or server side of the
// pointer.  Many plugins *only* apply to nodes that have pointers declared, and will
// get compilation errors for non-pointer nodes.
//
// Other Blueprint plugins apply client and server modifications by calling [AddSrcModifier]
// or [AddDstModifier].
//
// Internally the pointer keeps track of the modifier nodes that have been applied to
// the pointer.
//
// Typically there is an address node internally that separates the client side and server
// side of the pointer.  This is not implemented here directly -- a plugin should explicitly
// add an address node if it has separate client and server processes.  For example, a service
// will only add an address node when it is deployed over RPC.
//
// The server side of a pointer is usually instantiated lazily, because typically the server
// does not want to live in the same namespace as the client.
//
// The method [InstantiateDst] is used by other Blueprint plugins that wish to explicitly
// instantiate the server side of a pointer in a particular namespace.
package pointer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// A PointerDef provides methods for plugins to add client or server side modifiers
// to a pointer.
//
// Stored as metadata within a wiring spec.
type PointerDef struct {
	name         string
	srcHead      string
	srcModifiers []string
	srcTail      string
	dstHead      string
	dstModifiers []string
	dst          string
}

func (ptr PointerDef) String() string {
	b := strings.Builder{}
	b.WriteString("[")
	b.WriteString(strings.Join(ptr.srcModifiers, " -> "))
	b.WriteString("] -> [")
	b.WriteString(strings.Join(ptr.dstModifiers, " -> "))
	b.WriteString("]")
	return b.String()
}

// Creates a pointer called name that points to the specified node dst.  ptrType is the
// node type of dst.
//
// Any plugin that defines client and server nodes should typically declare a pointer to
// the server node.  This will provide some useful functionality:
//
// First, declaring a pointer will enable other plugins to apply client or server modifiers
// to the pointer.
//
// Second, pointers provide functionality for lazily instantiating the server side of
// the pointer if the server is not explicitly instantiated by the wiring spec.
func CreatePointer(spec wiring.WiringSpec, name string, ptrType any, dst string) *PointerDef {
	ptr := &PointerDef{}
	ptr.name = name
	ptr.srcModifiers = nil
	ptr.srcHead = name + ".src"
	ptr.srcTail = ptr.srcHead
	ptr.dstHead = dst
	ptr.dstModifiers = nil
	ptr.dst = dst

	spec.Alias(ptr.srcTail, ptr.dstHead)

	spec.Define(name, ptrType, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var node ir.IRNode
		if err := namespace.Get(ptr.srcHead, &node); err != nil {
			return nil, err
		}

		namespace.Defer(func() error {
			_, err := ptr.InstantiateDst(namespace)
			return err
		})

		return node, nil
	})

	spec.SetProperty(name, "ptr", ptr)

	return ptr
}

// Gets the PointerDef metadata for a pointer name that was defined using CreatePointer
func GetPointer(spec wiring.WiringSpec, name string) *PointerDef {
	var ptr *PointerDef
	spec.GetProperty(name, "ptr", &ptr)
	return ptr
}

// Appends a modifier node called modifierName to the client side modifiers of a pointer.
//
// Plugins use this method if they want to wrap the client side of a service, for example
// to add functionality like tracing, or to make calls over RPC.
//
// A pointer can have multiple modifiers applied to it.  They will be applied in the order
// that AddSrcModifier was called.
//
// The return value of AddSrcModifier is the name of the _next_ client side modifier.  This
// can be used within the BuildFunc of modifierName.
func (ptr *PointerDef) AddSrcModifier(spec wiring.WiringSpec, modifierName string) string {
	spec.Alias(ptr.srcTail, modifierName)
	ptr.srcTail = modifierName + ".ptr.src.next"
	spec.Alias(ptr.srcTail, ptr.dstHead)
	ptr.srcModifiers = append(ptr.srcModifiers, modifierName)

	return ptr.srcTail
}

// Appends a modifier node called modifierName to the server side modifiers of a pointer.
//
// Plugins use this method if they want to wrap the server side of a service, for example
// to add functionality like tracing, or to deploy a service with RPC.
//
// A pointer can have multiple modifiers applied to it.  They will be applied in the order
// that AddDstModifier was caleld.
//
// The return value of AddDstModifier is the name of the _previous_ server side modifier.  This
// can be used within the BuildFunc of modifierName.
func (ptr *PointerDef) AddDstModifier(spec wiring.WiringSpec, modifierName string) string {
	nextDst := ptr.dstHead
	ptr.dstHead = modifierName
	spec.Alias(ptr.srcTail, ptr.dstHead)
	ptr.dstModifiers = append([]string{ptr.dstHead}, ptr.dstModifiers...)
	return nextDst
}

// If any pointer modifiers are addresses, this will instantiate the server side of the addresses.
//
// This is primarily used by namespace plugins.
func (ptr *PointerDef) InstantiateDst(namespace wiring.Namespace) (ir.IRNode, error) {
	namespace.Info("Instantiating pointer %s.dst from namespace %s", ptr.name, namespace.Name())
	for _, modifier := range ptr.dstModifiers {
		var addr address.Node
		err := namespace.Get(modifier, &addr)

		// Want to find the final dstModifier that points to an address, then instantiate the address
		if err == nil {
			dstName, err := address.PointsTo(namespace, modifier)
			if err != nil {
				return nil, err
			}
			if addr.GetDestination() != nil {
				// Destination has already been instantiated, stop instantiating now
				namespace.Info("Destination %s of %s has already been instantiated", dstName, addr.Name())
				return nil, nil
			} else {
				namespace.Info("Instantiating %s of %s", dstName, addr.Name())
				var dst ir.IRNode
				if err := namespace.Instantiate(dstName, &dst); err != nil {
					return nil, err
				}
				err = addr.SetDestination(dst)
				if err != nil {
					return nil, err
				}
			}
		} else {
			namespace.Info("Skipping %v, not an address", modifier)
		}
	}

	var node ir.IRNode
	err := namespace.Get(ptr.dst, &node)
	return node, err
}
