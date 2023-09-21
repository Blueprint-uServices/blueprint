package gogen

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"path/filepath"
	"runtime"

	cp "github.com/otiai10/copy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
)

/*
Implements the WorkspaceBuilder interface defined in golang/ir.go

The WorkspaceBuilder is used for accumulating local module directories into a golang workspace.
*/
type WorkspaceBuilderImpl struct {
	golang.WorkspaceBuilder
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
	workspaceDir, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, blueprint.Errorf("invalid workspace dir %v", workspaceDir)
	}
	if IsDir(workspaceDir) {
		return nil, blueprint.Errorf("workspace %s already exists", workspaceDir)
	}
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		return nil, blueprint.Errorf("unable to create workspace %s due to %s", workspaceDir, err.Error())
	}
	workspace := &WorkspaceBuilderImpl{}
	workspace.WorkspaceDir = workspaceDir
	workspace.ModuleDirs = make(map[string]string)
	workspace.Modules = make(map[string]string)
	return workspace, nil
}

func (workspace *WorkspaceBuilderImpl) Info() golang.WorkspaceInfo {
	return golang.WorkspaceInfo{
		Path: workspace.WorkspaceDir,
	}
}

func (workspace *WorkspaceBuilderImpl) Visit(nodes []blueprint.IRNode) error {
	for _, node := range nodes {
		if n, valid := node.(golang.ProvidesModule); valid {
			err := n.AddToWorkspace(workspace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (workspace *WorkspaceBuilderImpl) CreateModule(moduleName string, moduleVersion string) (string, error) {
	// Don't currently support multiple versions of the same module
	if _, moduleExists := workspace.ModuleDirs[moduleName]; moduleExists {
		return "", blueprint.Errorf("module %v %v already exists in output workspace %v", moduleName, moduleVersion, workspace.WorkspaceDir)
	}

	// Find an unused subdirectory for the module
	splits := strings.Split(moduleName, "/")
	moduleShortName := splits[len(splits)-1]
	moduleSubDir := moduleShortName
	for i := 0; ; i += 1 {
		if _, subDirInUse := workspace.Modules[moduleSubDir]; !subDirInUse {
			break
		}
		moduleSubDir = fmt.Sprintf("%s%v", moduleShortName, i)
	}

	// Create output directory
	moduleDir := filepath.Join(workspace.WorkspaceDir, moduleShortName)
	err := CheckDir(moduleDir, true)
	if err != nil {
		return "", blueprint.Errorf("cannot generate new module %s due to %s", moduleShortName, err.Error())
	}

	// Save the module
	workspace.Modules[moduleShortName] = moduleName
	workspace.ModuleDirs[moduleName] = moduleShortName

	// Create the go.mod file
	modfileContents := fmt.Sprintf("module %v\n\ngo 1.20", moduleName)
	modfile := filepath.Join(moduleDir, "go.mod")
	err = os.WriteFile(modfile, []byte(modfileContents), 0755)
	return moduleDir, err
}

func (workspace *WorkspaceBuilderImpl) AddLocalModule(shortName string, moduleSrcPath string) error {
	// First open and parse the go.mod file to make sure it exists and is valid
	modfileName := filepath.Join(moduleSrcPath, "go.mod")
	modfileData, err := os.ReadFile(modfileName)
	if err != nil {
		return blueprint.Errorf("unable to read go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	mod, err := modfile.Parse(modfileName, modfileData, nil)
	if err != nil {
		return blueprint.Errorf("unable to parse go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	modulePath := mod.Module.Mod.Path

	// Check we haven't already declared a different module with the same name
	if existingShortName, exists := workspace.ModuleDirs[modulePath]; exists {
		if existingShortName != shortName {
			return blueprint.Errorf("redeclaration of module %s as %s - already exists in %s", modulePath, shortName, existingShortName)
		} else {
			// TODO: here, check module versions are the same
			return nil
		}
	} else {
		workspace.ModuleDirs[modulePath] = shortName
	}
	if existingModulePath, exists := workspace.Modules[shortName]; exists {
		if existingModulePath != modulePath {
			return blueprint.Errorf("cannot copy module %s to %s as it already contains module %s", modulePath, shortName, existingModulePath)
		} else {
			return nil
		}
	} else {
		workspace.Modules[shortName] = modulePath
	}

	moduleDstPath := filepath.Join(workspace.WorkspaceDir, shortName)
	err = CheckDir(moduleDstPath, true)
	if err != nil {
		return err
	}

	return cp.Copy(moduleSrcPath, moduleDstPath)
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
		return nil, blueprint.Errorf("workspace unable to read %s due to %s", modFileName, err.Error())
	}
	f, err := modfile.Parse(modFileName, modFileData, nil)
	if err != nil {
		return nil, blueprint.Errorf("workspace unable to parse %s due to %s", modFileName, err.Error())
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
	// Generate the go.work file
	workFileName := filepath.Join(workspace.WorkspaceDir, "go.work")
	err := ExecuteTemplateToFile("go.work", goWorkTemplate, workspace, workFileName)
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
		return blueprint.Errorf("generated an invalid go.work file for workspace %v due to %v", workspace.WorkspaceDir, err.Error())
	}

	// Rewrite the go.mod files to redirect to local modules
	for moduleSubDir, moduleName := range workspace.Modules {
		workspace.updateModfile(moduleSubDir, moduleName)
	}

	// Resolve imported packages
	for moduleSubDir := range workspace.Modules {
		workspace.goModTidy(moduleSubDir)
	}

	return nil
}

/*
Updates the go.mod file for the module in the specified sub directory, to add replace directives
to all other modules located in the workspace
*/
func (workspace *WorkspaceBuilderImpl) updateModfile(moduleSubDir string, moduleName string) error {
	// Read in the go.mod file to update
	modFile, err := workspace.readModfile(moduleSubDir)
	if err != nil {
		return err
	}

	// Drop all existing replace directives
	modFile.Replace = nil

	// Now we add replace directives
	for otherModuleSubDir, otherModuleName := range workspace.Modules {
		if moduleName == otherModuleName {
			continue
		}
		otherModfile, err := workspace.readModfile(otherModuleSubDir)
		if err != nil {
			return err
		}

		// Add a replace for everything, regardless of whether the module uses it
		modFile.AddReplace(otherModfile.Module.Mod.Path, "", "../"+otherModuleSubDir, "")
	}

	// Format and overwrite the new modfile
	data, err := modFile.Format()
	if err != nil {
		return err
	}
	modFileName := filepath.Join(workspace.WorkspaceDir, moduleSubDir, "go.mod")
	return os.WriteFile(modFileName, data, 0755)
}

func rel(path string) string {
	pwd, err := os.Getwd()
	if err != nil {
		return path
	}
	s, err := filepath.Rel(pwd, path)
	if err != nil {
		return path
	}
	return s
}

func (workspace *WorkspaceBuilderImpl) goModTidy(moduleSubDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = filepath.Join(workspace.WorkspaceDir, moduleSubDir)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	slog.Info(fmt.Sprintf("go mod tidy (%v)", rel(cmd.Dir)))
	return cmd.Run()
}
