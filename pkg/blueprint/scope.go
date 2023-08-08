package blueprint

import (
	"fmt"
	"reflect"

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

Most scope implementations should extend the BasicScope struct
*/
type Scope interface {
	Get(name string) (IRNode, error)                    // Get a node from this scope, possibly building it
	GetProperty(name string, key string) ([]any, error) // Get a property from this scope
	Put(name string, node IRNode) error                 // Put a node into this scope
	Build() (IRNode, error)                             // Build a node from the scope; optional
}

/*
This is a scope implementation that most plugins will want to extend.  It takes care of most of the functionality
needed for a scope while providing the following methods that can be overridden:
  - LookupDef -- looks up a definition in the wiring spec by name
  - Accepts -- declares whether this scope should build nodes of a particular type, or if they should be built in the parent scope
  - AddEdge -- saves a node that was gotten from the parent scope
  - AddNode -- saves a node that was built in the current scope
*/
type BasicScope struct {
	Scope

	Name        string            // A name for this scope
	ParentScope Scope             // The parent scope that created this scope; can be nil
	Wiring      *WiringSpec       // The wiring spec
	Nodes       map[string]IRNode // Nodes that were created in this scope
	Edges       map[string]IRNode // Nodes from parent scopes that have been referenced from this scope
}

func (scope *BasicScope) InitBasicScope(name string, parent Scope, wiring *WiringSpec) {
	scope.Name = name
	scope.ParentScope = parent
	scope.Wiring = wiring
	scope.Nodes = make(map[string]IRNode)
	scope.Edges = make(map[string]IRNode)
}

// Augments debug messages with information about the scope
func (scope *BasicScope) Info(message string, args ...any) {
	slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s", reflect.TypeOf(scope).String(), scope.Name, message), args...))
}

// Augments debug messages with information about the scope
func (scope *BasicScope) Debug(message string, args ...any) {
	slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s", reflect.TypeOf(scope).String(), scope.Name, message), args...))
}

// Augments error messages with information about the scope
func (scope *BasicScope) Error(err error) error {
	slog.Error(fmt.Sprintf("%s %s: %s", reflect.TypeOf(scope).String(), scope.Name, err.Error()))
	return err
}

// Augments error messages with information about the scope
func (scope *BasicScope) Errorf(message string, args ...string) error {
	return scope.Error(fmt.Errorf(message, args))
}

// Looks up a definition in the wiring spec.  By default this directly consults the wiring spec,
// but this method can be overridden to perform more complex selective logic if needed.
func (scope *BasicScope) LookupDef(name string) (*WiringDef, error) {
	def := scope.Wiring.GetDef(name)
	if def == nil {
		return nil, fmt.Errorf("%s does not exist in the wiring spec of scope %s", name, scope.Name)
	}
	return def, nil
}

// Asks if the node of the specified type should be built in this scope, or in the parent scope.
// Most Scope implementations should override this method to be selective about which nodes should
// get built in this scope.  For example, a golang scope only accepts golang nodes.
func (scope *BasicScope) Accepts(nodeType any) bool {
	return true
}

// This is called after getting a node from the parent scope.  By default it just saves the node
// as an edge.  Scope implementations can override this method to do other things.
func (scope *BasicScope) AddEdge(name string, node IRNode) (IRNode, error) {
	scope.Edges[name] = node
	return node, nil
}

// This is called after building a node in the current scope.  By default it just saves the node
// on the scope.  Scope implementations can override this method to do other things.
func (scope *BasicScope) AddNode(name string, node IRNode) (IRNode, error) {
	scope.Nodes[name] = node
	return node, nil
}

func (scope *BasicScope) Get(name string) (IRNode, error) {
	// If it already exists, return it
	if node, ok := scope.Nodes[name]; ok {
		return node, nil
	}
	if node, ok := scope.Edges[name]; ok {
		return node, nil
	}

	// Look up the definition
	def, err := scope.LookupDef(name)
	if err != nil {
		return nil, scope.Error(err)
	}
	scope.Debug("got %s of type %s", name, reflect.TypeOf(def.NodeType).String())

	// See if the node should be created here or in the parent
	if !scope.Accepts(def.NodeType) {
		scope.Debug("getting %s from parent scope", name)
		if scope.ParentScope == nil {
			return nil, scope.Errorf("scope does not accept nodes of type %s but there is no parent scope to get them from", reflect.TypeOf(def.NodeType).Name())
		}
		node, err := scope.ParentScope.Get(name)
		if err != nil {
			return nil, scope.Error(err)
		}
		return scope.AddEdge(name, node)
	}

	// Build the node
	node, err := def.Build(scope)
	if err != nil {
		return nil, scope.Error(err)
	}

	// Check the built node is an IRNode
	ir_node, ok := node.(IRNode)
	if !ok {
		return nil, scope.Errorf("build function for %s did not return an IRNode", name)
	}

	// Check the built node was of the type registered in the wiring declaration
	if !reflect.TypeOf(node).AssignableTo(reflect.TypeOf(def.NodeType)) {
		return nil, scope.Errorf("expected %s to be a %s but it is a %s", name, reflect.TypeOf(def.NodeType).Name(), reflect.TypeOf(node).Name())
	}

	scope.Info("Built node %s of type %s", name, reflect.TypeOf(node).String())
	return scope.AddNode(name, ir_node)
}

func (scope *BasicScope) GetProperty(name string, key string) ([]any, error) {
	def, err := scope.LookupDef(name)
	if err != nil {
		return nil, scope.Error(err)
	}
	return def.GetProperty(key), nil
}

func (scope *BasicScope) Put(name string, node IRNode) error {
	if !scope.Accepts(node) {
		scope.Debug("putting %s into parent scope", name)
		if scope.ParentScope != nil {
			return scope.Errorf("scope does not accept nodes of type %s but there is no parent scope to get them from", reflect.TypeOf(node).Name())
		}
		err := scope.ParentScope.Put(name, node)
		if err != nil {
			return scope.Error(err)
		}
		_, err = scope.AddEdge(name, node)
		return err
	} else {
		_, err := scope.AddNode(name, node)
		return err
	}
}

func (scope *BasicScope) Build() (IRNode, error) {
	return nil, scope.Errorf("cannot build a scope of this type")
}

type BlueprintScope struct {
	BasicScope
}

func newBlueprintScope(wiring *WiringSpec) (*BlueprintScope, error) {
	scope := BlueprintScope{}
	scope.InitBasicScope(wiring.name, nil, wiring)
	return &scope, nil
}

func (scope *BlueprintScope) Build() (IRNode, error) {
	node := ApplicationNode{}
	node.name = scope.Name
	node.children = scope.Nodes
	return &node, nil
}
