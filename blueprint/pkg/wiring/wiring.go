package wiring

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/stringutil"
)

type BuildFunc func(Namespace) (ir.IRNode, error)

type WiringSpec interface {
	Define(name string, nodeType any, build BuildFunc) // Adds a named node definition to the spec that can be built with the provided build function
	GetDef(name string) *WiringDef                     // For use by plugins to access the defined build functions and metadata
	Defs() []string                                    // Returns names of all defined nodes

	Alias(name string, pointsto string)   // Defines an alias to another defined node; these can be recursive
	GetAlias(alias string) (string, bool) // Gets the value of the specified alias, if it exists

	SetProperty(name string, key string, value any)       // Sets a static property value in the wiring spec, replacing any existing value specified
	AddProperty(name string, key string, value any)       // Adds a static property value in the wiring spec
	GetProperty(name string, key string, dst any) error   // Gets a static property value from the wiring spec
	GetProperties(name string, key string, dst any) error // Gets all static property values from the wiring spec

	String() string // Returns a string representation of everything that has been defined

	// Errors while building a wiring spec are accumulated within the wiring spec, rather than as return values to calls
	AddError(err error) // Used by plugins to signify an error; the error will be returned by a call to Err or GetBlueprint
	Err() error         // Gets an error if there is currently one

	BuildIR(nodesToInstantiate ...string) (*ir.ApplicationNode, error) // After defining everything, this builds the IR for the specified named nodes (implicitly including dependencies of those nodes)
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
	errors  []error
}

func NewWiringSpec(name string) WiringSpec {
	spec := wiringSpecImpl{}
	spec.name = name
	spec.defs = make(map[string]*WiringDef)
	spec.aliases = make(map[string]string)
	spec.errors = nil
	return &spec
}

func (def *WiringDef) AddProperty(key string, value any) {
	def.Properties[key] = append(def.Properties[key], value)
}

func (def *WiringDef) GetProperty(key string, dst any) error {
	vs := def.Properties[key]
	if len(vs) == 1 {
		return copyResult(vs[0], dst)
	} else {
		return setZero(dst)
	}
}

func (def *WiringDef) GetProperties(key string, dst any) error {
	return copyResult(def.Properties[key], dst)
}

func (def *WiringDef) String() string {
	var b strings.Builder
	b.WriteString(def.Name)
	b.WriteString(" = ")
	b.WriteString(reflect.TypeOf(def.NodeType).Elem().Name())
	b.WriteString("(")
	var propStrings []string
	for propKey, values := range def.Properties {
		if propKey != "callsite" {
			var propValues []string
			for _, v := range values {
				propValues = append(propValues, fmt.Sprintf("%s", v))
			}
			propStrings = append(propStrings, fmt.Sprintf("%s=%s", propKey, strings.Join(propValues, ",")))
		}
	}
	b.WriteString(strings.Join(propStrings, ", "))
	b.WriteString(")")
	return b.String()
}

func (spec *wiringSpecImpl) resolveAlias(alias string) string {
	for {
		name, is_alias := spec.aliases[alias]
		if is_alias {
			alias = name
		} else {
			return alias
		}
	}
}

func (spec *wiringSpecImpl) getDef(name string, createIfAbsent bool) *WiringDef {
	if def, ok := spec.defs[name]; ok {
		return def
	} else if createIfAbsent {
		def := WiringDef{}
		def.Name = name
		def.Properties = make(map[string][]any)
		spec.defs[name] = &def
		delete(spec.aliases, name)
		return &def
	} else {
		return nil
	}
}

// Adds a named node to the spec that can be built with the provided build function.
// The nodeType is used as an indicator of where to build the node; the buildfunc is not required to actually return a node of that type
func (spec *wiringSpecImpl) Define(name string, nodeType any, build BuildFunc) {
	def := spec.getDef(name, true)
	def.NodeType = nodeType
	def.Build = build
	def.Properties["callsite"] = []any{blueprint.GetCallstack()}
}

// Primarily for use by plugins to build nodes; this will recursively resolve any aliases until a def is reached
func (spec *wiringSpecImpl) GetDef(name string) *WiringDef {
	name = spec.resolveAlias(name)
	return spec.getDef(name, false)
}

func (spec *wiringSpecImpl) Defs() []string {
	defs := make([]string, 0, len(spec.defs))
	for name := range spec.defs {
		defs = append(defs, name)
	}
	return defs
}

// Defines an alias to another node.  Deletes any existing def for the alias
func (spec *wiringSpecImpl) Alias(alias string, pointsto string) {
	_, exists := spec.defs[alias]
	if exists {
		delete(spec.defs, alias)
	}
	spec.aliases[alias] = pointsto
}

// If the provided name is an alias, returns what it points to.
//
//	Otherwise returns the empty string and false
func (spec *wiringSpecImpl) GetAlias(alias string) (string, bool) {
	name, exists := spec.aliases[alias]
	return name, exists
}

// Sets a static value in the wiring spec, replacing any existing values for the specified key
func (spec *wiringSpecImpl) SetProperty(name string, propKey string, propValue any) {
	def := spec.getDef(name, true)
	def.Properties[propKey] = []any{propValue}

}

// Adds a static value to the wiring spec, appending it to any existing values for the specified key
func (spec *wiringSpecImpl) AddProperty(name string, propKey string, propValue any) {
	def := spec.getDef(name, true)
	def.Properties[propKey] = append(def.Properties[propKey], propValue)
}

// Primarily for use by plugins to get configuration values
func (spec *wiringSpecImpl) GetProperty(name string, key string, dst any) error {
	def := spec.getDef(name, false)
	if def != nil {
		return def.GetProperty(key, dst)
	}
	return nil
}

// Primarily for use by plugins to get configuration values
func (spec *wiringSpecImpl) GetProperties(name string, key string, dst any) error {
	def := spec.getDef(name, false)
	if def != nil {
		return def.GetProperties(key, dst)
	}
	return nil
}

func (spec *wiringSpecImpl) String() string {
	var defStrings []string
	for _, def := range spec.defs {
		defStrings = append(defStrings, def.String())
	}
	for alias, pointsto := range spec.aliases {
		defStrings = append(defStrings, alias+" -> "+pointsto)
	}
	return fmt.Sprintf("%s = WiringSpec {\n%s\n}", spec.name, stringutil.Indent(strings.Join(defStrings, "\n"), 2))
}

func (spec *wiringSpecImpl) AddError(err error) {
	spec.errors = append(spec.errors, err)
}

type WiringError struct {
	Errors []error
}

func (e WiringError) Error() string {
	var errStrings []string
	for i, err := range e.Errors {
		errStrings = append(errStrings, fmt.Sprintf("Error %v: %v", i, err.Error()))
	}
	return strings.Join(errStrings, "\n")
}

func (spec *wiringSpecImpl) Err() error {
	if spec.errors == nil {
		return nil
	} else {
		return &WiringError{spec.errors}
	}
}

func (spec *wiringSpecImpl) BuildIR(nodesToInstantiate ...string) (*ir.ApplicationNode, error) {
	return BuildApplicationIR(spec, spec.name, nodesToInstantiate...)
}
