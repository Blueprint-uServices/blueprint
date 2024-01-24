// Package golang implements the golang namespace used by Blueprint applications at
// runtime to instantiate golang nodes.
//
// A golang namespace takes care of the following:
//   - receives string arguments from the calling environment
//   - instantiates nodes that live in this namespace
package golang

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slog"
)

// Constructs a node. Within a namespace, a BuildFunc will only be called once,
// when somebody calls [Namespace.Get] for the named node.
//
// Namespaces reuse built nodes; subsequent calls to [Namespace.Get] will return
// the same built instance as the first invocation.
//
// node is a runtime instance such as a service, a wrapper class, etc.
//
// If node implements the [Runnable] interface then in addition to building the node.
// a namespace will also invoke [Runnable.Run] in a separate goroutine.
type BuildFunc func(n *Namespace) (node any, err error)

// If the return value of a [BuildFunc] implements the [Runnable] interface then
// the Namespace will automatically call [Runnable.Run] in a separate goroutine
type Runnable interface {
	// [Namespace] will call Run in a separate goroutine.
	Run(ctx context.Context) error
}

// The NamespaceBuilder is used at runtime by golang nodes to
// accumulate node definitions and configuration values for a namespace.
//
// Use the [NewNamespaceBuilder] function to create a namespace builder.
//
// Ultimately one of the Build methods should be
// called to create and start the namespace.
type NamespaceBuilder struct {
	name        string
	buildFuncs  map[string]BuildFunc
	required    map[string]*argNode
	optional    map[string]*argNode
	instantiate []string

	// The first error encountered while defining nodes on the builder.
	// Errors encountered by [NamespaceBuilder] are cached here, and then
	// returned when [NamespaceBuilder.Build] is called.
	err         error
	flagsparsed bool // flags only get parsed once
}

// A namespace from which nodes can be fetched by name.
//
// Nodes in the namespace can either be:
//   - argument nodes passed in to the namespace
//   - nodes built within this namespace.
//
// A namespace is constructed using the [NamespaceBuilder] struct, which
// can be created with the [NewNamespaceBuilder] method.
//
// The standard usage of a namespace is by a golang process.  Any
// golang services, wrappers, etc. are created using a Namespace.
//
// Some plugins, such as the ClientPool plugin, also use child Namespaces
// within the golang process Namespace.
type Namespace struct {
	name       string
	buildFuncs map[string]BuildFunc
	built      map[string]any

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	parent *Namespace
}

type argNode struct {
	name        string
	description string
	flag        *string
}

// Instantiates a new NamespaceBuilder.
//
// The NamespaceBuilder accumulates node and config variable definitions.
// Once all definitions are added, the Build* methods are used
// to build the namespace.
func NewNamespaceBuilder(name string) *NamespaceBuilder {
	b := &NamespaceBuilder{}
	b.name = name
	b.buildFuncs = make(map[string]BuildFunc)
	b.required = make(map[string]*argNode)
	b.optional = make(map[string]*argNode)
	b.instantiate = []string{}
	b.flagsparsed = false

	return b
}

// Sets a node to the specified value.
//
// Typically this is used for setting configuration or argument variables.
func (b *NamespaceBuilder) Set(name string, value string) {
	slog.Info(fmt.Sprintf("%v = %v", name, value))
	b.Define(name, func(n *Namespace) (any, error) { return value, nil })
}

// Defines a build function for a node that can be built within this namespace.
//
// name gives a name to the node that will be built.
//
// build is a [BuildFunc] for building the node.  build is lazily invoked
// when Get(name) is called on the [Namespace]
func (b *NamespaceBuilder) Define(name string, build BuildFunc) {
	if _, exists := b.buildFuncs[name]; exists {
		slog.Warn(fmt.Sprintf("%v redefining %v; this might indicate a bad wiring spec", b.name, name))
	}
	b.buildFuncs[name] = build
}

// A utility function to deterministically convert a string into a
// a valid linux environment variable name.  This is done by converting
// all punctuation characters to underscores, and converting alphabetic
// characters to uppercase (for convention), e.g.
//
//	a.grpc_addr becomes A_GRPC_ADDR.
//
// Punctuation is converted to underscores, and alpha are made uppercase.
func EnvVar(name string) string {
	return strings.ToUpper(cleanName(name))
}

var r = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// Returns name with only alphanumeric characters and all other
// symbols converted to underscores.
//
// CleanName is primarily used by plugins to convert user-defined
// service names into names that are valid as e.g. environment variables,
// command line arguments, etc.
func cleanName(name string) string {
	cleanName := r.ReplaceAllString(name, "_")
	for len(cleanName) > 0 {
		if _, err := strconv.Atoi(cleanName[0:1]); err != nil {
			return cleanName
		} else {
			cleanName = cleanName[1:]
		}
	}
	return cleanName
}

// Indicates that name is a required node.  When the namespace is built,
// an error will be returned if any required nodes are missing.
//
// The typical usage of this is to eagerly validate that all command line
// arguments have been provided.
func (b *NamespaceBuilder) Required(name string, description string) {
	b.required[name] = &argNode{
		name:        name,
		description: fmt.Sprintf("%s.  Can also be set with environment variable %s.", description, EnvVar(name)),
		flag:        flag.String(name, "", description),
	}
}

// Indicates that name is an optional node.  An error will only be returned
// if the caller attempts to build the node.
//
// The typical usage of this is when using only a single client from a client
// library
func (b *NamespaceBuilder) Optional(name string, description string) {
	b.optional[name] = &argNode{
		name:        name,
		description: fmt.Sprintf("%s.  Can also be set with environment variable %s.", description, EnvVar(name)),
		flag:        flag.String(name, "", description),
	}
}

// Indicates that name should be eagerly built when the namespace is built.
//
// The typical usage of this is to ensure that servers get started for
// namespaces that run servers.
func (b *NamespaceBuilder) Instantiate(name string) {
	b.instantiate = append(b.instantiate, name)
}

// Builds and returns the namespace.  This will:
//   - check that all required nodes have been defined
//   - parse command line arguments looking for missing required nodes
//   - build any nodes that were specified with [Instantiate]
//
// Returns a [Namespace] where nodes can now be gotten.
func (b *NamespaceBuilder) Build(ctx context.Context) (*Namespace, error) {
	// Return any error accumulated by the builder
	if b.err != nil {
		return nil, b.err
	}

	// Parse cmd line flags
	b.parseFlags()

	// Check required argnodes
	if err := b.checkRequired(nil); err != nil {
		return nil, err
	}

	// Create the namespace
	n := &Namespace{}
	n.name = b.name
	n.buildFuncs = make(map[string]BuildFunc)
	maps.Copy(n.buildFuncs, b.buildFuncs)
	n.built = make(map[string]any)
	n.ctx, n.cancel = context.WithCancel(ctx)
	n.wg = &sync.WaitGroup{}

	// Instantiate Normal nodes
	for _, name := range b.instantiate {
		var node any
		if err := n.Get(name, &node); err != nil {
			return nil, err
		}
	}

	// Graceful shutdown on interrupt
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for sig := range signals {
			slog.Info(fmt.Sprintf("received %v\n", sig))
			n.Shutdown(false)
		}
	}()

	return n, nil
}

// Builds and returns the namespace.  This will:
//   - check that all required nodes have been defined either in
//     this namespace or in the parent namespace(s)
//   - NOT parse command line arguments
//   - build any nodes that were specified with [Instantiate],
//     fetching missing nodes from the parent namespace
func (b *NamespaceBuilder) BuildWithParent(parent *Namespace) (*Namespace, error) {
	// Return any error accumulated by the builder
	if b.err != nil {
		return nil, b.err
	}

	// Check required argnodes
	if err := b.checkRequired(parent); err != nil {
		return nil, err
	}

	// Create the namespace
	n := &Namespace{}
	n.name = b.name
	n.parent = parent
	n.buildFuncs = make(map[string]BuildFunc)
	maps.Copy(n.buildFuncs, b.buildFuncs)
	n.built = make(map[string]any)
	n.ctx, n.cancel = context.WithCancel(parent.ctx)
	n.wg = &sync.WaitGroup{}

	// Instantiate nodes
	for _, name := range b.instantiate {
		var node any
		if err := n.Get(name, &node); err != nil {
			return nil, err
		}
	}

	// Don't install an interrupt handler here because it will be taken care of by the parent

	return n, nil
}

// Parse required arguments from flags
func (b *NamespaceBuilder) parseFlags() {
	if b.flagsparsed {
		return
	}
	b.flagsparsed = true
	if len(b.required) == 0 && len(b.optional) == 0 {
		return
	}

	flag.Parse()

	for _, node := range b.required {
		envValue := os.Getenv(EnvVar(node.name))
		if _, exists := b.buildFuncs[node.name]; exists {
			slog.Warn(fmt.Sprintf("Ignoring command line arg for %v", node.name))
		} else if *node.flag != "" {
			if envValue != "" && envValue != *node.flag {
				slog.Warn(fmt.Sprintf("Using command line argument %v=%v and ignoring environment variable %v=%v", node.name, *node.flag, EnvVar(node.name), envValue))
			}
			b.Set(node.name, *node.flag)
		} else if envValue != "" {
			b.Set(node.name, envValue)
		}
	}

	for _, node := range b.optional {
		envValue := os.Getenv(EnvVar(node.name))
		if _, exists := b.buildFuncs[node.name]; exists {
			slog.Warn(fmt.Sprintf("Ignoring command line arg for %v\n", node.name))
		} else if *node.flag != "" {
			if envValue != "" && envValue != *node.flag {
				slog.Warn(fmt.Sprintf("Using command line argument %v=%v and ignoring environment variable %v=%v", node.name, *node.flag, EnvVar(node.name), envValue))
			}
			b.Set(node.name, *node.flag)
		} else if envValue != "" {
			b.Set(node.name, envValue)
		} else {
			name := node.name
			b.Define(node.name, func(n *Namespace) (any, error) {
				return nil, fmt.Errorf("Required argument %v is not set", name)
			})
		}
	}
}

func (b *NamespaceBuilder) checkRequired(parent *Namespace) error {
	missing := []string{}
	for _, node := range b.required {
		if _, exists := b.buildFuncs[node.name]; exists {
			continue
		}
		if parent != nil && parent.has(node.name) {
			continue
		}
		missing = append(missing, node.name)
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required argnodes [%v]", strings.Join(missing, ", "))
	}
	return nil
}

// Reports whether the namespace has a definition for name
func (n *Namespace) has(name string) bool {
	if _, isBuilt := n.built[name]; isBuilt {
		return true
	}
	if _, hasDef := n.buildFuncs[name]; hasDef {
		return true
	}
	if n.parent != nil {
		return n.parent.has(name)
	}
	return false
}

// Gets a node from this namespace.  If the node hasn't been built yet,
// it will be built.
func (n *Namespace) Get(name string, receiver any) error {
	if existing, exists := n.built[name]; exists {
		return backend.CopyResult(existing, receiver)
	}
	if build, exists := n.buildFuncs[name]; exists {
		built, err := build(n)
		if err != nil {
			slog.Error(fmt.Sprintf("%v error building %v", n.name, name))
			return err
		} else {
			switch v := built.(type) {
			case string:
				slog.Info(fmt.Sprintf("%v built %v (%v) = %v", n.name, name, reflect.TypeOf(built), v))
			default:
				slog.Info(fmt.Sprintf("%v built %v (%v)", n.name, name, reflect.TypeOf(built)))
			}
		}
		n.built[name] = built

		if runnable, isRunnable := built.(Runnable); isRunnable {
			slog.Info(fmt.Sprintf("%v running %v", n.name, name))
			n.wg.Add(1)
			if n.parent != nil {
				n.parent.wg.Add(1)
			}
			go func() {
				err := runnable.Run(n.ctx)
				if err != nil {
					slog.Error(fmt.Sprintf("%v error running node %v: %v", n.name, name, err.Error()))
					n.cancel()
				} else {
					slog.Info(fmt.Sprintf("%v %v exited", n.name, name))
				}
				n.wg.Done()
				if n.parent != nil {
					n.parent.wg.Done()
				}
			}()
		}

		return backend.CopyResult(built, receiver)
	}
	if n.parent != nil {
		return n.parent.Get(name, receiver)
	}
	return fmt.Errorf("%v unknown %v", n.name, name)
}

// ctx can be used by any [BuildFunc] that wants to start background goroutines,
// perform a blocking select, etc.
//
// ctx will be notified on the Done channel if the namespace is shutdown during blocking.
func (n *Namespace) Context() (ctx context.Context) {
	return n.ctx
}

// Stops any nodes (e.g. servers) that are running in this namespace.
func (n *Namespace) Shutdown(awaitCompletion bool) {
	n.cancel()
	if awaitCompletion {
		n.Await()
	}
}

// If any nodes in this namespace are running goroutines, waits for them to finish
func (n *Namespace) Await() {
	n.wg.Wait()
}
