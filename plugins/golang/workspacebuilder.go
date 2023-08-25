package golang

import (
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"golang.org/x/mod/modfile"
)

/*
This struct is used by plugins if they want to collect and combine Golang code and modules.

It collects together local modules into a single output workspace directory.  It also allows
for plugins to dynamically generate module code.
*/
type WorkspaceBuilder struct {
	VisitTracker
	WorkspaceDir string            // The directory containing this workspace
	ModuleDirs   map[string]string // map from FQ module name to directory name within WorkspaceDir
	Modules      map[string]string // map from directory name to FQ module name within WorkspaceDir
}

/*
Creates a new WorkspaceBuilder at the specified output dir.

Will return an error if the workspacedir already exists
*/
func NewWorkspaceBuilder(workspaceDir string) (*WorkspaceBuilder, error) {
	if isDir(workspaceDir) {
		return nil, fmt.Errorf("workspace %s already exists", workspaceDir)
	}
	err := os.Mkdir(workspaceDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace %s due to %s", workspaceDir, err.Error())
	}
	workspace := &WorkspaceBuilder{}
	workspace.visited = make(map[string]any)
	workspace.WorkspaceDir = workspaceDir
	workspace.ModuleDirs = make(map[string]string)
	workspace.Modules = make(map[string]string)
	return workspace, nil
}

/*
This method is used by plugins if they want to copy a locally-defined module into the generated workspace.

The specified moduleSrcPath must point to a valid Go module with a go.mod file.
*/
func (workspace *WorkspaceBuilder) AddLocalModule(shortName string, moduleSrcPath string) error {
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
	if err != nil {
		return err
	}

	// Check we haven't already declared a different module with the same name
	if existingShortName, exists := workspace.ModuleDirs[modulePath]; exists {
		if existingShortName != shortName {
			return fmt.Errorf("redeclaration of module %s as %s - already exists in %s", modulePath, shortName, existingShortName)
		}
	}
	if existingModulePath, exists := workspace.Modules[shortName]; exists {
		if existingModulePath != modulePath {
			return fmt.Errorf("cannot copy module %s to %s as it already contains module %s", modulePath, shortName, existingModulePath)
		}
	}

	// Copy the module to the destination -- for now ignore possibility of a different version of the same module (TODO)
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

func (workspace *WorkspaceBuilder) Finish() error {
	// For now, nothing to do
	return nil
}
