package gogen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Implements the ModuleBuilder interface defined in golang/ir.go

The ModuleBuilder is used by plugins that generate golang source files, and need a module to put that
code into.
*/
type ModuleBuilderImpl struct {
	ir.VisitTrackerImpl
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
		PackageName: module.Name + "/" + packageName,
		Name:        packageName,
		ShortName:   splits[len(splits)-1],
		Path:        filepath.Join(module.ModuleDir, filepath.Join(splits...)),
	}
	slog.Info(fmt.Sprintf("Creating package %v/%v", module.Name, packageName))
	return info, os.MkdirAll(info.Path, 0755)
}

func (module *ModuleBuilderImpl) Workspace() golang.WorkspaceBuilder {
	return module.workspace
}

func (module *ModuleBuilderImpl) Build(nodes []ir.IRNode) error {
	for _, node := range nodes {
		if n, valid := node.(golang.ProvidesInterface); valid {
			err := n.AddInterfaces(module)
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

func (module *ModuleBuilderImpl) ImplementsBuildContext() {}
