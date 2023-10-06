package clientpool

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Wraps the client side of a service with a client pool with N client instances
*/
func Create(wiring blueprint.WiringSpec, serviceName string, n int) {
	clientpool := serviceName + ".clientpool"

	// Get the pointer metadata
	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to create clientpool for " + serviceName + " as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(wiring, clientpool)

	// Define the client pool
	wiring.Define(clientpool, &ClientPool{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		pool := NewClientPoolNamespace(namespace, wiring, clientpool, n)

		err := pool.Get(clientNext, &pool.handler.IRNode.Client)
		return pool.handler.IRNode, err
	})
}

type (
	ClientpoolNamespace struct {
		blueprint.SimpleNamespace
		handler *clientpoolNamespaceHandler
	}

	clientpoolNamespaceHandler struct {
		blueprint.DefaultNamespaceHandler

		IRNode *ClientPool
	}
)

func NewClientPoolNamespace(parent blueprint.Namespace, wiring blueprint.WiringSpec, name string, n int) *ClientpoolNamespace {
	namespace := &ClientpoolNamespace{}
	namespace.handler = &clientpoolNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newClientPool(name, n)
	namespace.Init(name, "ClientPool", parent, wiring, namespace.handler)
	return namespace
}

// Golang processes can only contain golang nodes
func (namespace *clientpoolNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

// When a node is added to this namespace, we just attach it to the IRNode representing the process
func (handler *clientpoolNamespaceHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this namespace, we just attach it as an argument to the IRNode representing the process
func (handler *clientpoolNamespaceHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
