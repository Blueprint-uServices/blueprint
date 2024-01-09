// Package clientpool provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a ClientPool that can create N instances of clients to a service.
package clientpool

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/wiring"
	"github.com/Blueprint-uServices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Wraps the client side of serviceName with a client pool with n client instances
func Create(spec wiring.WiringSpec, serviceName string, numClients int) {
	poolName := serviceName + ".clientpool"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to create clientpool for " + serviceName + " as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(spec, poolName)

	// Define the client pool
	spec.Define(poolName, &ClientPool{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		pool := &ClientPool{PoolName: poolName, N: numClients}
		poolNamespace, err := namespace.DeriveNamespace(poolName, &ClientPoolNamespace{pool})
		if err != nil {
			return nil, err
		}
		return pool, poolNamespace.Get(clientNext, &pool.Client)
	})
}

// A [wiring.NamespaceHandler] used to build [ClientPool] IRNodes
type ClientPoolNamespace struct {
	*ClientPool
}

// Implements [wiring.NamespaceHandler]
func (pool *ClientPool) Accepts(nodeType any) bool {
	_, isGolangNode := nodeType.(golang.Node)
	return isGolangNode
}

// Implements [wiring.NamespaceHandler]
func (pool *ClientPool) AddEdge(name string, edge ir.IRNode) error {
	pool.Edges = append(pool.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (pool *ClientPool) AddNode(name string, node ir.IRNode) error {
	pool.Nodes = append(pool.Nodes, node)
	return nil
}
