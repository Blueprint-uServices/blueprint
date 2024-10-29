package gogen

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	cp "github.com/otiai10/copy"
	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
)

// Implements [golang.WorkspaceBuilder].
//
// Creates a golang workspace on the local filesystem
type WorkspaceBuilderImpl struct {
	ir.VisitTrackerImpl
	WorkspaceDir     string            // The directory containing this workspace
	ModuleDirs       map[string]string // map from FQ module name to directory name within WorkspaceDir
	Modules          map[string]string // map from directory name to FQ module name within WorkspaceDir
	GeneratedModules map[string]string // map from directory name to FQ module name within WorkspaceDir
}

// Creates a golang workspace in the specified directory on the local filesystem.
//
// Calls to [CreateModule] will create golang modules in subdirectories of the workspace.  The builder
// will automatically create the necessary go.mod files of modules, and go.work file of the workspace.
//
// The typical usage of this is by plugins such as the [goproc] plugin that accumulate
// golang nodes and generate code to run those nodes.
//
// After calling this method, the returned WorkspaceBuilder can be passed to golang nodes,
// to accumulate modules.
//
// After all modules have been accumulated, the caller should invoke [Finish], which will write the
// go.work file, insert replace directives in go.mod files for sibling modules, and invoke go mod tidy
// to resolve external package dependencies.
//
// Returns an error if the directory already exists.
func NewWorkspaceBuilder(workspaceDir string) (*WorkspaceBuilderImpl, error) {
	workspace := &WorkspaceBuilderImpl{}
	workspace.WorkspaceDir = workspaceDir
	workspace.ModuleDirs = make(map[string]string)
	workspace.Modules = make(map[string]string)
	workspace.GeneratedModules = make(map[string]string)
	return workspace, nil
}

// Implements [golang.WorkspaceBuilder]
func (workspace *WorkspaceBuilderImpl) Info() golang.WorkspaceInfo {
	return golang.WorkspaceInfo{
		Path: workspace.WorkspaceDir,
	}
}

// Implements [golang.WorkspaceBuilder]
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
	err := ioutil.CheckDir(moduleDir, true)
	if err != nil {
		return "", blueprint.Errorf("cannot generate new module %s due to %s", moduleShortName, err.Error())
	}

	// Save the module
	workspace.Modules[moduleShortName] = moduleName
	workspace.ModuleDirs[moduleName] = moduleShortName
	workspace.GeneratedModules[moduleShortName] = moduleName

	// Create the go.mod file
	modfileContents := fmt.Sprintf("module %v\n\ngo 1.22", moduleName)
	modfile := filepath.Join(moduleDir, "go.mod")

	return moduleDir, os.WriteFile(modfile, []byte(modfileContents), 0755)
}

// Implements [golang.WorkspaceBuilder]
func (workspace *WorkspaceBuilderImpl) AddLocalModule(shortName string, moduleSrcPath string) (string, error) {
	// First open and parse the go.mod file to make sure it exists and is valid
	modfileName := filepath.Join(moduleSrcPath, "go.mod")
	modfileData, err := os.ReadFile(modfileName)
	if err != nil {
		return "", blueprint.Errorf("unable to read go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	mod, err := modfile.Parse(modfileName, modfileData, nil)
	if err != nil {
		return "", blueprint.Errorf("unable to parse go.mod for %s at %s due to %s", shortName, modfileName, err.Error())
	}

	modulePath := mod.Module.Mod.Path

	// Check we haven't already declared a different module with the same name
	if existingShortName, exists := workspace.ModuleDirs[modulePath]; exists {
		if existingShortName != shortName {
			return "", blueprint.Errorf("redeclaration of module %s as %s - already exists in %s", modulePath, shortName, existingShortName)
		} else {
			// TODO: here, check module versions are the same
			return filepath.Join(workspace.WorkspaceDir, workspace.ModuleDirs[modulePath]), nil
		}
	} else {
		workspace.ModuleDirs[modulePath] = shortName
	}
	if existingModulePath, exists := workspace.Modules[shortName]; exists {
		if existingModulePath != modulePath {
			return "", blueprint.Errorf("cannot copy module %s to %s as it already contains module %s", modulePath, shortName, existingModulePath)
		} else {
			return filepath.Join(workspace.WorkspaceDir, shortName), nil
		}
	} else {
		workspace.Modules[shortName] = modulePath
	}

	moduleDstPath := filepath.Join(workspace.WorkspaceDir, shortName)
	err = ioutil.CheckDir(moduleDstPath, true)
	if err != nil {
		return "", err
	}

	slog.Info(fmt.Sprintf("Copying local module %s to workspace %s", shortName, workspace.WorkspaceDir))

	return moduleDstPath, cp.Copy(moduleSrcPath, moduleDstPath)
}

// Implements [golang.WorkspaceBuilder]
func (workspace *WorkspaceBuilderImpl) GetLocalModule(modulePath string) (string, bool) {
	shortName, exists := workspace.ModuleDirs[modulePath]
	return shortName, exists
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

var goWorkTemplate = `go 1.22.1

use (
	{{ range $dirName, $moduleName := .Modules }}./{{ $dirName }}
	{{ end }}
)
`

// This method should be called by plugins after all modules in a workspace have been combined.
//
// The method will do the following:
//   - creates a go.work file in the root of the workspace that points to all of the modules contained therein
//   - updates the go.mod files of all contained modules with 'replace' directives for any required modules that exist in the workspace
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

	// Resolve imported packages for generated modules
	for moduleSubDir := range workspace.GeneratedModules {
		workspace.goModTidy(moduleSubDir)
	}

	return nil
}

// Updates the go.mod file for the module in the specified sub directory, to add replace directives
// to all other modules located in the workspace
func (workspace *WorkspaceBuilderImpl) updateModfile(moduleSubDir string, moduleName string) error {
	// Read in the go.mod file to update
	modFile, err := workspace.readModfile(moduleSubDir)
	if err != nil {
		return err
	}

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

func (workspace *WorkspaceBuilderImpl) ImplementsBuildContext() {}
