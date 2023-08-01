package blueprint

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func Wiring() {
	fmt.Println("Hello wiring!")

}

type WiringSpec struct {
	builders map[string]BuildFunc
}

type Blueprint struct {
	wiring *WiringSpec
	nodes  map[string]IRNode
}

type Namespace struct {
}

type BuildFunc func(*Blueprint) (string, interface{}, error)

func NewWiringSpec() *WiringSpec {
	InitBlueprintCompilerLogging()

	spec := WiringSpec{}
	spec.builders = make(map[string]BuildFunc)
	return &spec
}

// Adds a named node to the spec that can be built with the provided build function.
// When a node is built, it also returns the scope of the node (process, app level instance, etc.)
func (wiring *WiringSpec) Add(name string, build BuildFunc) {
	wiring.builders[name] = build
}

func (wiring *WiringSpec) String() string {
	keys := []string{}
	for name, _ := range wiring.builders {
		keys = append(keys, name)
	}
	return strings.Join(keys, ", ")
}

// Gets the named node, possibly building it if none has been built yet in the current blueprint
func (blueprint *Blueprint) Get(name string) (IRNode, error) {
	node, ok := blueprint.nodes[name]
	if ok {
		return node, nil
	}

	builder, ok := blueprint.wiring.builders[name]
	if !ok {
		return nil, fmt.Errorf("wiring spec doesn't contain \"%s\".  Known nodes: %s", name, blueprint.wiring)
	}

	slog.Info("Building", "node", name)
	_, inode, err := builder(blueprint) // TODO: support / do something with scope
	if err != nil {
		return nil, err
	}

	node, ok = inode.(IRNode)
	if !ok {
		// TODO: support e.g. configuration strings as well as nodes
		return nil, fmt.Errorf("lookup of node %s returned something that is not an IRNode (possibly unimplemented): %s", name, inode)
	}
	blueprint.nodes[name] = node

	return node, nil
}

// Print the IR graph
func (blueprint *Blueprint) String() string {
	var b strings.Builder
	for _, node := range blueprint.nodes {
		b.WriteString(node.String())
		b.WriteString("\n")
	}
	return b.String()
}

// Enters a namespace which will accumulate nodes, possibly exposing them globally, possibly not
// The namespace is automatically exited upon return from the invoked build function, or by calling
// Exit explicitly
func (blueprint *Blueprint) Enter(namespace Namespace) {
}

// Exits a namespace; if there was no enter call in this build function, then exit does nothing
func (blueprint *Blueprint) Exit() {
}

func (wiring *WiringSpec) Build() *Blueprint {
	blueprint := Blueprint{}
	blueprint.wiring = wiring
	blueprint.nodes = make(map[string]IRNode)

	var err error
	if err != nil {
		slog.Error("Unable to build workflow spec, exiting", "error", err)
		os.Exit(1)
	}
	return &blueprint
}

// Instantiates one or more specific named nodes
func (blueprint *Blueprint) Instantiate(names ...string) {
	for _, name := range names {
		_, err := blueprint.Get(name)
		if err != nil {
			slog.Error("Unable to instantiate blueprint, exiting", "error", err)
			os.Exit(1)
		}
	}
}

// Instantiates any nodes that haven't yet been instantiated.  Although this is commonly used,
// it is preferred to explicitly instantiate nodes by name.
func (blueprint *Blueprint) InstantiateAll() {
	for name, _ := range blueprint.wiring.builders {
		_, err := blueprint.Get(name)
		if err != nil {
			slog.Error("Unable to instantiate blueprint, exiting", "error", err)
			os.Exit(1)
		}
	}
}
