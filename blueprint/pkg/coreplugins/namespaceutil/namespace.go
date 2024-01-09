// Package namespaceutil provides helper functionality for IRNodes that implement namespaces.
//
// Although the [wiring] package defines and implements the [wiring.Namespace] interface that
// is a central component of building Blueprint's IR, the methods of the namespace package are
// primarily intended to help Blueprint plugins in constructing IRNodes that correspond to
// wiring namespaces.
//
// A namespace node is an IRNode that can contain arbitrary nodes of a particular type.  For example,
// a Golang process node is a namespace node that can contain any instances of Golang services.
//
// Some plugins introduce new kinds of namespace nodes.  For example the Golang process node is introduced
// by the goproc plugin.  The functionality provided in the namespace package is intended to aid
// plugins that want to introduce new kinds of plugin.
//
// This package simplifies the task of adding child nodes to namespaces, with the [AddNodeTo] function.
// When used in conjunction with [InstantiateNamespace], child nodes will be automatically instantiated
// when a namespace node is instantiated.
//
// This package also takes care of adding modifiers to pointers to build the server side of pointers
// inside the correct namespaces
package namespaceutil

import (
	"fmt"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// An IRNode that also implements the [wiring.NamespaceHandler] interface,
// so that it can be directly used with the convenience methods defined in the [pointer] package.
type IRNamespace interface {
	ir.IRNode
	wiring.NamespaceHandler
}

var prop_CHILDREN = "Children"

// If a Blueprint plugin derives a namespace (e.g. a Process namespace that contains Golang nodes)
// then the plugin can use this method.
// When namespace gets instantiated, it will build child.  If child is a pointer, then
// the pointer is also modified, so that when child is instantiated, it is done so within namespace.
// The type parameter NamespaceNodeType is the namespace node type, e.g. Process
func AddNodeTo[NamespaceNodeType any](spec wiring.WiringSpec, namespaceName string, childName string) {
	ptr := pointer.GetPointer(spec, childName)
	if ptr == nil {
		// Not a pointer, no special handling needed
		spec.AddProperty(namespaceName, prop_CHILDREN, childName)
		return
	}

	// Add a modifier to child that enters the namespace before instantiating the child
	// Unlike most modifiers, we don't want this modifier to change which node the client side of the
	// pointer receives
	modifierName := fmt.Sprintf("%v.%v", childName, namespaceName)
	ptrNext := ptr.AddDstModifier(spec, modifierName, pointer.ModifierOpts{IsInterfaceNode: false})

	// We are proxying the ptrNext node, so must provide wiring spec options
	ptrNextDef := spec.GetDef(ptrNext)
	opts := wiring.WiringOpts{ProxyNode: true}
	if ptrNextDef != nil {
		opts.ReturnType = ptrNextDef.Options.ReturnType
	}

	// The modifier gets the namespace (creating if it doesn't exist) then immediately gets the next
	// modifier node from within the namespace
	var nodeType NamespaceNodeType
	spec.Define(modifierName, &nodeType, func(parentNamespace wiring.Namespace) (ir.IRNode, error) {
		// Namespace node must exist so that namespace exists
		var namespaceNode ir.IRNode
		if err := parentNamespace.Get(namespaceName, &namespaceNode); err != nil {
			return nil, err
		}

		// Get the namespace
		namespace, err := parentNamespace.GetNamespace(namespaceNode.Name())
		if err != nil {
			return nil, err
		}

		// Continue building the pointer from inside the namespace
		var ptrNextNode ir.IRNode
		err = namespace.Instantiate(ptrNext, &ptrNextNode)
		return ptrNextNode, err
	}, opts)

	// The namespace also instantiates the modifier
	spec.AddProperty(namespaceName, prop_CHILDREN, ptrNext)
}

// Used in conjunction with [AddNodeTo].  InstantiateNamespace derives a new child namespace
// from the provided parent namespace, then within that child namespace, instantiates all
// child nodes that were previously added using [AddNodeTo].  Child nodes are instantiated
// lazily, using [Namespace.Defer].
//
// The node argument is an [ir.IRNode] that also implements the [wiring.NamespaceHandler]
// interface, for receiving child nodes.
//
// Returns the child namespace.  If the child namespace has already been created, this method
// will return an error
func InstantiateNamespace(parentNamespace wiring.Namespace, namespaceNode IRNamespace) (wiring.Namespace, error) {
	namespace, err := parentNamespace.DeriveNamespace(namespaceNode.Name(), namespaceNode)
	if err != nil {
		return nil, err
	}

	namespace.Info("Deferring instantiation of child nodes")
	namespace.Defer(func() error {
		return instantiateNamespaceNodes(namespace)
	}, wiring.DeferOpts{Front: true})

	return namespace, err
}

// Instantiates the child nodes of namespace that were added using [AddNodeTo]
// Plugins that define custom namespaces can invoke this method in their BuildFunc
// Plugins should invoke this lazily with [namespace.Defer] to prevent instantiation cycles
func instantiateNamespaceNodes(namespace wiring.Namespace) error {
	namespaceName := namespace.Name()
	var nodeNames []string
	if err := namespace.GetProperties(namespaceName, prop_CHILDREN, &nodeNames); err != nil {
		return namespace.Error("InstantiateNamespace failed due to %v", err.Error())
	}
	namespace.Info("Deferred instantiation of children [%v]", strings.Join(nodeNames, ", "))

	for _, childName := range nodeNames {
		// Instantiate the child
		var child ir.IRNode
		if err := namespace.Get(childName, &child); err != nil {
			return err
		}
	}
	return nil
}
