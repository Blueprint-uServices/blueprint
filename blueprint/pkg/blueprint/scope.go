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
	Name() string                                         // The name of this scope
	Get(name string) (IRNode, error)                      // Get a node from this scope or a parent scope, possibly building it
	Instantiate(name string) (IRNode, error)              // The same as Get, but without creating a dependency (an edge) into the current scope
	GetProperty(name string, key string) (any, error)     // Get a property from this scope
	GetProperties(name string, key string) ([]any, error) // Get a property from this scope
	Put(name string, node IRNode) error                   // Put a node into this scope
	Defer(f func() error)                                 // Enqueue a function to be executed once finished building the current nodes

	Info(message string, args ...any)        // Logging
	Warn(message string, args ...any)        // Logging
	Error(message string, args ...any) error // Logging
}

/*
A SimpleScope implements all of the Scope methods and only requires users to implement a SimpleScopeHandler interface.
Most plugins will want to use SimpleScope rather than directly implementing Scope.

See the documentation of SimpleScopeHandler for methods to override.
*/
type SimpleScope struct {
	Scope

	ScopeName   string             // A name for this scope
	ScopeType   string             // The type of this scope
	ParentScope Scope              // The parent scope that created this scope; can be nil
	Wiring      WiringSpec         // The wiring spec
	Handler     SimpleScopeHandler // User-provided handler
	Seen        map[string]IRNode  // Cache of built nodes
	Added       map[string]any     // Nodes that have been passed to the handler
	Deferred    []func() error     // Deferred functions to execute

	stack []*WiringDef // Used when building; the stack of wiring defs currently being built
}

/*
Has four methods with default implementations that callers can override with custom logic:
  - LookupDef(name) - look up a WiringDef; default implementation directly consults the WiringSpec.
    callers can override this if they want to restrict, modify, or wrap definitions
    that get instantiated within this scope.
  - Accepts(nodeType) - should return true if the specified node type should be built within this scope,
    or false if we should ask the parent to build it instead.  Most scope implementations will only
    accept certain node types, and will thus want to override this method.  For example, a golang process
    will only accept golang nodes
  - AddNode(name, IRNode) - this is called when a node is created within this scope.  The SimpleScope
    internally saves the node for future lookups; callers might want to save the node e.g. as a child within
    a node that is being created.
  - AddEdge(name, IRNode) - this is called when a node was created by a parent scope but referenced within
    this scope.  The SimpleScope internally saves the node for future lookups; callers might want to save the
    node e.g. as an argument to the node that is being created
*/
type SimpleScopeHandler interface {
	Init(*SimpleScope)
	LookupDef(string) (*WiringDef, error)
	Accepts(any) bool
	AddEdge(string, IRNode) error
	AddNode(string, IRNode) error
}

type DefaultScopeHandler struct {
	SimpleScopeHandler
	Scope *SimpleScope

	Nodes map[string]IRNode
	Edges map[string]IRNode
}

func (handler *DefaultScopeHandler) Init(scope *SimpleScope) {
	handler.Scope = scope
	handler.Nodes = make(map[string]IRNode)
	handler.Edges = make(map[string]IRNode)
}

/*
Look up a WiringDef; default implementation directly consults the WiringSpec.

	callers can override this if they want to restrict, modify, or wrap definitions
	that get instantiated within this scope.
*/
func (handler *DefaultScopeHandler) LookupDef(name string) (*WiringDef, error) {
	def := handler.Scope.Wiring.GetDef(name)
	if def == nil {
		return nil, fmt.Errorf("%s does not exist in the wiring spec of scope %s", name, handler.Scope.Name())
	}
	return def, nil
}

/*
should return true if the specified node type should be built within this scope, or false if we should ask the parent to build it instead.  Most scope implementations will only

	accept certain node types, and will thus want to override this method.  For example, a golang process
	will only accept golang nodes
*/
func (handler *DefaultScopeHandler) Accepts(nodeType any) bool {
	return true
}

// This is called after getting a node from the parent scope.  By default it just saves the node
// as an edge.  Scope implementations can override this method to do other things.
func (handler *DefaultScopeHandler) AddEdge(name string, node IRNode) error {
	handler.Edges[node.Name()] = node
	return nil
}

// This is called after building a node in the current scope.  By default it just saves the node
// on the scope.  Scope implementations can override this method to do other things.
func (handler *DefaultScopeHandler) AddNode(name string, node IRNode) error {
	handler.Nodes[node.Name()] = node
	return nil
}

func (scope *SimpleScope) Init(name, scopetype string, parent Scope, wiring WiringSpec, handler SimpleScopeHandler) {
	scope.ScopeName = name
	scope.ScopeType = scopetype
	scope.ParentScope = parent
	scope.Wiring = wiring
	scope.Handler = handler
	scope.Seen = make(map[string]IRNode)
	scope.Added = make(map[string]any)
}
func (scope *SimpleScope) Name() string {
	return scope.ScopeName
}

func (scope *SimpleScope) Instantiate(name string) (IRNode, error) {
	return scope.get(name, false)
}

func (scope *SimpleScope) Get(name string) (IRNode, error) {
	return scope.get(name, true)
}

func (scope *SimpleScope) get(name string, addEdge bool) (IRNode, error) {
	// If it already exists, return it
	if node, ok := scope.Seen[name]; ok {
		return node, nil
	}

	// Look up the definition
	def, err := scope.Handler.LookupDef(name)
	if err != nil {
		return nil, err
	}

	// Track the defs being built
	scope.stack = append(scope.stack, def)
	defer func() {
		scope.stack = scope.stack[:len(scope.stack)-1]
	}()

	// If it's an alias, get the aliased node
	if def.Name != name {
		scope.Info("Resolved %s to %s", name, def.Name)
		node, err := scope.get(def.Name, addEdge)
		scope.Seen[name] = node
		return node, err
	}

	// See if the node should be created here or in the parent
	if !scope.Handler.Accepts(def.NodeType) {
		if scope.ParentScope == nil {
			return nil, scope.Error("Scope does not accept node %s of type %s but there is no parent scope to get them from", name, reflect.TypeOf(def.NodeType).String())
		}
		scope.Info("Getting %s of type %s from parent scope %s", name, reflect.TypeOf(def.NodeType).String(), scope.ParentScope.Name())
		var node IRNode
		if addEdge {
			node, err = scope.ParentScope.Get(name)
		} else {
			node, err = scope.ParentScope.Instantiate(name)
		}
		if err != nil {
			return nil, err
		}
		if _, already_added := scope.Added[node.Name()]; !already_added {
			if _, is_metadata := node.(IRMetadata); !is_metadata && addEdge {
				// Don't bother adding edges for metadata
				scope.Handler.AddEdge(name, node)
			}
			scope.Added[node.Name()] = true
		}
		scope.Seen[name] = node
		return node, nil
	}

	if def.Name == name {
		scope.Info("Building %s of type %s", name, reflect.TypeOf(def.NodeType).String())
	} else {
		scope.Info("Building %s (alias %s) of type %s", def.Name, name, reflect.TypeOf(def.NodeType).String())
	}

	// Build the node
	node, err := def.Build(scope)
	if err != nil {
		scope.Error("Unable to build %v: %s", name, err.Error())
		return nil, err
	}

	if _, already_added := scope.Added[node.Name()]; !already_added {
		scope.Handler.AddNode(name, node)
		scope.Added[node.Name()] = true
	}
	scope.Info("Finished building %s of type %s", name, reflect.TypeOf(node).String())
	scope.Seen[name] = node

	return node, nil
}

func (scope *SimpleScope) Put(name string, node IRNode) error {
	scope.Seen[name] = node

	if scope.Handler.Accepts(node) {
		scope.Handler.AddNode(name, node)
		scope.Info("%s of type %s added to scope", name, reflect.TypeOf(node).Elem().Name())
		return nil
	}

	if scope.ParentScope != nil {
		return scope.Error("%s of type %s does not belong in this scope, but cannot push to parent scope because no parent scope exists", name, reflect.TypeOf(node).Elem().Name())
	}

	scope.Info("%s of type %s does not belong in this scope; pushing to parent scope %s", name, reflect.TypeOf(node).Elem().Name(), scope.ParentScope)
	err := scope.ParentScope.Put(name, node)
	if err != nil {
		return err
	}
	scope.Handler.AddEdge(name, node)
	return err
}

func (scope *SimpleScope) Defer(f func() error) {
	if scope.ParentScope == nil {
		scope.Deferred = append(scope.Deferred, f)
	} else {
		scope.ParentScope.Defer(f)
	}
}

func (scope *SimpleScope) GetProperty(name string, key string) (any, error) {
	def, err := scope.Handler.LookupDef(name)
	if err != nil {
		return nil, err
	}
	return def.GetProperty(key), nil
}

func (scope *SimpleScope) GetProperties(name string, key string) ([]any, error) {
	def, err := scope.Handler.LookupDef(name)
	if err != nil {
		return nil, err
	}
	return def.GetProperties(key), nil
}

type blueprintScope struct {
	SimpleScope

	handler *blueprintScopeHandler
}

type blueprintScopeHandler struct {
	DefaultScopeHandler

	application *ApplicationNode
}

func newBlueprintScope(wiring WiringSpec, name string) (*blueprintScope, error) {
	scope := &blueprintScope{}
	handler := blueprintScopeHandler{}
	handler.Init(&scope.SimpleScope)
	handler.application = &ApplicationNode{}
	scope.handler = &handler
	scope.Init(name, "BlueprintApplication", nil, wiring, &handler)
	return scope, nil
}

func (scope *blueprintScope) Build() (IRNode, error) {
	node := &ApplicationNode{}
	node.name = scope.Name()

	// Execute deferred functions until empty
	for len(scope.Deferred) > 0 {
		next := scope.Deferred[0]
		scope.Deferred = scope.Deferred[1:]
		err := next()
		if err != nil {
			node.Children = scope.handler.Nodes
			return node, err
		}
	}
	node.Children = scope.handler.Nodes
	return node, nil
}

// Augments debug messages with information about the scope
func (scope *SimpleScope) Info(message string, args ...any) {
	if len(scope.stack) > 0 {
		src := scope.stack[len(scope.stack)-1]
		callstack := src.Properties["callsite"][0].(*WiringCallstack)
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", scope.ScopeType, scope.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s", scope.ScopeType, scope.Name(), message), args...))
	}
}

// Augments debug messages with information about the scope
func (scope *SimpleScope) Debug(message string, args ...any) {
	if len(scope.stack) > 0 {
		src := scope.stack[len(scope.stack)-1]
		callstack := src.Properties["callsite"][0].(*WiringCallstack)
		slog.Info(callstack.String())
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", scope.ScopeType, scope.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s", scope.ScopeType, scope.Name(), message), args...))
	}
}

// Augments debug messages with information about the scope
func (scope *SimpleScope) Error(message string, args ...any) error {
	formattedMessage := fmt.Sprintf(message, args...)
	if len(scope.stack) > 0 {
		src := scope.stack[len(scope.stack)-1]
		callstack := src.Properties["callsite"][0].(*WiringCallstack)
		slog.Error(fmt.Sprintf("%s %s: %s (%s)", scope.ScopeType, scope.Name(), formattedMessage, callstack.Stack[0].String()))
	} else {
		slog.Error(fmt.Sprintf("%s %s: %s", scope.ScopeType, scope.Name(), formattedMessage))
	}
	return fmt.Errorf(formattedMessage)
}
