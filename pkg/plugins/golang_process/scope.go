package golang_process

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
)

// Used during building to accumulate golang application-level nodes
// Logic of the scope is as follows:
//   - Golang application-level nodes get stored in this scope and will be instantiated by a Golang process node
//   - TODO
type GolangProcessScope struct {
	blueprint.Scope

	parentScope blueprint.Scope
	wiring      *blueprint.WiringSpec
	node        *GolangProcessNode
}

func NewGolangProcessScope(parentScope blueprint.Scope, wiring *blueprint.WiringSpec, name string) *GolangProcessScope {
	scope := GolangProcessScope{}
	scope.parentScope = parentScope
	scope.wiring = wiring
	scope.node = newGolangProcessNode(name)
	return &scope
}

// The GolangProcessScope will store any golang nodes locally, but will
// consult the parent scope for any other types of node
func (scope *GolangProcessScope) Get(name string) (blueprint.IRNode, error) {

	nodetype, build := scope.wiring.GetDef(name)
	if nodetype == nil {
		return nil, fmt.Errorf("could not find \"%s\" in wiring spec", name)
	}
	fmt.Printf("Got %s of type %s\n", name, reflect.TypeOf(nodetype))

	// Golang Processes can only contain golang nodes; if this isn't a golang node
	// then we just get the node from the parent scope and save it as an argument
	// that will be passed in to the process
	if _, ok := nodetype.(golang.Node); !ok {
		fmt.Println("Getting from parent scope")
		node, err := scope.parentScope.Get(name)
		if err != nil {
			return nil, err
		}
		scope.node.AddArg(node)
		return node, nil
	}

	// Build golang nodes and save them in the process scope
	node, err := build(scope)
	if err != nil {
		return nil, err
	}
	if golang_node, ok := node.(golang.Node); ok {
		scope.node.AddChild(golang_node)
		return golang_node, nil
	} else {
		return nil, fmt.Errorf("%s was declared as a node of type %s but was actually %s", name, nodetype, node)
	}
}

func (scope *GolangProcessScope) Build() (blueprint.IRNode, error) {
	return scope.node, nil
}

func (scope *GolangProcessScope) String() string {
	return "GolangProcess"
}

// Get(name string) (IRNode, error)
// Build() (IRNode, error)
// String() string
