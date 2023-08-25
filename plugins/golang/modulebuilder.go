package golang

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
This struct is used by plugins if they want to generate Golang code and collect it into a module.

After creating a module builder, plugins can directly create directories and copy files into
the ModuleDir.  Any go dependencies should be added with the Require function.

When finished building the module, plugins should call Finish to finish building the go.mod
file
*/
type ModuleBuilder struct {
	VisitTracker
	ShortName string            // The shortname of this module
	Name      string            // The FQ name of this module
	Workspace *WorkspaceBuilder // The workspace that this module exists within
	ModuleDir string            // The directory containing this module
	Requires  map[string]string // required dependencies of this module and version
}

/*
This method is used by plugins if they want to generate their own Go module from scratch.

After dependencies and code have been added to the module, plugins must call Generate
on the Generated Module to finish building it.
*/
func NewModuleBuilder(workspace *WorkspaceBuilder, shortName string, moduleName string) (*ModuleBuilder, error) {
	if _, exists := workspace.Modules[shortName]; exists {
		return nil, fmt.Errorf("cannot generate new module %s (%s) as directory %s already exists in the generated workspace", shortName, moduleName, shortName)
	}
	if _, exists := workspace.ModuleDirs[moduleName]; exists {
		return nil, fmt.Errorf("cannot generate new module %s (%s) as module %s already exists in the generated workspace", shortName, moduleName, moduleName)
	}

	module := &ModuleBuilder{}
	module.visited = make(map[string]any)
	module.Name = moduleName
	module.ShortName = shortName
	module.ModuleDir = filepath.Join(workspace.WorkspaceDir, shortName)
	module.Requires = make(map[string]string)
	module.Workspace = workspace

	err := checkDir(module.ModuleDir, true)
	if err != nil {
		return nil, fmt.Errorf("cannot generate new module %s due to %s", shortName, err.Error())
	}

	workspace.ModuleDirs[moduleName] = shortName
	workspace.Modules[shortName] = moduleName

	return module, nil
}

/*
This method is used by plugins when contributing code to a generated module, to add any dependencies to the go.mod file.

When later generating the go.mod file, any dependencies that exist within the generated workspace will also have a `replace`
directive that points them to the local version of the module
*/
func (module *ModuleBuilder) Require(dependencyName string, version string) error {
	if existingVersion, exists := module.Requires[dependencyName]; exists {
		if existingVersion != version {
			return fmt.Errorf("module %s has two conflicting versions for dependency %s: %s and %s", module.Name, dependencyName, version, existingVersion)
		}
	} else {
		module.Requires[dependencyName] = version
	}
	return nil
}

func (module *ModuleBuilder) Finish() error {
	mod := `module ` + module.Name + `

go 1.19

`
	for reqName, reqVer := range module.Requires {
		mod += `require ` + reqName + ` ` + reqVer + "\n"
	}

	// TODO: add redirects to modules in same workspace dir, e.g.
	// replace gitlab.mpi-sws.org/cld/blueprint/plugins => ../../../plugins`

	modFileName := filepath.Join(module.ModuleDir, "go.mod")
	return os.WriteFile(modFileName, []byte(mod), 0755)
}
