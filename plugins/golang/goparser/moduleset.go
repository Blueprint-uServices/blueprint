package goparser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
)

// Represents a set of code modules that have been parsed.
type ParsedModuleSet struct {
	Modules    map[string]*ParsedModule // Map from FQ module name to module object
	ModuleDirs map[string]*ParsedModule // Map from module SrcDir to module object
	Parent     *ParsedModuleSet         // Another module set to consult for modules if not present in this one
}

// Returns a new [*ParsedModuleSet], optionally with a parent module set, which can be nil.
func New(parent *ParsedModuleSet) *ParsedModuleSet {
	return &ParsedModuleSet{
		Modules:    make(map[string]*ParsedModule),
		ModuleDirs: make(map[string]*ParsedModule),
		Parent:     parent,
	}
}

// Adds a module to the parsed module set.  This is the preferred method for
// adding modules, versus [AddModule].
//
// info for a module can be acquired by calling [FindPackageModule] or [FindModule]
func (set *ParsedModuleSet) Add(info *ModuleInfo) (*ParsedModule, error) {
	mod, err := set.AddModule(info.Dir)
	if err != nil {
		return nil, err
	}
	mod.IsLocal = info.IsLocal
	return mod, nil
}

// Manually parse and add a module to the parsed module set
//
// If the srcDir has already been parsed, then this function will do nothing.
//
// If [set.Parent] is not nil, then the module will be copied from [set.Parent] rather than
// re-parsed.
//
// If [set] already contains a module with the same FQ module name as the one in srcDir then
// this function will return an error.
//
// The parsed module will be assumed to be a local module; if it is not, then set [IsLocal] to false
func (set *ParsedModuleSet) AddModule(srcDir string) (*ParsedModule, error) {
	if err := set.AddModules(srcDir); err != nil {
		return nil, err
	}
	return set.ModuleDirs[srcDir], nil
}

// Manually parse and add multiple modules to the set.
//
// Equivalent to calling [AddModule] for each srcDir. If a srcDir has already
// been parsed then it will not be re-parsed.
//
// Returns an error if any of the modules cannot be parsed.
func (set *ParsedModuleSet) AddModules(srcDirs ...string) error {
	var err error
	var newModules []*ParsedModule

	for _, srcDir := range srcDirs {
		srcDir = filepath.Clean(srcDir)

		// Have we parsed this module already?
		if _, exists := set.ModuleDirs[srcDir]; exists {
			continue
		}

		// Has a parent parsed this module already?
		var mod *ParsedModule
		for parent := set.Parent; parent != nil; parent = parent.Parent {
			if parentMod, existsInParent := parent.ModuleDirs[srcDir]; existsInParent {
				mod = parentMod
				break
			}
		}

		// Parse it
		if mod == nil {
			if mod, err = parseModule(srcDir); err != nil {
				return err
			}
		}

		// Does the same module exist in multiple different directories?
		if existingMod, exists := set.Modules[mod.Name]; exists {
			return blueprint.Errorf("redeclaration of module %v found in %v and %v", mod.Name, existingMod.SrcDir, srcDir)
		}

		newModules = append(newModules, mod)
	}

	// No errors encountered; save the new modules
	for _, mod := range newModules {
		set.ModuleDirs[mod.SrcDir] = mod
		set.Modules[mod.Name] = mod
	}
	return nil
}

// Parses and adds all modules in the specified workspaceDir
func (set *ParsedModuleSet) AddWorkspace(workspaceDir string) error {
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		return blueprint.Errorf("unable to read workspace directory %v due to %v", workspaceDir, err.Error())
	}

	var srcDirs []string
	for _, e := range entries {
		if e.IsDir() {
			srcDirs = append(srcDirs, filepath.Join(workspaceDir, e.Name()))
		}
	}
	return set.AddModules(srcDirs...)
}

// Gets the [*ParsedPackage] for the specified name, possibly searching
// for and parsing the package.
//
// This method will return an error if the package was not found, or if
// there was a parse error
func (set *ParsedModuleSet) GetPackage(name string) (*ParsedPackage, error) {
	// See if we've parsed the package
	for moduleName, mod := range set.Modules {
		if strings.HasPrefix(name, moduleName) {
			if pkg, exists := mod.Packages[name]; exists {
				return pkg, nil
			}
		}
	}

	// If we have a parent, look there
	if set.Parent != nil {
		return set.Parent.GetPackage(name)
	}

	// At this point we are the root parent, and the package isnt found.
	// Let's look for the package on the gopath.
	info, err := FindPackageModule(name)
	if err != nil {
		return nil, err
	}

	// We found it on the gopath, so let's parse the module
	mod, err := set.AddModule(info.Dir)
	if err != nil {
		return nil, err
	}
	mod.IsLocal = info.IsLocal

	// Return the package if it exists
	if pkg, exists := mod.Packages[name]; exists {
		return pkg, nil
	} else {
		return nil, blueprint.Errorf("package %v not found in module %v", name, mod.Name)
	}
}

// Looks up the specified struct, possibly searching for and parsing the package.
//
// Returns an error if the package cannot be found or parsed.
//
// Returns the [*ParsedStruct] if found, or nil if no such struct exists in the package.
func (set *ParsedModuleSet) FindStruct(pkgName string, name string) (*ParsedStruct, error) {
	pkg, err := set.GetPackage(pkgName)
	if err != nil {
		return nil, err
	}

	if struc, exists := pkg.Structs[name]; exists {
		return struc, nil
	}
	return nil, nil
}

// Looks up the specified interface, possibly searching for and parsing the package.
//
// Returns an error if the package cannot be found or parsed.
//
// Returns the [*ParsedInterface] if found, or nil if no such interface exists in the package.
func (set *ParsedModuleSet) FindInterface(pkgName string, name string) (*ParsedInterface, error) {
	pkg, err := set.GetPackage(pkgName)
	if err != nil {
		return nil, err
	}

	if intf, exists := pkg.Interfaces[name]; exists {
		return intf, nil
	}
	return nil, nil
}
