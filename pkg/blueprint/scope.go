package blueprint

import (
	"fmt"

	"golang.org/x/exp/slog"
)

/*
A Scope is used during the IR-building process to accumulate built nodes.

Blueprint has several basic out-of-the-box scopes that are used when building applications.  A plugin can implement
its own custom scope.  Implementing a custom Scope is useful to achieve any of the following:
  - Scopes are the mechanism for limiting the visibility and addressibility of nodes
  - Scopes are the mechanism for templating nodes (e.g. to implement replication of nodes)
  - Scopes are the mechanism that determine addressing arguments that must be passed into Namespaces

Scopes are useful for implementing Namespace nodes, because Namespace nodes accumulate Nodes of a particular type.
For example, to build a GoProcess that contains Golang object instances, there will be a Scope that accumulates
Golang object instance nodes during the building process, and then creates a GoProcess namespace node.

Scopes are not the same as Namespaces, but to implement a Namespace will require a custom Scope for accumulating nodes.
*/
type Scope interface {
	Get(name string) (IRNode, error)
	Build() (IRNode, error)
}

// The Base scope used by Blueprint to accumulate the Blueprint application as a whole
type blueprintScope struct {
	Scope

	wiring *WiringSpec
	nodes  map[string]IRNode
}

func newBlueprintScope(wiring *WiringSpec) (*blueprintScope, error) {
	scope := blueprintScope{}
	scope.wiring = wiring
	scope.nodes = make(map[string]IRNode)
	return &scope, nil
}

func (scope *blueprintScope) Get(name string) (IRNode, error) {
	node, ok := scope.nodes[name]
	if ok {
		return node, nil
	}

	_, build := scope.wiring.GetDef(name)
	if build == nil {
		return nil, fmt.Errorf("wiring spec doesn't contain \"%s\".  Known nodes: %s", name, scope.wiring)
	}

	slog.Info("Building", "node", name)
	inode, err := build(scope)
	if err != nil {
		return nil, err
	}

	node, ok = inode.(IRNode)
	if !ok {
		// TODO: support e.g. configuration strings as well as nodes
		return nil, fmt.Errorf("lookup of node %s returned something that is not an IRNode (possibly unimplemented): %s", name, inode)
	}
	scope.nodes[name] = node

	return node, nil
}

func (scope *blueprintScope) Close() (interface{}, error) {
	node := ApplicationNode{}
	node.children = scope.nodes
	return &node, nil
}

func (scope *blueprintScope) Build() (IRNode, error) {
	node := ApplicationNode{}
	node.name = scope.wiring.name
	node.children = scope.nodes
	return &node, nil
}
