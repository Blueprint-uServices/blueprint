package wiring

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"

// Builds the IR of an application using the definitions of the provided spec.  Returns
// an [ir.ApplicationNode] of the application.
//
// Callers should typically provide nodesToInstantiate to specify which nodes should
// be instantiated in the application.  This method will recursively instantiate any
// dependencies.
//
// If nodesToInstantiate is empty, all nodes will be instantiated,
// but this might not result in an application with the desired topology.  Hence
// the recommended approach is to explicitly specify which nodes to instantiate.
func BuildApplicationIR(spec WiringSpec, name string, nodesToInstantiate ...string) (*ir.ApplicationNode, error) {
	// Create a root namespace for the application
	namespace := newRootNamespace(spec, name)

	// If no nodes were specified, then instead we will instantiate all defined nodes
	if len(nodesToInstantiate) == 0 {
		nodesToInstantiate = spec.Defs()
	}

	// Queue up the nodes to be built
	for _, nodeName := range nodesToInstantiate {
		namespace.instantiate(nodeName)
	}

	// Build 'em all
	return namespace.buildApplication()
}

// A root namespace used when building the IR for an application
type rootNamespace struct {
	SimpleNamespace

	handler *rootNamespaceHandler
}

type rootNamespaceHandler struct {
	DefaultNamespaceHandler

	application *ir.ApplicationNode
}

func newRootNamespace(spec WiringSpec, name string) *rootNamespace {
	namespace := &rootNamespace{}
	handler := rootNamespaceHandler{}
	handler.Init(&namespace.SimpleNamespace)
	handler.application = &ir.ApplicationNode{}
	namespace.handler = &handler
	namespace.Init(name, "BlueprintApplication", nil, spec, &handler)
	return namespace
}

// Adds nodeName to be built when buildApplication is invoked
func (namespace *rootNamespace) instantiate(nodeName string) {
	namespace.Defer(func() error {
		namespace.Info("Instantiating %v", nodeName)
		var node ir.IRNode
		return namespace.Get(nodeName, &node)
	})
}

// Builds all nodes that were added using instantiate as well as any
// recursively dependent nodes
func (namespace *rootNamespace) buildApplication() (*ir.ApplicationNode, error) {
	node := &ir.ApplicationNode{ApplicationName: namespace.Name()}

	// Execute deferred functions until empty
	for len(namespace.Deferred) > 0 {
		next := namespace.Deferred[0]
		namespace.Deferred = namespace.Deferred[1:]
		err := next()
		if err != nil {
			node.Children = namespace.handler.Nodes
			return node, err
		}
	}

	// Attach all nodes created within the namespace to the application node
	node.Children = namespace.handler.Nodes
	return node, nil
}
