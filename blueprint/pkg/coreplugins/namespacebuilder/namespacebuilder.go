// Package namespacebuilder provides a simple wiring spec utility struct called NamespaceBuilder
// that provides a simple way for plugins to create basic namespaces.
package namespacebuilder

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// NamespaceBuilder is a utility struct for use by plugins that want to create child namespaces
// of a particular node type.  For example, a golang process node wants to create a namespace
// that collects golang instances.
//
// NamespaceBuilder is the simplest way for a plugin to create child namespaces and most plugins
// should use this.
//
// Instances can be created with [Create] or [CustomCreate].
type NamespaceBuilder struct {
	// A name for this namespace; it is only used when logging things during compilation
	Name string

	// All nodes that were explicitly instantiated through calls to [NamespaceBuilder.Instantiate]
	// These nodes can either be ContainedNodes or ArgNodes, and will show up in one of those slices
	InstantiatedNodes map[string]ir.IRNode

	// All nodes that have been instantiated explicitly or implicitly within this namespace
	ContainedNodes []ir.IRNode

	// All nodes that are required by this namespace, but are passed in from the parent namespace
	ArgNodes []ir.IRNode

	Namespace wiring.Namespace

	spec    wiring.WiringSpec
	accepts func(any) bool
}

type namespaceBuilderHandler struct {
	wiring.DefaultNamespaceHandler
	builder *NamespaceBuilder
}

// Creates a NamespaceBuilder that will internally build any node of type [T].  Other node
// types will be recursively built in the parent namespace.
func Create[T any](parent wiring.Namespace, spec wiring.WiringSpec, namespacetype, name string) *NamespaceBuilder {
	accepts := func(nodeType any) bool {
		_, isT := nodeType.(T)
		return isT
	}
	return CustomCreate(parent, spec, namespacetype, name, accepts)
}

// Creates a NamespaceBuilder that will internally build any node for which accepts returns true.  Other node
// types will be recursively built in the parent namespace.
func CustomCreate(parent wiring.Namespace, spec wiring.WiringSpec, namespacetype, name string, accepts func(any) bool) *NamespaceBuilder {
	builder := &NamespaceBuilder{
		Name:    name,
		spec:    spec,
		accepts: accepts,
	}

	{
		namespace := &wiring.SimpleNamespace{}
		handler := &namespaceBuilderHandler{}
		handler.builder = builder

		handler.Init(namespace)
		namespace.Init(name, namespacetype, parent, spec, handler)

		builder.Namespace = namespace
	}

	builder.InstantiatedNodes = make(map[string]ir.IRNode)

	return builder
}

// Gets the specified nodes in this namespace.
// If a node is not a pointer, then the node is just instantiated using Get
// If a node is a pointer, then this will instantiate the server-side of
// the pointer.  See [NamespaceBuilder.Get] for client-side instantiation.
// Note that depending on the node type, it might be recursively fetched from the
// parent namespace, so it is possible for the node to be either a ContainedNode
// or an ArgNode
func (b *NamespaceBuilder) Instantiate(names ...string) (err error) {
	for _, childName := range names {
		var child ir.IRNode
		ptr := pointer.GetPointer(b.spec, childName)
		if ptr == nil {
			err = b.Namespace.Get(childName, &child)
		} else {
			child, err = ptr.InstantiateDst(b.Namespace)
		}
		if err != nil {
			return
		}
		b.InstantiatedNodes[childName] = child
	}
	return nil
}

// Inspects the propertyName property of [b.Name] and instantiates the node names
// stored in that property.
//
// Returns an error if propertyName doesn't exist, or if it isn't a string property,
// or if the nodes couldn't be built.
func (b *NamespaceBuilder) InstantiateFromProperty(propertyName string) error {
	var nodeNames []string
	if err := b.Namespace.GetProperties(b.Name, propertyName, &nodeNames); err != nil {
		return blueprint.Errorf("%v InstantiateFromProperty %v failed due to %s", b.Name, propertyName, err.Error())
	}
	b.Namespace.Info("%v = %s", propertyName, strings.Join(nodeNames, ", "))
	return b.Instantiate(nodeNames...)
}

// Gets the specified nodes in this namespace.
// If a node is not a pointer, then the node is just instantiated using Get
// If a node is a pointer, then this will only instantiate the client-side of
// the pointer.  See [NamespaceBuilder.Instantiate] for server-side instantiation.
// Note that depending on the node type, it might be recursively fetched from the
// parent namespace, so it is possible for the node to be either a ContainedNode
// or an ArgNode
func (b *NamespaceBuilder) InstantiateClients(names ...string) error {
	for _, childName := range names {
		var child ir.IRNode
		if err := b.Namespace.Get(childName, &child); err != nil {
			return err
		}
		b.InstantiatedNodes[childName] = child
	}
	return nil
}

// Inspects the propertyName property of [b.Name] and instantiates clients for
// the node names stored in that property.
//
// Returns an error if propertyName doesn't exist, or if it isn't a string property,
// or if the nodes couldn't be built.
func (b *NamespaceBuilder) InstantiateClientsFromProperty(propertyName string) error {
	var nodeNames []string
	if err := b.Namespace.GetProperties(b.Name, propertyName, &nodeNames); err != nil {
		return blueprint.Errorf("%v InstantiateClientsFromProperty %v failed due to %s", b.Name, propertyName, err.Error())
	}
	b.Namespace.Info("%v = %s", propertyName, strings.Join(nodeNames, ", "))
	return b.InstantiateClients(nodeNames...)
}

// Implements wiring.DefaultNamespaceHandler
func (h *namespaceBuilderHandler) Accepts(nodeType any) bool {
	return h.builder.accepts(nodeType)
}

// Implements wiring.DefaultNamespaceHandler
func (h *namespaceBuilderHandler) AddNode(name string, node ir.IRNode) error {
	h.builder.ContainedNodes = append(h.builder.ContainedNodes, node)
	return nil
}

// Implements wiring.DefaultNamespaceHandler
func (h *namespaceBuilderHandler) AddEdge(name string, node ir.IRNode) error {
	h.builder.ArgNodes = append(h.builder.ArgNodes, node)
	return nil
}
