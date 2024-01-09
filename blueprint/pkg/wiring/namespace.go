package wiring

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/logging"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
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

	// Creates and returns a child namespace within this namespaces.
	// handler will be used to determine what nodes can be built in the child namespace, and
	// handler's callbacks will be invoked when nodes get created within the child namespace.
	//
	// Subsequently, the namespace can be retrieved with GetNamespace
	//
	// Returns an error if the namespace has already been created
	DeriveNamespace(name string, handler NamespaceHandler) (Namespace, error)

	// Returns the child namespace with the given name.  The child namespace must have been created in this
	// namespace, using DeriveNamespace, otherwise an error will be returned.
	GetNamespace(name string) (Namespace, error)

	// Enqueue a function to be executed after all currently-queued functions have finished executing.
	// Most plugins should not need to use this.
	// [DeferOpts] can be optionally specified.
	Defer(f func() error, options ...DeferOpts)

	// Log an info-level message
	Info(message string, args ...any)

	// Log a warn-level message
	Warn(message string, args ...any)

	// Log an error-level message
	Error(message string, args ...any) error
}

// Options for deferred functions provided with [Namespace.Defer]
type DeferOpts struct {
	// Defaults to false. If set to true, pushes the deferred function to the front of the queue instead of the back.
	Front bool
}

var defaultDeferOpts = DeferOpts{
	Front: false,
}

// namespaceimpl is a base implementation of a [Namespace] that provides implementations of most methods.
//
// Most plugins that want to implement a [Namespace] will want to use [namespaceimpl] and only provide
// a [NamespaceHandler] implementation for a few of the custom namespace logics.
//
// See the documentation of [NamespaceHandler] for methods to implement.
type namespaceimpl struct {
	Namespace

	NamespaceName   string               // A name for this namespace
	NamespaceType   string               // The type of this namespace
	ParentNamespace Namespace            // The parent namespace that created this namespace; can be nil
	Wiring          WiringSpec           // The wiring spec
	Handler         NamespaceHandler     // User-provided handler
	Seen            map[string]ir.IRNode // Cache of built nodes
	Added           map[string]any       // Nodes that have been passed to the handler
	Deferred        []func() error       // Deferred functions to execute
	ChildNamespaces map[string]Namespace // Child namespaces

	stack []*WiringDef // Used when building; the stack of wiring defs currently being built
}

// NamespaceHandler is an interface intended for use by any Blueprint plugin that wants to
// provide a custom namespace.
type NamespaceHandler interface {
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

// Implements [Namespace]
func (namespace *namespaceimpl) Name() string {
	return namespace.NamespaceName
}

// Implements [Namespace]
func (namespace *namespaceimpl) Instantiate(name string, dst any) error {
	return namespace.get(name, false, dst)
}

// Implements [Namespace]
func (namespace *namespaceimpl) Get(name string, dst any) error {
	return namespace.get(name, true, dst)
}

func (namespace *namespaceimpl) lookupDef(name string) (*WiringDef, error) {
	def := namespace.Wiring.GetDef(name)
	if def == nil {
		return nil, blueprint.Errorf("%s does not exist in the wiring spec of namespace %s", name, namespace.NamespaceName)
	}
	return def, nil
}

func (namespace *namespaceimpl) get(name string, addEdge bool, dst any) error {
	// If it already exists, return it
	if node, ok := namespace.Seen[name]; ok {
		return copyResult(node, dst)
	}

	// Look up the definition
	def, err := namespace.lookupDef(name)
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
			if _, is_metadata := node.(ir.IRMetadata); !is_metadata && addEdge && !def.Options.ProxyNode {
				// Don't bother adding edges for metadata or proxy nodes, or instantiate (vs get) nodes
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
		if !def.Options.ProxyNode {
			namespace.Handler.AddNode(name, node)
		}
		namespace.Added[node.Name()] = true
	}
	namespace.Info("Finished building %s of type %s", name, reflect.TypeOf(node).String())
	namespace.Seen[name] = node
	return copyResult(node, dst)
}

// Implements [Namespace]
func (namespace *namespaceimpl) Put(name string, node ir.IRNode) error {
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

// Implements [Namespace]
func (namespace *namespaceimpl) Defer(f func() error, options ...DeferOpts) {
	opts := defaultDeferOpts
	if len(options) > 0 {
		opts = options[0]
	}
	if namespace.ParentNamespace == nil {
		if opts.Front {
			namespace.Deferred = append([]func() error{f}, namespace.Deferred...)
		} else {
			namespace.Deferred = append(namespace.Deferred, f)
		}
	} else {
		namespace.ParentNamespace.Defer(f, options...)
	}
}

// Implements [Namespace]
func (namespace *namespaceimpl) GetProperty(name string, key string, dst any) error {
	def, err := namespace.lookupDef(name)
	if err != nil {
		return err
	}
	return def.GetProperty(key, dst)
}

// Implements [Namespace]
func (namespace *namespaceimpl) GetProperties(name string, key string, dst any) error {
	def, err := namespace.lookupDef(name)
	if err != nil {
		return err
	}
	return def.GetProperties(key, dst)
}

// Implements [Namespace]
func (namespace *namespaceimpl) DeriveNamespace(name string, handler NamespaceHandler) (Namespace, error) {
	if _, exists := namespace.ChildNamespaces[name]; exists {
		return nil, namespace.Error("attempt to create child namespace %v that already exists", name)
	}

	child := &namespaceimpl{
		NamespaceName:   name,
		NamespaceType:   reflect.TypeOf(handler).Elem().Name(),
		ParentNamespace: namespace,
		Wiring:          namespace.Wiring,
		Handler:         handler,
		Seen:            make(map[string]ir.IRNode),
		Added:           make(map[string]any),
		ChildNamespaces: make(map[string]Namespace),
	}
	namespace.ChildNamespaces[name] = child
	namespace.Info("Created child namespace %v", name)
	return child, nil
}

// Implements [Namespace]
func (namespace *namespaceimpl) GetNamespace(name string) (Namespace, error) {
	if child, exists := namespace.ChildNamespaces[name]; exists {
		return child, nil
	}
	return nil, namespace.Error("child namespace %v does not exist", name)
}

// Implements [Namespace]
func (namespace *namespaceimpl) Info(message string, args ...any) {
	if len(namespace.stack) > 0 {
		src := namespace.stack[len(namespace.stack)-1]
		callstack := src.Properties["callsite"][0].(*logging.Callstack)
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", namespace.NamespaceType, namespace.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Info(fmt.Sprintf(fmt.Sprintf("%s %s: %s", namespace.NamespaceType, namespace.Name(), message), args...))
	}
}

// Implements [Namespace]
func (namespace *namespaceimpl) Debug(message string, args ...any) {
	if len(namespace.stack) > 0 {
		src := namespace.stack[len(namespace.stack)-1]
		callstack := src.Properties["callsite"][0].(*logging.Callstack)
		slog.Info(callstack.String())
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s (%s)", namespace.NamespaceType, namespace.Name(), message, callstack.Stack[0].String()), args...))
	} else {
		slog.Debug(fmt.Sprintf(fmt.Sprintf("%s %s: %s", namespace.NamespaceType, namespace.Name(), message), args...))
	}
}

// Implements [Namespace]
func (namespace *namespaceimpl) Error(message string, args ...any) error {
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
