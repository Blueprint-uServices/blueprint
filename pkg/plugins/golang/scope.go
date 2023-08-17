package golang

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

type ProcessScope struct {
	blueprint.SimpleScope
	handler *processScopeHandler
}

// Used during building to accumulate golang application-level nodes
// Logic of the scope is as follows:
//   - Golang application-level nodes get stored in this scope and will be instantiated by a Golang process node
//   - TODO
type processScopeHandler struct {
	blueprint.DefaultScopeHandler

	irNode *Process
}

func NewGolangProcessScope(parentScope blueprint.Scope, wiring blueprint.WiringSpec, name string) *ProcessScope {
	scope := &ProcessScope{}
	scope.handler = &processScopeHandler{}
	scope.handler.Init(&scope.SimpleScope)
	scope.handler.irNode = newGolangProcessNode(name)
	scope.Init(name, "GolangProcess", parentScope, wiring, scope.handler)
	return scope
}

// Asks if the node of the specified type should be built in this scope, or in the parent scope.
// Most Scope implementations should override this method to be selective about which nodes should
// get built in this scope.  For example, a golang scope only accepts golang nodes.
func (scope *processScopeHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(Node)
	return ok
}

func (handler *processScopeHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.irNode.AddChild(node)
}

func (handler *processScopeHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.irNode.AddArg(node)
	return nil
}

func (scope *ProcessScope) Build() (blueprint.IRNode, error) {
	return scope.handler.irNode, nil
}
