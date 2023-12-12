package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

var prop_CHILDREN = "Children"

// Adds a child node to an existing process
func AddChildToProcess(spec wiring.WiringSpec, procName, childName string) {
	spec.AddProperty(procName, prop_CHILDREN, childName)
}

// Adds a process that explicitly instantiates all of the children provided.
// The process will also implicitly instantiate any of the dependencies of the children
func CreateProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddChildToProcess(spec, procName, childName)
	}

	// The process node is simply a namespace that accepts [golang.Node] nodes
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace := wiring.CreateNamespace(spec, namespace, proc)
		_, err := pointer.InstantiateFromProperty(spec, procNamespace, prop_CHILDREN)
		return proc, err
	})

	return procName
}

// Creates a process that contains clients to the specified children.  This is for convenience in
// serving as a starting point to write a custom client
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	for _, childName := range children {
		AddChildToProcess(spec, procName, childName)
	}

	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace := wiring.CreateNamespace(spec, namespace, proc)
		_, err := pointer.InstantiateClientsFromProperty(spec, procNamespace, prop_CHILDREN)
		return proc, err
	})

	return procName
}
