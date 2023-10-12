package procgen

import (
	"os"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/process"
)

/*
Implements the ProcWorkspaceBuilder interface defined in process/ir.go

The ProcWorkspaceBuilder is used for accumulating process artifacts
*/
type ProcWorkspaceBuilderImpl struct {
	blueprint.VisitTrackerImpl
	WorkspaceDir string
	ProcDirs     map[string]string // map from proc name to directory
}

/*
Creates a new ProcWorkspaceBuilder at the specified output dir.

Will return an error if the workspacedir already exists
*/
func NewProcWorkspaceBuilder(workspaceDir string) (*ProcWorkspaceBuilderImpl, error) {
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
	workspace := &ProcWorkspaceBuilderImpl{}
	workspace.WorkspaceDir = workspaceDir
	workspace.ProcDirs = make(map[string]string)
	return workspace, nil
}

func (workspace *ProcWorkspaceBuilderImpl) Info() process.ProcWorkspaceInfo {
	return process.ProcWorkspaceInfo{Path: workspace.WorkspaceDir}
}
