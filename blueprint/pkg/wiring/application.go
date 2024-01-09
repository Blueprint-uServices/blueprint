package wiring

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

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
	// Create the root application namespace
	app := &ir.ApplicationNode{ApplicationName: name}

	namespace := &namespaceimpl{
		NamespaceName:   name,
		NamespaceType:   "BlueprintApplication",
		ParentNamespace: nil,
		Wiring:          spec,
		Handler:         &applicationNamespaceHandler{app: app},
		Seen:            make(map[string]ir.IRNode),
		Added:           make(map[string]any),
		ChildNamespaces: make(map[string]Namespace),
	}

	// If no nodes were specified, then instead we will instantiate all defined nodes
	if len(nodesToInstantiate) == 0 {
		nodesToInstantiate = spec.Defs()
	}

	// Queue up the nodes to be built
	for i := range nodesToInstantiate {
		nodeName := nodesToInstantiate[i]
		namespace.Defer(func() error {
			namespace.Info("Instantiating %v", nodeName)
			var node ir.IRNode
			return namespace.Get(nodeName, &node)
		}, DeferOpts{Front: true})
	}

	// Execute deferred functions until empty
	for len(namespace.Deferred) > 0 {
		next := namespace.Deferred[0]
		namespace.Deferred = namespace.Deferred[1:]
		if err := next(); err != nil {
			return app, err
		}
	}
	return app, nil
}

type applicationNamespaceHandler struct {
	NamespaceHandler
	app *ir.ApplicationNode
}

// NamespaceHandler
func (handler *applicationNamespaceHandler) Accepts(any) bool {
	return true
}

// NamespaceHandler
func (handler *applicationNamespaceHandler) AddEdge(name string, node ir.IRNode) error {
	return fmt.Errorf("BlueprintApplication %v encountered unexpected edge %v %v", handler.app.ApplicationName, name, node)
}

// NamespaceHandler
func (handler *applicationNamespaceHandler) AddNode(name string, node ir.IRNode) error {
	handler.app.Children = append(handler.app.Children, node)
	return nil
}
