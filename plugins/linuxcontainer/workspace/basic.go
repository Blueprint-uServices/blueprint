package workspace

import (
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ioutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/linuxgen"
)

/*
The base implementation of the linux.ProcessWorkspace defined in linux/ir.go

This workspace performs the basic actions that are (presumed to be) common
to all process workspaces:
 (a) gather each process's artifacts into process subdirectories
 (b) allow each process to declare a run command
 (c) allow each process to provide a build file
 (d) generate a root build.sh that invokes each process's build file
 (e) generate a root run.sh that invokes each process's run command

Note that the Docker process workspace extends this workspace to enable
processes to additionally provide Dockerfile build commands in lieu of
a build.sh script
*/

type BasicWorkspace struct {
	blueprint.VisitTrackerImpl

	info linux.ProcessWorkspaceInfo

	ProcDirs map[string]string // map from proc name to directory

	Build *linuxgen.BuildScript
	Run   *linuxgen.RunScript
}

func NewBasicWorkspace(name string, dir string) *BasicWorkspace {
	return &BasicWorkspace{
		info: linux.ProcessWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "basic",
		},
		ProcDirs: make(map[string]string),
		Build:    linuxgen.NewBuildScript(dir, "build.sh"),
		Run:      linuxgen.NewRunScript(name, dir, "run.sh"),
	}
}

func (workspace *BasicWorkspace) Info() linux.ProcessWorkspaceInfo {
	return workspace.info
}

// Creates a subdirectory for a process to output its artifacts.
// Saves the metadata about the process
func (ws *BasicWorkspace) CreateProcessDir(name string) (string, error) {
	// Only alphanumeric and underscores are allowed in a proc name
	name = blueprint.CleanName(name)

	// Can't redefine a procdir that already exists
	if _, exists := ws.ProcDirs[name]; exists {
		return "", blueprint.Errorf("process dir %v already exists in output procworkspace %v", name, ws.info.Path)
	}

	// Create the dir
	procDir := filepath.Join(ws.info.Path, name)
	if err := ioutil.CheckDir(procDir, true); err != nil {
		return "", blueprint.Errorf("cannot generate process to output workspace %v due to %v", name, err.Error())
	}
	ws.ProcDirs[name] = procDir

	return procDir, nil
}

// Adds a build script provided by a process
func (ws *BasicWorkspace) AddBuildScript(path string) error {
	return ws.Build.Add(path)
}

// TODO: this might be unnecessary
func (ws *BasicWorkspace) AddArg(node blueprint.IRNode) {
	ws.Run.Require(node)
}

// Adds a command to the run.sh file for running the specified process node
func (ws *BasicWorkspace) DeclareRunCommand(name string, runfunc string, deps ...blueprint.IRNode) error {
	// Generate the runfunc
	runfunc_impl, err := linuxgen.GenerateRunFunc(name, runfunc, deps...)
	ws.Run.Add(name, runfunc_impl, deps...)
	return err
}

/*
Creates a build.sh and a run.sh file in the root of the proc workspace

When invoked, the build.sh file will sequentially invoke any
build scripts that were provided by processes in the workspace.

The build.sh will typically be invoked by e.g. a Dockerfile
*/
func (ws *BasicWorkspace) Finish() error {
	// Generate the build.sh
	if err := ws.Build.GenerateBuildScript(); err != nil {
		return err
	}

	// Generate the run.sh
	return ws.Run.GenerateRunScript()
}

func (ws *BasicWorkspace) ImplementsBuildContext()     {}
func (ws *BasicWorkspace) ImplementsProcessWorkspace() {}
