package simplenosqldb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

/* Creates a simple queue instance with the specified name */
func Define(spec wiring.WiringSpec, qName string) string {
	// Define the queue backend
	backendName := qName + ".backend"
	spec.Define(backendName, &SimpleQueue{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newSimpleQueue(qName)
	})

	// Mandate that this queue with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := qName + ".dst"
	spec.Alias(dstName, backendName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(spec, qName, &SimpleQueue{}, dstName)

	return qName
}
