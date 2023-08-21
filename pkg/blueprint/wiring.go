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

type Blueprint struct {
	applicationScope *blueprintScope
	wiring           *wiringSpecImpl
}

type BuildFunc func(Scope) (IRNode, error)

type WiringSpec interface {
	Define(name string, nodeType any, build BuildFunc) // Adds a named node definition to the spec that can be built with the provided build function
	GetDef(name string) *WiringDef                     // For use by plugins to access the defined build functions and metadata

	Alias(name string, pointsto string)   // Defines an alias to another defined node; these can be recursive
	GetAlias(alias string) (string, bool) // Gets the value of the specified alias, if it exists

	SetProperty(name string, key string, value any) // Sets a static property value in the wiring spec, replacing any existing value specified
	AddProperty(name string, key string, value any) // Adds a static property value in the wiring spec
	GetProperty(name string, key string) any        // Gets a static property value from the wiring spec
	GetProperties(name string, key string) []any    // Gets all static property values from the wiring spec

	String() string // Returns a string representation of everything that has been defined

	GetBlueprint() *Blueprint // After defining everything, this method provides the means to then build everything.
}

type WiringDef struct {
	Name       string
	NodeType   any
	Build      BuildFunc
	Properties map[string][]any
}

type wiringSpecImpl struct {
	WiringSpec
	name    string
	defs    map[string]*WiringDef
	aliases map[string]string
}

func NewWiringSpec(name string) WiringSpec {
	InitBlueprintCompilerLogging()

	spec := wiringSpecImpl{}
	spec.name = name
	spec.defs = make(map[string]*WiringDef)
	spec.aliases = make(map[string]string)
	return &spec
}

func (def *WiringDef) AddProperty(key string, value any) {
	def.Properties[key] = append(def.Properties[key], value)
}

func (def *WiringDef) GetProperty(key string) any {
	vs := def.Properties[key]
	if len(vs) == 1 {
		return vs[0]
	} else {
		return nil
	}
}

func (def *WiringDef) GetProperties(key string) []any {
	return def.Properties[key]
}

func (def *WiringDef) String() string {
	var b strings.Builder
	b.WriteString(def.Name)
	b.WriteString(" = ")
	b.WriteString(reflect.TypeOf(def.NodeType).Elem().Name())
	b.WriteString("(")
	var propStrings []string
	for propKey, values := range def.Properties {
		var propValues []string
		for _, v := range values {
			propValues = append(propValues, fmt.Sprintf("%s", v))
		}
		propStrings = append(propStrings, fmt.Sprintf("%s=%s", propKey, strings.Join(propValues, ",")))
	}
	b.WriteString(strings.Join(propStrings, ", "))
	b.WriteString(")")
	return b.String()
}

func (wiring *wiringSpecImpl) resolveAlias(alias string) string {
	for {
		name, is_alias := wiring.aliases[alias]
		if is_alias {
			alias = name
		} else {
			return alias
		}
	}
}

func (wiring *wiringSpecImpl) getDef(name string, createIfAbsent bool) *WiringDef {
	if def, ok := wiring.defs[name]; ok {
		return def
	} else if createIfAbsent {
		def := WiringDef{}
		def.Name = name
		def.Properties = make(map[string][]any)
		wiring.defs[name] = &def
		delete(wiring.aliases, name)
		return &def
	} else {
		return nil
	}
}

// Adds a named node to the spec that can be built with the provided build function.
// The nodeType is used as an indicator of where to build the node; the buildfunc is not required to actually return a node of that type
func (wiring *wiringSpecImpl) Define(name string, nodeType any, build BuildFunc) {
	def := wiring.getDef(name, true)
	def.NodeType = nodeType
	def.Build = build
}

// Primarily for use by plugins to build nodes; this will recursively resolve any aliases until a def is reached
func (wiring *wiringSpecImpl) GetDef(name string) *WiringDef {
	name = wiring.resolveAlias(name)
	return wiring.getDef(name, false)
}

// Defines an alias to another node.  Deletes any existing def for the alias
func (wiring *wiringSpecImpl) Alias(alias string, pointsto string) {
	_, exists := wiring.defs[alias]
	if exists {
		delete(wiring.defs, alias)
	}
	wiring.aliases[alias] = pointsto
}

// If the provided name is an alias, returns what it points to.
//
//	Otherwise returns the empty string and false
func (wiring *wiringSpecImpl) GetAlias(alias string) (string, bool) {
	name, exists := wiring.aliases[alias]
	return name, exists
}

// Sets a static value in the wiring spec, replacing any existing values for the specified key
func (wiring *wiringSpecImpl) SetProperty(name string, propKey string, propValue any) {
	def := wiring.getDef(name, true)
	def.Properties[propKey] = []any{propValue}

}

// Adds a static value to the wiring spec, appending it to any existing values for the specified key
func (wiring *wiringSpecImpl) AddProperty(name string, propKey string, propValue any) {
	def := wiring.getDef(name, true)
	def.Properties[propKey] = append(def.Properties[propKey], propValue)
}

// Primarily for use by plugins to get configuration values
func (wiring *wiringSpecImpl) GetProperty(name string, key string) any {
	def := wiring.getDef(name, false)
	if def != nil {
		return def.GetProperty(key)
	}
	return nil
}

// Primarily for use by plugins to get configuration values
func (wiring *wiringSpecImpl) GetProperties(name string, key string) []any {
	def := wiring.getDef(name, false)
	if def != nil {
		return def.GetProperties(key)
	}
	return nil
}

func (wiring *wiringSpecImpl) String() string {
	var defStrings []string
	for _, def := range wiring.defs {
		defStrings = append(defStrings, def.String())
	}
	for alias, pointsto := range wiring.aliases {
		defStrings = append(defStrings, alias+" -> "+pointsto)
	}
	return fmt.Sprintf("%s = WiringSpec {\n%s\n}", wiring.name, Indent(strings.Join(defStrings, "\n"), 2))
}

func (wiring *wiringSpecImpl) GetBlueprint() *Blueprint {
	blueprint := Blueprint{}

	scope, err := newBlueprintScope(wiring, wiring.name)
	if err != nil {
		slog.Error("Unable to build workflow spec, exiting", "error", err)
		os.Exit(1)
	}
	blueprint.applicationScope = scope
	blueprint.wiring = wiring
	return &blueprint
}

// Instantiates one or more specific named nodes
func (blueprint *Blueprint) Instantiate(names ...string) {
	for _, name := range names {
		nameToGet := name
		blueprint.applicationScope.Defer(func() error {
			blueprint.applicationScope.Info("Instantiating %v", nameToGet)
			_, err := blueprint.applicationScope.Get(nameToGet)
			return err
		})
	}
}

// Instantiates any nodes that haven't yet been instantiated.  Although this is commonly used,
// it is preferred to explicitly instantiate nodes by name.
func (blueprint *Blueprint) InstantiateAll() {
	for name, _ := range blueprint.wiring.defs {
		nameToGet := name
		blueprint.applicationScope.Defer(func() error {
			blueprint.applicationScope.Info("Instantiating %v", nameToGet)
			_, err := blueprint.applicationScope.Get(nameToGet)
			return err
		})
	}
}

func (blueprint *Blueprint) Build() (*ApplicationNode, error) {
	node, err := blueprint.applicationScope.Build()
	if err != nil {
		return nil, err
	}
	return node.(*ApplicationNode), err
}
