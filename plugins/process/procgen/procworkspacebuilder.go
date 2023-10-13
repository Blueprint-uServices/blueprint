package procgen

import (
	"os"
	"path/filepath"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ioutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/process"
)

/*
Implements the ProcWorkspaceBuilder interface defined in process/ir.go

The ProcWorkspaceBuilder is used for accumulating process artifacts
*/
type ProcWorkspaceBuilderImpl struct {
	blueprint.VisitTrackerImpl
	WorkspaceDir  string
	ProcDirs      map[string]string      // map from proc name to directory
	BuildScripts  map[string]BuildScript // map from PathInWorkspace to BuildScript
	BuildFileName string                 // the name of the build file generated for the workspace
}

type BuildScript struct {
	HostPath        string // Fully qualified path on the local file system
	PathInWorkspace string // Path within the workspace to the script
	DirInWorkspace  string // Path within the workspace to the directory containing the script
	FileName        string // Filename of the script
}

/*
Creates a new ProcWorkspaceBuilder at the specified output dir.

Will return an error if the workspacedir already exists
*/
func NewProcWorkspaceBuilder(workspaceDir string) (*ProcWorkspaceBuilderImpl, error) {
	workspace := &ProcWorkspaceBuilderImpl{}
	workspace.WorkspaceDir = filepath.Clean(workspaceDir)
	workspace.ProcDirs = make(map[string]string)
	workspace.BuildScripts = make(map[string]BuildScript)
	workspace.BuildFileName = "build.sh"
	return workspace, nil
}

func (builder *ProcWorkspaceBuilderImpl) Info() process.ProcWorkspaceInfo {
	return process.ProcWorkspaceInfo{Path: builder.WorkspaceDir}
}

// Creates a subdirectory and saves its metadata
func (builder *ProcWorkspaceBuilderImpl) CreateProcessDir(name string) (string, error) {
	// Only alphanumeric and underscores are allowed in a proc name
	name = blueprint.CleanName(name)

	// Can't redefine a procdir that already exists
	if _, exists := builder.ProcDirs[name]; exists {
		return "", blueprint.Errorf("process dir %v already exists in output procworkspace %v", name, builder.WorkspaceDir)
	}

	// Create the dir
	procDir := filepath.Join(builder.WorkspaceDir, name)
	if err := ioutil.CheckDir(procDir, true); err != nil {
		return "", blueprint.Errorf("cannot generate process to output workspace %v due to %v", name, err.Error())
	}
	builder.ProcDirs[name] = procDir

	return procDir, nil
}

func (builder *ProcWorkspaceBuilderImpl) AddBuildScript(path string) error {
	pathInWorkspace, err := filepath.Rel(builder.WorkspaceDir, filepath.Clean(path))
	if err != nil {
		return blueprint.Errorf("procworkspace only supports build scripts located within the workspace; got %v which does not reside in %v; error: %v", path, builder.WorkspaceDir, err.Error())
	}

	if _, exists := builder.BuildScripts[pathInWorkspace]; exists {
		return blueprint.Errorf("workspace already contains build script %v", pathInWorkspace)
	}

	dirInWorkspace, fileName := filepath.Split(pathInWorkspace)

	info := BuildScript{
		HostPath:        path,
		PathInWorkspace: filepath.ToSlash(pathInWorkspace),
		DirInWorkspace:  filepath.ToSlash(filepath.Clean(dirInWorkspace)),
		FileName:        filepath.Clean(fileName),
	}
	builder.BuildScripts[pathInWorkspace] = info
	return nil
}

var buildScriptTemplate = `#!/bin/bash

{{range $name, $script := .BuildScripts}}
echo "Executing {{.PathInWorkspace}}"
cd {{.DirInWorkspace}}
chmod +x {{.FileName}}
./{{.FileName}}
cd -
{{end}}
echo "
`

/*
Creates a build.sh file in the root of the proc workspace.

When invoked, the build.sh file will sequentially invoke any
build scripts that were provided by processes in the workspace.

The build.sh will typically be invoked by e.g. a Dockerfile
*/
func (builder *ProcWorkspaceBuilderImpl) Finish() error {
	filename := filepath.Join(builder.WorkspaceDir, builder.BuildFileName)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New(builder.BuildFileName).Parse(buildScriptTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, builder)
}

func (builder *ProcWorkspaceBuilderImpl) ImplementsBuildContext() {}
