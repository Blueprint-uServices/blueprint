package clientpool

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Wraps the client side of a service with a client pool with N client instances
*/
func Create(spec wiring.WiringSpec, serviceName string, n int) {
	clientpool := serviceName + ".clientpool"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to create clientpool for " + serviceName + " as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(spec, clientpool)

	// Define the client pool
	spec.Define(clientpool, &ClientPool{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		pool := NewClientPoolNamespace(namespace, spec, clientpool, n)

		err := pool.Get(clientNext, &pool.handler.IRNode.Client)
		return pool.handler.IRNode, err
	})
}

type (
	ClientpoolNamespace struct {
		wiring.SimpleNamespace
		handler *clientpoolNamespaceHandler
	}

	clientpoolNamespaceHandler struct {
		wiring.DefaultNamespaceHandler

		IRNode *ClientPool
	}
)

func NewClientPoolNamespace(parent wiring.Namespace, spec wiring.WiringSpec, name string, n int) *ClientpoolNamespace {
	namespace := &ClientpoolNamespace{}
	namespace.handler = &clientpoolNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newClientPool(name, n)
	namespace.Init(name, "ClientPool", parent, spec, namespace.handler)
	return namespace
}

// Golang clientpools can only contain golang nodes
func (namespace *clientpoolNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

// When a node is added to this namespace, we just attach it to the IRNode representing the clientpool
func (handler *clientpoolNamespaceHandler) AddNode(name string, node ir.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this namespace, we just attach it as an argument to the IRNode representing the clientpool
func (handler *clientpoolNamespaceHandler) AddEdge(name string, node ir.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
