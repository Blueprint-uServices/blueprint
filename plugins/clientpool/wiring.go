// Package clientpool is a plugin that wraps the client side of a service to use a pool of N clients, disallowing
// callers from making more than N concurrent oustanding calls each to the service.
//
// # Wiring Spec Usage
//
// To use the clientpool plugin in your wiring spec, simply apply it to an application-level service instance:
//
//	clientpool.Create(spec, "my_service", 10)
//
// # Description
//
// When applied, the clientpool plugin instantiates N instances of clients to a service, and callers have exclusive
// access to a client when making a call.  This effectively limits the caller-side to only having N outstanding calls at
// a time, with any extra calls blocking until a previous call completes and a client becomes available.
// By contrast, the default Blueprint behavior is for all callers to share a single client that allows an unlimited
// number of concurrent calls.
//
// After applying the clientpool plugin to a service, you can continue to apply application-level
// modifiers to the service.
//
// # Artifacts Generated
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

// Create can be used by wiring specs to add a clientpool to the client side of a service.
//
// This will modify the client-side of serviceName so that all calls are made using a pool of numClients clients.
//
// At runtime, clients to serviceName will only be able to have up to numClients concurrent calls outstanding,
// before subsequent calls block.
//
// serviceName must be an application-level service instance, e.g. clientpool must be applied to the service
// before deploying the service over RPC or to a process.
//
// After calling [Create] you can continue to apply application-level modifiers to serviceName.
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
