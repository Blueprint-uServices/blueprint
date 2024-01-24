package gogen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
)

// Implements [golang.ModuleBuilder].
//
// Creates a module on the local filesystem with a user-provided module name.
type ModuleBuilderImpl struct {
	ir.VisitTrackerImpl
	Name      string                // The FQ name of this module
	workspace *WorkspaceBuilderImpl // The workspace that this module exists within
	ModuleDir string                // The directory containing this module

	modfile *modfile.File // The parsed go.mod file for this module
}

// Creates a module within the provided workspace, and returns a [ModuleBuilderImpl] that
// can be used to accumulate code in the created module.
//
// The typical usage of this is by plugins such as the [goproc] plugin that accumulate
// golang nodes and generate code to run those nodes.
//
// After calling this method, the returned ModuleBuilder can be passed to golang nodes,
// to accumulate the interfaces and funcs of those nodes.
//
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
func NewModuleBuilder(workspace *WorkspaceBuilderImpl, moduleName string) (*ModuleBuilderImpl, error) {
	moduleDir, err := workspace.CreateModule(moduleName, "v0.0.0")
	if err != nil {
		return nil, err
	}

	module := &ModuleBuilderImpl{}
	module.Name = moduleName
	module.ModuleDir = moduleDir
	module.workspace = workspace
	module.modfile, err = loadmodfile(moduleDir)
	return module, err
}

// Implements [golang.ModuleBuilder]
func (module *ModuleBuilderImpl) Info() golang.ModuleInfo {
	return golang.ModuleInfo{
		Name:    module.Name,
		Version: "v0.0.0",
		Path:    module.ModuleDir,
	}
}

// Implements [golang.ModuleBuilder]
func (module *ModuleBuilderImpl) CreatePackage(packageName string) (golang.PackageInfo, error) {
	if packageName == "" {
		packageName = "main"
	}
	info := golang.PackageInfo{}
	if packageName == "main" {
		info = golang.PackageInfo{
			PackageName: module.Name,
			Name:        "main",
			ShortName:   "main",
			Path:        module.ModuleDir,
		}
	} else {
		splits := strings.Split(packageName, "/")
		info = golang.PackageInfo{
			PackageName: module.Name + "/" + packageName,
			Name:        packageName,
			ShortName:   splits[len(splits)-1],
			Path:        filepath.Join(module.ModuleDir, filepath.Join(splits...)),
		}
	}
	if s, err := os.Stat(info.Path); err == nil && s.IsDir() {
		// Package already exists
		return info, nil
	} else {
		slog.Info(fmt.Sprintf("Creating package %v/%v", module.Name, packageName))
		return info, os.MkdirAll(info.Path, 0755)
	}
}

// Implements [golang.ModuleBuilder]
func (module *ModuleBuilderImpl) Require(moduleName string, version string) error {
	if version == "" {
		return blueprint.Errorf("%s go.mod require needs a version for %s", module.Name, moduleName)
	}
	slog.Info(fmt.Sprintf("require %s %s", moduleName, version))
	module.modfile.AddNewRequire(moduleName, version, false)
	return savemodfile(module.modfile, filepath.Join(module.ModuleDir, "go.mod"))
}

// Implements [golang.ModuleBuilder]
func (module *ModuleBuilderImpl) Workspace() golang.WorkspaceBuilder {
	return module.workspace
}

func (module *ModuleBuilderImpl) ImplementsBuildContext() {}

func loadmodfile(moduleDir string) (*modfile.File, error) {
	modfileName := filepath.Join(moduleDir, "go.mod")
	modfileData, err := os.ReadFile(modfileName)
	if err != nil {
		return nil, blueprint.Errorf("unable to read go.mod %s due to %s", modfileName, err.Error())
	}

	mod, err := modfile.Parse(modfileName, modfileData, nil)
	if err != nil {
		return nil, blueprint.Errorf("unable to parse go.mod %s due to %s", modfileName, err.Error())
	}

	return mod, nil
}

func savemodfile(f *modfile.File, filename string) error {
	bytes, err := f.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0644)
}
