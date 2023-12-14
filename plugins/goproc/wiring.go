package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

var prop_CHILDREN = "Children"

// // Adds a child node to an existing process
// func AddChildToProcess(spec wiring.WiringSpec, procName, childName string) {
// 	spec.AddProperty(procName, prop_CHILDREN, childName)
// }

// // Adds a process that explicitly instantiates all of the children provided.
// // The process will also implicitly instantiate any of the dependencies of the children
// func CreateProcess(spec wiring.WiringSpec, procName string, children ...string) string {
// 	// If any children were provided in this call, add them to the process via a property
// 	for _, childName := range children {
// 		AddChildToProcess(spec, procName, childName)
// 	}

// 	// The process node is simply a namespace that accepts [golang.Node] nodes
// 	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
// 		proc := newGolangProcessNode(procName)
// 		procNamespace := wiring.CreateNamespace(spec, namespace, proc)
// 		_, err := pointer.InstantiateFromProperty(spec, procNamespace, prop_CHILDREN)
// 		return proc, err
// 	})

// 	return procName
// }

// // Creates a process that contains clients to the specified children.  This is for convenience in
// // serving as a starting point to write a custom client
// func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
// 	for _, childName := range children {
// 		AddChildToProcess(spec, procName, childName)
// 	}

// 	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
// 		proc := newGolangProcessNode(procName)
// 		procNamespace := wiring.CreateNamespace(spec, namespace, proc)
// 		_, err := pointer.InstantiateClientsFromProperty(spec, procNamespace, prop_CHILDREN)
// 		return proc, err
// 	})

// 	return procName
// }

// Adds a child node to an existing process
func AddChildToProcess(spec wiring.WiringSpec, procName, childName string) {
	ptr := pointer.GetPointer(spec, childName)
	if ptr == nil {
		// The simple case: the child is not a pointer, nothing else needs to happen
		spec.AddProperty(procName, prop_CHILDREN, childName)
		return
	}

	modifierName := childName + ".process"
	ptrNext := ptr.AddDstModifier(spec, modifierName)
	spec.Define(modifierName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		// Get or build the process.  All nodes within the process will now be queued up for
		// delayed / lazy instantiation.
		var proc *Process
		if err := namespace.Get(procName, &proc); err != nil {
			return nil, err
		}

		// Immediately build and get the next modifier from within the process
		var nextNode ir.IRNode
		err := proc.namespace.Get(ptrNext, &nextNode)
		return nextNode, err
	})

	// By adding ptrNext to the proc children, when procName is instantiated, it will then
	// lazily create ptrNext
	spec.AddProperty(procName, prop_CHILDREN, ptrNext)
}

// Deploys serviceName as a process; this automatically creates a process and deploys serviceName in it
// Returns serviceName
func DeployService(spec wiring.WiringSpec, serviceName string) string {
	procName := serviceName + "_process"
	CreateProcess(spec, procName, serviceName)
	return serviceName
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

		// Get the child nodes that will later be instantiated
		var childNodes []string
		if err := namespace.GetProperties(procName, prop_CHILDREN, &childNodes); err != nil {
			return nil, err
		}

		// Create the process but don't instantiate children yet; do that later
		proc := newGolangProcessNode(procName)
		proc.namespace = wiring.CreateNamespace(spec, namespace, proc)
		proc.namespace.Defer(func() error {
			for _, childNode := range childNodes {
				var dst ir.IRNode
				err := proc.namespace.Get(childNode, &dst)
				if err != nil {
					return err
				}

				if addr, isAddr := dst.(address.Node); isAddr && addr.GetDestination() == nil {
					nextNode, err := address.PointsTo(proc.namespace, childNode)
					if err != nil {
						return err
					}

					err = proc.namespace.Get(nextNode, &dst)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		return proc, nil
	})

	return procName
}

// Creates a process that contains clients to the specified children.  This is for convenience in
// serving as a starting point to write a custom client
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	// No modifiers or special treatment needed when instantiating the client side
	for _, childName := range children {
		spec.AddProperty(procName, prop_CHILDREN, childName)
	}

	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace := wiring.CreateNamespace(spec, namespace, proc)
		_, err := pointer.InstantiateClientsFromProperty(spec, procNamespace, prop_CHILDREN)
		return proc, err
	})

	return procName
}
