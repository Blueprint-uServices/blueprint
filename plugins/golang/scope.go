package golang

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

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
	_, ok := nodeType.(Node)
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
