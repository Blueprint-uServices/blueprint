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

type WiringDef struct {
	Name     string
	Nodetype any
	Build    BuildFunc
}

type WiringSpec struct {
	name string
	defs map[string]WiringDef
}

type Namespace struct {
}

type Blueprint struct {
	ApplicationScope Scope
	Wiring           *WiringSpec
}

type BuildFunc func(Scope) (any, error)

func NewWiringSpec(name string) *WiringSpec {
	InitBlueprintCompilerLogging()

	spec := WiringSpec{}
	spec.name = name
	spec.defs = make(map[string]WiringDef)
	return &spec
}

// Adds a named node to the spec that can be built with the provided build function.
// When a node is built, it also returns the scope of the node (process, app level instance, etc.)
func (wiring *WiringSpec) Add(name string, nodeType any, build BuildFunc) {
	def := WiringDef{}
	def.Name = name
	def.Nodetype = nodeType
	def.Build = build
	wiring.defs[name] = def
}

func (wiring *WiringSpec) GetDef(name string) (any, BuildFunc) {
	if def, ok := wiring.defs[name]; ok {
		return def.Nodetype, def.Build
	}
	return nil, nil
}

func (wiring *WiringSpec) String() string {
	keys := []string{}
	for name, _ := range wiring.defs {
		keys = append(keys, name)
	}
	return strings.Join(keys, ", ")
}

func (wiring *WiringSpec) Blueprint() *Blueprint {
	blueprint := Blueprint{}

	scope, err := newBlueprintScope(wiring)
	if err != nil {
		slog.Error("Unable to build workflow spec, exiting", "error", err)
		os.Exit(1)
	}
	blueprint.ApplicationScope = scope
	blueprint.Wiring = wiring
	return &blueprint
}

// Instantiates one or more specific named nodes
func (blueprint *Blueprint) Instantiate(names ...string) {
	for _, name := range names {
		_, err := blueprint.ApplicationScope.Get(name)
		if err != nil {
			slog.Error("Unable to instantiate blueprint, exiting", "error", err)
			os.Exit(1)
		}
	}
}

// Instantiates any nodes that haven't yet been instantiated.  Although this is commonly used,
// it is preferred to explicitly instantiate nodes by name.
func (blueprint *Blueprint) InstantiateAll() {
	for name, _ := range blueprint.Wiring.defs {
		_, err := blueprint.ApplicationScope.Get(name)
		if err != nil {
			slog.Error("Unable to instantiate blueprint, exiting", "error", err)
			os.Exit(1)
		}
	}
}

func (blueprint *Blueprint) Build() (*ApplicationNode, error) {
	node, err := blueprint.ApplicationScope.Build()
	if err != nil {
		return nil, err
	}
	return node.(*ApplicationNode), err
}
