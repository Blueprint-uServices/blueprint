package golang_process

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"golang.org/x/exp/slog"
)

// Used during building to accumulate golang application-level nodes
// Logic of the scope is as follows:
//   - Golang application-level nodes get stored in this scope and will be instantiated by a Golang process node
//   - TODO
type GolangProcessScope struct {
	blueprint.Scope

	visible     map[string]blueprint.IRNode
	parentScope blueprint.Scope
	wiring      *blueprint.WiringSpec
	node        *GolangProcessNode
}

func NewGolangProcessScope(parentScope blueprint.Scope, wiring *blueprint.WiringSpec, name string) *GolangProcessScope {
	scope := GolangProcessScope{}
	scope.visible = make(map[string]blueprint.IRNode)
	scope.parentScope = parentScope
	scope.wiring = wiring
	scope.node = newGolangProcessNode(name)
	return &scope
}

func (scope *GolangProcessScope) SetVisible(name string, node blueprint.IRNode) {
	scope.visible[name] = node
	scope.parentScope.SetVisible(name, node)
}

func (scope *GolangProcessScope) Visible(name string) bool {
	_, is_visible := scope.visible[name]
	return is_visible || scope.parentScope.Visible(name)
}

// The GolangProcessScope will store any golang nodes locally, but will
// consult the parent scope for any other types of node
func (scope *GolangProcessScope) Get(name string) (blueprint.IRNode, error) {

	// First, check to see if the node already exists in this scope.  If so just return it.
	if node, exists := scope.visible[name]; exists {
		slog.Debug("GolangProcess %s has already seen %s, returning existing node", scope.node.InstanceName, name)
		return node, nil
	}

	// Next, make sure the specified name has actually been defined
	def := scope.wiring.GetDef(name)
	if def == nil {
		err := fmt.Errorf("GolangProcess %s unable to find %s in the wiring spec", scope.node.InstanceName, name)
		slog.Error(err.Error())
		return nil, err
	}
	slog.Debug("Got %s of type %s\n", name, reflect.TypeOf(def.NodeType))

	// Check if this is a golang object node.  If not, we just get the node from the parent.
	if _, ok := def.NodeType.(golang.Node); !ok {
		slog.Debug("GolangProcess %s getting non-Golang node %s from parent", scope.node.InstanceName, name)
		node, err := scope.parentScope.Get(name)
		if err != nil {
			return nil, err
		}
		scope.visible[name] = node
		scope.node.AddArg(node)
		return node, nil
	}

	// We now know that the node is a golang node.

	// If the golang node exists in the parent scope, it is a visibility error
	if scope.parentScope.Visible(name) {
		err := fmt.Errorf("GolangProcess %s wants to instantiate a \"%s\" node but it has already been instantiated elsewhere", scope.node.InstanceName, name)
		slog.Error(err.Error())
		return nil, err
	}

	// Build the node
	node, err := def.Build(scope)
	if err != nil {
		err := fmt.Errorf("GolangProcess %s unable to build node %s, reason: %s", scope.node.InstanceName, name, err.Error())
		slog.Error(err.Error())
		return nil, err
	}

	// Check it's a golang node
	golang_node, is_golang_node := node.(golang.Node)
	if !is_golang_node {
		err = fmt.Errorf("GolangProcess %s expected %s to be a Golang node, but found %s instead", scope.node.InstanceName, name, golang_node)
		slog.Error(err.Error())
		return nil, err
	}

	// This scope doesn't restrict node visibility, so advertise the node's existence to the parent
	slog.Info("GolangProcess built child node", "GolangProcess", scope.node.InstanceName, "childNode", name)
	scope.SetVisible(name, golang_node)
	scope.node.AddChild(golang_node)
	return golang_node, nil
}

func (scope *GolangProcessScope) GetProperty(name string, key string) ([]any, error) {
	return scope.parentScope.GetProperty(name, key)
}

func (scope *GolangProcessScope) Build() (blueprint.IRNode, error) {
	return scope.node, nil
}

func (scope *GolangProcessScope) String() string {
	return "GolangProcess"
}
