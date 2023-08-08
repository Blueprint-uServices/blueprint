package golang_process

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
)

// Used during building to accumulate golang application-level nodes
// Logic of the scope is as follows:
//   - Golang application-level nodes get stored in this scope and will be instantiated by a Golang process node
//   - TODO
type GolangProcessScope struct {
	blueprint.BasicScope
}

func NewGolangProcessScope(parentScope blueprint.Scope, wiring *blueprint.WiringSpec, name string) *GolangProcessScope {
	scope := GolangProcessScope{}
	scope.InitBasicScope(name, parentScope, wiring)
	return &scope
}

// Asks if the node of the specified type should be built in this scope, or in the parent scope.
// Most Scope implementations should override this method to be selective about which nodes should
// get built in this scope.  For example, a golang scope only accepts golang nodes.
func (scope *GolangProcessScope) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

func (scope *GolangProcessScope) Build() (blueprint.IRNode, error) {
	procnode := newGolangProcessNode(scope.Name)
	for _, node := range scope.Nodes {
		err := procnode.AddChild(node)
		if err != nil {
			return nil, err
		}
	}
	for _, edge := range scope.Edges {
		procnode.AddArg(edge)
	}
	return procnode, nil
}
