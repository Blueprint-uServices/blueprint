package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Adds a child node to an existing process
func AddToProcess(spec wiring.WiringSpec, procName, childName string) {
	namespaceutil.AddNodeTo[Process](spec, procName, childName)
}

// Wraps serviceName with a modifier that deploys the service inside a Golang process
func Deploy(spec wiring.WiringSpec, serviceName string) string {
	procName := serviceName + "_proc"
	CreateProcess(spec, procName, serviceName)
	return serviceName
}

// Creates a process with a given name, and adds the provided nodes as children.  This method
// is only needed when creating processes with more than one child node; otherwise it is easier
// to use [Deploy]
func CreateProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddToProcess(spec, procName, childName)
	}

	// The process node is simply a namespace that accepts [golang.Node] nodes
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		_, err := namespaceutil.InstantiateNamespace(namespace, &GolangProcessNamespace{proc})
		return proc, err
	})

	return procName
}

// Creates a process that contains clients to the specified children.  This is for convenience in
// serving as a starting point to write a custom client
func CreateClientProcess(spec wiring.WiringSpec, procName string, children ...string) string {
	spec.Define(procName, &Process{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		proc := newGolangProcessNode(procName)
		procNamespace, err := namespace.DeriveNamespace(procName, &GolangProcessNamespace{proc})
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			var childNode ir.IRNode
			if err := procNamespace.Get(child, &childNode); err != nil {
				return nil, err
			}
		}
		return proc, err
	})

	return procName
}

// A [wiring.NamespaceHandler] used to build [Process] IRNodes
type GolangProcessNamespace struct {
	*Process
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) Accepts(nodeType any) bool {
	_, isGolangNode := nodeType.(golang.Node)
	return isGolangNode
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) AddEdge(name string, edge ir.IRNode) error {
	proc.Edges = append(proc.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (proc *GolangProcessNamespace) AddNode(name string, node ir.IRNode) error {
	proc.Nodes = append(proc.Nodes, node)
	return nil
}
