package workflowspec

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
)

// Representation of a parsed workflow spec.
//
// This code makes heavy use of the Golang code parser defined in the Golang plugin.  That
// code parser extracts structs, interfaces, and function definitions from a set of golang
// modules.
//
// This code adds functionality that:
//   - Identifies valid service interfaces
//   - Matches structs to interfaces that they implement
//   - Finds constructors of structs
type WorkflowSpec struct {
	Modules *goparser.ParsedModuleSet
}

// Returns a new [WorkflowSpec] for parsing workflow modules from scratch.
//
// Most plugins typically don't need to reference the workflow spec directly
// and can just call [GetService] or [GetServiceByName].
//
// Similarly most plugins shouldn't need to construct their own workflow
// spec from scratch, and instead should be able to make use of the existing
// cached one (through calling Get())
func New() *WorkflowSpec {
	return new(goparser.New(nil))
}

// Derives a [WorkflowSpec] from an existing one, making use of all of the
// modules already loaded in this workflow spec.
func (spec *WorkflowSpec) Derive() *WorkflowSpec {
	return new(goparser.New(spec.Modules))
}

func new(modules *goparser.ParsedModuleSet) *WorkflowSpec {
	spec := &WorkflowSpec{
		Modules: modules,
	}
	return spec
}

// Looks up the specified module, parses it, and adds it to the workflow spec.
func (spec *WorkflowSpec) ParseModule(moduleName string) error {
	mod, err := goparser.FindPackageModule(moduleName)
	if err != nil {
		return err
	}
	_, err = spec.Modules.Add(mod)
	return err
}

// Parses the specified module info and adds it to the workflow spec.
func (spec *WorkflowSpec) Parse(modInfo *goparser.ModuleInfo) error {
	_, err := spec.Modules.Add(modInfo)
	return err
}

// Looks up the named service in the workflow spec.  When a wiring spec instantiates a workflow spec
// service, this method will ultimately get called.
//
// Returns the service and a constructor
func (spec *WorkflowSpec) get(pkgName, name string, modInfo *goparser.ModuleInfo) (*Service, error) {
	// Parse the module
	mod, err := spec.Modules.Add(modInfo)
	if err != nil {
		return nil, err
	}

	// Find the package within the module
	pkg, pkgExists := mod.Packages[pkgName]
	if !pkgExists {
		return nil, err
	}

	// Return either the interface or struct definition
	if iface, hasIface := pkg.Interfaces[name]; hasIface {
		return spec.makeServiceFromInterface(iface)
	}
	if struc, hasStruc := pkg.Structs[name]; hasStruc {
		return spec.makeServiceFromStruct(struc)
	}
	return nil, blueprint.Errorf("unable to find service %v in workflow spec", name)
}
