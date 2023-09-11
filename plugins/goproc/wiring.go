package goproc

import (
	"fmt"
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

	wiring.Define(procName, &Process{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		process := NewGolangProcessScope(scope, wiring, procName)

		childNames, err := scope.GetProperties(procName, "Children")
		if err != nil {
			return nil, fmt.Errorf("unable to build Golang process as the \"Children\" property is not defined: %s", err.Error())
		}
		var childNameStrings []string
		for _, childName := range childNames {
			childNameStrings = append(childNameStrings, childName.(string))
		}
		process.Info("%v children to build (%s)", len(childNames), strings.Join(childNameStrings, ", "))

		// Instantiate all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, childName := range childNames {
			ptr := pointer.GetPointer(wiring, childName.(string))
			if ptr == nil {
				// for non-pointer types, just get the child node
				_, err := process.Get(childName.(string))
				if err != nil {
					return nil, err
				}
			} else {
				// for pointer nodes, only instantiate the dst side of the pointer
				dst, err := ptr.InstantiateDst(process)
				if err != nil {
					return nil, err
				}
				process.handler.IRNode.Instantiate(dst)
			}
		}

		// Instantiate and return the service
		return process.handler.IRNode, nil
	})

	return procName
}

// Used during building to accumulate golang application-level nodes
// Non-golang nodes will just be recursively fetched from the parent scope
type ProcessScope struct {
	blueprint.SimpleScope
	handler *processScopeHandler
}

type processScopeHandler struct {
	blueprint.DefaultScopeHandler

	IRNode *Process
}

// Creates a process `name` within the provided parent scope
func NewGolangProcessScope(parentScope blueprint.Scope, wiring blueprint.WiringSpec, name string) *ProcessScope {
	scope := &ProcessScope{}
	scope.handler = &processScopeHandler{}
	scope.handler.Init(&scope.SimpleScope)
	scope.handler.IRNode = newGolangProcessNode(name)
	scope.Init(name, "GolangProcess", parentScope, wiring, scope.handler)
	return scope
}

// Golang processes can only contain golang nodes
func (scope *processScopeHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

// When a node is added to this scope, we just attach it to the IRNode representing the process
func (handler *processScopeHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this scope, we just attach it as an argument to the IRNode representing the process
func (handler *processScopeHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
