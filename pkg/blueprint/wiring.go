package blueprint

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"golang.org/x/exp/slog"
)

func Wiring() {
	fmt.Println("Hello wiring!")

}

type WiringDef struct {
	Name       string
	NodeType   any
	Build      BuildFunc
	Properties map[string][]any
}

func (def *WiringDef) AddProperty(key string, value any) {
	def.Properties[key] = append(def.Properties[key], value)
}

func (def *WiringDef) GetProperty(key string) []any {
	return def.Properties[key]
}

func (def *WiringDef) String() string {
	var b strings.Builder
	b.WriteString(def.Name)
	b.WriteString("(")
	b.WriteString(reflect.TypeOf(def.NodeType).Name())
	b.WriteString(")[")
	var propStrings []string
	for propKey, values := range def.Properties {
		var propValues []string
		for _, v := range values {
			propValues = append(propValues, fmt.Sprintf("%s", v))
		}
		propStrings = append(propStrings, fmt.Sprintf("%s=%s", propKey, strings.Join(propValues, ",")))
	}
	b.WriteString(strings.Join(propStrings, "; "))
	b.WriteString("]")
	return b.String()
}

type WiringSpec struct {
	name string
	defs map[string]*WiringDef
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
	spec.defs = make(map[string]*WiringDef)
	return &spec
}

func (wiring *WiringSpec) getDef(name string, createIfAbsent bool) *WiringDef {
	if def, ok := wiring.defs[name]; ok {
		return def
	} else if createIfAbsent {
		def := WiringDef{}
		def.Name = name
		def.Properties = make(map[string][]any)
		wiring.defs[name] = &def
		return &def
	} else {
		return nil
	}
}

// Adds a named node to the spec that can be built with the provided build function.
// When a node is built, it also returns the scope of the node (process, app level instance, etc.)
func (wiring *WiringSpec) Define(name string, nodeType any, build BuildFunc) {
	def := wiring.getDef(name, true)
	def.NodeType = nodeType
	def.Build = build
}

// Primarily for use by plugins to build nodes
func (wiring *WiringSpec) GetDef(name string) *WiringDef {
	return wiring.getDef(name, false)
}

// Adds a static value to the wiring spec, appending it to any existing values for the specified key
func (wiring *WiringSpec) AddProperty(name string, propKey string, propValue any) {
	def := wiring.getDef(name, true)
	def.Properties[propKey] = append(def.Properties[propKey], propValue)
}

// Primarily for use by plugins to get configuration values
func (wiring *WiringSpec) GetProperty(name string, key string) []any {
	def := wiring.getDef(name, false)
	if def != nil {
		return def.Properties[key]
	} else {
		return nil
	}
}

func (wiring *WiringSpec) String() string {
	var defStrings []string
	for _, def := range wiring.defs {
		defStrings = append(defStrings, def.String())
	}
	return fmt.Sprintf("%s = WiringSpec {\n%s\n}", wiring.name, Indent(strings.Join(defStrings, "\n"), 2))
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
