package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"golang.org/x/mod/modfile"
)

type ModuleBuilderImpl struct {
	ModuleBuilder
	tracker   irutil.VisitTrackerImpl
	ShortName string                // The shortname of this module
	Name      string                // The FQ name of this module
	workspace *WorkspaceBuilderImpl // The workspace that this module exists within
	ModuleDir string                // The directory containing this module
	Requires  map[string]string     // required dependencies of this module and version
	Replaces  map[string]string     // dependencies that need to be redirected to local workspace modules; this is calculated lazily
}

/*
This method is used by plugins if they want to generate their own Go module from scratch.

After dependencies and code have been added to the module, plugins must call Generate
on the Generated Module to finish building it.
*/
func NewModuleBuilder(workspace *WorkspaceBuilderImpl, shortName string, moduleName string) (*ModuleBuilderImpl, error) {
	if _, exists := workspace.Modules[shortName]; exists {
		return nil, fmt.Errorf("cannot generate new module %s (%s) as directory %s already exists in the generated workspace", shortName, moduleName, shortName)
	}
	if _, exists := workspace.ModuleDirs[moduleName]; exists {
		return nil, fmt.Errorf("cannot generate new module %s (%s) as module %s already exists in the generated workspace", shortName, moduleName, moduleName)
	}

	module := &ModuleBuilderImpl{}
	module.Name = moduleName
	module.ShortName = shortName
	module.ModuleDir = filepath.Join(workspace.WorkspaceDir, shortName)
	module.Requires = make(map[string]string)
	module.workspace = workspace

	err := checkDir(module.ModuleDir, true)
	if err != nil {
		return nil, fmt.Errorf("cannot generate new module %s due to %s", shortName, err.Error())
	}

	workspace.ModuleDirs[moduleName] = shortName
	workspace.Modules[shortName] = moduleName

	return module, nil
}

func (module *ModuleBuilderImpl) Workspace() WorkspaceBuilder {
	return module.workspace
}

func (module *ModuleBuilderImpl) Require(dependencyName string, version string) error {
	if version == "" {
		version = "v0.0.0"
	}
	if existingVersion, exists := module.Requires[dependencyName]; exists {
		if existingVersion != version {
			return fmt.Errorf("module %s requires two conflicting versions for dependency %s: %s and %s", module.Name, dependencyName, version, existingVersion)
		}
	} else {
		module.Requires[dependencyName] = version
	}
	return nil
}

func (module *ModuleBuilderImpl) Visited(name string) bool {
	return module.tracker.Visited(name)
}

var goModTemplate = `module {{.Name}}

go 1.20

{{ range $name, $version := .Requires }}require {{ $name }} {{ $version }}
{{ end }}
`

/*
This method should be used by plugins after all files have finished being added to the module

The method will do the following:
  - creates a go.mod file in the root of the module that includes all the required modules
*/
func (module *ModuleBuilderImpl) Finish() error {
	t, err := template.New("go.mod").Parse(goModTemplate)
	if err != nil {
		return err
	}

	// Create the go.mod file
	modFileName := filepath.Join(module.ModuleDir, "go.mod")
	f, err := os.OpenFile(modFileName, os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// Generate the file
	err = t.Execute(f, module)
	if err != nil {
		return fmt.Errorf("%v unable to generate go.mod file due to %v", module.ShortName, err.Error())
	}

	// Parse it to double check it is valid
	fWritten, err := os.ReadFile(modFileName)
	if err != nil {
		return err
	}
	_, err = modfile.Parse(modFileName, fWritten, nil)
	if err != nil {
		return fmt.Errorf("generated an invalid go.mod file for module %v due to %v", module.ShortName, err.Error())
	}

	return nil
}
