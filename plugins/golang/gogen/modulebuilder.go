package gogen

import (
	"os"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

/*
Implements the ModuleBuilder interface defined in golang/ir.go

The ModuleBuilder is used by plugins that generate golang source files, and need a module to put that
code into.
*/
type ModuleBuilderImpl struct {
	golang.ModuleBuilder
	tracker   irutil.VisitTrackerImpl
	Name      string                // The FQ name of this module
	workspace *WorkspaceBuilderImpl // The workspace that this module exists within
	ModuleDir string                // The directory containing this module
}

/*
This method is used by plugins if they want to generate their own Go module from scratch.

After dependencies and code have been added to the module, plugins must call Generate
on the Generated Module to finish building it.
*/
func NewModuleBuilder(workspace *WorkspaceBuilderImpl, moduleName string) (*ModuleBuilderImpl, error) {
	moduleDir, err := workspace.CreateModule(moduleName, "v0.0.0")
	if err != nil {
		return nil, err
	}

	module := &ModuleBuilderImpl{}
	module.Name = moduleName
	module.ModuleDir = moduleDir
	module.workspace = workspace
	return module, nil
}

func (module *ModuleBuilderImpl) Workspace() golang.WorkspaceBuilder {
	return module.workspace
}

func (module *ModuleBuilderImpl) Info() golang.ModuleInfo {
	return golang.ModuleInfo{
		Name:    module.Name,
		Version: "v0.0.0",
		Path:    module.ModuleDir,
	}
}

func (module *ModuleBuilderImpl) CreatePackage(packageName string) (golang.PackageInfo, error) {
	splits := strings.Split(packageName, "/")
	info := golang.PackageInfo{
		Name:      packageName,
		ShortName: splits[len(splits)-1],
		Path:      filepath.Join(module.ModuleDir, filepath.Join(splits...)),
	}
	return info, os.MkdirAll(info.Path, 0755)
}

func (module *ModuleBuilderImpl) Visit(nodes []blueprint.IRNode) error {
	for _, node := range nodes {
		if n, valid := node.(golang.GeneratesInterfaces); valid {
			err := n.GenerateInterfaces(module)
			if err != nil {
				return err
			}
		}
	}
	for _, node := range nodes {
		if n, valid := node.(golang.GeneratesFuncs); valid {
			err := n.GenerateFuncs(module)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (module *ModuleBuilderImpl) Visited(name string) bool {
	return module.tracker.Visited(name)
}
