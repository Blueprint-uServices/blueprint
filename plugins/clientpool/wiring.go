// Package clientpool provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a ClientPool that can create N instances of clients to a service.
package clientpool

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
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
		poolNode := &ClientPool{PoolName: poolName, N: numClients}
		poolNamespace := wiring.CreateNamespace(spec, namespace, poolNode)
		err := poolNamespace.Get(clientNext, &poolNode.Client)
		return poolNode, err
	})
}
