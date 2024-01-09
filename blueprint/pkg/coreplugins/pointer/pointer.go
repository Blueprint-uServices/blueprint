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

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// A PointerDef provides methods for plugins to add client or server side modifiers
// to a pointer.
//
// Stored as metadata within a wiring spec.
type PointerDef struct {
	name string

	srcHead      string
	srcModifiers []string
	srcTail      string

	// The node that is the interface from src to dst.  Typically this is an address
	interfaceNode string

	dstHead      string
	dstModifiers []string
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

// Additional options that can be specified when creating a pointer.
// If not specified, defaults are used.
type PointerOpts struct {
	// If specified, applies [RequireUniqueness] to the pointer destination
	// before creating the pointer.  Set to nil to disable.  Defaults to
	// &ir.ApplicationNode{}
	RequireUniqueness any
}

// Additional options that can be specified when adding a modifier to a pointer.
// If not specified, defaults are used.
type ModifierOpts struct {
	// Defaults to true.  When true, the pointer's interface node is updated
	IsInterfaceNode bool
}

var defaultPointerOpts = PointerOpts{
	RequireUniqueness: &ir.ApplicationNode{},
}

var defaultModifierOpts = ModifierOpts{
	IsInterfaceNode: true,
}

// Creates a pointer called name that points to the specified node dst.
// Type parameter [SrcNodeType] is the nodeType of the client side of the pointer.
//
// Any plugin that defines client and server nodes should typically declare a pointer to
// the server node.  Declaring a pointer will enable other plugins to apply client or
// server modifiers to the pointer.  Additionally, pointers will automatically instantiate
// the server side of the pointer when using addresses
//
// Additional pointer options can be specified by providing optional PointerOpts.
func CreatePointer[SrcNodeType any](spec wiring.WiringSpec, name string, dst string, options ...PointerOpts) *PointerDef {
	opts := defaultPointerOpts
	if len(options) > 0 {
		opts = options[0]
	}

	if opts.RequireUniqueness != nil {
		dstName := name + ".dst"
		spec.Alias(dstName, dst)
		RequireUniqueness(spec, dstName, opts.RequireUniqueness)
		dst = dstName
	}

	ptr := &PointerDef{}
	ptr.name = name
	ptr.srcModifiers = nil
	ptr.srcHead = name + ".src"
	ptr.srcTail = ptr.srcHead
	ptr.interfaceNode = dst
	ptr.dstHead = dst
	ptr.dstModifiers = []string{dst}

	spec.Alias(ptr.srcTail, ptr.interfaceNode)

	var ptrType SrcNodeType
	spec.Define(name, ptrType, func(namespace wiring.Namespace) (ir.IRNode, error) {
		// This is the lazy implicit instantiation of the Dst side of the pointer, if
		// it hasn't explicitly been instantiated somewhere in the wiring spec.
		namespace.Defer(func() error {
			return ptr.InstantiateDst(namespace)
		}, wiring.DeferOpts{Front: false})

		var node ir.IRNode
		err := namespace.Get(ptr.srcHead, &node)
		return node, err
	})

	spec.SetProperty(name, "ptr", ptr)

	return ptr
}

// Gets the PointerDef metadata for a pointer name that was defined using [CreatePointer]
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
	spec.Alias(ptr.srcTail, ptr.interfaceNode)
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
func (ptr *PointerDef) AddDstModifier(spec wiring.WiringSpec, modifierName string, options ...ModifierOpts) string {
	opts := defaultModifierOpts
	if len(options) > 0 {
		opts = options[0]
	}
	nextDst := ptr.dstHead
	ptr.dstHead = modifierName
	if opts.IsInterfaceNode {
		ptr.interfaceNode = ptr.dstHead
		spec.Alias(ptr.srcTail, ptr.interfaceNode)
	}
	ptr.dstModifiers = append([]string{ptr.dstHead}, ptr.dstModifiers...)
	return nextDst
}

// AddAddrModifier is a special case of AddDstModifier where the modifier is an address node.
//
// It immediately instantiates the address, and returns it.  It defers instantiation of the
// server side of the address.
//
// The return value of AddAddrModifier is the name of the _previous_ server side modifier.  This
// can be used within the BuildFunc of the destination (PointsTo) of addrName
func (ptr *PointerDef) AddAddrModifier(spec wiring.WiringSpec, addrName string) string {
	// Get the address metadata
	def := address.GetAddress(spec, addrName)
	if def == nil {
		return ""
	}

	// Add a modifier to instantiate address PointsTo
	nextDst := ptr.AddDstModifier(spec, def.PointsTo)

	// Set the pointer interface to be the address, rather than the node
	ptr.interfaceNode = addrName
	spec.Alias(ptr.srcTail, ptr.interfaceNode)

	return nextDst
}

// If any pointer modifiers are addresses, this will instantiate the server side of the addresses.
//
// This is primarily used by namespace plugins.
func (ptr *PointerDef) InstantiateDst(namespace wiring.Namespace) error {
	namespace.Info("Instantiating pointer %s.dst from namespace %s", ptr.name, namespace.Name())

	// Instantiating dst starts from the interface node between src and dst
	var node ir.IRNode
	err := namespace.Get(ptr.interfaceNode, &node)
	if err != nil {
		return err
	}

	// We only have to handle the case where the interface between src and dst is an addr, which means
	// the dst side of the addr might not yet be instantiated.  If it's not an addr, then either the client,
	// or the above call to namespace.Get, will have triggered dst instantiation
	addr, isAddr := node.(address.Node)
	if !isAddr {
		return nil
	}

	// The addr destination might already have been explicitly instantiated, or instantiated by a different client
	if addr.GetDestination() != nil {
		return nil
	}

	// Getting the first dst modifier should cause all of the dst side of the pointer to be instantiated
	var dst ir.IRNode
	if namespace.Instantiate(ptr.dstModifiers[0], &dst) != nil {
		return err
	}

	// Currently we don't support multiple addresses within a pointer; so getting the first dst modifier
	// should cause a cascade where the addr destination gets set.  Error out if not.
	if addr.GetDestination() == nil {
		return namespace.Error("Attempted to instantiate the server-side of address %v starting with %v but the server failed to instantiate", addr.Name(), ptr.dstModifiers[0])
	}
	return nil
}
