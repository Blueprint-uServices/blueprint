package goproc

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Adds a child node to an existing process
func AddChildToProcess(wiring blueprint.WiringSpec, procName, childName string) {
	wiring.AddProperty(procName, "Children", childName)
}

// Adds a process that explicitly instantiates all of the children provided.
// The process will also implicitly instantiate any of the dependencies of the children
func CreateProcess(wiring blueprint.WiringSpec, procName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddChildToProcess(wiring, procName, childName)
	}

	wiring.Define(procName, &Process{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		process := NewGolangProcessNamespace(namespace, wiring, procName)

		var childNames []string
		if err := namespace.GetProperties(procName, "Children", &childNames); err != nil {
			return nil, blueprint.Errorf("unable to build Golang process as the \"Children\" property is invalid: %s", err.Error())
		}
		process.Info("%v children to build (%s)", len(childNames), strings.Join(childNames, ", "))

		// Instantiate all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, childName := range childNames {
			ptr := pointer.GetPointer(wiring, childName)
			if ptr == nil {
				// for non-pointer types, just get the child node
				_, err := process.Get(childName)
				if err != nil {
					return nil, err
				}
			} else {
				// for pointer nodes, only instantiate the dst side of the pointer
				_, err := ptr.InstantiateDst(process)
				if err != nil {
					return nil, err
				}
			}
		}

		// Instantiate and return the service
		return process.handler.IRNode, nil
	})

	return procName
}

// Creates a process that contains clients to the specified children.  This is for convenience in
// serving as a starting point to write a custom client
func CreateClientProcess(wiring blueprint.WiringSpec, procName string, children ...string) string {
	for _, childName := range children {
		AddChildToProcess(wiring, procName, childName)
	}

	wiring.Define(procName, &Process{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		process := NewGolangProcessNamespace(namespace, wiring, procName)

		var childNames []string
		if err := namespace.GetProperties(procName, "Children", &childNames); err != nil {
			return nil, blueprint.Errorf("unable to build Golang process as the \"Children\" property is not defined: %s", err.Error())
		}
		process.Info("%v children to build (%s)", len(childNames), strings.Join(childNames, ", "))

		// Instantiate all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, childName := range childNames {
			_, err := process.Get(childName)
			if err != nil {
				return nil, err
			}
		}

		// Instantiate and return the service
		return process.handler.IRNode, nil
	})

	return procName
}

// Used during building to accumulate golang application-level nodes
// Non-golang nodes will just be recursively fetched from the parent namespace
type ProcessNamespace struct {
	blueprint.SimpleNamespace
	handler *processNamespaceHandler
}

type processNamespaceHandler struct {
	blueprint.DefaultNamespaceHandler

	IRNode *Process
}

// Creates a process `name` within the provided parent namespace
func NewGolangProcessNamespace(parentNamespace blueprint.Namespace, wiring blueprint.WiringSpec, name string) *ProcessNamespace {
	namespace := &ProcessNamespace{}
	namespace.handler = &processNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newGolangProcessNode(name)
	namespace.Init(name, "GolangProcess", parentNamespace, wiring, namespace.handler)
	return namespace
}

// Golang processes can only contain golang nodes
func (namespace *processNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

// When a node is added to this namespace, we just attach it to the IRNode representing the process
func (handler *processNamespaceHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this namespace, we just attach it as an argument to the IRNode representing the process
func (handler *processNamespaceHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
