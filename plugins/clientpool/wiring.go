// Package clientpool provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a ClientPool that can create N instances of clients to a service.
package clientpool

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/namespacebuilder"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Wraps the client side of serviceName with a client pool with n client instances
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
		pool := namespacebuilder.Create[golang.Node](namespace, spec, "ClientPool", clientpool)
		var client golang.Service
		err := pool.Namespace.Get(clientNext, &client)
		return newClientPool(clientpool, n, client, pool.ArgNodes, pool.ContainedNodes), err
	})
}
