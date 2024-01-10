// Package clientpool is a plugin for adding a client pool to the client side of service calls.
//
// By default, Blueprint instantiate one client to a service and there is no concurrency control
// or rate limiting of calls using that client.
//
// When applied, the clientpool plugin instantiates N instances of clients to a service, and
// each client instance can only be used by one caller at a time, effectively rate-limiting to
// N outstanding calls at a time.
//
// To use the clientpool plugin in your wiring spec, simply apply it to an application-level service instance:
//
//	clientpool.Create(spec, "my_service", 10)
//
// After applying the clientpool plugin to a service, you can continue to apply application-level
// modifiers to the service.
//
// During compilation, the clientpool plugin will generate a client-side wrapper class.  The plugin
// also utilizes some code in the [runtime/plugins/clientpool] package.
//
// [runtime/plugins/clientpool]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/clientpool
package clientpool

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Modifies the client-side of an application-level service so that all calls to serviceName
// are made using a pool of numClients clients.  At runtime, clients to serviceName will only
// be able to have up to numClients concurrent calls outstanding before subsequent calls block.
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
		poolNamespace, err := namespace.DeriveNamespace(poolName, &clientPoolNamespace{pool})
		if err != nil {
			return nil, err
		}
		return pool, poolNamespace.Get(clientNext, &pool.Client)
	})
}

// A [wiring.NamespaceHandler] used to build [ClientPool] IRNodes
type clientPoolNamespace struct {
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
