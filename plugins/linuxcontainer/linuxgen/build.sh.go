package linuxgen

import (
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
)

/*
Within a process workspace, a single build.sh file will be generated
at the root of the workspace.  This build.sh file will invoke
individual build scripts of each of the processes in the workspace.

The default linux workspace uses this functionality and
generates a build.sh file.

The Docker container workspace also uses this functionality as a default,
but some processes can instead provide Dockerfile commands, so often the
generated build.sh will be empty.
*/

type BuildScript struct {
	WorkspaceDir string
	FileName     string
	FilePath     string
	Scripts      map[string]*scriptInfo
}

/*
Creates a new build.sh that will invoke multiple build scripts
of processes in subdirectories of the workspace
*/
func NewBuildScript(workspaceDir, fileName string) *BuildScript {
	return &BuildScript{
		WorkspaceDir: workspaceDir,
		FileName:     fileName,
		FilePath:     filepath.Join(workspaceDir, fileName),
		Scripts:      make(map[string]*scriptInfo),
	}
}

/*
Adds a process's build script to the workspace's build.sh

filePath should be a fully qualified path to a build script
that resides within a subdirectory of the workspace

Returns an error if the script resides outside of the workspace
*/
func (d *BuildScript) Add(filePath string) error {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return blueprint.Errorf("invalid build script path %v", filePath)
	}
	pathInWorkspace, err := filepath.Rel(d.WorkspaceDir, filepath.Clean(filePath))
	if err != nil {
		return blueprint.Errorf("workspace build.sh only supports build scripts located within the workspace; got %v which does not reside in %v; error: %v", filePath, d.WorkspaceDir, err.Error())
	}

	dirInWorkspace, fileName := filepath.Split(pathInWorkspace)

	d.Scripts[filePath] = &scriptInfo{
		HostPath:        filePath,
		PathInWorkspace: filepath.ToSlash(pathInWorkspace),
		DirInWorkspace:  filepath.ToSlash(filepath.Clean(dirInWorkspace)),
		FileName:        fileName,
	}
	return nil
}

func (d *BuildScript) GenerateBuildScript() error {
	return ExecuteTemplateToFile("build.sh", buildScriptTemplate, d, d.FilePath)
}

var buildScriptTemplate = `#!/bin/bash

{{range $name, $script := .Scripts}}
echo "Executing {{.PathInWorkspace}}"
cd {{.DirInWorkspace}}
chmod +x {{.FileName}}
./{{.FileName}}
cd -
{{end}}
`

// A build script in a process subdirectory
type scriptInfo struct {
	HostPath        string // Fully qualified path on the local file system
	PathInWorkspace string // Path within the workspace to the script
	DirInWorkspace  string // Path within the workspace to the directory containing the script
	FileName        string // Filename of the script
}
