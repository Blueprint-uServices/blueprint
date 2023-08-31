package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	cp "github.com/otiai10/copy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"golang.org/x/mod/modfile"
)

type WorkspaceBuilderImpl struct {
	WorkspaceBuilder
	tracker      irutil.VisitTrackerImpl
	WorkspaceDir string            // The directory containing this workspace
	ModuleDirs   map[string]string // map from FQ module name to directory name within WorkspaceDir
	Modules      map[string]string // map from directory name to FQ module name within WorkspaceDir
}

/*
Creates a new WorkspaceBuilder at the specified output dir.

Will return an error if the workspacedir already exists
*/
func NewWorkspaceBuilder(workspaceDir string) (*WorkspaceBuilderImpl, error) {
	if isDir(workspaceDir) {
		return nil, fmt.Errorf("workspace %s already exists", workspaceDir)
	}
	err := os.Mkdir(workspaceDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace %s due to %s", workspaceDir, err.Error())
	}
	workspace := &WorkspaceBuilderImpl{}
	workspace.WorkspaceDir = workspaceDir
	workspace.ModuleDirs = make(map[string]string)
	workspace.Modules = make(map[string]string)
	return workspace, nil
}

func (workspace *WorkspaceBuilderImpl) AddLocalModule(shortName string, moduleSrcPath string) error {
	// First open and parse the go.mod file to make sure it exists and is valid
	modfileName := filepath.Join(moduleSrcPath, "go.mod")
	modfileData, err := os.ReadFile(modfileName)
	if err != nil {
		return fmt.Errorf("unable to read go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	mod, err := modfile.Parse(modfileName, modfileData, nil)
	if err != nil {
		return fmt.Errorf("unable to parse go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	modulePath := mod.Module.Mod.Path

	// Check we haven't already declared a different module with the same name
	if existingShortName, exists := workspace.ModuleDirs[modulePath]; exists {
		if existingShortName != shortName {
			return fmt.Errorf("redeclaration of module %s as %s - already exists in %s", modulePath, shortName, existingShortName)
		}
		// TODO: here, check module versions are the same
	} else {
		workspace.ModuleDirs[modulePath] = shortName
	}
	if existingModulePath, exists := workspace.Modules[shortName]; exists {
		if existingModulePath != modulePath {
			return fmt.Errorf("cannot copy module %s to %s as it already contains module %s", modulePath, shortName, existingModulePath)
		}
	} else {
		workspace.Modules[shortName] = modulePath
	}

	moduleDstPath := filepath.Join(workspace.WorkspaceDir, shortName)
	err = checkDir(moduleDstPath, true)
	if err != nil {
		return err
	}

	err = cp.Copy(moduleSrcPath, moduleDstPath)
	if err != nil {
		return err
	}

	return nil
}

func (workspace *WorkspaceBuilderImpl) GetLocalModule(modulePath string) (string, bool) {
	shortName, exists := workspace.ModuleDirs[modulePath]
	return shortName, exists
}

/*
This method is used by plugins if they want to copy a locally-defined module into the generated workspace.

The specified relativeModuleSrcPath must point to a valid Go module with a go.mod file, relative to the calling
file's location.
*/
func (workspace *WorkspaceBuilderImpl) AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) error {
	_, callingFile, _, _ := runtime.Caller(1)
	dir, _ := filepath.Split(callingFile)
	moduleSrcPath := filepath.Join(dir, relativeModuleSrcPath)
	return workspace.AddLocalModule(shortName, moduleSrcPath)

}

func (workspace *WorkspaceBuilderImpl) readModfile(moduleSubDir string) (*modfile.File, error) {
	modFileName := filepath.Join(workspace.WorkspaceDir, moduleSubDir, "go.mod")
	modFileData, err := os.ReadFile(modFileName)
	if err != nil {
		return nil, fmt.Errorf("workspace unable to read %s due to %s", modFileName, err.Error())
	}
	f, err := modfile.Parse(modFileName, modFileData, nil)
	if err != nil {
		return nil, fmt.Errorf("workspace unable to parse %s due to %s", modFileName, err.Error())
	}
	if f.Module.Mod.Version == "" {
		f.Module.Mod.Version = "v0.0.0"
	}
	return f, nil
}

func (workspace *WorkspaceBuilderImpl) Visited(name string) bool {
	return workspace.tracker.Visited(name)
}

var goWorkTemplate = `go 1.20

use (
	{{ range $dirName, $moduleName := .Modules }}./{{ $dirName }}
	{{ end }}
)
`

/*
This method should be used by plugins after all modules in a workspace have been combined.

The method will do the following:
  - creates a go.work file in the root of the workspace that points to all of the modules contained therein
  - updates the go.mod files of all contained modules with 'replace' directives for any required modules that exist in the workspace
*/
func (workspace *WorkspaceBuilderImpl) Finish() error {
	t, err := template.New("go.work").Parse(goWorkTemplate)
	if err != nil {
		return err
	}

	// Create the go.work file
	workFileName := filepath.Join(workspace.WorkspaceDir, "go.work")
	f, err := os.OpenFile(workFileName, os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// Generate the file
	err = t.Execute(f, workspace)
	if err != nil {
		return err
	}

	// Parse it to double check it is valid
	fWritten, err := os.ReadFile(workFileName)
	if err != nil {
		return err
	}
	_, err = modfile.ParseWork(workFileName, fWritten, nil)
	if err != nil {
		return fmt.Errorf("generated an invalid go.work file for workspace %v due to %v", workspace.WorkspaceDir, err.Error())
	}

	// Rewrite the go.mod files
	for moduleSubDir, _ := range workspace.Modules {
		// Read in the go.mod file to update
		modFile, err := workspace.readModfile(moduleSubDir)
		if err != nil {
			return err
		}

		// Drop all existing replace directives
		modFile.Replace = nil

		// Check all of the 'require' statements, add 'replace' as needed to redirect to the local subdir.  Also validate the version while we're at it
		for _, requireStmt := range modFile.Require {
			if dependencySubDir, dependencyIsLocal := workspace.ModuleDirs[requireStmt.Mod.Path]; dependencyIsLocal {
				// Read the go.mod of the dependency
				targetModFile, err := workspace.readModfile(dependencySubDir)
				if err != nil {
					return err
				}

				requiredVersion := requireStmt.Mod.Version
				localVersion := targetModFile.Module.Mod.Version
				if localVersion != requiredVersion {
					return fmt.Errorf("dependency version mismatch for module %v which requires module %v version %v but version %v was found in workspace %v", moduleSubDir, dependencySubDir, requiredVersion, localVersion, workspace.WorkspaceDir)
				}

				replacePath := "../" + dependencySubDir
				modFile.AddReplace(requireStmt.Mod.Path, requireStmt.Mod.Version, replacePath, "")
			}
		}

		// Format the new modfile
		data, err := modFile.Format()
		if err != nil {
			return err
		}

		// Overwrite it
		modFileName := filepath.Join(workspace.WorkspaceDir, moduleSubDir, "go.mod")
		err = os.WriteFile(modFileName, data, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
