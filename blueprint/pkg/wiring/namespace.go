package wiring

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/logging"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"golang.org/x/exp/slog"
)

// Namespace is a dependency injection container used by Blueprint plugins during Blueprint's IR construction process.
// Namespaces instantiate and store IRNodes.  A root Blueprint application is itself a Namespace.
//
// A Namespace argument is passed to the [BuildFunc] when an IRNode is being built.  An IRNode can potentially
// be built multiple times, in different namespaces.
//
// If an IRNode depends on other IRNodes, those others can be fetched by calling [Namespace.Get].  If those IRNodes
// haven't yet been built, then their BuildFuncs will also be invoked, recursively.  Conversely, if those IRNodes
// are already built, then the built instance is re-used.
//
// Namespaces are hierarchical and a namespace implementation can choose to only support a subset of IRNodes.
// In this case, [Namespace.Get] on an unsupported IRNode will recursively get the node on the parent namespace.
// Namespaces inspect the nodeType argument of [WiringSpec.Define] to make this decision.
type Namespace interface {
	// Returns the name of this namespace
	Name() string

	// Gets an IRNode with the specified name from this namespace, placing the result in the pointer dst.
	// dst should typically be a pointer to an IRNode type.
	// If the node has already been built, it returns the existing built node.
	// If the node hasn't yet been built, the node's [BuildFunc] will be called and the result will be
	// cached and returned.
	// This call might recursively call [Get] on a parent namespace depending on the [nodeType] registered
	// for name.
	Get(name string, dst any) error

	// The same as [Get] but without creating a depending (an edge) into the current namespace.  Most
	// plugins should use [Get] instead.
	Instantiate(name string, dst any) error

	// Gets a property from the wiring spec; dst should be a pointer to a value
	GetProperty(name string, key string, dst any) error

	// Gets a slice of properties from the wiring spec; dst should be a pointer to a slice
	GetProperties(name string, key string, dst any) error

	// Puts a node into this namespace
	Put(name string, node ir.IRNode) error

	// Enqueue a function to be executed after all currently-queued functions have finished executing.
	// Most plugins should not need to use this.
	Defer(f func() error)

	// Log an info-level message
	Info(message string, args ...any)

	// Log a warn-level message
	Warn(message string, args ...any)

	// Log an error-level message
	Error(message string, args ...any) error
}

// SimpleNamespace is a base implementation of a [Namespace] that provides implementations of most methods.
//
// Most plugins that want to implement a [Namespace] will want to use [SimpleNamespace] and only provide
// a [SimpleNamespaceHandler] implementation for a few of the custom namespace logics.
//
// See the documentation of [SimpleNamespaceHandler] for methods to implement.
type SimpleNamespace struct {
	Namespace

	NamespaceName   string                 // A name for this namespace
	NamespaceType   string                 // The type of this namespace
	ParentNamespace Namespace              // The parent namespace that created this namespace; can be nil
	Wiring          WiringSpec             // The wiring spec
	Handler         SimpleNamespaceHandler // User-provided handler
	Seen            map[string]ir.IRNode   // Cache of built nodes
	Added           map[string]any         // Nodes that have been passed to the handler
	Deferred        []func() error         // Deferred functions to execute

	stack []*WiringDef // Used when building; the stack of wiring defs currently being built
}

// SimpleNamespaceHandler is an interface intended for use by any Blueprint plugin that wants to
// provide a custom namespace.
//
// The plugin should implement the methods of this handler and then create a [SimpleNamespace] and call
// [SimpleNamespace.Init]
type SimpleNamespaceHandler interface {
	// Initialize the handler with a namespace
	Init(*SimpleNamespace)

	// Look up a [WiringDef] from the wiring spec.
	LookupDef(string) (*WiringDef, error)

	// Reports true if this namespace can build nodes of the specified node type.
	//
	// For some node type T, if Accepts(T) returns false, then nodes of type T will
	// not be built in this namespace and instead the parent namespace will be called.
	Accepts(any) bool

	// After a node has been gotten from the parent namespace, AddEdge will be
	// called to inform the handler that the node should be passed to this namespace
	// as an argument.
	AddEdge(string, ir.IRNode) error

	// After a node has been built in this namespace, AddNode will be called
	// to enable the handler to save the built node.
	AddNode(string, ir.IRNode) error
}

// A basic [SimpleNamespaceHandler] implementation that accepts nodes of all types.
type DefaultNamespaceHandler struct {
	SimpleNamespaceHandler
	Namespace *SimpleNamespace

	Nodes []ir.IRNode
	Edges []ir.IRNode
}

func (handler *DefaultNamespaceHandler) Init(namespace *SimpleNamespace) {
	handler.Namespace = namespace
}

func (handler *DefaultNamespaceHandler) LookupDef(name string) (*WiringDef, error) {
	def := handler.Namespace.Wiring.GetDef(name)
	if def == nil {
		return nil, blueprint.Errorf("%s does not exist in the wiring spec of namespace %s", name, handler.Namespace.Name())
	}
	return def, nil
}

func (handler *DefaultNamespaceHandler) Accepts(nodeType any) bool {
	return true
}

func (handler *DefaultNamespaceHandler) AddEdge(name string, node ir.IRNode) error {
	handler.Edges = append(handler.Edges, node)
	return nil
}

func (handler *DefaultNamespaceHandler) AddNode(name string, node ir.IRNode) error {
	handler.Nodes = append(handler.Nodes, node)
	return nil
}

// Initializes a SimpleNamespace.  To do so, a parent namespace, wiring spec,
// and [SimpleNamespaceHandler] implementation must be provided.
func (namespace *SimpleNamespace) Init(name, namespacetype string, parent Namespace, wiring WiringSpec, handler SimpleNamespaceHandler) {
	namespace.NamespaceName = name
	namespace.NamespaceType = namespacetype
	namespace.ParentNamespace = parent
	namespace.Wiring = wiring
	namespace.Handler = handler
	namespace.Seen = make(map[string]ir.IRNode)
	namespace.Added = make(map[string]any)
}

func (namespace *SimpleNamespace) Name() string {
	return namespace.NamespaceName
}

func (namespace *SimpleNamespace) Instantiate(name string, dst any) error {
	return namespace.get(name, false, dst)
}

func (namespace *SimpleNamespace) Get(name string, dst any) error {
	return namespace.get(name, true, dst)
}

func (namespace *SimpleNamespace) get(name string, addEdge bool, dst any) error {
	// If it already exists, return it
	if node, ok := namespace.Seen[name]; ok {
		return copyResult(node, dst)
	}

	// Look up the definition
	def, err := namespace.Handler.LookupDef(name)
	if err != nil {
		return err
	}

	// Track the defs being built
	namespace.stack = append(namespace.stack, def)
	defer func() {
		namespace.stack = namespace.stack[:len(namespace.stack)-1]
	}()

	// If it's an alias, get the aliased node
	if def.Name != name {
		namespace.Info("Resolved %s to %s", name, def.Name)
		var node ir.IRNode
		err := namespace.get(def.Name, addEdge, &node)
		namespace.Seen[name] = node
		if err != nil {
			return err
		}
		return copyResult(node, dst)
	}

	// See if the node should be created here or in the parent
	if !namespace.Handler.Accepts(def.NodeType) {
		if namespace.ParentNamespace == nil {
			return namespace.Error("Namespace does not accept node %s of type %s but there is no parent namespace to get them from", name, reflect.TypeOf(def.NodeType).String())
		}
		namespace.Info("Getting %s of type %s from parent namespace %s", name, reflect.TypeOf(def.NodeType).String(), namespace.ParentNamespace.Name())
		var node ir.IRNode
		if addEdge {
			err = namespace.ParentNamespace.Get(name, &node)
		} else {
			err = namespace.ParentNamespace.Instantiate(name, &node)
		}
		if err != nil {
			return err
		}
		if _, already_added := namespace.Added[node.Name()]; !already_added {
			if _, is_metadata := node.(ir.IRMetadata); !is_metadata && addEdge {
				// Don't bother adding edges for metadata
				namespace.Handler.AddEdge(name, node)
			}
			namespace.Added[node.Name()] = true
		}
		namespace.Seen[name] = node
		return copyResult(node, dst)
	}

	if def.Name == name {
		namespace.Info("Building %s of type %s", name, reflect.TypeOf(def.NodeType).String())
	} else {
		namespace.Info("Building %s (alias %s) of type %s", def.Name, name, reflect.TypeOf(def.NodeType).String())
	}

	// Build the node
	node, err := def.Build(namespace)
	if err != nil {
		namespace.Error("Unable to build %v: %s", name, err.Error())
		return err
	}

	if _, already_added := namespace.Added[node.Name()]; !already_added {
		namespace.Handler.AddNode(name, node)
		namespace.Added[node.Name()] = true
	}
	namespace.Info("Finished building %s of type %s", name, reflect.TypeOf(node).String())
	namespace.Seen[name] = node
	return copyResult(node, dst)
}

func (namespace *SimpleNamespace) Put(name string, node ir.IRNode) error {
	namespace.Seen[name] = node

	if namespace.Handler.Accepts(node) {
		namespace.Handler.AddNode(name, node)
		namespace.Info("%s of type %s added to namespace", name, reflect.TypeOf(node).Elem().Name())
		return nil
	}

	if namespace.ParentNamespace != nil {
		return namespace.Error("%s of type %s does not belong in this namespace, but cannot push to parent namespace because no parent namespace exists", name, reflect.TypeOf(node).Elem().Name())
	}

	namespace.Info("%s of type %s does not belong in this namespace; pushing to parent namespace %s", name, reflect.TypeOf(node).Elem().Name(), namespace.ParentNamespace)
	err := namespace.ParentNamespace.Put(name, node)
	if err != nil {
		return err
	}
	namespace.Handler.AddEdge(name, node)
	return err
}

func (namespace *SimpleNamespace) Defer(f func() error) {
	if namespace.ParentNamespace == nil {
		namespace.Deferred = append(namespace.Deferred, f)
	} else {
		namespace.ParentNamespace.Defer(f)
	}
}

func (namespace *SimpleNamespace) GetProperty(name string, key string, dst any) error {
	def, err := namespace.Handler.LookupDef(name)
	if err != nil {
		return err
	}
	return def.GetProperty(key, dst)
}

func (namespace *SimpleNamespace) GetProperties(name string, key string, dst any) error {
	def, err := namespace.Handler.LookupDef(name)
	if err != nil {
		return err
	}
	return def.GetProperties(key, dst)
}

func (namespace *SimpleNamespace) Info(message string, args ...any) {
	if len(namespace.stack) > 0 {
		src := namespace.stack[len(namespace.stack)-1]
		callstack := src.Properties["callsite"][0].(*logging.Callstack)
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", namespace.NamespaceType, namespace.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s", namespace.NamespaceType, namespace.Name(), message), args...))
	}
}

func (namespace *SimpleNamespace) Debug(message string, args ...any) {
	if len(namespace.stack) > 0 {
		src := namespace.stack[len(namespace.stack)-1]
		callstack := src.Properties["callsite"][0].(*logging.Callstack)
		slog.Info(callstack.String())
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", namespace.NamespaceType, namespace.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s", namespace.NamespaceType, namespace.Name(), message), args...))
	}
}

func (namespace *SimpleNamespace) Error(message string, args ...any) error {
	formattedMessage := fmt.Sprintf(message, args...)
	if len(namespace.stack) > 0 {
		src := namespace.stack[len(namespace.stack)-1]
		callstack := src.Properties["callsite"][0].(*logging.Callstack)
		slog.Error(fmt.Sprintf("%s %s: %s (%s)", namespace.NamespaceType, namespace.Name(), formattedMessage, callstack.Stack[0].String()))
	} else {
		slog.Error(fmt.Sprintf("%s %s: %s", namespace.NamespaceType, namespace.Name(), formattedMessage))
	}
	return fmt.Errorf(formattedMessage)
}
